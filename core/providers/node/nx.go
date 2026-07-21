package node

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/railwayapp/railpack/core/generate"
)

// isNx reports whether this is an Nx workspace we can drive with nx CLI fallbacks.
func (p *NodeProvider) isNx(ctx *generate.GenerateContext) bool {
	if !ctx.App.HasFile("nx.json") {
		return false
	}

	if p.workspace != nil && p.workspace.HasDependency("nx") {
		return true
	}

	return false
}

// isNextAppPackage identifies a deployable Next.js package in a monorepo.
// Root packages that only list next as a shared dependency (common in Nx scaffolds)
// are excluded unless they have a next.config or a next build script.
func (p *NodeProvider) isNextAppPackage(pkg *WorkspacePackage, ctx *generate.GenerateContext) bool {
	if pkg == nil || pkg.PackageJson == nil {
		return false
	}

	if p.packageHasNextConfig(ctx, pkg.Path) {
		return true
	}

	if !pkg.PackageJson.hasDependency("next") {
		return false
	}

	if pkg.PackageJson.BuildScriptContains("next build") {
		return true
	}

	// Workspace packages with next but no scripts (Nx inferred targets)
	return pkg.Path != ""
}

func (p *NodeProvider) packageHasNextConfig(ctx *generate.GenerateContext, packagePath string) bool {
	for _, name := range nextConfigFiles {
		path := name
		if packagePath != "" {
			path = filepath.Join(packagePath, name)
		}
		if ctx.App.HasFile(path) {
			return true
		}
	}
	return false
}

func (p *NodeProvider) getNxNextPackages(ctx *generate.GenerateContext) []*WorkspacePackage {
	packages := []*WorkspacePackage{}
	if p.workspace == nil {
		return packages
	}

	if p.isNextAppPackage(p.workspace.Root, ctx) {
		packages = append(packages, p.workspace.Root)
	}

	for _, pkg := range p.workspace.Packages {
		if p.isNextAppPackage(pkg, ctx) {
			packages = append(packages, pkg)
		}
	}

	return packages
}

// nxProjectName returns the name Nx uses for a package (package.json name, or directory basename).
func nxProjectName(pkg *WorkspacePackage) string {
	if pkg.PackageJson != nil && pkg.PackageJson.Name != "" {
		return pkg.PackageJson.Name
	}
	if pkg.Path != "" {
		return filepath.Base(pkg.Path)
	}
	return ""
}

func matchesNxAppSelector(pkg *WorkspacePackage, selector string) bool {
	if selector == "" {
		return false
	}

	if pkg.PackageJson != nil && pkg.PackageJson.Name == selector {
		return true
	}

	if pkg.Path == selector {
		return true
	}

	if pkg.Path != "" && filepath.Base(pkg.Path) == selector {
		return true
	}

	// "@scope/web" matches selector "web"
	if pkg.PackageJson != nil {
		name := pkg.PackageJson.Name
		if i := strings.LastIndex(name, "/"); i >= 0 && name[i+1:] == selector {
			return true
		}
	}

	return false
}

// resolveNxDeployPackage picks the Nx app to build/start when root scripts are absent.
// Returns the workspace package, Nx project name, and whether a selection was made.
func (p *NodeProvider) resolveNxDeployPackage(ctx *generate.GenerateContext) (*WorkspacePackage, string, bool) {
	if !p.isNx(ctx) {
		return nil, "", false
	}

	packages := p.getNxNextPackages(ctx)
	if len(packages) == 0 {
		return nil, "", false
	}

	if selector, _ := ctx.Env.GetConfigVariable("NX_APP"); selector != "" {
		for _, pkg := range packages {
			if matchesNxAppSelector(pkg, selector) {
				name := nxProjectName(pkg)
				ctx.Logger.LogInfo("Using Nx app %s (RAILPACK_NX_APP=%s)", name, selector)
				return pkg, name, true
			}
		}

		ctx.Logger.LogWarn("RAILPACK_NX_APP=%s did not match any Next.js app package", selector)
		return nil, "", false
	}

	if len(packages) == 1 {
		pkg := packages[0]
		name := nxProjectName(pkg)
		ctx.Logger.LogInfo("Using Nx app %s", name)
		return pkg, name, true
	}

	names := make([]string, 0, len(packages))
	for _, pkg := range packages {
		names = append(names, nxProjectName(pkg))
	}
	ctx.Logger.LogWarn("Multiple Next.js apps found in Nx workspace (%s). Set RAILPACK_NX_APP to choose one.", strings.Join(names, ", "))
	return nil, "", false
}

// nxBuildCommand returns the shell command to build an Nx project.
func nxBuildCommand(projectName string) string {
	return fmt.Sprintf("nx build %s", projectName)
}

// nxNextStartCommand returns a production start command for a Next.js app package.
func nxNextStartCommand(pkg *WorkspacePackage) string {
	if pkg.Path == "" {
		return DefaultNextStartCommand
	}
	return fmt.Sprintf("cd %s && %s", pkg.Path, DefaultNextStartCommand)
}
