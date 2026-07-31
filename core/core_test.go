package core

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/railwayapp/railpack/core/app"
	"github.com/railwayapp/railpack/core/logger"
	"github.com/railwayapp/railpack/core/mise"
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
			buildResult, err := GenerateBuildPlan(userApp, env, &GenerateBuildPlanOptions{})
			require.NoError(t, err)

			if !buildResult.Success {
				t.Fatalf("failed to generate build plan for %s: %s", entry.Name(), buildResult.Logs)
			}

			plan := buildResult.Plan

			// Remove the generated-mise-toml asset since the versions may change between runs
			for _, step := range plan.Steps {
				for name := range step.Assets {
					if name == "generated-mise-toml" {
						step.Assets[name] = "[generated-mise-toml]"
					}
				}
			}

			snaps.MatchStandaloneJSON(t, plan)
		})
	}
}

func TestFailedBuildResult(t *testing.T) {
	temporary := fmt.Errorf("failed to ensure mise is installed: %w",
		&mise.TemporaryError{URL: "https://github.com/jdx/mise", Err: errors.New("i/o timeout")})

	result, err := failedBuildResult(logger.NewLogger(), temporary)
	require.ErrorIs(t, err, temporary, "transient failures must reach the caller so they can be retried")
	require.False(t, result.Success)
	require.NotEmpty(t, result.Logs)

	result, err = failedBuildResult(logger.NewLogger(), errors.New("no start command was found"))
	require.NoError(t, err, "deterministic failures are reported in the build result only")
	require.False(t, result.Success)
	require.NotEmpty(t, result.Logs)
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

func TestGetConfig_MergesEnvironmentAndFileSecrets(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, defaultConfigFileName)
	err := os.WriteFile(configPath, []byte(`{"secrets":["VITE_APP_TITLE","FILE_ONLY_SECRET"]}`), 0644)
	require.NoError(t, err)

	userApp, err := app.NewApp(tempDir)
	require.NoError(t, err)

	envVars := map[string]string{
		"DATABASE_URL":      "postgres://localhost/mydb",
		"SENTRY_AUTH_TOKEN": "sntrx_abc",
		"VITE_APP_TITLE":    "MyApp",
		"MY_SECRET":         "s3cret",
	}

	env := app.NewEnvironment(&envVars)
	config, err := GetConfig(userApp, env, &GenerateBuildPlanOptions{}, logger.NewLogger())
	require.NoError(t, err)

	require.ElementsMatch(t, []string{
		"DATABASE_URL",
		"SENTRY_AUTH_TOKEN",
		"VITE_APP_TITLE",
		"MY_SECRET",
		"FILE_ONLY_SECRET",
	}, config.Secrets)
}

func TestGenerateBuildPlan_DockerignoreMetadata(t *testing.T) {
	appPath := "../examples/dockerignore"
	userApp, err := app.NewApp(appPath)
	require.NoError(t, err)

	env := app.NewEnvironment(nil)
	buildResult, err := GenerateBuildPlan(userApp, env, &GenerateBuildPlanOptions{})
	require.NoError(t, err)

	require.True(t, buildResult.Success)
	require.NotNil(t, buildResult.Metadata)
	require.Equal(t, "true", buildResult.Metadata["dockerIgnore"])
	require.NotEmpty(t, buildResult.Plan.Exclude)
}

func TestProviderDefaultExcludes(t *testing.T) {
	tests := []struct {
		name     string
		appPath  string
		excludes []string
	}{
		{name: "node", appPath: "../examples/node-npm", excludes: []string{".git", "node_modules"}},
		{name: "bun", appPath: "../examples/node-bun", excludes: []string{".git", "node_modules"}},
		{name: "deno", appPath: "../examples/deno-2", excludes: []string{".git", "node_modules"}},
		{name: "python", appPath: "../examples/python-pip", excludes: []string{".git", ".venv", "venv"}},
		{name: "other provider", appPath: "../examples/go-mod", excludes: []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userApp, err := app.NewApp(tt.appPath)
			require.NoError(t, err)

			buildResult := GenerateBuildPlan(userApp, app.NewEnvironment(nil), &GenerateBuildPlanOptions{})
			require.True(t, buildResult.Success)
			require.Equal(t, tt.excludes, buildResult.Plan.Exclude)
		})
	}
}
