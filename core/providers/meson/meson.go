package meson

import (
	"fmt"
	"path/filepath"

	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/plan"
)

type MesonProvider struct{}

func (p *MesonProvider) Name() string {
	return "meson"
}

func (p *MesonProvider) Detect(ctx *generate.GenerateContext) (bool, error) {
	return ctx.App.HasFile("meson.build"), nil
}

func (p *MesonProvider) Initialize(ctx *generate.GenerateContext) error {
	return nil
}

func (p *MesonProvider) StartCommandHelp() string {
	return ""
}

func (p *MesonProvider) Plan(ctx *generate.GenerateContext) error {

	packages := ctx.GetMiseStepBuilder()

	packages.Default("meson", "latest")
	packages.Default("ninja", "latest")

	build := ctx.NewCommandStep("build")
	build.AddInput(plan.NewStepLayer(packages.Name()))
	build.AddInput(ctx.NewLocalLayer())
	build.AddCommands([]plan.Command{
		plan.NewExecCommand("mkdir /build"),
		plan.NewExecCommand("meson setup /build"),
		plan.NewExecCommand("meson compile -C /build"),
	})

	ctx.Deploy.StartCmd = fmt.Sprintf("/build/%s", filepath.Base(ctx.GetAppSource()))
	ctx.Deploy.AddInputs([]plan.Layer{
		plan.NewStepLayer(build.Name(), plan.NewIncludeFilter([]string{"/build/"})),
	})

	return nil
}
