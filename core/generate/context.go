package generate

import (
	"fmt"
	"maps"
	"slices"
	"sort"
	"strings"

	"github.com/charmbracelet/log"
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
	Build(p *plan.BuildPlan, options *BuildStepOptions) error
}

type GenerateContext struct {
	App             *a.App
	Env             *a.Environment
	Config          *config.Config
	dockerignoreCtx *plan.DockerignoreContext

	BaseImage string
	Steps     []StepBuilder
	Deploy    *DeployBuilder

	Caches *CacheContext

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

// is a user-provided command entry a "spread" command?
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

	dockerignoreCtx, err := plan.NewDockerignoreContext(app)
	if err != nil {
		return nil, fmt.Errorf("failed to parse .dockerignore: %w", err)
	}

	if dockerignoreCtx.HasFile {
		logger.LogInfo("Found .dockerignore file, applying filters")
		log.Debugf("Dockerignore patterns: %v", dockerignoreCtx.Excludes)
	}

	ctx := &GenerateContext{
		App:             app,
		Env:             env,
		Config:          config,
		Steps:           make([]StepBuilder, 0),
		Deploy:          NewDeployBuilder(),
		Caches:          NewCacheContext(),
		Metadata:        NewMetadata(),
		Resolver:        resolver,
		Logger:          logger,
		dockerignoreCtx: dockerignoreCtx,
	}

	ctx.applyPackagesFromConfig()

	if dockerignoreCtx.HasFile {
		ctx.Metadata.SetBool("dockerIgnore", true)
	}

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

	buildPlan := plan.NewBuildPlan()

	// Merge exclude patterns from .dockerignore and railpack.json
	excludePatterns := []string{}
	excludePatterns = append(excludePatterns, c.dockerignoreCtx.Excludes...)
	excludePatterns = append(excludePatterns, c.Config.Exclude...)
	if len(excludePatterns) > 0 {
		buildPlan.Exclude = excludePatterns
	}

	buildStepOptions := &BuildStepOptions{
		ResolvedPackages: resolvedPackages,
		Caches:           c.Caches,
	}

	for _, stepBuilder := range c.Steps {
		err := stepBuilder.Build(buildPlan, buildStepOptions)

		if err != nil {
			return nil, nil, fmt.Errorf("failed to build step: %w", err)
		}
	}

	buildPlan.Caches = c.Caches.Caches
	c.Deploy.Build(buildPlan, buildStepOptions)

	buildPlan.Normalize()
	buildPlan.Secrets, buildPlan.Steps = resolveSecrets(c.Config, buildPlan.Steps)

	return buildPlan, resolvedPackages, nil
}

func (o *BuildStepOptions) NewAptInstallCommand(pkgs []string) plan.Command {
	pkgs = utils.RemoveDuplicates(pkgs)
	sort.Strings(pkgs)

	// sh -c is required because && is a shell operator that needs a shell to interpret it
	return plan.NewExecCommand("sh -c 'apt-get update && apt-get install -y "+strings.Join(pkgs, " ")+"'", plan.ExecOptions{
		CustomName: "install apt packages: " + strings.Join(pkgs, " "),
	})
}

func (c *GenerateContext) applyPackagesFromConfig() {
	miseStep := c.GetMiseStepBuilder()

	// railpack.json supports defining custom packages, if we find them we seed the mise builder versions with those user-specified values
	// other more specific version definitions (such as package.json, ENV vars, etc) will take precedence over these
	for _, pkg := range slices.Sorted(maps.Keys(c.Config.Packages)) {
		version := c.Config.Packages[pkg]
		pkgRef := miseStep.Default(pkg, version)
		// `custom config` and not `railpack.json` is used since the source of the custom config could be a CLI flag or custom config file
		miseStep.Version(pkgRef, version, "custom config")
	}
}

