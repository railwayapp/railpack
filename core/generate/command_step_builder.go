package generate

import (
	"maps"

	"github.com/charmbracelet/log"
	a "github.com/railwayapp/railpack/core/app"
	"github.com/railwayapp/railpack/core/plan"
)

type CommandStepBuilder struct {
	DisplayName string
	Commands    []plan.Command
	Inputs      []plan.Layer
	Assets      map[string]string
	Variables   map[string]string
	Caches      []string
	Secrets     []string
	app         *a.App
	env         *a.Environment
}

func (c *GenerateContext) NewCommandStep(name string) *CommandStepBuilder {
	step := &CommandStepBuilder{
		DisplayName: c.GetStepName(name),
		Inputs:      []plan.Layer{},
		Commands:    []plan.Command{},
		Assets:      map[string]string{},
		Variables:   map[string]string{},
		Caches:      []string{},
		Secrets:     []string{"*"},
		app:         c.App,
		env:         c.Env,
	}

	// Remove any existing step with the same name
	for i, existingStep := range c.Steps {
		if existingStep.Name() == step.Name() {
			c.Steps = append(c.Steps[:i], c.Steps[i+1:]...)
			break
		}
	}

	c.Steps = append(c.Steps, step)

	return step
}

func (b *CommandStepBuilder) AddInput(input plan.Layer) {
	b.Inputs = append(b.Inputs, input)
}

func (b *CommandStepBuilder) AddInputs(inputs []plan.Layer) {
	b.Inputs = append(b.Inputs, inputs...)
}

func (b *CommandStepBuilder) AddVariables(variables map[string]string) {
	maps.Copy(b.Variables, variables)
}

func (b *CommandStepBuilder) AddCache(name string) {
	if name == "" {
		log.Error("Cache name is empty")
		return
	}

	b.Caches = append(b.Caches, name)
}

func (b *CommandStepBuilder) AddCommand(command plan.Command) {
	b.AddCommands([]plan.Command{command})
}

func (b *CommandStepBuilder) AddCommands(commands []plan.Command) {
	if b.Commands == nil {
		b.Commands = []plan.Command{}
	}
	b.Commands = append(b.Commands, commands...)
}

func (b *CommandStepBuilder) AddEnvVars(envVars map[string]string) {
	maps.Copy(b.Variables, envVars)
}

func (b *CommandStepBuilder) AddPaths(paths []string) {
	commands := []plan.Command{}
	for _, path := range paths {
		commands = append(commands, plan.NewPathCommand(path))
	}
	b.AddCommands(commands)
}

func (b *CommandStepBuilder) UseSecretsWithPrefixes(prefixes []string) {
	for _, prefix := range prefixes {
		b.UseSecretsWithPrefix(prefix)
	}
}

func (b *CommandStepBuilder) UseSecretsWithPrefix(prefix string) {
	secrets := b.env.GetSecretsWithPrefix(prefix)
	b.Secrets = append(b.Secrets, secrets...)
}

func (b *CommandStepBuilder) UseSecrets(secrets []string) {
	if b.env.GetVariable("CI") != "" {
		b.Secrets = append(b.Secrets, secrets...)
	}
}

func (b *CommandStepBuilder) Name() string {
	return b.DisplayName
}

func (b *CommandStepBuilder) Build(p *plan.BuildPlan, options *BuildStepOptions) error {
	step := plan.NewStep(b.DisplayName)

	step.Inputs = b.Inputs
	step.Commands = b.Commands
	step.Assets = b.Assets
	step.Caches = b.Caches
	step.Variables = b.Variables
	step.Secrets = b.Secrets

	p.Steps = append(p.Steps, *step)

	return nil
}
