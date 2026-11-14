package gleam

import "github.com/invopop/jsonschema"

type GleamConfig struct {
	IncludeSource *bool `json:"includeSource,omitempty" jsonschema:"description=Include source files in the runtime image"`
}

func (*GleamProvider) ConfigSchema() *jsonschema.Schema {
	reflector := jsonschema.Reflector{DoNotReference: true, Anonymous: true}
	return reflector.Reflect(&GleamConfig{})
}
