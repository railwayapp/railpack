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

		patterns, err := CheckAndParseDockerignore(testApp)
		require.NoError(t, err)
		require.Nil(t, patterns)
	})

	t.Run("valid dockerignore file", func(t *testing.T) {
		examplePath := filepath.Join("..", "..", "examples", "dockerignore")
		testApp, err := app.NewApp(examplePath)
		require.NoError(t, err)

		patterns, err := CheckAndParseDockerignore(testApp)

		require.NoError(t, err)
		require.NotNil(t, patterns)

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
			require.Contains(t, patterns, expected, "Expected pattern %s not found in patterns", expected)
		}
	})

	t.Run("dockerignore with negation", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "dockerignore-test")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		dockerignorePath := filepath.Join(tempDir, ".dockerignore")
		err = os.WriteFile(dockerignorePath, []byte("node_modules\n!/.vscode/tasks.json\n"), 0644)
		require.NoError(t, err)

		testApp, err := app.NewApp(tempDir)
		require.NoError(t, err)

		patterns, err := CheckAndParseDockerignore(testApp)
		require.NoError(t, err)
		require.Contains(t, patterns, "node_modules")
		require.Contains(t, patterns, "!.vscode/tasks.json")
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
		patterns, err := CheckAndParseDockerignore(testApp)
		require.Error(t, err)
		require.Contains(t, err.Error(), "error reading .dockerignore")
		require.Nil(t, patterns)
	})
}

func TestDockerignoreContext(t *testing.T) {
	t.Run("new context", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "dockerignore-test")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		testApp, err := app.NewApp(tempDir)
		require.NoError(t, err)

		ctx := NewDockerignoreContext(testApp)
		require.NotNil(t, ctx)
		require.Equal(t, testApp, ctx.app)
		require.False(t, ctx.parsed)
		require.Nil(t, ctx.patterns)
	})

	t.Run("parse caching", func(t *testing.T) {
		examplePath := filepath.Join("..", "..", "examples", "dockerignore")
		testApp, err := app.NewApp(examplePath)
		require.NoError(t, err)

		ctx := NewDockerignoreContext(testApp)

		// First parse
		patterns1, err1 := ctx.Parse()
		require.NoError(t, err1)
		require.True(t, ctx.parsed)

		// Second parse should return cached results
		patterns2, err2 := ctx.Parse()
		require.NoError(t, err2)
		require.Equal(t, patterns1, patterns2)
	})

	t.Run("parse with logging", func(t *testing.T) {
		examplePath := filepath.Join("..", "..", "examples", "dockerignore")
		testApp, err := app.NewApp(examplePath)
		require.NoError(t, err)

		ctx := NewDockerignoreContext(testApp)

		// Mock logger that captures calls
		logCalls := []string{}
		mockLogger := &mockLogger{logFunc: func(format string, args ...interface{}) {
			logCalls = append(logCalls, format)
		}}

		patterns, err := ctx.ParseWithLogging(mockLogger)
		require.NoError(t, err)
		require.NotNil(t, patterns)

		// Should have logged that dockerignore was found
		require.Contains(t, logCalls, "Found .dockerignore file, applying filters")
	})

	t.Run("parse nonexistent file", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "dockerignore-test")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		testApp, err := app.NewApp(tempDir)
		require.NoError(t, err)

		ctx := NewDockerignoreContext(testApp)

		patterns, err := ctx.Parse()
		require.NoError(t, err)
		require.Nil(t, patterns)
		require.True(t, ctx.parsed) // Should still mark as parsed
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

		ctx := NewDockerignoreContext(testApp)
		patterns, err := ctx.Parse()

		require.Error(t, err)
		require.Nil(t, patterns)
		require.False(t, ctx.parsed) // Should not mark as parsed on error
	})
}

// Mock logger for testing
type mockLogger struct {
	logFunc func(string, ...interface{})
}

func (m *mockLogger) LogInfo(format string, args ...interface{}) {
	if m.logFunc != nil {
		m.logFunc(format, args...)
	}
}
