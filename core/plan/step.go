package plan

import (
	"encoding/json"

	"github.com/invopop/jsonschema"
)

type Step struct {
	Name      string            `json:"name,omitempty" jsonschema:"description=The name of the step"`
	Inputs    []Layer           `json:"inputs,omitempty" jsonschema:"description=The inputs for this step"`
	Commands  []Command         `json:"commands,omitempty" jsonschema:"description=The commands to run in this step"`
	Secrets   []string          `json:"secrets,omitempty" jsonschema:"description=The secrets that this step uses"`
	Assets    map[string]string `json:"assets,omitempty" jsonschema:"description=The assets available to this step. The key is the name of the asset that is referenced in a file command"`
	Variables map[string]string `json:"variables,omitempty" jsonschema:"description=The variables available to this step. The key is the name of the variable that is referenced in a variable command"`
	Caches    []string          `json:"caches,omitempty" jsonschema:"description=The caches available to all commands in this step. Each cache must refer to a cache at the top level of the plan"`
}

func NewStep(name string) *Step {
	return &Step{
		Name:      name,
		Assets:    make(map[string]string),
		Variables: make(map[string]string),
		Secrets:   []string{"*"}, // default to using all secrets
	}
}

func (s *Step) AddCommands(commands []Command) {
	if s.Commands == nil {
		s.Commands = []Command{}
	}
	s.Commands = append(s.Commands, commands...)
}

func (s *Step) UnmarshalJSON(data []byte) error {
	type Alias Step
	aux := &struct {
		Commands *[]json.RawMessage `json:"commands"`
		*Alias
	}{
		Alias: (*Alias)(s),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.Commands != nil {
		s.Commands = []Command{}
		for _, rawCmd := range *aux.Commands {
			cmd, err := UnmarshalCommand(rawCmd)
			if err != nil {
				return err
			}
			s.Commands = append(s.Commands, cmd)
		}
	}

	return nil
}

func (Step) JSONSchemaExtend(schema *jsonschema.Schema) {
	// Remove name from the schema
	var required []string
	for _, prop := range schema.Required {
		if prop != "name" {
			required = append(required, prop)
		}
	}
	schema.Required = required
	schema.Properties.Delete("name")

	// Add proper schemas for the commands
	var commandsDescription string
	if currCommandsSchema, ok := schema.Properties.Get("commands"); ok {
		commandsDescription = currCommandsSchema.Description
	}

	commandSchema := &jsonschema.Schema{
		Type:        "array",
		Description: commandsDescription,
		Items:       CommandsSchema(),
	}

	schema.Properties.Set("commands", commandSchema)
}

func CommandsSchema() *jsonschema.Schema {
	execSchema := generateSchemaWithComments(ExecCommand{})
	pathSchema := generateSchemaWithComments(PathCommand{})
	copySchema := generateSchemaWithComments(CopyCommand{})
	fileSchema := generateSchemaWithComments(FileCommand{})

	availableCommands := []*jsonschema.Schema{execSchema, pathSchema, copySchema, fileSchema}

	// Add string schema type as an additional valid command type
	stringSchema := &jsonschema.Schema{
		Type:        "string",
		Description: "Strings will be parsed and interpreted as a command to run",
	}
	availableCommands = append([]*jsonschema.Schema{stringSchema}, availableCommands...)

	return &jsonschema.Schema{
		OneOf: availableCommands,
	}
}

func generateSchemaWithComments(v any) *jsonschema.Schema {
	r := jsonschema.Reflector{
		Anonymous:      true,
		DoNotReference: true,
	}
	return r.Reflect(v)
}
