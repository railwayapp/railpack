package node

import (
	_ "embed"
	"fmt"
	"path"

	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/plan"
)

const (
	DefaultCaddyfilePath = "/Caddyfile"
	OUTPUT_DIR_VAR       = "SPA_OUTPUT_DIR"
)

//go:embed Caddyfile.template
var caddyfileTemplate string

func (p *NodeProvider) isSPA(ctx *generate.GenerateContext) bool {
	if ctx.Env.IsConfigVariableTruthy("NO_SPA") {
		return false
	}

	// Setting the output dir directly will force an SPA build
	if value, _ := ctx.Env.GetConfigVariable(OUTPUT_DIR_VAR); value != "" {
		return true
	}

	// If there is a custom start command, we don't want to deploy with Caddy as an SPA
	if p.hasCustomStartCommand(ctx) {
		return false
	}

	isVite := p.isVite(ctx)
	isAstro := p.isAstroSPA(ctx)
	isCRA := p.isCRA(ctx)
	isAngular := p.isAngular(ctx)

	return (isVite || isAstro || isCRA || isAngular) && p.getOutputDirectory(ctx) != ""
}

func (p *NodeProvider) getSPAFramework(ctx *generate.GenerateContext) string {
	if !p.isSPA(ctx) {
		return ""
	}

	if p.isVite(ctx) {
		return "vite"
	} else if p.isAstro(ctx) {
		return "astro"
	} else if p.isCRA(ctx) {
		return "CRA"
	} else if p.isAngular(ctx) {
		return "Angular"
	}

	return ""
}

func (p *NodeProvider) DeploySPA(ctx *generate.GenerateContext, build *generate.CommandStepBuilder) error {
	outputDir := p.getOutputDirectory(ctx)
	spaFramework := p.getSPAFramework(ctx)

	ctx.Logger.LogInfo("Deploying as %s static site", spaFramework)
	ctx.Logger.LogInfo("Output directory: %s", outputDir)

	data := map[string]interface{}{
		"DIST_DIR": path.Join("/app", outputDir),
	}

	caddyfileTemplate, err := ctx.TemplateFiles([]string{"Caddyfile.template", "Caddyfile"}, caddyfileTemplate, data)
	if err != nil {
		return err
	}

	if caddyfileTemplate.Filename != "" {
		ctx.Logger.LogInfo("Using custom Caddyfile: %s", caddyfileTemplate.Filename)
	}

	installCaddyStep := ctx.NewInstallBinStepBuilder("packages:caddy")
	installCaddyStep.Default("caddy", "latest")

	caddy := ctx.NewCommandStep("caddy")
	caddy.AddInput(plan.NewStepLayer(installCaddyStep.Name()))
	caddy.AddCommands([]plan.Command{
		plan.NewFileCommand(DefaultCaddyfilePath, "Caddyfile"),
		plan.NewExecCommand(fmt.Sprintf("caddy fmt --overwrite %s", DefaultCaddyfilePath)),
	})
	caddy.Assets = map[string]string{
		"Caddyfile": caddyfileTemplate.Contents,
	}

	ctx.Deploy.StartCmd = fmt.Sprintf("caddy run --config %s --adapter caddyfile 2>&1", DefaultCaddyfilePath)

	ctx.Deploy.AddInputs([]plan.Layer{
		installCaddyStep.GetLayer(),
		plan.NewStepLayer(caddy.Name(), plan.InputOptions{
			Include: []string{DefaultCaddyfilePath},
		}),
		plan.NewStepLayer(build.Name(), plan.InputOptions{
			Include: []string{outputDir},
		}),
	})

	// ctx.Deploy.Inputs = []plan.Layer{
	// 	ctx.DefaultRuntimeInput(),
	// 	plan.NewStepLayer(installCaddyStep.Name(), plan.InputOptions{
	// 		Include: installCaddyStep.GetOutputPaths(),
	// 	}),
	// 	plan.NewStepLayer(caddy.Name(), plan.InputOptions{
	// 		Include: []string{DefaultCaddyfilePath},
	// 	}),
	// 	plan.NewStepLayer(build.Name(), plan.InputOptions{
	// 		Include: []string{outputDir},
	// 	}),
	// }

	return nil
}

func (p *NodeProvider) getOutputDirectory(ctx *generate.GenerateContext) string {
	outputDir := ""

	if dir, _ := ctx.Env.GetConfigVariable(OUTPUT_DIR_VAR); dir != "" {
		outputDir = dir
	} else if p.isVite(ctx) {
		outputDir = p.getViteOutputDirectory(ctx)
	} else if p.isAstroSPA(ctx) {
		outputDir = p.getAstroOutputDirectory(ctx)
	} else if p.isCRA(ctx) {
		outputDir = p.getCRAOutputDirectory(ctx)
	} else if p.isAngular(ctx) {
		outputDir = p.getAngularOutputDirectory(ctx)
	}

	return outputDir
}

func (p *NodeProvider) hasCustomStartCommand(ctx *generate.GenerateContext) bool {
	startCommand := ctx.Config.Deploy.StartCmd
	if startCommand == "" {
		startCommand = p.packageJson.Scripts["start"]
	}
	isAngularDefaultStartCommand := startCommand == DefaultAngularStartCommand
	isCRAStartCommand := startCommand == DefaultCRAStartCommand
	return startCommand != "" && !isAngularDefaultStartCommand && !isCRAStartCommand
}
