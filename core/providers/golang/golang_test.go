package golang

import (
	"encoding/json"
	"testing"

	"github.com/railwayapp/railpack/core/app"
	"github.com/railwayapp/railpack/core/config"
	testingUtils "github.com/railwayapp/railpack/core/testing"
	"github.com/stretchr/testify/require"
)

func TestGolang(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		detected     bool
		hasGoMod     bool
		hasWorkspace bool
		goVersion    string
		cgoEnabled   bool
	}{
		{
			name:      "go mod",
			path:      "../../../examples/go-mod",
			detected:  true,
			hasGoMod:  true,
			goVersion: "1.25.3",
		},
		{
			name:      "go cmd dirs",
			path:      "../../../examples/go-cmd-dirs",
			detected:  true,
			hasGoMod:  true,
			goVersion: "1.25.3",
		},
		{
			name:         "go workspaces",
			path:         "../../../examples/go-workspaces",
			detected:     true,
			hasGoMod:     false,
			hasWorkspace: true,
			goVersion:    "1.25",
		},
		{
			name:     "node",
			path:     "../../../examples/node-npm",
			detected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := testingUtils.CreateGenerateContext(t, tt.path)
			provider := GoProvider{}
			detected, err := provider.Detect(ctx)
			require.NoError(t, err)
			require.Equal(t, tt.detected, detected)

			if detected {
				err = provider.Initialize(ctx)
				require.NoError(t, err)

				err = provider.Plan(ctx)
				require.NoError(t, err)

				require.Equal(t, tt.hasGoMod, provider.isGoMod(ctx))
				require.Equal(t, tt.hasWorkspace, provider.isGoWorkspace(ctx))
				require.Equal(t, tt.cgoEnabled, provider.hasCGOEnabled(ctx))

				if tt.goVersion != "" {
					goVersion := ctx.Resolver.Get("go")
					require.Equal(t, tt.goVersion, goVersion.Version)
				}

				if tt.hasWorkspace {
					packages := provider.GoWorkspacePackages(ctx)
					require.Greater(t, len(packages), 0, "workspace should have at least one package")
				}
			}
		})
	}
}

func TestGoProviderConfigSchema(t *testing.T) {
	schema := (&GoProvider{}).ConfigSchema()
	require.Equal(t, "object", schema.Type)

	versionSchema, ok := schema.Properties.Get("version")
	require.True(t, ok)
	require.Equal(t, "string", versionSchema.Type)
	require.Equal(t, "The Go version to install", versionSchema.Description)
}

func TestGoVersionPrecedence(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		env       map[string]string
		wantGoVer string
	}{
		{
			name: "provider config overrides provider environment",
			path: "../../../examples/go-workspaces",
			env: map[string]string{
				"RAILPACK_GOLANG_VERSION": "1.25",
			},
			wantGoVer: "1.26",
		},
		{
			name: "GO_VERSION overrides provider config",
			path: "../../../examples/go-workspaces",
			env: map[string]string{
				"RAILPACK_GOLANG_VERSION": "1.25",
				"RAILPACK_GO_VERSION":     "1.27",
			},
			wantGoVer: "1.27",
		},
		{
			name: "mise config overrides provider config and GO_VERSION",
			path: "../../../examples/go-mod",
			env: map[string]string{
				"RAILPACK_GOLANG_VERSION": "1.25",
				"RAILPACK_GO_VERSION":     "1.27",
			},
			wantGoVer: "1.25.3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := testingUtils.CreateGenerateContext(t, tt.path)
			ctx.Env = app.NewEnvironment(&tt.env)
			fileConfig := config.EmptyConfig()
			require.NoError(t, json.Unmarshal([]byte(`{"golang":{"version":"1.26"}}`), fileConfig))
			ctx.Config = config.Merge(ctx.Config, fileConfig)

			provider := GoProvider{}
			require.NoError(t, provider.Initialize(ctx))
			provider.InstallGoPackages(ctx, ctx.GetMiseStepBuilder())

			require.Equal(t, tt.wantGoVer, ctx.Resolver.Get("go").Version)
		})
	}
}
