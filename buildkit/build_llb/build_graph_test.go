package build_llb

import (
	"testing"

	"github.com/moby/buildkit/client/llb"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/railwayapp/railpack/core/plan"
	"github.com/stretchr/testify/require"
)

func TestBuildGraphNoCache(t *testing.T) {
	p := &plan.BuildPlan{
		Steps: []plan.Step{
			{
				Name:   "test",
				Caches: []string{"test-cache"},
			},
		},
		Caches: map[string]*plan.Cache{
			"test-cache": {
				Directory: "/root/.cache",
			},
		},
	}

	localState := llb.Local("context")
	cacheStore := NewBuildKitCacheStore("")
	platform := specs.Platform{OS: "linux", Architecture: "amd64"}

	t.Run("with NoCache=false", func(t *testing.T) {
		g, err := NewBuildGraph(p, &localState, cacheStore, "", &platform, "", false)
		require.NoError(t, err)
		require.False(t, g.NoCache)
	})

	t.Run("with NoCache=true", func(t *testing.T) {
		g, err := NewBuildGraph(p, &localState, cacheStore, "", &platform, "", true)
		require.NoError(t, err)
		require.True(t, g.NoCache)
		// NoCache should NOT affect if caches are enabled or not, it just affects the layer cache
		opts, err := g.getCacheMountOptions([]string{"test-cache"})
		require.NoError(t, err)
		require.Len(t, opts, 1)
	})
}

func TestNewBuildGraphRejectsCyclicPlan(t *testing.T) {
	p := &plan.BuildPlan{
		Steps: []plan.Step{
			{Name: "A", Inputs: []plan.Layer{plan.NewStepLayer("B"), plan.NewStepLayer("C")}},
			{Name: "B", Inputs: []plan.Layer{plan.NewStepLayer("A")}},
			{Name: "C", Inputs: []plan.Layer{plan.NewStepLayer("A")}},
		},
	}
	localState := llb.Scratch()
	platform := specs.Platform{OS: "linux", Architecture: "amd64"}

	// Shared cycles can disappear if reduction runs before validation.
	g, err := NewBuildGraph(p, &localState, NewBuildKitCacheStore(""), "", &platform, "", false)
	require.ErrorContains(t, err, "cycle detected")
	require.Nil(t, g)
}

func TestBuildGraphPreservesDescendantEnvironmentOverrides(t *testing.T) {
	p := &plan.BuildPlan{
		Steps: []plan.Step{
			{Name: "A", Variables: map[string]string{"MODE": "base"}},
			{Name: "B", Inputs: []plan.Layer{plan.NewStepLayer("A")}, Variables: map[string]string{"MODE": "derived"}},
			{Name: "C", Inputs: []plan.Layer{plan.NewStepLayer("B"), plan.NewStepLayer("A")}},
		},
		Deploy: plan.Deploy{Base: plan.NewStepLayer("C")},
	}
	localState := llb.Scratch()
	platform := specs.Platform{OS: "linux", Architecture: "amd64"}
	g, err := NewBuildGraph(p, &localState, NewBuildKitCacheStore(""), "", &platform, "", false)
	require.NoError(t, err)

	nodeC, exists := g.graph.GetNode("C")
	require.True(t, exists)
	require.Len(t, nodeC.GetParents(), 1)
	require.Equal(t, "B", nodeC.GetParents()[0].GetName())

	_, err = g.GenerateLLB()
	require.NoError(t, err)
	// The redundant A input must not overwrite B's more specific environment.
	require.Equal(t, "derived", nodeC.(*StepNode).InputEnv.EnvVars["MODE"])
}
