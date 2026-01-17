package plan

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/railwayapp/railpack/core/app"
	"github.com/stretchr/testify/require"
)

func TestCheckAndParseDockerignore(t *testing.T) {
	t.Run("nonexistent dockerignore", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "dockerignore-test")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		testApp, err := app.NewApp(tempDir)
		require.NoError(t, err)

		excludes, includes, err := CheckAndParseDockerignore(testApp)
		require.NoError(t, err)
		require.Nil(t, excludes)
		require.Nil(t, includes)
	})

	t.Run("valid dockerignore file", func(t *testing.T) {
		examplePath := filepath.Join("..", "..", "examples", "dockerignore")
		testApp, err := app.NewApp(examplePath)
		require.NoError(t, err)

		excludes, includes, err := CheckAndParseDockerignore(testApp)

		require.NoError(t, err)
		require.NotNil(t, excludes)
		require.NotNil(t, includes)
		require.Contains(t, includes, "negation_test/should_exist.txt")
		require.Contains(t, includes, "negation_test/existing_folder")

		// Verify some expected patterns from examples/dockerignore/.dockerignore
		// Note: patterns are parsed by the moby/patternmatcher library
		expectedPatterns := []string{
			".vscode",
			".copier", // Leading slash is stripped
			".env-specific",
			".env*",
			"__pycache__", // Trailing slash is stripped
			"test",        // Leading slash is stripped
			"tmp/*",       // Leading slash is stripped
			"*.log",
			"Justfile",
			"TODO*",     // Leading slash is stripped
			"README.md", // Leading slash is stripped
			"docker-compose*.yml",
		}

		for _, expected := range expectedPatterns {
			require.Contains(t, excludes, expected, "Expected pattern %s not found in excludes", expected)
		}
	})

	t.Run("inaccessible dockerignore", func(t *testing.T) {
		// Create a temporary directory and file
		tempDir, err := os.MkdirTemp("", "dockerignore-test")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		dockerignorePath := filepath.Join(tempDir, ".dockerignore")
		err = os.WriteFile(dockerignorePath, []byte("*.log\nnode_modules\n"), 0644)
		require.NoError(t, err)

		// Make the file unreadable (this simulates permission errors)
		err = os.Chmod(dockerignorePath, 0000)
		require.NoError(t, err)
		defer func() { _ = os.Chmod(dockerignorePath, 0644) }() // Restore permissions for cleanup

		testApp, err := app.NewApp(tempDir)
		require.NoError(t, err)

		// This should fail with a permission error
		excludes, includes, err := CheckAndParseDockerignore(testApp)
		require.Error(t, err)
		require.Contains(t, err.Error(), "error reading .dockerignore")
		require.Nil(t, excludes)
		require.Nil(t, includes)
	})
}

func TestSeparatePatterns(t *testing.T) {
	t.Run("only exclude patterns", func(t *testing.T) {
		patterns := []string{"*.log", "node_modules", "/tmp"}
		excludes, includes := separatePatterns(patterns)

		require.Equal(t, patterns, excludes)
		require.Empty(t, includes)
	})

	t.Run("only include patterns", func(t *testing.T) {
		patterns := []string{"!important.log", "!keep/this"}
		excludes, includes := separatePatterns(patterns)

		require.Empty(t, excludes)
		require.Equal(t, []string{"important.log", "keep/this"}, includes)
	})

	t.Run("mixed patterns", func(t *testing.T) {
		patterns := []string{"*.log", "!important.log", "node_modules", "!node_modules/keep"}
		excludes, includes := separatePatterns(patterns)

		require.Equal(t, []string{"*.log", "node_modules"}, excludes)
		require.Equal(t, []string{"important.log", "node_modules/keep"}, includes)
	})

	t.Run("empty patterns", func(t *testing.T) {
		patterns := []string{}
		excludes, includes := separatePatterns(patterns)

		require.Empty(t, excludes)
		require.Empty(t, includes)
	})

	t.Run("empty string patterns", func(t *testing.T) {
		patterns := []string{"", "*.log", "", "!keep.log"}
		excludes, includes := separatePatterns(patterns)

		require.Equal(t, []string{"", "*.log", ""}, excludes)
		require.Equal(t, []string{"keep.log"}, includes)
	})
}

