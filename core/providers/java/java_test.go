package java

import (
	"os"
	"path/filepath"
	"testing"

	testingUtils "github.com/railwayapp/railpack/core/testing"
	"github.com/stretchr/testify/require"
)

func TestGradleStartCommand(t *testing.T) {
	t.Run("spring boot gradle project places port option before -jar and targets canonical app.jar", func(t *testing.T) {
		tempDir := t.TempDir()

		err := os.WriteFile(filepath.Join(tempDir, "gradlew"), []byte("#!/bin/sh"), 0o755)
		require.NoError(t, err)

		buildGradle := `
			plugins {
				id 'org.springframework.boot' version '3.2.0'
				id 'java'
			}
		`
		err = os.WriteFile(filepath.Join(tempDir, "build.gradle"), []byte(buildGradle), 0o644)
		require.NoError(t, err)

		ctx := testingUtils.CreateGenerateContext(t, tempDir)
		provider := JavaProvider{}

		startCmd := provider.getStartCmd(ctx)

		require.Contains(t, startCmd, "-Dserver.port=$PORT")
		require.NotContains(t, startCmd, "-jar -Dserver.port")
		require.Contains(t, startCmd, "-jar app.jar")
	})

	t.Run("plain gradle project without spring boot omits port option and targets canonical app.jar", func(t *testing.T) {
		tempDir := t.TempDir()

		err := os.WriteFile(filepath.Join(tempDir, "gradlew"), []byte("#!/bin/sh"), 0o755)
		require.NoError(t, err)

		err = os.WriteFile(filepath.Join(tempDir, "build.gradle"), []byte("plugins { id 'java' }"), 0o644)
		require.NoError(t, err)

		ctx := testingUtils.CreateGenerateContext(t, tempDir)
		provider := JavaProvider{}

		startCmd := provider.getStartCmd(ctx)

		require.NotContains(t, startCmd, "-Dserver.port")
		require.Contains(t, startCmd, "-jar app.jar")
	})
}
