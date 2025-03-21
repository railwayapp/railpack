package generate

import (
	"fmt"
	"sort"
	"strings"

	a "github.com/railwayapp/railpack/core/app"
	"github.com/railwayapp/railpack/core/config"
	"github.com/railwayapp/railpack/core/logger"
	"github.com/railwayapp/railpack/core/mise"
	"github.com/railwayapp/railpack/core/plan"
	"github.com/railwayapp/railpack/core/resolver"
	"github.com/railwayapp/railpack/internal/utils"
)

type BuildStepOptions struct {
	ResolvedPackages map[string]*resolver.ResolvedPackage
	Caches           *CacheContext
}

type StepBuilder interface {
	Name() string
	Build(options *BuildStepOptions) (*plan.Step, error)
}

type GenerateContext struct {
	App    *a.App
	Env    *a.Environment
	Config *config.Config

	BaseImage string
	Steps     []StepBuilder
	Deploy    *DeployBuilder

	Caches  *CacheContext
	Secrets []string

	SubContexts []string

	Metadata        *Metadata
	Resolver        *resolver.Resolver
	MiseStepBuilder *MiseStepBuilder

	Logger *logger.Logger
}

type Command interface {
	IsSpread() bool
}

type CommandWrapper struct {
	Command plan.Command
}

func (c CommandWrapper) IsSpread() bool {
	if execCmd, ok := c.Command.(plan.ExecCommand); ok {
		return execCmd.Cmd == plan.ShellCommandString("...") || execCmd.Cmd == "..."
	}
	return false
}

func NewGenerateContext(app *a.App, env *a.Environment, config *config.Config, logger *logger.Logger) (*GenerateContext, error) {
	resolver, err := resolver.NewResolver(mise.InstallDir)
	if err != nil {
		return nil, err
	}

	ctx := &GenerateContext{
		App:      app,
		Env:      env,
		Config:   config,
		Steps:    make([]StepBuilder, 0),
		Deploy:   NewDeployBuilder(),
		Caches:   NewCacheContext(),
		Secrets:  []string{},
		Metadata: NewMetadata(),
		Resolver: resolver,
		Logger:   logger,
	}

	// The default runtime image should include the runtime apt packages
	// ctx.Deploy.Inputs = append(ctx.Deploy.Inputs, ctx.DefaultRuntimeInput())

	return ctx, nil
}

func (c *GenerateContext) GetMiseStepBuilder() *MiseStepBuilder {
	if c.MiseStepBuilder == nil {
		c.MiseStepBuilder = c.newMiseStepBuilder()
	}
	return c.MiseStepBuilder
}

func (c *GenerateContext) EnterSubContext(subContext string) *GenerateContext {
	c.SubContexts = append(c.SubContexts, subContext)
	return c
}

func (c *GenerateContext) ExitSubContext() *GenerateContext {
	c.SubContexts = c.SubContexts[:len(c.SubContexts)-1]
	return c
}

func (c *GenerateContext) GetStepName(name string) string {
	subContextNames := strings.Join(c.SubContexts, ":")
	if subContextNames != "" {
		return name + ":" + subContextNames
	}
	return name
}

func (c *GenerateContext) GetStepByName(name string) *StepBuilder {
	for _, step := range c.Steps {
		if step.Name() == name {
			return &step
		}
	}
	return nil
}

func (c *GenerateContext) ResolvePackages() (map[string]*resolver.ResolvedPackage, error) {
	return c.Resolver.ResolvePackages()
}

// Generate a build plan from the context
func (c *GenerateContext) Generate() (*plan.BuildPlan, map[string]*resolver.ResolvedPackage, error) {
	c.applyConfig()

	// Resolve all package versions into a fully qualified and valid version
	resolvedPackages, err := c.ResolvePackages()
	if err != nil {
		return nil, nil, err
	}

	// Create the actual build plan
	buildPlan := plan.NewBuildPlan()

	buildStepOptions := &BuildStepOptions{
		ResolvedPackages: resolvedPackages,
		Caches:           c.Caches,
	}

	for _, stepBuilder := range c.Steps {
		step, err := stepBuilder.Build(buildStepOptions)

		if err != nil {
			return nil, nil, fmt.Errorf("failed to build step: %w", err)
		}

		buildPlan.AddStep(*step)
	}

	buildPlan.Caches = c.Caches.Caches
	buildPlan.Secrets = utils.RemoveDuplicates(c.Secrets)
	c.Deploy.Build(buildPlan, buildStepOptions)

	return buildPlan, resolvedPackages, nil
}

