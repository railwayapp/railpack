package generate

import (
	"github.com/railwayapp/railpack/core/plan"
	"github.com/railwayapp/railpack/core/resolver"
)

type ImageStepBuilder struct {
	DisplayName      string
	Resolver         *resolver.Resolver
	Packages         []*resolver.PackageRef
	ResolveStepImage func(options *BuildStepOptions) string
	AptPackages      []string
}

func (c *GenerateContext) NewImageStep(name string, resolveStepImage func(options *BuildStepOptions) string) *ImageStepBuilder {
	step := &ImageStepBuilder{
		DisplayName:      c.GetStepName(name),
		Resolver:         c.Resolver,
		ResolveStepImage: resolveStepImage,
	}

	c.Steps = append(c.Steps, step)

	return step
}

func (b *ImageStepBuilder) Default(name string, defaultVersion string) resolver.PackageRef {
	for _, pkg := range b.Packages {
		if pkg.Name == name {
			return *pkg
		}
	}

	pkg := b.Resolver.Default(name, defaultVersion)
	b.Packages = append(b.Packages, &pkg)
	return pkg
}

func (b *ImageStepBuilder) Version(name resolver.PackageRef, version string, source string) {
	b.Resolver.Version(name, version, source)
}

func (b *ImageStepBuilder) SetVersionAvailable(ref resolver.PackageRef, isVersionAvailable func(version string) bool) {
	b.Resolver.SetVersionAvailable(ref, isVersionAvailable)
}

func (b *ImageStepBuilder) Name() string {
	return b.DisplayName
}

func (b *ImageStepBuilder) Build(p *plan.BuildPlan, options *BuildStepOptions) error {
	image := b.ResolveStepImage(options)

	step := plan.NewStep(b.DisplayName)

	step.Inputs = []plan.Layer{
		plan.NewImageLayer(image),
	}

	step.Secrets = []string{}

	if len(b.AptPackages) > 0 {
		step.Commands = []plan.Command{
			options.NewAptInstallCommand(b.AptPackages),
		}
	}

	p.Steps = append(p.Steps, *step)

	return nil
}
