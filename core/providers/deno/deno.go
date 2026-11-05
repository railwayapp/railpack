package deno

import (
	"fmt"

	"github.com/railwayapp/railpack/core/app"
	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/plan"
)

const (
	DEFAULT_DENO_VERSION = "2"
	ROOT_CACHE           = "/root/.cache"
)

var denoDefaultExcludePatterns = []string{"node_modules"}

type DenoProvider struct {
	config        DenoConfig
	mainFile      string
	versionSource string
}

func (p *DenoProvider) Name() string {
	return "deno"
}

func (p *DenoProvider) Detect(ctx *generate.GenerateContext) (bool, error) {
	hasDenoJson := ctx.App.HasFile("deno.json") || ctx.App.HasFile("deno.jsonc")
	return hasDenoJson, nil
}

func (p *DenoProvider) Initialize(ctx *generate.GenerateContext) error {
	if err := ctx.Config.ProviderConfigWithEnv(p.Name(), ctx.Env.Variables, &p.config); err != nil {
		return fmt.Errorf("invalid deno provider config: %w", err)
	}

	p.versionSource = "railpack config > deno > version"
	var fileConfig DenoConfig
	if err := ctx.Config.ProviderConfig(p.Name(), &fileConfig); err != nil {
		return fmt.Errorf("invalid deno provider config: %w", err)
	}
	if fileConfig.Version == "" {
		if _, envName := ctx.Env.GetConfigVariable("DENO_VERSION"); envName != "" {
			p.versionSource = envName + " environment variable"
		}
	}

	p.mainFile = findMainFile(ctx.App)
	return nil
}

func (p *DenoProvider) Plan(ctx *generate.GenerateContext) error {
	ctx.SetProviderDefaultExcludes(denoDefaultExcludePatterns)

	miseStep := ctx.GetMiseStepBuilder()
	p.InstallMisePackages(ctx, miseStep)

	build := ctx.NewCommandStep("build")
	build.AddInput(plan.NewStepLayer(miseStep.Name()))
	p.Build(ctx, build)

	ctx.Deploy.AddInputs([]plan.Layer{
		miseStep.GetLayer(),
		plan.NewStepLayer(build.Name(), plan.Filter{
			Include: []string{".", ROOT_CACHE},
		}),
	})
	ctx.Deploy.StartCmd = p.GetStartCommand(ctx)

	return nil
}

func (p *DenoProvider) CleansePlan(buildPlan *plan.BuildPlan) {}

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

	build.AddInput(plan.NewLocalLayer())
	build.AddCommands([]plan.Command{
		plan.NewExecCommand(fmt.Sprintf("deno cache %s", p.mainFile)),
	})
}

func (p *DenoProvider) InstallMisePackages(ctx *generate.GenerateContext, miseStep *generate.MiseStepBuilder) {
	deno := miseStep.Default("deno", DEFAULT_DENO_VERSION)

	miseStep.UseMiseVersions(ctx, []string{"deno"})

	if p.config.Version != "" {
		miseStep.Version(deno, p.config.Version, p.versionSource)
	}
}

// selects the entrypoint for a Deno app. It prefers an explicit main.* file in the
// project root, falling back to the first .ts/.js/.mjs/.mts file found anywhere in the tree.
func findMainFile(app *app.App) string {
	if name, _, _ := app.ReadFirstFileOf("main.ts", "main.js", "main.mjs", "main.mts"); name != "" {
		return name
	}

	matches, err := app.FindFiles("**/*.{ts,js,mjs,mts}")
	if err != nil || len(matches) == 0 {
		return ""
	}
	return matches[0]
}
