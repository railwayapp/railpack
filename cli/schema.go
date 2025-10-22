package cli

import (
	"context"
	"encoding/json"
	"os"

	"github.com/railwayapp/railpack/core/config"
	"github.com/urfave/cli/v3"
)

var SchemaCommand = &cli.Command{
	Name:                  "schema",
	Usage:                 "outputs the JSON schema for the Railpack config",
	EnableShellCompletion: true,
	Flags:                 []cli.Flag{},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		schema := config.GetJsonSchema()

		schemaJson, err := json.MarshalIndent(schema, "", "  ")
		if err != nil {
			return cli.Exit(err, 1)
		}

		if _, err := os.Stdout.Write(schemaJson); err != nil {
			return cli.Exit(err, 1)
		}
		if _, err := os.Stdout.Write([]byte("\n")); err != nil {
			return cli.Exit(err, 1)
		}

		return nil
	},
}
