package config

type NodeConfig struct {
	Version         string   `json:"version,omitempty" jsonschema:"description=Override the Node.js version for the node provider"`
	BunVersion      string   `json:"bunVersion,omitempty" jsonschema:"description=Override the Bun version for the node provider"`
	NoSpa           bool     `json:"noSpa,omitempty" jsonschema:"description=Disable SPA mode for the node provider"`
	SpaOutputDir    string   `json:"spaOutputDir,omitempty" jsonschema:"description=Specify the SPA output directory for the node provider"`
	PruneDeps       bool     `json:"pruneDeps,omitempty" jsonschema:"description=Remove development dependencies for the node provider"`
	PruneCmd        string   `json:"pruneCmd,omitempty" jsonschema:"description=Specify a custom prune command for the node provider"`
	InstallPatterns []string `json:"installPatterns,omitempty" jsonschema:"description=Additional file patterns to include during dependency installation for the node provider"`
	AngularProject  string   `json:"angularProject,omitempty" jsonschema:"description=Specify the Angular project name for the node provider"`
}
