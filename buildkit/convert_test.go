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
