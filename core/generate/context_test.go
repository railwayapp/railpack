package generate

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/railwayapp/railpack/core/app"
	"github.com/railwayapp/railpack/core/config"
	"github.com/railwayapp/railpack/core/logger"
	"github.com/railwayapp/railpack/core/plan"
	"github.com/stretchr/testify/require"
)

type TestProvider struct{}

func (p *TestProvider) Plan(ctx *GenerateContext) error {
	// mise
	mise := ctx.GetMiseStepBuilder()
	nodeRef := mise.Default("node", "18")
	mise.Version(nodeRef, "18", "test")

	// commands
	installStep := ctx.NewCommandStep("install")
	installStep.AddCommand(plan.NewExecCommand("npm install", plan.ExecOptions{}))
	installStep.AddInput(mise.GetLayer())
	installStep.Secrets = []string{}

	buildStep := ctx.NewCommandStep("build")
	buildStep.AddCommand(plan.NewExecCommand("npm run build", plan.ExecOptions{}))
	buildStep.AddInput(plan.NewStepLayer(installStep.Name()))

	ctx.Deploy.DeployInputs = []plan.Layer{
		plan.NewStepLayer(buildStep.Name()),
	}

	return nil
}

func CreateTestContext(t *testing.T, path string) *GenerateContext {
	t.Helper()

	userApp, err := app.NewApp(path)
	require.NoError(t, err)

	env := app.NewEnvironment(nil)
	config := config.EmptyConfig()

	ctx, err := NewGenerateContext(userApp, env, config, logger.NewLogger())
	require.NoError(t, err)

	return ctx
}

func TestGenerateContext(t *testing.T) {
	ctx := CreateTestContext(t, "../../examples/node-npm")
	provider := &TestProvider{}
	require.NoError(t, provider.Plan(ctx))

	// User defined config
	configJSON := `{
		"packages": {
			"node": "20.18.2",
			"go": "1.23.5",
			"python": "3.13.1"
		},
		"aptPackages": ["curl"],
		"steps": {
			"build": {
				"commands": ["echo building"]
			}
		},
		"secrets": ["RAILWAY_SECRET_1", "RAILWAY_SECRET_2"],
		"deploy": {
			"startCommand": "echo hello",
			"variables": {
				"HELLO": "world"
			}
		}
	}`

	var config config.Config
	require.NoError(t, json.Unmarshal([]byte(configJSON), &config))

	ctx.Config = &config

	buildPlan, _, err := ctx.Generate()
	require.NoError(t, err)

	buildPlanJSON, err := json.MarshalIndent(buildPlan, "", "  ")
	require.NoError(t, err)

	var actualPlan map[string]any
	require.NoError(t, json.Unmarshal(buildPlanJSON, &actualPlan))

	serializedPlan, err := json.MarshalIndent(actualPlan, "", "  ")
	require.NoError(t, err)

	snaps.MatchJSON(t, serializedPlan)
}

