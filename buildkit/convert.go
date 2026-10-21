// converts a railpack build plan to a BuildKit LLB state and image config
package buildkit

import (
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/util/system"
	dockerspec "github.com/moby/docker-image-spec/specs-go/v1"
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
	// Do not use cache when building
	NoCache bool
}

const WorkingDir = "/app"

func ConvertPlanToLLB(plan *p.BuildPlan, opts ConvertPlanOptions) (*llb.State, *Image, error) {
	platform := opts.BuildPlatform

	// by default, the whole directory is transferred into context, we don't need to explicitly include it
	localOpts := []llb.LocalOption{
		llb.SharedKeyHint("local"),
		llb.SessionID(opts.SessionID),
		llb.WithCustomName("loading ."),
	}

	// note that exclude patterns can contain inverse (inclusions) patterns. The llb.IncludePatterns should *not* be used for this
	if len(plan.Exclude) > 0 {
		localOpts = append(localOpts, llb.ExcludePatterns(plan.Exclude))
	}

	localState := llb.Local("context", localOpts...)

	cacheStore := build_llb.NewBuildKitCacheStore(opts.CacheKey)
	graph, err := build_llb.NewBuildGraph(plan, &localState, cacheStore, opts.SecretsHash, &platform, opts.GitHubToken, opts.NoCache)
	if err != nil {
		return nil, nil, err
	}

	graphOutput, err := graph.GenerateLLB()
	if err != nil {
		return nil, nil, err
	}

	state := getStartState(*graphOutput.State)
	imageEnv := getImageEnv(graphOutput, plan)

	entrypoint, cmd := containerProcess(plan.Deploy)
	healthcheck, err := plan.Deploy.Healthcheck.ToImageConfig()
	if err != nil {
		return nil, nil, err
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
		Config: dockerspec.DockerOCIImageConfig{
			ImageConfig: specs.ImageConfig{
				Env:        imageEnv,
				WorkingDir: WorkingDir,
				Entrypoint: entrypoint,
				Cmd:        cmd,
			},
			DockerOCIImageConfigExt: dockerspec.DockerOCIImageConfigExt{
				Healthcheck: healthcheck,
			},
		},
	}

	return &state, &image, nil
}

// Uses /bin/bash -c when no entrypoint is configured so the start command can use shell syntax.
func containerProcess(deploy p.Deploy) (entrypoint []string, cmd []string) {
	if len(deploy.Entrypoint) == 0 {
		startCommand := deploy.StartCmd
		if startCommand == "" {
			// Start an interactive shell when no command is configured.
			startCommand = "/bin/bash"
		}
		return []string{"/bin/bash", "-c"}, []string{startCommand}
	}

	// No start command means the entrypoint array is the whole process.
	if deploy.StartCmd == "" {
		return deploy.Entrypoint, nil
	}

	return deploy.Entrypoint, []string{deploy.StartCmd}
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

	envMap := make(map[string]string, len(graphOutput.GraphEnv.EnvVars)+len(plan.Deploy.Variables)+2)
	maps.Copy(envMap, graphOutput.GraphEnv.EnvVars)
	maps.Copy(envMap, plan.Deploy.Variables)

	envMap["PATH"] = pathString
	envMap["RAILPACK_BUILT_AT"] = strconv.FormatInt(time.Now().Unix(), 10)

	envVars := make([]string, 0, len(envMap))
	for _, k := range slices.Sorted(maps.Keys(envMap)) {
		v := envMap[k]
		envVars = append(envVars, fmt.Sprintf("%s=%s", k, v))
	}

	return envVars
}
