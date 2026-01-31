package config

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
)

func TestEmptyConfig(t *testing.T) {
	config := EmptyConfig()
	require.NotNil(t, config)
	require.Empty(t, config.Caches)
	require.Empty(t, config.Packages)
	require.Empty(t, config.BuildAptPackages)
	require.Empty(t, config.Steps)
	require.Nil(t, config.Provider)

	require.NotNil(t, config.Deploy)
	require.Nil(t, config.Deploy.Inputs)
	require.Nil(t, config.Deploy.Paths)
}

func TestMergeConfigSmall(t *testing.T) {
	config1JSON := `{
		"baseImage": "ubuntu:20.04",
		"packages": {
			"python": "latest",
			"node": "22"
		},
		"aptPackages": ["git"],
		"steps": {
			"install": {
				"dependsOn": ["packages"],
				"commands": [
					"echo first"
				],
				"variables": {
					"HELLO": "world"
				},
				"caches": ["pip"]
			}
		},
		"deploy": {
			"variables": {
				"SHARED": "one",
				"HELLO": "world"
			}
		}
	}`

	config2JSON := `{
		"baseImage": "secondd",
		"packages": {
			"node": "23",
			"bun": "latest"
		},
		"steps": {
			"install": {
				"variables": {
					"another": "boop"
				}
			}
		},
		"deploy": {
			"variables": {
				"SHARED": "two",
				"FOO": "bar"
			}
		}
	}`

	expectedJSON := `{
		"baseImage": "secondd",
		"packages": {
			"python": "latest",
			"node": "23",
			"bun": "latest"
		},
		"aptPackages": ["git"],
		"steps": {
			"install": {
				"dependsOn": ["packages"],
				"commands": [
					"echo first"
				],
				"variables": {
					"HELLO": "world",
					"another": "boop"
				},
				"caches": ["pip"]
			}
		},
		"deploy": {
			"variables": {
				"HELLO": "world",
				"SHARED": "two",
				"FOO": "bar"
			}
		},
		"caches": {}
	}`

	var config1, config2, expected Config
	require.NoError(t, json.Unmarshal([]byte(config1JSON), &config1))
	require.NoError(t, json.Unmarshal([]byte(config2JSON), &config2))
	require.NoError(t, json.Unmarshal([]byte(expectedJSON), &expected))

	result := Merge(&config1, &config2)

	if diff := cmp.Diff(expected, *result); diff != "" {
		t.Errorf("configs mismatch (-want +got):\n%s", diff)
	}

}

func TestMergeConfig(t *testing.T) {
	config1JSON := `{
		"baseImage": "ubuntu:20.04",
		"packages": {
			"python": "latest",
			"node": "22"
		},
		"aptPackages": ["git"],
		"caches": {
			"npm": {
				"directory": "/root/.npm",
				"type": "locked"
			},
			"pip": {
				"directory": "/root/.cache/pip"
			}
		},
		"secrets": ["SECRET_1", "API_KEY"],
		"steps": {
			"install": {
				"name": "install",
				"secrets": ["*"],
				"assets": {
					"package.json": "content1",
					"requirements.txt": "content2"
				},
				"commands": [
					{"type": "exec", "cmd": "npm install", "caches": ["npm"], "customName": "Install NPM deps"},
					{"type": "path", "path": "/usr/local/bin"},
					{"type": "variable", "name": "NODE_ENV", "value": "production"},
					{"type": "copy", "src": "/src", "dest": "/app", "image": "alpine:latest"},
					{"type": "file", "path": "/app", "name": "config.json", "mode": 384, "customName": "Write config"}
				]
			},
			"build": {
				"name": "build",
				"commands": [
					{"type": "exec", "cmd": "config 1 a"},
					{"type": "exec", "cmd": "config 1 b"}
				]
			}
		},
		"deploy": {
			"startCommand": "python app.py"
		}
	}`

	config2JSON := `{
		"providers": ["node"],
		"baseImage": "ubuntu:22.04",
		"packages": {
			"node": "23"
		},
		"aptPackages": ["curl"],
		"caches": {
			"npm": {
				"directory": "/root/.npm-new",
				"type": "shared"
			},
			"go": {
				"directory": "/root/.cache/go-build"
			}
		},
		"secrets": ["SECRET_2"],
		"steps": {
			"install": {
				"name": "install",
				"secrets": ["*"],
				"assets": {
					"package.json": "content3"
				}
			},
			"build": {
				"name": "build",
				"secrets": [],
				"commands": [
					{"type": "exec", "cmd": "config 2 a"},
					{"type": "exec", "cmd": "config 2 b"}
				]
			}
		},
		"deploy": {
			"aptPackages": ["curl"],
			"startCommand": "node server.js",
			"paths": ["/usr/local/bin", "/app/bin"]
		}
	}`

	expectedJSON := `{
		"providers": ["node"],
		"baseImage": "ubuntu:22.04",
		"packages": {
			"python": "latest",
			"node": "23"
		},
		"aptPackages": ["curl"],
		"caches": {
			"npm": {
				"directory": "/root/.npm-new",
				"type": "shared"
			},
			"go": {
				"directory": "/root/.cache/go-build"
			},
			"pip": {
				"directory": "/root/.cache/pip"
			}
		},
		"secrets": ["SECRET_2"],
		"steps": {
			"install": {
				"name": "install",
				"secrets": ["*"],
				"assets": {
					"package.json": "content3",
					"requirements.txt": "content2"
				},
				"commands": [
					{"type": "exec", "cmd": "npm install", "caches": ["npm"], "customName": "Install NPM deps"},
					{"type": "path", "path": "/usr/local/bin"},
					{"type": "variable", "name": "NODE_ENV", "value": "production"},
					{"type": "copy", "src": "/src", "dest": "/app", "image": "alpine:latest"},
					{"type": "file", "path": "/app", "name": "config.json", "mode": 384, "customName": "Write config"}
				]
			},
			"build": {
				"name": "build",
				"secrets": [],
				"commands": [
					{"type": "exec", "cmd": "config 2 a"},
					{"type": "exec", "cmd": "config 2 b"}
				]
			}
		},
		"deploy": {
			"aptPackages": ["curl"],
			"startCommand": "node server.js",
			"paths": ["/usr/local/bin", "/app/bin"]
		}
	}`

	var config1, config2, expected Config
	require.NoError(t, json.Unmarshal([]byte(config1JSON), &config1))
	require.NoError(t, json.Unmarshal([]byte(config2JSON), &config2))
	require.NoError(t, json.Unmarshal([]byte(expectedJSON), &expected))

	result := Merge(&config1, &config2)

	if diff := cmp.Diff(expected, *result); diff != "" {
		t.Errorf("configs mismatch (-want +got):\n%s", diff)
	}
}

