package config

type GolangConfig struct {
	Version         string `json:"version,omitempty" jsonschema:"description=Override the Go version for the golang provider"`
	Bin             string `json:"bin,omitempty" jsonschema:"description=Specify which command in cmd/ to build for the golang provider"`
	WorkspaceModule string `json:"workspaceModule,omitempty" jsonschema:"description=Specify which workspace module to build for the golang provider"`
	CgoEnabled      bool   `json:"cgoEnabled,omitempty" jsonschema:"description=Enable CGO for non-static binary compilation in the golang provider"`
}
