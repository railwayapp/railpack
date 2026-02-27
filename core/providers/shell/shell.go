package shell

import (
	"errors"

	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/plan"
	shellconfig "github.com/railwayapp/railpack/core/providers/shell/config"
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

	miseStepName := ctx.GetMiseStepBuilder().Name()
	var buildBaseLayer plan.Layer

	// If install step is configured (e.g. via RAILPACK_INSTALL_CMD) we add it so the user-supplied install config is run properly
	if _, ok := ctx.Config.Steps["install"]; ok {
		install := ctx.NewCommandStep("install")
		// Install step needs mise base to access any tools installed in the mise step (e.g. via RAILPACK_PACKAGES)
		install.AddInput(plan.NewStepLayer(miseStepName))
		install.AddInput(ctx.NewLocalLayer())

		// If we have an install step, the build step should be based on the result of the install step
		// so that artifacts from the install step are available during the build.
		buildBaseLayer = plan.NewStepLayer(install.Name())
	} else {
		// If no install step is configured, the build step is based directly on the mise step.
		buildBaseLayer = plan.NewStepLayer(miseStepName)
	}

	build := ctx.NewCommandStep("build")
	build.AddInput(buildBaseLayer)
	build.AddInput(ctx.NewLocalLayer())
	build.AddCommands(
		[]plan.Command{
			plan.NewExecCommand("chmod +x " + p.scriptName),
		},
	)

	ctx.Deploy.AddInputs([]plan.Layer{
		ctx.GetMiseStepBuilder().GetLayer(),
		plan.NewStepLayer(build.Name(), plan.Filter{
			Include: []string{"."},
		}),
	})

	return nil
}

func (p *ShellProvider) CleansePlan(buildPlan *plan.BuildPlan) {}

func (p *ShellProvider) StartCommandHelp() string {
	return ""
}

func providerConfig(ctx *generate.GenerateContext) *shellconfig.ShellConfig {
	if ctx.Config == nil {
		return nil
	}

	return ctx.Config.Shell
}

func shellScript(ctx *generate.GenerateContext) (string, string) {
	if scriptName, envVarName := ctx.Env.GetConfigVariable("SHELL_SCRIPT"); scriptName != "" {
		return scriptName, envVarName
	}

	providerConfig := providerConfig(ctx)
	if providerConfig != nil && providerConfig.Script != "" {
		return providerConfig.Script, "shell.script"
	}

	return "", ""
}

// determine shell script to use for container start
func getScript(ctx *generate.GenerateContext) string {
	scriptName, source := shellScript(ctx)
	if scriptName == "" {
		scriptName = StartScriptName
	}

	if ctx.App.HasFile(scriptName) {
		return scriptName
	}

	if source != "" {
		ctx.Logger.LogWarn("%s %s script not found", source, scriptName)
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
