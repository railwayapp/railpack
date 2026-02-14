package config

type DenoConfig struct {
	Version string `json:"version,omitempty" jsonschema:"description=Override the Deno version for the deno provider"`
}
