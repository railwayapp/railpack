package buildkit

import (
	"context"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	specs "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/railwayapp/railpack/buildkit/build_llb"
	"github.com/railwayapp/railpack/core"
	"github.com/railwayapp/railpack/core/app"
	"github.com/railwayapp/railpack/core/plan"
	"github.com/stretchr/testify/require"
)

func TestGetImageEnvIncludesBuiltAt(t *testing.T) {
	graphOutput := &build_llb.BuildGraphOutput{
		GraphEnv: build_llb.NewGraphEnvironment(),
	}
	buildPlan := plan.NewBuildPlan()
	buildPlan.Deploy.Variables = map[string]string{
		"RAILPACK_BUILT_AT": "custom",
	}
	before := time.Now().Unix()

	env := getImageEnv(graphOutput, buildPlan)
	after := time.Now().Unix()

	builtAt := ""
	for _, variable := range env {
		if value, ok := strings.CutPrefix(variable, "RAILPACK_BUILT_AT="); ok {
			builtAt = value
		}
	}
	timestamp, err := strconv.ParseInt(builtAt, 10, 64)
	require.NoError(t, err)
	require.GreaterOrEqual(t, timestamp, before)
	require.LessOrEqual(t, timestamp, after)
	require.True(t, slices.IsSorted(env))
}

func TestImagePathHasNoDuplicateEntries(t *testing.T) {
	userApp, err := app.NewApp(filepath.Join("..", "examples", "ruby-with-node"))
	require.NoError(t, err)

	buildResult, err := core.GenerateBuildPlan(userApp, app.NewEnvironment(nil), &core.GenerateBuildPlanOptions{})
	require.NoError(t, err)
	require.True(t, buildResult.Success)

	_, image, err := ConvertPlanToLLB(buildResult.Plan, ConvertPlanOptions{
		BuildPlatform: specs.Platform{OS: "linux", Architecture: "amd64"},
	})
	require.NoError(t, err)

	path := pathFromEnv(t, image.Config.Env)
	require.Empty(t, duplicatePathEntries(path))
}

func pathFromEnv(t *testing.T, env []string) string {
	t.Helper()

	for _, envVar := range env {
		if value, ok := strings.CutPrefix(envVar, "PATH="); ok {
			return value
		}
	}
	t.Fatal("PATH not found in image environment")
	return ""
}

func duplicatePathEntries(path string) []string {
	seen := map[string]bool{}
	duplicates := []string{}
	for entry := range strings.SplitSeq(path, ":") {
		if seen[entry] {
			duplicates = append(duplicates, entry)
		}
		seen[entry] = true
	}
	return duplicates
}

func TestDeployVariablesDoNotAffectLLB(t *testing.T) {
	firstPlan := plan.NewBuildPlan()
	firstPlan.Deploy.Base = plan.NewImageLayer("alpine:latest")
	firstPlan.Deploy.Variables = map[string]string{"VALUE": "one"}
	secondPlan := plan.NewBuildPlan()
	secondPlan.Deploy.Base = plan.NewImageLayer("alpine:latest")
	secondPlan.Deploy.Variables = map[string]string{"VALUE": "two"}
	opts := ConvertPlanOptions{
		BuildPlatform: specs.Platform{
			OS:           "linux",
			Architecture: "amd64",
		},
	}

	firstState, firstImage, err := ConvertPlanToLLB(firstPlan, opts)
	require.NoError(t, err)
	secondState, secondImage, err := ConvertPlanToLLB(secondPlan, opts)
	require.NoError(t, err)

	firstDefinition, err := firstState.Marshal(context.Background())
	require.NoError(t, err)
	secondDefinition, err := secondState.Marshal(context.Background())
	require.NoError(t, err)

	require.Equal(t, firstDefinition.ToPB(), secondDefinition.ToPB())
	require.NotEqual(t, firstImage.Config.Env, secondImage.Config.Env)
}

