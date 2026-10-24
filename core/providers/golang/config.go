package golang

import "github.com/invopop/jsonschema"

type GoConfig struct {
	Version string `json:"version,omitempty" jsonschema:"description=The Go version to install"`
}

func (*GoProvider) ConfigSchema() *jsonschema.Schema {
	reflector := jsonschema.Reflector{DoNotReference: true, Anonymous: true}
	return reflector.Reflect(&GoConfig{})
}
