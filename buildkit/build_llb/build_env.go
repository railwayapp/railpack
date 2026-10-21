package build_llb

import (
	"maps"
	"slices"
)

type BuildEnvironment struct {
	PathList []string
	EnvVars  map[string]string
}

func NewGraphEnvironment() BuildEnvironment {
	return BuildEnvironment{
		PathList: make([]string, 0),
		EnvVars:  make(map[string]string),
	}
}

// Merges the other environment into the current environment
func (e *BuildEnvironment) Merge(other BuildEnvironment) {
	e.PathList = appendUniquePaths(e.PathList, other.PathList...)
	maps.Copy(e.EnvVars, other.EnvVars)
}

func appendUniquePaths(existing []string, additions ...string) []string {
	// Steps with shared ancestors inherit the same paths, so merging their outputs can
	// otherwise duplicate entries in downstream steps and the final image PATH. Keep
	// the first occurrence to preserve path precedence.
	for _, path := range additions {
		if slices.Contains(existing, path) {
			continue
		}
		existing = append(existing, path)
	}
	return existing
}

func (e *BuildEnvironment) PushPath(path string) {
	if slices.Contains(e.PathList, path) {
		return
	}
	e.PathList = append([]string{path}, e.PathList...)
}

func (e *BuildEnvironment) AddEnvVar(key, value string) {
	e.EnvVars[key] = value
}
