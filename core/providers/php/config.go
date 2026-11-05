package php

import "github.com/invopop/jsonschema"

type PhpConfig struct {
	Version string `json:"version,omitempty" jsonschema:"description=The PHP version to install"`
	RootDir string `json:"rootDir,omitempty" jsonschema:"description=The document root directory to serve"`
}

func (*PhpProvider) ConfigSchema() *jsonschema.Schema {
	reflector := jsonschema.Reflector{DoNotReference: true, Anonymous: true}
	return reflector.Reflect(&PhpConfig{})
}
