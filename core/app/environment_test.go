package app

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseKeyValueStrings(t *testing.T) {
	t.Setenv("FROM_ENV", "resolved")

	variables, err := ParseKeyValueStrings([]string{
		"VAR1=value1",
		"VAR2=value2",
		"RAILPACK_APT_PACKAGES=apt1,apt2",
		"COMMA=this has, a comma",
		"RAILPACK_TRUTHY_CASE=True ",
		"RAILPACK_TRUTHY_INT_CASE= 1 ",
		"HELLO+WORLD=boop",
		"FROM_ENV",
		"EMPTY=",
		"MISSING",
	})
	env := NewEnvironment(&variables)

	require.NoError(t, err)
	require.Equal(t, env.GetVariable("VAR1"), "value1")
	require.Equal(t, env.GetVariable("VAR2"), "value2")
	require.Equal(t, env.GetVariable("RAILPACK_APT_PACKAGES"), "apt1,apt2")
	require.Equal(t, env.GetVariable("COMMA"), "this has, a comma")
	require.Equal(t, env.IsConfigVariableTruthy("TRUTHY_CASE"), true)
	require.Equal(t, env.IsConfigVariableTruthy("TRUTHY_INT_CASE"), true)
	require.Equal(t, env.GetVariable("HELLO+WORLD"), "boop")
	require.Equal(t, env.GetVariable("FROM_ENV"), "resolved")
	require.Contains(t, env.Variables, "EMPTY")
	require.Empty(t, env.GetVariable("EMPTY"))
	require.NotContains(t, env.Variables, "MISSING")
}

func TestEnvironmentVariables(t *testing.T) {
	env := NewEnvironment(nil)

	require.Empty(t, env.GetVariable("MISSING"))
	env.SetVariable("NAME", "value")
	require.Equal(t, "value", env.GetVariable("NAME"))
	require.Equal(t, "RAILPACK_NAME", env.ConfigVariable("NAME"))
}

func TestGetConfigVariable(t *testing.T) {
	envVars := map[string]string{
		"RAILPACK_PACKAGES": " \n node@22 jq@latest \t",
		"RAILPACK_EMPTY":    " \n\t ",
	}
	env := NewEnvironment(&envVars)

	tests := []struct {
		name               string
		variable           string
		expectedValue      string
		expectedConfigName string
	}{
		{
			name:               "trims surrounding whitespace",
			variable:           "PACKAGES",
			expectedValue:      "node@22 jq@latest",
			expectedConfigName: "RAILPACK_PACKAGES",
		},
		{
			name:               "preserves the name of an empty variable",
			variable:           "EMPTY",
			expectedValue:      "",
			expectedConfigName: "RAILPACK_EMPTY",
		},
		{
			name:     "returns empty values for a missing variable",
			variable: "MISSING",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, configName := env.GetConfigVariable(tt.variable)

			require.Equal(t, tt.expectedValue, value)
			require.Equal(t, tt.expectedConfigName, configName)
		})
	}
}

func TestGetConfigVariableList(t *testing.T) {
	envVars := map[string]string{
		"RAILPACK_PACKAGES": " node@22 jq@latest ",
		"RAILPACK_SINGLE":   "python",
		"RAILPACK_EMPTY":    "  ",
	}
	env := NewEnvironment(&envVars)

	tests := []struct {
		name               string
		variable           string
		expectedValues     []string
		expectedConfigName string
	}{
		{
			name:               "splits a space-separated list",
			variable:           "PACKAGES",
			expectedValues:     []string{"node@22", "jq@latest"},
			expectedConfigName: "RAILPACK_PACKAGES",
		},
		{
			name:               "returns a single item",
			variable:           "SINGLE",
			expectedValues:     []string{"python"},
			expectedConfigName: "RAILPACK_SINGLE",
		},
		{
			name:     "returns no list for an empty variable",
			variable: "EMPTY",
		},
		{
			name:     "returns no list for a missing variable",
			variable: "MISSING",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values, configName := env.GetConfigVariableList(tt.variable)

			require.Equal(t, tt.expectedValues, values)
			require.Equal(t, tt.expectedConfigName, configName)
		})
	}
}

func TestIsConfigVariableTruthy(t *testing.T) {
	envVars := map[string]string{
		"RAILPACK_TRUE":        "true",
		"RAILPACK_TRUE_CASE":   " True ",
		"RAILPACK_TRUE_INT":    "\n1\t",
		"RAILPACK_FALSE":       "false",
		"RAILPACK_FALSE_INT":   "0",
		"RAILPACK_OTHER_VALUE": "yes",
		"RAILPACK_EMPTY":       "  ",
	}
	env := NewEnvironment(&envVars)

	tests := []struct {
		name     string
		variable string
		expected bool
	}{
		{name: "true", variable: "TRUE", expected: true},
		{name: "mixed case and whitespace", variable: "TRUE_CASE", expected: true},
		{name: "integer true and whitespace", variable: "TRUE_INT", expected: true},
		{name: "false", variable: "FALSE", expected: false},
		{name: "integer false", variable: "FALSE_INT", expected: false},
		{name: "unsupported truthy value", variable: "OTHER_VALUE", expected: false},
		{name: "empty", variable: "EMPTY", expected: false},
		{name: "missing", variable: "MISSING", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, env.IsConfigVariableTruthy(tt.variable))
		})
	}
}

func TestGetSecretsWithPrefix(t *testing.T) {
	envVars := map[string]string{
		"SECRET_":        "exact prefix",
		"SECRET_ONE":     "one",
		"SECRET_TWO":     "two",
		"SECRETISH":      "not a match",
		"OTHER_SECRET":   "not a match",
		"RAILPACK_TOKEN": "not a match",
	}
	env := NewEnvironment(&envVars)

	require.ElementsMatch(t, []string{
		"SECRET_",
		"SECRET_ONE",
		"SECRET_TWO",
	}, env.GetSecretsWithPrefix("SECRET_"))
}
