package dotnet

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"maps"
	"path"
	"regexp"
	"strings"

	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/plan"
	"github.com/railwayapp/railpack/internal/utils"
)

const (
	DEFAULT_DOTNET_VERSION   = "8.0"
	DOTNET_ROOT              = "/mise/installs/dotnet"
	DOTNET_DEPENDENCIES_ROOT = "/root/.nuget/packages"
)

type DotnetProvider struct {
}

func (p *DotnetProvider) Name() string {
	return "Dotnet"
}

func (p *DotnetProvider) Detect(ctx *generate.GenerateContext) (bool, error) {
	return ctx.App.HasMatch("*.csproj"), nil
}

func (p *DotnetProvider) Initialize(ctx *generate.GenerateContext) error {
	return nil
}

func (p *DotnetProvider) Plan(ctx *generate.GenerateContext) error {
	miseStep := ctx.GetMiseStepBuilder()
	p.InstallMisePackages(ctx, miseStep)

	install := ctx.NewCommandStep("install")
	install.AddInput(plan.NewStepLayer(miseStep.Name()))
	p.Install(ctx, install)

	build := ctx.NewCommandStep("build")
	build.AddInput(plan.NewStepLayer(miseStep.Name()))
	build.AddInput(plan.NewStepLayer(install.Name(), plan.Filter{
		Include: []string{"obj/", DOTNET_DEPENDENCIES_ROOT},
	}))
	p.Build(ctx, build)

	envVars := p.GetEnvVars(ctx)
	// Required for internationalization
	ctx.Deploy.AddAptPackages([]string{"libicu-dev"})
	ctx.Deploy.AddInputs([]plan.Layer{
		plan.NewStepLayer(miseStep.Name(), plan.Filter{
			// Need to include the dotnet runtime for the binary to run
			Include: []string{envVars["DOTNET_ROOT"]},
		}),
		plan.NewStepLayer(build.Name(), plan.Filter{
			Include: []string{"out"},
		}),
	})
	ctx.Deploy.StartCmd = p.GetStartCommand(ctx)
	maps.Copy(ctx.Deploy.Variables, envVars)

	return nil
}

func (p *DotnetProvider) CleansePlan(buildPlan *plan.BuildPlan) {}

func (p *DotnetProvider) StartCommandHelp() string {
	return "To start your Dotnet application, Railpack will look for:\n\n" +
		"1. A .csproj file in your project root\n\n" +
		"The project will be run with `./out`"
}

func (p *DotnetProvider) GetStartCommand(ctx *generate.GenerateContext) string {
	projFiles, err := ctx.App.FindFiles("*.csproj")
	if err != nil || len(projFiles) == 0 {
		return ""
	}
	projFile := projFiles[0]
	projName := strings.TrimSuffix(projFile, ".csproj")
	return fmt.Sprintf("./out/%s", projName)
}

func (p *DotnetProvider) Install(ctx *generate.GenerateContext, install *generate.CommandStepBuilder) {
	maps.Copy(install.Variables, p.GetEnvVars(ctx))
	install.AddCommands([]plan.Command{
		plan.NewCopyCommand("nuget.config*"),
		plan.NewCopyCommand("*.csproj"),
		plan.NewCopyCommand("global.json*"),
		plan.NewExecCommand(fmt.Sprintf("mkdir -p %s", DOTNET_DEPENDENCIES_ROOT)),
		plan.NewExecCommand(`dotnet restore`),
	})
}

func (p *DotnetProvider) Build(ctx *generate.GenerateContext, build *generate.CommandStepBuilder) {
	maps.Copy(build.Variables, p.GetEnvVars(ctx))
	build.AddInput(ctx.NewLocalLayer())
	build.AddCommands([]plan.Command{
		plan.NewExecCommand("dotnet publish --no-restore -c Release -o out"),
	})
}

func (p *DotnetProvider) GetEnvVars(ctx *generate.GenerateContext) map[string]string {
	version := p.getDotnetVersion(ctx)
	return map[string]string{
		"ASPNETCORE_ENVIRONMENT":      "production",
		"ASPNETCORE_URLS":             "http://0.0.0.0:3000",
		"DOTNET_CLI_TELEMETRY_OPTOUT": "1",
		"DOTNET_ROOT":                 path.Join(DOTNET_ROOT, version),
	}
}

func (p *DotnetProvider) InstallMisePackages(ctx *generate.GenerateContext, miseStep *generate.MiseStepBuilder) {
	dotnet := miseStep.Default("dotnet", DEFAULT_DOTNET_VERSION)

	if files, err := ctx.App.FindFiles("*.csproj"); err == nil && len(files) > 0 {
		if data, err := ctx.App.ReadFile(files[0]); err == nil {
			var project *Project
			err = xml.Unmarshal([]byte(data), &project)
			if err != nil {
				fmt.Printf("Error parsing XML: %v\n", err)
				return
			}

			for _, pg := range project.PropertyGroups {
				if pg.TargetFramework != "" {
					version := extractVersionFromCsproj(pg.TargetFramework)
					if version != "" {
						miseStep.Version(dotnet, version, "csproj")
						break
					}
				}

				if pg.TargetFrameworks != "" {
					frameworks := strings.Split(pg.TargetFrameworks, ";")
					if len(frameworks) > 0 {
						version := extractVersionFromCsproj(frameworks[0])
						if version != "" {
							miseStep.Version(dotnet, version, "csproj")
							break
						}
					}
				}
			}
		}
	}

	if globalJSON, err := ctx.App.ReadFile("global.json"); err == nil {
		var global *CSharpGlobalJSON
		if err := json.Unmarshal([]byte(globalJSON), &global); err == nil && global.SDK.Version != "" {
			version := utils.ExtractSemverVersion(global.SDK.Version)
			semver, err := utils.ParseSemver(version)

			// The global.json file needs to be respected, unlike csproj which can be fuzzy
			if err == nil {
				version := fmt.Sprintf("%d.%d.%d", semver.Major, semver.Minor, semver.Patch)
				miseStep.Version(dotnet, version, "global.json")
			}
		}
	}

	if envVersion, varName := ctx.Env.GetConfigVariable("DOTNET_VERSION"); envVersion != "" {
		miseStep.Version(dotnet, envVersion, varName)
	}

	miseStep.UseMiseVersions(ctx, []string{"dotnet"})
}

func (p *DotnetProvider) getDotnetVersion(ctx *generate.GenerateContext) string {
	miseStepBuilder := ctx.GetMiseStepBuilder()
	pkg := miseStepBuilder.Resolver.Get("dotnet")
	if pkg != nil && pkg.Version != "" {
		return pkg.Version
	}
	return DEFAULT_DOTNET_VERSION
}

type Project struct {
	PropertyGroups []PropertyGroup `xml:"PropertyGroup"`
}

type PropertyGroup struct {
	TargetFramework  string `xml:"TargetFramework"`
	TargetFrameworks string `xml:"TargetFrameworks"`
}

func extractVersionFromCsproj(framework string) string {
	// Match patterns like net6.0, netcoreapp3.1, net5.0-windows, etc.
	re := regexp.MustCompile(`(?:net|netcoreapp)(\d+\.\d+)`)
	matches := re.FindStringSubmatch(framework)
	if len(matches) > 1 {
		return matches[1]
	}

	return ""
}

type CSharpGlobalJSON struct {
	SDK SDK `json:"sdk"`
}

type SDK struct {
	Version string `json:"version"`
}
