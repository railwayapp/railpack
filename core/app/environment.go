package app

import (
	"fmt"
	"os"
	"strings"
)

type Environment struct {
	Variables map[string]string
}

func NewEnvironment(variables *map[string]string) *Environment {
	if variables == nil {
		variables = &map[string]string{}
	}

	return &Environment{Variables: *variables}
}

// ParseKeyValueStrings parses CLI arguments in NAME or NAME=VALUE form. A
// name-only entry resolves its value from the current process environment.
func ParseKeyValueStrings(pairs []string) (map[string]string, error) {
	vars := make(map[string]string)

	for _, pair := range pairs {
		name, value, hasValue := strings.Cut(pair, "=")
		if hasValue {
			vars[name] = value
			continue
		}

		if value, ok := os.LookupEnv(name); ok {
			vars[name] = value
		}
	}

	return vars, nil
}

// GetVariable returns the value of the given variable name
func (e *Environment) GetVariable(name string) string {
	return e.Variables[name]
}

// SetVariable stores a variable in the Environment
func (e *Environment) SetVariable(name, value string) {
	e.Variables[name] = value
}

// ConfigVariable returns the RAILPACK_ prefixed version of a variable name
func (e *Environment) ConfigVariable(name string) string {
	return fmt.Sprintf("RAILPACK_%s", name)
}

// returns the value of a RAILPACK_ prefixed variable with newlines removed
// Returns both the value and the name of the config variable
func (e *Environment) GetConfigVariable(name string) (string, string) {
	configVar := e.ConfigVariable(name)

	if val, exists := e.Variables[configVar]; exists {
		return strings.TrimSpace(val), configVar
	}
	return "", ""
}

// GetConfigVariableList returns a space-separated config variable as a list
// Returns both the list and the name of the config variable
func (e *Environment) GetConfigVariableList(name string) ([]string, string) {
	val, configVar := e.GetConfigVariable(name)
	if val == "" {
		return nil, ""
	}
	return strings.Split(val, " "), configVar
}

// checks if a RAILPACK_ prefixed variable is set to "1" or "true"
func (e *Environment) IsConfigVariableTruthy(name string) bool {
	if val, _ := e.GetConfigVariable(name); val != "" {
		lowerVal := strings.ToLower(val)
		return lowerVal == "1" || lowerVal == "true"
	}
	return false
}

// GetSecretsWithPrefix returns all secrets that have the given prefix
func (e *Environment) GetSecretsWithPrefix(prefix string) []string {
	secrets := []string{}
	for secretName := range e.Variables {
		if strings.HasPrefix(secretName, prefix) {
			secrets = append(secrets, secretName)
		}
	}
	return secrets
}
