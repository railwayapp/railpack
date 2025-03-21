package deno

import (
	"fmt"

	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/plan"
)

const (
	DEFAULT_DENO_VERSION = "2"
	ROOT_CACHE           = "/root/.cache"
)

type DenoJson struct {
	Tasks map[string]string `json:"tasks"`
}

type DenoProvider struct {
	mainFile string
}

func (p *DenoProvider) Name() string {
	return "deno"
}

func (p *DenoProvider) Detect(ctx *generate.GenerateContext) (bool, error) {
	hasDenoJson := ctx.App.HasMatch("deno.json") || ctx.App.HasMatch("deno.jsonc")
	return hasDenoJson, nil
}

func (p *DenoProvider) Initialize(ctx *generate.GenerateContext) error {
	p.mainFile = p.findMainFile(ctx)
	return nil
}

func (p *DenoProvider) Plan(ctx *generate.GenerateContext) error {
	miseStep := ctx.GetMiseStepBuilder()
	p.InstallMisePackages(ctx, miseStep)

	build := ctx.NewCommandStep("build")
	build.AddInput(plan.NewStepLayer(miseStep.Name()))
	p.Build(ctx, build)

	ctx.Deploy.AddInputs([]plan.Layer{
		miseStep.GetLayer(),
		plan.NewStepLayer(build.Name(), plan.InputOptions{
			Include: []string{".", ROOT_CACHE},
		}),
	})
	ctx.Deploy.StartCmd = p.GetStartCommand(ctx)

	return nil
}

func (p *DenoProvider) StartCommandHelp() string {
	return "To start your Deno application, Railpack will look for:\n\n" +
		"1. A main.ts, main.js, main.mjs, or main.mts file in your project root\n\n" +
		"2. If no main file is found, it will use the first .ts, .js, .mjs, or .mts file found in your project\n\n" +
		"The selected file will be run with `deno run --allow-all`"
}

func (p *DenoProvider) GetStartCommand(ctx *generate.GenerateContext) string {
	if p.mainFile == "" {
		return ""
	}

	return fmt.Sprintf("deno run --allow-all %s", p.mainFile)
}

func (p *DenoProvider) Build(ctx *generate.GenerateContext, build *generate.CommandStepBuilder) {
	if p.mainFile == "" {
		return
	}

	build.AddCommands([]plan.Command{
		plan.NewCopyCommand("."),
		plan.NewExecCommand(fmt.Sprintf("deno cache %s", p.mainFile)),
	})
}

func (p *DenoProvider) InstallMisePackages(ctx *generate.GenerateContext, miseStep *generate.MiseStepBuilder) {
	deno := miseStep.Default("deno", DEFAULT_DENO_VERSION)

	if envVersion, varName := ctx.Env.GetConfigVariable("DENO_VERSION"); envVersion != "" {
		miseStep.Version(deno, envVersion, varName)
	}
}

func (p *DenoProvider) findMainFile(ctx *generate.GenerateContext) string {
	files := []string{"main.ts", "main.js", "main.mjs", "main.mts"}
	for _, file := range files {
		if ctx.App.HasMatch(file) {
			return file
		}
	}

	files, err := ctx.App.FindFiles("**/*.{ts,js,mjs,mts}")
	if err != nil {
		return ""
	}

	if len(files) == 0 {
		return ""
	}

	return files[0]
}
