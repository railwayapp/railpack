package rust

import "github.com/invopop/jsonschema"

type RustConfig struct {
	Version        string `json:"version,omitempty" jsonschema:"description=The Rust toolchain version to install"`
	Bin            string `json:"bin,omitempty" jsonschema:"description=The Cargo binary to run when a project has multiple binaries"`
	CargoWorkspace string `json:"cargoWorkspace,omitempty" jsonschema:"description=The Cargo workspace package to build and run"`
}

func (*RustProvider) ConfigSchema() *jsonschema.Schema {
	reflector := jsonschema.Reflector{DoNotReference: true, Anonymous: true}
	return reflector.Reflect(&RustConfig{})
}
