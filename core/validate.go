package core

import (
	"fmt"
	"strings"

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

	if !validateStartCommand(plan, logger, options) {
		return false
	}

	if !validateEntrypoint(plan, logger) {
		return false
	}

	if !validateHealthcheck(plan, logger) {
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

func validateStartCommand(plan *plan.BuildPlan, logger *logger.Logger, options *ValidatePlanOptions) bool {
	if plan.Deploy.StartCmd != "" || len(plan.Deploy.Entrypoint) > 0 {
		return true
	}

	msg := "No start command detected. Specify a start command: https://railpack.com/config/file#deploy"
	if options.ProviderToUse != nil {
		if providerHelp := options.ProviderToUse.StartCommandHelp(); providerHelp != "" {
			msg += "\n\n" + providerHelp
		}
	}

	if options.ErrorMissingStartCommand {
		logger.LogError("%s", msg)
		return false
	}

	logger.LogWarn("%s", msg)
	return true
}

func validateEntrypoint(buildPlan *plan.BuildPlan, logger *logger.Logger) bool {
	if buildPlan.Deploy.Entrypoint == nil {
		return true
	}
	if len(buildPlan.Deploy.Entrypoint) == 0 {
		logger.LogError("deploy.entrypoint must include at least one argument")
		logger.LogSuggestion("See how to configure an entrypoint", "/config/file/#entrypoint")
		return false
	}
	for _, arg := range buildPlan.Deploy.Entrypoint {
		if arg == "" {
			logger.LogError("deploy.entrypoint arguments must be non-empty")
			logger.LogSuggestion("See how to configure an entrypoint", "/config/file/#entrypoint")
			return false
		}
	}
	return true
}

func validateHealthcheck(buildPlan *plan.BuildPlan, logger *logger.Logger) bool {
	healthcheck := buildPlan.Deploy.Healthcheck
	if healthcheck == nil {
		return true
	}

	_, err := healthcheck.ToImageConfig()
	if err == nil {
		return true
	}

	// Conversion reports every invalid field together instead of stopping at the first.
	if joined, ok := err.(interface{ Unwrap() []error }); ok {
		for _, item := range joined.Unwrap() {
			logger.LogError("%s", item.Error())
		}
		return false
	}

	logger.LogError("%s", err.Error())
	return false
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
		logger.LogError("`%s` inputs must be an image or step input\n\n%s", stepName, firstInput.String())
		return false
	}

	// and does not have any include or exclude
	if len(firstInput.Include) > 0 || len(firstInput.Exclude) > 0 {
		logger.LogError("the first input of `%s` cannot have any includes or excludes.\n\n%s", stepName, firstInput.String())
		return false
	}

	return true
}

func validateDeployLayers(plan *plan.BuildPlan, logger *logger.Logger) bool {
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

	var fileTree strings.Builder
	fileTree.WriteString("./\n")

	for i, dir := range dirs {
		prefix := "├── "
		if i == len(dirs)-1 && len(files) == 0 {
			prefix = "└── "
		}
		fmt.Fprintf(&fileTree, "%s%s/\n", prefix, dir)
	}

	for i, file := range files {
		prefix := "├── "
		if i == len(files)-1 {
			prefix = "└── "
		}
		fmt.Fprintf(&fileTree, "%s%s\n", prefix, file)
	}

	var errorMsg strings.Builder
	errorMsg.WriteString("Railpack could not determine how to build the app.\n\n")
	errorMsg.WriteString("The following languages are supported:\n")
	for _, provider := range providerNames {
		fmt.Fprintf(&errorMsg, "- %s\n", provider)
	}

	errorMsg.WriteString("\nThe app contents that Railpack analyzed contains:\n\n")
	errorMsg.WriteString(fileTree.String())
	errorMsg.WriteString("\n")
	errorMsg.WriteString("Check out the docs for more information: https://railpack.com")

	return errorMsg.String()
}
