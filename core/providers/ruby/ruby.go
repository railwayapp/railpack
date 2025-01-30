package ruby

import (
	"fmt"
	"strings"

	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/plan"
)

const (
	DEFAULT_RUBY_VERSION = "latest"
)

type RubyProvider struct{}

func (p *RubyProvider) Name() string {
	return "ruby"
}

func (p *RubyProvider) Detect(ctx *generate.GenerateContext) (bool, error) {
	hasRuby := ctx.App.HasMatch("Gemfile")

	return hasRuby, nil
}

func (p *RubyProvider) Plan(ctx *generate.GenerateContext) error {
	packages := p.packages(ctx)

	_ = p.install(ctx, packages)

	ctx.Start.Command = fmt.Sprintf("ruby %s", "app.rb")

	return nil
}

func (p *RubyProvider) packages(ctx *generate.GenerateContext) *generate.MiseStepBuilder {
	packages := ctx.GetMiseStepBuilder()

	ruby := packages.Default("ruby", DEFAULT_RUBY_VERSION)

	if envVersion, varName := ctx.Env.GetConfigVariable("RUBY_VERSION"); envVersion != "" {
		packages.Version(ruby, envVersion, varName)
	}

	if version := p.gemfileRubyVersion(ctx); version != "" {
		packages.Version(ruby, version, "Gemfile.lock")
	}

	return packages
}

func (p *RubyProvider) install(ctx *generate.GenerateContext, packages *generate.MiseStepBuilder) *generate.CommandStepBuilder {
	install := ctx.NewCommandStep("install")
	install.AddCommands([]plan.Command{
		// make sure gem is updated
		plan.NewExecCommand("gem update --system --no-document"),
		// install bundler
		plan.NewExecCommand("gem install -N bundler"),
		plan.NewCopyCommand("Gemfile"),
		plan.NewCopyCommand("Gemfile.lock"),
		plan.NewExecCommand("bundle install"),
	})

	install.DependsOn = []string{packages.DisplayName}

	return install
}

// scan for a Gemfile.lock and return the Ruby version
func (p *RubyProvider) gemfileRubyVersion(ctx *generate.GenerateContext) string {
	if gemfileLock, err := ctx.App.ReadFile("Gemfile.lock"); err == nil {
		foundRubyVersion := false
		lines := strings.Split(string(gemfileLock), "\n")
		for _, line := range lines {
			// Look for the "RUBY VERSION" line
			if strings.HasPrefix(line, "RUBY VERSION") {
				foundRubyVersion = true
				continue
			}

			// The Ruby version follows "RUBY VERSION"
			if foundRubyVersion {
				fields := strings.Fields(line)
				if len(fields) >= 2 && fields[0] == "ruby" {
					fmt.Println("Ruby Version:", fields[1])
					return fields[1]
				}
			}
		}
		// packages.Version(ruby, string(versionFile), "Gemfile.lock")
	}

	return ""
}
