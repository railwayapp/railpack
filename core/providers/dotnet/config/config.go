package config

type DotnetConfig struct {
	Version string `json:"version,omitempty" jsonschema:"description=Override the .NET version for the dotnet provider"`
}
