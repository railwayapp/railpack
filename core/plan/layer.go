package plan

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/invopop/jsonschema"
)

type Layer struct {
	Image  string `json:"image,omitempty" jsonschema:"description=The image to use as input"`
	Step   string `json:"step,omitempty" jsonschema:"description=The step to use as input"`
	Local  bool   `json:"local,omitempty" jsonschema:"description=Whether to use local files as input"`
	Spread bool   `json:"spread,omitempty" jsonschema:"description=Whether to spread the input"`

	Filter
}

func NewStepLayer(stepName string, filter ...Filter) Layer {
	input := Layer{
		Step: stepName,
	}

	if len(filter) > 0 {
		input.Include = filter[0].Include
		input.Exclude = filter[0].Exclude
	}

	return input
}

func NewImageLayer(image string, filter ...Filter) Layer {
	input := Layer{
		Image: image,
	}

	if len(filter) > 0 {
		input.Include = filter[0].Include
		input.Exclude = filter[0].Exclude
	}

	return input
}

func NewLocalLayer() Layer {
	return Layer{
		Local:  true,
		Filter: NewIncludeFilter([]string{"."}),
	}
}

func (i Layer) IsEmpty() bool {
	return i.Step == "" && i.Image == "" && !i.Local && !i.Spread
}

func (i Layer) IsSpread() bool {
	return i.Spread
}

func (i *Layer) String() string {
	bytes, _ := json.Marshal(i)
	return string(bytes)
}

func (i *Layer) DisplayName() string {
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

// Supports two types of inputs:
//
// Object Notation:
//   - Step Layer: {"step": "build", "include": ["src/**/*.go"], "exclude": ["*_test.go"]}
//     References the output of a named build step with optional file filtering
//   - Image Layer: {"image": "golang:1.21", "include": ["."], "exclude": ["tmp"]}
//     Uses a Docker image as input with optional file filtering
//   - Local Layer: {"local": true, "include": ["src"], "exclude": ["node_modules"]}
//     Uses local files from the build context with filtering
//
// String Shortcuts: ".", "...", "$stepname"
func (i *Layer) UnmarshalJSON(data []byte) error {
	// First try normal JSON unmarshal for object notation
	type Alias Layer
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(i),
	}
	if err := json.Unmarshal(data, &aux); err == nil {
		return nil
	}

	// If object unmarshaling fails, try string shortcuts
	str := string(data)

	// Remove quotes from JSON string
	str = strings.Trim(str, "\"")
	switch str {
	case ".":
		// "." represents a local layer with current directory
		*i = NewLocalLayer()
		return nil
	case "...":
		// Creates a spread layer that expands to include all previous layers' files
		*i = Layer{Spread: true}
		return nil
	default:
		// "$stepname" represents a reference to another step
		if strings.HasPrefix(str, "$") {
			stepName := strings.TrimPrefix(str, "$")
			*i = NewStepLayer(stepName)
			return nil
		}
		return fmt.Errorf("invalid input format: %s", str)
	}
}

func (Layer) JSONSchema() *jsonschema.Schema {
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
		Enum:        []interface{}{".", "..."},
	}

	availableInputs := []*jsonschema.Schema{stepSchema, imageSchema, localSchema, stringSchema}

	return &jsonschema.Schema{
		OneOf: availableInputs,
	}
}
