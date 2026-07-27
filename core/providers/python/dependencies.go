package python

import (
	"regexp"
	"strings"

	"github.com/railwayapp/railpack/core/generate"
)

var (
	// Extract the leading distribution name from requirements.txt and PEP 508 dependency specs.
	// This is not used for Poetry or Pipenv tables, where the distribution name is the TOML key.
	pythonDependencyNameRegex = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*`)

	// Normalize the interchangeable package-name separators defined by Python packaging conventions.
	pythonDependencyNameSeparatorRegex = regexp.MustCompile(`[-_.]+`)
)

type pyprojectManifest struct {
	Project struct {
		Dependencies []string `toml:"dependencies"`
	} `toml:"project"`
	Tool struct {
		Poetry struct {
			Dependencies map[string]any `toml:"dependencies"`
		} `toml:"poetry"`
	} `toml:"tool"`
}

type pipfileManifest struct {
	Packages map[string]any `toml:"packages"`
}

func (p *PythonProvider) usesProductionDep(ctx *generate.GenerateContext, dep string) bool {
	if p.hasRequirements(ctx) {
		return requirementsUseDependency(ctx, dep)
	}

	if p.hasPyproject(ctx) {
		return pyprojectsUseProductionDependency(ctx, dep)
	}

	if p.hasPipfile(ctx) {
		return pipfileUsesProductionDependency(ctx, dep)
	}

	return false
}

func requirementsUseDependency(ctx *generate.GenerateContext, dep string) bool {
	contents, err := ctx.App.ReadFile("requirements.txt")
	if err != nil {
		return false
	}

	for line := range strings.Lines(contents) {
		if dependencySpecMatches(line, dep) {
			return true
		}
	}

	return false
}

func pyprojectsUseProductionDependency(ctx *generate.GenerateContext, dep string) bool {
	files, err := ctx.App.FindFiles("**/pyproject.toml")
	if err != nil {
		return false
	}

	for _, file := range files {
		var manifest pyprojectManifest
		if err := ctx.App.ReadTOML(file, &manifest); err != nil {
			continue
		}

		for _, dependency := range manifest.Project.Dependencies {
			if dependencySpecMatches(dependency, dep) {
				return true
			}
		}

		for dependency := range manifest.Tool.Poetry.Dependencies {
			if normalizedDependencyName(dependency) == normalizedDependencyName(dep) {
				return true
			}
		}
	}

	return false
}

func pipfileUsesProductionDependency(ctx *generate.GenerateContext, dep string) bool {
	var manifest pipfileManifest
	if err := ctx.App.ReadTOML("Pipfile", &manifest); err != nil {
		return false
	}

	for dependency := range manifest.Packages {
		if normalizedDependencyName(dependency) == normalizedDependencyName(dep) {
			return true
		}
	}

	return false
}

func dependencySpecMatches(spec string, dep string) bool {
	name := pythonDependencyNameRegex.FindString(strings.TrimSpace(spec))
	return normalizedDependencyName(name) == normalizedDependencyName(dep)
}

func normalizedDependencyName(name string) string {
	return pythonDependencyNameSeparatorRegex.ReplaceAllString(strings.ToLower(name), "-")
}
