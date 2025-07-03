package ruby

import (
	"bufio"
	"fmt"
	"maps"
	"regexp"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/plan"
	"github.com/railwayapp/railpack/core/providers/node"
	"github.com/railwayapp/railpack/internal/utils"
)

const (
	DEFAULT_RUBY_VERSION = "3.4.2"
)

type RubyProvider struct{}

func (p *RubyProvider) Name() string {
	return "ruby"
}

func (p *RubyProvider) Initialize(ctx *generate.GenerateContext) error {
	return nil
}

func (p *RubyProvider) Detect(ctx *generate.GenerateContext) (bool, error) {
	hasRuby := ctx.App.HasMatch("Gemfile")
	return hasRuby, nil
}

func (p *RubyProvider) Plan(ctx *generate.GenerateContext) error {
	miseStep := ctx.GetMiseStepBuilder()
	p.InstallMisePackages(ctx, miseStep)

	install := ctx.NewCommandStep("install")
	install.AddInput(plan.NewStepLayer(miseStep.Name()))
	installOutputs := p.Install(ctx, install)
	p.addMetadata(ctx)

	nodeProvider := &node.NodeProvider{}
	nodeDetected, err := nodeProvider.Detect(ctx)
	if err != nil {
		return err
	}
	if nodeDetected || p.usesDep(ctx, "execjs") {
		nodeProvider.InstallMisePackages(ctx, miseStep)
	}

	var (
		buildNode *generate.CommandStepBuilder
		pruneNode *generate.CommandStepBuilder
	)
	if nodeDetected {
		err := nodeProvider.Initialize(ctx)
		if err != nil {
			return err
		}

		nodeProvider.InstallMisePackages(ctx, miseStep)
		installNode := ctx.NewCommandStep("install:node")
		installNode.AddInput(plan.NewStepLayer(miseStep.Name()))
		nodeProvider.InstallNodeDeps(ctx, installNode)

		pruneNode = ctx.NewCommandStep("prune:node")
		pruneNode.AddInput(plan.NewStepLayer(installNode.Name()))
		nodeProvider.PruneNodeDeps(ctx, pruneNode)

		buildNode = ctx.NewCommandStep("build:node")
		buildNode.AddInputs([]plan.Layer{
			plan.NewStepLayer(install.Name()),
			plan.NewStepLayer(installNode.Name(), plan.Filter{
				Include: append([]string{"."}, miseStep.GetOutputPaths()...),
			}),
		})
		nodeProvider.Build(ctx, buildNode)
	}

	build := ctx.NewCommandStep("build")
	build.AddInput(plan.NewStepLayer(install.Name()))
	buildOutputs := p.Build(ctx, build)

	ctx.Deploy.StartCmd = p.GetStartCommand(ctx)
	maps.Copy(ctx.Deploy.Variables, p.GetRubyEnvVars(ctx))
	p.AddRuntimeDeps(ctx)

	ctx.Deploy.AddInputs([]plan.Layer{
		plan.NewStepLayer(miseStep.Name(), plan.Filter{
			Include: miseStep.GetOutputPaths(),
		}),
		plan.NewStepLayer(install.Name(), plan.Filter{
			Include: installOutputs,
		}),
		plan.NewStepLayer(build.Name(), plan.Filter{
			Include: buildOutputs,
		}),
	})

	if buildNode != nil && pruneNode != nil {
		ctx.Deploy.AddInputs([]plan.Layer{
			plan.NewStepLayer(pruneNode.Name(), plan.Filter{
				Include: []string{"/app/node_modules"},
			}),
			plan.NewStepLayer(buildNode.Name(), plan.Filter{
				Include: []string{"."},
				Exclude: []string{"node_modules", ".yarn"},
			}),
		})
	}

	return nil
}

func (p *RubyProvider) GetStartCommand(ctx *generate.GenerateContext) string {
	startCommand := ""
	app := ctx.App

	if p.usesRails(ctx) {
		if app.HasMatch("rails") {
			return "bundle exec rails server -b 0.0.0.0 -p ${PORT:-3000}"
		} else {
			return "bundle exec bin/rails server -b 0.0.0.0 -p ${PORT:-3000} -e $RAILS_ENV"
		}
	} else if app.HasMatch("config/environment.rb") && app.HasMatch("script") {
		return "bundle exec ruby script/server -p ${PORT:-3000}"
	} else if app.HasMatch("config.ru") {
		return "bundle exec rackup config.ru -p ${PORT:-3000}"
	} else if app.HasMatch("Rakefile") {
		return "bundle exec rake"
	}

	return startCommand
}

func (p *RubyProvider) StartCommandHelp() string {
	return "To start your Ruby application, Railpack will automatically:\n\n" +
		"1. Start the Rails server if a Rails application is detected\n" +
		"2. Start the Ruby server if a config/environment.rb and script/server file is found\n" +
		"3. Start the Rack server if a config.ru file is found\n" +
		"4. Run the Rakefile if a Rakefile is found\n\n" +
		"Otherwise, it will not start any server by default."
}

