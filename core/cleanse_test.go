package core

import (
	"reflect"
	"testing"

	"github.com/railwayapp/railpack/core/logger"
	"github.com/railwayapp/railpack/core/plan"
	"github.com/railwayapp/railpack/core/providers/node"
)

func newTestLogger() *logger.Logger { return logger.NewLogger() }

// helper to create a basic build plan with a node_modules cache (when withCache true)
func buildPlan(withCache bool) *plan.BuildPlan {
	p := plan.NewBuildPlan()
	if withCache {
		p.Caches["node_modules"] = &plan.Cache{Directory: node.NODE_MODULES_CACHE, Type: plan.CacheTypeShared}
	}
	return p
}

func TestCleanse_CachePresent_StepDoesNotRemoveNodeModules(t *testing.T) {
	p := buildPlan(true)
	step := plan.Step{Name: "build", Caches: []string{"node_modules"}}
	step.Commands = []plan.Command{plan.NewExecShellCommand("echo 'nothing to see'")}
	p.Steps = append(p.Steps, step)

	cleansePlanStructure(p, newTestLogger())

	// should remain mounted
	if !reflect.DeepEqual(p.Steps[0].Caches, []string{"node_modules"}) {
		t.Fatalf("expected cache to remain since step doesn't remove node_modules, got %#v", p.Steps[0].Caches)
	}
}

func TestCleanse_CachePresent_StepRemovesNodeModules(t *testing.T) {
	p := buildPlan(true)
	step := plan.Step{Name: "build", Caches: []string{"node_modules"}}
	step.Commands = []plan.Command{plan.NewExecShellCommand("rm -rf node_modules && echo done")}
	p.Steps = append(p.Steps, step)

	cleansePlanStructure(p, newTestLogger())

	if len(p.Steps[0].Caches) != 0 { // should be removed (allow nil or empty)
		t.Fatalf("expected cache to be removed (nil or empty), got %#v", p.Steps[0].Caches)
	}
}

func TestCleanse_InstallStepAlwaysKeepsCache(t *testing.T) {
	p := buildPlan(true)
	install := plan.Step{Name: "install", Caches: []string{"node_modules"}}
	install.Commands = []plan.Command{plan.NewExecShellCommand("npm ci")}
	p.Steps = append(p.Steps, install)

	cleansePlanStructure(p, newTestLogger())

	if !reflect.DeepEqual(p.Steps[0].Caches, []string{"node_modules"}) { // should remain even though npm ci matches removal heuristic
		t.Fatalf("expected install step cache to remain, got %#v", p.Steps[0].Caches)
	}
}
