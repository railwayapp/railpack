package core

import (
	"regexp"

	"github.com/railwayapp/railpack/core/logger"
	"github.com/railwayapp/railpack/core/plan"
	"github.com/railwayapp/railpack/core/providers/node"
)

// Regexes for matching commands that intentionally remove node_modules or perform
// clean installs (which implicitly delete the directory) so we can avoid mounting
// the node_modules cache in those steps.
var (
	// Matches "npm ci" with flexible whitespace, using word boundaries
	npmCiCommandRegex = regexp.MustCompile(`(?i)\bnpm\s+ci\b`)

	// Matches common delete commands targeting node_modules
	removeNodeModulesRegex = regexp.MustCompile(`(?i)\b(?:rm\s+-r[f]?|rmdir|rimraf)\s+(?:\S*\/)?node_modules\b`)
)

// willRemoveNodeModules determines if any command in the provided slice removes
// the node_modules directory either directly (rm/rimraf) or indirectly (npm ci).
// this is brittle & imperfect: https://github.com/railwayapp/railpack/pull/259
func willRemoveNodeModules(commands []plan.Command) bool {
	for _, cmd := range commands {
		if execCmd, ok := cmd.(plan.ExecCommand); ok {
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
		if cacheDef.Directory == node.NODE_MODULES_CACHE {
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

		// It's important that we do not result in an array with a zeroed string, which is why we are using this ugly loop
		var newCaches []string
		for _, name := range step.Caches {
			if name != "" && name != nodeModulesCacheKey {
				newCaches = append(newCaches, name)
			}
		}

		buildPlan.Steps[i].Caches = newCaches
	}
}
