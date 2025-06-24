package node

import (
	"encoding/json"
	"maps"
	"strings"
)

type WorkspacesConfig struct {
	Packages []string `json:"packages"`
}

type PackageJson struct {
	Name            string            `json:"name"`
	Version         string            `json:"version"`
	Scripts         map[string]string `json:"scripts"`
	PackageManager  *string           `json:"packageManager"`
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
	Engines         map[string]string `json:"engines"`
	Main            string            `json:"main"`
	Workspaces      []string          `json:"-"`
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

func (p *PackageJson) hasDependency(dependency string) bool {
	if p.Dependencies != nil {
		if _, ok := p.Dependencies[dependency]; ok {
			return true
		}
	}

	if p.DevDependencies != nil {
		if _, ok := p.DevDependencies[dependency]; ok {
			return true
		}
	}

	return false
}

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

func (p *PackageJson) GetPackageManagerInfo() (string, string) {
	if p == nil || p.PackageManager == nil {
		return "", ""
	}

	pmString := *p.PackageManager
	parts := strings.Split(pmString, "@")
	if len(parts) == 2 {
		versionParts := strings.Split(parts[1], "+")
		return parts[0], versionParts[0]
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
		Workspaces interface{} `json:"workspaces"`
	}{
		Alias: (*Alias)(p),
	}

	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	// Handle workspaces field based on its type
	switch w := aux.Workspaces.(type) {
	case []interface{}:
		p.Workspaces = make([]string, len(w))
		for i, v := range w {
			if s, ok := v.(string); ok {
				p.Workspaces[i] = s
			}
		}
	case map[string]interface{}:
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