func TestMergeConfigStart(t *testing.T) {
	config1JSON := `{
		"deploy": {
			"startCommand": "python app.py"
		}
	}`

	config2JSON := `{
		"packages": {
			"node": "23"
		}
	}`

	expectedJSON := `{
		"packages": {
			"node": "23"
		},
		"deploy": {
			"startCommand": "python app.py"
		},
		"steps": {},
		"caches": {}
	}`

	var config1, config2, expected Config
	require.NoError(t, json.Unmarshal([]byte(config1JSON), &config1))
	require.NoError(t, json.Unmarshal([]byte(config2JSON), &config2))
	require.NoError(t, json.Unmarshal([]byte(expectedJSON), &expected))

	result := Merge(&config1, &config2)

	if diff := cmp.Diff(expected, *result); diff != "" {
		t.Errorf("configs mismatch (-want +got):\n%s", diff)
	}
}

func TestGetJsonSchema(t *testing.T) {
	schema := GetJsonSchema()
	require.NotEmpty(t, schema)

	require.NotContains(t, schema.Required, "provider")

	schemaJson, err := json.MarshalIndent(schema, "", "  ")
	require.NoError(t, err)
	require.NotEmpty(t, schemaJson)
}

func TestConfigExcludeInclude(t *testing.T) {
	configJSON := `{
		"exclude": [
			"node_modules",
			"*.log",
			".env",
			"!important.log"
		]
	}`

	var config Config
	require.NoError(t, json.Unmarshal([]byte(configJSON), &config))

	require.Equal(t, []string{"node_modules", "*.log", ".env", "!important.log"}, config.Exclude)
}

func TestMergeConfigExcludeInclude(t *testing.T) {
	config1JSON := `{
		"exclude": ["node_modules", ".env", "!important.log"]
	}`

	config2JSON := `{
		"exclude": ["*.tmp", "!keep.tmp"]
	}`

	expectedJSON := `{
		"exclude": ["*.tmp", "!keep.tmp"],
		"steps": {},
		"packages": {},
		"caches": {},
		"deploy": {}
	}`

	var config1, config2, expected Config
	require.NoError(t, json.Unmarshal([]byte(config1JSON), &config1))
	require.NoError(t, json.Unmarshal([]byte(config2JSON), &config2))
	require.NoError(t, json.Unmarshal([]byte(expectedJSON), &expected))

	result := Merge(&config1, &config2)

	if diff := cmp.Diff(expected, *result); diff != "" {
		t.Errorf("configs mismatch (-want +got):\n%s", diff)
	}
}