func (p *RubyProvider) Install(ctx *generate.GenerateContext, install *generate.CommandStepBuilder) []string {
	install.Secrets = []string{}
	install.UseSecretsWithPrefixes([]string{"RUBY", "GEM", "BUNDLE"})
	envVars := p.GetRubyEnvVars(ctx)
	install.AddEnvVars(envVars)
	bundlerVersion := parseBundlerVersionFromGemfile(ctx)
	commands := []plan.Command{
		plan.NewExecCommand(fmt.Sprintf("gem install -N %s", bundlerVersion)),
		plan.NewCopyCommand("Gemfile"),
		plan.NewCopyCommand("Gemfile.lock"),
	}

	for _, path := range parseLocalPathsFromGemfile(ctx) {
		commands = append(commands, plan.NewCopyCommand(path))
	}

	commands = append(commands, plan.NewExecCommand("bundle install"))

	if p.usesDep(ctx, "bootsnap") {
		commands = append(commands, plan.NewExecCommand("bundle exec bootsnap precompile --gemfile"))
	}

	install.AddCommands(commands)
	install.AddPaths([]string{envVars["GEM_PATH"]})
	return []string{envVars["GEM_HOME"]}
}

func (p *RubyProvider) Build(ctx *generate.GenerateContext, build *generate.CommandStepBuilder) []string {
	build.Secrets = []string{}
	build.UseSecretsWithPrefixes([]string{"RAILS", "BUNDLE", "BOOTSNAP", "SPROCKETS", "WEBPACKER", "ASSET", "DISABLE_SPRING"})
	build.AddEnvVars(p.GetRubyEnvVars(ctx))
	build.AddCommand(plan.NewCopyCommand("."))
	outputs := []string{"/app"}
	// Only compile assets if a Rails app have an asset pipeline gem
	// installed (e.g. sprockets, propshaft). Rails API-only apps [0]
	// do not come with the asset pipelines because they have no assets.
	// [0] https://guides.rubyonrails.org/api_app.html
	if p.usesRails(ctx) && p.usesAssetPipeline(ctx) {
		build.AddCommand(plan.NewExecCommand("bundle exec rake assets:precompile"))
	}

	if p.usesRails(ctx) && p.usesDep(ctx, "bootsnap") {
		build.AddCommand(plan.NewExecCommand("bundle exec bootsnap precompile app/ lib/"))
		outputs = append(outputs, "lib/")
	}

	return outputs
}

func (p *RubyProvider) AddRuntimeDeps(ctx *generate.GenerateContext) {
	packages := []string{"libyaml-dev"}

	if p.usesPostgres(ctx) {
		packages = append(packages, "libpq-dev")
	}

	if p.usesMysql(ctx) {
		packages = append(packages, "default-libmysqlclient-dev")
	}

	if p.usesDep(ctx, "magick") {
		packages = append(packages, "libmagickwand-dev")
	}

	if p.usesDep(ctx, "vips") {
		packages = append(packages, "libvips-dev")
	}

	if p.usesDep(ctx, "charlock_holmes") {
		packages = append(packages, "libicu-dev", "libxml2-dev", "libxslt-dev")
	}

	ctx.Deploy.AddAptPackages(packages)
}

func (p *RubyProvider) GetBuilderDeps(ctx *generate.GenerateContext) *generate.MiseStepBuilder {
	miseStep := ctx.GetMiseStepBuilder()
	miseStep.SupportingAptPackages = append(miseStep.SupportingAptPackages, "procps")

	if p.usesPostgres(ctx) {
		miseStep.SupportingAptPackages = append(miseStep.SupportingAptPackages, "libpq-dev")
	}

	if p.usesMysql(ctx) {
		miseStep.SupportingAptPackages = append(miseStep.SupportingAptPackages, "default-libmysqlclient-dev")
	}

	return miseStep
}

