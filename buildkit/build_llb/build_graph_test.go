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
		require.False(t, isCacheDisabled("test-cache"))
	})

	t.Run("with NoCache=true", func(t *testing.T) {
		g, err := NewBuildGraph(p, &localState, cacheStore, "", &platform, "", true)
		require.NoError(t, err)
		require.True(t, g.NoCache)
		// NoCache should NOT affect if caches are enabled or not, it just affects the layer cache
		require.False(t, isCacheDisabled("test-cache"))

		opts, err := g.getCacheMountOptions([]string{"test-cache"})
		require.NoError(t, err)
		require.Len(t, opts, 1)
	})
}
