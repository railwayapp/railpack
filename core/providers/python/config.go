package python

import "github.com/invopop/jsonschema"

type PythonConfig struct {
	Version string `json:"version,omitempty" jsonschema:"description=The Python version to install"`
}

func (*PythonProvider) ConfigSchema() *jsonschema.Schema {
	reflector := jsonschema.Reflector{DoNotReference: true, Anonymous: true}
	return reflector.Reflect(&PythonConfig{})
}
