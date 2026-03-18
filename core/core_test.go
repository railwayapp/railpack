package core

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/railwayapp/railpack/core/app"
	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/logger"
	"github.com/railwayapp/railpack/core/resolver"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	v := m.Run()
	snaps.Clean(m, snaps.CleanOpts{Sort: true})
	os.Exit(v)
}

// generate snapshot plan JSON for each build example and assert against it
func TestGenerateBuildPlanForExamples(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)

	// Get all the examples
	examplesDir := filepath.Join(filepath.Dir(wd), "examples")
	entries, err := os.ReadDir(examplesDir)
	require.NoError(t, err)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// For each example, generate a build plan that we can snapshot test
		t.Run(entry.Name(), func(t *testing.T) {
			examplePath := filepath.Join(examplesDir, entry.Name())

			userApp, err := app.NewApp(examplePath)
			require.NoError(t, err)

			env := app.NewEnvironment(nil)
			buildResult := GenerateBuildPlan(userApp, env, &GenerateBuildPlanOptions{})

			if !buildResult.Success {
				t.Fatalf("failed to generate build plan for %s: %s", entry.Name(), buildResult.Logs)
			}

			plan := buildResult.Plan

			// Remove the mise.toml asset since the versions may change between runs
			for _, step := range plan.Steps {
				for name := range step.Assets {
					if name == "mise.toml" {
						step.Assets[name] = "[mise.toml]"
					}
				}
			}

			snaps.MatchStandaloneJSON(t, plan)
		})
	}
}

func TestGenerateConfigFromFile_NotFound(t *testing.T) {
	// Use an existing example app directory so relative paths resolve
	appPath := "../examples/config-file"
	userApp, err := app.NewApp(appPath)
	require.NoError(t, err)

	env := app.NewEnvironment(nil)
	l := logger.NewLogger()

	options := &GenerateBuildPlanOptions{ConfigFilePath: "does-not-exist.railpack.json"}
	cfg, genErr := GenerateConfigFromFile(userApp, env, options, l)

	require.Error(t, genErr, "expected an error when explicit config file does not exist")
	require.Nil(t, cfg, "config should be nil on error")
}

func TestGenerateConfigFromFile_Malformed(t *testing.T) {
	appPath := "../examples/config-file"
	userApp, err := app.NewApp(appPath)
	require.NoError(t, err)

	env := app.NewEnvironment(nil)
	l := logger.NewLogger()

	options := &GenerateBuildPlanOptions{ConfigFilePath: "railpack.malformed.json"}
	cfg, genErr := GenerateConfigFromFile(userApp, env, options, l)

	require.Error(t, genErr, "expected an error for malformed JSON config file")
	require.Nil(t, cfg, "config should be nil on error")
}

func TestGenerateBuildPlan_DockerignoreMetadata(t *testing.T) {
	appPath := "../examples/dockerignore"
	userApp, err := app.NewApp(appPath)
	require.NoError(t, err)

	env := app.NewEnvironment(nil)
	buildResult := GenerateBuildPlan(userApp, env, &GenerateBuildPlanOptions{})

	require.True(t, buildResult.Success)
	require.NotNil(t, buildResult.Metadata)
	require.Equal(t, "true", buildResult.Metadata["dockerIgnore"])
}

func TestAddAdditionalMiseToolsToPackages(t *testing.T) {
	nodeVersion := "24.0.0"
	bunVersion := "1.3.10"

	// Simulate what provider already resolved (node + bun already in table)
	resolvedPackages := map[string]*resolver.ResolvedPackage{
		"node": {Name: "node", ResolvedVersion: &nodeVersion, Source: "mise.toml"},
		"bun":  {Name: "bun", ResolvedVersion: &bunVersion, Source: "mise.toml"},
	}

	// Simulate what mise returns from app mise.toml
	misePackages := map[string]*generate.MisePackageInfo{
		"node":   {Version: "24.0.0", RequestedVersion: "24", Source: "mise.toml"},
		"bun":    {Version: "1.3.10", RequestedVersion: "latest", Source: "mise.toml"},
		"python": {Version: "3.13.0", RequestedVersion: "3.13", Source: "mise.toml"},
		"go":     {Version: "1.23.0", RequestedVersion: "1.23", Source: "mise.toml"},
	}

	addAdditionalMiseToolsFromPackages(misePackages, resolvedPackages)

	// node and bun should remain unchanged
	require.Equal(t, "mise.toml", resolvedPackages["node"].Source)
	require.Equal(t, "mise.toml", resolvedPackages["bun"].Source)

	// python and go should be added with requested versions from app mise.toml
	require.Contains(t, resolvedPackages, "python")
	require.Equal(t, "mise.toml", resolvedPackages["python"].Source)
	require.Equal(t, "3.13.0", *resolvedPackages["python"].ResolvedVersion)
	require.Equal(t, "3.13", *resolvedPackages["python"].RequestedVersion)

	require.Contains(t, resolvedPackages, "go")
	require.Equal(t, "mise.toml", resolvedPackages["go"].Source)
	require.Equal(t, "1.23.0", *resolvedPackages["go"].ResolvedVersion)
	require.Equal(t, "1.23", *resolvedPackages["go"].RequestedVersion)

	// total should now be 4
	require.Len(t, resolvedPackages, 4)
}
