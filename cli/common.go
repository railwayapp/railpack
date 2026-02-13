package cli

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/railwayapp/railpack/core"
	a "github.com/railwayapp/railpack/core/app"
	"github.com/railwayapp/railpack/core/config"
	"github.com/railwayapp/railpack/core/plan"
	"github.com/railwayapp/railpack/internal/utils"
	"github.com/urfave/cli/v3"
)

var Version string // This will be set by main

func commonPlanFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringSliceFlag{
			Name:    "env",
			Aliases: []string{"e"},
			Usage:   "environment variables to set",
		},
		&cli.StringSliceFlag{
			Name:  "secret",
			Usage: "secrets to expose (can lookup from environment)",
		},
		&cli.StringSliceFlag{
			Name:  "previous",
			Usage: "versions of packages used for previous builds (e.g. 'package@version')",
		},
		&cli.StringFlag{
			Name:  "build-cmd",
			Usage: "build command to use",
		},
		&cli.StringFlag{
			Name:  "start-cmd",
			Usage: "start command to use",
		},
		&cli.StringFlag{
			Name:  "config-file",
			Usage: "relative path to railpack config file (default: railpack.json)",
		},
		&cli.BoolFlag{
			Name:  "error-missing-start",
			Usage: "error if no start command is found",
		},
	}
}

func GenerateBuildResultForCommand(cmd *cli.Command) (*core.BuildResult, *a.App, *a.Environment, error) {
	directory := cmd.Args().First()

	if directory == "" {
		return nil, nil, nil, cli.Exit("directory argument is required", 1)
	}

	app, err := a.NewApp(directory)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error creating app: %w", err)
	}

	log.Debugf("Building %s", app.Source)

	envsArgs := cmd.StringSlice("env")
	secretsArgs := cmd.StringSlice("secret")

	env, err := a.FromEnvs(secretsArgs)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error creating env: %w", err)
	}

	vars, err := a.ParseKeyValueStrings(envsArgs, false)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error parsing env vars: %w", err)
	}

	// if --verbose is passed as a CLI global argument, enable verbose mise logging so the user don't have to understand
	// the railpack build system deeply to get this debugging information.
	if cmd.Bool("verbose") && env.GetVariable("MISE_VERBOSE") == "" {
		env.SetVariable("MISE_VERBOSE", "1")
	}

	previousVersions := utils.ParsePackageWithVersion(cmd.StringSlice("previous"))

	generateOptions := &core.GenerateBuildPlanOptions{
		RailpackVersion:          Version,
		BuildCommand:             cmd.String("build-cmd"),
		StartCommand:             cmd.String("start-cmd"),
		PreviousVersions:         previousVersions,
		ConfigFilePath:           cmd.String("config-file"),
		ErrorMissingStartCommand: cmd.Bool("error-missing-start"),
		Vars:                     vars,
	}

	buildResult := core.GenerateBuildPlan(app, env, generateOptions)

	return buildResult, app, env, nil
}

// add $schema link to resulting map JSON for improved IDE experience when manually editing
func addSchemaToPlanMap(p *plan.BuildPlan) (map[string]any, error) {
	if p == nil {
		return map[string]any{"$schema": config.SchemaUrl}, nil
	}
	planBytes, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	var planMap map[string]any
	if err := json.Unmarshal(planBytes, &planMap); err != nil {
		return nil, err
	}
	planMap["$schema"] = config.SchemaUrl
	return planMap, nil
}
