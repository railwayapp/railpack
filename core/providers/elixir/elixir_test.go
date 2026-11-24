package elixir

import (
	"testing"

	"github.com/railwayapp/railpack/core/resolver"
	testingUtils "github.com/railwayapp/railpack/core/testing"
	"github.com/stretchr/testify/require"
)

func TestElixirUsesDefaultVersionWhenUnspecified(t *testing.T) {
	ctx := testingUtils.CreateGenerateContext(t, t.TempDir())
	provider := ElixirProvider{}

	provider.InstallMisePackages(ctx, ctx.GetMiseStepBuilder())

	elixir := ctx.Resolver.Get("elixir")
	require.NotNil(t, elixir)
	require.Equal(t, DEFAULT_ELIXIR_VERSION, elixir.Version)
	require.Equal(t, resolver.DefaultSource, elixir.Source)
}

func TestElixir(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		detected     bool
		expectedMain string
	}{
		{
			name:     "deno project with main.ts",
			path:     "../../../examples/elixir-phoenix",
			detected: true,
		},
		{
			name:     "non-deno project",
			path:     "../../../examples/node-npm",
			detected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := testingUtils.CreateGenerateContext(t, tt.path)
			provider := ElixirProvider{}

			detected, err := provider.Detect(ctx)
			require.NoError(t, err)
			require.Equal(t, tt.detected, detected)

			if detected {
				err = provider.Initialize(ctx)
				require.NoError(t, err)

				err = provider.Plan(ctx)
				require.NoError(t, err)
			}
		})
	}
}