func TestGenerateContextAppliesConfiguredAptPackages(t *testing.T) {
	t.Run("deprecated build packages remain additive", func(t *testing.T) {
		ctx := CreateTestContext(t, "../../examples/node-npm")
		ctx.GetMiseStepBuilder().AddSupportingAptPackage("gcc")

		cfg := config.EmptyConfig()
		cfg.BuildAptPackages = []string{"curl"}
		ctx.Config = cfg

		ctx.applyConfig()

		require.Equal(t, []string{"gcc", "curl"}, ctx.GetMiseStepBuilder().SupportingAptPackages)
		require.Len(t, ctx.Logger.Logs, 2)
		require.Equal(t, logger.Deprecation, ctx.Logger.Logs[0].Level)
		require.Contains(t, ctx.Logger.Logs[0].Msg, "in the future")
		require.Equal(t, logger.Suggestion, ctx.Logger.Logs[1].Level)
		require.Contains(t, ctx.Logger.Logs[1].Msg, "Add `...` to `buildAptPackages`")
		require.Equal(t, "/guides/installing-packages", ctx.Logger.Logs[1].DocsPath)
	})

	t.Run("build packages explicitly extend generated packages", func(t *testing.T) {
		ctx := CreateTestContext(t, "../../examples/node-npm")
		ctx.GetMiseStepBuilder().AddSupportingAptPackage("gcc")

		cfg := config.EmptyConfig()
		cfg.BuildAptPackages = []string{"...", "curl"}
		ctx.Config = cfg

		ctx.applyConfig()

		require.Equal(t, []string{"gcc", "curl"}, ctx.GetMiseStepBuilder().SupportingAptPackages)
		require.Empty(t, ctx.Logger.Logs)
	})

	t.Run("deploy packages replace generated packages", func(t *testing.T) {
		ctx := CreateTestContext(t, "../../examples/node-npm")
		ctx.Deploy.AddAptPackages([]string{"libnss3"})

		cfg := config.EmptyConfig()
		cfg.Deploy.AptPackages = []string{"curl"}
		ctx.Config = cfg

		ctx.applyConfig()

		require.Equal(t, []string{"curl"}, ctx.Deploy.AptPackages)
		require.Len(t, ctx.Logger.Logs, 1)
		require.Equal(t, logger.Suggestion, ctx.Logger.Logs[0].Level)
		require.Contains(t, ctx.Logger.Logs[0].Msg, "Add `...` to `deploy.aptPackages`")
		require.Equal(t, "/guides/installing-packages", ctx.Logger.Logs[0].DocsPath)
	})

	t.Run("deploy packages explicitly extend generated packages", func(t *testing.T) {
		ctx := CreateTestContext(t, "../../examples/node-npm")
		ctx.Deploy.AddAptPackages([]string{"libnss3"})

		cfg := config.EmptyConfig()
		cfg.Deploy.AptPackages = []string{"...", "curl"}
		ctx.Config = cfg

		ctx.applyConfig()

		require.Equal(t, []string{"libnss3", "curl"}, ctx.Deploy.AptPackages)
		require.Empty(t, ctx.Logger.Logs)
	})
}

func TestGenerateContextAppliesConfiguredDeployBase(t *testing.T) {
	t.Run("direct deploy base", func(t *testing.T) {
		ctx := CreateTestContext(t, "../../examples/node-npm")
		cfg := config.EmptyConfig()
		cfg.Deploy.Base = &plan.Layer{Image: "debian:bookworm-slim"}
		ctx.Config = cfg

		buildPlan, _, err := ctx.Generate()
		require.NoError(t, err)
		require.Equal(t, plan.NewImageLayer("debian:bookworm-slim"), buildPlan.Deploy.Base)
	})

	t.Run("runtime apt step uses configured deploy base", func(t *testing.T) {
		ctx := CreateTestContext(t, "../../examples/node-npm")
		cfg := config.EmptyConfig()
		cfg.Deploy.Base = &plan.Layer{Image: "debian:bookworm-slim"}
		cfg.Deploy.AptPackages = []string{"curl"}
		ctx.Config = cfg

		buildPlan, _, err := ctx.Generate()
		require.NoError(t, err)
		require.Equal(t, plan.NewStepLayer("packages:apt:runtime"), buildPlan.Deploy.Base)

		var runtimeAptStep *plan.Step
		for i := range buildPlan.Steps {
			if buildPlan.Steps[i].Name == "packages:apt:runtime" {
				runtimeAptStep = &buildPlan.Steps[i]
				break
			}
		}

		require.NotNil(t, runtimeAptStep)
		require.Equal(t, []plan.Layer{plan.NewImageLayer("debian:bookworm-slim")}, runtimeAptStep.Inputs)
	})
}

