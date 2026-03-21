package gleam

import (
	"strings"

	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/plan"
	gleamconfig "github.com/railwayapp/railpack/core/providers/gleam/config"
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

func (p *GleamProvider) CleansePlan(buildPlan *plan.BuildPlan) {}

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
	ctx.GetMiseStepBuilder().UseMiseVersions(ctx, []string{"gleam", "erlang"})

	runtimeMiseStep := ctx.NewMiseStepBuilder("packages:mise:runtime")
	p.installErlang(runtimeMiseStep)
	runtimeMiseStep.UseMiseVersions(ctx, []string{"erlang"})

	outPath := "build/erlang-shipment/."

	includedPath := outPath
	if p.includeSource(ctx) {
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

func (p *GleamProvider) providerConfig(ctx *generate.GenerateContext) *gleamconfig.GleamConfig {
	if ctx.Config == nil {
		return nil
	}

	return ctx.Config.Gleam
}

func (p *GleamProvider) includeSource(ctx *generate.GenerateContext) bool {
	if envValue, _ := ctx.Env.GetConfigVariable("GLEAM_INCLUDE_SOURCE"); envValue != "" {
		envValue = strings.ToLower(envValue)
		return envValue == "1" || envValue == "true"
	}

	providerConfig := p.providerConfig(ctx)
	return providerConfig != nil && providerConfig.IncludeSource
}
