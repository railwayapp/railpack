package node

const (
	// Static dir is resolved relative to the server entry (dist/server/), matching TanStack docs.
	DefaultTanstackSrvxStartCommand = "srvx --prod -s ../client dist/server/server.js"
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

// True when TanStack Start has no start script and we fall back to srvx.
func (p *NodeProvider) usesTanstackSrvxFallback() bool {
	if !p.isTanstackStart() {
		return false
	}

	return p.getScripts(p.packageJson, "start") == ""
}

func (p *NodeProvider) getTanstackStartCommand() string {
	if !p.usesTanstackSrvxFallback() {
		return ""
	}

	return p.packageManager.ExecCommand(DefaultTanstackSrvxStartCommand)
}
