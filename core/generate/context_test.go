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

	t.Run("runtime bootstrap step uses configured deploy base", func(t *testing.T) {
		ctx := CreateTestContext(t, "../../examples/node-npm")
		cfg := config.EmptyConfig()
		cfg.Deploy.Base = &plan.Layer{Image: "debian:bookworm-slim"}
		cfg.Deploy.AptPackages = []string{"curl"}
		ctx.Config = cfg

		buildPlan, _, err := ctx.Generate()
		require.NoError(t, err)
		require.Equal(t, plan.NewStepLayer(MiseBootstrapRuntimeStepName), buildPlan.Deploy.Base)

		var runtimeBootstrapStep *plan.Step
		for i := range buildPlan.Steps {
			if buildPlan.Steps[i].Name == MiseBootstrapRuntimeStepName {
				runtimeBootstrapStep = &buildPlan.Steps[i]
				break
			}
		}

		require.NotNil(t, runtimeBootstrapStep)
		require.Equal(
			t,
			[]plan.Layer{plan.NewImageLayer("debian:bookworm-slim")},
			runtimeBootstrapStep.Inputs,
		)
	})
}

func TestGenerateContextAppliesProjectMiseBootstrapConfig(t *testing.T) {
	ctx := CreateTestContext(t, "../../examples/config-file")
	provider := &TestProvider{}
	require.NoError(t, provider.Plan(ctx))

	buildPlan, _, err := ctx.Generate()
	require.NoError(t, err)

	var buildBootstrapStep *plan.Step
	var runtimeBootstrapStep *plan.Step
	for i := range buildPlan.Steps {
		switch buildPlan.Steps[i].Name {
		case MiseBootstrapBuildStepName:
			buildBootstrapStep = &buildPlan.Steps[i]
		case MiseBootstrapRuntimeStepName:
			runtimeBootstrapStep = &buildPlan.Steps[i]
		}
	}

	require.NotNil(t, buildBootstrapStep)
	require.NotNil(t, runtimeBootstrapStep)
	require.Equal(t, []string{"*"}, buildBootstrapStep.Secrets)
	require.Empty(t, runtimeBootstrapStep.Secrets)
	require.Contains(
		t,
		runtimeBootstrapStep.Inputs,
		miseBootstrapRepositoryLayer(),
	)
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
