package plan

import "path/filepath"

type Filter struct {
	Include []string `json:"include,omitempty" jsonschema:"description=Files or directories to include"`
	Exclude []string `json:"exclude,omitempty" jsonschema:"description=Files or directories to exclude"`
}

func normalizeFilterPaths(paths []string) []string {
	if paths == nil {
		return nil
	}
	normalized := make([]string, len(paths))
	for i, p := range paths {
		normalized[i] = filepath.ToSlash(p)
	}
	return normalized
}

func NewFilter(include []string, exclude []string) Filter {
	return Filter{
		Include: normalizeFilterPaths(include),
		Exclude: normalizeFilterPaths(exclude),
	}
}

func NewIncludeFilter(include []string) Filter {
	return Filter{
		Include: normalizeFilterPaths(include),
	}
}
