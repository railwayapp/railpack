package cpp

import (
	"fmt"
	"path/filepath"

	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/plan"
)

type CppProvider struct{}

func (p *CppProvider) Name() string {
	return "cpp"
}

func (p *CppProvider) CleansePlan(buildPlan *plan.BuildPlan) {}

func (p *CppProvider) Detect(ctx *generate.GenerateContext) (bool, error) {
	_, c := p.DetectCmake(ctx)
	_, m := p.DetectMeson(ctx)
	return c || m, nil
}

func (p *CppProvider) Initialize(ctx *generate.GenerateContext) error {
	return nil
}

func (p *CppProvider) StartCommandHelp() string {
	return ""
}

func (p *CppProvider) Plan(ctx *generate.GenerateContext) error {

	packages := ctx.GetMiseStepBuilder()

	buildsystem, found := p.DetectCmake(ctx)
	if !found {
		buildsystem, _ = p.DetectMeson(ctx)
	}

	buildsystem.Install(ctx, packages)

	build := ctx.NewCommandStep("build")
	build.AddInput(plan.NewStepLayer(packages.Name()))
	build.AddInput(ctx.NewLocalLayer())
	build.AddCommand(plan.NewExecCommand("mkdir /build"))
	buildsystem.Build(build)

	ctx.Deploy.StartCmd = fmt.Sprintf("/build/%s", filepath.Base(ctx.GetAppSource()))
	ctx.Deploy.AddInputs([]plan.Layer{
		plan.NewStepLayer(build.Name(), plan.NewIncludeFilter([]string{"/build/"})),
	})

	return nil
}
