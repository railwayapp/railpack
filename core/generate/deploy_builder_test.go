package generate

import (
	"testing"

	"github.com/railwayapp/railpack/core/plan"
	"github.com/stretchr/testify/assert"
)

func TestHasIncludeForStep(t *testing.T) {
	tests := []struct {
		name     string
		inputs   []plan.Layer
		stepName string
		path     string
		expected bool
	}{
		{
			name:     "empty inputs",
			inputs:   []plan.Layer{},
			stepName: "build",
			path:     ".",
			expected: false,
		},
		{
			name: "exact match",
			inputs: []plan.Layer{
				{Step: "build", Filter: plan.Filter{Include: []string{"."}}},
			},
			stepName: "build",
			path:     ".",
			expected: true,
		},
		{
			name: "dot covers any path",
			inputs: []plan.Layer{
				{Step: "build", Filter: plan.Filter{Include: []string{"."}}},
			},
			stepName: "build",
			path:     "/app/dist",
			expected: true,
		},
		{
			name: "any path covers dot",
			inputs: []plan.Layer{
				{Step: "build", Filter: plan.Filter{Include: []string{"/root/.cache", "."}}},
			},
			stepName: "build",
			path:     ".",
			expected: true,
		},
		{
			name: "different step name",
			inputs: []plan.Layer{
				{Step: "install", Filter: plan.Filter{Include: []string{"."}}},
			},
			stepName: "build",
			path:     ".",
			expected: false,
		},
		{
			name: "specific path match",
			inputs: []plan.Layer{
				{Step: "build", Filter: plan.Filter{Include: []string{"/app/node_modules"}}},
			},
			stepName: "build",
			path:     "/app/node_modules",
			expected: true,
		},
		{
			name: "specific path no match",
			inputs: []plan.Layer{
				{Step: "build", Filter: plan.Filter{Include: []string{"/app/node_modules"}}},
			},
			stepName: "build",
			path:     "/app/dist",
			expected: false,
		},
		{
			name: "specific path does not cover dot",
			inputs: []plan.Layer{
				{Step: "build", Filter: plan.Filter{Include: []string{"/app/node_modules"}}},
			},
			stepName: "build",
			path:     ".",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewDeployBuilder()
			builder.DeployInputs = tt.inputs
			result := builder.HasIncludeForStep(tt.stepName, tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}
