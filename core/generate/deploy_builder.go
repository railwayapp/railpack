package generate

import (
	"github.com/railwayapp/railpack/core/plan"
)

type DeployBuilder struct {
	Base         plan.Layer
	DeployInputs []plan.Layer
	StartCmd     string
	Variables    map[string]string
	Paths        []string
	AptPackages  []string
}

func NewDeployBuilder() *DeployBuilder {
	return &DeployBuilder{
		Base:         plan.NewImageLayer(plan.RailpackRuntimeImage),
		DeployInputs: []plan.Layer{},
		StartCmd:     "",
		Variables:    map[string]string{},
		Paths:        []string{},
		AptPackages:  []string{},
	}
}

func (b *DeployBuilder) SetInputs(layers []plan.Layer) {
	b.DeployInputs = layers
}

func (b *DeployBuilder) AddInputs(layers []plan.Layer) {
	b.DeployInputs = append(b.DeployInputs, layers...)
}

func (b *DeployBuilder) HasIncludeForStep(stepName string, path string) bool {
	for _, layer := range b.DeployInputs {
		if layer.Step != stepName {
			continue
		}
		for _, inc := range layer.Include {
			// exact match, or existing include is "." which covers everything
			if inc == path || inc == "." {
				return true
			}
		}
	}
	return false
}

func (b *DeployBuilder) AddAptPackages(packages []string) {
	b.AptPackages = append(b.AptPackages, packages...)
}

func (b *DeployBuilder) Build(p *plan.BuildPlan, options *BuildStepOptions) {
	baseLayer := b.Base

	if len(b.AptPackages) > 0 {
		runtimeAptStep := plan.NewStep("packages:apt:runtime")
		runtimeAptStep.Inputs = []plan.Layer{baseLayer}
		runtimeAptStep.AddCommands([]plan.Command{
			options.NewAptInstallCommand(b.AptPackages),
		})
		runtimeAptStep.Caches = options.Caches.GetAptCaches()
		runtimeAptStep.Secrets = []string{}
		p.Steps = append(p.Steps, *runtimeAptStep)
		baseLayer = plan.NewStepLayer(runtimeAptStep.Name)
	}

	p.Deploy.Base = baseLayer

	p.Deploy.Inputs = append(p.Deploy.Inputs, b.DeployInputs...)
	p.Deploy.StartCmd = b.StartCmd
	p.Deploy.Variables = b.Variables
	p.Deploy.Paths = b.Paths
}
