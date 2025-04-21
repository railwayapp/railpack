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
			name: "no overlap",
			input: []plan.Layer{
				plan.NewStepLayer("install", plan.NewIncludeFilter([]string{"node_modules"})),
				plan.NewStepLayer("build", plan.NewIncludeFilter([]string{"."})),
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
