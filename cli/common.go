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
			Name:  "env",
			Usage: "environment variables to set",
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
			Usage: "path to config file to use",
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

	env, err := a.FromEnvs(envsArgs)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error creating env: %w", err)
	}

	previousVersions := utils.ParsePackageWithVersion(cmd.StringSlice("previous"))

	generateOptions := &core.GenerateBuildPlanOptions{
		RailpackVersion:          Version,
		BuildCommand:             cmd.String("build-cmd"),
		StartCommand:             cmd.String("start-cmd"),
		PreviousVersions:         previousVersions,
		ConfigFilePath:           cmd.String("config-file"),
		ErrorMissingStartCommand: cmd.Bool("error-missing-start"),
	}

	buildResult := core.GenerateBuildPlan(app, env, generateOptions)

	return buildResult, app, env, nil
}

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