func TestDockerignoreContext(t *testing.T) {
	t.Run("new context", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "dockerignore-test")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		testApp, err := app.NewApp(tempDir)
		require.NoError(t, err)

		ctx, err := NewDockerignoreContext(testApp)
		require.NoError(t, err)
		require.NotNil(t, ctx)
		require.False(t, ctx.HasFile)
		require.Nil(t, ctx.Excludes)
		require.Nil(t, ctx.Includes)
	})

	t.Run("context with dockerignore file", func(t *testing.T) {
		examplePath := filepath.Join("..", "..", "examples", "dockerignore")
		testApp, err := app.NewApp(examplePath)
		require.NoError(t, err)

		ctx, err := NewDockerignoreContext(testApp)
		require.NoError(t, err)
		require.True(t, ctx.HasFile)
		require.NotNil(t, ctx.Excludes)
		require.NotNil(t, ctx.Includes)
		require.Contains(t, ctx.Includes, "negation_test/should_exist.txt")
		require.Contains(t, ctx.Includes, "negation_test/existing_folder")
	})

	t.Run("parse nonexistent file", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "dockerignore-test")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		testApp, err := app.NewApp(tempDir)
		require.NoError(t, err)

		ctx, err := NewDockerignoreContext(testApp)
		require.NoError(t, err)
		require.False(t, ctx.HasFile)
		require.Nil(t, ctx.Excludes)
		require.Nil(t, ctx.Includes)
	})

	t.Run("parse error handling", func(t *testing.T) {
		// Create a temporary directory with an inaccessible .dockerignore
		tempDir, err := os.MkdirTemp("", "dockerignore-test")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		dockerignorePath := filepath.Join(tempDir, ".dockerignore")
		err = os.WriteFile(dockerignorePath, []byte("*.log\n"), 0644)
		require.NoError(t, err)

		// Make the file unreadable
		err = os.Chmod(dockerignorePath, 0000)
		require.NoError(t, err)
		defer func() { _ = os.Chmod(dockerignorePath, 0644) }()

		testApp, err := app.NewApp(tempDir)
		require.NoError(t, err)

		ctx, err := NewDockerignoreContext(testApp)
		require.Error(t, err)
		require.Nil(t, ctx)
	})
}

func TestDockerignoreDuplicatePatterns(t *testing.T) {
	t.Run("duplicate patterns removed", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "dockerignore-test")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		// Create test files
		err = os.WriteFile(filepath.Join(tempDir, "keep.txt"), []byte("exists"), 0644)
		require.NoError(t, err)

		// Create .dockerignore with duplicate patterns
		dockerignoreContent := `*.log
*.log
node_modules
!keep.txt
!keep.txt
`
		err = os.WriteFile(filepath.Join(tempDir, ".dockerignore"), []byte(dockerignoreContent), 0644)
		require.NoError(t, err)

		testApp, err := app.NewApp(tempDir)
		require.NoError(t, err)

		ctx, err := NewDockerignoreContext(testApp)
		require.NoError(t, err)

		// Count occurrences of each pattern
		logCount := 0
		nodeModulesCount := 0
		for _, pattern := range ctx.Excludes {
			if pattern == "*.log" {
				logCount++
			}
			if pattern == "node_modules" {
				nodeModulesCount++
			}
		}

		keepCount := 0
		for _, pattern := range ctx.Includes {
			if pattern == "keep.txt" {
				keepCount++
			}
		}

		// Verify no duplicates exist
		require.LessOrEqual(t, logCount, 1, "*.log pattern should appear at most once")
		require.LessOrEqual(t, nodeModulesCount, 1, "node_modules pattern should appear at most once")
		require.LessOrEqual(t, keepCount, 1, "keep.txt pattern should appear at most once")
	})
}

func TestCheckAndParseDockerignoreWithNegation(t *testing.T) {
	t.Run("negated patterns with existing and non-existing files and folders", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "dockerignore-test")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		// Create test files
		err = os.MkdirAll(filepath.Join(tempDir, "negation_test", "existing_folder"), 0755)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(tempDir, "negation_test", "should_exist.txt"), []byte("exists"), 0644)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(tempDir, "negation_test", "existing_folder", "file.txt"), []byte("exists"), 0644)
		require.NoError(t, err)

		// Create .dockerignore with mixed negation cases
		dockerignoreContent := `
negation_test/*
!negation_test/should_exist.txt
!negation_test/should_not_exist.txt
!negation_test/folder_does_not_exist/
!negation_test/existing_folder/
`
		err = os.WriteFile(filepath.Join(tempDir, ".dockerignore"), []byte(dockerignoreContent), 0644)
		require.NoError(t, err)

		testApp, err := app.NewApp(tempDir)
		require.NoError(t, err)

		excludes, includes, err := CheckAndParseDockerignore(testApp)
		require.NoError(t, err)

		// Check excludes
		require.Contains(t, excludes, "negation_test/*")

		// Check includes - should only contain the file/folder that actually exists
		require.Contains(t, includes, "negation_test/should_exist.txt")
		require.Contains(t, includes, "negation_test/existing_folder")
		require.Contains(t, includes, "negation_test/existing_folder")
		require.NotContains(t, includes, "negation_test/should_not_exist.txt")
		require.NotContains(t, includes, "negation_test/folder_does_not_exist/")
		require.Len(t, includes, 2)
	})
}
