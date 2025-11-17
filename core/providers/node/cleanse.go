// there are certain commands which `rm -rf node_modules`, which will cause a cache checksum docker error.
// this cleanse operation imperfectly detects this case and removes node_modules from the cache keys to avoid this common user error.
// more info here: https://github.com/railwayapp/railpack/pull/259

package node

import (
	"regexp"

	"github.com/railwayapp/railpack/core/plan"
)

var (
	npmCiCommandRegex      = regexp.MustCompile(`(?i)\bnpm\s+ci\b`)
	removeNodeModulesRegex = regexp.MustCompile(`(?i)\b(?:rm\s+-r[f]?|rmdir|rimraf)\s+(?:\S*\/)?node_modules\b`)
)

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

func (p *NodeProvider) CleansePlan(buildPlan *plan.BuildPlan) {
	var nodeModulesCacheKey string
	for cacheName, cacheDef := range buildPlan.Caches {
		if cacheDef.Directory == NODE_MODULES_CACHE {
			nodeModulesCacheKey = cacheName
			break
		}
	}

	if nodeModulesCacheKey == "" {
		return
	}

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
