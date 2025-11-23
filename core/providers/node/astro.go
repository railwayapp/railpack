package node

import (
	"regexp"
	"strings"

	"github.com/railwayapp/railpack/core/generate"
)

const (
	DefaultAstroOutputDirectory = "dist"
)

func (p *NodeProvider) isAstroPackage(pkg *WorkspacePackage, ctx *generate.GenerateContext) bool {
	astroConfigMjs := "astro.config.mjs"
	astroConfigTs := "astro.config.ts"
	if pkg.Path != "" {
		astroConfigMjs = pkg.Path + "/astro.config.mjs"
		astroConfigTs = pkg.Path + "/astro.config.ts"
	}

	hasAstroConfig := ctx.App.HasFile(astroConfigMjs) || ctx.App.HasFile(astroConfigTs)
	hasAstroBuildCommand := strings.Contains(strings.ToLower(pkg.PackageJson.GetScript("build")), "astro build")

	return hasAstroConfig && hasAstroBuildCommand
}

func (p *NodeProvider) isAstro(ctx *generate.GenerateContext) bool {
	return p.isAstroPackage(p.workspace.Root, ctx)
}

func (p *NodeProvider) isAstroSPA(ctx *generate.GenerateContext) bool {
	if !p.isAstro(ctx) {
		return false
	}

	configFileContents := p.getAstroConfigFileContents(ctx)
	hasServerOutput := strings.Contains(configFileContents, "output: 'server'")
	hasAdapter := p.hasDependency("@astrojs/node") || p.hasDependency("@astrojs/vercel") || p.hasDependency("@astrojs/cloudflare") || p.hasDependency("@astrojs/netlify")

	return !hasServerOutput && !hasAdapter
}

func (p *NodeProvider) getAstroOutputDirectory(ctx *generate.GenerateContext) string {
	configFileContents := p.getAstroConfigFileContents(ctx)
	if configFileContents != "" {
		// Look for outDir in config
		outDirRegex := regexp.MustCompile(`outDir:\s*['"](.+?)['"]`)
		matches := outDirRegex.FindStringSubmatch(configFileContents)
		if len(matches) > 1 {
			return matches[1]
		}
	}

	return DefaultAstroOutputDirectory
}

func (p *NodeProvider) getAstroConfigFileContents(ctx *generate.GenerateContext) string {
	configFile := ""

	if ctx.App.HasFile("astro.config.mjs") {
		contents, err := ctx.App.ReadFile("astro.config.mjs")
		if err == nil {
			configFile = contents
		}
	} else if ctx.App.HasFile("astro.config.ts") {
		contents, err := ctx.App.ReadFile("astro.config.ts")
		if err == nil {
			configFile = contents
		}
	}

	return configFile
}

func (p *NodeProvider) getAstroEnvVars() map[string]string {
	envVars := map[string]string{
		"HOST": "0.0.0.0",
	}

	return envVars
}
