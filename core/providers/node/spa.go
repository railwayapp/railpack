package node

import (
	_ "embed"
	"fmt"
	"path"

	"github.com/charmbracelet/log"
	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/plan"
	"github.com/railwayapp/railpack/core/providers/staticfile"
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

	// Setting the output dir directly via a ENV will force an SPA build regardless of framework detection
	if value, _ := ctx.Env.GetConfigVariable(OUTPUT_DIR_VAR); value != "" {
		return true
	}

	// If there is a custom start command, we don't want to deploy with Caddy as an SPA
	if p.hasCustomStartCommand(ctx) {
		// it's easy for a user to trip over this wire and not understand that it would impact SPA deployment since using the start script
		// is somewhat if a railpack-convention, so let's make it clear to them.
		ctx.Logger.LogInfo("Custom start command detected, skipping Caddy start")
		return false
	}

	if p.isReactRouter(ctx) {
		return p.isReactRouterSPA(ctx) && p.getOutputDirectory(ctx) != ""
	}

	isVite := p.isVite(ctx)
	isAstro := p.isAstroSPA(ctx)
	isNext := p.isNextSPA(ctx)
	isCRA := p.isCRA(ctx)
	isAngular := p.isAngular(ctx)
	isExpoSPA := p.isExpoSPA(ctx)

	return (isVite || isAstro || isNext || isCRA || isAngular || isExpoSPA) && p.getOutputDirectory(ctx) != ""
}

// returns the canonical lowercase SPA framework name, or "" when none is detected.
func (p *NodeProvider) getSPAName(ctx *generate.GenerateContext) string {
	if p.isReactRouter(ctx) {
		if p.isReactRouterSPA(ctx) {
			return "react-router"
		}
		return ""
	}
	if p.isVite(ctx) {
		return "vite"
	}
	if p.isAstro(ctx) {
		return "astro"
	}
	if p.isNextSPA(ctx) {
		return "next"
	}
	if p.isCRA(ctx) {
		return "cra"
	}
	if p.isAngular(ctx) {
		return "angular"
	}
	if p.isExpoSPA(ctx) {
		return "expo"
	}

	// This can happen when the output directory environment variable forces SPA mode.
	log.Infof("No SPA framework detected")
	return ""
}

func (p *NodeProvider) DeploySPA(ctx *generate.GenerateContext, build *generate.CommandStepBuilder) error {
	outputDir := p.getOutputDirectory(ctx)
	spaFramework := p.getSPAName(ctx)

	ctx.Logger.LogInfo("Deploying as %s static site", spaFramework)
	ctx.Logger.LogInfo("Output directory: %s", outputDir)

	// default all paths to use the root index.html by default on SPA apps, but allow the user to override
	indexFallback := true
	if indexFallbackConfig := staticfile.GetIndexFallback(ctx); indexFallbackConfig != nil {
		indexFallback = *indexFallbackConfig
	}

	data := map[string]any{
		"DIST_DIR":      path.Join("/app", outputDir),
		"IndexFallback": indexFallback,
	}

	// TODO this template stuff is a bit odd: I don't see the use case for passing these specific variables to the template.
	//      if the user is customizing the Caddyfile, they can just hardcode what they want into the Caddyfile?
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
		plan.NewStepLayer(caddy.Name(), plan.Filter{
			Include: []string{DefaultCaddyfilePath},
		}),
		plan.NewStepLayer(build.Name(), plan.Filter{
			Include: []string{outputDir},
		}),
	})

	return nil
}

func (p *NodeProvider) getOutputDirectory(ctx *generate.GenerateContext) string {
	outputDir := ""

	if dir, _ := ctx.Env.GetConfigVariable(OUTPUT_DIR_VAR); dir != "" {
		outputDir = dir
	} else if p.isReactRouter(ctx) {
		if p.isReactRouterSPA(ctx) {
			outputDir = p.getReactRouterOutputDirectory(ctx)
		}
	} else if p.isVite(ctx) {
		outputDir = p.getViteOutputDirectory(ctx)
	} else if p.isAstroSPA(ctx) {
		outputDir = p.getAstroOutputDirectory(ctx)
	} else if p.isNextSPA(ctx) {
		outputDir = p.getNextOutputDirectory(ctx)
	} else if p.isCRA(ctx) {
		outputDir = p.getCRAOutputDirectory(ctx)
	} else if p.isAngular(ctx) {
		outputDir = p.getAngularOutputDirectory(ctx)
	} else if p.isExpoSPA(ctx) {
		outputDir = p.getExpoOutputDirectory(ctx)
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
	isExpoStartCommand := startCommand == DefaultExpoStartCommand
	isNextStartCommand := startCommand == DefaultNextStartCommand
	isReactRouterStartCommand := startCommand == DefaultReactRouterStartCommand

	return startCommand != "" && !isAngularDefaultStartCommand && !isCRAStartCommand && !isExpoStartCommand && !isNextStartCommand && !isReactRouterStartCommand
}
