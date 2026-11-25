package node

import (
	"os"
	"path/filepath"
	"testing"

	testingUtils "github.com/railwayapp/railpack/core/testing"
	"github.com/stretchr/testify/require"
)

func writeViteApp(t *testing.T, packageJson string, viteConfig string) string {
	t.Helper()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "package.json"), []byte(packageJson), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "vite.config.ts"), []byte(viteConfig), 0o644))
	return dir
}

func TestViteNitroDeploysAsServer(t *testing.T) {
	t.Run("fresh 3 registers nitro through its own vite plugin", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, "../../../examples/node-vite-fresh")
		provider := NodeProvider{}
		require.NoError(t, provider.Initialize(ctx))

		require.True(t, provider.usesViteNitro(ctx))
		require.False(t, provider.isVite(ctx))
		require.False(t, provider.isSPA(ctx))
		require.Equal(t, DefaultNitroStartCommand, provider.GetStartCommand(ctx))

		require.NoError(t, provider.Plan(ctx))
		require.Equal(t, DefaultNitroStartCommand, ctx.Deploy.StartCmd)
	})

	t.Run("nitro vite plugin without a framework", func(t *testing.T) {
		dir := writeViteApp(t,
			`{"scripts": {"build": "vite build"}, "dependencies": {"nitro": "latest", "vite": "^8"}}`,
			`import { defineConfig } from "vite"; import { nitro } from "nitro/vite"; export default defineConfig({ plugins: [nitro()] });`,
		)
		ctx := testingUtils.CreateGenerateContext(t, dir)
		provider := NodeProvider{}
		require.NoError(t, provider.Initialize(ctx))

		require.True(t, provider.usesViteNitro(ctx))
		require.False(t, provider.isSPA(ctx))
		require.Equal(t, DefaultNitroStartCommand, provider.GetStartCommand(ctx))
	})

	t.Run("start script takes precedence over the nitro output", func(t *testing.T) {
		dir := writeViteApp(t,
			`{"scripts": {"build": "vite build", "start": "node server.mjs"}, "dependencies": {"nitro": "latest", "vite": "^8"}}`,
			`import { nitro } from "nitro/vite"; export default { plugins: [nitro()] };`,
		)
		ctx := testingUtils.CreateGenerateContext(t, dir)
		provider := NodeProvider{}
		require.NoError(t, provider.Initialize(ctx))

		require.Equal(t, "npm run start", provider.GetStartCommand(ctx))
	})

	t.Run("nitro dependency unused by the vite config stays a static site", func(t *testing.T) {
		dir := writeViteApp(t,
			`{"scripts": {"build": "vite build"}, "dependencies": {"nitro": "latest", "vite": "^8"}}`,
			`export default {};`,
		)
		ctx := testingUtils.CreateGenerateContext(t, dir)
		provider := NodeProvider{}
		require.NoError(t, provider.Initialize(ctx))

		require.False(t, provider.usesViteNitro(ctx))
		require.True(t, provider.isSPA(ctx))
	})
}
