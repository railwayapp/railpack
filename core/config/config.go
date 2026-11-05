package config

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/invopop/jsonschema"
	"github.com/railwayapp/railpack/core/app"
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
	Provider         *string                    `json:"provider,omitempty" jsonschema:"enum=php,enum=golang,enum=java,enum=rust,enum=ruby,enum=elixir,enum=python,enum=deno,enum=dotnet,enum=node,enum=gleam,enum=cpp,enum=staticfile,enum=shell,description=The provider to use"`
	BuildAptPackages []string                   `json:"buildAptPackages,omitempty" jsonschema:"description=List of apt packages to install during the build step"`
	Steps            map[string]*StepConfig     `json:"steps,omitempty" jsonschema:"description=Map of step names to step definitions"`
	Deploy           *DeployConfig              `json:"deploy,omitempty" jsonschema:"description=Deploy configuration"`
	Packages         map[string]string          `json:"packages,omitempty" jsonschema:"description=Map of package name to package version"`
	Caches           map[string]*plan.Cache     `json:"caches,omitempty" jsonschema:"description=Map of cache name to cache definitions. The cache key can be referenced in an exec command"`
	Secrets          []string                   `json:"secrets,omitempty" jsonschema:"description=Secrets that should be made available to commands that have useSecrets set to true"`
	Exclude          []string                   `json:"exclude,omitempty" jsonschema:"description=File patterns to exclude from the build context (alternative to .dockerignore). Supports negation patterns starting with ! to include files that would otherwise be excluded"`
	ProviderConfigs  map[string]json.RawMessage `json:"-"`
}

func EmptyConfig() *Config {
	return &Config{
		Steps:           make(map[string]*StepConfig),
		Packages:        make(map[string]string),
		Caches:          make(map[string]*plan.Cache),
		Deploy:          &DeployConfig{},
		ProviderConfigs: make(map[string]json.RawMessage),
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

func (c *Config) ProviderConfig(name string, target any) error {
	data, ok := c.ProviderConfigs[name]
	if !ok {
		return nil
	}
	return json.Unmarshal(data, target)
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

		providerConfigs := mergeProviderConfigs(result.ProviderConfigs, config.ProviderConfigs)
		utils.MergeStructs(result, config)
		result.ProviderConfigs = providerConfigs
	}

	return result
}

func mergeProviderConfigs(dst, src map[string]json.RawMessage) map[string]json.RawMessage {
	merged := make(map[string]json.RawMessage, len(dst)+len(src))
	for name, value := range dst {
		merged[name] = value
	}

	for name, value := range src {
		merged[name] = mergeProviderConfigValue(merged[name], value)
	}

	return merged
}

func mergeProviderConfigValue(dst, src json.RawMessage) json.RawMessage {
	var dstProperties, srcProperties map[string]json.RawMessage
	if err := json.Unmarshal(dst, &dstProperties); err != nil || dstProperties == nil {
		return src
	}
	if err := json.Unmarshal(src, &srcProperties); err != nil || srcProperties == nil {
		return src
	}
	for property, propertyValue := range srcProperties {
		if existingValue, exists := dstProperties[property]; exists {
			dstProperties[property] = mergeProviderConfigValue(existingValue, propertyValue)
		} else {
			dstProperties[property] = propertyValue
		}
	}
	merged, _ := json.Marshal(dstProperties)
	return merged
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

func (c *Config) UnmarshalJSON(data []byte) error {
	type configAlias Config
	var config configAlias
	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}

	var properties map[string]json.RawMessage
	if err := json.Unmarshal(data, &properties); err != nil {
		return err
	}

	knownProperties := map[string]struct{}{
		"$schema": {}, "provider": {}, "buildAptPackages": {}, "steps": {}, "deploy": {},
		"packages": {}, "caches": {}, "secrets": {}, "exclude": {},
	}
	providerConfigs := make(map[string]json.RawMessage)
	for name, value := range properties {
		if _, known := knownProperties[name]; known {
			continue
		}
		providerConfigs[name] = value
	}

	*c = Config(config)
	c.ProviderConfigs = providerConfigs
	return nil
}

func (c Config) MarshalJSON() ([]byte, error) {
	type configAlias Config
	data, err := json.Marshal(configAlias(c))
	if err != nil {
		return nil, err
	}

	var properties map[string]json.RawMessage
	if err := json.Unmarshal(data, &properties); err != nil {
		return nil, err
	}
	for name, value := range c.ProviderConfigs {
		properties[name] = value
	}
	return json.Marshal(properties)
}

func (Config) JSONSchemaExtend(schema *jsonschema.Schema) {
	schema.Properties.Set("$schema", &jsonschema.Schema{
		Type:        "string",
		Description: "The schema for this config",
	})
}

func GetJsonSchema(providerSchemas map[string]*jsonschema.Schema) *jsonschema.Schema {
	r := jsonschema.Reflector{
		DoNotReference: true,
		Anonymous:      true,
	}

	schema := r.Reflect(&Config{})
	for name, providerSchema := range providerSchemas {
		schema.Properties.Set(name, providerSchema)
	}
	return schema
}

func (c *Config) ProviderConfigWithEnv(name string, envVariables map[string]string, target any) error {
	if err := unmarshalProviderConfigEnv(name, envVariables, target); err != nil {
		return err
	}
	return c.ProviderConfig(name, target)
}

func unmarshalProviderConfigEnv(name string, envVariables map[string]string, target any) error {
	environment := app.NewEnvironment(&envVariables)
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Pointer || targetValue.IsNil() || targetValue.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("provider config target must be a non-nil pointer to a struct")
	}

	configValue := targetValue.Elem()
	configType := configValue.Type()
	for i := 0; i < configType.NumField(); i++ {
		field := configType.Field(i)
		if field.PkgPath != "" {
			continue
		}

		jsonName := strings.Split(field.Tag.Get("json"), ",")[0]
		if jsonName == "-" {
			continue
		}
		if jsonName == "" {
			jsonName = field.Name
		}

		envValue, envName := environment.GetConfigVariable(strings.ToUpper(name) + "_" + configFieldEnvName(jsonName))
		if envName == "" {
			continue
		}

		configField := configValue.Field(i)
		encodedValue := []byte(envValue)
		if configField.Kind() == reflect.String {
			var err error
			encodedValue, err = json.Marshal(envValue)
			if err != nil {
				return fmt.Errorf("invalid provider config environment variable %s: %w", envName, err)
			}
		}
		if err := json.Unmarshal(encodedValue, configField.Addr().Interface()); err != nil {
			return fmt.Errorf("invalid provider config environment variable %s: %w", envName, err)
		}
	}
	return nil
}

func configFieldEnvName(name string) string {
	var envName strings.Builder
	for i, char := range name {
		if char == '-' || char == '_' {
			envName.WriteByte('_')
			continue
		}
		if char >= 'A' && char <= 'Z' && i > 0 {
			envName.WriteByte('_')
		}
		envName.WriteRune(char)
	}
	return strings.ToUpper(envName.String())
}
