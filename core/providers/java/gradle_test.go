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
		want            string
		wantSource      string
	}{
		{
			name:            "exact version",
			distributionURL: `https\://services.gradle.org/distributions/gradle-9.5.0-bin.zip`,
			want:            "9.5.0",
			wantSource:      "gradle-wrapper.properties",
		},
		{
			name:            "major and minor",
			distributionURL: `https\://services.gradle.org/distributions/gradle-8.13-bin.zip`,
			want:            "8.13",
			wantSource:      "gradle-wrapper.properties",
		},
		{
			name:            "all distribution",
			distributionURL: `https\://services.gradle.org/distributions/gradle-8.13-all.zip`,
			want:            "8.13",
			wantSource:      "gradle-wrapper.properties",
		},
		{
			name: "milestone falls back to the release it precedes",
			// The version regex stops at the first non-numeric segment, so a
			// milestone URL yields the plain version rather than the preview tag.
			distributionURL: `https\://services.gradle.org/distributions/gradle-9.0-milestone-1-bin.zip`,
			want:            "9.0",
			wantSource:      "gradle-wrapper.properties",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := gradleApp(t, "distributionUrl="+tt.distributionURL+"\n")

			ctx := testingUtils.CreateGenerateContext(t, dir)
			provider := JavaProvider{}
			provider.setGradleVersion(ctx)

			pkg := ctx.GetMiseStepBuilder().Resolver.Get("gradle")
			require.NotNil(t, pkg)
			require.Equal(t, tt.want, pkg.Version)
			require.Equal(t, tt.wantSource, pkg.Source)
		})
	}

	t.Run("no wrapper properties keeps the default", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "gradlew"), []byte("#!/bin/sh\n"), 0o644))

		ctx := testingUtils.CreateGenerateContext(t, dir)
		provider := JavaProvider{}
		provider.setGradleVersion(ctx)

		pkg := ctx.GetMiseStepBuilder().Resolver.Get("gradle")
		require.NotNil(t, pkg)
		require.Equal(t, DEFAULT_GRADLE_VERSION, pkg.Version)
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
