package main

import (
	"context"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/muesli/termenv"
	"github.com/railwayapp/railpack/cli"
	urfave "github.com/urfave/cli/v3"
)

var (
	verbose bool
	version = "dev" // This will be overwritten by goreleaser
)

func main() {
	cli.Version = version

	logger := log.Default()
	logger.SetTimeFormat("")
	urfaveLogWriter := logger.StandardLog(log.StandardLogOptions{
		ForceLevel: log.ErrorLevel,
	}).Writer()
	urfave.ErrWriter = urfaveLogWriter

	if os.Getenv("FORCE_COLOR") != "" {
		lipgloss.SetColorProfile(termenv.TrueColor)
	}

	// https://no-color.org
	if os.Getenv("NO_COLOR") != "" {
		lipgloss.SetColorProfile(termenv.Ascii)
	}

	commands := []*urfave.Command{
		cli.BuildCommand,
		cli.PrepareCommand,
		cli.InfoCommand,
		cli.PlanCommand,
		cli.SchemaCommand,
		cli.FrontendCommand,
	}

	// Disable the slice flag separator for all commands
	for _, command := range commands {
		command.DisableSliceFlagSeparator = true
	}

	cmd := &urfave.Command{
		Name:                  "railpack",
		Usage:                 "Automatically analyze and generate build plans for applications",
		EnableShellCompletion: true,
		Version:               cli.Version,
		Flags: []urfave.Flag{
			&urfave.BoolFlag{
				Name:        "verbose",
				Usage:       "Enable verbose logging",
				Value:       false,
				Destination: &verbose,
			},
		},
		Before: func(ctx context.Context, cmd *urfave.Command) (context.Context, error) {
			configureLogging(verbose)

			return ctx, nil
		},
		Commands: commands,
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

func configureLogging(verbose bool) {
	log.SetTimeFormat("")

	if verbose {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
}
