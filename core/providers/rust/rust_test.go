package rust

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/railwayapp/railpack/core/app"
	"github.com/railwayapp/railpack/core/config"
	testingUtils "github.com/railwayapp/railpack/core/testing"
	"github.com/stretchr/testify/require"
)

func TestRustConfigVersion(t *testing.T) {
	tests := []struct {
		name       string
		fileConfig string
		env        map[string]string
		want       string
	}{
		{
			name: "environment version remains supported",
			env:  map[string]string{"RAILPACK_RUST_VERSION": "1.88"},
			want: "1.88",
		},
		{
			name:       "config version",
			fileConfig: `{"rust":{"version":"1.90"}}`,
			want:       "1.90",
		},
		{
			name:       "config version overrides environment",
			fileConfig: `{"rust":{"version":"1.90"}}`,
			env:        map[string]string{"RAILPACK_RUST_VERSION": "1.88"},
			want:       "1.90",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			require.NoError(t, os.MkdirAll(filepath.Join(dir, "src"), 0o755))
			require.NoError(t, os.WriteFile(filepath.Join(dir, "Cargo.toml"), []byte("[package]\nname = \"test-app\"\nversion = \"0.1.0\"\nedition = \"2021\"\n"), 0o644))
			require.NoError(t, os.WriteFile(filepath.Join(dir, "src", "main.rs"), []byte("fn main() {}\n"), 0o644))

			ctx := testingUtils.CreateGenerateContext(t, dir)
			ctx.Env = app.NewEnvironment(&tt.env)
			fileConfig := config.EmptyConfig()
			if tt.fileConfig != "" {
				require.NoError(t, json.Unmarshal([]byte(tt.fileConfig), fileConfig))
			}
			ctx.Config = config.Merge(ctx.Config, fileConfig)

			provider := RustProvider{}
			require.NoError(t, provider.Initialize(ctx))
			provider.InstallMisePackages(ctx, ctx.GetMiseStepBuilder())
			require.Equal(t, tt.want, ctx.Resolver.Get("rust").Version)
		})
	}
}

func TestRustConfigBinary(t *testing.T) {
	tests := []struct {
		name       string
		fileConfig string
		env        map[string]string
		wantBin    string
	}{
		{
			name:    "environment selector remains supported",
			env:     map[string]string{"RAILPACK_RUST_BIN": "server"},
			wantBin: "server",
		},
		{
			name:       "config selector",
			fileConfig: `{"rust":{"bin":"worker"}}`,
			wantBin:    "worker",
		},
		{
			name:       "config selector overrides environment",
			fileConfig: `{"rust":{"bin":"worker"}}`,
			env:        map[string]string{"RAILPACK_RUST_BIN": "server"},
			wantBin:    "worker",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			require.NoError(t, os.MkdirAll(filepath.Join(dir, "src", "bin"), 0o755))
			require.NoError(t, os.WriteFile(filepath.Join(dir, "Cargo.toml"), []byte("[package]\nname = \"test-app\"\nversion = \"0.1.0\"\nedition = \"2021\"\n"), 0o644))
			require.NoError(t, os.WriteFile(filepath.Join(dir, "src", "main.rs"), []byte("fn main() {}\n"), 0o644))
			require.NoError(t, os.WriteFile(filepath.Join(dir, "src", "bin", "server.rs"), []byte("fn main() {}\n"), 0o644))
			require.NoError(t, os.WriteFile(filepath.Join(dir, "src", "bin", "worker.rs"), []byte("fn main() {}\n"), 0o644))

			ctx := testingUtils.CreateGenerateContext(t, dir)
			ctx.Env = app.NewEnvironment(&tt.env)
			fileConfig := config.EmptyConfig()
			if tt.fileConfig != "" {
				require.NoError(t, json.Unmarshal([]byte(tt.fileConfig), fileConfig))
			}
			ctx.Config = config.Merge(ctx.Config, fileConfig)

			provider := RustProvider{}
			require.NoError(t, provider.Initialize(ctx))
			require.Equal(t, "./bin/"+tt.wantBin, provider.GetStartCommand(ctx))
		})
	}
}

func TestRustConfigCargoWorkspace(t *testing.T) {
	tests := []struct {
		name       string
		fileConfig string
		env        map[string]string
		want       string
	}{
		{
			name: "environment selector remains supported",
			env:  map[string]string{"RAILPACK_CARGO_WORKSPACE": "legacy-app"},
			want: "legacy-app",
		},
		{
			name:       "config selector",
			fileConfig: `{"rust":{"cargoWorkspace":"app"}}`,
			want:       "app",
		},
		{
			name:       "config selector overrides environment",
			fileConfig: `{"rust":{"cargoWorkspace":"app"}}`,
			env:        map[string]string{"RAILPACK_CARGO_WORKSPACE": "legacy-app"},
			want:       "app",
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

			provider := RustProvider{}
			require.NoError(t, provider.Initialize(ctx))
			require.Equal(t, tt.want, provider.resolveCargoWorkspace(ctx))
		})
	}
}

func TestRustProviderConfigSchema(t *testing.T) {
	tests := []struct {
		property string
		wantType string
	}{
		{property: "version", wantType: "string"},
		{property: "bin", wantType: "string"},
		{property: "cargoWorkspace", wantType: "string"},
	}

	schema := (&RustProvider{}).ConfigSchema()
	require.Equal(t, "object", schema.Type)
	for _, tt := range tests {
		t.Run(tt.property, func(t *testing.T) {
			propertySchema, ok := schema.Properties.Get(tt.property)
			require.True(t, ok)
			require.Equal(t, tt.wantType, propertySchema.Type)
		})
	}
}

func TestRust(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		detected    bool
		rustVersion string
	}{
		{
			name:        "rust system deps",
			path:        "../../../examples/rust-system-deps",
			detected:    true,
			rustVersion: "1.85.1",
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
			provider := RustProvider{}
			detected, err := provider.Detect(ctx)
			require.NoError(t, err)
			require.Equal(t, tt.detected, detected)

			if detected {
				err = provider.Initialize(ctx)
				require.NoError(t, err)

				err = provider.Plan(ctx)
				require.NoError(t, err)

				if tt.rustVersion != "" {
					rustVersion := ctx.Resolver.Get("rust")
					require.Equal(t, tt.rustVersion, rustVersion.Version)
				}
			}
		})
	}
}
