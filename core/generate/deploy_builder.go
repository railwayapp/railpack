package generate

import (
	"github.com/railwayapp/railpack/core/plan"
)

type DeployBuilder struct {
	Base         plan.Layer
	DeployLaters []plan.Layer
	StartCmd     string
	Variables    map[string]string
	Paths        []string
	AptPackages  []string
}

func NewDeployBuilder() *DeployBuilder {
	return &DeployBuilder{
		Base:         plan.Layer{},
		DeployLaters: []plan.Layer{},
		StartCmd:     "",
		Variables:    map[string]string{},
		Paths:        []string{},
		AptPackages:  []string{},
	}
}

func (b *DeployBuilder) AddInputs(layers []plan.Layer) {
	b.DeployLaters = append(b.DeployLaters, layers...)
}

func (b *DeployBuilder) AddAptPackages(packages []string) {
	b.AptPackages = append(b.AptPackages, packages...)
}

func (b *DeployBuilder) Build(p *plan.BuildPlan, options *BuildStepOptions) {
	baseLayer := plan.NewImageLayer(plan.RailpackRuntimeImage)

	if len(b.AptPackages) > 0 {
		runtimeAptStep := plan.NewStep("packages:apt:runtime")
		runtimeAptStep.Inputs = []plan.Layer{baseLayer}
		runtimeAptStep.AddCommands([]plan.Command{
			options.NewAptInstallCommand(b.AptPackages),
		})
		runtimeAptStep.Caches = options.Caches.GetAptCaches()
		p.Steps = append(p.Steps, *runtimeAptStep)
		baseLayer = plan.NewStepLayer(runtimeAptStep.Name)
	}

	p.Deploy.Base = &baseLayer

	p.Deploy.Inputs = append(p.Deploy.Inputs, b.DeployLaters...)
	p.Deploy.StartCmd = b.StartCmd
	p.Deploy.Variables = b.Variables
	p.Deploy.Paths = b.Paths
}
