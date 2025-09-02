package cli

import (
	"encoding/json"
	"testing"

	"github.com/railwayapp/railpack/core/config"
	"github.com/railwayapp/railpack/core/plan"
	"github.com/stretchr/testify/require"
)

func TestAddSchemaToPlanMap_IncludesSchemaAndPreservesPlan(t *testing.T) {
	// Create a minimal non-empty plan
	p := plan.NewBuildPlan()
	p.Deploy.StartCmd = "run"
	p.Secrets = []string{"FOO"}

	// Marshal the original plan to a map for comparison
	baseBytes, err := json.Marshal(p)
	require.NoError(t, err)
	var baseMap map[string]any
	require.NoError(t, json.Unmarshal(baseBytes, &baseMap))

	// Get the map with schema and validate
	outMap, err := addSchemaToPlanMap(p)
	require.NoError(t, err)
	require.Equal(t, config.SchemaUrl, outMap["$schema"])

	// Remove the added $schema and compare with the original
	delete(outMap, "$schema")
	require.Equal(t, baseMap, outMap)
}

func TestAddSchemaToPlanMap_NilPlan(t *testing.T) {
	outMap, err := addSchemaToPlanMap(nil)
	require.NoError(t, err)
	require.Len(t, outMap, 1)
	require.Equal(t, config.SchemaUrl, outMap["$schema"])
}
