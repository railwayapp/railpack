package config

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/railwayapp/railpack/core/plan"
	"github.com/railwayapp/railpack/internal/utils"
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
	providerSchema, ok := schema.Properties.Get("provider")
	require.True(t, ok)
	require.ElementsMatch(t, []any{
		"php",
		"golang",
		"java",
		"rust",
		"ruby",
		"elixir",
		"python",
		"deno",
		"dotnet",
		"node",
		"gleam",
		"cpp",
		"staticfile",
		"shell",
	}, providerSchema.Enum)

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

func TestDeployOutputsShorthand(t *testing.T) {
	jsonConfig := `{
		"steps": {
			"build": {
				"inputs": [".", "$install"],
				"commands": ["echo build"],
				"deployOutputs": [
					// The bare "dist" entry exercises shorthand alongside a full filter object.
					"dist",
					{"include": ["public"], "exclude": ["public/**/*.map"]}
				]
			}
		},
		"deploy": {
			"inputs": [{"step": "build", "include": ["other"]}]
		}
	}`
	configJSON, err := utils.StandardizeJSON([]byte(jsonConfig))
	require.NoError(t, err)
	var c Config

	err = json.Unmarshal(configJSON, &c)

	require.NoError(t, err)
	step := c.Steps["build"]

	require.NotNil(t, step)
	require.Equal(t, DeployOutputs{
		plan.NewIncludeFilter([]string{"dist"}),
		plan.NewFilter([]string{"public"}, []string{"public/**/*.map"}),
	}, step.DeployOutputs)
	// Layer shortcuts and object inputs must retain their own decoding rules.
	require.Equal(t, []plan.Layer{plan.NewLocalLayer(), plan.NewStepLayer("install")}, step.Inputs)
	require.Len(t, step.Commands, 1)
	require.Equal(t, []plan.Layer{plan.NewStepLayer("build", plan.NewIncludeFilter([]string{"other"}))}, c.Deploy.Inputs)
}

func TestDeployOutputsSchemaAcceptsPathStringsAndFilterObjects(t *testing.T) {
	schema := GetJsonSchema()

	steps, ok := schema.Properties.Get("steps")

	require.True(t, ok)

	outputs, ok := steps.AdditionalProperties.Properties.Get("deployOutputs")

	require.True(t, ok)
	require.Equal(t, "array", outputs.Type)
	require.Len(t, outputs.Items.OneOf, 2)
	require.Equal(t, "string", outputs.Items.OneOf[0].Type)
	object := outputs.Items.OneOf[1]

	require.Equal(t, "object", object.Type)
	for _, name := range []string{"include", "exclude"} {
		property, ok := object.Properties.Get(name)

		require.True(t, ok)
		require.Equal(t, "array", property.Type)
		require.Equal(t, "string", property.Items.Type)
	}

	// Layer inputs still need their schema for step references and input shortcuts.
	inputs, ok := steps.AdditionalProperties.Properties.Get("inputs")

	require.True(t, ok)

	stepInput, ok := inputs.Items.OneOf[0].Properties.Get("step")

	require.True(t, ok)
	require.Equal(t, "string", stepInput.Type)
}
