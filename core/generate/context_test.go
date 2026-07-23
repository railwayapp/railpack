package generate

import (
	"encoding/json"
	"os"
	"path/filepath"
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

func TestGenerateContextDockerignore(t *testing.T) {
	t.Run("context with dockerignore", func(t *testing.T) {
		ctx := CreateTestContext(t, "../../examples/dockerignore")

		// Verify dockerignore was parsed during context creation
		require.NotNil(t, ctx.dockerignoreCtx)

		// Verify metadata indicates dockerignore presence
		require.Equal(t, "true", ctx.Metadata.Get("dockerIgnore"))

		// Test NewLocalLayer with dockerignore patterns
		layer := ctx.NewLocalLayer()
		require.True(t, layer.Local)
		require.NotNil(t, layer.Filter)

		// Should have exclude patterns from .dockerignore
		require.NotEmpty(t, layer.Exclude)
		require.Contains(t, layer.Exclude, ".vscode")
		require.Contains(t, layer.Exclude, "*.log")
		require.Contains(t, layer.Exclude, "__pycache__") // Trailing slash is stripped by parser

		// Should have default include pattern
		require.Equal(t, []string{".", "negation_test/should_exist.txt", "negation_test/existing_folder"}, layer.Include)
	})

	t.Run("context without dockerignore", func(t *testing.T) {
		ctx := CreateTestContext(t, "../../examples/node-npm")

		// Verify dockerignore context exists but has no patterns
		require.NotNil(t, ctx.dockerignoreCtx)

		// Verify metadata does not indicate dockerignore presence
		require.Empty(t, ctx.Metadata.Get("dockerIgnore"))

		// Test NewLocalLayer without dockerignore patterns
		layer := ctx.NewLocalLayer()
		require.True(t, layer.Local)

		// Should use default behavior when no dockerignore patterns exist
		require.NotNil(t, layer.Filter)
		require.Equal(t, []string{"."}, layer.Include)
		require.Empty(t, layer.Exclude)
	})

	t.Run("context creation with no dockerignore", func(t *testing.T) {
		// Test with a directory that exists but has no .dockerignore file
		ctx := CreateTestContext(t, "../../examples/node-npm")

		// Should succeed even without .dockerignore file
		require.NotNil(t, ctx)
		require.NotNil(t, ctx.dockerignoreCtx)

		// Verify parsing works with no file present
		require.Nil(t, ctx.dockerignoreCtx.Excludes)
		require.Nil(t, ctx.dockerignoreCtx.Includes)
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
