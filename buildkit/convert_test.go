package buildkit

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	specs "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/railwayapp/railpack/core"
	"github.com/railwayapp/railpack/core/app"
	"github.com/railwayapp/railpack/core/plan"
	"github.com/stretchr/testify/require"
)

func TestImageEnvironmentDoesNotInjectBuildTime(t *testing.T) {
	buildPlan := plan.NewBuildPlan()
	buildPlan.Deploy.Base = plan.NewImageLayer("alpine:latest")
	buildPlan.Deploy.Variables = map[string]string{"EXAMPLE": "value"}
	_, image, err := ConvertPlanToLLB(buildPlan, ConvertPlanOptions{
		BuildPlatform: specs.Platform{OS: "linux", Architecture: "amd64"},
	})
	require.NoError(t, err)
	require.Contains(t, image.Config.Env, "EXAMPLE=value")
	require.NotContains(t, strings.Join(image.Config.Env, "\n"), "RAILPACK_BUILT_AT=")
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
