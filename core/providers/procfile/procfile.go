// this provider is unique: it used solely to extract a start command
package procfile

import "github.com/railwayapp/railpack/core/generate"

type ProcfileProvider struct{}

func (p *ProcfileProvider) Name() string {
	return "procfile"
}

func (p *ProcfileProvider) Plan(ctx *generate.GenerateContext) (bool, error) {
	if _, err := ctx.App.ReadFile("Procfile"); err != nil {
		return false, nil
	}

	parsedProcfile := map[string]string{}
	if err := ctx.App.ReadYAML("Procfile", &parsedProcfile); err != nil {
		return false, err
	}

	webCommand := parsedProcfile["web"]
	workerCommand := parsedProcfile["worker"]

	if webCommand != "" {
		ctx.Logger.LogInfo("Found web command in Procfile")
		ctx.Deploy.StartCmd = webCommand
	} else if workerCommand != "" {
		ctx.Logger.LogInfo("Found worker command in Procfile")
		ctx.Deploy.StartCmd = workerCommand
	} else if len(parsedProcfile) > 0 {
		for processType, command := range parsedProcfile {
			if command != "" {
				ctx.Logger.LogInfo("Found %s command in Procfile", processType)
				ctx.Deploy.StartCmd = command
				break
			}
		}
	}

	return false, nil
}