func TestContainerProcess(t *testing.T) {

	t.Run("default entrypoint uses the start command", func(t *testing.T) {
		image := convertDeployToImage(t, plan.Deploy{StartCmd: "npm start"})
		require.Equal(t, []string{"/bin/bash", "-c"}, image.Config.Entrypoint)
		require.Equal(t, []string{"npm start"}, image.Config.Cmd)
	})

	t.Run("missing start command falls back to bash", func(t *testing.T) {
		image := convertDeployToImage(t, plan.Deploy{})
		require.Equal(t, []string{"/bin/bash", "-c"}, image.Config.Entrypoint)
		require.Equal(t, []string{"/bin/bash"}, image.Config.Cmd)
	})

	t.Run("custom entrypoint keeps the start command as one argument", func(t *testing.T) {
		image := convertDeployToImage(t, plan.Deploy{
			Entrypoint: []string{"./entrypoint.sh"},
			StartCmd:   "echo hello-from-cmd",
		})
		require.Equal(t, []string{"./entrypoint.sh"}, image.Config.Entrypoint)
		require.Equal(t, []string{"echo hello-from-cmd"}, image.Config.Cmd)
	})

	t.Run("custom entrypoint without a start command omits cmd", func(t *testing.T) {
		image := convertDeployToImage(t, plan.Deploy{Entrypoint: []string{"./server", "--port", "8080"}})
		require.Equal(t, []string{"./server", "--port", "8080"}, image.Config.Entrypoint)
		require.Nil(t, image.Config.Cmd)
	})

	t.Run("empty entrypoint keeps the default shell", func(t *testing.T) {
		image := convertDeployToImage(t, plan.Deploy{Entrypoint: []string{}, StartCmd: "npm start"})
		require.Equal(t, []string{"/bin/bash", "-c"}, image.Config.Entrypoint)
		require.Equal(t, []string{"npm start"}, image.Config.Cmd)
	})

}

func TestImageHealthcheck(t *testing.T) {
	t.Run("omitted healthcheck stays unset", func(t *testing.T) {
		image := convertDeployToImage(t, plan.Deploy{})
		require.Nil(t, image.Config.Healthcheck)
	})

	t.Run("command-only healthcheck uses dockerfile defaults", func(t *testing.T) {
		image := convertDeployToImage(t, plan.Deploy{Healthcheck: &plan.Healthcheck{Command: "true"}})
		require.Equal(t, []string{"CMD", "/bin/bash", "-c", "true"}, image.Config.Healthcheck.Test)
		require.Equal(t, 30*time.Second, image.Config.Healthcheck.Interval)
		require.Equal(t, 30*time.Second, image.Config.Healthcheck.Timeout)
		require.Equal(t, time.Duration(0), image.Config.Healthcheck.StartPeriod)
		require.Equal(t, 5*time.Second, image.Config.Healthcheck.StartInterval)
		require.Equal(t, 3, image.Config.Healthcheck.Retries)
	})

	t.Run("healthcheck overrides replace defaults", func(t *testing.T) {
		image := convertDeployToImage(t, plan.Deploy{Healthcheck: &plan.Healthcheck{
			Command:  "true",
			Interval: "30s",
			Timeout:  "5s",
			Retries:  3,
		}})
		require.Equal(t, 30*time.Second, image.Config.Healthcheck.Interval)
		require.Equal(t, 5*time.Second, image.Config.Healthcheck.Timeout)
		require.Equal(t, 5*time.Second, image.Config.Healthcheck.StartInterval)
		require.Equal(t, 3, image.Config.Healthcheck.Retries)
	})
}

func convertDeployToImage(t *testing.T, deploy plan.Deploy) *Image {
	t.Helper()
	buildPlan := plan.NewBuildPlan()
	buildPlan.Deploy = deploy
	buildPlan.Deploy.Base = plan.NewImageLayer("alpine:latest")
	_, image, err := ConvertPlanToLLB(buildPlan, ConvertPlanOptions{
		BuildPlatform: specs.Platform{OS: "linux", Architecture: "amd64"},
	})
	require.NoError(t, err)
	return image
}
