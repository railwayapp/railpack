package plan

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/invopop/jsonschema"
)

type Input struct {
	Image   string   `json:"image,omitempty" jsonschema:"description=The image to use as input"`
	Step    string   `json:"step,omitempty" jsonschema:"description=The step to use as input"`
	Local   bool     `json:"local,omitempty" jsonschema:"description=Whether to use local files as input"`
	Spread  bool     `json:"spread,omitempty" jsonschema:"description=Whether to spread the input"`
	Include []string `json:"include,omitempty" jsonschema:"description=Files or directories to include"`
	Exclude []string `json:"exclude,omitempty" jsonschema:"description=Files or directories to exclude"`
}

type InputOptions struct {
	Include []string
	Exclude []string
}

func NewStepInput(stepName string, options ...InputOptions) Input {
	input := Input{
		Step: stepName,
	}

	if len(options) > 0 {
		input.Include = options[0].Include
		input.Exclude = options[0].Exclude
	}

	return input
}

func NewImageInput(image string, options ...InputOptions) Input {
	input := Input{
		Image: image,
	}

	if len(options) > 0 {
		input.Include = options[0].Include
		input.Exclude = options[0].Exclude
	}
	return input
}

func NewLocalInput(path string) Input {
	return Input{
		Local:   true,
		Include: []string{path},
	}
}

func (i *Input) String() string {
	bytes, _ := json.Marshal(i)
	return string(bytes)
}

func (i *Input) DisplayName() string {
	include := strings.Join(i.Include, ", ")

	if i.Local {
		return fmt.Sprintf("local %s", include)
	}

	if i.Spread {
		return fmt.Sprintf("spread %s", include)
	}

	if i.Step != "" {
		return fmt.Sprintf("$%s", i.Step)
	}

	if i.Image != "" {
		return i.Image
	}

	return fmt.Sprintf("input %s", include)
}

func (i Input) IsSpread() bool {
	return i.Spread
}

func (i *Input) UnmarshalJSON(data []byte) error {
	// First try normal JSON unmarshal
	type Alias Input
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(i),
	}
	if err := json.Unmarshal(data, &aux); err == nil {
		return nil
	}

	str := string(data)

	str = strings.Trim(str, "\"")
	switch str {
	case ".":
		*i = NewLocalInput(".")
		return nil
	case "...":
		*i = Input{Spread: true}
		return nil
	default:
		if strings.HasPrefix(str, "$") {
			stepName := strings.TrimPrefix(str, "$")
			*i = NewStepInput(stepName)
			return nil
		}
		return fmt.Errorf("invalid input format: %s", str)
	}
}

func (Input) JSONSchema() *jsonschema.Schema {
	// Create common schemas for include/exclude
	includeSchema := &jsonschema.Schema{
		Type:        "array",
		Description: "Files or directories to include",
		Items: &jsonschema.Schema{
			Type: "string",
		},
	}
	excludeSchema := &jsonschema.Schema{
		Type:        "array",
		Description: "Files or directories to exclude",
		Items: &jsonschema.Schema{
			Type: "string",
		},
	}

	// Step input schema
	stepSchema := &jsonschema.Schema{
		Type:       "object",
		Properties: jsonschema.NewProperties(),
	}
	stepSchema.Properties.Set("step", &jsonschema.Schema{
		Type:        "string",
		Description: "The step to use as input",
	})
	stepSchema.Properties.Set("include", includeSchema)
	stepSchema.Properties.Set("exclude", excludeSchema)
	stepSchema.Required = []string{"step"}

	// Image input schema
	imageSchema := &jsonschema.Schema{
		Type:       "object",
		Properties: jsonschema.NewProperties(),
	}
	imageSchema.Properties.Set("image", &jsonschema.Schema{
		Type:        "string",
		Description: "The image to use as input",
	})
	imageSchema.Properties.Set("include", includeSchema)
	imageSchema.Properties.Set("exclude", excludeSchema)
	imageSchema.Required = []string{"image"}

	// Local input schema
	localSchema := &jsonschema.Schema{
		Type:       "object",
		Properties: jsonschema.NewProperties(),
	}
	localSchema.Properties.Set("local", &jsonschema.Schema{
		Type:        "boolean",
		Description: "Whether to use local files as input",
	})
	localSchema.Properties.Set("include", includeSchema)
	localSchema.Properties.Set("exclude", excludeSchema)
	localSchema.Required = []string{"local"}

	// String input schema
	stringSchema := &jsonschema.Schema{
		Type:        "string",
		Description: "Strings will be parsed and interpreted as an input. Valid formats are: '.', '...', or '$step'",
		Enum:        []interface{}{".", "...", "$step"},
	}

	availableInputs := []*jsonschema.Schema{stepSchema, imageSchema, localSchema, stringSchema}

	return &jsonschema.Schema{
		OneOf: availableInputs,
	}
}