func (c *GenerateContext) applyConfig() {
	c.applyPackagesFromConfig()
	c.applyBuildAptPackages()

	// Apply the cache config to the context
	maps.Copy(c.Caches.Caches, c.Config.Caches)

	// Update deploy from config
	if c.Config.Deploy != nil {
		if c.Config.Deploy.Base != nil && !c.Config.Deploy.Base.IsEmpty() {
			c.Deploy.Base = *c.Config.Deploy.Base
		}

		if c.Config.Deploy.StartCmd != "" {
			c.Deploy.StartCmd = c.Config.Deploy.StartCmd
		}

		c.applyDeployAptPackages()
		c.Deploy.DeployInputs = plan.Spread(c.Config.Deploy.Inputs, c.Deploy.DeployInputs)
		c.Deploy.Paths = plan.SpreadStrings(c.Config.Deploy.Paths, c.Deploy.Paths)
		maps.Copy(c.Deploy.Variables, c.Config.Deploy.Variables)
	}

	// A spread retains generated deploy composition; any explicit list without one takes full control.
	replacesGeneratedDeployInputs := c.Config.Deploy != nil &&
		c.Config.Deploy.Inputs != nil &&
		!slices.ContainsFunc(c.Config.Deploy.Inputs, plan.Layer.IsSpread)

	// Apply step config to the context
	for _, name := range slices.Sorted(maps.Keys(c.Config.Steps)) {
		configStep := c.Config.Steps[name]

		var commandStepBuilder *CommandStepBuilder

		if existingStep := c.GetStepByName(name); existingStep != nil {
			if csb, ok := (*existingStep).(*CommandStepBuilder); ok {
				commandStepBuilder = csb
			} else {
				log.Warnf("Step `%s` exists, but it is not a command step. Skipping...", name)
				continue
			}
		} else {
			// If no build step found, create a new one
			// Run the build in the builder context and copy the /app contents to the final image
			commandStepBuilder = c.NewCommandStep(name)
			commandStepBuilder.AddInput(plan.NewStepLayer(c.GetMiseStepBuilder().Name()))
		}

		commandStepBuilder.Inputs = plan.Spread(configStep.Inputs, commandStepBuilder.Inputs)
		commandStepBuilder.Commands = plan.Spread(configStep.Commands, commandStepBuilder.Commands)
		commandStepBuilder.Secrets = plan.SpreadStrings(configStep.Secrets, commandStepBuilder.Secrets)
		commandStepBuilder.Caches = plan.SpreadStrings(configStep.Caches, commandStepBuilder.Caches)
		commandStepBuilder.AddEnvVars(configStep.Variables)
		maps.Copy(commandStepBuilder.Assets, configStep.Assets)

		// Convert the deploy outputs into layers that will be added to the deploy.
		// Skip if the path is already covered by existing inputs from this step
		// (e.g. provider already added "." so we don't duplicate it from --build-cmd).
		outputFilters := []plan.Filter{plan.NewIncludeFilter([]string{"."})}
		if configStep.DeployOutputs != nil {
			// if deploy outputs are explicitly set on a step, then always use them, regardless of deploy configuration
			// TODO I don't like this and find it confusing: deploy.inputs should be able to override step-level deploy outputs
			outputFilters = configStep.DeployOutputs
		} else if replacesGeneratedDeployInputs || c.Deploy.HasInputForStep(name) {
			// if no deployOutput is specified on a step, the user has not specified a "..." in deploy.inputs, and
			continue
		}
		for _, filter := range outputFilters {
			if slices.ContainsFunc(filter.Include, func(inc string) bool {
				return c.Deploy.HasIncludeForStep(name, inc)
			}) {
				continue
			}
			c.Deploy.AddInputs([]plan.Layer{plan.NewStepLayer(name, filter)})
		}
	}

	c.notifyCustomAptDebianUpgrade()
}

// TODO(2026-10-17): remove this Debian upgrade notice for custom apt packages.
func (c *GenerateContext) notifyCustomAptDebianUpgrade() {
	hasCustom := false
	for _, pkg := range c.Config.BuildAptPackages {
		if pkg != "" && pkg != "..." {
			hasCustom = true
			break
		}
	}
	if !hasCustom && c.Config.Deploy != nil {
		for _, pkg := range c.Config.Deploy.AptPackages {
			if pkg != "" && pkg != "..." {
				hasCustom = true
				break
			}
		}
	}
	if !hasCustom {
		return
	}

	c.Logger.LogInfo("The debian base image has been upgraded and you may experience issues with custom apt packages. Report any issues here: https://github.com/railwayapp/railpack/issues")
}

func (c *GenerateContext) applyBuildAptPackages() {
	configuredPackages := c.Config.BuildAptPackages
	if configuredPackages == nil {
		return
	}

	if !slices.Contains(configuredPackages, "...") {
		// TODO the names of these configs will probably change in a future release as well...
		c.Logger.LogDeprecation("`buildAptPackages` without a `...` entry will replace Railpack packages in the future")
		c.Logger.LogSuggestion("Add `...` to `buildAptPackages` to retain Railpack packages", "/guides/installing-packages")

		// TODO: Remove this implicit spread so lists without "..." replace generated packages.
		configuredPackages = append([]string{"..."}, configuredPackages...)
	}

	miseStep := c.GetMiseStepBuilder()
	miseStep.SupportingAptPackages = plan.SpreadStrings(configuredPackages, miseStep.SupportingAptPackages)
}

func (c *GenerateContext) applyDeployAptPackages() {
	configuredPackages := c.Config.Deploy.AptPackages
	if configuredPackages != nil && !slices.Contains(configuredPackages, "...") {
		c.Logger.LogSuggestion("Add `...` to `deploy.aptPackages` to retain Railpack packages", "/guides/installing-packages")
	}

	c.Deploy.AptPackages = plan.SpreadStrings(configuredPackages, c.Deploy.AptPackages)
}

// in order to get around a circular dependency issue, we need to define discrete getters to interface with
// the mise package version logic.

func (c *GenerateContext) GetAppSource() string {
	return c.App.Source
}

func (c *GenerateContext) GetLogger() *logger.Logger {
	return c.Logger
}
