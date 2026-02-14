package java

import (
	"os"
	"path/filepath"
	"testing"

	testingUtils "github.com/railwayapp/railpack/core/testing"
	"github.com/stretchr/testify/require"
)

func TestJavaProviderConfigFromFile(t *testing.T) {
	t.Run("jdk version from provider config", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, "../../../examples/java-maven")
		testingUtils.ClearConfigVariable(ctx, "JDK_VERSION")
		testingUtils.ClearConfigVariable(ctx, "GRADLE_VERSION")
		testingUtils.SetConfigFromJSON(t, ctx, `{
			"java": {
				"version": "17"
			}
		}`)

		provider := JavaProvider{}
		require.NoError(t, provider.Initialize(ctx))
		require.NoError(t, provider.Plan(ctx))

		jdk := ctx.Resolver.Get("java")
		require.Equal(t, "17", jdk.Version)
		require.Equal(t, "java.version", jdk.Source)
	})

	t.Run("gradle version from provider config", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, createGradleApp(t))
		testingUtils.ClearConfigVariable(ctx, "JDK_VERSION")
		testingUtils.ClearConfigVariable(ctx, "GRADLE_VERSION")
		testingUtils.SetConfigFromJSON(t, ctx, `{
			"java": {
				"gradleVersion": "7"
			}
		}`)

		provider := JavaProvider{}
		require.NoError(t, provider.Initialize(ctx))
		require.NoError(t, provider.Plan(ctx))

		gradle := ctx.Resolver.Get("gradle")
		require.Equal(t, "7", gradle.Version)
		require.Equal(t, "java.gradleVersion", gradle.Source)
	})

	t.Run("jdk env var takes precedence over provider config", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, "../../../examples/java-maven")
		testingUtils.ClearConfigVariable(ctx, "JDK_VERSION")
		testingUtils.ClearConfigVariable(ctx, "GRADLE_VERSION")
		ctx.Env.SetVariable("RAILPACK_JDK_VERSION", "11")
		testingUtils.SetConfigFromJSON(t, ctx, `{
			"java": {
				"version": "17"
			}
		}`)

		provider := JavaProvider{}
		require.NoError(t, provider.Initialize(ctx))
		require.NoError(t, provider.Plan(ctx))

		jdk := ctx.Resolver.Get("java")
		require.Equal(t, "11", jdk.Version)
		require.Equal(t, "RAILPACK_JDK_VERSION", jdk.Source)
	})

	t.Run("gradle env var takes precedence over provider config", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, createGradleApp(t))
		testingUtils.ClearConfigVariable(ctx, "JDK_VERSION")
		testingUtils.ClearConfigVariable(ctx, "GRADLE_VERSION")
		ctx.Env.SetVariable("RAILPACK_GRADLE_VERSION", "6")
		testingUtils.SetConfigFromJSON(t, ctx, `{
			"java": {
				"gradleVersion": "8"
			}
		}`)

		provider := JavaProvider{}
		require.NoError(t, provider.Initialize(ctx))
		require.NoError(t, provider.Plan(ctx))

		gradle := ctx.Resolver.Get("gradle")
		require.Equal(t, "6", gradle.Version)
		require.Equal(t, "RAILPACK_GRADLE_VERSION", gradle.Source)
	})
}

func createGradleApp(t *testing.T) string {
	t.Helper()

	appDir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(appDir, "gradlew"), []byte("#!/usr/bin/env sh\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(appDir, "build.gradle"), []byte("plugins {}\n"), 0644))

	return appDir
}
