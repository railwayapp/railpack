package config

import (
	"encoding/json"

	"github.com/invopop/jsonschema"
	"github.com/railwayapp/railpack/core/plan"
	"github.com/railwayapp/railpack/internal/utils"
)

const (
	SchemaUrl = "https://schema.railpack.com"
)

type DeployConfig struct {
	AptPackages []string          `json:"aptPackages,omitempty" jsonschema:"description=List of apt packages to include at runtime"`
	Base        *plan.Layer       `json:"base,omitempty" jsonschema:"description=The base image to use for the deploy step"`
	Inputs      []plan.Layer      `json:"inputs,omitempty" jsonschema:"description=The inputs for the deploy step"`
	StartCmd    string            `json:"startCommand,omitempty" jsonschema:"description=The command to run in the container"`
	Variables   map[string]string `json:"variables,omitempty" jsonschema:"description=The variables available to this step. The key is the name of the variable that is referenced in a variable command"`
	Paths       []string          `json:"paths,omitempty" jsonschema:"description=The paths to prepend to the $PATH environment variable"`
}

type StepConfig struct {
	plan.Step
	DeployOutputs []plan.Filter `json:"deployOutputs,omitempty" jsonschema:"description=Parts of this step that should be included in the final image. If empty, the /app directory will be used."`
}

type Config struct {
	Provider         *string                `json:"provider" jsonschema:"description=The provider to use"`
	BuildAptPackages []string               `json:"buildAptPackages,omitempty" jsonschema:"description=List of apt packages to install during the build step"`
	Steps            map[string]*StepConfig `json:"steps,omitempty" jsonschema:"description=Map of step names to step definitions"`
	Deploy           *DeployConfig          `json:"deploy,omitempty" jsonschema:"description=Deploy configuration"`
	Packages         map[string]string      `json:"packages,omitempty" jsonschema:"description=Map of package name to package version"`
	Caches           map[string]*plan.Cache `json:"caches,omitempty" jsonschema:"description=Map of cache name to cache definitions. The cache key can be referenced in an exec command"`
	Secrets          []string               `json:"secrets,omitempty" jsonschema:"description=Secrets that should be made available to commands that have useSecrets set to true"`
}

func EmptyConfig() *Config {
	return &Config{
		Steps:    make(map[string]*StepConfig),
		Packages: make(map[string]string),
		Caches:   make(map[string]*plan.Cache),
		Deploy:   &DeployConfig{Variables: make(map[string]string)},
	}
}

func (c *Config) GetOrCreateStep(name string) *StepConfig {
	if existingStep, exists := c.Steps[name]; exists {
		return existingStep
	}

	step := &StepConfig{
		Step: *plan.NewStep(name),
	}
	c.Steps[name] = step

	return step
}

// Merge combines multiple configs by merging their values with later configs taking precedence
func Merge(configs ...*Config) *Config {
	if len(configs) == 0 {
		return EmptyConfig()
	}

	result := EmptyConfig()
	for _, config := range configs {
		if config == nil {
			continue
		}

		utils.MergeStructs(result, config)
	}

	return result
}

func (s *StepConfig) UnmarshalJSON(data []byte) error {
	var temp struct {
		DeployOutputs []plan.Filter `json:"deployOutputs,omitempty"`
	}
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}
	s.DeployOutputs = temp.DeployOutputs

	return s.Step.UnmarshalJSON(data)
}

func (Config) JSONSchemaExtend(schema *jsonschema.Schema) {
	schema.Properties.Set("$schema", &jsonschema.Schema{
		Type:        "string",
		Description: "The schema for this config",
	})
}

func GetJsonSchema() *jsonschema.Schema {
	r := jsonschema.Reflector{
		DoNotReference: true,
		Anonymous:      true,
	}

	schema := r.Reflect(&Config{})
	return schema
}
