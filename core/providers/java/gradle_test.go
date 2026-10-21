package java

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	testingUtils "github.com/railwayapp/railpack/core/testing"
)

func TestSetGradleVersion(t *testing.T) {
	tests := []struct {
		name            string
		distributionURL string
		wantVersion     string
		wantJDK         string
	}{
		{
			name:            "patch version",
			distributionURL: `https\://services.gradle.org/distributions/gradle-9.5.0-bin.zip`,
			wantVersion:     "9.5.0",
			wantJDK:         DEFAULT_JDK_VERSION,
		},
		{
			name:            "major and minor",
			distributionURL: `https\://services.gradle.org/distributions/gradle-8.13-bin.zip`,
			wantVersion:     "8.13",
			wantJDK:         DEFAULT_JDK_VERSION,
		},
		{
			name:            "all distribution",
			distributionURL: `https\://services.gradle.org/distributions/gradle-8.13-all.zip`,
			wantVersion:     "8.13",
			wantJDK:         DEFAULT_JDK_VERSION,
		},
		{
			name:            "milestone strips suffix",
			distributionURL: `https\://services.gradle.org/distributions/gradle-9.0-milestone-1-bin.zip`,
			wantVersion:     "9.0",
			wantJDK:         DEFAULT_JDK_VERSION,
		},
		{
			name:            "gradle 5 selects jdk 8",
			distributionURL: `https\://services.gradle.org/distributions/gradle-5.6.4-bin.zip`,
			wantVersion:     "5.6.4",
			wantJDK:         "8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := gradleApp(t, "distributionUrl="+tt.distributionURL+"\n")

			ctx := testingUtils.CreateGenerateContext(t, dir)
			provider := JavaProvider{}
			provider.setGradleVersion(ctx)
			provider.setJDKVersion(ctx, ctx.GetMiseStepBuilder())

			gradlePkg := ctx.GetMiseStepBuilder().Resolver.Get("gradle")
			require.NotNil(t, gradlePkg)
			require.Equal(t, tt.wantVersion, gradlePkg.Version)
			require.Equal(t, "gradle-wrapper.properties", gradlePkg.Source)

			jdkPkg := ctx.GetMiseStepBuilder().Resolver.Get("java")
			require.NotNil(t, jdkPkg)
			require.Equal(t, tt.wantJDK, jdkPkg.Version)
		})
	}

	t.Run("no wrapper properties keeps default", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "gradlew"), []byte("#!/bin/sh\n"), 0o644))

		ctx := testingUtils.CreateGenerateContext(t, dir)
		provider := JavaProvider{}
		provider.setGradleVersion(ctx)

		pkg := ctx.GetMiseStepBuilder().Resolver.Get("gradle")
		require.NotNil(t, pkg)
		require.Equal(t, DEFAULT_GRADLE_VERSION, pkg.Version)
	})

	t.Run("java-gradle example", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, "../../../examples/java-gradle")

		provider := JavaProvider{}
		provider.setGradleVersion(ctx)

		pkg := ctx.GetMiseStepBuilder().Resolver.Get("gradle")
		require.NotNil(t, pkg)
		require.Equal(t, "8.13", pkg.Version)
		require.Equal(t, "gradle-wrapper.properties", pkg.Source)
	})
}

func gradleApp(t *testing.T, wrapperProps string) string {
	t.Helper()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "gradlew"), []byte("#!/bin/sh\n"), 0o644))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "gradle", "wrapper"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "gradle", "wrapper", "gradle-wrapper.properties"), []byte(wrapperProps), 0o644))

	return dir
}
