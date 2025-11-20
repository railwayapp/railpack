// converts a railpack build plan to a BuildKit LLB state and image config
package buildkit

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/util/system"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/railwayapp/railpack/buildkit/build_llb"
	p "github.com/railwayapp/railpack/core/plan"
)

type ConvertPlanOptions struct {
	BuildPlatform specs.Platform

	// Hash of all the secrets values that can be used to invalidate the layer cache when a secret changes
	SecretsHash string

	// Unique value prepended to all cache mount keys
	CacheKey string

	// BuildKit session ID
	SessionID string

	// Token used to make authenticated API requests to GitHub to increase rate limits
	GitHubToken string

	// Dockerignore context for filtering local files
	DockerignoreContext *p.DockerignoreContext
}

const (
	WorkingDir = "/app"
)

func ConvertPlanToLLB(plan *p.BuildPlan, opts ConvertPlanOptions) (*llb.State, *Image, error) {
	platform := opts.BuildPlatform

	localOpts := []llb.LocalOption{
		llb.SharedKeyHint("local"),
		llb.SessionID(opts.SessionID),
		llb.WithCustomName("loading ."),
		llb.FollowPaths([]string{"."}),
	}

	excludePatterns := opts.DockerignoreContext.GetExcludePatterns()
	if len(excludePatterns) > 0 {
		localOpts = append(localOpts, llb.ExcludePatterns(excludePatterns))
	}

	localState := llb.Local("context", localOpts...)

	cacheStore := build_llb.NewBuildKitCacheStore(opts.CacheKey)
	graph, err := build_llb.NewBuildGraph(plan, &localState, cacheStore, opts.SecretsHash, &platform, opts.GitHubToken)
	if err != nil {
		return nil, nil, err
	}

	graphOutput, err := graph.GenerateLLB()
	if err != nil {
		return nil, nil, err
	}

	state := getStartState(*graphOutput.State)
	imageEnv := getImageEnv(graphOutput, plan)

	startCommand := plan.Deploy.StartCmd
	if startCommand == "" {
		startCommand = "/bin/bash"
	}

	image := Image{
		Image: specs.Image{
			Platform: specs.Platform{
				OS:           platform.OS,
				Architecture: platform.Architecture,
			},
			RootFS: specs.RootFS{
				Type: "layers",
			},
		},
		Variant: platform.Variant,
		Config: specs.ImageConfig{
			Env:        imageEnv,
			WorkingDir: WorkingDir,
			Entrypoint: []string{"/bin/bash", "-c"},
			Cmd:        []string{startCommand},
		},
	}

	return &state, &image, nil
}

func getStartState(buildState llb.State) llb.State {
	startState := buildState.Dir(WorkingDir)
	return startState
}

func getImageEnv(graphOutput *build_llb.BuildGraphOutput, plan *p.BuildPlan) []string {
	paths := []string{}
	paths = append(paths, plan.Deploy.Paths...)
	paths = append(paths, graphOutput.GraphEnv.PathList...)
	paths = append(paths, system.DefaultPathEnvUnix)
	slices.Sort(paths)
	pathString := strings.Join(paths, ":")

	envMap := make(map[string]string, len(graphOutput.GraphEnv.EnvVars)+len(plan.Deploy.Variables)+1)
	maps.Copy(envMap, graphOutput.GraphEnv.EnvVars)
	maps.Copy(envMap, plan.Deploy.Variables)

	envMap["PATH"] = pathString

	envVars := make([]string, 0, len(envMap))
	for _, k := range slices.Sorted(maps.Keys(envMap)) {
		v := envMap[k]
		envVars = append(envVars, fmt.Sprintf("%s=%s", k, v))
	}

	return envVars
}
