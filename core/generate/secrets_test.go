package generate

import (
	"encoding/json"
	"testing"

	"github.com/railwayapp/railpack/core/config"
	"github.com/railwayapp/railpack/core/plan"
	"github.com/stretchr/testify/require"
)

func TestResolveSecrets(t *testing.T) {
	tests := []struct {
		name              string
		configuredSecrets []string
		steps             []plan.Step
		expected          []string
		expectedByStep    [][]string
	}{
		{
			name:           "promotes a step secret",
			steps:          []plan.Step{{Secrets: []string{"TOKEN"}}},
			expected:       []string{"TOKEN"},
			expectedByStep: [][]string{{"TOKEN"}},
		},
		{
			name: "promotes secrets from multiple steps",
			steps: []plan.Step{
				{Secrets: []string{"TOKEN"}},
				{Secrets: []string{"API_KEY"}},
			},
			expected:       []string{"TOKEN", "API_KEY"},
			expectedByStep: [][]string{{"TOKEN"}, {"API_KEY"}},
		},
		{
			name:              "deduplicates configured and step secrets",
			configuredSecrets: []string{"TOKEN"},
			steps: []plan.Step{
				{Secrets: []string{"TOKEN"}},
				{Secrets: []string{"TOKEN"}},
			},
			expected:       []string{"TOKEN"},
			expectedByStep: [][]string{{"TOKEN"}, {"TOKEN"}},
		},
		{
			name:              "expands selectors from configured secrets",
			configuredSecrets: []string{"GLOBAL"},
			steps: []plan.Step{
				{Secrets: []string{"*"}},
				{Secrets: []string{"...", "TOKEN"}},
			},
			expected:       []string{"GLOBAL", "TOKEN"},
			expectedByStep: [][]string{{"GLOBAL"}, {"TOKEN"}},
		},
		{
			name:              "wildcard excludes secrets inferred from other steps",
			configuredSecrets: []string{"GLOBAL"},
			steps: []plan.Step{
				{Secrets: []string{"LOCAL"}},
				{Secrets: []string{"*"}},
			},
			expected:       []string{"GLOBAL", "LOCAL"},
			expectedByStep: [][]string{{"LOCAL"}, {"GLOBAL"}},
		},
		{
			name:              "combines wildcard with an explicit step secret",
			configuredSecrets: []string{"GLOBAL"},
			steps:             []plan.Step{{Secrets: []string{"*", "LOCAL"}}},
			expected:          []string{"GLOBAL", "LOCAL"},
			expectedByStep:    [][]string{{"GLOBAL", "LOCAL"}},
		},
		{
			name: "empty and omitted step secrets contribute nothing",
			steps: []plan.Step{
				{},
				{Secrets: []string{}},
			},
			expected:       []string{},
			expectedByStep: [][]string{{}, {}},
		},
		{
			name:           "preserves wildcard when no secrets exist",
			steps:          []plan.Step{{Secrets: []string{"*"}}},
			expected:       []string{},
			expectedByStep: [][]string{{"*"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buildConfig := config.EmptyConfig()
			buildConfig.Secrets = tt.configuredSecrets
			secrets, steps := resolveSecrets(buildConfig, tt.steps)

			require.ElementsMatch(t, tt.expected, secrets)
			require.Len(t, secrets, len(tt.expected))

			for i, expected := range tt.expectedByStep {
				require.ElementsMatch(t, expected, steps[i].Secrets)
				require.Len(t, steps[i].Secrets, len(expected))
			}
		})
	}
}

func TestResolveSecretsDoesNotMutateSteps(t *testing.T) {
	buildConfig := config.EmptyConfig()
	buildConfig.Secrets = []string{"GLOBAL"}
	steps := []plan.Step{{Secrets: []string{"*"}}}

	_, resolvedSteps := resolveSecrets(buildConfig, steps)

	require.Equal(t, []string{"*"}, steps[0].Secrets)
	require.Equal(t, []string{"GLOBAL"}, resolvedSteps[0].Secrets)
}

func TestGenerateContextPromotesSecretsFromReferencedSteps(t *testing.T) {
	ctx := CreateTestContext(t, "../../examples/node-npm")
	configJSON := `{
		"steps": {
			"included": {
				"commands": ["echo included"],
				"secrets": ["INCLUDED_SECRET"]
			},
			"removed": {
				"commands": ["echo removed"],
				"secrets": ["REMOVED_SECRET"]
			}
		},
		"deploy": {
			"inputs": [{"step": "included"}]
		}
	}`

	var cfg config.Config
	require.NoError(t, json.Unmarshal([]byte(configJSON), &cfg))
	ctx.Config = &cfg

	buildPlan, _, err := ctx.Generate()
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"INCLUDED_SECRET"}, buildPlan.Secrets)
	require.NotContains(t, buildPlan.Secrets, "REMOVED_SECRET")

	for _, step := range buildPlan.Steps {
		require.NotEqual(t, "removed", step.Name)
	}
}
