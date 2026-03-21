package testing

import (
	"encoding/json"
	"testing"

	"github.com/railwayapp/railpack/core/config"
	"github.com/railwayapp/railpack/core/generate"
	"github.com/stretchr/testify/require"
)

func SetConfigFromJSON(t *testing.T, ctx *generate.GenerateContext, configJSON string) {
	t.Helper()

	parsedConfig := config.EmptyConfig()
	require.NoError(t, json.Unmarshal([]byte(configJSON), parsedConfig))
	ctx.Config = config.Merge(ctx.Config, parsedConfig)
}

func ClearConfigVariable(ctx *generate.GenerateContext, variableName string) {
	delete(ctx.Env.Variables, ctx.Env.ConfigVariable(variableName))
}
