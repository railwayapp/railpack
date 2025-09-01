package core

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/railwayapp/railpack/core/app"
	"github.com/railwayapp/railpack/core/logger"
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

// Ensure that providing an absolute path to a valid config file works
func TestGenerateConfigFromFile_AbsolutePath(t *testing.T) {
	appPath := "../examples/config-file"
	userApp, err := app.NewApp(appPath)
	require.NoError(t, err)

	absConfigPath, err := filepath.Abs(filepath.Join(appPath, "railpack.json"))
	require.NoError(t, err)

	env := app.NewEnvironment(nil)
	l := logger.NewLogger()

	options := &GenerateBuildPlanOptions{ConfigFilePath: absConfigPath}
	cfg, genErr := GenerateConfigFromFile(userApp, env, options, l)

	require.NoError(t, genErr, "expected no error for absolute config path")
	require.NotNil(t, cfg, "config should not be nil")
	// Basic sanity check: provider should be carried through from example config
	require.Equal(t, "node", *cfg.Provider)
}
