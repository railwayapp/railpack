package plan

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCheckAndParseDockerignore(t *testing.T) {
	t.Run("nonexistent dockerignore", func(t *testing.T) {
		excludes, includes, err := CheckAndParseDockerignore("/nonexistent/path")
		require.NoError(t, err)
		require.Nil(t, excludes)
		require.Nil(t, includes)
	})

	t.Run("valid dockerignore file", func(t *testing.T) {
		examplePath := filepath.Join("..", "..", "examples", "dockerignore")
		excludes, includes, err := CheckAndParseDockerignore(examplePath)

		require.NoError(t, err)
		require.NotNil(t, excludes)
		require.Nil(t, includes) // No include patterns (starting with !) in the test file

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
		ctx := NewDockerignoreContext("/some/path")
		require.NotNil(t, ctx)
		require.Equal(t, "/some/path", ctx.repoPath)
		require.False(t, ctx.parsed)
		require.Nil(t, ctx.excludes)
		require.Nil(t, ctx.includes)
	})

	t.Run("parse caching", func(t *testing.T) {
		examplePath := filepath.Join("..", "..", "examples", "dockerignore")
		ctx := NewDockerignoreContext(examplePath)

		// First parse
		excludes1, includes1, err1 := ctx.Parse()
		require.NoError(t, err1)
		require.True(t, ctx.parsed)

		// Second parse should return cached results
		excludes2, includes2, err2 := ctx.Parse()
		require.NoError(t, err2)
		require.Equal(t, excludes1, excludes2)
		require.Equal(t, includes1, includes2)
	})

	t.Run("parse with logging", func(t *testing.T) {
		examplePath := filepath.Join("..", "..", "examples", "dockerignore")
		ctx := NewDockerignoreContext(examplePath)

		// Mock logger that captures calls
		logCalls := []string{}
		mockLogger := &mockLogger{logFunc: func(format string, args ...interface{}) {
			logCalls = append(logCalls, format)
		}}

		excludes, includes, err := ctx.ParseWithLogging(mockLogger)
		require.NoError(t, err)
		require.NotNil(t, excludes)
		require.Nil(t, includes)

		// Should have logged that dockerignore was found
		require.Contains(t, logCalls, "Found .dockerignore file, applying filters")
	})

	t.Run("parse nonexistent file", func(t *testing.T) {
		ctx := NewDockerignoreContext("/nonexistent/path")

		excludes, includes, err := ctx.Parse()
		require.NoError(t, err)
		require.Nil(t, excludes)
		require.Nil(t, includes)
		require.True(t, ctx.parsed) // Should still mark as parsed
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