// func (c *GenerateContext) DefaultRuntimeInput() plan.Layer {
// 	return c.DefaultRuntimeInputWithPackages([]string{})
// }

// func (c *GenerateContext) DefaultRuntimeInputWithPackages(additionalAptPackages []string) plan.Layer {
// 	aptPackages := append(c.Config.Deploy.AptPackages, additionalAptPackages...)

// 	if len(aptPackages) == 0 {
// 		return plan.NewImageLayer(plan.RailpackRuntimeImage)
// 	}

// 	runtimeAptStep := c.NewAptStepBuilder("runtime")
// 	runtimeAptStep.Packages = aptPackages
// 	runtimeAptStep.AddInput(plan.NewImageLayer(plan.RailpackRuntimeImage))

// 	return plan.NewStepLayer(runtimeAptStep.Name())
// }

func (o *BuildStepOptions) NewAptInstallCommand(pkgs []string) plan.Command {
	pkgs = utils.RemoveDuplicates(pkgs)
	sort.Strings(pkgs)

	return plan.NewExecCommand("sh -c 'apt-get update && apt-get install -y "+strings.Join(pkgs, " ")+"'", plan.ExecOptions{
		CustomName: "install apt packages: " + strings.Join(pkgs, " "),
	})
}

func (c *GenerateContext) applyConfig() {
	// miseStep := c.GetMiseStepBuilder()
	// for _, pkg := range slices.Sorted(maps.Keys(c.Config.Packages)) {
	// 	version := c.Config.Packages[pkg]
	// 	pkgRef := miseStep.Default(pkg, version)
	// 	miseStep.Version(pkgRef, version, "custom config")
	// }

	// // Apply the cache config to the context
	// maps.Copy(c.Caches.Caches, c.Config.Caches)
	// c.Secrets = plan.SpreadStrings(c.Config.Secrets, c.Secrets)

	// // Apply step config to the context
	// for _, name := range slices.Sorted(maps.Keys(c.Config.Steps)) {
	// 	configStep := c.Config.Steps[name]

	// 	var commandStepBuilder *CommandStepBuilder

	// 	if existingStep := c.GetStepByName(name); existingStep != nil {
	// 		if csb, ok := (*existingStep).(*CommandStepBuilder); ok {
	// 			commandStepBuilder = csb
	// 		} else {
	// 			log.Warnf("Step `%s` exists, but it is not a command step. Skipping...", name)
	// 			continue
	// 		}
	// 	} else {
	// 		// If no build step found, create a new one
	// 		// Run the build in the builder context and copy the /app contents to the final image
	// 		commandStepBuilder = c.NewCommandStep(name)
	// 		commandStepBuilder.AddInput(plan.NewStepLayer(miseStep.Name()))
	// 		c.Deploy.Inputs = append(c.Deploy.Inputs, plan.NewStepLayer(commandStepBuilder.Name(), plan.InputOptions{
	// 			Include: []string{"."},
	// 		}))
	// 	}

	// 	commandStepBuilder.Commands = plan.Spread(configStep.Commands, commandStepBuilder.Commands)
	// 	commandStepBuilder.Inputs = plan.Spread(configStep.Inputs, commandStepBuilder.Inputs)

	// 	commandStepBuilder.Secrets = plan.SpreadStrings(configStep.Secrets, commandStepBuilder.Secrets)

	// 	commandStepBuilder.Caches = plan.SpreadStrings(configStep.Caches, commandStepBuilder.Caches)
	// 	commandStepBuilder.AddEnvVars(configStep.Variables)
	// 	maps.Copy(commandStepBuilder.Assets, configStep.Assets)
	// }

	// // Update deploy from config
	// if c.Config.Deploy != nil {
	// 	if c.Config.Deploy.StartCmd != "" {
	// 		c.Deploy.StartCmd = c.Config.Deploy.StartCmd
	// 	}

	// 	c.Deploy.Inputs = plan.Spread(c.Config.Deploy.Inputs, c.Deploy.Inputs)
	// 	c.Deploy.Paths = plan.SpreadStrings(c.Config.Deploy.Paths, c.Deploy.Paths)
	// 	maps.Copy(c.Deploy.Variables, c.Config.Deploy.Variables)
	// }

}
