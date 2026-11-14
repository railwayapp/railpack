package elixir

import "github.com/invopop/jsonschema"

type ElixirConfig struct {
	Version       string `json:"version,omitempty" jsonschema:"description=The Elixir version to install"`
	ErlangVersion string `json:"erlangVersion,omitempty" jsonschema:"description=The Erlang/OTP version to install"`
}

func (*ElixirProvider) ConfigSchema() *jsonschema.Schema {
	reflector := jsonschema.Reflector{DoNotReference: true, Anonymous: true}
	return reflector.Reflect(&ElixirConfig{})
}
