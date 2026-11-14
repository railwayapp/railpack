package node

import "github.com/invopop/jsonschema"

type NodeConfig struct {
	Version string `json:"version,omitempty" jsonschema:"description=The Node.js version to install"`
}

func (*NodeProvider) ConfigSchema() *jsonschema.Schema {
	reflector := jsonschema.Reflector{DoNotReference: true, Anonymous: true}
	return reflector.Reflect(&NodeConfig{})
}
