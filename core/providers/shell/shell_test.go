package shell

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/railwayapp/railpack/core/app"
	testingUtils "github.com/railwayapp/railpack/core/testing"
	"github.com/stretchr/testify/require"
)

func TestDetect(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "shell script",
			path: "../../../examples/shell-script",
			want: true,
		},
		{
			name: "node project",
			path: "../../../examples/node-npm",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := testingUtils.CreateGenerateContext(t, tt.path)
			provider := ShellProvider{}
			got, err := provider.Detect(ctx)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestShellProviderConfigFromFile(t *testing.T) {
	t.Run("shell script from provider config", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, createShellApp(t))
		testingUtils.ClearConfigVariable(ctx, "SHELL_SCRIPT")
		testingUtils.SetConfigFromJSON(t, ctx, `{
			"shell": {
				"script": "deploy.sh"
			}
		}`)

		provider := ShellProvider{}
		require.NoError(t, provider.Initialize(ctx))
		require.Equal(t, "deploy.sh", provider.scriptName)
	})

	t.Run("shell env var takes precedence over provider config", func(t *testing.T) {
		ctx := testingUtils.CreateGenerateContext(t, createShellApp(t))
		testingUtils.ClearConfigVariable(ctx, "SHELL_SCRIPT")
		ctx.Env.SetVariable("RAILPACK_SHELL_SCRIPT", "start.sh")
		testingUtils.SetConfigFromJSON(t, ctx, `{
			"shell": {
				"script": "deploy.sh"
			}
		}`)

		provider := ShellProvider{}
		require.NoError(t, provider.Initialize(ctx))
		require.Equal(t, "start.sh", provider.scriptName)
	})
}

func TestInitialize(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		wantScriptName string
		hasError       bool
	}{
		{
			name:           "default script",
			path:           "../../../examples/shell-script",
			wantScriptName: StartScriptName,
			hasError:       false,
		},
		{
			name:           "default script",
			path:           "../../../examples/config-file",
			wantScriptName: "",
			hasError:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := testingUtils.CreateGenerateContext(t, tt.path)
			provider := ShellProvider{}
			err := provider.Initialize(ctx)

			if tt.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.wantScriptName, provider.scriptName)
			}
		})
	}
}

func TestGetScript(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		envVars        map[string]string
		wantScriptName string
	}{
		{
			name:           "default script",
			path:           "../../../examples/shell-script",
			envVars:        nil,
			wantScriptName: StartScriptName,
		},
		{
			name:           "custom script from env",
			path:           "../../../examples/shell-script",
			envVars:        map[string]string{"SHELL_SCRIPT": "start.sh"},
			wantScriptName: "start.sh",
		},
		{
			name:           "non-existent script from env",
			path:           "../../../examples/shell-script",
			envVars:        map[string]string{"SHELL_SCRIPT": "nonexistent.sh"},
			wantScriptName: StartScriptName, // Falls back to default
		},
		{
			name:           "no script",
			path:           "../../../examples/node-npm", // No shell script here
			envVars:        nil,
			wantScriptName: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := testingUtils.CreateGenerateContext(t, tt.path)

			if tt.envVars != nil {
				// Create a new environment with the test environment variables
				envVars := tt.envVars // Create a local copy
				ctx.Env = app.NewEnvironment(&envVars)
			}

			scriptName := getScript(ctx)
			require.Equal(t, tt.wantScriptName, scriptName)
		})
	}
}

func TestDetectShellInterpreter(t *testing.T) {
	tests := []struct {
		name            string
		scriptContent   string
		wantInterpreter string
	}{
		{
			name:            "bash shebang",
			scriptContent:   "#!/bin/bash\necho 'hello'",
			wantInterpreter: "bash",
		},
		{
			name:            "bash with env",
			scriptContent:   "#!/usr/bin/env bash\necho 'hello'",
			wantInterpreter: "bash",
		},
		{
			name:            "sh shebang",
			scriptContent:   "#!/bin/sh\necho 'hello'",
			wantInterpreter: "sh",
		},
		{
			name:            "dash shebang",
			scriptContent:   "#!/bin/dash\necho 'hello'",
			wantInterpreter: "sh",
		},
		{
			name:            "zsh shebang",
			scriptContent:   "#!/bin/zsh\necho 'hello'",
			wantInterpreter: "zsh",
		},
		{
			name:            "no shebang",
			scriptContent:   "echo 'hello'",
			wantInterpreter: "sh",
		},
		{
			name:            "empty file",
			scriptContent:   "",
			wantInterpreter: "sh",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			scriptPath := filepath.Join(tmpDir, "test.sh")
			err := os.WriteFile(scriptPath, []byte(tt.scriptContent), 0644)
			require.NoError(t, err)

			ctx := testingUtils.CreateGenerateContext(t, tmpDir)
			got, err := detectShellInterpreter(ctx, "test.sh")
			require.NoError(t, err)
			require.Equal(t, tt.wantInterpreter, got)
		})
	}
}

func createShellApp(t *testing.T) string {
	t.Helper()

	appDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(appDir, "start.sh"), []byte("#!/bin/sh\necho start\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(appDir, "deploy.sh"), []byte("#!/bin/sh\necho deploy\n"), 0644))

	return appDir
}
