package gleam

import (
	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/plan"
)

type GleamProvider struct{}

func (p *GleamProvider) Name() string {
	return "gleam"
}

func (p *GleamProvider) Detect(ctx *generate.GenerateContext) (bool, error) {
	return ctx.App.HasFile("gleam.toml"), nil
}

func (p *GleamProvider) Initialize(ctx *generate.GenerateContext) error {
	return nil
}

func (p *GleamProvider) StartCommandHelp() string {
	return ""
}

func (p *GleamProvider) Plan(ctx *generate.GenerateContext) error {
	build := ctx.NewCommandStep("build")
	build.AddInput(plan.NewStepLayer(ctx.GetMiseStepBuilder().Name()))
	build.AddInput(ctx.NewLocalLayer())
	build.AddCommand(plan.NewExecCommand("gleam export erlang-shipment"))

	p.installErlang(ctx.GetMiseStepBuilder())
	ctx.GetMiseStepBuilder().Default("gleam", "latest")

	runtimeMiseStep := ctx.NewMiseStepBuilder("packages:mise:runtime")
	p.installErlang(runtimeMiseStep)

	outPath := "build/erlang-shipment/."
	ctx.Deploy.AddInputs([]plan.Layer{
		runtimeMiseStep.GetLayer(),
		plan.NewStepLayer(build.Name(), plan.Filter{
			Include: []string{outPath},
		}),
	})

	ctx.Deploy.StartCmd = "./build/erlang-shipment/entrypoint.sh run"

	return nil
}

func (p *GleamProvider) installErlang(step *generate.MiseStepBuilder) {
	step.Default("erlang", "latest")
}
