package java

import (
	"testing"

	"github.com/stretchr/testify/require"

	testingUtils "github.com/railwayapp/railpack/core/testing"
)

func TestSetGradleVersionUsesFullVersion(t *testing.T) {
	ctx := testingUtils.CreateGenerateContext(t, "../../../examples/java-gradle")

	provider := JavaProvider{}
	provider.setGradleVersion(ctx)

	pkg := ctx.GetMiseStepBuilder().Resolver.Get("gradle")
	require.NotNil(t, pkg)
	require.Equal(t, "8.13", pkg.Version)
	require.Equal(t, "gradle-wrapper.properties", pkg.Source)
}
