package app

import (
	"fmt"
	"os"
	"regexp"
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

// FromEnvs collects variables from the given environment variable names
func FromEnvs(envs []string) (*Environment, error) {
	env := NewEnvironment(nil)
	re := regexp.MustCompile(`([A-Za-z0-9_+\-]*)(?:=?)(.*)`)

	for _, e := range envs {
		matches := re.FindStringSubmatch(e)
		if len(matches) < 3 {
			continue
		}

		name := matches[1]
		value := matches[2]

		if value == "" {
			// No value, pull from current environment
			if v, ok := os.LookupEnv(name); ok {
				env.SetVariable(name, v)
			}
		} else {
			// Use provided name, value pair
			env.SetVariable(name, value)
		}
	}

	return env, nil
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