func TestGenerateContextDeployInputs(t *testing.T) {
	t.Run("explicit inputs suppress implicit outputs from every configured step", func(t *testing.T) {
		ctx := CreateTestContext(t, "../../examples/node-npm")
		provider := &TestProvider{}
		require.NoError(t, provider.Plan(ctx))

		configJSON := `{
			"steps": {
				"install": {
					"commands": ["echo installing"]
				},
				"build": {
					"commands": ["echo building"]
				}
			},
			"deploy": {
				"inputs": [
					{
						"step": "build",
						"include": ["apps/landing/.next/standalone"]
					}
				]
			}
		}`

		var config config.Config
		require.NoError(t, json.Unmarshal([]byte(configJSON), &config))
		ctx.Config = &config

		buildPlan, _, err := ctx.Generate()
		require.NoError(t, err)

		require.Equal(t, []plan.Layer{
			plan.NewStepLayer("build", plan.NewIncludeFilter([]string{"apps/landing/.next/standalone"})),
		}, buildPlan.Deploy.Inputs)
	})

	t.Run("omitted inputs preserve implicit outputs", func(t *testing.T) {
		ctx := CreateTestContext(t, "../../examples/node-npm")
		configJSON := `{
			"steps": {
				"custom": {
					"commands": ["echo custom"]
				}
			},
			"deploy": {
				"startCommand": "echo hello"
			}
		}`

		var config config.Config
		require.NoError(t, json.Unmarshal([]byte(configJSON), &config))
		ctx.Config = &config

		buildPlan, _, err := ctx.Generate()
		require.NoError(t, err)

		require.Equal(t, []plan.Layer{
			plan.NewStepLayer("custom", plan.NewIncludeFilter([]string{"."})),
		}, buildPlan.Deploy.Inputs)
	})

	t.Run("explicit deploy outputs remain additive", func(t *testing.T) {
		ctx := CreateTestContext(t, "../../examples/node-npm")
		configJSON := `{
			"steps": {
				"build": {
					"commands": ["echo building"],
					"deployOutputs": [
						{"include": ["dist"]}
					]
				}
			},
			"deploy": {
				"inputs": [
					{"step": "build", "include": ["other"]}
				]
			}
		}`

		var config config.Config
		require.NoError(t, json.Unmarshal([]byte(configJSON), &config))
		ctx.Config = &config

		buildPlan, _, err := ctx.Generate()
		require.NoError(t, err)

		require.Equal(t, []plan.Layer{
			plan.NewStepLayer("build", plan.NewIncludeFilter([]string{"other"})),
			plan.NewStepLayer("build", plan.NewIncludeFilter([]string{"dist"})),
		}, buildPlan.Deploy.Inputs)
	})

	t.Run("empty inputs suppress implicit outputs", func(t *testing.T) {
		ctx := CreateTestContext(t, "../../examples/node-npm")
		provider := &TestProvider{}
		require.NoError(t, provider.Plan(ctx))

		configJSON := `{
			"steps": {
				"install": {
					"commands": ["echo installing"]
				},
				"build": {
					"commands": ["echo building"]
				}
			},
			"deploy": {
				"inputs": []
			}
		}`

		var config config.Config
		require.NoError(t, json.Unmarshal([]byte(configJSON), &config))
		ctx.Config = &config

		buildPlan, _, err := ctx.Generate()
		require.NoError(t, err)

		require.Empty(t, buildPlan.Deploy.Inputs)
	})

	t.Run("spread inputs preserve generated and implicit outputs", func(t *testing.T) {
		ctx := CreateTestContext(t, "../../examples/node-npm")
		provider := &TestProvider{}
		require.NoError(t, provider.Plan(ctx))

		configJSON := `{
			"steps": {
				"custom": {
					"commands": ["echo custom"]
				}
			},
			"deploy": {
				"inputs": ["..."]
			}
		}`

		var config config.Config
		require.NoError(t, json.Unmarshal([]byte(configJSON), &config))
		ctx.Config = &config

		buildPlan, _, err := ctx.Generate()
		require.NoError(t, err)

		require.Equal(t, []plan.Layer{
			plan.NewStepLayer("build"),
			plan.NewStepLayer("custom", plan.NewIncludeFilter([]string{"."})),
		}, buildPlan.Deploy.Inputs)
	})
}

