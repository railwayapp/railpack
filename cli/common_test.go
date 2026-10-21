package cli

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/railwayapp/railpack/core/config"
	"github.com/railwayapp/railpack/core/mise"
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

func TestExitCodeForError(t *testing.T) {
	temporary := &mise.TemporaryError{URL: "https://example.com/mise.tar.gz", Err: fmt.Errorf("i/o timeout")}

	tests := []struct {
		name     string
		err      error
		expected int
	}{
		{name: "temporary", err: temporary, expected: ExitCodeTransient},
		{name: "wrapped temporary", err: fmt.Errorf("failed to ensure mise is installed: %w", temporary), expected: ExitCodeTransient},
		{name: "deterministic", err: fmt.Errorf("no start command was found"), expected: ExitCodeFailure},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, exitCodeForError(tt.err))
		})
	}
}

func TestAddSchemaToPlanMap_NilPlan(t *testing.T) {
	outMap, err := addSchemaToPlanMap(nil)
	require.NoError(t, err)
	require.Len(t, outMap, 1)
	require.Equal(t, config.SchemaUrl, outMap["$schema"])
}
