package node

import (
	"fmt"
	"path/filepath"

	"github.com/railwayapp/railpack/core/app"
)

type Workspace struct {
	Root     *WorkspacePackage
	Packages []*WorkspacePackage
}

type WorkspacePackage struct {
	Path        string
	PackageJson *PackageJson
}

type PnpmWorkspace struct {
	Packages []string `yaml:"packages"`
}

// NewWorkspace creates a new workspace from a package.json file
func NewWorkspace(app *app.App) (*Workspace, error) {
	packageJson, err := readPackageJson(app, "package.json")
	if err != nil {
		return nil, fmt.Errorf("error reading root package.json: %w", err)
	}

	workspace := &Workspace{
		Root: &WorkspacePackage{
			Path:        "",
			PackageJson: packageJson,
		},
		Packages: []*WorkspacePackage{},
	}

	// Try to read PNPM workspace config first
	if app.HasMatch("pnpm-workspace.yaml") {
		var pnpmWorkspace PnpmWorkspace
		if err := app.ReadYAML("pnpm-workspace.yaml", &pnpmWorkspace); err == nil && len(pnpmWorkspace.Packages) > 0 {
			workspace.Root.PackageJson.Workspaces = pnpmWorkspace.Packages
		}
	}

	if len(workspace.Root.PackageJson.Workspaces) > 0 {
		if err := workspace.findWorkspacePackages(app); err != nil {
			return nil, err
		}
	}

	return workspace, nil
}

// findWorkspacePackages finds all packages in the workspace using the workspace patterns
func (w *Workspace) findWorkspacePackages(app *app.App) error {
	for _, pattern := range w.Root.PackageJson.Workspaces {
		// For each workspace pattern, we need to:
		// 1. Find all package.json files in that pattern
		// 2. Read each package.json file
		// 3. Add it to our list of packages

		pattern = convertWorkspacePattern(pattern)
		matches, err := app.FindFiles(pattern)
		if err != nil {
			continue
		}

		for _, match := range matches {
			packageJson, err := readPackageJson(app, match)
			if err != nil {
				continue
			}

			dir := filepath.Dir(match)
			w.Packages = append(w.Packages, &WorkspacePackage{
				Path:        dir,
				PackageJson: packageJson,
			})
		}
	}

	return nil
}

// convertWorkspacePattern converts npm/pnpm workspace patterns to glob patterns
func convertWorkspacePattern(pattern string) string {
	// npm/pnpm uses packages/* or packages/** for glob patterns
	// - packages/* -> packages/*/package.json (single level)
	// - packages/** -> packages/**/package.json (recursive)
	if len(pattern) >= 2 && pattern[len(pattern)-2:] == "/*" {
		// Single level pattern (packages/*)
		return pattern[:len(pattern)-1] + "*/package.json"
	}

	if len(pattern) >= 3 && pattern[len(pattern)-3:] == "/**" {
		// Recursive pattern (packages/**)
		return pattern[:len(pattern)-2] + "**/package.json"
	}

	// Direct path or other pattern, just append package.json
	return filepath.Join(pattern, "package.json")
}

// readPackageJson reads a package.json file from the given path
func readPackageJson(app *app.App, path string) (*PackageJson, error) {
	packageJson := NewPackageJson()
	if err := app.ReadJSON(path, packageJson); err != nil {
		return nil, err
	}
	return packageJson, nil
}

// HasWorkspaces returns true if this is a workspace root
func (w *Workspace) HasWorkspaces() bool {
	return len(w.Root.PackageJson.Workspaces) > 0
}

// GetPackage returns a workspace package by path
func (w *Workspace) GetPackage(path string) *WorkspacePackage {
	for _, pkg := range w.Packages {
		if pkg.Path == path {
			return pkg
		}
	}
	return nil
}

func (w *Workspace) HasDependency(dependency string) bool {
	if w.Root.PackageJson.hasDependency(dependency) {
		return true
	}

	for _, pkg := range w.Packages {
		if pkg.PackageJson.hasDependency(dependency) {
			return true
		}
	}

	return false
}

func (w *Workspace) AllPackageJson() []*PackageJson {
	packageJsons := []*PackageJson{w.Root.PackageJson}

	for _, pkg := range w.Packages {
		packageJsons = append(packageJsons, pkg.PackageJson)
	}

	return packageJsons
}
