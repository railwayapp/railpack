package node

import "github.com/railwayapp/railpack/core/generate"

const DefaultSvelteKitStartCommand = "node build"

func (p *NodeProvider) configureSvelteKit(ctx *generate.GenerateContext, build *generate.CommandStepBuilder) {
	if !p.isSvelteKit() || !p.usesSvelteKitAdapterAuto() {
		return
	}

	// Causes @sveltejs/adapter-auto to select adapter-node, so the build emits a Node server.
	// Oddly, GCP_BUILDPACKS is the only env var adapter-auto recognizes for Node; there is no generic/Railway signal.
	ctx.Logger.LogInfo("SvelteKit with adapter-auto detected, forcing adapter-node")
	build.AddVariables(map[string]string{"GCP_BUILDPACKS": "true"})
}

func (p *NodeProvider) getSvelteKitStartCommand() string {
	if !p.isSvelteKit() || !p.usesSvelteKitAdapterAuto() {
		return ""
	}

	return DefaultSvelteKitStartCommand
}

func (p *NodeProvider) isSvelteKit() bool {
	return p.isSvelteKitPackage(p.workspace.Root)
}

func (p *NodeProvider) isSvelteKitPackage(pkg *WorkspacePackage) bool {
	if pkg == nil || pkg.PackageJson == nil {
		return false
	}

	return pkg.PackageJson.hasDependency("svelte") && pkg.PackageJson.hasDependency("@sveltejs/kit")
}

func (p *NodeProvider) usesSvelteKitAdapterAuto() bool {
	return p.hasDependency("@sveltejs/adapter-auto")
}
