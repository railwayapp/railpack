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
			return cli.Exit(err, exitCodeForError(err))
		}

		// Pretty print the result to stdout unless hidden
		if !cmd.Bool("hide-pretty-plan") {
			core.PrettyPrintBuildResult(buildResult, core.PrintOptions{Version: Version})
		}

		if !buildResult.Success {
			// still write the info file so callers can read the failure from it rather than stdout
			if err := writeInfoFile(cmd, buildResult); err != nil {
				return cli.Exit(err, ExitCodeFailure)
			}

			os.Exit(ExitCodeFailure)
			return nil
		}

		// Show plan to stdout if requested
		if cmd.Bool("show-plan") {
			planMap, err := addSchemaToPlanMap(buildResult.Plan)
			if err != nil {
				return cli.Exit(err, ExitCodeFailure)
			}
			serialized, err := json.MarshalIndent(planMap, "", "  ")
			if err != nil {
				return cli.Exit(err, ExitCodeFailure)
			}
			_, _ = os.Stdout.Write(serialized)
		}

		// Save plan if requested
		if planOut := cmd.String("plan-out"); planOut != "" {
			// Include $schema in the plan JSON for editor support
			planMap, err := addSchemaToPlanMap(buildResult.Plan)
			if err != nil {
				return cli.Exit(err, ExitCodeFailure)
			}
			if err := writeJSONFile(planOut, planMap, "Build plan written to %s"); err != nil {
				return cli.Exit(err, ExitCodeFailure)
			}
		}

		if err := writeInfoFile(cmd, buildResult); err != nil {
			return cli.Exit(err, ExitCodeFailure)
		}

		return nil
	},
}

// writes the build result to the --info-out file, if one was requested. The plan is
// dropped from the copy since it is written separately with --plan-out.
func writeInfoFile(cmd *cli.Command, buildResult *core.BuildResult) error {
	infoOut := cmd.String("info-out")
	if infoOut == "" {
		return nil
	}

	info := *buildResult
	info.Plan = nil

	return writeJSONFile(infoOut, &info, "Build result info written to %s")
}

func writeJSONFile(path string, data any, logMessage string) error {
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
