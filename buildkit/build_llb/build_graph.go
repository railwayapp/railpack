package build_llb

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/util/system"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/railwayapp/railpack/buildkit/graph"
	"github.com/railwayapp/railpack/core/plan"
)

type BuildGraph struct {
	graph      *graph.Graph
	BaseState  *llb.State
	CacheStore *BuildKitCacheStore
	Plan       *plan.BuildPlan
	Platform   *specs.Platform
	LocalState *llb.State

	secretsFile     *llb.State
	usedSecretsBase *llb.State
}

type BuildGraphOutput struct {
	State    *llb.State
	GraphEnv BuildEnvironment
}

func NewBuildGraph(plan *plan.BuildPlan, baseState *llb.State, localState *llb.State, cacheStore *BuildKitCacheStore, secretsHash string, platform *specs.Platform) (*BuildGraph, error) {
	var secretsFile *llb.State
	if secretsHash != "" {
		st := llb.Scratch().File(llb.Mkfile("/secrets-hash", 0644, []byte(secretsHash)), llb.WithCustomName("[railpack] secrets hash"))
		secretsFile = &st
	}
	usedSecretsBase := llb.Image("alpine:latest", llb.WithCustomName("[railpack] loading secrets"))

	g := &BuildGraph{
		graph:      graph.NewGraph(),
		BaseState:  baseState,
		CacheStore: cacheStore,
		Plan:       plan,
		Platform:   platform,
		LocalState: localState,

		secretsFile:     secretsFile,
		usedSecretsBase: &usedSecretsBase,
	}

	// Create a node for each step
	for i := range plan.Steps {
		step := &plan.Steps[i]
		node := &StepNode{
			Step:      step,
			Processed: false,
			OutputEnv: NewGraphEnvironment(),
		}

		g.graph.AddNode(node)
	}

	// Add dependencies to each node
	for _, node := range g.graph.GetNodes() {
		llbNode := node.(*StepNode)
		for _, depName := range llbNode.Step.DependsOn {
			if depNode, exists := g.graph.GetNode(depName); exists {
				parents := llbNode.GetParents()
				parents = append(parents, depNode)
				llbNode.SetParents(parents)

				children := depNode.GetChildren()
				children = append(children, node)
				depNode.SetChildren(children)
			}
		}
	}

	g.graph.ComputeTransitiveDependencies()
	return g, nil
}

// GenerateLLB generates the LLB state for the build graph
func (g *BuildGraph) GenerateLLB() (*BuildGraphOutput, error) {
	// Get processing order using topological sort
	order, err := g.graph.ComputeProcessingOrder()
	if err != nil {
		return nil, err
	}

	// Process all nodes in order
	for _, node := range order {
		llbNode := node.(*StepNode)
		if err := g.processNode(llbNode); err != nil {
			return nil, err
		}
	}

	// Find all leaf nodes and get their states
	var leafNodes []*StepNode
	graphEnv := NewGraphEnvironment()

	for _, node := range g.graph.GetNodes() {
		llbNode := node.(*StepNode)
		if len(llbNode.GetChildren()) == 0 && llbNode.State != nil {
			leafNodes = append(leafNodes, llbNode)
			graphEnv.Merge(llbNode.OutputEnv)
		}
	}

	// If no leaf states, return base state
	if len(leafNodes) == 0 {
		return &BuildGraphOutput{
			State:    g.BaseState,
			GraphEnv: graphEnv,
		}, nil
	}

	// If only one leaf state, return it
	if len(leafNodes) == 1 {
		return &BuildGraphOutput{
			State:    leafNodes[0].State,
			GraphEnv: graphEnv,
		}, nil
	}

	result := g.mergeNodes(leafNodes)

	return &BuildGraphOutput{
		State:    &result,
		GraphEnv: graphEnv,
	}, nil
}

// mergeNodes merges the states of the given nodes into a single state
// This essentially creates a scratch file system and then copies the contents of each node's state into it
func (g *BuildGraph) mergeNodes(nodes []*StepNode) llb.State {
	stateNames := []string{}
	for _, node := range nodes {
		stateNames = append(stateNames, node.Step.Name)
	}

	states := []llb.State{}
	for _, node := range nodes {
		states = append(states, *node.State)
	}

	result := llb.Scratch()
	for i, state := range states {
		result = result.File(llb.Copy(state, "/", "/", &llb.CopyInfo{
			CreateDestPath: true,
			FollowSymlinks: true,
			AllowWildcard:  true,
		}), llb.WithCustomNamef("copy from %s", stateNames[i]))
	}

	return result
}

