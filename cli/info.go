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

var InfoCommand = &cli.Command{
	Name:                  "info",
	Aliases:               []string{"i"},
	Usage:                 "get as much information as possible about an app",
	ArgsUsage:             "DIRECTORY",
	EnableShellCompletion: true,
	Flags: append([]cli.Flag{
		&cli.StringFlag{
			Name:  "format",
			Usage: "output format. one of: pretty, json",
			Value: "pretty",
		},
		&cli.StringFlag{
			Name:  "out",
			Usage: "output file name",
		},
		&cli.StringFlag{
			Name:    "workspace",
			Aliases: []string{"w"},
			Usage:   "specify a workspace directory within a monorepo",
		},
	}, commonPlanFlags()...),
	Action: func(ctx context.Context, cmd *cli.Command) error {
		buildResult, _, _, err := GenerateBuildResultForCommand(cmd)
		if err != nil {
			return cli.Exit(err, 1)
		}

		format := cmd.String("format")

		var buildResultString string
		if format == "pretty" {
			buildResultString = core.FormatBuildResult(buildResult, core.PrintOptions{
				Metadata: true,
				Version:  Version,
			})
		} else {
			serializedResult, err := json.MarshalIndent(buildResult, "", "  ")
			if err != nil {
				return cli.Exit(err, 1)
			}
			buildResultString = string(serializedResult)
		}

		output := cmd.String("out")
		if output == "" {
			// Write to stdout if no output file specified
			os.Stdout.Write([]byte(buildResultString))
			os.Stdout.Write([]byte("\n"))
			return nil
		} else {
			if err := os.MkdirAll(filepath.Dir(output), 0755); err != nil {
				return cli.Exit(err, 1)
			}

			err = os.WriteFile(output, []byte(buildResultString), 0644)
			if err != nil {
				return cli.Exit(err, 1)
			}

			log.Infof("Plan written to %s", output)
		}

		if !buildResult.Success {
			os.Exit(1)
			return nil
		}

		return nil
	},
}
