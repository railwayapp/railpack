package node

import (
	"encoding/json"
	"path"
	"strings"

	"github.com/railwayapp/railpack/core/generate"
)

const (
	DefaultAngularStartCommand = "ng serve"
)

func (p *NodeProvider) isAngular(ctx *generate.GenerateContext) bool {
	hasAngularDep := p.hasDependency("@angular/core")
	hasAngularConfig := ctx.App.HasFile("angular.json")
	hasAngularBuildCommand := strings.Contains(strings.ToLower(p.packageJson.GetScript("build")), "ng build")

	return hasAngularDep && hasAngularConfig && hasAngularBuildCommand
}

type AngularConfig struct {
	Projects map[string]struct {
		Architect struct {
			Build struct {
				Builder string `json:"builder"`
				Options struct {
					OutputPath string `json:"outputPath"`
					Browser    string `json:"browser,omitempty"`
				} `json:"options"`
			} `json:"build"`
		} `json:"architect"`
	} `json:"projects"`
}

func (p *NodeProvider) getAngularOutputDirectory(ctx *generate.GenerateContext) string {
	angularJson, err := ctx.App.ReadFile("angular.json")
	if err != nil {
		return ""
	}

	var config AngularConfig
	if err := json.Unmarshal([]byte(angularJson), &config); err != nil {
		return ""
	}

	var projectName string
	if name, _ := ctx.Env.GetConfigVariable("ANGULAR_PROJECT"); name != "" {
		projectName = name
	} else {
		providerConfig := providerConfig(ctx)
		if providerConfig != nil {
			projectName = providerConfig.AngularProject
		}
	}

	for name, project := range config.Projects {
		if projectName != "" && projectName != name {
			continue
		}

		outputPath := project.Architect.Build.Options.OutputPath

		// Check if it's using the new application builder or has a browser field
		if project.Architect.Build.Builder == "@angular-devkit/build-angular:application" ||
			project.Architect.Build.Options.Browser != "" {
			return path.Join(outputPath, "browser")
		}

		return outputPath
	}

	return ""
}
