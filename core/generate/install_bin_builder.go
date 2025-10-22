package generate

import (
	"fmt"
	"maps"

	"github.com/railwayapp/railpack/core/plan"
	"github.com/railwayapp/railpack/core/resolver"
)

const (
	BinDir = "/railpack"
)

type InstallBinStepBuilder struct {
	DisplayName           string
	Resolver              *resolver.Resolver
	SupportingAptPackages []string
	Package               resolver.PackageRef
}

func (c *GenerateContext) NewInstallBinStepBuilder(name string) *InstallBinStepBuilder {
	step := &InstallBinStepBuilder{
		DisplayName: c.GetStepName(name),
		Resolver:    c.Resolver,
		Package:     resolver.PackageRef{},
	}

	c.Steps = append(c.Steps, step)

	return step
}

func (b *InstallBinStepBuilder) Name() string {
	return b.DisplayName
}

func (b *InstallBinStepBuilder) Default(name string, defaultVersion string) resolver.PackageRef {
	b.Package = b.Resolver.Default(name, defaultVersion)
	return b.Package
}

func (b *InstallBinStepBuilder) GetOutputPaths() []string {
	return []string{b.getBinPath()}
}

func (b *InstallBinStepBuilder) Version(name resolver.PackageRef, version string, source string) {
	b.Resolver.Version(name, version, source)
}

func (b *InstallBinStepBuilder) GetLayer() plan.Layer {
	return plan.NewStepLayer(b.Name(), plan.Filter{
		Include: b.GetOutputPaths(),
	})
}

func (b *InstallBinStepBuilder) Build(p *plan.BuildPlan, options *BuildStepOptions) error {
	packageVersion := options.ResolvedPackages[b.Package.Name].ResolvedVersion
	if packageVersion == nil {
		return fmt.Errorf("package %s not found", b.Package.Name)
	}

	step := plan.NewStep(b.DisplayName)
	step.Secrets = []string{}
	step.Inputs = []plan.Layer{
		plan.NewImageLayer(plan.RailpackBuilderImage),
	}

	binPath := b.getBinPath()

	step.AddCommands([]plan.Command{
		plan.NewExecCommand(fmt.Sprintf("mise install-into %s@%s %s", b.Package.Name, *packageVersion, binPath)),
		plan.NewPathCommand(binPath),
		plan.NewPathCommand(fmt.Sprintf("%s/bin", binPath)),
	})

	// Add user variables to this step
	maps.Copy(step.Variables, options.UserVariables)

	p.Steps = append(p.Steps, *step)

	return nil
}

func (b *InstallBinStepBuilder) getBinPath() string {
	return fmt.Sprintf("%s/%s", BinDir, b.Package.Name)
}
