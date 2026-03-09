// build step that installs packages defined by mise configuration, provider configuration, or user configuration
package generate

import (
	"encoding/json"
	"fmt"
	"maps"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	a "github.com/railwayapp/railpack/core/app"
	"github.com/railwayapp/railpack/core/mise"
	"github.com/railwayapp/railpack/core/plan"
	"github.com/railwayapp/railpack/core/resolver"
)

const (
	MisePackageStepName = "packages:mise"
	// System-level config at /etc/mise/config.toml is auto-trusted by mise
	MiseInstallCommand = "mise install"
)

var (
	RailpackBuilderImage = fmt.Sprintf("ghcr.io/railwayapp/railpack-builder:mise-%s", mise.Version)
)

// represents a app-local mise package
type MisePackageInfo struct {
	Version string
	Source  string
}

// MiseListSource represents the source of a mise tool installation
type MiseListSource struct {
	Type string `json:"type"`
	Path string `json:"path"`
}

// MiseListTool represents a tool in the mise list output
type MiseListTool struct {
	Version          string         `json:"version"`
	RequestedVersion string         `json:"requested_version"`
	InstallPath      string         `json:"install_path"`
	Source           MiseListSource `json:"source"`
	Installed        bool           `json:"installed"`
	// --current ensures Active=true for all entries
	Active bool `json:"active"`
}

// MisePackageListOutput represents the full output of `mise list --current --json`
type MisePackageListOutput map[string][]MiseListTool

type MiseStepBuilder struct {
	DisplayName           string
	Resolver              *resolver.Resolver
	SupportingAptPackages []string
	MisePackages          []*resolver.PackageRef
	SupportingMiseFiles   []string
	Assets                map[string]string
	Inputs                []plan.Layer
	Variables             map[string]string
	app                   *a.App
	env                   *a.Environment
}

func (c *GenerateContext) NewMiseStepBuilder(displayName string) *MiseStepBuilder {
	supportingAptPackages := c.Config.BuildAptPackages

	step := &MiseStepBuilder{
		DisplayName:           displayName,
		Resolver:              c.Resolver,
		MisePackages:          []*resolver.PackageRef{},
		SupportingAptPackages: append(supportingAptPackages, c.Config.BuildAptPackages...),
		Assets:                map[string]string{},
		Inputs:                []plan.Layer{},
		Variables:             map[string]string{},
		app:                   c.App,
		env:                   c.Env,
	}

	c.Steps = append(c.Steps, step)

	return step
}

func (c *GenerateContext) newMiseStepBuilder() *MiseStepBuilder {
	step := c.NewMiseStepBuilder(MisePackageStepName)

	return step
}

func (b *MiseStepBuilder) AddSupportingAptPackage(name string) {
	b.SupportingAptPackages = append(b.SupportingAptPackages, name)
}

func (b *MiseStepBuilder) AddInput(input plan.Layer) {
	b.Inputs = append(b.Inputs, input)
}

func (b *MiseStepBuilder) Default(name string, defaultVersion string) resolver.PackageRef {
	for _, pkg := range b.MisePackages {
		if pkg.Name == name {
			return *pkg
		}
	}

	pkg := b.Resolver.Default(name, defaultVersion)
	b.MisePackages = append(b.MisePackages, &pkg)
	return pkg
}

func (b *MiseStepBuilder) Version(name resolver.PackageRef, version string, source string) {
	b.Resolver.Version(name, version, source)
}

func (b *MiseStepBuilder) SkipMiseInstall(name resolver.PackageRef) {
	b.Resolver.SetSkipMiseInstall(name, true)
}

