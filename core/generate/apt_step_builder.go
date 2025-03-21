package generate

import (
	"fmt"

	"github.com/railwayapp/railpack/core/plan"
)

type AptStepBuilder struct {
	DisplayName string
	Packages    []string
	Inputs      []plan.Layer
}

func (c *GenerateContext) NewAptStepBuilder(name string) *AptStepBuilder {
	step := &AptStepBuilder{
		DisplayName: c.GetStepName(fmt.Sprintf("packages:%s", name)),
		Packages:    []string{},
		Inputs:      []plan.Layer{},
	}

	c.Steps = append(c.Steps, step)

	return step
}

func (b *AptStepBuilder) AddInput(input plan.Layer) {
	b.Inputs = append(b.Inputs, input)
}

func (b *AptStepBuilder) AddAptPackage(pkg string) {
	b.Packages = append(b.Packages, pkg)
}

func (b *AptStepBuilder) Name() string {
	return b.DisplayName
}

func (b *AptStepBuilder) Build(options *BuildStepOptions) (*plan.Step, error) {
	step := plan.NewStep(b.DisplayName)

	step.AddCommands([]plan.Command{
		options.NewAptInstallCommand(b.Packages),
	})

	step.Caches = options.Caches.GetAptCaches()
	step.Inputs = b.Inputs

	// Does not use any secrets
	step.Secrets = []string{}

	return step, nil
}