func TestGenerateContextDockerignore(t *testing.T) {
	t.Run("dockerignore patterns precede config patterns", func(t *testing.T) {
		ctx := CreateTestContext(t, "../../examples/dockerignore")
		configExcludes := []string{"!the.log", "config-only"}
		ctx.Config.Exclude = configExcludes

		buildPlan, _, err := ctx.Generate()
		require.NoError(t, err)

		expected := append(slices.Clone(ctx.dockerignoreCtx.Excludes), configExcludes...)
		require.Equal(t, expected, buildPlan.Exclude)
	})

	t.Run("context with dockerignore", func(t *testing.T) {
		ctx := CreateTestContext(t, "../../examples/dockerignore")

		// Verify dockerignore was parsed during context creation
		require.NotNil(t, ctx.dockerignoreCtx)

		// Verify metadata indicates dockerignore presence
		require.Equal(t, "true", ctx.Metadata.Get("dockerIgnore"))

		// Verify dockerignore patterns are in the context
		require.NotEmpty(t, ctx.dockerignoreCtx.Excludes)
		require.Contains(t, ctx.dockerignoreCtx.Excludes, ".vscode")
		require.Contains(t, ctx.dockerignoreCtx.Excludes, "*.log")
		require.Contains(t, ctx.dockerignoreCtx.Excludes, "__pycache__")

		// Negation patterns are also in Excludes (with ! prefix)
		require.Contains(t, ctx.dockerignoreCtx.Excludes, "!negation_test/should_exist.txt")
		require.Contains(t, ctx.dockerignoreCtx.Excludes, "!negation_test/existing_folder")

		// Verify dockerignore patterns are correctly moved to the plan level
		buildPlan, _, err := ctx.Generate()
		require.NoError(t, err)
		require.Contains(t, buildPlan.Exclude, ".vscode")
		require.Contains(t, buildPlan.Exclude, "*.log")
		require.Contains(t, buildPlan.Exclude, "!negation_test/should_exist.txt")
		require.Contains(t, buildPlan.Exclude, "!negation_test/existing_folder")
	})

	t.Run("context without dockerignore", func(t *testing.T) {
		ctx := CreateTestContext(t, "../../examples/node-npm")

		// Verify dockerignore context exists but has no patterns
		require.NotNil(t, ctx.dockerignoreCtx)

		// Verify metadata does not indicate dockerignore presence
		require.Empty(t, ctx.Metadata.Get("dockerIgnore"))
	})

	t.Run("context creation with no dockerignore", func(t *testing.T) {
		// Test with a directory that exists but has no .dockerignore file
		ctx := CreateTestContext(t, "../../examples/node-npm")

		// Should succeed even without .dockerignore file
		require.NotNil(t, ctx)
		require.NotNil(t, ctx.dockerignoreCtx)

		// Verify parsing works with no file present
		require.Nil(t, ctx.dockerignoreCtx.Excludes)
	})

	t.Run("context creation fails with invalid dockerignore", func(t *testing.T) {
		// Create a temporary directory with an inaccessible .dockerignore
		tempDir, err := os.MkdirTemp("", "dockerignore-test")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tempDir) }()

		dockerignorePath := filepath.Join(tempDir, ".dockerignore")
		err = os.WriteFile(dockerignorePath, []byte("*.log\nnode_modules\n"), 0644)
		require.NoError(t, err)

		// Make the file unreadable to simulate a parsing error
		err = os.Chmod(dockerignorePath, 0000)
		require.NoError(t, err)
		defer func() { _ = os.Chmod(dockerignorePath, 0644) }()

		// Create app with the temp directory
		userApp, err := app.NewApp(tempDir)
		require.NoError(t, err)

		env := app.NewEnvironment(nil)
		config := config.EmptyConfig()

		// Context creation should fail due to dockerignore parsing error
		ctx, err := NewGenerateContext(userApp, env, config, logger.NewLogger())
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to parse .dockerignore")
		require.Nil(t, ctx)
	})
}

