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
			require.Equal(t, commandPrefix+DefaultTanstackSrvxStartCommand, provider.getTanstackStartCommand())
		}
	})

	t.Run("oob template uses srvx fallback", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, "../../../examples/tanstack-latest")
		provider := NodeProvider{}

		err := provider.Initialize(ctx)
		require.NoError(t, err)

		require.True(t, provider.isTanstackStart())
		require.True(t, provider.usesTanstackSrvxFallback())
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
		require.False(t, provider.usesTanstackSrvxFallback())
		require.False(t, provider.isSPA(ctx))
		require.Equal(t, "bun run start", provider.GetStartCommand(ctx))

		err = provider.Plan(ctx)
		require.NoError(t, err)
	})
}
