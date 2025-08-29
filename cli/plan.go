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
	Usage:                 "generate a build plan for a directory or GitHub repository",
	ArgsUsage:             "DIRECTORY_OR_GITHUB_URL",
	Description:           "Generate a build plan for a local directory or GitHub repository.\nFor GitHub repos, use format: github.com/owner/repo or https://github.com/owner/repo",
	EnableShellCompletion: true,
	Flags: append([]cli.Flag{
		&cli.StringFlag{
			Name:    "out",
			Aliases: []string{"o"},
			Usage:   "output file name",
		},
	}, commonPlanFlags()...),
	Action: func(ctx context.Context, cmd *cli.Command) error {
		buildResult, app, _, err := GenerateBuildResultForCommand(cmd)
		if err != nil {
			return cli.Exit(err, 1)
		}
		defer func() {
			if app != nil && app.IsRemote {
				authStatus := "unauthenticated"
				if app.GitHubClient != nil && app.GitHubClient.Token != "" {
					authStatus = "authenticated"
				}
				log.Infof("Generated plan for GitHub repo: %s (%s)", app.RemoteURL, authStatus)
			}
		}()

		// Include $schema in the generated plan JSON for editor support
		planMap, err := addSchemaToPlanMap(buildResult.Plan)
		if err != nil {
			return cli.Exit(err, 1)
		}

		// Add success field and error information if available
		planMap["success"] = buildResult.Success

		// Extract error message from logs if build failed
		if !buildResult.Success && len(buildResult.Logs) > 0 {
			for _, log := range buildResult.Logs {
				if string(log.Level) == "error" {
					planMap["error"] = log.Msg
					break
				}
			}
		}

		serializedPlan, err := json.MarshalIndent(planMap, "", "  ")
		if err != nil {
			return cli.Exit(err, 1)
		}
		buildResultString := serializedPlan

		output := cmd.String("out")
		if output == "" {
			// Write to stdout if no output file specified
			os.Stdout.Write([]byte(buildResultString))
			os.Stdout.Write([]byte("\n"))
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
