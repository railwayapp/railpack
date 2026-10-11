package node

import (
	"testing"

	testingUtils "github.com/railwayapp/railpack/core/testing"
	"github.com/stretchr/testify/require"
)

func TestTanstackStartSrvxFallback(t *testing.T) {
	t.Run("uses package manager executor", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, "../../../examples/tanstack-latest")
		provider := NodeProvider{}

		err := provider.Initialize(ctx)
		require.NoError(t, err)

		tests := map[PackageManager]string{
			PackageManagerNpm:  "npx ",
			PackageManagerPnpm: "pnpm exec ",
			PackageManagerBun:  "bunx ",
		}
		for packageManager, commandPrefix := range tests {
			provider.packageManager = packageManager
			require.Equal(t, commandPrefix+DefaultTanstackSrvxStartCommand, provider.getTanstackStartCommand(ctx))
		}
	})

	t.Run("oob template uses srvx fallback", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, "../../../examples/tanstack-latest")
		provider := NodeProvider{}

		err := provider.Initialize(ctx)
		require.NoError(t, err)

		require.True(t, provider.isTanstackStart())
		require.True(t, provider.usesTanstackStartFallback())
		require.False(t, provider.usesTanstackNitro(ctx))
		require.False(t, provider.isSPA(ctx))
		require.False(t, provider.isVite(ctx))
		expectedStartCommand := "npx " + DefaultTanstackSrvxStartCommand
		require.Equal(t, expectedStartCommand, provider.GetStartCommand(ctx))

		err = provider.Plan(ctx)
		require.NoError(t, err)
		require.Equal(t, expectedStartCommand, ctx.Deploy.StartCmd)
	})

	t.Run("explicit start skips srvx fallback", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, "../../../examples/node-tanstack-start")
		provider := NodeProvider{}

		err := provider.Initialize(ctx)
		require.NoError(t, err)

		require.True(t, provider.isTanstackStart())
		require.False(t, provider.usesTanstackStartFallback())
		require.False(t, provider.isSPA(ctx))
		require.Equal(t, "bun run start", provider.GetStartCommand(ctx))

		err = provider.Plan(ctx)
		require.NoError(t, err)
	})

	t.Run("nitro without start script uses the nitro server output", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, "../../../examples/tanstack-nitro-nostart")
		provider := NodeProvider{}

		err := provider.Initialize(ctx)
		require.NoError(t, err)

		require.True(t, provider.isTanstackStart())
		require.True(t, provider.usesTanstackStartFallback())
		require.True(t, provider.usesTanstackNitro(ctx))
		require.False(t, provider.isSPA(ctx))
		require.False(t, provider.isVite(ctx))
		require.Equal(t, DefaultTanstackNitroStartCommand, provider.GetStartCommand(ctx))

		err = provider.Plan(ctx)
		require.NoError(t, err)
		require.Equal(t, DefaultTanstackNitroStartCommand, ctx.Deploy.StartCmd)
	})
}

func TestTanstackStartMisplacedDependency(t *testing.T) {
	t.Run("react-start in devDependencies deploys as SPA with a suggestion", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, "../../../examples/tanstack-devdeps-spa")
		provider := NodeProvider{}

		err := provider.Initialize(ctx)
		require.NoError(t, err)

		require.False(t, provider.isTanstackStart())
		require.True(t, provider.hasMisplacedTanstackStart())
		require.True(t, provider.isSPA(ctx))
	})

	t.Run("react-start in dependencies is not misplaced", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, "../../../examples/tanstack-latest")
		provider := NodeProvider{}

		err := provider.Initialize(ctx)
		require.NoError(t, err)

		require.False(t, provider.hasMisplacedTanstackStart())
	})
}
