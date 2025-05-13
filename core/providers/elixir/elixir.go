package elixir

import (
	"bufio"
	"fmt"
	"maps"
	"regexp"
	"strings"

	"github.com/railwayapp/railpack/core/app"
	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/plan"
	"github.com/railwayapp/railpack/core/providers/node"
	"github.com/railwayapp/railpack/internal/utils"
)

const (
	DEFAULT_ERLANG_VERSION = "27.3"
	DEFAULT_ELIXIR_VERSION = "1.18"
	APP_BIN_PATH           = "/app/bin/server"
	MIX_ROOT               = "/root/.mix"
)

type ElixirProvider struct {
}

func (p *ElixirProvider) Name() string {
	return "Elixir"
}

func (p *ElixirProvider) Detect(ctx *generate.GenerateContext) (bool, error) {
	hasMixFile := ctx.App.HasMatch("mix.exs")
	return hasMixFile, nil
}

func (p *ElixirProvider) Initialize(ctx *generate.GenerateContext) error {
	return nil
}

func (p *ElixirProvider) Plan(ctx *generate.GenerateContext) error {
	miseStep := ctx.GetMiseStepBuilder()
	p.InstallMisePackages(ctx, miseStep)

	install := ctx.NewCommandStep("install")
	install.AddInput(plan.NewStepLayer(miseStep.Name()))
	installOutputPaths := p.Install(ctx, install)
	maps.Copy(install.Variables, p.GetEnvVars(ctx))

	build := ctx.NewCommandStep("build")
	build.AddInput(plan.NewStepLayer(miseStep.Name()))
	build.AddInput(plan.NewStepLayer(install.Name(), plan.Filter{
		Include: installOutputPaths,
	}))
	maps.Copy(build.Variables, p.GetEnvVars(ctx))
	buildOutputPaths := p.Build(ctx, build)

	maps.Copy(ctx.Deploy.Variables, p.GetEnvVars(ctx))
	ctx.Deploy.AddInputs([]plan.Layer{
		plan.NewStepLayer(build.Name(), plan.Filter{
			Include: buildOutputPaths,
		}),
	})
	ctx.Deploy.StartCmd = p.GetStartCommand(ctx)

	// Node (if necessary)
	if err := p.InstallNode(ctx, build); err != nil {
		return err
	}

	return nil
}

func (p *ElixirProvider) StartCommandHelp() string {
	return "To start your Elixir application, Railpack will look for:\n\n" +
		"1. A mix.exs file in your project root\n\n" +
		"The start command will run your application server using the generated release."
}

func (p *ElixirProvider) GetStartCommand(ctx *generate.GenerateContext) string {
	binName := p.findBinName(ctx)
	return fmt.Sprintf("/app/_build/prod/rel/%s/bin/%s start", binName, binName)
}

func (p *ElixirProvider) Install(ctx *generate.GenerateContext, install *generate.CommandStepBuilder) []string {
	install.AddCommands([]plan.Command{
		plan.NewExecCommand("mix local.hex --force"),
		plan.NewExecCommand("mix local.rebar --force"),
		plan.NewCopyCommand("mix.exs"),
		plan.NewCopyCommand("mix.lock"),
		plan.NewExecCommand("mix deps.get --only prod"),
		plan.NewExecCommand("mkdir -p config"),
		plan.NewCopyCommand("config/config.exs*", "config/"),
		plan.NewCopyCommand("config/prod.exs*", "config/"),
		plan.NewExecCommand("mix deps.compile"),
	})
	if matches := ctx.App.FindFilesWithContent("mix.exs", regexp.MustCompile(`assets\.setup`)); len(matches) > 0 {
		install.AddCommand(plan.NewExecCommand("mix assets.setup"))
	}
	return []string{"deps", "_build", "config", "mix.exs", "mix.lock", MIX_ROOT}
}

func (p *ElixirProvider) InstallNode(ctx *generate.GenerateContext, build *generate.CommandStepBuilder) error {
	// All providers assume they're running in the application root
	// but Phoenix puts it in the assets folder, so we have to lie to the provider
	assetsApp, err := app.NewApp(ctx.App.Source + "/assets")
	if err != nil {
		// If the assets folder doesn't exist, then it isn't an error, we just don't need to install Node
		return nil
	}
	defer func(originalApp *app.App) { ctx.App = originalApp }(ctx.App)
	ctx.App = assetsApp

	nodeProvider := node.NodeProvider{}
	isNode, err := nodeProvider.Detect(ctx)
	if err != nil {
		return err
	}
	if !isNode {
		return nil
	}

	err = nodeProvider.Initialize(ctx)
	if err != nil {
		return err
	}

	miseStep := ctx.GetMiseStepBuilder()
	nodeProvider.InstallMisePackages(ctx, miseStep)

	installNode := ctx.NewCommandStep("install:node")
	installNode.AddInput(plan.NewStepLayer(miseStep.Name()))
	nodeProvider.InstallNodeDeps(ctx, installNode)

	// Again, the provider thinks it's in the root folder, but is actually in assets
	// So we have to modify all copy commands
	for idx, cmd := range installNode.Commands {
		if copyCmd, ok := cmd.(plan.CopyCommand); ok {
			copyCmd.Src = "assets/" + copyCmd.Src
			installNode.Commands[idx] = copyCmd
		}
	}

	// esbuild knows how to load node_modules from the root, so we don't have to copy it to the assets folder
	build.AddInput(plan.NewStepLayer(installNode.Name(), plan.Filter{
		Include: []string{"node_modules"},
	}))

	return nil
}

