package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
		&cli.BoolFlag{
			Name:  "show-plan",
			Usage: "show the generated build plan",
		},
	}, commonPlanFlags()...),
	Action: func(ctx context.Context, cmd *cli.Command) error {
		buildResult, _, _, err := GenerateBuildResultForCommand(cmd)
		if err != nil {
			return cli.Exit(err, 1)
		}

		format := cmd.String("format")

		var rendered bytes.Buffer
		if format == "pretty" {
			rendered.WriteString(core.FormatBuildResult(buildResult, core.PrintOptions{
				Metadata: true,
				Version:  Version,
			}))

			if cmd.Bool("show-plan") {
				planMap, err := addSchemaToPlanMap(buildResult.Plan)
				if err != nil {
					return cli.Exit(err, 1)
				}
				serializedPlan, err := json.MarshalIndent(planMap, "", "  ")
				if err != nil {
					return cli.Exit(err, 1)
				}

				core.PrettyPrintSectionHeader(&rendered, "Generated railpack-plan.json")
				core.PrettyPrintJSON(&rendered, serializedPlan)
			} else {
				fmt.Fprintf(
					&rendered,
					"Use %s to view generated railpack-plan.json\n",
					core.FormatHighlight("--show-plan"),
				)
			}
		} else {
			serializedResult, err := json.MarshalIndent(buildResult, "", "  ")
			if err != nil {
				return cli.Exit(err, 1)
			}
			fmt.Fprintln(&rendered, string(serializedResult))
		}

		output := cmd.String("out")
		if output == "" {
			// Write to stdout if no output file specified
			_, _ = os.Stdout.Write(rendered.Bytes())
			return nil
		} else {
			if err := os.MkdirAll(filepath.Dir(output), 0755); err != nil {
				return cli.Exit(err, 1)
			}

			err = os.WriteFile(output, rendered.Bytes(), 0644)
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
