package core

import (
	"testing"

	"github.com/railwayapp/railpack/core/app"
	"github.com/railwayapp/railpack/core/logger"
	"github.com/railwayapp/railpack/core/plan"
	"github.com/railwayapp/railpack/core/providers"
	"github.com/stretchr/testify/require"
)

type mockProvider struct {
	providers.Provider
	startCommandHelp string
}

func (m *mockProvider) StartCommandHelp() string {
	return m.startCommandHelp
}

func TestValidatePlan(t *testing.T) {
	logger := logger.NewLogger()
	testApp, _ := app.NewApp(".")
	mockProvider := &mockProvider{startCommandHelp: "Add a start command"}

	t.Run("valid plan", func(t *testing.T) {
		buildPlan := plan.NewBuildPlan()
		buildStep := plan.NewStep("build")
		buildStep.Commands = []plan.Command{plan.NewExecShellCommand("npm build")}
		buildStep.Inputs = []plan.Layer{plan.NewImageLayer("node:18")}
		buildPlan.Steps = append(buildPlan.Steps, *buildStep)
		buildPlan.Deploy = plan.Deploy{
			Base:     plan.NewImageLayer("node:18"),
			StartCmd: "npm start",
			Inputs:   []plan.Layer{plan.NewStepLayer("build", plan.Filter{Include: []string{"."}})},
		}

		options := &ValidatePlanOptions{
			ErrorMissingStartCommand: true,
			ProviderToUse:            mockProvider,
		}
		require.True(t, ValidatePlan(buildPlan, testApp, logger, options))
	})
}

func TestValidateCommands(t *testing.T) {
	logger := logger.NewLogger()
	testApp, _ := app.NewApp(".")

	t.Run("plan with commands", func(t *testing.T) {
		buildPlan := plan.NewBuildPlan()
		buildStep := plan.NewStep("build")
		buildStep.Commands = []plan.Command{plan.NewExecShellCommand("npm build")}
		buildPlan.Steps = append(buildPlan.Steps, *buildStep)
		require.True(t, validateCommands(buildPlan, testApp, logger))
	})

	t.Run("plan without commands", func(t *testing.T) {
		buildPlan := plan.NewBuildPlan()
		buildStep := plan.NewStep("build")
		buildPlan.Steps = append(buildPlan.Steps, *buildStep)
		require.False(t, validateCommands(buildPlan, testApp, logger))
	})
}

func TestValidateStartCommand(t *testing.T) {
	mockProvider := &mockProvider{startCommandHelp: "Add a start command"}

	t.Run("with start command", func(t *testing.T) {
		logger := logger.NewLogger()
		buildPlan := plan.NewBuildPlan()
		buildPlan.Deploy = plan.Deploy{
			StartCmd: "npm start",
		}
		options := &ValidatePlanOptions{
			ErrorMissingStartCommand: true,
			ProviderToUse:            mockProvider,
		}
		require.True(t, validateStartCommand(buildPlan, logger, options))
		require.Equal(t, 0, len(logger.Logs))
	})

	t.Run("without start command (error)", func(t *testing.T) {
		loggerInst := logger.NewLogger()
		buildPlan := plan.NewBuildPlan()
		options := &ValidatePlanOptions{
			ErrorMissingStartCommand: true,
			ProviderToUse:            mockProvider,
		}
		require.False(t, validateStartCommand(buildPlan, loggerInst, options))
		require.Equal(t, 1, len(loggerInst.Logs))
		require.Equal(t, logger.Error, loggerInst.Logs[0].Level)
		require.Contains(t, loggerInst.Logs[0].Msg, "No start command detected")
		require.Contains(t, loggerInst.Logs[0].Msg, "Add a start command")
	})

	t.Run("without start command (warning)", func(t *testing.T) {
		loggerInst := logger.NewLogger()
		buildPlan := plan.NewBuildPlan()
		options := &ValidatePlanOptions{
			ErrorMissingStartCommand: false,
			ProviderToUse:            mockProvider,
		}
		require.True(t, validateStartCommand(buildPlan, loggerInst, options))
		require.Equal(t, 1, len(loggerInst.Logs))
		require.Equal(t, logger.Warn, loggerInst.Logs[0].Level)
		require.Contains(t, loggerInst.Logs[0].Msg, "No start command detected")
		require.Contains(t, loggerInst.Logs[0].Msg, "Add a start command")
	})
}

func TestValidateInputs(t *testing.T) {
	logger := logger.NewLogger()

	t.Run("valid inputs", func(t *testing.T) {
		inputs := []plan.Layer{
			plan.NewImageLayer("node:18"),
			plan.NewStepLayer("build", plan.Filter{Include: []string{"src"}}),
		}
		require.True(t, validateInputs(inputs, "test", logger))
	})

	t.Run("no inputs", func(t *testing.T) {
		inputs := []plan.Layer{}
		require.False(t, validateInputs(inputs, "test", logger))
	})

	t.Run("invalid first input", func(t *testing.T) {
		inputs := []plan.Layer{
			plan.NewLocalLayer(),
		}
		require.False(t, validateInputs(inputs, "test", logger))
	})

	t.Run("first input with includes", func(t *testing.T) {
		inputs := []plan.Layer{
			plan.NewImageLayer("node:18", plan.Filter{Include: []string{"src"}}),
		}
		require.False(t, validateInputs(inputs, "test", logger))
	})
}
