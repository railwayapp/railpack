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

	bootstrapBuilder := NewMiseBootstrapStepBuilder(
		b.DisplayName,
		plan.NewImageLayer(image),
		b.AptPackages,
		options.MiseBootstrapProject.ConfigFiles,
	)
	bootstrapBuilder.CopyMise = true
	bootstrapBuilder.RunHooks = false
	bootstrapBuilder.ApplyProjectPackages = options.MiseBootstrapProject.HasAptPackages
	bootstrapBuilder.HasProjectHooks = options.MiseBootstrapProject.HasPackageHooks
	if bootstrapBuilder.IsRequired() && bootstrapBuilder.HasProjectHooks {
		// Run repository hooks once on the full builder image, then reuse their output.
		bootstrapBuilder.Inputs = []plan.Layer{miseBootstrapRepositoryLayer()}
	}
	if bootstrapBuilder.IsRequired() {
		step, err := bootstrapBuilder.Build(options)
		if err != nil {
			return err
		}

		p.Steps = append(p.Steps, *step)
		return nil
	}

	step := plan.NewStep(b.DisplayName)
	step.Secrets = []string{}
	step.Inputs = []plan.Layer{plan.NewImageLayer(image)}
	p.Steps = append(p.Steps, *step)

	return nil
}
