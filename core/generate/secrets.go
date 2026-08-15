package generate

import (
	"slices"

	"github.com/railwayapp/railpack/core/config"
	"github.com/railwayapp/railpack/core/plan"
	"github.com/railwayapp/railpack/internal/utils"
)

// Secret resolution from config declarations into build plans belongs in this file
// Concrete secrets are named secrets rather than selectors such as "*" and "..."
// The wildcard catalog contains concrete top-level config names, while the plan catalog
// also contains concrete names inferred from individual steps

// Resolves step selectors and builds the catalog of configured and inferred secrets
func resolveSecrets(buildConfig *config.Config, buildPlan *plan.BuildPlan) {
	configuredSecrets := concreteSecrets(buildConfig.Secrets)
	buildPlan.Secrets = collectSecrets(configuredSecrets, buildPlan.Steps)

	for i := range buildPlan.Steps {
		buildPlan.Steps[i].Secrets = resolveStepSecrets(
			buildPlan.Steps[i].Secrets,
			configuredSecrets,
		)
	}
}

// Collects the concrete configured and step-level names in their original order
func collectSecrets(configuredSecrets []string, steps []plan.Step) []string {
	secrets := slices.Clone(configuredSecrets)
	for _, step := range steps {
		secrets = append(secrets, concreteSecrets(step.Secrets)...)
	}

	return utils.RemoveDuplicates(secrets)
}

// Expands wildcards from configured names without leaking secrets inferred from other steps
func resolveStepSecrets(stepSecrets, configuredSecrets []string) []string {
	resolvedSecrets := make([]string, 0, len(stepSecrets)+len(configuredSecrets))
	for _, secret := range stepSecrets {
		switch secret {
		case "*":
			resolvedSecrets = append(resolvedSecrets, configuredSecrets...)
		case "...":
			continue
		default:
			resolvedSecrets = append(resolvedSecrets, secret)
		}
	}

	return utils.RemoveDuplicates(resolvedSecrets)
}

// Removes access selectors that are not concrete secret names
func concreteSecrets(secrets []string) []string {
	return slices.DeleteFunc(slices.Clone(secrets), func(secret string) bool {
		return secret == "*" || secret == "..."
	})
}
