package build_llb

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/railwayapp/railpack/core/plan"
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
