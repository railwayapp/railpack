package elixir

import (
	"bufio"
	"fmt"
	"maps"
	"regexp"
	"strings"

	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/plan"
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

	if miseStep.Resolver.Get("node") != nil {
		nodeModules := ctx.NewCommandStep("node_modules")
		nodeModules.AddInput(plan.NewStepLayer(miseStep.Name()))
		nodeModules.AddCache(ctx.Caches.AddCache("npm-install", "/root/.npm"))
		maps.Copy(nodeModules.Variables, p.GetEnvVars(ctx))
		nodeModulesOutputPaths := p.NodeModules(ctx, nodeModules)

		build.AddInput(plan.NewStepLayer(nodeModules.Name(), plan.Filter{
			Include: nodeModulesOutputPaths,
		}))
	}

	maps.Copy(ctx.Deploy.Variables, p.GetEnvVars(ctx))
	ctx.Deploy.AddInputs([]plan.Layer{
		plan.NewStepLayer(build.Name(), plan.Filter{
			Include: buildOutputPaths,
		}),
	})
	ctx.Deploy.StartCmd = p.GetStartCommand(ctx)

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
	return []string{"deps", "_build", "config", "mix.exs", "mix.lock", MIX_ROOT}
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
		plan.NewExecCommand("mix release"),
	})

	return []string{"_build/prod/rel"}
}

func (p *ElixirProvider) NodeModules(ctx *generate.GenerateContext, build *generate.CommandStepBuilder) []string {
	build.AddCommands([]plan.Command{
		plan.NewExecCommand("mkdir -p assets"),
		plan.NewCopyCommand("assets/package*.json", "assets/"),
	})

	if ctx.App.HasMatch("assets/package-lock.json") {
		build.AddCommand(plan.NewExecCommand("npm ci --prefix assets"))
	} else {
		build.AddCommand(plan.NewExecCommand("npm install --prefix assets"))
	}

	return []string{"assets"}
}

var elixirVersionRegex = regexp.MustCompile(`(elixir:[\s].*[> ])([\w|\.]*)`)

func (p *ElixirProvider) InstallMisePackages(ctx *generate.GenerateContext, miseStep *generate.MiseStepBuilder) {
	elixir := miseStep.Default("elixir", DEFAULT_ELIXIR_VERSION)

	if envVersion, varName := ctx.Env.GetConfigVariable("ELIXIR_VERSION"); envVersion != "" {
		miseStep.Version(elixir, envVersion, varName)
	}

	if versionFile, err := ctx.App.ReadFile(".elixir-version"); err == nil {
		miseStep.Version(elixir, strings.TrimSpace(string(versionFile)), ".elixir-version")
	}

	if mixExs, err := ctx.App.ReadFile("mix.exs"); err == nil {
		if match := elixirVersionRegex.FindStringSubmatch(mixExs); len(match) > 2 {
			version := utils.ExtractSemverVersion(match[2])
			if version != "" {
				fmt.Println("mix.exs", version)
				miseStep.Version(elixir, version, "mix.exs")
			}
		}
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

	if envVersion, varName := ctx.Env.GetConfigVariable("ERLANG_VERSION"); envVersion != "" {
		miseStep.Version(erlang, envVersion, varName)
	}

	if versionFile, err := ctx.App.ReadFile(".erlang-version"); err == nil {
		miseStep.Version(erlang, strings.TrimSpace(string(versionFile)), ".erlang-version")
	}

	versionParts := strings.Split(elixirVersion, "-otp-")
	if len(versionParts) > 1 {
		otpVersion := versionParts[1]
		otpSemverVersion := utils.ExtractSemverVersion(otpVersion)
		if _, err := utils.ParseSemver(otpSemverVersion); err == nil {
			miseStep.Version(erlang, otpSemverVersion, "resolved compatible OTP version")
		}
	}

	if ctx.App.HasMatch("assets/package.json") {
		miseStep.Default("node", "latest")
	}
}

func (p *ElixirProvider) GetEnvVars(ctx *generate.GenerateContext) map[string]string {
	return map[string]string{
		"LANG":                       "en_US.UTF-8",
		"LANGUAGE":                   "en_US:en",
		"LC_ALL":                     "en_US.UTF-8",
		"ELIXIR_ERL_OPTIONS":         "+fnu",
		"MIX_ENV":                    "prod",
		"NODE_ENV":                   "production",
		"NPM_CONFIG_PRODUCTION":      "false",
		"NPM_CONFIG_UPDATE_NOTIFIER": "false",
		"NPM_CONFIG_FUND":            "false",
		"CI":                         "true",
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
