package config

import (
	"encoding/json"
	"testing"

	"github.com/railwayapp/railpack/core/plan"
	"github.com/stretchr/testify/require"
)

func TestDeployOutputsDecodePathsAndFilters(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected DeployOutputs
	}{
		{"shorthand", `["dist"]`, DeployOutputs{plan.NewIncludeFilter([]string{"dist"})}},
		{"escaped path", `["build/\"quoted\""]`, DeployOutputs{plan.NewIncludeFilter([]string{`build/"quoted"`})}},
		{"object", `[{"include":["dist","public"],"exclude":["**/*.map"]}]`, DeployOutputs{plan.NewFilter([]string{"dist", "public"}, []string{"**/*.map"})}},
		{"exclude only", `[{"exclude":["tmp"]}]`, DeployOutputs{plan.NewFilter(nil, []string{"tmp"})}},
		{"empty object", `[{}]`, DeployOutputs{{}}},
		{"disable outputs", `[]`, DeployOutputs{}},
		{"generated outputs", `null`, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var outputs DeployOutputs

			err := json.Unmarshal([]byte(tt.input), &outputs)

			require.NoError(t, err)
			require.Equal(t, tt.expected, outputs)

			// Both input forms serialize to build-plan objects and retain empty/nil semantics.
			encoded, err := json.Marshal(outputs)

			require.NoError(t, err)

			var decoded DeployOutputs

			err = json.Unmarshal(encoded, &decoded)

			require.NoError(t, err)
			require.Equal(t, outputs, decoded)
		})
	}
}

func TestDeployOutputsRejectInvalidValues(t *testing.T) {
	for _, input := range []string{`"dist"`, `{}`, `[null]`, `[true]`, `[42]`, `[[]]`, `[{"include":"dist"}]`, `[{"exclude":[1]}]`} {
		t.Run(input, func(t *testing.T) {
			var outputs DeployOutputs

			err := json.Unmarshal([]byte(input), &outputs)

			require.Error(t, err)
		})
	}
}

func TestDeployOutputsShorthandReplacesPreviousFilters(t *testing.T) {
	// This old exclusion would hide dist's contents if shorthand retained the previous filter.
	outputs := DeployOutputs{plan.NewFilter([]string{"old"}, []string{"dist/**"})}

	err := json.Unmarshal([]byte(`["dist"]`), &outputs)

	// The shorthand replaces the whole output list, leaving no stale exclusions.
	require.NoError(t, err)
	require.Equal(t, DeployOutputs{plan.NewIncludeFilter([]string{"dist"})}, outputs)
}
