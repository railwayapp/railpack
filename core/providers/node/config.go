package node

import "github.com/invopop/jsonschema"

type NodeConfig struct {
	Version        string         `json:"version,omitempty" jsonschema:"description=The Node.js version to install"`
	PackageManager PackageManager `json:"packageManager,omitempty" jsonschema:"enum=npm,enum=pnpm,enum=yarn,enum=bun,description=The package manager to use"`
}

func (*NodeProvider) ConfigSchema() *jsonschema.Schema {
	reflector := jsonschema.Reflector{DoNotReference: true, Anonymous: true}
	return reflector.Reflect(&NodeConfig{})
}
