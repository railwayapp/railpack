package generate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/railwayapp/railpack/core/app"
	"github.com/railwayapp/railpack/core/resolver"
	"github.com/stretchr/testify/require"
)

func TestGetPackageVersions(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mise-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	ctx := CreateTestContext(t, "../../examples/python-uv-tool-versions")

	// Create a resolver
	resolver, err := resolver.NewResolver(tempDir)
	require.NoError(t, err)

	builder := &MiseStepBuilder{
		Resolver: resolver,
		app:      ctx.App,
		env:      ctx.Env,
	}

	packages, err := builder.GetMisePackageVersions(ctx)
	require.NoError(t, err)

	// Expected packages from the example
	expected := map[string]struct{}{
		"python": {},
		"uv":     {},
	}

	// Ensure ONLY expected packages are present
	require.Len(t, packages, len(expected), "unexpected number of packages returned")
	for name := range packages {
		_, ok := expected[name]
		require.True(t, ok, "unexpected package found: %s", name)
	}

	// The python-uv-tool-versions example should have python and uv defined
	require.Contains(t, packages, "python")
	require.Contains(t, packages, "uv")

	// Verify versions are not empty
	require.NotEmpty(t, packages["python"].Version)
	require.NotEmpty(t, packages["uv"].Version)

	// Verify python version starts with "3.9" (as defined in .tool-versions)
	require.True(t, strings.HasPrefix(packages["python"].Version, "3.9"))

	// Verify uv version starts with "0.7" (as defined in .tool-versions)
	require.True(t, strings.HasPrefix(packages["uv"].Version, "0.7"))

	// Verify source types are set
	require.NotEmpty(t, packages["python"].Source)
	require.NotEmpty(t, packages["uv"].Source)
}

func TestGetPackageVersionsWithNoToolVersions(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mise-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	ctx := CreateTestContext(t, "../../examples/node-tanstack-start")

	// Create a resolver
	resolver, err := resolver.NewResolver(tempDir)
	require.NoError(t, err)

	builder := &MiseStepBuilder{
		Resolver: resolver,
		app:      ctx.App,
		env:      ctx.Env,
	}

	packages, err := builder.GetMisePackageVersions(ctx)
	require.NoError(t, err)

	// Should return empty map for directory with no .tool-versions
	require.Empty(t, packages)
}

func TestGetSupportingMiseConfigFiles(t *testing.T) {
	dir := t.TempDir()

	included := []string{
		"mise.toml",
		"mise.local.toml",
		"mise.production.toml",
		"mise.production.local.toml",
		".mise.local.toml",
		".config/mise.local.toml",
		".config/mise.production.local.toml",
		"mise/config.local.toml",
		"mise/config.production.toml",
		".mise/config.local.toml",
		".mise/config.production.toml",
		".config/mise/config.local.toml",
		".config/mise/config.production.local.toml",
		".config/mise/mise.toml",
		".config/mise/mise.local.toml",
		"mise/conf.d/tools.toml",
		"mise/conf.d/tools.local.toml",
		"mise/conf.d/tools.production.toml",
		".mise/conf.d/tools.toml",
		".config/mise/conf.d/tools.toml",
	}
	lockfiles := []string{
		"mise.lock",
		"mise.local.lock",
		"mise.production.lock",
		"mise.production.local.lock",
		".config/mise.local.lock",
		".config/mise.production.local.lock",
		"mise/mise.lock",
		"mise/mise.local.lock",
		"mise/mise.production.lock",
		".mise/mise.lock",
		".mise/mise.local.lock",
		".mise/mise.production.lock",
		".config/mise/mise.lock",
		".config/mise/mise.local.lock",
		".config/mise/mise.production.local.lock",
	}
	excluded := []string{
		"other.toml",
		"mise/tasks/build.toml",
		"mise/conf.d/nested/skip.toml",
		// conf.d fragments lock into the parent directory, not this path.
		"mise/conf.d/mise.lock",
	}

	for _, file := range append(append(included, lockfiles...), excluded...) {
		path := filepath.Join(dir, file)
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		require.NoError(t, os.WriteFile(path, []byte("[tools]\nnode = \"22\"\n"), 0o644))
	}

	userApp, err := app.NewApp(dir)
	require.NoError(t, err)

	got := (&MiseStepBuilder{app: userApp}).getSupportingMiseConfigFiles()
	require.ElementsMatch(t, append(included, lockfiles...), got)
}

func TestPartitionMiseConfigFiles(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"mise.toml":       "[tools]\nnode = \"22\"\n",
		"mise.local.toml": "[tools]\nnode = \"20\"\n",
		".nvmrc":          "20\n",
	}
	for name, contents := range files {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(contents), 0o644))
	}

	userApp, err := app.NewApp(dir)
	require.NoError(t, err)
	require.NoError(t, userApp.SetExcludePatterns([]string{"mise.local.toml", ".nvmrc"}))

	included, excluded := (&MiseStepBuilder{app: userApp}).partitionMiseConfigFiles()
	require.Equal(t, []string{"mise.toml"}, included)
	require.ElementsMatch(t, []string{"mise.local.toml", ".nvmrc"}, excluded)
}

func TestMiseLockfilePath(t *testing.T) {
	cases := map[string]string{
		"mise.toml":                         "mise.lock",
		"mise.local.toml":                   "mise.local.lock",
		"mise.production.toml":              "mise.production.lock",
		"mise.production.local.toml":        "mise.production.local.lock",
		".mise.toml":                        "mise.lock",
		".mise.local.toml":                  "mise.local.lock",
		".config/mise.toml":                 ".config/mise.lock",
		".config/mise.production.toml":      ".config/mise.production.lock",
		".mise/config.toml":                 ".mise/mise.lock",
		".mise/config.local.toml":           ".mise/mise.local.lock",
		".mise/conf.d/foo.toml":             ".mise/mise.lock",
		".mise/conf.d/foo.local.toml":       ".mise/mise.local.lock",
		".config/mise/conf.d/foo.toml":      ".config/mise/mise.lock",
		"mise/conf.d/tools.toml":            "mise/mise.lock",
		"mise/conf.d/tools.production.toml": "mise/mise.lock",
		"mise/config.toml":                  "mise/mise.lock",
		"mise/config.production.local.toml": "mise/mise.production.local.lock",
		".config/mise/config.toml":          ".config/mise/mise.lock",
		".config/mise/mise.toml":            ".config/mise/mise.lock",
		".config/mise/mise.local.toml":      ".config/mise/mise.local.lock",
		"rust-toolchain.toml":               "mise.lock",
	}
	for configPath, lockPath := range cases {
		require.Equal(t, lockPath, miseLockfilePath(configPath), configPath)
	}
}
