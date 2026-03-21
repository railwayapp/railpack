package deno

import (
	"testing"

	testingUtils "github.com/railwayapp/railpack/core/testing"
	"github.com/stretchr/testify/require"
)

func TestDeno(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		detected     bool
		expectedMain string
	}{
		{
			name:         "deno project with main.ts",
			path:         "../../../examples/deno-2",
			detected:     true,
			expectedMain: "main.ts",
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
			provider := DenoProvider{}

			detected, err := provider.Detect(ctx)
			require.NoError(t, err)
			require.Equal(t, tt.detected, detected)

			if detected {
				err = provider.Initialize(ctx)
				require.NoError(t, err)

				require.Equal(t, tt.expectedMain, provider.mainFile)

				err = provider.Plan(ctx)
				require.NoError(t, err)

				// Verify start command format
				if provider.mainFile != "" {
					expectedCmd := "deno run --allow-all " + provider.mainFile
					require.Equal(t, expectedCmd, ctx.Deploy.StartCmd)
				} else {
					require.Empty(t, ctx.Deploy.StartCmd)
				}
			}
		})
	}
}

func TestDenoProviderConfigFromFile(t *testing.T) {
	t.Run("deno version from provider config", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, "../../../examples/deno-2")
		testingUtils.ClearConfigVariable(ctx, "DENO_VERSION")
		testingUtils.SetConfigFromJSON(t, ctx, `{
			"deno": {
				"version": "2.1.0"
			}
		}`)

		provider := DenoProvider{}
		require.NoError(t, provider.Initialize(ctx))
		require.NoError(t, provider.Plan(ctx))

		denoVersion := ctx.Resolver.Get("deno")
		require.Equal(t, "2.1.0", denoVersion.Version)
		require.Equal(t, "deno.version", denoVersion.Source)
	})

	t.Run("deno env var takes precedence over provider config", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, "../../../examples/deno-2")
		testingUtils.ClearConfigVariable(ctx, "DENO_VERSION")
		ctx.Env.SetVariable("RAILPACK_DENO_VERSION", "2.4.0")
		testingUtils.SetConfigFromJSON(t, ctx, `{
			"deno": {
				"version": "2.1.0"
			}
		}`)

		provider := DenoProvider{}
		require.NoError(t, provider.Initialize(ctx))
		require.NoError(t, provider.Plan(ctx))

		denoVersion := ctx.Resolver.Get("deno")
		require.Equal(t, "2.4.0", denoVersion.Version)
		require.Equal(t, "RAILPACK_DENO_VERSION", denoVersion.Source)
	})
}
