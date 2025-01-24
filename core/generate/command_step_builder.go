package generate

import (
	"github.com/railwayapp/railpack-go/core/plan"
	"github.com/railwayapp/railpack-go/core/utils"
)

type CommandStepBuilder struct {
	DisplayName string
	DependsOn   []string
	Commands    []plan.Command
	Outputs     []string
	Assets      map[string]string
}

func (c *GenerateContext) NewCommandStep(name string) *CommandStepBuilder {
	step := &CommandStepBuilder{
		DisplayName: c.GetStepName(name),
		DependsOn:   []string{MisePackageStepName},
		Commands:    []plan.Command{},
		Outputs:     []string{},
		Assets:      map[string]string{},
	}

	c.Steps = append(c.Steps, step)

	return step
}

func (b *CommandStepBuilder) DependOn(name string) {
	b.DependsOn = append(b.DependsOn, name)
}

func (b *CommandStepBuilder) AddCommand(command plan.Command) {
	b.Commands = append(b.Commands, command)
}

func (b *CommandStepBuilder) AddCommands(commands []plan.Command) {
	b.Commands = append(b.Commands, commands...)
}

func (b *CommandStepBuilder) Build(options *BuildStepOptions) (*plan.Step, error) {
	step := plan.NewStep(b.DisplayName)

	step.DependsOn = utils.RemoveDuplicates(b.DependsOn)
	step.Outputs = utils.RemoveDuplicates(b.Outputs)
	step.Commands = b.Commands
	step.Assets = b.Assets

	return step, nil
}