// processNode processes a node and its parents to determine the state to build upon
func (g *BuildGraph) processNode(node *StepNode) error {
	// If already processed, we're done
	if node.Processed {
		return nil
	}

	// Check if all parents are processed
	for _, parent := range node.GetParents() {
		parentNode := parent.(*StepNode)
		if !parentNode.Processed {
			// If this node is marked in-progress, we have a dependency violation
			if node.InProgress {
				return fmt.Errorf("dependency violation: %s waiting for unprocessed parent %s",
					node.Step.Name, parentNode.Step.Name)
			}

			// Mark this node as in-progress and process the parent
			node.InProgress = true
			if err := g.processNode(parentNode); err != nil {
				node.InProgress = false
				return err
			}
			node.InProgress = false
		}
	}

	// Determine the state to build upon
	var currentState *llb.State
	currentGraphEnv := NewGraphEnvironment()

	// Merge the output envs of all the parent nodes
	for _, parent := range node.GetParents() {
		parentNode := parent.(*StepNode)
		currentGraphEnv.Merge(parentNode.OutputEnv)
	}

	if len(node.GetParents()) == 0 {
		currentState = g.BaseState
	} else {
		parentNodes := make([]*StepNode, len(node.GetParents()))
		for i, parent := range node.GetParents() {
			parentNode := parent.(*StepNode)

			if parentNode.State == nil {
				return fmt.Errorf("parent %s of %s has nil state",
					parentNode.Step.Name, node.Step.Name)
			}

			parentNodes[i] = parentNode
		}

		merged := g.mergeNodes(parentNodes)
		currentState = &merged
	}

	// if len(node.GetParents()) == 0 {
	// 	currentState = g.BaseState
	// } else if len(node.GetParents()) == 1 {
	// 	// If only one parent, use its state directly
	// 	parentNode := node.GetParents()[0].(*StepNode)
	// 	currentState = parentNode.State
	// } else {
	// 	// If multiple parents, merge their states
	// 	parentNodes := make([]*StepNode, len(node.GetParents()))

	// 	for i, parent := range node.GetParents() {
	// 		parentNode := parent.(*StepNode)
	// 		if parentNode.State == nil {
	// 			return fmt.Errorf("parent %s of %s has nil state",
	// 				parentNode.Step.Name, node.Step.Name)
	// 		}

	// 		parentNodes[i] = parentNode
	// 	}

	// 	merged := g.mergeNodes(parentNodes)
	// 	currentState = &merged
	// }

	node.InputEnv = currentGraphEnv

	// Convert this node's step to LLB
	stepState, err := g.convertNodeToLLB(node, currentState)
	if err != nil {
		return err
	}

	node.State = stepState
	node.Processed = true

	return nil
}

// convertNodeToLLB converts a step node to an LLB state
func (g *BuildGraph) convertNodeToLLB(node *StepNode, baseState *llb.State) (*llb.State, error) {
	state := *baseState
	state = state.Dir("/app")

	state, err := g.getNodeStartingState(state, node)
	if err != nil {
		return nil, err
	}

	// Process the step commands
	if len(node.Step.Commands) > 0 {
		for _, cmd := range node.Step.Commands {
			var err error
			state, err = g.convertCommandToLLB(node, cmd, state, node.Step)
			if err != nil {
				return nil, err
			}
		}
	}

	if len(node.Step.Outputs) > 0 {
		if len(node.Step.Outputs) == 1 {
			// Special case for a single output - avoid creating intermediate state
			output := node.Step.Outputs[0]
			state = baseState.File(llb.Copy(state, output, output, &llb.CopyInfo{
				CreateDestPath:     true,
				AllowWildcard:      true,
				AllowEmptyWildcard: true,
				FollowSymlinks:     true,
			}), llb.WithCustomNamef("copying %s", output))
		} else {
			// For multiple outputs, copy them sequentially to maintain efficiency
			result := llb.Scratch()

			// Copy each output path individually
			for _, output := range node.Step.Outputs {
				result = result.File(llb.Copy(state, output, output, &llb.CopyInfo{
					CreateDestPath:     true,
					AllowWildcard:      true,
					AllowEmptyWildcard: true,
					FollowSymlinks:     true,
				}), llb.WithCustomNamef("copying %s", output))
			}

			// Apply to base state
			state = baseState.File(llb.Copy(result, "/", "/", &llb.CopyInfo{
				CreateDestPath: true,
				FollowSymlinks: true,
				AllowWildcard:  true,
			}), llb.WithCustomNamef("combined outputs: %s", node.Step.Name))
		}
	}

	return &state, nil
}

