package node

import (
	"regexp"
	"strings"

	"github.com/railwayapp/railpack/core/generate"
)

const (
	DefaultViteOutputDirectory = "dist"
)

func (p *NodeProvider) isVitePackage(pkg *WorkspacePackage, ctx *generate.GenerateContext) bool {
	hasViteBuildCommand := strings.Contains(strings.ToLower(pkg.PackageJson.GetScript("build")), "vite build")
	hasVitePackage := pkg.PackageJson.hasDependency("vite")
	hasBuildCommand := pkg.PackageJson.HasScript("build")

	viteConfigJS := "vite.config.js"
	viteConfigTS := "vite.config.ts"
	if pkg.Path != "" {
		viteConfigJS = pkg.Path + "/vite.config.js"
		viteConfigTS = pkg.Path + "/vite.config.ts"
	}

	hasViteConfig := ctx.App.HasFile(viteConfigJS) || ctx.App.HasFile(viteConfigTS)

	// SvelteKit does not build as a static site by default
	if p.isSvelteKitPackage(pkg) {
		return false
	}

	// We only consider a package "vite" if
	// 1. It has a vite package.json dependency
	// 2. It has a build command. This is to ensure that there will actually be a directory to serve
	// 3. It has a vite config file OR has a vite build command in the build command
	return hasVitePackage && hasBuildCommand && (hasViteBuildCommand || hasViteConfig)
}

func (p *NodeProvider) isVite(ctx *generate.GenerateContext) bool {
	return p.isVitePackage(p.workspace.Root, ctx)
}

func (p *NodeProvider) getViteOutputDirectory(ctx *generate.GenerateContext) string {
	configContent := ""

	if ctx.App.HasFile("vite.config.js") {
		content, err := ctx.App.ReadFile("vite.config.js")
		if err == nil {
			configContent = content
		}
	} else if ctx.App.HasFile("vite.config.ts") {
		content, err := ctx.App.ReadFile("vite.config.ts")
		if err == nil {
			configContent = content
		}
	}

	if configContent != "" {
		// Look for outDir in config
		outDirRegex := regexp.MustCompile(`outDir:\s*['"](.+?)['"]`)
		matches := outDirRegex.FindStringSubmatch(configContent)
		if len(matches) > 1 {
			return matches[1]
		}
	}

	// Check for outDir in build script
	if p.packageJson.Scripts != nil {
		if buildScript, ok := p.packageJson.Scripts["build"]; ok {
			outDirRegex := regexp.MustCompile(`vite\s+build(?:\s+-[^\s]*)*\s+(?:--outDir)\s+([^-\s;]+)`)
			matches := outDirRegex.FindStringSubmatch(buildScript)
			if len(matches) > 1 {
				return matches[1]
			}
		}
	}

	return DefaultViteOutputDirectory
}

func (p *NodeProvider) isSvelteKitPackage(pkg *WorkspacePackage) bool {
	return pkg.PackageJson.hasDependency("svelte") && pkg.PackageJson.hasDependency("@sveltejs/kit")
}
