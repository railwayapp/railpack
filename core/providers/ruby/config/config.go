package config

type RubyConfig struct {
	Version string `json:"version,omitempty" jsonschema:"description=Override the Ruby version for the ruby provider"`
}
