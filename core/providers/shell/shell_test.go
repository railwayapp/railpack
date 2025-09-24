package shell

import (
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
