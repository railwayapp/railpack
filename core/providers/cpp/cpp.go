package cpp

import (
	"fmt"
	"path/filepath"

	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/plan"
)

type CppProvider struct {
	config CppConfig
}

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
	if err := ctx.Config.ProviderConfig(p.Name(), &p.config); err != nil {
		return fmt.Errorf("invalid cpp provider config: %w", err)
	}
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
	p.configureMisePackages(packages, buildsystem)

	build := ctx.NewCommandStep("build")
	build.AddInput(plan.NewStepLayer(packages.Name()))
	build.AddInput(plan.NewLocalLayer())
	build.AddCommand(plan.NewExecCommand("mkdir /build"))
	buildsystem.Build(build)

	ctx.Deploy.StartCmd = fmt.Sprintf("/build/%s", filepath.Base(ctx.GetAppSource()))
	ctx.Deploy.AddInputs([]plan.Layer{
		plan.NewStepLayer(build.Name(), plan.NewIncludeFilter([]string{"/build/"})),
	})

	return nil
}

func (p *CppProvider) configureMisePackages(packages *generate.MiseStepBuilder, buildsystem buildSystem) {
	switch buildsystem.(type) {
	case *cmake:
		if p.config.CmakeVersion != "" {
			cmake := packages.Default("cmake", "latest")
			packages.Version(cmake, p.config.CmakeVersion, "railpack config > cpp > cmakeVersion")
		}
	case *meson:
		if p.config.MesonVersion != "" {
			meson := packages.Default("meson", "latest")
			packages.Version(meson, p.config.MesonVersion, "railpack config > cpp > mesonVersion")
		}
	}

	if p.config.NinjaVersion != "" {
		ninja := packages.Default("ninja", "latest")
		packages.Version(ninja, p.config.NinjaVersion, "railpack config > cpp > ninjaVersion")
	}
}
