package cmake

import (
	"fmt"
	"path/filepath"

	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/plan"
)

type CmakeProvider struct{}

func (p *CmakeProvider) Name() string {
	return "cmake"
}

func (p *CmakeProvider) Detect(ctx *generate.GenerateContext) (bool, error) {
	return ctx.App.HasFile("CMakeLists.txt"), nil
}

func (p *CmakeProvider) Initialize(ctx *generate.GenerateContext) error {
	return nil
}

func (p *CmakeProvider) StartCommandHelp() string {
	return ""
}

func (p *CmakeProvider) Plan(ctx *generate.GenerateContext) error {

	packages := ctx.GetMiseStepBuilder()

	packages.Default("cmake", "latest")
	packages.Default("ninja", "latest")

	build := ctx.NewCommandStep("build")
	build.AddInput(plan.NewStepLayer(packages.Name()))
	build.AddInput(ctx.NewLocalLayer())
	build.AddCommands([]plan.Command{
		plan.NewExecCommand("mkdir /build"),
		plan.NewExecCommand("cmake -B /build -GNinja /app"),
		plan.NewExecCommand("cmake --build /build"),
	})

	ctx.Deploy.StartCmd = fmt.Sprintf("/build/%s", filepath.Base(ctx.GetAppSource()))
	ctx.Deploy.AddInputs([]plan.Layer{
		plan.NewStepLayer(build.Name(), plan.NewIncludeFilter([]string{"/build/"})),
	})

	return nil
}
