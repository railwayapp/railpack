package cli

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
	"github.com/urfave/cli/v3"
)

var PlanCommand = &cli.Command{
	Name:                  "plan",
	Aliases:               []string{"p"},
	Usage:                 "generate a build plan for a directory",
	ArgsUsage:             "DIRECTORY",
	EnableShellCompletion: true,
	Flags: append([]cli.Flag{
		&cli.StringFlag{
			Name:    "out",
			Aliases: []string{"o"},
			Usage:   "output file name",
		},
	}, commonPlanFlags()...),
	Action: func(ctx context.Context, cmd *cli.Command) error {
		buildResult, _, _, err := GenerateBuildResultForCommand(cmd)
		if err != nil {
			return cli.Exit(err, exitCodeForError(err))
		}

		// Include $schema in the generated plan JSON for editor support
		planMap, err := addSchemaToPlanMap(buildResult.Plan)
		if err != nil {
			return cli.Exit(err, ExitCodeFailure)
		}

		// serialize the plan to a string to dump to the plan file
		buildResultString, err := json.MarshalIndent(planMap, "", "  ")
		if err != nil {
			return cli.Exit(err, ExitCodeFailure)
		}

		output := cmd.String("out")
		if output == "" {
			// Write to stdout if no output file specified
			_, _ = os.Stdout.Write([]byte(buildResultString))
			_, _ = os.Stdout.Write([]byte("\n"))
			return nil
		} else {
			if err := os.MkdirAll(filepath.Dir(output), 0755); err != nil {
				return cli.Exit(err, ExitCodeFailure)
			}

			err = os.WriteFile(output, []byte(buildResultString), 0644)
			if err != nil {
				return cli.Exit(err, ExitCodeFailure)
			}

			log.Infof("Plan written to %s", output)
		}

		return nil
	},
}
