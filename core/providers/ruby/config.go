package ruby

import "github.com/invopop/jsonschema"

type RubyConfig struct {
	Version string `json:"version,omitempty" jsonschema:"description=The Ruby version to install"`
}

func (*RubyProvider) ConfigSchema() *jsonschema.Schema {
	reflector := jsonschema.Reflector{DoNotReference: true, Anonymous: true}
	return reflector.Reflect(&RubyConfig{})
}
