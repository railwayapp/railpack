package shell

import (
	"errors"

	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/plan"
)

const (
	StartScriptName = "start.sh"
)

type ShellProvider struct {
	scriptName string
}

func (p *ShellProvider) Name() string {
	return "shell"
}

func (p *ShellProvider) Detect(ctx *generate.GenerateContext) (bool, error) {
	return getScript(ctx) != "", nil
}

func (p *ShellProvider) Initialize(ctx *generate.GenerateContext) error {
	p.scriptName = getScript(ctx)

	if p.scriptName == "" {
		return errors.New("start shell script could not be found")
	}

	return nil
}

func (p *ShellProvider) Plan(ctx *generate.GenerateContext) error {
	ctx.Deploy.StartCmd = "sh " + p.scriptName

	ctx.Logger.LogInfo("Using shell script: %s", p.scriptName)

	build := ctx.NewCommandStep("build")
	build.AddInput(plan.NewImageLayer(plan.RailpackRuntimeImage))
	build.AddInput(plan.NewLocalLayer("."))
	build.AddCommands(
		[]plan.Command{
			plan.NewExecCommand("chmod +x " + p.scriptName),
		},
	)

	ctx.Deploy.AddInputs([]plan.Layer{
		plan.NewStepLayer(build.Name(), plan.Filter{
			Include: []string{"."},
		}),
	})

	return nil
}

func (p *ShellProvider) StartCommandHelp() string {
	return ""
}

// determine shell script to use for container start
func getScript(ctx *generate.GenerateContext) string {
	scriptName, envVarName := ctx.Env.GetConfigVariable("SHELL_SCRIPT")
	if scriptName == "" {
		scriptName = StartScriptName
	}

	if ctx.App.HasFile(scriptName) {
		return scriptName
	}

	if envVarName != "" {
		ctx.Logger.LogWarn("%s %s script not found", envVarName, scriptName)
	} else {
		ctx.Logger.LogWarn("script %s not found", scriptName)
	}

	return ""
}
