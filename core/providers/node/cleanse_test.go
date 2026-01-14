package node

import (
	"reflect"
	"testing"

	"github.com/railwayapp/railpack/core/plan"
)

func buildPlan(withCache bool) *plan.BuildPlan {
	p := plan.NewBuildPlan()
	if withCache {
		p.Caches["node_modules"] = &plan.Cache{Directory: NODE_MODULES_CACHE, Type: plan.CacheTypeShared}
	}
	return p
}

func TestCleanse_CachePresent_StepDoesNotRemoveNodeModules(t *testing.T) {
	p := buildPlan(true)
	step := plan.Step{Name: "build", Caches: []string{"node_modules"}}
	step.Commands = []plan.Command{plan.NewExecShellCommand("echo 'nothing to see'")}
	p.Steps = append(p.Steps, step)

	provider := &NodeProvider{}
	provider.CleansePlan(p)

	if !reflect.DeepEqual(p.Steps[0].Caches, []string{"node_modules"}) {
		t.Fatalf("expected cache to remain since step doesn't remove node_modules, got %#v", p.Steps[0].Caches)
	}
}

func TestCleanse_CachePresent_StepRemovesNodeModules(t *testing.T) {
	p := buildPlan(true)
	step := plan.Step{Name: "build", Caches: []string{"node_modules"}}
	step.Commands = []plan.Command{plan.NewExecShellCommand("rm -rf node_modules && echo done")}
	p.Steps = append(p.Steps, step)

	provider := &NodeProvider{}
	provider.CleansePlan(p)

	if len(p.Steps[0].Caches) != 0 {
		t.Fatalf("expected cache to be removed (nil or empty), got %#v", p.Steps[0].Caches)
	}
}

func TestCleanse_InstallStepAlwaysKeepsCache(t *testing.T) {
	p := buildPlan(true)
	install := plan.Step{Name: "install", Caches: []string{"node_modules"}}
	install.Commands = []plan.Command{plan.NewExecShellCommand("npm ci")}
	p.Steps = append(p.Steps, install)

	provider := &NodeProvider{}
	provider.CleansePlan(p)

	if !reflect.DeepEqual(p.Steps[0].Caches, []string{"node_modules"}) {
		t.Fatalf("expected install step cache to remain, got %#v", p.Steps[0].Caches)
	}
}
