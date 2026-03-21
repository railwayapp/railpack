package gleam

import (
	"testing"

	testingUtils "github.com/railwayapp/railpack/core/testing"
	"github.com/stretchr/testify/require"
)

func TestGleamProviderConfigFromFile(t *testing.T) {
	t.Run("include source from provider config", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, "../../../examples/gleam")
		testingUtils.ClearConfigVariable(ctx, "GLEAM_INCLUDE_SOURCE")
		testingUtils.SetConfigFromJSON(t, ctx, `{
			"gleam": {
				"includeSource": true
			}
		}`)

		provider := GleamProvider{}
		require.NoError(t, provider.Initialize(ctx))
		require.NoError(t, provider.Plan(ctx))

		require.True(t, ctx.Deploy.HasIncludeForStep("build", "."))
	})

	t.Run("env var takes precedence over provider config", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, "../../../examples/gleam")
		testingUtils.ClearConfigVariable(ctx, "GLEAM_INCLUDE_SOURCE")
		ctx.Env.SetVariable("RAILPACK_GLEAM_INCLUDE_SOURCE", "false")
		testingUtils.SetConfigFromJSON(t, ctx, `{
			"gleam": {
				"includeSource": true
			}
		}`)

		provider := GleamProvider{}
		require.NoError(t, provider.Initialize(ctx))
		require.NoError(t, provider.Plan(ctx))

		require.False(t, ctx.Deploy.HasIncludeForStep("build", "."))
		require.True(t, ctx.Deploy.HasIncludeForStep("build", "build/erlang-shipment/."))
	})
}
