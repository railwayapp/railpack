package config

type GleamConfig struct {
	IncludeSource bool `json:"includeSource,omitempty" jsonschema:"description=Include the project source tree in the final image for the gleam provider"`
}
