package config

type ElixirConfig struct {
	Version       string `json:"version,omitempty" jsonschema:"description=Override the Elixir version for the elixir provider"`
	ErlangVersion string `json:"erlangVersion,omitempty" jsonschema:"description=Override the Erlang version for the elixir provider"`
}
