package node

import (
	"regexp"
	"strings"

	"github.com/railwayapp/railpack/core/generate"
)

const (
	DefaultViteOutputDirectory = "dist"

	// The Nitro Vite plugin builds a self-contained server, matching the Nuxt output layout.
	DefaultNitroStartCommand = "node .output/server/index.mjs"
)

func viteConfigFiles(pkg *WorkspacePackage) []string {
	viteConfigJS := "vite.config.js"
	viteConfigTS := "vite.config.ts"
	if pkg.Path != "" {
		viteConfigJS = pkg.Path + "/vite.config.js"
		viteConfigTS = pkg.Path + "/vite.config.ts"
	}

	return []string{viteConfigJS, viteConfigTS}
}

func (p *NodeProvider) isVitePackage(pkg *WorkspacePackage, ctx *generate.GenerateContext) bool {
	hasViteBuildCommand := strings.Contains(strings.ToLower(pkg.PackageJson.GetScript("build")), "vite build")
	hasVitePackage := pkg.PackageJson.hasDependency("vite")
	hasBuildCommand := pkg.PackageJson.HasScript("build")

	viteConfigs := viteConfigFiles(pkg)
	hasViteConfig := ctx.App.HasFile(viteConfigs[0]) || ctx.App.HasFile(viteConfigs[1])

	// SvelteKit, TanStack Start and Nitro apps do not build as static sites by default
	if p.isSvelteKitPackage(pkg) || p.isTanstackStartPackage(pkg) || p.usesViteNitroPackage(pkg, ctx) {
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

// True when the app builds with the Nitro Vite plugin, which outputs a self-contained
// server to .output/ instead of static assets in dist/. Fresh 3 registers nitro() inside
// its own fresh() plugin, so there nitro is only a transitive dependency.
func (p *NodeProvider) usesViteNitroPackage(pkg *WorkspacePackage, ctx *generate.GenerateContext) bool {
	if pkg == nil || pkg.PackageJson == nil {
		return false
	}

	hasNitro := pkg.PackageJson.hasDependency("nitro")
	hasFresh := pkg.PackageJson.hasDependency("fresh")
	if !hasNitro && !hasFresh {
		return false
	}

	_, configContent, _ := ctx.App.ReadFirstFileOf(viteConfigFiles(pkg)...)
	return (hasNitro && strings.Contains(configContent, "nitro")) ||
		(hasFresh && strings.Contains(configContent, "fresh/vite"))
}

func (p *NodeProvider) usesViteNitro(ctx *generate.GenerateContext) bool {
	return p.usesViteNitroPackage(p.workspace.Root, ctx)
}

func (p *NodeProvider) getViteOutputDirectory(ctx *generate.GenerateContext) string {
	_, configContent, _ := ctx.App.ReadFirstFileOf("vite.config.js", "vite.config.ts")

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
