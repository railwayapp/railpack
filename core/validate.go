package core

import (
	"fmt"

	"github.com/railwayapp/railpack/core/app"
	"github.com/railwayapp/railpack/core/logger"
	"github.com/railwayapp/railpack/core/plan"
	"github.com/railwayapp/railpack/core/providers"
	"github.com/railwayapp/railpack/internal/utils"
)

type ValidatePlanOptions struct {
	ErrorMissingStartCommand bool
	ProviderToUse            providers.Provider
}

func ValidatePlan(plan *plan.BuildPlan, app *app.App, logger *logger.Logger, options *ValidatePlanOptions) bool {
	if !validateCommands(plan, app, logger) {
		return false
	}

	if options.ErrorMissingStartCommand && !validateStartCommand(plan, logger, options.ProviderToUse) {
		return false
	}

	for _, step := range plan.Steps {
		if !validateInputs(step.Inputs, step.Name, logger) {
			return false
		}
	}

	return validateDeployLayers(plan, logger)
}

// validateCommands checks if the plan has at least one command
func validateCommands(plan *plan.BuildPlan, app *app.App, logger *logger.Logger) bool {
	var atLeastOneCommand = false
	for _, step := range plan.Steps {
		if len(step.Commands) > 0 {
			atLeastOneCommand = true
		}
	}

	if !atLeastOneCommand {
		logger.LogError("%s", getNoProviderError(app))
		return false
	}

	return true
}

// validateStartCommand checks if the plan has a start command
func validateStartCommand(plan *plan.BuildPlan, logger *logger.Logger, provider providers.Provider) bool {
	if plan.Deploy.StartCmd == "" {
		startCmdHelp := "No start command was found."

		if provider != nil {
			if providerHelp := provider.StartCommandHelp(); providerHelp != "" {
				startCmdHelp += "\n\n" + providerHelp
			}
		}

		logger.LogError("%s", startCmdHelp)

		return false
	}

	return true
}

// validateInputs checks that
// 1. the step has at least one input
// 2. the first input is an image or step input
// 3. the first input does not have any includes or excludes
func validateInputs(inputs []plan.Layer, stepName string, logger *logger.Logger) bool {
	if len(inputs) == 0 {
		logger.LogError("step `%s` has no inputs", stepName)
		return false
	}

	// Check that the first input is an image or step input
	firstInput := inputs[0]
	if firstInput.Image == "" && firstInput.Step == "" {
		logger.LogError("%s inputs must be an image or step input", stepName)
		return false
	}

	// and does not have any include or exclude
	if len(firstInput.Include) > 0 || len(firstInput.Exclude) > 0 {
		logger.LogError("the first input of %s cannot have any includes or excludes.\n\n%s", stepName, firstInput.String())
		return false
	}

	return true
}

func validateDeployLayers(plan *plan.BuildPlan, logger *logger.Logger) bool {
	fmt.Printf("PLAN DEPLOY: %+v\n", plan.Deploy.Base)

	if plan.Deploy.Base.Image == "" && plan.Deploy.Base.Step == "" {
		logger.LogError("deploy.base is required")
		return false
	}

	return true
}

func getNoProviderError(app *app.App) string {
	providerNames := []string{}
	for _, provider := range providers.GetLanguageProviders() {
		providerNames = append(providerNames, utils.CapitalizeFirst(provider.Name()))
	}

	files, _ := app.FindFiles("*")
	dirs, _ := app.FindDirectories("*")

	fileTree := "./\n"

	for i, dir := range dirs {
		prefix := "├── "
		if i == len(dirs)-1 && len(files) == 0 {
			prefix = "└── "
		}
		fileTree += fmt.Sprintf("%s%s/\n", prefix, dir)
	}

	for i, file := range files {
		prefix := "├── "
		if i == len(files)-1 {
			prefix = "└── "
		}
		fileTree += fmt.Sprintf("%s%s\n", prefix, file)
	}

	errorMsg := "Railpack could not determine how to build the app.\n\n"
	errorMsg += "The following languages are supported:\n"
	for _, provider := range providerNames {
		errorMsg += fmt.Sprintf("- %s\n", provider)
	}

	errorMsg += "\nThe app contents that Railpack analyzed contains:\n\n"
	errorMsg += fileTree
	errorMsg += "\n"
	errorMsg += "Check out the docs for more information: https://railpack.com"

	return errorMsg
}
