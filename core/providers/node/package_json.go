package node

import (
	"encoding/json"
	"maps"
	"strings"
)

type WorkspacesConfig struct {
	Packages []string `json:"packages"`
}

// a single `devEngines` entry, e.g. {"name": "pnpm", "version": "10.5.0"}
type DevEngineDependency struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// https://nodejs.org/api/packages.html#devengines
// each field is either a single object or an array of objects
type DevEngines struct {
	Runtime        []DevEngineDependency `json:"runtime"`
	PackageManager []DevEngineDependency `json:"packageManager"`
}

type PackageJson struct {
	Name            string            `json:"name"`
	Version         string            `json:"version"`
	Scripts         map[string]string `json:"scripts"`
	PackageManager  *string           `json:"packageManager"`
	DevEngines      *DevEngines       `json:"devEngines"`
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
	Engines         map[string]string `json:"engines"`
	Main            string            `json:"main"`
	Workspaces      []string          `json:"-"`
}

// devEngines entries accept either a single object or an array of objects, so both shapes are unmarshalled here.
func (d *DevEngines) UnmarshalJSON(data []byte) error {
	type Alias DevEngines
	aux := &struct {
		Runtime        any `json:"runtime"`
		PackageManager any `json:"packageManager"`
		*Alias
	}{
		Alias: (*Alias)(d),
	}

	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	d.Runtime = parseDevEngineDependencies(aux.Runtime)
	d.PackageManager = parseDevEngineDependencies(aux.PackageManager)

	return nil
}

func parseDevEngineDependencies(value any) []DevEngineDependency {
	if value == nil {
		return nil
	}

	raw, err := json.Marshal(value)
	if err != nil {
		return nil
	}

	var dependencies []DevEngineDependency
	if err := json.Unmarshal(raw, &dependencies); err == nil {
		return dependencies
	}

	var dependency DevEngineDependency
	if err := json.Unmarshal(raw, &dependency); err == nil {
		return []DevEngineDependency{dependency}
	}

	return nil
}

// the name and version of the package manager declared in `devEngines.packageManager`.
// returns empty strings when it is not declared.
func (p *PackageJson) GetDevEnginePackageManager() (string, string) {
	if p == nil || p.DevEngines == nil {
		return "", ""
	}

	for _, dependency := range p.DevEngines.PackageManager {
		name := strings.TrimSpace(dependency.Name)
		if name != "" {
			return name, strings.TrimSpace(dependency.Version)
		}
	}

	return "", ""
}

// the version of a runtime declared in `devEngines.runtime`, e.g. "node".
// returns an empty string when that runtime is not declared.
func (p *PackageJson) GetDevEngineRuntimeVersion(name string) string {
	if p == nil || p.DevEngines == nil {
		return ""
	}

	for _, dependency := range p.DevEngines.Runtime {
		if strings.EqualFold(strings.TrimSpace(dependency.Name), name) {
			return strings.TrimSpace(dependency.Version)
		}
	}

	return ""
}

func NewPackageJson() *PackageJson {
	return &PackageJson{
		Scripts:    map[string]string{},
		Engines:    map[string]string{},
		Workspaces: []string{},
	}
}

func (p *PackageJson) HasScript(name string) bool {
	return p.Scripts != nil && p.Scripts[name] != ""
}

func (p *PackageJson) GetScript(name string) string {
	if p.Scripts == nil {
		return ""
	}

	return p.Scripts[name]
}

func (p *PackageJson) BuildScriptContains(value string) bool {
	return strings.Contains(p.GetScript("build"), value)
}

func (p *PackageJson) hasDependency(dependency string) bool {
	return p.hasProductionDependency(dependency) || p.hasDevDependency(dependency)
}

// is there a dependency that requires the entire app to be loaded into the context
func (p *PackageJson) hasLocalDependency() bool {
	allDeps := make(map[string]string)
	maps.Copy(allDeps, p.Dependencies)
	maps.Copy(allDeps, p.DevDependencies)

	for _, dependency := range allDeps {
		if strings.HasPrefix(dependency, "file:") {
			return true
		}
	}

	return false
}

// parse a packageManager string in the format "name@version" or "name@version+extra".
// Returns the package manager name and version as separate strings.
// returns empty strings for both name and version if it can't be parsed
func (p *PackageJson) GetPackageManagerInfo() (string, string) {
	if p == nil || p.PackageManager == nil {
		return "", ""
	}

	pmString := strings.TrimSpace(*p.PackageManager)
	parts := strings.Split(pmString, "@")
	if len(parts) == 2 {
		versionParts := strings.Split(parts[1], "+")
		return strings.TrimSpace(parts[0]), strings.TrimSpace(versionParts[0])
	}

	return "", ""
}

func (p *PackageJson) UnmarshalJSON(data []byte) error {
	type WorkspacesObject struct {
		Packages []string `json:"packages"`
	}

	type Alias PackageJson
	aux := &struct {
		*Alias
		Workspaces any `json:"workspaces"`
	}{
		Alias: (*Alias)(p),
	}

	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	// Handle workspaces field based on its type
	switch w := aux.Workspaces.(type) {
	case []any:
		p.Workspaces = make([]string, len(w))
		for i, v := range w {
			if s, ok := v.(string); ok {
				p.Workspaces[i] = s
			}
		}
	case map[string]any:
		// Try to unmarshal as WorkspacesObject
		var wo WorkspacesObject
		if b, err := json.Marshal(w); err == nil {
			if err := json.Unmarshal(b, &wo); err == nil {
				p.Workspaces = wo.Packages
			}
		}
	}

	return nil
}
