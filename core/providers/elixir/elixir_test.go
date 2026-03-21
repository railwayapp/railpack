package elixir

import (
	"testing"

	testingUtils "github.com/railwayapp/railpack/core/testing"
	"github.com/stretchr/testify/require"
)

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

func TestElixirProviderConfigFromFile(t *testing.T) {
	t.Run("elixir version from provider config", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, "../../../examples/elixir-ecto")
		testingUtils.ClearConfigVariable(ctx, "ELIXIR_VERSION")
		testingUtils.ClearConfigVariable(ctx, "ERLANG_VERSION")
		testingUtils.SetConfigFromJSON(t, ctx, `{
			"elixir": {
				"version": "1.18.4"
			}
		}`)

		provider := ElixirProvider{}
		require.NoError(t, provider.Initialize(ctx))
		require.NoError(t, provider.Plan(ctx))

		elixirVersion := ctx.Resolver.Get("elixir")
		require.Equal(t, "1.18.4", elixirVersion.Version)
		require.Equal(t, "elixir.version", elixirVersion.Source)
	})

	t.Run("erlang version from provider config", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, "../../../examples/elixir-ecto")
		testingUtils.ClearConfigVariable(ctx, "ELIXIR_VERSION")
		testingUtils.ClearConfigVariable(ctx, "ERLANG_VERSION")
		testingUtils.SetConfigFromJSON(t, ctx, `{
			"elixir": {
				"erlangVersion": "27.3.4"
			}
		}`)

		provider := ElixirProvider{}
		require.NoError(t, provider.Initialize(ctx))
		require.NoError(t, provider.Plan(ctx))

		erlangVersion := ctx.Resolver.Get("erlang")
		require.Equal(t, "27.3.4", erlangVersion.Version)
		require.Equal(t, "elixir.erlangVersion", erlangVersion.Source)
	})

	t.Run("elixir env var takes precedence over provider config", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, "../../../examples/elixir-ecto")
		testingUtils.ClearConfigVariable(ctx, "ELIXIR_VERSION")
		testingUtils.ClearConfigVariable(ctx, "ERLANG_VERSION")
		ctx.Env.SetVariable("RAILPACK_ELIXIR_VERSION", "1.16.3")
		testingUtils.SetConfigFromJSON(t, ctx, `{
			"elixir": {
				"version": "1.18.4"
			}
		}`)

		provider := ElixirProvider{}
		require.NoError(t, provider.Initialize(ctx))
		require.NoError(t, provider.Plan(ctx))

		elixirVersion := ctx.Resolver.Get("elixir")
		require.Equal(t, "1.16.3", elixirVersion.Version)
		require.Equal(t, "RAILPACK_ELIXIR_VERSION", elixirVersion.Source)
	})

	t.Run("erlang env var takes precedence over provider config", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, "../../../examples/elixir-ecto")
		testingUtils.ClearConfigVariable(ctx, "ELIXIR_VERSION")
		testingUtils.ClearConfigVariable(ctx, "ERLANG_VERSION")
		ctx.Env.SetVariable("RAILPACK_ERLANG_VERSION", "26.2.5")
		testingUtils.SetConfigFromJSON(t, ctx, `{
			"elixir": {
				"erlangVersion": "27.3.4"
			}
		}`)

		provider := ElixirProvider{}
		require.NoError(t, provider.Initialize(ctx))
		require.NoError(t, provider.Plan(ctx))

		erlangVersion := ctx.Resolver.Get("erlang")
		require.Equal(t, "26.2.5", erlangVersion.Version)
		require.Equal(t, "RAILPACK_ERLANG_VERSION", erlangVersion.Source)
	})
}
