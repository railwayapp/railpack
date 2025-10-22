package core

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/railwayapp/railpack/core/app"
	"github.com/railwayapp/railpack/core/config"
	"github.com/stretchr/testify/require"
)

func TestGenerateConfigFromEnvironment(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected string
	}{
		{
			name:    "empty environment",
			envVars: map[string]string{},
			expected: `{
				"steps": {},
				"packages": {},
				"caches": {},
				"deploy": {"variables": {}}
			}`,
		},

		{
			name: "kitchen sink",
			envVars: map[string]string{
				"RAILPACK_INSTALL_CMD":         "npm install",
				"RAILPACK_BUILD_CMD":           "npm run build",
				"RAILPACK_START_CMD":           "npm start",
				"RAILPACK_PACKAGES":            "node@18 python@3.9",
				"RAILPACK_BUILD_APT_PACKAGES":  "build-essential libssl-dev",
				"RAILPACK_DEPLOY_APT_PACKAGES": "libssl-dev",
			},
			expected: `{
				"steps": {
					"install": {
						"name": "install",
						"commands": [
							{ "src": ".", "dest": "." },
							"npm install"
						],
						"secrets": ["*"],
						"assets": {},
						"variables": {}
					},
					"build": {
						"name": "build",
						"commands": [
							{ "src": ".", "dest": "." },
							"npm run build"
						],
						"secrets": ["*"],
						"assets": {},
						"variables": {}
					}
				},
				"buildAptPackages": ["build-essential", "libssl-dev"],
				"packages": {
					"node": "18",
					"python": "3.9"
				},
				"caches": {},
				"deploy": {
					"startCommand": "npm start",
					"aptPackages": ["libssl-dev"],
					"variables": {}
				},
				"secrets": ["RAILPACK_BUILD_APT_PACKAGES", "RAILPACK_BUILD_CMD", "RAILPACK_DEPLOY_APT_PACKAGES",
					"RAILPACK_INSTALL_CMD", "RAILPACK_PACKAGES", "RAILPACK_START_CMD"]
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := app.NewEnvironment(&tt.envVars)
			gotConfig := GenerateConfigFromEnvironment(env)

			serializedConfig := config.Config{}
			err := json.Unmarshal([]byte(tt.expected), &serializedConfig)
			require.NoError(t, err)

			if diff := cmp.Diff(serializedConfig, *gotConfig); diff != "" {
				t.Errorf("GenerateConfigFromEnvironment() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