// GetMisePackageVersions gets all package versions from mise that are defined in the app directory environment
// this can include additional packages defined outside the app directory, but we filter those out
func (b *MiseStepBuilder) GetMisePackageVersions(ctx *GenerateContext) (map[string]*MisePackageInfo, error) {
	miseInstance, err := mise.New(mise.InstallDir)
	if err != nil {
		return nil, err
	}

	appDir := ctx.GetAppSource()
	output, err := miseInstance.GetCurrentList(appDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get package versions: %w", err)
	}

	var listOutput MisePackageListOutput
	if err := json.Unmarshal([]byte(output), &listOutput); err != nil {
		return nil, fmt.Errorf("failed to parse mise list output: %w", err)
	}

	packages := make(map[string]*MisePackageInfo)

	for toolName, tools := range listOutput {
		var appDirTools []MiseListTool
		for _, tool := range tools {
			// Only include tools that are sourced from within the app directory
			if strings.HasPrefix(tool.Source.Path, appDir) {
				appDirTools = append(appDirTools, tool)
			}
		}

		if len(appDirTools) > 1 {
			versions := make([]string, len(appDirTools))
			for i, tool := range appDirTools {
				versions[i] = tool.Version
			}

			// this is possible, although in practice it should be extremely rare
			ctx.GetLogger().LogWarn("Multiple versions of tool '%s' found: %v. Using the first one: %s",
				toolName, versions, versions[0])
		}

		if len(appDirTools) > 0 {
			firstTool := appDirTools[0]
			packages[toolName] = &MisePackageInfo{
				Version: firstTool.Version,
				// include the source so we can surface this to the user so they understand where the package version came from
				Source: firstTool.Source.Type,
			}
		}
	}

	return packages, nil
}

// Use mise-specified versions for all packages in the input list
func (b *MiseStepBuilder) UseMiseVersions(ctx *GenerateContext, packages []string) {
	miseVersions, err := b.GetMisePackageVersions(ctx)
	if err != nil {
		ctx.Logger.LogWarn("Failed to get package versions from mise: %s", err.Error())
		return
	}

	if miseVersions == nil {
		return
	}

	for _, packageName := range packages {
		if pkg := miseVersions[packageName]; pkg != nil {
			// Find the existing package reference
			for _, pkgRef := range b.MisePackages {
				if pkgRef.Name == packageName {
					b.Version(*pkgRef, pkg.Version, pkg.Source)
					break
				}
			}
		}
	}
}

func (b *MiseStepBuilder) Name() string {
	return b.DisplayName
}

func (b *MiseStepBuilder) GetOutputPaths() []string {
	if len(b.MisePackages) == 0 && len(b.getSupportingMiseConfigFiles()) == 0 {
		return []string{}
	}

	return []string{"/mise/shims", "/mise/installs", "/usr/local/bin/mise", "/etc/mise/config.toml", "/root/.local/state/mise"}
}

func (b *MiseStepBuilder) GetLayer() plan.Layer {
	outputPaths := b.GetOutputPaths()
	if len(outputPaths) == 0 {
		return plan.Layer{}
	}

	return plan.NewStepLayer(b.Name(), plan.Filter{
		Include: outputPaths,
	})
}

