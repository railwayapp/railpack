package zig

import (
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/plan"
)

const (
	ZIG_OUT_BIN   = "zig-out/bin"
	ZIG_CACHE_KEY = "zig"
	ZIG_CACHE_DIR = "/root/.cache/zig"
)

// Tried in order at each position: a `.name` assignment, a string literal, or a comment.
// The assignment has to beat the string pattern so `.name = "my_app"` is captured.
// Strings have to beat comments so "https://..." is not treated as a comment, and a
// `//` comment or a `\\` multiline-string line is consumed whole so a `.name` inside it cannot match.
// `.name` is an enum literal since Zig 0.14 (`.name = .my_app`) and a string before that.
var zonNameToken = regexp.MustCompile(
	`(?m)\.name\s*=\s*(?:\.([A-Za-z_][A-Za-z0-9_]*)|"([^"]+)")|"(?:\\.|[^"\\])*"|//.*|^[ \t]*\\\\.*`,
)

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
	build.AddCommand(plan.NewExecCommand("zig build --release=safe"))

	binaryName := p.binaryName(ctx)

	ctx.Deploy.StartCmd = fmt.Sprintf("./%s/%s", ZIG_OUT_BIN, binaryName)
	ctx.Deploy.AddInputs([]plan.Layer{
		plan.NewStepLayer(build.Name(), plan.NewIncludeFilter([]string{ZIG_OUT_BIN})),
	})

	return nil
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
	for _, match := range zonNameToken.FindAllStringSubmatch(contents, -1) {
		if match[1] != "" {
			return match[1]
		}
		if match[2] != "" {
			return match[2]
		}
	}

	return ""
}
