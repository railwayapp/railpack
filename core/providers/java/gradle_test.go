package java

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/railwayapp/railpack/core/app"
	"github.com/railwayapp/railpack/core/config"
	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/logger"
	"github.com/stretchr/testify/require"
)

func TestSetGradleVersionUsesFullWrapperVersion(t *testing.T) {
	tests := []struct {
		name            string
		distributionURL string
		expectedVersion string
	}{
		{
			name:            "minor version",
			distributionURL: `https\://services.gradle.org/distributions/gradle-8.13-bin.zip`,
			expectedVersion: "8.13",
		},
		{
			name:            "patch version",
			distributionURL: `https\://services.gradle.org/distributions/gradle-9.5.0-bin.zip`,
			expectedVersion: "9.5.0",
		},
		{
			name:            "prerelease version",
			distributionURL: `https\://services.gradle.org/distributions/gradle-9.6.0-rc-1-all.zip`,
			expectedVersion: "9.6.0-rc-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createGradleWrapperContext(t, tt.distributionURL)

			provider := JavaProvider{}
			provider.setGradleVersion(ctx)

			gradle := ctx.Resolver.Get("gradle")
			require.NotNil(t, gradle)
			require.Equal(t, tt.expectedVersion, gradle.Version)
			require.Equal(t, "gradle-wrapper.properties", gradle.Source)
		})
	}
}

func createGradleWrapperContext(t *testing.T, distributionURL string) *generate.GenerateContext {
	t.Helper()

	appDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(appDir, "gradlew"), []byte("#!/bin/sh\n"), 0o755))

	wrapperDir := filepath.Join(appDir, "gradle", "wrapper")
	require.NoError(t, os.MkdirAll(wrapperDir, 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(wrapperDir, "gradle-wrapper.properties"),
		[]byte("distributionUrl="+distributionURL+"\n"),
		0o644,
	))

	userApp, err := app.NewApp(appDir)
	require.NoError(t, err)

	ctx, err := generate.NewGenerateContext(
		userApp,
		app.NewEnvironment(nil),
		config.EmptyConfig(),
		logger.NewLogger(),
	)
	require.NoError(t, err)

	return ctx
}