// Adds the input environment to the base state of the node
// This includes things like the environment variables and accumulated paths
func (g *BuildGraph) getNodeStartingState(baseState llb.State, node *StepNode) (llb.State, error) {
	envVars := make(map[string]string)

	// Collect all environment variables first
	for k, v := range node.InputEnv.EnvVars {
		envVars[k] = v
		node.OutputEnv.AddEnvVar(k, v)
	}
	for k, v := range node.Step.Variables {
		envVars[k] = v
		node.OutputEnv.AddEnvVar(k, v)
	}

	var state llb.State
	if node.Step.StartingImage != "" {
		state = llb.Image(node.Step.StartingImage, llb.Platform(*g.Platform)).Dir("/app")
	} else {
		state = baseState
	}

	for _, k := range slices.Sorted(maps.Keys(envVars)) {
		state = state.AddEnv(k, envVars[k])
	}

	if len(node.InputEnv.PathList) > 0 {
		pathString := strings.Join(node.InputEnv.PathList, ":")
		state = state.AddEnvf("PATH", "%s:%s", pathString, system.DefaultPathEnvUnix)
		for _, path := range node.InputEnv.PathList {
			node.OutputEnv.AddPath(path)
		}
	}

	return state, nil

	// state := baseState

	// // If a custom image is specified, we build off that instead of the parent's state
	// if node.Step.StartingImage != "" {
	// 	state = llb.Image(node.Step.StartingImage, llb.Platform(*g.Platform)).Dir("/app")
	// }

	// // Add all the variables coming from the parent nodes
	// for _, k := range slices.Sorted(maps.Keys(node.InputEnv.EnvVars)) {
	// 	v := node.InputEnv.EnvVars[k]
	// 	state = state.AddEnv(k, v)
	// 	node.OutputEnv.AddEnvVar(k, v)
	// }

	// // Add all the variables coming from the step
	// for _, k := range slices.Sorted(maps.Keys(node.Step.Variables)) {
	// 	v := node.Step.Variables[k]
	// 	state = state.AddEnv(k, v)
	// 	node.OutputEnv.AddEnvVar(k, v)
	// }

	// // Add all the paths coming from the parent nodes
	// for _, path := range node.InputEnv.PathList {
	// 	newState, err := g.convertCommandToLLB(node, plan.PathCommand{Path: path}, state, node.Step)
	// 	if err != nil {
	// 		return state, err
	// 	}
	// 	state = newState
	// }

	// return state, nil
}

func (g *BuildGraph) convertCommandToLLB(node *StepNode, cmd plan.Command, state llb.State, step *plan.Step) (llb.State, error) {
	switch cmd := cmd.(type) {
	case plan.ExecCommand:
		return g.convertExecCommandToLLB(node, cmd, state)
	case plan.PathCommand:
		return g.convertPathCommandToLLB(node, cmd, state)
	case plan.CopyCommand:
		return g.convertCopyCommandToLLB(cmd, state)
	case plan.FileCommand:
		return g.convertFileCommandToLLB(cmd, state, step)
	}
	return state, nil
}

// convertExecCommandToLLB converts an exec command to an LLB state
func (g *BuildGraph) convertExecCommandToLLB(node *StepNode, cmd plan.ExecCommand, state llb.State) (llb.State, error) {
	opts := []llb.RunOption{llb.Shlex(cmd.Cmd)}
	if cmd.CustomName != "" {
		opts = append(opts, llb.WithCustomName(cmd.CustomName))
	}

	if len(node.Step.Secrets) > 0 {
		secretOpts := []llb.RunOption{}
		for _, secret := range g.Plan.Secrets {
			secretOpts = append(secretOpts, llb.AddSecret(secret, llb.SecretID(secret), llb.SecretAsEnv(true), llb.SecretAsEnvName(secret)))
		}
		opts = append(opts, secretOpts...)

		if g.secretsFile != nil {
			secretOpts = g.getSecretMountOptions(node, secretOpts)
			opts = append(opts, secretOpts...)
		}
	}

	if len(node.Step.Caches) > 0 {
		cacheOpts, err := g.getCacheMountOptions(node.Step.Caches)
		if err != nil {
			return state, err
		}
		opts = append(opts, cacheOpts...)
	}

	s := state.Run(opts...).Root()
	return s, nil
}