func (p *RubyProvider) InstallMisePackages(ctx *generate.GenerateContext, miseStep *generate.MiseStepBuilder) {
	ruby := miseStep.Default("ruby", DEFAULT_RUBY_VERSION)

	if envVersion, varName := ctx.Env.GetConfigVariable("RUBY_VERSION"); envVersion != "" {
		miseStep.Version(ruby, envVersion, varName)
	}

	if versionFile, err := ctx.App.ReadFile(".ruby-version"); err == nil {
		miseStep.Version(ruby, utils.ExtractSemverVersion(string(versionFile)), ".ruby-version")
	}

	if constraint := extractRubyConstraintFromGemfile(ctx); constraint != "" {
		allVersions, err := miseStep.Resolver.GetMise().GetAllVersions("ruby", "")
		if err == nil && len(allVersions) > 0 {
			constraintObj, err := semver.NewConstraint(constraint)
			if err == nil {
				var matchingVersions []*semver.Version
				for _, vstr := range allVersions {
					if strings.Contains(vstr, "-dev") || strings.Contains(vstr, "-preview") ||
						strings.Contains(vstr, "-rc") || strings.Contains(vstr, "dev") {
						continue
					}
					v, err := semver.NewVersion(vstr)
					if err == nil && constraintObj.Check(v) {
						matchingVersions = append(matchingVersions, v)
					}
				}
				if len(matchingVersions) > 0 {
					sort.Sort(sort.Reverse(semver.Collection(matchingVersions)))
					miseStep.Version(ruby, matchingVersions[0].String(), "Gemfile")
				}
			}
		}
	}

	miseStep.AddSupportingAptPackage("libyaml-dev")
	version := p.getRubyVersion(ctx)
	version = utils.ExtractSemverVersion(version)
	semver, err := utils.ParseSemver(version)
	// YJIT in Ruby 3.1+ requires rustc to install
	if err == nil && semver != nil && semver.Major >= 3 && semver.Minor > 1 {
		miseStep.AddSupportingAptPackage("rustc")
		miseStep.AddSupportingAptPackage("cargo")
	}
}

func (p *RubyProvider) getRubyVersion(ctx *generate.GenerateContext) string {
	miseStepBuilder := ctx.GetMiseStepBuilder()
	pkg := miseStepBuilder.Resolver.Get("ruby")
	if pkg != nil && pkg.Version != "" {
		return pkg.Version
	}
	return DEFAULT_RUBY_VERSION
}

func (p *RubyProvider) GetRubyEnvVars(ctx *generate.GenerateContext) map[string]string {
	return map[string]string{
		"BUNDLE_GEMFILE":   "/app/Gemfile",
		"GEM_PATH":         "/usr/local/bundle",
		"GEM_HOME":         "/usr/local/bundle",
		"MALLOC_ARENA_MAX": "2",
	}
}

func (p *RubyProvider) usesPostgres(ctx *generate.GenerateContext) bool {
	return p.usesDep(ctx, "pg")
}

func (p *RubyProvider) usesMysql(ctx *generate.GenerateContext) bool {
	return p.usesDep(ctx, "mysql")
}

func (p *RubyProvider) usesRails(ctx *generate.GenerateContext) bool {
	contents, err := ctx.App.ReadFile("config/application.rb")
	if err != nil {
		return false
	}

	return strings.Contains(contents, "Rails::Application")
}

func (p *RubyProvider) usesAssetPipeline(ctx *generate.GenerateContext) bool {
	return p.usesDep(ctx, "sprockets") || p.usesDep(ctx, "propshaft")
}

func (p *RubyProvider) usesDep(ctx *generate.GenerateContext, dep string) bool {
	for _, file := range []string{"Gemfile", "Gemfile.lock"} {
		if contents, err := ctx.App.ReadFile(file); err == nil {
			if strings.Contains(string(contents), dep) {
				return true
			}

		}
	}
	return false
}

func (p *RubyProvider) addMetadata(ctx *generate.GenerateContext) {
	ctx.Metadata.SetBool("rubyRails", p.usesRails(ctx))
	ctx.Metadata.SetBool("rubyAssetPipeline", p.usesAssetPipeline(ctx))
	ctx.Metadata.SetBool("rubyBootsnap", p.usesDep(ctx, "bootsnap"))
}

var (
	gemfileVersionRegex = regexp.MustCompile(`ruby\s+(?:'|")([^'"]+)(?:'|")`)
)

func parseBundlerVersionFromGemfile(ctx *generate.GenerateContext) string {
	gemfileLock, err := ctx.App.ReadFile("Gemfile.lock")
	if err != nil {
		return "bundler"
	}

	scanner := bufio.NewScanner(strings.NewReader(gemfileLock))

	foundBundledWith := false
	for scanner.Scan() {
		line := scanner.Text()

		if foundBundledWith {
			return fmt.Sprintf("bundler:%s", strings.TrimSpace(line))
		}

		if strings.Contains(line, "BUNDLED WITH") {
			foundBundledWith = true
		}
	}

	// note: we don't need to worry about the scanner error because we are just returning "bundler" in that case anyway
	return "bundler"
}

func parseLocalPathsFromGemfile(ctx *generate.GenerateContext) []string {
	gemfile, err := ctx.App.ReadFile("Gemfile")
	if err != nil {
		return []string{}
	}

	scanner := bufio.NewScanner(strings.NewReader(string(gemfile)))
	paths := []string{}

	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "path:") {
			p := strings.TrimSpace(strings.Split(line, "path:")[1])
			paths = append(paths, strings.Trim(p, "\""))
		}
	}

	return paths
}

func extractRubyConstraintFromGemfile(ctx *generate.GenerateContext) string {
	gemfile, err := ctx.App.ReadFile("Gemfile")
	if err != nil {
		return ""
	}
	matches := gemfileVersionRegex.FindStringSubmatch(string(gemfile))
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}
