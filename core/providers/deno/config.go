package deno

import "github.com/invopop/jsonschema"

type DenoConfig struct {
	Version string `json:"version,omitempty" jsonschema:"description=The Deno version to install"`
}

func (*DenoProvider) ConfigSchema() *jsonschema.Schema {
	reflector := jsonschema.Reflector{DoNotReference: true, Anonymous: true}
	return reflector.Reflect(&DenoConfig{})
}
