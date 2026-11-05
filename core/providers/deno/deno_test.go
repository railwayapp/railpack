package deno

import (
	"encoding/json"
	"testing"

	"github.com/railwayapp/railpack/core/app"
	"github.com/railwayapp/railpack/core/config"
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

func TestDenoConfigVersion(t *testing.T) {
	tests := []struct {
		name       string
		fileConfig string
		env        map[string]string
		want       string
		wantSource string
	}{
		{
			name:       "environment variable",
			env:        map[string]string{"RAILPACK_DENO_VERSION": "2.1"},
			want:       "2.1",
			wantSource: "RAILPACK_DENO_VERSION environment variable",
		},
		{
			name:       "config file",
			fileConfig: `{"deno":{"version":"2.2"}}`,
			want:       "2.2",
			wantSource: "railpack config > deno > version",
		},
		{
			name:       "config file overrides environment variable",
			fileConfig: `{"deno":{"version":"2.2"}}`,
			env:        map[string]string{"RAILPACK_DENO_VERSION": "2.1"},
			want:       "2.2",
			wantSource: "railpack config > deno > version",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := testingUtils.CreateGenerateContext(t, "../../../examples/deno-2")
			ctx.Env = app.NewEnvironment(&tt.env)

			fileConfig := config.EmptyConfig()
			if tt.fileConfig != "" {
				require.NoError(t, json.Unmarshal([]byte(tt.fileConfig), fileConfig))
			}
			ctx.Config = config.Merge(ctx.Config, fileConfig)

			provider := DenoProvider{}
			require.NoError(t, provider.Initialize(ctx))
			require.Equal(t, tt.wantSource, provider.versionSource)
			require.NoError(t, provider.Plan(ctx))
			require.Equal(t, tt.want, ctx.Resolver.Get("deno").Version)
		})
	}
}

func TestDenoProviderConfigSchema(t *testing.T) {
	tests := []struct {
		property string
		wantType string
	}{
		{property: "version", wantType: "string"},
	}

	schema := (&DenoProvider{}).ConfigSchema()
	require.Equal(t, "object", schema.Type)
	for _, tt := range tests {
		t.Run(tt.property, func(t *testing.T) {
			propertySchema, ok := schema.Properties.Get(tt.property)
			require.True(t, ok)
			require.Equal(t, tt.wantType, propertySchema.Type)
		})
	}
}