func (p *ElixirProvider) Build(ctx *generate.GenerateContext, build *generate.CommandStepBuilder) []string {
	build.AddCommands([]plan.Command{
		plan.NewCopyCommand("priv*", "."),
		plan.NewCopyCommand("lib*", "."),
		plan.NewCopyCommand("assets*", "."),
	})
	if matches := ctx.App.FindFilesWithContent("mix.exs", regexp.MustCompile(`assets\.deploy`)); len(matches) > 0 {
		build.AddCommand(plan.NewExecCommand("mix assets.deploy"))
	}
	if matches := ctx.App.FindFilesWithContent("mix.exs", regexp.MustCompile(`ecto\.deploy`)); len(matches) > 0 {
		build.AddCommand(plan.NewExecCommand("mix ecto.deploy"))
	}
	build.AddCommands([]plan.Command{
		plan.NewExecCommand("mix compile"),
		plan.NewCopyCommand("config/runtime.exs*", "config/"),
		plan.NewCopyCommand("rel*", "."),
		plan.NewExecCommand("mix release"),
	})

	return []string{"_build/prod/rel"}
}

var elixirVersionRegex = regexp.MustCompile(`(elixir:[\s].*[> ])([\w|\.]*)`)

func (p *ElixirProvider) InstallMisePackages(ctx *generate.GenerateContext, miseStep *generate.MiseStepBuilder) {
	elixir := miseStep.Default("elixir", DEFAULT_ELIXIR_VERSION)

	if mixExs, err := ctx.App.ReadFile("mix.exs"); err == nil {
		if match := elixirVersionRegex.FindStringSubmatch(mixExs); len(match) > 2 {
			version := utils.ExtractSemverVersion(match[2])
			if version != "" {
				miseStep.Version(elixir, version, "mix.exs")
			}
		}
	}

	if versionFile, err := ctx.App.ReadFile(".elixir-version"); err == nil {
		miseStep.Version(elixir, strings.TrimSpace(string(versionFile)), ".elixir-version")
	}

	if envVersion, varName := ctx.Env.GetConfigVariable("ELIXIR_VERSION"); envVersion != "" {
		miseStep.Version(elixir, envVersion, varName)
	}

	pkgs, err := miseStep.Resolver.ResolvePackages()
	erlang := miseStep.Default("erlang", DEFAULT_ERLANG_VERSION)
	elixirVersion := DEFAULT_ELIXIR_VERSION
	if err == nil && pkgs["elixir"] != nil && pkgs["elixir"].ResolvedVersion != nil {
		elixirVersion = *pkgs["elixir"].ResolvedVersion
	}

	elixirSemverVersion := utils.ExtractSemverVersion(elixirVersion)
	elixirSemver, err := utils.ParseSemver(elixirSemverVersion)

	if err == nil {
		compatibleErlangVersion := getCompatibleErlangVersion(fmt.Sprintf("%d.%d", elixirSemver.Major, elixirSemver.Minor))
		miseStep.Version(erlang, compatibleErlangVersion, "default compatible OTP version")
	}

	versionParts := strings.Split(elixirVersion, "-otp-")
	if len(versionParts) > 1 {
		otpVersion := versionParts[1]
		otpSemverVersion := utils.ExtractSemverVersion(otpVersion)
		if _, err := utils.ParseSemver(otpSemverVersion); err == nil {
			miseStep.Version(erlang, otpSemverVersion, "resolved compatible OTP version")
		}
	}

	if versionFile, err := ctx.App.ReadFile(".erlang-version"); err == nil {
		miseStep.Version(erlang, strings.TrimSpace(string(versionFile)), ".erlang-version")
	}

	if envVersion, varName := ctx.Env.GetConfigVariable("ERLANG_VERSION"); envVersion != "" {
		miseStep.Version(erlang, envVersion, varName)
	}
}

func (p *ElixirProvider) GetEnvVars(ctx *generate.GenerateContext) map[string]string {
	return map[string]string{
		"LANG":               "en_US.UTF-8",
		"LANGUAGE":           "en_US:en",
		"LC_ALL":             "en_US.UTF-8",
		"ELIXIR_ERL_OPTIONS": "+fnu",
		"MIX_ENV":            "prod",
	}
}

func (p *ElixirProvider) findBinName(ctx *generate.GenerateContext) string {
	configFile, err := ctx.App.ReadFile("mix.exs")
	if err != nil {
		return ""
	}

	scanner := bufio.NewScanner(strings.NewReader(configFile))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "app: :") {
			binName := strings.Split(strings.Replace(line, "app:", "", 1), ":")[1]
			binName = strings.TrimSpace(strings.Trim(binName, ",'\""))
			return binName
		}
	}

	if err := scanner.Err(); err != nil {
		return ""
	}

	return ""
}

// See: https://hexdocs.pm/elixir/1.18.3/compatibility-and-deprecations.html#between-elixir-and-erlang-otp
func getCompatibleErlangVersion(elixirVersion string) string {
	switch elixirVersion {
	case "1.0", "1.1":
		return "18"
	case "1.2", "1.3":
		return "19"
	case "1.4":
		return "20"
	case "1.5":
		return "20"
	case "1.6":
		return "21"
	case "1.7", "1.8", "1.9":
		return "22"
	case "1.10":
		return "23"
	case "1.11", "1.12":
		return "24"
	case "1.13":
		return "25"
	case "1.14":
		return "26"
	case "1.15", "1.16":
		return "26"
	case "1.17", "1.18":
		return "27"
	default:
		return DEFAULT_ERLANG_VERSION
	}
}
