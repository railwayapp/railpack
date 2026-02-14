package gleam

import (
	"encoding/json"
	"testing"

	"github.com/railwayapp/railpack/core/config"
	"github.com/railwayapp/railpack/core/generate"
	testingUtils "github.com/railwayapp/railpack/core/testing"
	"github.com/stretchr/testify/require"
)

func TestGleamProviderConfigFromFile(t *testing.T) {
	t.Run("include source from provider config", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, "../../../examples/gleam")
		clearConfigVariable(ctx, "GLEAM_INCLUDE_SOURCE")
		setConfigFromJSON(t, ctx, `{
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
		clearConfigVariable(ctx, "GLEAM_INCLUDE_SOURCE")
		ctx.Env.SetVariable("RAILPACK_GLEAM_INCLUDE_SOURCE", "false")
		setConfigFromJSON(t, ctx, `{
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

func setConfigFromJSON(t *testing.T, ctx *generate.GenerateContext, configJSON string) {
	t.Helper()

	var cfg config.Config
	require.NoError(t, json.Unmarshal([]byte(configJSON), &cfg))
	ctx.Config = &cfg
}

func clearConfigVariable(ctx *generate.GenerateContext, variableName string) {
	delete(ctx.Env.Variables, ctx.Env.ConfigVariable(variableName))
}
