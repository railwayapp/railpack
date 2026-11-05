package zig

import "github.com/invopop/jsonschema"

type ZigConfig struct {
	Version string `json:"version,omitempty" jsonschema:"description=The Zig version to install"`
}

func (*ZigProvider) ConfigSchema() *jsonschema.Schema {
	reflector := jsonschema.Reflector{DoNotReference: true, Anonymous: true}
	return reflector.Reflect(&ZigConfig{})
}
