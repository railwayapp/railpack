package java

import (
	"os"
	"path/filepath"
	"testing"

	testingUtils "github.com/railwayapp/railpack/core/testing"
	"github.com/stretchr/testify/require"
)

func TestGradleStartCommand(t *testing.T) {
	t.Run("spring boot gradle project places port option before -jar and supports root build/libs", func(t *testing.T) {
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

		// JVM system property must precede -jar flag
		require.Contains(t, startCmd, "-Dserver.port=$PORT $JAVA_OPTS -jar")
		require.NotContains(t, startCmd, "-jar -Dserver.port", "JVM options must not be placed after -jar")

		// Must search both single-project and multi-project build output directories
		require.Contains(t, startCmd, "build/libs/*jar */build/libs/*jar")
	})

	t.Run("plain gradle project without spring boot omits port option", func(t *testing.T) {
		tempDir := t.TempDir()

		err := os.WriteFile(filepath.Join(tempDir, "gradlew"), []byte("#!/bin/sh"), 0o755)
		require.NoError(t, err)

		err = os.WriteFile(filepath.Join(tempDir, "build.gradle"), []byte("plugins { id 'java' }"), 0o644)
		require.NoError(t, err)

		ctx := testingUtils.CreateGenerateContext(t, tempDir)
		provider := JavaProvider{}

		startCmd := provider.getStartCmd(ctx)

		require.NotContains(t, startCmd, "-Dserver.port")
		require.Contains(t, startCmd, "build/libs/*jar */build/libs/*jar")
	})
}
