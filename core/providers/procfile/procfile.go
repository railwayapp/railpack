// this provider is unique: it used solely to extract a start command
package procfile

import (
	"github.com/railwayapp/railpack/core/generate"
)

type ProcfileProvider struct{}

func (p *ProcfileProvider) Name() string {
	return "procfile"
}

func (p *ProcfileProvider) Plan(ctx *generate.GenerateContext) (bool, error) {
	contents, err := ctx.App.ReadFile("Procfile")
	if err != nil {
		return false, nil
	}

	entries, err := ParseProcfile(contents)
	if err != nil {
		return false, err
	}
	if len(entries) == 0 {
		return false, nil
	}

	webCommand := ""
	workerCommand := ""
	for _, entry := range entries {
		switch entry.Name {
		case "web":
			webCommand = entry.Command
		case "worker":
			workerCommand = entry.Command
		}
	}

	if webCommand != "" {
		ctx.Logger.LogInfo("Found web command in Procfile")
		ctx.Deploy.StartCmd = webCommand
		return false, nil
	}

	if workerCommand != "" {
		ctx.Logger.LogInfo("Found worker command in Procfile")
		ctx.Deploy.StartCmd = workerCommand
		return false, nil
	}

	// use the first process in file order so the result is deterministic
	ctx.Logger.LogInfo("Found %s command in Procfile", entries[0].Name)
	ctx.Deploy.StartCmd = entries[0].Command

	return false, nil
}
