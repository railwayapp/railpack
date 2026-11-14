package java

import (
	"encoding/json"
	"testing"

	"github.com/railwayapp/railpack/core/app"
	"github.com/railwayapp/railpack/core/config"
	testingUtils "github.com/railwayapp/railpack/core/testing"
	"github.com/stretchr/testify/require"
)

func TestJavaConfigVersionPrecedence(t *testing.T) {
	tests := []struct {
		name       string
		env        map[string]string
		configJSON string
		wantJava   string
		wantGradle string
	}{
		{
			name: "provider config overrides environment and project versions",
			env: map[string]string{
				"RAILPACK_JAVA_VERSION":        "11",
				"RAILPACK_JAVA_GRADLE_VERSION": "6",
			},
			configJSON: `{"java":{"version":"21","gradleVersion":"8"}}`,
			wantJava:   "21",
			wantGradle: "8",
		},
		{
			name: "generated environment variables are applied",
			env: map[string]string{
				"RAILPACK_JAVA_VERSION":        "17",
				"RAILPACK_JAVA_GRADLE_VERSION": "7",
			},
			wantJava:   "17",
			wantGradle: "7",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := testingUtils.CreateGenerateContext(t, "../../../examples/java-gradle")
			ctx.Env = app.NewEnvironment(&tt.env)
			if tt.configJSON != "" {
				fileConfig := config.EmptyConfig()
				require.NoError(t, json.Unmarshal([]byte(tt.configJSON), fileConfig))
				ctx.Config = config.Merge(ctx.Config, fileConfig)
			}

			provider := JavaProvider{}
			require.NoError(t, provider.Initialize(ctx))
			provider.setGradleVersion(ctx)
			provider.setJDKVersion(ctx, ctx.GetMiseStepBuilder())

			require.Equal(t, tt.wantGradle, ctx.Resolver.Get("gradle").Version)
			require.Equal(t, tt.wantJava, ctx.Resolver.Get("java").Version)
		})
	}
}

func TestJavaLegacyConfigVariables(t *testing.T) {
	tests := []struct {
		name       string
		env        map[string]string
		wantJava   string
		wantGradle string
		wantLogs   int
	}{
		{
			name: "legacy variables are copied and deprecated",
			env: map[string]string{
				"RAILPACK_JDK_VERSION":    "17",
				"RAILPACK_GRADLE_VERSION": "7",
			},
			wantJava:   "17",
			wantGradle: "7",
			wantLogs:   2,
		},
		{
			name: "new variables take precedence",
			env: map[string]string{
				"RAILPACK_JDK_VERSION":         "17",
				"RAILPACK_JAVA_VERSION":        "21",
				"RAILPACK_GRADLE_VERSION":      "7",
				"RAILPACK_JAVA_GRADLE_VERSION": "8",
			},
			wantJava:   "21",
			wantGradle: "8",
			wantLogs:   2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := testingUtils.CreateGenerateContext(t, "../../../examples/java-gradle")
			ctx.Env = app.NewEnvironment(&tt.env)

			provider := JavaProvider{}
			require.NoError(t, provider.Initialize(ctx))
			provider.setGradleVersion(ctx)
			provider.setJDKVersion(ctx, ctx.GetMiseStepBuilder())

			require.Equal(t, tt.wantGradle, ctx.Resolver.Get("gradle").Version)
			require.Equal(t, tt.wantJava, ctx.Resolver.Get("java").Version)
			require.Len(t, ctx.Logger.Logs, tt.wantLogs)
			for _, log := range ctx.Logger.Logs {
				require.Equal(t, "deprecation", string(log.Level))
				require.Contains(t, log.Msg, "deprecated")
			}
		})
	}
}
