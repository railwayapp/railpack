package config

import (
	"encoding/json"
	"fmt"

	"github.com/invopop/jsonschema"
	"github.com/railwayapp/railpack/core/plan"
)

// Keep config shorthand separate from the filters embedded in build-plan layers.
type DeployOutputs []plan.Filter

// Normalize path strings and filter objects into the same build-plan representation.
func (d *DeployOutputs) UnmarshalJSON(data []byte) error {
	var entries []json.RawMessage
	if err := json.Unmarshal(data, &entries); err != nil {
		return err
	}

	// Preserve nil (use generated outputs) versus [] (disable this step's outputs).
	var filters DeployOutputs
	if entries != nil {
		filters = make(DeployOutputs, len(entries))
	}
	for index, entry := range entries {
		var value any
		if err := json.Unmarshal(entry, &value); err != nil {
			return err
		}

		switch value := value.(type) {
		case string:
			filters[index] = plan.NewIncludeFilter([]string{value})
		case map[string]any:
			// Decode the object into Filter to validate the include/exclude field types.
			if err := json.Unmarshal(entry, &filters[index]); err != nil {
				return fmt.Errorf("deployOutputs[%d]: %w", index, err)
			}
		default:
			return fmt.Errorf("deployOutputs[%d]: expected a path string or filter object", index)
		}
	}
	*d = filters
	return nil
}

// Let config editors validate shorthand without changing the schema for plan.Filter.
func (DeployOutputs) JSONSchema() *jsonschema.Schema {
	reflector := jsonschema.Reflector{Anonymous: true, DoNotReference: true}
	return &jsonschema.Schema{
		Type: "array",
		Items: &jsonschema.Schema{
			OneOf: []*jsonschema.Schema{
				{Type: "string", Description: "Files or directories to include"},
				reflector.Reflect(&plan.Filter{}),
			},
		},
	}
}
