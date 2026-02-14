package config

type StaticfileConfig struct {
	Root string `json:"root,omitempty" jsonschema:"description=Override the root directory served by the staticfile provider"`
}
