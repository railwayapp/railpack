package config

type ShellConfig struct {
	Script string `json:"script,omitempty" jsonschema:"description=Specify which shell script to run for the shell provider"`
}
