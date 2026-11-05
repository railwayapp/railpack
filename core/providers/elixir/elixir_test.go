package elixir

import (
	"encoding/json"
	"testing"

	"github.com/railwayapp/railpack/core/app"
	"github.com/railwayapp/railpack/core/config"
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

func TestElixirConfigVersions(t *testing.T) {
	tests := []struct {
		name       string
		fileConfig string
		env        map[string]string
		wantElixir string
		wantErlang string
	}{
		{
			name: "environment variables remain supported",
			env: map[string]string{
				"RAILPACK_ELIXIR_VERSION": "1.17",
				"RAILPACK_ERLANG_VERSION": "26.2",
			},
			wantElixir: "1.17",
			wantErlang: "26.2",
		},
		{
			name:       "config file versions",
			fileConfig: `{"elixir":{"version":"1.16","erlangVersion":"25.3"}}`,
			wantElixir: "1.16",
			wantErlang: "25.3",
		},
		{
			name:       "config file overrides environment variables",
			fileConfig: `{"elixir":{"version":"1.16","erlangVersion":"25.3"}}`,
			env: map[string]string{
				"RAILPACK_ELIXIR_VERSION": "1.17",
				"RAILPACK_ERLANG_VERSION": "26.2",
			},
			wantElixir: "1.16",
			wantErlang: "25.3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := testingUtils.CreateGenerateContext(t, t.TempDir())
			ctx.Env = app.NewEnvironment(&tt.env)

			fileConfig := config.EmptyConfig()
			if tt.fileConfig != "" {
				require.NoError(t, json.Unmarshal([]byte(tt.fileConfig), fileConfig))
			}
			ctx.Config = config.Merge(ctx.Config, fileConfig)

			provider := ElixirProvider{}
			require.NoError(t, provider.Initialize(ctx))
			provider.InstallMisePackages(ctx, ctx.GetMiseStepBuilder())

			require.Equal(t, tt.wantElixir, ctx.Resolver.Get("elixir").Version)
			require.Equal(t, tt.wantErlang, ctx.Resolver.Get("erlang").Version)
		})
	}
}

func TestElixirProviderConfigSchema(t *testing.T) {
	tests := []struct {
		property string
		wantType string
	}{
		{property: "version", wantType: "string"},
		{property: "erlangVersion", wantType: "string"},
	}

	schema := (&ElixirProvider{}).ConfigSchema()
	require.Equal(t, "object", schema.Type)
	for _, tt := range tests {
		t.Run(tt.property, func(t *testing.T) {
			propertySchema, ok := schema.Properties.Get(tt.property)
			require.True(t, ok)
			require.Equal(t, tt.wantType, propertySchema.Type)
		})
	}
}