// convertPathCommandToLLB converts a path command to an LLB state
func (g *BuildGraph) convertPathCommandToLLB(node *StepNode, cmd plan.PathCommand, state llb.State) (llb.State, error) {
	node.OutputEnv.AddPath(cmd.Path)
	pathList := node.getPathList()
	pathString := strings.Join(pathList, ":")

	s := state.AddEnvf("PATH", "%s:%s", pathString, system.DefaultPathEnvUnix)
	return s, nil
}

// convertCopyCommandToLLB converts a copy command to an LLB state
func (g *BuildGraph) convertCopyCommandToLLB(cmd plan.CopyCommand, state llb.State) (llb.State, error) {
	var src llb.State
	if cmd.Image != "" {
		src = llb.Image(cmd.Image, llb.Platform(*g.Platform))
	} else {
		src = *g.LocalState
	}

	s := state.File(llb.Copy(src, cmd.Src, cmd.Dest, &llb.CopyInfo{
		CreateDestPath:      true,
		FollowSymlinks:      true,
		CopyDirContentsOnly: false,
		AllowWildcard:       true,
		AllowEmptyWildcard:  true,
	}))

	return s, nil
}

// convertFileCommandToLLB converts a file command to an LLB state
func (g *BuildGraph) convertFileCommandToLLB(cmd plan.FileCommand, state llb.State, step *plan.Step) (llb.State, error) {
	asset, ok := step.Assets[cmd.Name]
	if !ok {
		return state, fmt.Errorf("asset %q not found", cmd.Name)
	}

	// Create parent directories for the file
	parentDir := filepath.Dir(cmd.Path)
	if parentDir != "/" {
		s := state.File(llb.Mkdir(parentDir, 0755, llb.WithParents(true)))
		state = s
	}

	var mode os.FileMode = 0644
	if cmd.Mode != 0 {
		mode = cmd.Mode
	}

	fileAction := llb.Mkfile(cmd.Path, mode, []byte(asset))
	s := state.File(fileAction)
	if cmd.CustomName != "" {
		s = state.File(fileAction, llb.WithCustomName(cmd.CustomName))
	}

	return s, nil
}

func (g *BuildGraph) getSecretMountOptions(node *StepNode, secretOpts []llb.RunOption) []llb.RunOption {
	opts := []llb.RunOption{}

	if len(node.Step.Secrets) == 0 || g.secretsFile == nil {
		return secretOpts
	}

	// If all secrets are included, we can just copy the secrets hash file to the new state
	if slices.Contains(node.Step.Secrets, "*") {
		opts = append(opts, llb.AddMount("/secrets-hash", *g.secretsFile))
	} else {
		// If not all secrets are included, we want to compute the hash of only the used secrets
		secrets := slices.Clone(node.Step.Secrets)
		slices.Sort(secrets)
		secretsString := "$" + strings.Join(secrets, " $")

		// Hash all the secrets into a single file
		hashCommand := fmt.Sprintf("sh -c 'echo \"%s\" | sha256sum > /used-secrets-hash'", secretsString)

		usedSecretsState := g.usedSecretsBase.
			File(llb.Copy(*g.secretsFile, "/secrets-hash", "/secrets-hash"),
				llb.WithCustomName("[railpack] copy secrets hash")).
			Run(append([]llb.RunOption{
				llb.Shlex(hashCommand),
				llb.WithCustomName("[railpack] hash used secrets")},
				secretOpts...)...).Root()

		usedSecretsHash := llb.Scratch().File(
			llb.Copy(usedSecretsState, "/used-secrets-hash", "/used-secrets-hash"),
			llb.WithCustomName("[railpack] copy used secrets hash"))

		opts = append(secretOpts, llb.AddMount("/used-secrets-hash", usedSecretsHash))
	}

	return opts
}

// getCacheMountOptions returns the llb.RunOption slice for the given cache keys
func (g *BuildGraph) getCacheMountOptions(cacheKeys []string) ([]llb.RunOption, error) {
	var opts []llb.RunOption

	for _, cacheKey := range cacheKeys {
		if planCache, ok := g.Plan.Caches[cacheKey]; ok {
			cache := g.CacheStore.GetCache(cacheKey, planCache)
			cacheType := llb.CacheMountShared
			if planCache.Type == plan.CacheTypeLocked {
				cacheType = llb.CacheMountLocked
			}

			opts = append(opts,
				llb.AddMount(planCache.Directory, *cache.cacheState, llb.AsPersistentCacheDir(cache.cacheKey, cacheType)),
			)
		} else {
			return nil, fmt.Errorf("cache with key %q not found", cacheKey)
		}
	}
	return opts, nil
}
