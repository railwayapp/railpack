package shell

import (
	"github.com/charmbracelet/log"
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
	return nil
}

func (p *ShellProvider) Plan(ctx *generate.GenerateContext) error {
	ctx.Deploy.StartCmd = "sh " + p.scriptName

	ctx.Logger.LogInfo("Using shell script: %s", p.scriptName)

	setup := ctx.NewCommandStep("setup")
	setup.AddInput(ctx.MiseStepBuilder.GetLayer())
	setup.AddCommands(
		[]plan.Command{
			plan.NewCopyCommand("."),
			plan.NewExecCommand("chmod +x " + p.scriptName),
		},
	)

	ctx.Deploy.AddInputs([]plan.Layer{
		plan.NewStepLayer(setup.Name()),
	})

	return nil
}

func (p *ShellProvider) StartCommandHelp() string {
	return ""
}

func getScript(ctx *generate.GenerateContext) string {
	if scriptName, envVarName := ctx.Env.GetConfigVariable("SHELL_SCRIPT"); scriptName != "" {
		hasScript := ctx.App.HasMatch(scriptName)
		if hasScript {
			return scriptName
		}

		log.Warnf("%s %s script not found", envVarName, scriptName)
	}

	hasScript := ctx.App.HasMatch(StartScriptName)
	if hasScript {
		return StartScriptName
	}

	return ""
}
