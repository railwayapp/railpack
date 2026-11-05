package dotnet

import "github.com/invopop/jsonschema"

type DotnetConfig struct {
	Version string `json:"version,omitempty" jsonschema:"description=The .NET SDK version to install"`
}

func (*DotnetProvider) ConfigSchema() *jsonschema.Schema {
	reflector := jsonschema.Reflector{DoNotReference: true, Anonymous: true}
	return reflector.Reflect(&DotnetConfig{})
}
