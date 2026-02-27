package config

type PhpConfig struct {
	Extensions []string `json:"extensions,omitempty" jsonschema:"description=Additional PHP extensions to install for the php provider"`
}
