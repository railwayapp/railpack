package golang

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/plan"
)

const (
	DEFAULT_GO_VERSION = "1.23"
	GO_BUILD_CACHE_KEY = "go-build"
	GO_BINARY_NAME     = "out"
	GO_PATH            = "/go"
)

type GoProvider struct{}

func (p *GoProvider) Name() string {
	return "golang"
}

func (p *GoProvider) Detect(ctx *generate.GenerateContext) (bool, error) {
	return p.isGoMod(ctx) || ctx.App.HasMatch("main.go"), nil
}

func (p *GoProvider) Initialize(ctx *generate.GenerateContext) error {
	return nil
}

func (p *GoProvider) Plan(ctx *generate.GenerateContext) error {
	builder := p.GetBuilder(ctx)
	p.InstallGoPackages(ctx, builder)

	install := ctx.NewCommandStep("install")
	install.AddInput(plan.NewStepLayer(builder.Name()))
	p.InstallGoDeps(ctx, install)

	build := ctx.NewCommandStep("build")
	build.AddInput(plan.NewStepLayer(install.Name()))
	p.Build(ctx, build)

	ctx.Deploy.StartCmd = fmt.Sprintf("./%s", GO_BINARY_NAME)

	runtimePkgs := []string{"tzdata"}
	if p.hasCGOEnabled(ctx) {
		ctx.Logger.LogInfo("CGO is enabled")
		runtimePkgs = append(runtimePkgs, "libc6")
	}

	ctx.Deploy.AddAptPackages(runtimePkgs)
	ctx.Deploy.AddInputs([]plan.Layer{
		plan.NewStepLayer(build.Name(), plan.Filter{
			Include: []string{"."},
		}),
	})

	p.addMetadata(ctx)

	return nil
}

func (p *GoProvider) Build(ctx *generate.GenerateContext, build *generate.CommandStepBuilder) {
	var buildCmd string

	flags := "-w -s"
	baseBuildCmd := fmt.Sprintf("go build -ldflags=\"%s\" -o %s", flags, GO_BINARY_NAME)

	if binName, _ := ctx.Env.GetConfigVariable("GO_BIN"); binName != "" {
		buildCmd = fmt.Sprintf("%s ./cmd/%s", baseBuildCmd, binName)
	} else if p.isGoMod(ctx) && p.hasRootGoFiles(ctx) {
		buildCmd = baseBuildCmd
	} else if dirs, err := ctx.App.FindDirectories("cmd/*"); err == nil && len(dirs) > 0 {
		// Try to find a command in the cmd directory
		cmdName := filepath.Base(dirs[0])
		buildCmd = fmt.Sprintf("%s ./cmd/%s", baseBuildCmd, cmdName)
	} else if p.isGoMod(ctx) {
		buildCmd = baseBuildCmd
	} else if ctx.App.HasMatch("main.go") {
		buildCmd = fmt.Sprintf("%s main.go", baseBuildCmd)
	}

	build.AddCommand(plan.NewCopyCommand("."))

	if buildCmd == "" {
		return
	}

	build.AddCache(p.goBuildCache(ctx))
	build.AddCommands([]plan.Command{
		plan.NewExecCommand(buildCmd),
	})
}

func (p *GoProvider) InstallGoDeps(ctx *generate.GenerateContext, install *generate.CommandStepBuilder) {
	install.AddEnvVars(map[string]string{
		"GOPATH": GO_PATH,
		"GOBIN":  fmt.Sprintf("%s/bin", GO_PATH),
	})
	install.AddCommands([]plan.Command{
		plan.NewPathCommand(fmt.Sprintf("%s/bin", GO_PATH)),
	})

	if !p.isGoMod(ctx) {
		return
	}

	install.AddCache(p.goBuildCache(ctx))
	install.AddCommand(plan.NewCopyCommand("go.mod"))

	if ctx.App.HasMatch("go.sum") {
		install.AddCommand(plan.NewCopyCommand("go.sum"))
	}

	install.AddCommand(plan.NewExecCommand("go mod download"))

	ctx.Logger.LogInfo("Using go mod")

	if !p.hasCGOEnabled(ctx) {
		install.AddEnvVars(map[string]string{"CGO_ENABLED": "0"})
	}
}

func (p *GoProvider) InstallGoPackages(ctx *generate.GenerateContext, miseStep *generate.MiseStepBuilder) {
	goPkg := miseStep.Default("go", DEFAULT_GO_VERSION)

	if goModContents, err := ctx.App.ReadFile("go.mod"); err == nil {
		// Split content into lines and look for "go X.XX" line
		lines := strings.Split(string(goModContents), "\n")
		for _, line := range lines {
			if strings.HasPrefix(strings.TrimSpace(line), "go ") {
				// Extract version number
				if goVersion := strings.TrimSpace(strings.TrimPrefix(line, "go")); goVersion != "" {
					miseStep.Version(goPkg, goVersion, "go.mod")
					break
				}
			}
		}
	}

	if envVersion, varName := ctx.Env.GetConfigVariable("GO_VERSION"); envVersion != "" {
		miseStep.Version(goPkg, envVersion, varName)
	}
}

func (p *GoProvider) GetBuilder(ctx *generate.GenerateContext) *generate.MiseStepBuilder {
	miseStep := ctx.GetMiseStepBuilder()

	if p.hasCGOEnabled(ctx) {
		miseStep.SupportingAptPackages = append(miseStep.SupportingAptPackages, "gcc", "g++", "libc6-dev")
	}

	return miseStep
}

func (p *GoProvider) addMetadata(ctx *generate.GenerateContext) {
	ctx.Metadata.SetBool("goMod", p.isGoMod(ctx))
	ctx.Metadata.SetBool("goRootFile", p.hasRootGoFiles(ctx))
	ctx.Metadata.SetBool("goGin", p.isGin(ctx))
	ctx.Metadata.SetBool("goCGO", p.hasCGOEnabled(ctx))
}

func (p *GoProvider) goBuildCache(ctx *generate.GenerateContext) string {
	return ctx.Caches.AddCache(GO_BUILD_CACHE_KEY, "/root/.cache/go-build")
}

func (p *GoProvider) hasRootGoFiles(ctx *generate.GenerateContext) bool {
	if files, err := ctx.App.FindFiles("*.go"); err == nil {
		for _, file := range files {
			if filepath.Dir(file) == "." {
				return true
			}
		}
	}
	return false
}

func (p *GoProvider) isGin(ctx *generate.GenerateContext) bool {
	if goModContents, err := ctx.App.ReadFile("go.mod"); err == nil {
		return strings.Contains(string(goModContents), "github.com/gin-gonic/gin")
	}

	return false
}

func (p *GoProvider) hasCGOEnabled(ctx *generate.GenerateContext) bool {
	return ctx.Env.GetVariable("CGO_ENABLED") == "1"
}

func (p *GoProvider) isGoMod(ctx *generate.GenerateContext) bool {
	return ctx.App.HasMatch("go.mod")
}

func (p *GoProvider) StartCommandHelp() string {
	return "To configure your start command, Railpack will check:\n\n" +
		"1. Create a main.go file in your project root\n\n" +
		"2. Create a command in the cmd directory (e.g., cmd/server/main.go)\n\n" +
		"3. Set the GO_BIN environment variable to specify which command to build"
}
