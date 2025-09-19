// this provider is distinct from the SPA functionality used by node providers
// it is meant to simply serve static files over HTTP

package staticfile

import (
	_ "embed"
	"fmt"

	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/plan"
)

//go:embed Caddyfile.template
var caddyfileTemplate string

const (
	StaticfileConfigName = "Staticfile"
	CaddyfilePath        = "Caddyfile"
)

type StaticfileConfig struct {
	RootDir string `yaml:"root"`
}

type StaticfileProvider struct {
	RootDir string
}

func (p *StaticfileProvider) Name() string {
	return "staticfile"
}

func (p *StaticfileProvider) Initialize(ctx *generate.GenerateContext) error {
	rootDir, err := getRootDir(ctx)
	if err != nil {
		return err
	}

	p.RootDir = rootDir

	return nil
}

func (p *StaticfileProvider) Detect(ctx *generate.GenerateContext) (bool, error) {
	rootDir, err := getRootDir(ctx)
	if rootDir != "" && err == nil {
		return true, nil
	}

	return false, nil
}

func (p *StaticfileProvider) Plan(ctx *generate.GenerateContext) error {
	miseStep := ctx.GetMiseStepBuilder()
	miseStep.Default("caddy", "latest")

	build := ctx.NewCommandStep("build")
	build.AddInput(plan.NewStepLayer(miseStep.Name()))
	build.AddInput(ctx.NewLocalLayer())

	err := p.addCaddyfileToStep(ctx, build)
	if err != nil {
		return err
	}

	ctx.Deploy.AddInputs([]plan.Layer{
		miseStep.GetLayer(),
		plan.NewStepLayer(build.Name(), plan.Filter{
			Include: []string{"."},
		}),
	})

	ctx.Deploy.StartCmd = fmt.Sprintf("caddy run --config %s --adapter caddyfile 2>&1", CaddyfilePath)

	return nil
}

func (p *StaticfileProvider) StartCommandHelp() string {
	return ""
}

func (p *StaticfileProvider) addCaddyfileToStep(ctx *generate.GenerateContext, setup *generate.CommandStepBuilder) error {
	ctx.Logger.LogInfo("Using root dir: %s", p.RootDir)

	data := map[string]interface{}{
		"STATIC_FILE_ROOT": p.RootDir,
	}

	caddyfileTemplate, err := ctx.TemplateFiles([]string{"Caddyfile.template", "Caddyfile"}, caddyfileTemplate, data)
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

func getRootDir(ctx *generate.GenerateContext) (string, error) {
	if rootDir, _ := ctx.Env.GetConfigVariable("STATIC_FILE_ROOT"); rootDir != "" {
		return rootDir, nil
	}

	staticfileConfig, err := getStaticfileConfig(ctx)
	if staticfileConfig != nil && err == nil {
		return staticfileConfig.RootDir, nil
	}

	if ctx.App.HasMatch("public") {
		return "public", nil
	} else if ctx.App.HasMatch("index.html") {
		return ".", nil
	}

	return "", fmt.Errorf("no static file root dir found")
}

func getStaticfileConfig(ctx *generate.GenerateContext) (*StaticfileConfig, error) {
	if !ctx.App.HasMatch(StaticfileConfigName) {
		return nil, nil
	}

	staticfileData := StaticfileConfig{}
	if err := ctx.App.ReadYAML(StaticfileConfigName, &staticfileData); err != nil {
		return nil, err
	}

	return &staticfileData, nil
}
