package generate

import (
	"fmt"

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

func (b *DeployBuilder) Build(p *plan.BuildPlan, options *BuildStepOptions) error {
	baseLayer := b.Base

	bootstrapBuilder := NewMiseBootstrapStepBuilder(
		MiseBootstrapRuntimeStepName,
		baseLayer,
		b.AptPackages,
		options.MiseBootstrapProject.ConfigFiles,
	)
	bootstrapBuilder.CopyMise = true
	bootstrapBuilder.RunHooks = false
	bootstrapBuilder.ApplyProjectPackages = options.MiseBootstrapProject.HasAptPackages
	bootstrapBuilder.HasProjectHooks = options.MiseBootstrapProject.HasPackageHooks
	if bootstrapBuilder.IsRequired() && bootstrapBuilder.HasProjectHooks {
		// Repository hooks run in the builder where bootstrap utilities are available.
		bootstrapBuilder.Inputs = []plan.Layer{miseBootstrapRepositoryLayer()}
	}
	if bootstrapBuilder.IsRequired() {
		bootstrapStep, err := bootstrapBuilder.Build(options)
		if err != nil {
			return fmt.Errorf("failed to build runtime bootstrap step: %w", err)
		}

		p.Steps = append(p.Steps, *bootstrapStep)
		baseLayer = plan.NewStepLayer(bootstrapStep.Name)
	}

	p.Deploy.Base = baseLayer

	p.Deploy.Inputs = append(p.Deploy.Inputs, b.DeployInputs...)
	p.Deploy.StartCmd = b.StartCmd
	p.Deploy.Variables = b.Variables
	p.Deploy.Paths = b.Paths

	return nil
}
