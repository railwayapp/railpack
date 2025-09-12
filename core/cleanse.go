package core

// This file contains logic related to post-generation cleansing / normalization
// of build plans. Keeping this separate from `core.go` keeps provider selection
// and config merge logic focused while allowing targeted tests for plan
// mutation behavior.

import (
	"regexp"

	"github.com/charmbracelet/log"
	"github.com/railwayapp/railpack/core/logger"
	"github.com/railwayapp/railpack/core/plan"
)

// Regexes for matching commands that intentionally remove node_modules or perform
// clean installs (which implicitly delete the directory) so we can avoid mounting
// the node_modules cache in those steps.
var (
	npmCiCommandRegex      = regexp.MustCompile(`.*npm\s+ci\b.*`)
	removeNodeModulesRegex = regexp.MustCompile(`(^|\s)(rm\s+-rf\s+|rimraf\s+)(\./)?node_modules(\s|;|&|$)`)
)

const (
	// NODE_MODULES_CACHE is the path inside the build container we treat as the
	// node modules cache. Steps which delete node_modules should not mount this
	// cache so the deletion stays local to the layer they're producing and does
	// not wipe the shared cache directory.
	NODE_MODULES_CACHE = "/app/node_modules/.cache"
)

// willRemoveNodeModules determines if any command in the provided slice removes
// the node_modules directory either directly (rm/rimraf) or indirectly (npm ci).
func willRemoveNodeModules(commands []plan.Command) bool {
	for _, cmd := range commands {
		if execCmd, ok := cmd.(plan.ExecCommand); ok {
			log.Debugf("Inspecting build command: %s", execCmd.Cmd)
			if npmCiCommandRegex.MatchString(execCmd.Cmd) || removeNodeModulesRegex.MatchString(execCmd.Cmd) {
				return true
			}
		}
	}
	return false
}

// cleansePlanStructure applies mutations to the build plan structure after it
// is generated but before validation / serialization. Today this focuses on
// detaching the node_modules cache from steps that explicitly remove
// node_modules so the global cache isn't invalidated unintentionally.
func cleansePlanStructure(buildPlan *plan.BuildPlan, logger *logger.Logger) {
	// let's get the cache key name that has a Directory of NODE_MODULES_CACHE
	var nodeModulesCacheKey string
	for cacheName, cacheDef := range buildPlan.Caches {
		if cacheDef.Directory == NODE_MODULES_CACHE {
			nodeModulesCacheKey = cacheName
			break
		}
	}

	if nodeModulesCacheKey == "" {
		// no node_modules cache defined, nothing to do
		return
	}

	// Only detach the node modules cache from steps that remove node_modules themselves.
	// Keep the global cache definition so earlier steps (like install) can still mount it.
	for i, step := range buildPlan.Steps {
		if step.Name == "install" || !willRemoveNodeModules(step.Commands) {
			continue
		}

		before := len(step.Caches)
		if before == 0 {
			continue
		}

		var newCaches []string
		for _, name := range step.Caches {
			if name != "" && name != nodeModulesCacheKey {
				newCaches = append(newCaches, name)
			}
		}

		buildPlan.Steps[i].Caches = newCaches
	}
}
