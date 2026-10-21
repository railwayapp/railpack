package node

import (
	"testing"

	"github.com/railwayapp/railpack/core/generate"
	testingUtils "github.com/railwayapp/railpack/core/testing"
	"github.com/stretchr/testify/require"
)

func TestSvelteKitAdapterAuto(t *testing.T) {
	ctx := testingUtils.CreateGenerateContext(t, "../../../examples/node-svelte-kit")
	provider := NodeProvider{}
	require.NoError(t, provider.Initialize(ctx))
	require.NoError(t, provider.Plan(ctx))

	require.Equal(t, DefaultSvelteKitStartCommand, ctx.Deploy.StartCmd)

	step := ctx.GetStepByName("build")
	require.NotNil(t, step)
	build, ok := (*step).(*generate.CommandStepBuilder)
	require.True(t, ok)
	require.Equal(t, "true", build.Variables["GCP_BUILDPACKS"])
}
