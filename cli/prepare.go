package cli

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
	"github.com/railwayapp/railpack/core"
	"github.com/urfave/cli/v3"
)

var PrepareCommand = &cli.Command{
	Name:                  "prepare",
	Aliases:               []string{"p"},
	Usage:                 "prepares all the files necessary for a platform to build an app with the BuildKit frontend",
	ArgsUsage:             "DIRECTORY",
	EnableShellCompletion: true,
	Flags: append([]cli.Flag{
		&cli.StringFlag{
			Name:  "plan-out",
			Usage: "output file for the JSON serialized build plan",
		},
		&cli.StringFlag{
			Name:  "info-out",
			Usage: "output file for the JSON serialized build result info",
		},
		&cli.BoolFlag{
			Name:  "show-plan",
			Usage: "dump the build plan to stdout",
		},
		&cli.BoolFlag{
			Name:  "hide-pretty-plan",
			Usage: "hide the pretty-printed build result output",
		},
	}, commonPlanFlags()...),
	Action: func(ctx context.Context, cmd *cli.Command) error {
		buildResult, _, _, err := GenerateBuildResultForCommand(cmd)
		if err != nil {
			return cli.Exit(err, 1)
		}

		// Pretty print the result to stdout unless hidden
		if !cmd.Bool("hide-pretty-plan") {
			core.PrettyPrintBuildResult(buildResult, core.PrintOptions{Version: Version})
		}

		if !buildResult.Success {
			os.Exit(1)
			return nil
		}

		// Show plan to stdout if requested
		if cmd.Bool("show-plan") {
			planMap, err := addSchemaToPlanMap(buildResult.Plan)
			if err != nil {
				return cli.Exit(err, 1)
			}
			serialized, err := json.MarshalIndent(planMap, "", "  ")
			if err != nil {
				return cli.Exit(err, 1)
			}
			os.Stdout.Write(serialized)
		}

		// Save plan if requested
		if planOut := cmd.String("plan-out"); planOut != "" {
			// Include $schema in the plan JSON for editor support
			planMap, err := addSchemaToPlanMap(buildResult.Plan)
			if err != nil {
				return cli.Exit(err, 1)
			}
			if err := writeJSONFile(planOut, planMap, "Build plan written to %s"); err != nil {
				return cli.Exit(err, 1)
			}
		}

		// Save info if requested
		if infoOut := cmd.String("info-out"); infoOut != "" {
			buildResult.Plan = nil
			if err := writeJSONFile(infoOut, buildResult, "Build result info written to %s"); err != nil {
				return cli.Exit(err, 1)
			}
		}

		return nil
	},
}

func writeJSONFile(path string, data interface{}, logMessage string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	serialized, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(path, serialized, 0644); err != nil {
		return err
	}

	log.Debugf(logMessage, path)
	return nil
}
