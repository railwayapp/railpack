package java

import (
	"encoding/json"
	"testing"

	"github.com/railwayapp/railpack/core/app"
	"github.com/railwayapp/railpack/core/config"
	testingUtils "github.com/railwayapp/railpack/core/testing"
	"github.com/stretchr/testify/require"
)

func TestJavaConfigOverridesEnvironmentAndProjectVersions(t *testing.T) {
	ctx := testingUtils.CreateGenerateContext(t, "../../../examples/java-gradle")
	ctx.Env = app.NewEnvironment(&map[string]string{
		"RAILPACK_JDK_VERSION":    "11",
		"RAILPACK_GRADLE_VERSION": "6",
	})
	fileConfig := config.EmptyConfig()
	require.NoError(t, json.Unmarshal([]byte(`{"java":{"version":"21","gradleVersion":"8"}}`), fileConfig))
	ctx.Config = config.Merge(ctx.Config, fileConfig)

	provider := JavaProvider{}
	require.NoError(t, provider.Initialize(ctx))
	provider.setGradleVersion(ctx)
	provider.setJDKVersion(ctx, ctx.GetMiseStepBuilder())

	require.Equal(t, "8", ctx.Resolver.Get("gradle").Version)
	require.Equal(t, "21", ctx.Resolver.Get("java").Version)
}
