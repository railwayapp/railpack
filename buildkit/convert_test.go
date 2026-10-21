package buildkit

import (
	"context"
	"encoding/json"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/solver/pb"
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

func TestGetSourceStateLocal(t *testing.T) {
	buildPlan := &plan.BuildPlan{Exclude: []string{"node_modules", "!keep"}}

	ops := marshalOps(t, getSourceState(buildPlan, ConvertPlanOptions{SessionID: "session-id"}))

	sources := sourceOps(ops)
	require.Len(t, sources, 1)
	require.Equal(t, "local://context", sources[0].Identifier)
	var excludes []string
	require.NoError(t, json.Unmarshal([]byte(sources[0].Attrs[pb.AttrExcludePatterns]), &excludes))
	require.Equal(t, buildPlan.Exclude, excludes)
}

func TestGetSourceStateGit(t *testing.T) {
	gitState := llb.Git("https://github.com/org/repo.git", "main")

	tests := []struct {
		name         string
		exclude      []string
		wantExcludes [][]string
	}{
		{name: "no_excludes", exclude: nil, wantExcludes: nil},
		{name: "excludes", exclude: []string{"node_modules", "!keep"}, wantExcludes: [][]string{{"node_modules", "!keep"}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buildPlan := &plan.BuildPlan{Exclude: tt.exclude}

			ops := marshalOps(t, getSourceState(buildPlan, ConvertPlanOptions{ContextState: &gitState}))

			sources := sourceOps(ops)
			require.Len(t, sources, 1)
			require.True(t, strings.HasPrefix(sources[0].Identifier, "git://"), "source = %q, want a git source", sources[0].Identifier)
			var gotExcludes [][]string
			for _, cp := range copyActions(ops) {
				gotExcludes = append(gotExcludes, cp.ExcludePatterns)
			}
			require.Equal(t, tt.wantExcludes, gotExcludes, "copy exclude patterns for plan exclude %q", tt.exclude)
		})
	}
}
func marshalOps(t *testing.T, st llb.State) []*pb.Op {
	t.Helper()

	def, err := st.Marshal(context.Background())
	require.NoError(t, err)

	ops := make([]*pb.Op, 0, len(def.Def))
	for _, dt := range def.Def {
		var op pb.Op
		require.NoError(t, op.UnmarshalVT(dt))
		ops = append(ops, &op)
	}
	return ops
}

func sourceOps(ops []*pb.Op) []*pb.SourceOp {
	var sources []*pb.SourceOp
	for _, op := range ops {
		if src := op.GetSource(); src != nil {
			sources = append(sources, src)
		}
	}
	return sources
}

func copyActions(ops []*pb.Op) []*pb.FileActionCopy {
	var copies []*pb.FileActionCopy
	for _, op := range ops {
		for _, action := range op.GetFile().GetActions() {
			if cp := action.GetCopy(); cp != nil {
				copies = append(copies, cp)
			}
		}
	}
	return copies
}
