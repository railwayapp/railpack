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
	ArgsUsage:             "DIRECTORY_OR_GITHUB_URL",
	Description:           "Scan a local directory or GitHub repository to detect language, frameworks, and generate build information.\nFor GitHub repos, use format: github.com/owner/repo or https://github.com/owner/repo",
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
				log.Infof("Scanned GitHub repo: %s (%s)", app.RemoteURL, authStatus)
			}
		}()

		format := cmd.String("format")

		var buildResultString string
		if format == "pretty" {
			opts := core.PrintOptions{
				Metadata: true,
				Version:  Version,
			}
			if app != nil && app.IsRemote {
				opts.GitHubURL = app.RemoteURL
				opts.GitHubAuth = app.GitHubClient != nil && app.GitHubClient.Token != ""
			}
			buildResultString = core.FormatBuildResult(buildResult, opts)
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
