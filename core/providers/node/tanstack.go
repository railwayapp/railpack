package node

import (
	"strings"

	"github.com/railwayapp/railpack/core/generate"
)

const (
	// Static dir is resolved relative to the server entry (dist/server/), matching TanStack docs.
	DefaultTanstackSrvxStartCommand = "srvx --prod -s ../client dist/server/server.js"

	// The Nitro Vite plugin builds a self-contained server, matching the Nuxt output layout.
	DefaultTanstackNitroStartCommand = "node .output/server/index.mjs"
)

func (p *NodeProvider) isTanstackStart() bool {
	return p.isTanstackStartPackage(p.workspace.Root)
}

func (p *NodeProvider) isTanstackStartPackage(pkg *WorkspacePackage) bool {
	if pkg == nil || pkg.PackageJson == nil {
		return false
	}

	return pkg.PackageJson.hasProductionDependency("@tanstack/react-start")
}

// True when TanStack Start has no start script and we fall back to a default server command.
func (p *NodeProvider) usesTanstackStartFallback() bool {
	if !p.isTanstackStart() {
		return false
	}

	return p.getScripts(p.packageJson, "start") == ""
}

// True when the app builds with the Nitro Vite plugin, which outputs a self-contained
// server to .output/ instead of the default dist/ layout. The TanStack scaffolder's
// nitro deployment option adds the dependency and plugin without adding a start script.
func (p *NodeProvider) usesTanstackNitro(ctx *generate.GenerateContext) bool {
	if !p.isTanstackStart() || !p.hasDependency("nitro") {
		return false
	}

	_, configContent, _ := ctx.App.ReadFirstFileOf("vite.config.js", "vite.config.ts")
	return strings.Contains(configContent, "nitro")
}

func (p *NodeProvider) getTanstackStartCommand(ctx *generate.GenerateContext) string {
	if !p.usesTanstackStartFallback() {
		return ""
	}

	// Nitro builds .output/, so the srvx command (which serves dist/) would crash on startup.
	if p.usesTanstackNitro(ctx) {
		return DefaultTanstackNitroStartCommand
	}

	return p.packageManager.ExecCommand(DefaultTanstackSrvxStartCommand)
}

// True when @tanstack/react-start is only in devDependencies. Detection requires a
// production dependency, so these apps silently deploy as static Vite sites with
// SSR and server routes broken.
func (p *NodeProvider) hasMisplacedTanstackStart() bool {
	pkg := p.workspace.Root
	if pkg == nil || pkg.PackageJson == nil {
		return false
	}

	return !pkg.PackageJson.hasProductionDependency("@tanstack/react-start") &&
		pkg.PackageJson.hasDevDependency("@tanstack/react-start")
}
