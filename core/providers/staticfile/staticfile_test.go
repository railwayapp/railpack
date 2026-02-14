package staticfile

import (
	"encoding/json"
	"testing"

	"github.com/railwayapp/railpack/core/config"
	"github.com/railwayapp/railpack/core/generate"
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
			name: "index",
			path: "../../../examples/staticfile-index",
			want: true,
		},
		{
			name: "config",
			path: "../../../examples/staticfile-config",
			want: true,
		},
		{
			name: "npm",
			path: "../../../examples/node-npm",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := testingUtils.CreateGenerateContext(t, tt.path)
			provider := StaticfileProvider{}
			got, err := provider.Detect(ctx)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestGetRootDir(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		envVars     map[string]string
		configJSON  string
		want        string
		expectError bool
	}{
		{
			name: "from env var",
			path: "../../../examples/staticfile-index",
			envVars: map[string]string{
				"RAILPACK_STATIC_FILE_ROOT": "/custom/path",
			},
			want:        "/custom/path",
			expectError: false,
		},
		{
			name:       "from provider config",
			path:       "../../../examples/staticfile-index",
			envVars:    map[string]string{},
			configJSON: `{"staticfile":{"root":"dist"}}`,
			want:       "dist",
		},
		{
			name: "env var takes precedence over provider config",
			path: "../../../examples/staticfile-index",
			envVars: map[string]string{
				"RAILPACK_STATIC_FILE_ROOT": "/custom/path",
			},
			configJSON: `{"staticfile":{"root":"dist"}}`,
			want:       "/custom/path",
		},
		{
			name:        "from staticfile config",
			path:        "../../../examples/staticfile-config",
			envVars:     map[string]string{},
			want:        "hello",
			expectError: false,
		},
		{
			name:        "from index.html",
			path:        "../../../examples/staticfile-index",
			envVars:     map[string]string{},
			want:        ".",
			expectError: false,
		},
		{
			name:        "no root dir",
			path:        "../../../examples/node-npm",
			envVars:     map[string]string{},
			want:        "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := testingUtils.CreateGenerateContext(t, tt.path)
			for k, v := range tt.envVars {
				ctx.Env.SetVariable(k, v)
			}

			if tt.configJSON != "" {
				setConfigFromJSON(t, ctx, tt.configJSON)
			}

			got, err := getRootDir(ctx)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func setConfigFromJSON(t *testing.T, ctx *generate.GenerateContext, configJSON string) {
	t.Helper()

	var cfg config.Config
	require.NoError(t, json.Unmarshal([]byte(configJSON), &cfg))
	ctx.Config = &cfg
}
