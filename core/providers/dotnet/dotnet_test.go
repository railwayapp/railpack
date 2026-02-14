package dotnet

import (
	"encoding/json"
	"testing"

	"github.com/railwayapp/railpack/core/config"
	"github.com/railwayapp/railpack/core/generate"
	testingUtils "github.com/railwayapp/railpack/core/testing"
	"github.com/stretchr/testify/require"
)

func TestDotnet(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		detected bool
	}{
		{
			name:     "dotnet project",
			path:     "../../../examples/dotnet-cli",
			detected: true,
		},
		{
			name:     "non-dotnet project",
			path:     "../../../examples/node-npm",
			detected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := testingUtils.CreateGenerateContext(t, tt.path)
			provider := DotnetProvider{}

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

func TestDotnetProviderConfigFromFile(t *testing.T) {
	t.Run("dotnet version from provider config", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, "../../../examples/dotnet-cli")
		clearConfigVariable(ctx, "DOTNET_VERSION")
		setConfigFromJSON(t, ctx, `{
			"dotnet": {
				"version": "9.0"
			}
		}`)

		provider := DotnetProvider{}
		require.NoError(t, provider.Initialize(ctx))
		require.NoError(t, provider.Plan(ctx))

		dotnetVersion := ctx.Resolver.Get("dotnet")
		require.Equal(t, "9.0", dotnetVersion.Version)
		require.Equal(t, "dotnet.version", dotnetVersion.Source)
	})

	t.Run("dotnet env var takes precedence over provider config", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, "../../../examples/dotnet-cli")
		clearConfigVariable(ctx, "DOTNET_VERSION")
		ctx.Env.SetVariable("RAILPACK_DOTNET_VERSION", "8.0")
		setConfigFromJSON(t, ctx, `{
			"dotnet": {
				"version": "9.0"
			}
		}`)

		provider := DotnetProvider{}
		require.NoError(t, provider.Initialize(ctx))
		require.NoError(t, provider.Plan(ctx))

		dotnetVersion := ctx.Resolver.Get("dotnet")
		require.Equal(t, "8.0", dotnetVersion.Version)
		require.Equal(t, "RAILPACK_DOTNET_VERSION", dotnetVersion.Source)
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
