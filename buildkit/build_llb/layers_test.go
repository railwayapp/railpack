package build_llb

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/railwayapp/railpack/core/plan"
	"github.com/stretchr/testify/require"
)

func TestShouldLLBMerge(t *testing.T) {
	tests := []struct {
		name     string
		input    []plan.Layer
		expected bool
	}{
		{
			name:     "no layers",
			input:    []plan.Layer{},
			expected: true,
		},

		{
			name: "no overlap with excludes",
			input: []plan.Layer{
				plan.NewStepLayer("install", plan.NewIncludeFilter([]string{"/app/node_modules"})),
				plan.NewStepLayer("build", plan.NewFilter([]string{"."}, []string{"node_modules"})),
				plan.NewStepLayer("build", plan.NewIncludeFilter([]string{"/root/.cache"})),
			},
			expected: true,
		},

		{
			name: "no overlap different roots",
			input: []plan.Layer{
				plan.NewStepLayer("mise", plan.NewIncludeFilter([]string{"/mise/shims", "/mise/installs"})),
				plan.NewStepLayer("build", plan.NewIncludeFilter([]string{"/root/.cache"})),
			},
			expected: true,
		},

		{
			name: "overlapping include",
			input: []plan.Layer{
				plan.NewStepLayer("build", plan.NewIncludeFilter([]string{"."})),
				plan.NewStepLayer("build", plan.NewIncludeFilter([]string{".", "/root/.cache"})),
			},
			expected: false,
		},

		{
			name: "overlapping with exclude",
			input: []plan.Layer{
				plan.NewStepLayer("build", plan.NewFilter([]string{"/root/.cache", "."}, []string{"node_modules", ".yarn"})),
				plan.NewStepLayer("build", plan.NewFilter([]string{"/something/else", "."}, []string{})),
			},
			expected: false,
		},

		{
			name: "path contains no exclude",
			input: []plan.Layer{
				plan.NewStepLayer("install", plan.NewIncludeFilter([]string{"/app/node_modules"})),
				plan.NewStepLayer("build", plan.NewIncludeFilter([]string{"/app"})),
			},
			expected: false,
		},

		{
			name: "overlap excluded by containing layer",
			input: []plan.Layer{
				plan.NewStepLayer("install", plan.NewIncludeFilter([]string{"/app/node_modules"})),
				plan.NewStepLayer("build", plan.NewFilter([]string{"/app"}, []string{"node_modules"})),
			},
			expected: true,
		},

		{
			name: "nested path excluded",
			input: []plan.Layer{
				plan.NewStepLayer("install", plan.NewIncludeFilter([]string{"/app/node_modules/.cache"})),
				plan.NewStepLayer("build", plan.NewFilter([]string{"/app"}, []string{"node_modules"})),
			},
			expected: true,
		},

		{
			name: "relative path overlap not excluded",
			input: []plan.Layer{
				plan.NewStepLayer("mise", plan.NewIncludeFilter([]string{".nvmrc"})),
				plan.NewStepLayer("build", plan.NewFilter([]string{"."}, []string{"node_modules", ".yarn"})),
			},
			expected: false,
		},

		{
			name: "relative path overlap is excluded",
			input: []plan.Layer{
				plan.NewStepLayer("mise", plan.NewIncludeFilter([]string{".nvmrc"})),
				plan.NewStepLayer("build", plan.NewFilter([]string{"."}, []string{"node_modules", ".nvmrc"})),
			},
			expected: true,
		},

		{
			name: "multiple overlaps some excluded",
			input: []plan.Layer{
				plan.NewStepLayer("install", plan.NewIncludeFilter([]string{"/app/node_modules", "/app/dist"})),
				plan.NewStepLayer("build", plan.NewFilter([]string{"/app"}, []string{"node_modules"})),
			},
			expected: false, // /app/dist is not excluded, so still overlap
		},

		{
			name: "all overlaps excluded",
			input: []plan.Layer{
				plan.NewStepLayer("install", plan.NewIncludeFilter([]string{"/app/node_modules", "/app/.yarn"})),
				plan.NewStepLayer("build", plan.NewFilter([]string{"/app"}, []string{"node_modules", ".yarn"})),
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldLLBMerge(tt.input)
			if diff := cmp.Diff(tt.expected, got); diff != "" {
				t.Errorf("shouldLLBMerge() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestIsPathExcluded(t *testing.T) {
	tests := []struct {
		name     string
		relPath  string
		excludes []string
		expected bool
	}{
		{
			name:     "no excludes",
			relPath:  "node_modules",
			excludes: []string{},
			expected: false,
		},
		{
			name:     "exact match",
			relPath:  "node_modules",
			excludes: []string{"node_modules"},
			expected: true,
		},
		{
			name:     "nested path with excluded parent",
			relPath:  "node_modules/.cache",
			excludes: []string{"node_modules"},
			expected: true,
		},
		{
			name:     "dotfile excluded",
			relPath:  ".nvmrc",
			excludes: []string{".nvmrc"},
			expected: true,
		},
		{
			name:     "dotfile not excluded",
			relPath:  ".nvmrc",
			excludes: []string{"node_modules", ".yarn"},
			expected: false,
		},
		{
			name:     "deep nested path",
			relPath:  "node_modules/foo/bar/baz",
			excludes: []string{"node_modules"},
			expected: true,
		},
		{
			name:     "partial name no match",
			relPath:  "node_modules_backup",
			excludes: []string{"node_modules"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isPathExcluded(tt.relPath, tt.excludes)
			require.Equal(t, tt.expected, got)
		})
	}
}

func TestPathOverlap(t *testing.T) {
	tests := []struct {
		name     string
		paths1   []string
		paths2   []string
		expected bool
	}{
		{
			name:     "no overlap",
			paths1:   []string{"/app/node_modules"},
			paths2:   []string{"/app/dist"},
			expected: false,
		},
		{
			name:     "direct overlap",
			paths1:   []string{"/app/node_modules", "/app/dist"},
			paths2:   []string{"/app/dist", "/app/src"},
			expected: true,
		},
		{
			name:     "prefix overlap",
			paths1:   []string{"/app/node_modules/foo"},
			paths2:   []string{"/app/node_modules"},
			expected: true,
		},
		{
			name:     "root path overlap",
			paths1:   []string{"/app/dist"},
			paths2:   []string{"/app"},
			expected: true,
		},
		{
			name:     "different roots no overlap",
			paths1:   []string{"/app/node_modules"},
			paths2:   []string{"/var/lib"},
			expected: false,
		},
		{
			name:     "similar names no overlap",
			paths1:   []string{"/app-foo"},
			paths2:   []string{"/app"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasPathOverlap(tt.paths1, tt.paths2)
			require.Equal(t, tt.expected, got)
		})
	}
}
