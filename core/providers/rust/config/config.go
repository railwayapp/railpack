package config

type RustConfig struct {
	Version   string `json:"version,omitempty" jsonschema:"description=Override the Rust version for the rust provider"`
	Bin       string `json:"bin,omitempty" jsonschema:"description=Specify which binary to start for the rust provider"`
	Workspace string `json:"workspace,omitempty" jsonschema:"description=Specify which Cargo workspace package to build for the rust provider"`
}
