package java

import "github.com/invopop/jsonschema"

type JavaConfig struct {
	Version       string `json:"version,omitempty" jsonschema:"description=The JDK version to install"`
	GradleVersion string `json:"gradleVersion,omitempty" jsonschema:"description=The Gradle version to install"`
}

func (*JavaProvider) ConfigSchema() *jsonschema.Schema {
	reflector := jsonschema.Reflector{DoNotReference: true, Anonymous: true}
	return reflector.Reflect(&JavaConfig{})
}
