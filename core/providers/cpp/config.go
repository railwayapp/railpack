package cpp

import "github.com/invopop/jsonschema"

type CppConfig struct {
	CmakeVersion string `json:"cmakeVersion,omitempty" jsonschema:"description=The CMake version to install"`
	MesonVersion string `json:"mesonVersion,omitempty" jsonschema:"description=The Meson version to install"`
	NinjaVersion string `json:"ninjaVersion,omitempty" jsonschema:"description=The Ninja version to install"`
}

func (*CppProvider) ConfigSchema() *jsonschema.Schema {
	reflector := jsonschema.Reflector{DoNotReference: true, Anonymous: true}
	return reflector.Reflect(&CppConfig{})
}
