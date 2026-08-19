package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// newExcludedApp builds a throwaway app tree and applies exclude patterns to it.
func newExcludedApp(t *testing.T, files []string, patterns []string) *App {
	t.Helper()

	dir := t.TempDir()
	for _, file := range files {
		path := filepath.Join(dir, file)
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0755))
		require.NoError(t, os.WriteFile(path, []byte("{}"), 0644))
	}

	app, err := NewApp(dir)
	require.NoError(t, err)
	require.NoError(t, app.SetExcludePatterns(patterns))

	return app
}

func TestFindFilesRespectsExcludePatterns(t *testing.T) {
	t.Run("no patterns means no filtering", func(t *testing.T) {
		app := newExcludedApp(t, []string{"package.json", "docs/package.json"}, nil)

		files, err := app.FindFiles("**/package.json")
		require.NoError(t, err)
		require.ElementsMatch(t, []string{"package.json", "docs/package.json"}, files)
	})

	t.Run("drops a file matched directly", func(t *testing.T) {
		app := newExcludedApp(t,
			[]string{"package.json", "docs/package.json"},
			[]string{"docs/package.json"},
		)

		files, err := app.FindFiles("**/package.json")
		require.NoError(t, err)
		require.Equal(t, []string{"package.json"}, files)
	})

	t.Run("drops files under an excluded directory", func(t *testing.T) {
		app := newExcludedApp(t,
			[]string{"package.json", "web/node_modules/left-pad/package.json", "web/package.json"},
			[]string{"**/node_modules"},
		)

		files, err := app.FindFiles("**/package.json")
		require.NoError(t, err)
		require.ElementsMatch(t, []string{"package.json", "web/package.json"}, files)
	})

	t.Run("honours negation patterns", func(t *testing.T) {
		app := newExcludedApp(t,
			[]string{"keep/a.json", "keep/b.json"},
			[]string{"keep/*", "!keep/a.json"},
		)

		files, err := app.FindFiles("keep/*.json")
		require.NoError(t, err)
		require.Equal(t, []string{"keep/a.json"}, files)
	})

	t.Run("filters directories and HasMatch too", func(t *testing.T) {
		app := newExcludedApp(t,
			[]string{"docs/marketing/index.json"},
			[]string{"docs"},
		)

		dirs, err := app.FindDirectories("*")
		require.NoError(t, err)
		require.Empty(t, dirs)
		require.False(t, app.HasMatch("docs/**"))
	})

	t.Run("invalid patterns are reported", func(t *testing.T) {
		app := newExcludedApp(t, []string{"package.json"}, nil)
		require.Error(t, app.SetExcludePatterns([]string{"["}))
	})
}
