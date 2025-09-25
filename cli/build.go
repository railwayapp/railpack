package cli

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"

	"github.com/railwayapp/railpack/buildkit"
	"github.com/railwayapp/railpack/core"
	"github.com/railwayapp/railpack/core/app"
	"github.com/railwayapp/railpack/core/plan"
	"github.com/urfave/cli/v3"
)

var BuildCommand = &cli.Command{
	Name:                  "build",
	Aliases:               []string{"b"},
	Usage:                 "build an image with BuildKit",
	ArgsUsage:             "DIRECTORY",
	EnableShellCompletion: true,
	Flags: append([]cli.Flag{
		&cli.StringFlag{
			Name:  "name",
			Usage: "name of the image to build",
		},
		&cli.StringFlag{
			Name:  "output",
			Usage: "output the final filesystem to a local directory",
		},
		&cli.StringFlag{
			Name:  "platform",
			Usage: "platform to build for (e.g. linux/amd64, linux/arm64)",
		},
		&cli.StringFlag{
			Name:  "progress",
			Usage: "buildkit progress output mode. Values: auto, plain, tty",
			Value: "auto",
		},
		&cli.BoolFlag{
			Name:  "show-plan",
			Usage: "Show the build plan before building. This is useful for development and debugging.",
			Value: false,
		},
		&cli.StringFlag{
			Name:  "cache-key",
			Usage: "Unique id to prefix to cache keys",
		},
		&cli.BoolFlag{
			Name:   "dump-llb",
			Hidden: true,
			Value:  false,
		},
	}, commonPlanFlags()...),
	Action: func(ctx context.Context, cmd *cli.Command) error {
		buildResult, app, env, err := GenerateBuildResultForCommand(cmd)
		if err != nil {
			return cli.Exit(err, 1)
		}

		core.PrettyPrintBuildResult(buildResult, core.PrintOptions{Version: Version})

		if !buildResult.Success {
			os.Exit(1)
			return nil
		}

		if cmd.Bool("show-plan") {
			planMap, err := addSchemaToPlanMap(buildResult.Plan)
			if err != nil {
				return cli.Exit(err, 1)
			}

			serializedPlan, err := json.MarshalIndent(planMap, "", "  ")
			if err != nil {
				return cli.Exit(err, 1)
			}
			fmt.Println(string(serializedPlan))
		}

		err = validateSecrets(buildResult.Plan, env)
		if err != nil {
			return cli.Exit(err, 1)
		}

		secretsHash := getSecretsHash(env)

		platformStr := cmd.String("platform")
		err = buildkit.BuildWithBuildkitClient(app.Source, buildResult.Plan, buildkit.BuildWithBuildkitClientOptions{
			ImageName:    cmd.String("name"),
			DumpLLB:      cmd.Bool("dump-llb"),
			OutputDir:    cmd.String("output"),
			ProgressMode: cmd.String("progress"),
			CacheKey:     cmd.String("cache-key"),
			SecretsHash:  secretsHash,
			Secrets:      env.Variables,
			Platform:     platformStr,
			GitHubToken:  os.Getenv("GITHUB_TOKEN"),
		})
		if err != nil {
			return cli.Exit(err, 1)
		}

		return nil
	},
}

func validateSecrets(plan *plan.BuildPlan, env *app.Environment) error {
	for _, secret := range plan.Secrets {
		if _, ok := env.Variables[secret]; !ok {
			return fmt.Errorf("missing environment variable: %s. Please set the envvar with --env %s=%s", secret, secret, "...")
		}
	}
	return nil
}

func getSecretsHash(env *app.Environment) string {
	secretsValue := ""
	for _, v := range env.Variables {
		secretsValue += v
	}
	hasher := sha256.New()
	hasher.Write([]byte(secretsValue))
	return fmt.Sprintf("%x", hasher.Sum(nil))
}
