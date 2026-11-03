package python

import (
	"regexp"
	"slices"
	"strings"

	"github.com/railwayapp/railpack/core/generate"
)

var (
	// dependency strings carry the package name inline; poetry and pipenv tables use it as the key
	pythonDependencyNameRegex = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*`)

	pythonDependencyNameSeparatorRegex = regexp.MustCompile(`[-_.]+`)
)

type pyprojectManifest struct {
	Project struct {
		Dependencies   []string `toml:"dependencies"`
		RequiresPython string   `toml:"requires-python"`
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

// keeps the normalized package name and selected extras needed by production dependency detection
type pythonDependency struct {
	name   string
	extras []string
}

// follows install-manifest precedence so development-only dependencies cannot enable production integrations
func (p *PythonProvider) usesProductionDep(ctx *generate.GenerateContext, requestedDependency string) bool {
	if p.hasRequirements(ctx) {
		return requirementsUseDependency(ctx, requestedDependency)
	}

	if p.hasPyproject(ctx) {
		return pyprojectsUseProductionDependency(ctx, requestedDependency)
	}

	if p.hasPipfile(ctx) {
		return pipfileUsesProductionDependency(ctx, requestedDependency)
	}

	return false
}

func requirementsUseDependency(ctx *generate.GenerateContext, requestedDependency string) bool {
	contents, err := ctx.App.ReadFile("requirements.txt")
	if err != nil {
		return false
	}

	for dependencyString := range strings.Lines(contents) {
		if dependencyStringDeclaresPackageAndExtras(dependencyString, requestedDependency) {
			return true
		}
	}

	return false
}

// includes workspace package manifests so their production dependencies can enable runtime integrations
func pyprojectsUseProductionDependency(ctx *generate.GenerateContext, requestedDependency string) bool {
	pyprojectPaths, err := ctx.App.FindFiles("**/pyproject.toml")
	if err != nil {
		return false
	}

	for _, pyprojectPath := range pyprojectPaths {
		var manifest pyprojectManifest
		if err := ctx.App.ReadTOML(pyprojectPath, &manifest); err != nil {
			continue
		}

		for _, dependencyString := range manifest.Project.Dependencies {
			if dependencyStringDeclaresPackageAndExtras(dependencyString, requestedDependency) {
				return true
			}
		}

		for packageName, packageConfig := range manifest.Tool.Poetry.Dependencies {
			if dependencyTableDeclaresPackageAndExtras(packageName, packageConfig, requestedDependency) {
				return true
			}
		}
	}

	return false
}

func pipfileUsesProductionDependency(ctx *generate.GenerateContext, requestedDependency string) bool {
	var manifest pipfileManifest
	if err := ctx.App.ReadTOML("Pipfile", &manifest); err != nil {
		return false
	}

	for packageName, packageConfig := range manifest.Packages {
		if dependencyTableDeclaresPackageAndExtras(packageName, packageConfig, requestedDependency) {
			return true
		}
	}

	return false
}

func dependencyStringDeclaresPackageAndExtras(dependencyString string, requestedDependency string) bool {
	declaredDependency := parseDependencyNameAndExtras(dependencyString)
	requiredDependency := parseDependencyNameAndExtras(requestedDependency)
	return declaredDependency.matchesNameAndRequestedExtras(requiredDependency)
}

// extracts dependency names and extras for framework/server detection
//
// - accepts requirement strings from requirements.txt and pyproject.toml
// - leaves version constraints and environment markers to the package manager
//
// "fastapi[other,standard]>=0.115" -> {name: "fastapi", extras: ["other", "standard"]}
// https://packaging.python.org/en/latest/specifications/dependency-specifiers/#extras
func parseDependencyNameAndExtras(dependencyString string) pythonDependency {
	dependencyString = strings.TrimSpace(dependencyString)
	packageName := pythonDependencyNameRegex.FindString(dependencyString)
	dependency := pythonDependency{name: normalizedDependencyName(packageName)}
	afterPackageName := strings.TrimSpace(strings.TrimPrefix(dependencyString, packageName))
	if !strings.HasPrefix(afterPackageName, "[") {
		return dependency
	}

	extrasList, _, closed := strings.Cut(strings.TrimPrefix(afterPackageName, "["), "]")
	if !closed {
		return dependency
	}
	for extraName := range strings.SplitSeq(extrasList, ",") {
		extraName = strings.TrimSpace(extraName)
		if extraName == "" {
			continue
		}
		dependency.extras = append(dependency.extras, normalizedDependencyName(extraName))
	}
	return dependency
}

// reads poetry/pipenv package options so extras participate in the same checks as dependency strings
//
// ("fastapi", {version: "*", extras: ["other", "standard"]}, "fastapi[standard]") -> true
func dependencyTableDeclaresPackageAndExtras(packageName string, packageConfig any, requestedDependency string) bool {
	declaredDependency := parseDependencyNameAndExtras(packageName)
	if packageOptions, ok := packageConfig.(map[string]any); ok {
		configuredExtras, _ := packageOptions["extras"].([]any)
		for _, configuredExtra := range configuredExtras {
			if extraName, ok := configuredExtra.(string); ok {
				declaredDependency.extras = append(declaredDependency.extras, normalizedDependencyName(extraName))
			}
		}
	}
	requiredDependency := parseDependencyNameAndExtras(requestedDependency)
	return declaredDependency.matchesNameAndRequestedExtras(requiredDependency)
}

func (declaredDependency pythonDependency) matchesNameAndRequestedExtras(requestedDependency pythonDependency) bool {
	if declaredDependency.name == "" || declaredDependency.name != requestedDependency.name {
		return false
	}
	for _, requestedExtra := range requestedDependency.extras {
		if !slices.Contains(declaredDependency.extras, requestedExtra) {
			return false
		}
	}
	return true
}

// avoids missed dependency matches because packaging treats case and runs of "-", "_", and "." as equivalent
//
// "extra_name" -> "extra-name"
// https://peps.python.org/pep-0685/#specification
func normalizedDependencyName(packageOrExtraName string) string {
	return pythonDependencyNameSeparatorRegex.ReplaceAllString(strings.ToLower(packageOrExtraName), "-")
}
