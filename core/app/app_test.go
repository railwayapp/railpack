package app

import (
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestPackageJSON represents the package.json structure for testing
type PackageJSON struct {
	Name    string `json:"name"`
	Engines struct {
		Node string `json:"node"`
	} `json:"engines"`
	Scripts struct {
		PostStart string
		PreStart  string
		Start     string
	}
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`

	AllScripts map[string]string `json:"scripts"`
}

func TestApp(t *testing.T) {
	app, err := NewApp("../../examples/node-bun")
	require.NoError(t, err)

	content, err := app.ReadFile("package.json")
	require.NoError(t, err)
	require.Contains(t, content, "node-bun")

	var packageJSON PackageJSON
	err = app.ReadJSON("package.json", &packageJSON)
	require.NoError(t, err)
	require.Equal(t, packageJSON.Name, "node-bun")

	files, err := app.FindFiles("*.ts")
	require.NoError(t, err)
	require.Equal(t, len(files), 1)
	require.Equal(t, files[0], "index.ts")
}

func TestAppAbsolutePath(t *testing.T) {
	relPath := "../../examples/node-bun"
	absPath, err := filepath.Abs(relPath)
	require.NoError(t, err)

	app, err := NewApp(absPath)
	require.NoError(t, err)

	require.Equal(t, app.Source, absPath)
}

func TestAppReadJsonWithComments(t *testing.T) {
	app, err := NewApp("../../examples/config-file")
	require.NoError(t, err)

	var config map[string]interface{}
	err = app.ReadJSON("hello.jsonc", &config)
	require.NoError(t, err)
	require.Equal(t, config["hello"], "world")
}

func TestFindFilesWithContent(t *testing.T) {
	app, err := NewApp("../../examples/node-bun")
	require.NoError(t, err)

	// Test finding files containing "process.stdout.write"
	regex := regexp.MustCompile(`process\.stdout\.write`)
	matches := app.FindFilesWithContent("*.ts", regex)
	require.Equal(t, len(matches), 1)
	require.Equal(t, matches[0], "index.ts")

	// Test finding files with non-existent pattern
	regex = regexp.MustCompile(`nonexistent`)
	matches = app.FindFilesWithContent("*.ts", regex)
	require.Empty(t, matches)

	// Test with invalid glob pattern
	regex = regexp.MustCompile(`test`)
	matches = app.FindFilesWithContent("[invalid", regex)
	require.Empty(t, matches)
}
