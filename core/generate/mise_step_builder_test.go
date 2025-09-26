package generate

import (
	"os"
	"strings"
	"testing"

	"github.com/railwayapp/railpack/core/resolver"
	"github.com/stretchr/testify/require"
)

func TestGetPackageVersions(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mise-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

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
	defer os.RemoveAll(tempDir)

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