func TestGenerateContextProviderDefaultExcludes(t *testing.T) {
	defaultExcludes := []string{".git", "node_modules"}
	expectedWarning := logger.Msg{
		Level: logger.Warn,
		Msg: "Applying default exclude patterns: `.git`, `node_modules`. " +
			"Add a `.dockerignore` file or `exclude` patterns to `railpack.json` to define your own exclusions. " +
			"https://railpack.com/config/excluding-files",
	}
	expectedSuggestion := logger.Msg{
		Level:    logger.Suggestion,
		Msg:      "Add a `.dockerignore` file to exclude files from the build context.",
		DocsPath: "/config/excluding-files",
	}

	assertNoExcludeLogs := func(t *testing.T, logs []logger.Msg) {
		t.Helper()
		for _, msg := range logs {
			require.NotEqual(t, "/config/excluding-files", msg.DocsPath)
			require.NotContains(t, msg.Msg, "https://railpack.com/config/excluding-files")
		}
	}

	t.Run("applies provider defaults when no user excludes are present", func(t *testing.T) {
		ctx := CreateTestContext(t, t.TempDir())
		ctx.SetProviderDefaultExcludes(defaultExcludes)

		buildPlan, _, err := ctx.Generate()
		require.NoError(t, err)
		require.Equal(t, defaultExcludes, buildPlan.Exclude)
		require.Contains(t, ctx.Logger.Logs, expectedWarning)
	})

	t.Run("does not apply defaults when provider has none", func(t *testing.T) {
		ctx := CreateTestContext(t, t.TempDir())

		buildPlan, _, err := ctx.Generate()
		require.NoError(t, err)
		require.Empty(t, buildPlan.Exclude)
		require.Contains(t, ctx.Logger.Logs, expectedSuggestion)
	})

	t.Run("dockerignore overrides provider defaults", func(t *testing.T) {
		appDir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(appDir, ".dockerignore"), []byte("custom\n"), 0644))

		ctx := CreateTestContext(t, appDir)
		ctx.SetProviderDefaultExcludes(defaultExcludes)

		buildPlan, _, err := ctx.Generate()
		require.NoError(t, err)
		require.Equal(t, []string{"custom"}, buildPlan.Exclude)
		assertNoExcludeLogs(t, ctx.Logger.Logs)
	})

	t.Run("empty dockerignore overrides provider defaults", func(t *testing.T) {
		appDir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(appDir, ".dockerignore"), []byte("# intentionally empty\n"), 0644))

		ctx := CreateTestContext(t, appDir)
		ctx.SetProviderDefaultExcludes(defaultExcludes)

		buildPlan, _, err := ctx.Generate()
		require.NoError(t, err)
		require.Empty(t, buildPlan.Exclude)
		assertNoExcludeLogs(t, ctx.Logger.Logs)
	})

	t.Run("config excludes override provider defaults", func(t *testing.T) {
		ctx := CreateTestContext(t, t.TempDir())
		ctx.Config.Exclude = []string{"custom"}
		ctx.SetProviderDefaultExcludes(defaultExcludes)

		buildPlan, _, err := ctx.Generate()
		require.NoError(t, err)
		require.Equal(t, []string{"custom"}, buildPlan.Exclude)
		assertNoExcludeLogs(t, ctx.Logger.Logs)
	})

	t.Run("empty config excludes override provider defaults", func(t *testing.T) {
		ctx := CreateTestContext(t, t.TempDir())
		ctx.Config.Exclude = []string{}
		ctx.SetProviderDefaultExcludes(defaultExcludes)

		buildPlan, _, err := ctx.Generate()
		require.NoError(t, err)
		require.Empty(t, buildPlan.Exclude)
		assertNoExcludeLogs(t, ctx.Logger.Logs)
	})
}
