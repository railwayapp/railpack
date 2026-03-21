package rust

import (
	"os"
	"path/filepath"
	"testing"

	testingUtils "github.com/railwayapp/railpack/core/testing"
	"github.com/stretchr/testify/require"
)

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

func TestRustProviderConfigFromFile(t *testing.T) {
	t.Run("rust version from provider config", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, createRustApp(t))
		testingUtils.ClearConfigVariable(ctx, "RUST_VERSION")
		testingUtils.SetConfigFromJSON(t, ctx, `{
			"rust": {
				"version": "1.88.0"
			}
		}`)

		provider := RustProvider{}
		require.NoError(t, provider.Initialize(ctx))
		require.NoError(t, provider.Plan(ctx))

		rustVersion := ctx.Resolver.Get("rust")
		require.Equal(t, "1.88.0", rustVersion.Version)
		require.Equal(t, "rust.version", rustVersion.Source)
	})

	t.Run("rust env var takes precedence over provider config", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, createRustApp(t))
		testingUtils.ClearConfigVariable(ctx, "RUST_VERSION")
		ctx.Env.SetVariable("RAILPACK_RUST_VERSION", "1.90.0")
		testingUtils.SetConfigFromJSON(t, ctx, `{
			"rust": {
				"version": "1.88.0"
			}
		}`)

		provider := RustProvider{}
		require.NoError(t, provider.Initialize(ctx))
		require.NoError(t, provider.Plan(ctx))

		rustVersion := ctx.Resolver.Get("rust")
		require.Equal(t, "1.90.0", rustVersion.Version)
		require.Equal(t, "RAILPACK_RUST_VERSION", rustVersion.Source)
	})

	t.Run("rust bin from provider config", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, createRustMultiBinApp(t))
		testingUtils.ClearConfigVariable(ctx, "RUST_BIN")
		testingUtils.SetConfigFromJSON(t, ctx, `{
			"rust": {
				"bin": "worker"
			}
		}`)

		provider := RustProvider{}
		require.NoError(t, provider.Initialize(ctx))
		require.NoError(t, provider.Plan(ctx))

		require.Equal(t, "./bin/worker", ctx.Deploy.StartCmd)
	})

	t.Run("rust bin env var takes precedence over provider config", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, createRustMultiBinApp(t))
		testingUtils.ClearConfigVariable(ctx, "RUST_BIN")
		ctx.Env.SetVariable("RAILPACK_RUST_BIN", "server")
		testingUtils.SetConfigFromJSON(t, ctx, `{
			"rust": {
				"bin": "worker"
			}
		}`)

		provider := RustProvider{}
		require.NoError(t, provider.Initialize(ctx))
		require.NoError(t, provider.Plan(ctx))

		require.Equal(t, "./bin/server", ctx.Deploy.StartCmd)
	})

	t.Run("rust workspace from provider config", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, createRustWorkspaceApp(t))
		testingUtils.ClearConfigVariable(ctx, "CARGO_WORKSPACE")
		testingUtils.SetConfigFromJSON(t, ctx, `{
			"rust": {
				"workspace": "worker"
			}
		}`)

		provider := RustProvider{}
		require.NoError(t, provider.Initialize(ctx))
		require.NoError(t, provider.Plan(ctx))

		require.Equal(t, "./bin/worker", ctx.Deploy.StartCmd)
	})

	t.Run("cargo workspace env var takes precedence over provider config", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, createRustWorkspaceApp(t))
		testingUtils.ClearConfigVariable(ctx, "CARGO_WORKSPACE")
		ctx.Env.SetVariable("RAILPACK_CARGO_WORKSPACE", "api")
		testingUtils.SetConfigFromJSON(t, ctx, `{
			"rust": {
				"workspace": "worker"
			}
		}`)

		provider := RustProvider{}
		require.NoError(t, provider.Initialize(ctx))
		require.NoError(t, provider.Plan(ctx))

		require.Equal(t, "./bin/api", ctx.Deploy.StartCmd)
	})
}

func createRustApp(t *testing.T) string {
	t.Helper()

	appDir := t.TempDir()

	require.NoError(t, os.MkdirAll(filepath.Join(appDir, "src"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(appDir, "Cargo.toml"), []byte(`[package]
name = "rust-app"
version = "0.1.0"
edition = "2021"
`), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(appDir, "src", "main.rs"), []byte(`fn main() {}`), 0644))

	return appDir
}

func createRustMultiBinApp(t *testing.T) string {
	t.Helper()

	appDir := t.TempDir()

	require.NoError(t, os.MkdirAll(filepath.Join(appDir, "src", "bin"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(appDir, "Cargo.toml"), []byte(`[package]
name = "rust-app"
version = "0.1.0"
edition = "2021"
`), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(appDir, "src", "bin", "server.rs"), []byte(`fn main() {}`), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(appDir, "src", "bin", "worker.rs"), []byte(`fn main() {}`), 0644))

	return appDir
}

func createRustWorkspaceApp(t *testing.T) string {
	t.Helper()

	appDir := t.TempDir()

	require.NoError(t, os.MkdirAll(filepath.Join(appDir, "api", "src"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(appDir, "worker", "src"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(appDir, "Cargo.toml"), []byte(`[workspace]
members = ["api", "worker"]
`), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(appDir, "api", "Cargo.toml"), []byte(`[package]
name = "api"
version = "0.1.0"
edition = "2021"
`), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(appDir, "worker", "Cargo.toml"), []byte(`[package]
name = "worker"
version = "0.1.0"
edition = "2021"
`), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(appDir, "api", "src", "main.rs"), []byte(`fn main() {}`), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(appDir, "worker", "src", "main.rs"), []byte(`fn main() {}`), 0644))

	return appDir
}
