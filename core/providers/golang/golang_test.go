package golang

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/railwayapp/railpack/core/config"
	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/plan"
	testingUtils "github.com/railwayapp/railpack/core/testing"
	"github.com/stretchr/testify/require"
)

func TestGolang(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		detected     bool
		hasGoMod     bool
		hasWorkspace bool
		goVersion    string
		cgoEnabled   bool
	}{
		{
			name:      "go mod",
			path:      "../../../examples/go-mod",
			detected:  true,
			hasGoMod:  true,
			goVersion: "1.25.3",
		},
		{
			name:      "go cmd dirs",
			path:      "../../../examples/go-cmd-dirs",
			detected:  true,
			hasGoMod:  true,
			goVersion: "1.25.3",
		},
		{
			name:         "go workspaces",
			path:         "../../../examples/go-workspaces",
			detected:     true,
			hasGoMod:     false,
			hasWorkspace: true,
			goVersion:    "1.25",
		},
		{
			name:     "node",
			path:     "../../../examples/node-npm",
			detected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := testingUtils.CreateGenerateContext(t, tt.path)
			provider := GoProvider{}
			detected, err := provider.Detect(ctx)
			require.NoError(t, err)
			require.Equal(t, tt.detected, detected)

			if detected {
				err = provider.Initialize(ctx)
				require.NoError(t, err)

				err = provider.Plan(ctx)
				require.NoError(t, err)

				require.Equal(t, tt.hasGoMod, provider.isGoMod(ctx))
				require.Equal(t, tt.hasWorkspace, provider.isGoWorkspace(ctx))
				require.Equal(t, tt.cgoEnabled, provider.hasCGOEnabled(ctx))

				if tt.goVersion != "" {
					goVersion := ctx.Resolver.Get("go")
					require.Equal(t, tt.goVersion, goVersion.Version)
				}

				if tt.hasWorkspace {
					packages := provider.GoWorkspacePackages(ctx)
					require.Greater(t, len(packages), 0, "workspace should have at least one package")
				}
			}
		})
	}
}

func TestGolangProviderConfigFromFile(t *testing.T) {
	t.Run("go version from provider config", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, createGoCmdDirsApp(t))
		setConfigFromJSON(t, ctx, `{
			"golang": {
				"version": "1.24.6"
			}
		}`)

		provider := GoProvider{}
		require.NoError(t, provider.Initialize(ctx))
		require.NoError(t, provider.Plan(ctx))

		goVersion := ctx.Resolver.Get("go")
		require.Equal(t, "1.24.6", goVersion.Version)
	})

	t.Run("go bin from provider config", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, createGoCmdDirsApp(t))
		setConfigFromJSON(t, ctx, `{
			"golang": {
				"bin": "worker"
			}
		}`)

		provider := GoProvider{}
		require.NoError(t, provider.Initialize(ctx))
		require.NoError(t, provider.Plan(ctx))

		buildCmd := getBuildExecCommand(t, ctx)
		require.Equal(t, `go build -ldflags="-w -s" -o out ./cmd/worker`, buildCmd)
	})

	t.Run("workspace module from provider config", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, "../../../examples/go-workspaces")
		setConfigFromJSON(t, ctx, `{
			"golang": {
				"workspaceModule": "shared"
			}
		}`)

		provider := GoProvider{}
		require.NoError(t, provider.Initialize(ctx))
		require.NoError(t, provider.Plan(ctx))

		buildCmd := getBuildExecCommand(t, ctx)
		require.Equal(t, `go build -ldflags="-w -s" -o out ./shared`, buildCmd)
	})

	t.Run("cgo enabled from provider config", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, createGoCmdDirsApp(t))
		setConfigFromJSON(t, ctx, `{
			"golang": {
				"cgoEnabled": true
			}
		}`)

		provider := GoProvider{}
		require.NoError(t, provider.Initialize(ctx))
		require.NoError(t, provider.Plan(ctx))

		require.True(t, provider.hasCGOEnabled(ctx))
		require.Contains(t, ctx.Deploy.AptPackages, "libc6")

		installVariables := getStepVariables(t, ctx, "install")
		_, hasCGOEnabled := installVariables["CGO_ENABLED"]
		require.False(t, hasCGOEnabled)
	})

	t.Run("cgo env var takes precedence over provider config", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, createGoCmdDirsApp(t))
		ctx.Env.SetVariable("CGO_ENABLED", "0")
		setConfigFromJSON(t, ctx, `{
			"golang": {
				"cgoEnabled": true
			}
		}`)

		provider := GoProvider{}
		require.NoError(t, provider.Initialize(ctx))
		require.NoError(t, provider.Plan(ctx))

		require.False(t, provider.hasCGOEnabled(ctx))
		require.NotContains(t, ctx.Deploy.AptPackages, "libc6")

		installVariables := getStepVariables(t, ctx, "install")
		require.Equal(t, "0", installVariables["CGO_ENABLED"])
	})
}

// Keep config setup in one place so each test case only describes the provider
// behavior it is validating.
func setConfigFromJSON(t *testing.T, ctx *generate.GenerateContext, configJSON string) {
	t.Helper()

	var cfg config.Config
	require.NoError(t, json.Unmarshal([]byte(configJSON), &cfg))
	ctx.Config = &cfg
}

// These tests assert selection logic by inspecting the generated build command,
// since bin/workspace-module choices are encoded there.
func getBuildExecCommand(t *testing.T, ctx *generate.GenerateContext) string {
	t.Helper()

	for _, step := range ctx.Steps {
		if step.Name() != "build" {
			continue
		}

		buildStep, ok := step.(*generate.CommandStepBuilder)
		require.True(t, ok)

		for _, command := range buildStep.Commands {
			execCommand, ok := command.(plan.ExecCommand)
			if ok {
				return execCommand.Cmd
			}
		}
	}

	t.Fatalf("build step exec command not found")
	return ""
}

func getStepVariables(t *testing.T, ctx *generate.GenerateContext, stepName string) map[string]string {
	t.Helper()

	for _, step := range ctx.Steps {
		if step.Name() != stepName {
			continue
		}

		commandStep, ok := step.(*generate.CommandStepBuilder)
		require.True(t, ok)

		return commandStep.Variables
	}

	t.Fatalf("%s step not found", stepName)
	return nil
}

// Build a temporary multi-command Go app so bin selection can be tested
// deterministically without mutating shared example fixtures.
func createGoCmdDirsApp(t *testing.T) string {
	t.Helper()

	appDir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(appDir, "go.mod"), []byte(`module example.com/testapp

go 1.25
`), 0644))

	require.NoError(t, os.MkdirAll(filepath.Join(appDir, "cmd", "server"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(appDir, "cmd", "worker"), 0755))

	goMainFile := []byte(`package main

func main() {}
`)
	require.NoError(t, os.WriteFile(filepath.Join(appDir, "cmd", "server", "main.go"), goMainFile, 0644))
	require.NoError(t, os.WriteFile(filepath.Join(appDir, "cmd", "worker", "main.go"), goMainFile, 0644))

	return appDir
}
