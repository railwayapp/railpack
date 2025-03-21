package generate

import "github.com/railwayapp/railpack/core/plan"

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

func (b *DeployBuilder) Build(p *plan.BuildPlan) {
	baseLayer := plan.NewImageLayer(plan.RailpackRuntimeImage)
	p.Deploy.Base = &baseLayer

	p.Deploy.Inputs = append(p.Deploy.Inputs, b.DeployLaters...)
	p.Deploy.StartCmd = b.StartCmd
	p.Deploy.Variables = b.Variables
	p.Deploy.Paths = b.Paths
}
