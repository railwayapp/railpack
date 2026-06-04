// this provider is distinct from the SPA functionality used by node providers
// it is meant to simply serve static files over HTTP

package staticfile

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/plan"
)

//go:embed Caddyfile.template
var caddyfileTemplate string

const (
	StaticfileConfigName = "Staticfile"
	CaddyfilePath        = "Caddyfile"
)

// Staticfile in a directory is parsed into this structure
type StaticfileConfig struct {
	RootDir string `yaml:"root"`
	// IndexFallback controls whether to fall back to index.html for unmatched routes (default: true)
	IndexFallback *bool `yaml:"index_fallback"`
}

type StaticfileProvider struct{}

func (p *StaticfileProvider) Name() string {
	return "staticfile"
}

func (p *StaticfileProvider) Initialize(ctx *generate.GenerateContext) error {
	return nil
}

func (p *StaticfileProvider) Detect(ctx *generate.GenerateContext) (bool, error) {
	_, err := getRootDir(ctx)
	if err == nil {
		return true, nil
	}

	return false, nil
}

func (p *StaticfileProvider) Plan(ctx *generate.GenerateContext) error {
	rootDir, err := getRootDir(ctx)
	if err != nil {
		return err
	}

	// NOTE `latest` is not used intentionally here as the Caddyfile template should be tested against major caddy upgrades
	installCaddyStep := ctx.NewInstallBinStepBuilder("packages:caddy")
	installCaddyStep.Default("caddy", "2")

	build := ctx.NewCommandStep("build")
	build.AddInput(plan.NewStepLayer(installCaddyStep.Name()))
	build.AddInput(ctx.NewLocalLayer())

	if err := p.addCaddyfileToStep(ctx, build, rootDir, getIndexFallback(ctx)); err != nil {
		return err
	}

	ctx.Deploy.AddInputs([]plan.Layer{
		installCaddyStep.GetLayer(),
		plan.NewStepLayer(build.Name(), plan.Filter{
			Include: []string{"."},
		}),
	})

	ctx.Deploy.StartCmd = fmt.Sprintf("caddy run --config %s --adapter caddyfile 2>&1", CaddyfilePath)

	return nil
}

func (p *StaticfileProvider) CleansePlan(buildPlan *plan.BuildPlan) {}

func (p *StaticfileProvider) StartCommandHelp() string {
	return "Railpack serves static files using Caddy. To configure the static file root, Railpack will check:\n\n" +
		"1. The RAILPACK_STATIC_FILE_ROOT environment variable\n\n" +
		"2. The \"root\" field in a Staticfile in your project root:\n" +
		"   root: dist\n\n" +
		"3. A \"public\" directory\n\n" +
		"4. An index.html in your project root\n\n" +
		"To enable SPA-style index.html fallback for unmatched routes, set \"index_fallback: true\" in your Staticfile."
}

func (p *StaticfileProvider) addCaddyfileToStep(ctx *generate.GenerateContext, setup *generate.CommandStepBuilder, rootDir string, indexFallback bool) error {
	ctx.Logger.LogInfo("Using staticfile root dir: %s", rootDir)

	caddyfileTemplateVariables := map[string]any{
		"StaticFileRoot": rootDir,
		"IndexFallback":  indexFallback,
	}

	caddyfileTemplate, err := ctx.TemplateFiles([]string{"Caddyfile.template", "Caddyfile"}, caddyfileTemplate, caddyfileTemplateVariables)
	if err != nil {
		return err
	}

	if caddyfileTemplate.Filename != "" {
		ctx.Logger.LogInfo("Using custom Caddyfile: %s", caddyfileTemplate.Filename)
	}

	setup.AddCommands([]plan.Command{
		plan.NewFileCommand(CaddyfilePath, "Caddyfile"),
		plan.NewExecCommand("caddy fmt --overwrite Caddyfile"),
	})

	setup.Assets = map[string]string{
		"Caddyfile": caddyfileTemplate.Contents,
	}

	return nil
}

// only returns an empty string if no root dir can be found
func getRootDir(ctx *generate.GenerateContext) (string, error) {
	if rootDir, _ := ctx.Env.GetConfigVariable("STATIC_FILE_ROOT"); rootDir != "" {
		rootDir = strings.TrimSpace(rootDir)
		if rootDir != "" {
			return rootDir, nil
		}
	}

	staticfileConfig, err := getStaticfileConfig(ctx)
	if staticfileConfig != nil && err == nil {
		rootDir := strings.TrimSpace(staticfileConfig.RootDir)
		if rootDir != "" {
			return rootDir, nil
		}
	}

	if ctx.App.HasMatch("public") {
		return "public", nil
	} else if ctx.App.HasFile("index.html") {
		return ".", nil
	}

	return "", fmt.Errorf("no static file root dir found")
}

// IndexFallbackFromStaticfile returns index_fallback from a Staticfile when explicitly set.
func IndexFallbackFromStaticfile(ctx *generate.GenerateContext) *bool {
	config, err := getStaticfileConfig(ctx)
	if config != nil && err == nil && config.IndexFallback != nil {
		return config.IndexFallback
	}

	return nil
}

func getIndexFallback(ctx *generate.GenerateContext) bool {
	// TODO we probably want to add a ENV var for this config option in the future
	if indexFallback := IndexFallbackFromStaticfile(ctx); indexFallback != nil {
		return *indexFallback
	}

	return false
}

// convert a Staticfile in the app source into a struct that we can read options from
func getStaticfileConfig(ctx *generate.GenerateContext) (*StaticfileConfig, error) {
	if !ctx.App.HasFile(StaticfileConfigName) {
		return nil, nil
	}

	staticfileData := StaticfileConfig{}
	if err := ctx.App.ReadYAML(StaticfileConfigName, &staticfileData); err != nil {
		return nil, err
	}

	return &staticfileData, nil
}
