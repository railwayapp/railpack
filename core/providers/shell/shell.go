package shell

import (
	"errors"

	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/plan"
	"mvdan.cc/sh/v3/fileutil"
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
	interpreter, err := detectShellInterpreter(ctx, p.scriptName)
	if err != nil {
		return err
	}

	ctx.Deploy.StartCmd = interpreter + " " + p.scriptName
	ctx.Metadata.Set("detectedShellInterpreter", interpreter)

	// zsh is not included in the base image by default, but it's common enough that we should support it
	if interpreter == "zsh" {
		ctx.Logger.LogInfo("Installing zsh for shell script execution")
		ctx.Deploy.AddAptPackages([]string{"zsh"})
	}

	ctx.Logger.LogInfo("Using shell script: %s with interpreter: %s", p.scriptName, interpreter)

	build := ctx.NewCommandStep("build")
	build.AddInput(plan.NewImageLayer(plan.RailpackRuntimeImage))
	build.AddInput(ctx.NewLocalLayer())
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

func detectShellInterpreter(ctx *generate.GenerateContext, scriptName string) (string, error) {
	content, err := ctx.App.ReadFile(scriptName)
	if err != nil {
		return "", err
	}

	// fileutil.Shebang only recognizes POSIX-compliant shells (bash, sh, zsh, etc).
	// Non-POSIX shells like fish are not detected and will return empty string.
	// TODO in the future, we should add config for forcing a specific shell interpreter.
	interpreter := fileutil.Shebang([]byte(content))
	if interpreter == "" {
		return "sh", nil
	}

	return mapToAvailableShell(ctx, interpreter), nil
}

func mapToAvailableShell(ctx *generate.GenerateContext, shell string) string {
	switch shell {
	case "bash":
		return "bash"
	case "zsh":
		return "zsh"
	case "sh", "dash":
		return "sh"
	case "mksh", "ksh", "fish":
		ctx.Logger.LogWarn("Shell '%s' not available in runtime, using 'bash'", shell)
		return "bash"
	default:
		ctx.Logger.LogWarn("Unknown shell '%s', using 'sh'", shell)
		return "sh"
	}
}
