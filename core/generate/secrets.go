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
func resolveSecrets(buildConfig *config.Config, steps []plan.Step) ([]string, []plan.Step) {
	configuredSecrets := concreteSecrets(buildConfig.Secrets)
	secrets := collectSecrets(configuredSecrets, steps)
	preserveWildcard := len(secrets) == 0
	resolvedSteps := slices.Clone(steps)

	for i := range resolvedSteps {
		resolvedSteps[i].Secrets = resolveStepSecrets(
			resolvedSteps[i].Secrets,
			configuredSecrets,
			preserveWildcard,
		)
	}

	return secrets, resolvedSteps
}

// Collects the concrete configured and step-level names in their original order
func collectSecrets(configuredSecrets []string, steps []plan.Step) []string {
	secrets := slices.Clone(configuredSecrets)
	for _, step := range steps {
		secrets = append(secrets, concreteSecrets(step.Secrets)...)
	}

	return utils.RemoveDuplicates(secrets)
}

// Expands wildcards unless an empty catalog requires preserving cache semantics
func resolveStepSecrets(stepSecrets, configuredSecrets []string, preserveWildcard bool) []string {
	resolvedSecrets := make([]string, 0, len(stepSecrets)+len(configuredSecrets))
	for _, secret := range stepSecrets {
		switch secret {
		case "*":
			if preserveWildcard {
				resolvedSecrets = append(resolvedSecrets, secret)
			} else {
				resolvedSecrets = append(resolvedSecrets, configuredSecrets...)
			}
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