func (b *MiseStepBuilder) Build(p *plan.BuildPlan, options *BuildStepOptions) error {
	baseLayer := plan.NewImageLayer(RailpackBuilderImage)

	if len(b.SupportingAptPackages) > 0 {
		aptStep := plan.NewStep("packages:apt:build")
		aptStep.Inputs = []plan.Layer{baseLayer}
		aptStep.AddCommands([]plan.Command{
			options.NewAptInstallCommand(b.SupportingAptPackages),
		})
		aptStep.Caches = options.Caches.GetAptCaches()
		aptStep.Secrets = []string{}

		p.Steps = append(p.Steps, *aptStep)
		baseLayer = plan.NewStepLayer(aptStep.Name)
	}

	step := plan.NewStep(b.DisplayName)

	step.Inputs = []plan.Layer{baseLayer}

	supportingMiseConfigFiles := b.getSupportingMiseConfigFiles()
	if len(b.MisePackages) > 0 || len(supportingMiseConfigFiles) > 0 {
		step.AddCommands([]plan.Command{plan.NewPathCommand("/mise/shims")})
		// NOTE make sure to keep (some) of the variables below in sync with install_bin_builder
		maps.Copy(step.Variables, map[string]string{
			"MISE_DATA_DIR":     "/mise",
			"MISE_CONFIG_DIR":   "/mise",
			"MISE_CACHE_DIR":    "/mise/cache",
			"MISE_SHIMS_DIR":    "/mise/shims",
			"MISE_INSTALLS_DIR": "/mise/installs",
			// Don't verify the asset because recently released versions don't have a public key to verify against
			// https://github.com/railwayapp/railpack/issues/207
			"MISE_NODE_VERIFY": "false",
			// Enforces HTTPS and stricter security
			"MISE_PARANOID": "1",
			// Trust config files in the app directory to avoid trust warnings during build
			"MISE_TRUSTED_CONFIG_PATHS": "/app",
			// Enable mise to automatically read idiomatic version files
			"MISE_IDIOMATIC_VERSION_FILE_ENABLE_TOOLS": mise.IdiomaticVersionFileTools,
		})
		maps.Copy(step.Variables, b.Variables)

		// pass through the MISE_VERBOSE variable for detailed logging
		if verbose := b.env.GetVariable("MISE_VERBOSE"); verbose != "" {
			step.Variables["MISE_VERBOSE"] = verbose
		}

		// Add user mise config files if they exist
		for _, file := range supportingMiseConfigFiles {
			step.AddCommands([]plan.Command{
				plan.NewCopyCommand(file),
			})
		}

		// Setup mise commands
		packagesToInstall := make(map[string]string)
		for _, pkg := range b.MisePackages {
			resolved, ok := options.ResolvedPackages[pkg.Name]

			if ok && resolved.ResolvedVersion != nil && !b.Resolver.Get(pkg.Name).SkipMiseInstall {
				packagesToInstall[pkg.Name] = *resolved.ResolvedVersion
			}
		}

		miseToml, err := mise.GenerateMiseToml(packagesToInstall)
		if err != nil {
			return fmt.Errorf("failed to generate mise.toml: %w", err)
		}

		// use a `generated-` to make it clear to the plan reader that this is system-generated, not user provided
		b.Assets["generated-mise-toml"] = miseToml

		pkgNames := make([]string, 0, len(packagesToInstall))
		for k := range packagesToInstall {
			pkgNames = append(pkgNames, k)
		}
		sort.Strings(pkgNames)

		step.AddCommands([]plan.Command{
			plan.NewFileCommand("/etc/mise/config.toml", "generated-mise-toml", plan.FileOptions{
				CustomName: "create mise config",
			}),
			plan.NewExecCommand(MiseInstallCommand, plan.ExecOptions{
				CustomName: "install mise packages: " + strings.Join(pkgNames, ", "),
			}),
		})
	}

	step.Assets = b.Assets
	step.Secrets = []string{}

	p.Steps = append(p.Steps, *step)

	return nil
}

// https://mise.jdx.dev/configuration.html#idiomatic-version-files
var miseIdiomaticFiles = []string{
	".python-version",
	".python-versions",
	".node-version",
	".nvmrc",
	".ruby-version",
	"Gemfile",
	".go-version",
	".java-version",
	".sdkmanrc",
	".exenv-version",
	".deno-version",
	// .bun-version is a community convention, not officially supported by Bun
	".bun-version",
	".yvmrc",
}

// https://mise.jdx.dev/configuration.html#configuration-hierarchy
var miseConfigFiles = []string{
	"mise.toml",
	".mise.toml",
	"mise/config.toml",
	".mise/config.toml",
	".config/mise.toml",
	".config/mise/config.toml",
	".tool-versions",
}

// https://mise.jdx.dev/configuration.html#mise-toml
// the env-specific mise files are not as well documented, but we should look for them and include them in the install
// step in case a user specified a MISE_ENV. They won't negatively impact builds otherwise (outside of more frequency cache busting)
var miseConfigGlobs = []string{
	"mise.*.toml",
	".mise.*.toml",
	".config/mise/conf.d/*.toml",
}

// This logic casts a wide net to find any mise configuration that may exist in the app source.
// this enables the user to use mise config to configure the runtime, options, etc in a pretty granular way but also
// requires the user to understand how to enable mise for various environments if they have a more advanced configuration
func (b *MiseStepBuilder) getSupportingMiseConfigFiles() []string {
	seen := map[string]bool{}
	files := []string{}

	add := func(file string) {
		if !seen[file] {
			seen[file] = true
			files = append(files, file)
		}
	}

	for _, file := range slices.Concat(miseConfigFiles, miseIdiomaticFiles) {
		if b.app.HasFile(file) {
			add(file)
		}
	}

	for _, pattern := range miseConfigGlobs {
		matches, err := b.app.FindFiles(pattern)
		if err != nil {
			continue
		}
		for _, match := range matches {
			add(match)
		}
	}

	// For each directory containing a toml config, also check for a co-located mise.lock
	for _, file := range files {
		if !strings.HasSuffix(file, ".toml") {
			continue
		}
		if lockFile := filepath.Join(filepath.Dir(file), "mise.lock"); b.app.HasFile(lockFile) {
			add(lockFile)
		}
	}

	return files
}
