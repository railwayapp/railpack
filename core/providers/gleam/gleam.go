package gleam

import (
	"fmt"

	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/plan"
)

type GleamProvider struct {
	config        GleamConfig
	includeSource bool
}

func (p *GleamProvider) Name() string {
	return "gleam"
}

func (p *GleamProvider) Detect(ctx *generate.GenerateContext) (bool, error) {
	return ctx.App.HasFile("gleam.toml"), nil
}

func (p *GleamProvider) Initialize(ctx *generate.GenerateContext) error {
	if err := ctx.Config.ProviderConfig(p.Name(), &p.config); err != nil {
		return fmt.Errorf("invalid gleam provider config: %w", err)
	}
	p.includeSource = ctx.Env.IsConfigVariableTruthy("GLEAM_INCLUDE_SOURCE")
	if p.config.IncludeSource != nil {
		p.includeSource = *p.config.IncludeSource
	}
	return nil
}

func (p *GleamProvider) CleansePlan(buildPlan *plan.BuildPlan) {}

func (p *GleamProvider) StartCommandHelp() string {
	return ""
}

func (p *GleamProvider) Plan(ctx *generate.GenerateContext) error {
	build := ctx.NewCommandStep("build")
	build.AddInput(plan.NewStepLayer(ctx.GetMiseStepBuilder().Name()))
	build.AddInput(plan.NewLocalLayer())
	build.AddCommand(plan.NewExecCommand("gleam export erlang-shipment"))

	p.installErlang(ctx.GetMiseStepBuilder())
	ctx.GetMiseStepBuilder().Default("gleam", "latest")
	ctx.GetMiseStepBuilder().UseMiseVersions(ctx, []string{"gleam", "erlang"})

	runtimeMiseStep := ctx.NewMiseStepBuilder("packages:mise:runtime")
	p.installErlang(runtimeMiseStep)
	runtimeMiseStep.UseMiseVersions(ctx, []string{"erlang"})

	outPath := "build/erlang-shipment/."

	includedPath := outPath
	if p.includeSource {
		includedPath = "."
	}

	ctx.Deploy.AddInputs([]plan.Layer{
		runtimeMiseStep.GetLayer(),
		plan.NewStepLayer(build.Name(), plan.Filter{
			Include: []string{includedPath},
		}),
	})

	ctx.Deploy.StartCmd = "./build/erlang-shipment/entrypoint.sh run"

	return nil
}

func (p *GleamProvider) installErlang(step *generate.MiseStepBuilder) {
	step.Default("erlang", "latest")
}
