package zig

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/plan"
)

const (
	ZIG_OUT_BIN   = "zig-out/bin"
	ZIG_CACHE_KEY = "zig"
	ZIG_CACHE_DIR = "/root/.cache/zig"
)

// `.name` is an enum literal since Zig 0.14 (`.name = .my_app`) and a string before that
// (`.name = "my_app"`), so both spellings are matched.
var zonNameRegex = regexp.MustCompile(`\.name\s*=\s*(?:\.([A-Za-z_][A-Za-z0-9_]*)|"([^"]+)")`)

type ZigProvider struct{}

func (p *ZigProvider) Name() string {
	return "zig"
}

func (p *ZigProvider) Detect(ctx *generate.GenerateContext) (bool, error) {
	return ctx.App.HasFile("build.zig"), nil
}

func (p *ZigProvider) Initialize(ctx *generate.GenerateContext) error {
	return nil
}

func (p *ZigProvider) CleansePlan(buildPlan *plan.BuildPlan) {}

func (p *ZigProvider) StartCommandHelp() string {
	return "To start your Zig application, Railpack runs the executable that `zig build` writes to zig-out/bin.\n\n" +
		"The executable is named after the `.name` field in build.zig.zon."
}

func (p *ZigProvider) Plan(ctx *generate.GenerateContext) error {
	packages := ctx.GetMiseStepBuilder()
	packages.Default("zig", "latest")
	packages.UseMiseVersions(ctx, []string{"zig"})

	build := ctx.NewCommandStep("build")
	build.AddInput(plan.NewStepLayer(packages.Name()))
	build.AddInput(plan.NewLocalLayer())
	// the global cache holds compiled artifacts that are reused across builds
	build.AddCache(ctx.Caches.AddCache(ZIG_CACHE_KEY, ZIG_CACHE_DIR))
	build.AddCommand(plan.NewExecCommand(fmt.Sprintf("zig build %s", p.releaseFlag(ctx))))

	binaryName := p.binaryName(ctx)

	ctx.Deploy.StartCmd = fmt.Sprintf("./%s/%s", ZIG_OUT_BIN, binaryName)
	ctx.Deploy.AddInputs([]plan.Layer{
		plan.NewStepLayer(build.Name(), plan.NewIncludeFilter([]string{ZIG_OUT_BIN})),
	})

	return nil
}

func (p *ZigProvider) releaseFlag(ctx *generate.GenerateContext) string {
	if mode, _ := ctx.Env.GetConfigVariable("ZIG_RELEASE_MODE"); mode != "" {
		return fmt.Sprintf("--release=%s", mode)
	}

	return "--release=safe"
}

// the executable that `zig build` produces is named after `.name` in build.zig.zon,
// falling back to the app directory when that file is missing or unparseable
func (p *ZigProvider) binaryName(ctx *generate.GenerateContext) string {
	if contents, err := ctx.App.ReadFile("build.zig.zon"); err == nil {
		if name := parseZonName(contents); name != "" {
			return name
		}
	}

	return filepath.Base(ctx.GetAppSource())
}

func parseZonName(contents string) string {
	for _, line := range strings.Split(contents, "\n") {
		// skip commented lines so the documentation in a generated build.zig.zon is not matched
		if strings.HasPrefix(strings.TrimSpace(line), "//") {
			continue
		}

		if match := zonNameRegex.FindStringSubmatch(line); match != nil {
			if match[1] != "" {
				return match[1]
			}

			return match[2]
		}
	}

	return ""
}
