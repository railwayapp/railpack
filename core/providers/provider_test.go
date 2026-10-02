package providers

import (
	"testing"

	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/plan"
	"github.com/stretchr/testify/require"
)

// stubProvider is a minimal Provider used to verify custom provider injection.
type stubProvider struct{ name string }

func (p *stubProvider) Name() string                                   { return p.name }
func (p *stubProvider) Detect(*generate.GenerateContext) (bool, error) { return false, nil }
func (p *stubProvider) Initialize(*generate.GenerateContext) error     { return nil }
func (p *stubProvider) Plan(*generate.GenerateContext) error           { return nil }
func (p *stubProvider) CleansePlan(*plan.BuildPlan)                    {}
func (p *stubProvider) StartCommandHelp() string                       { return "" }

func TestGetProviderFrom(t *testing.T) {
	custom := []Provider{&stubProvider{name: "custom"}}
	require.NotNil(t, GetProviderFrom("custom", custom))
	require.NotNil(t, GetProviderFrom("Custom", custom))
	require.Nil(t, GetProviderFrom("node", custom))

	builtin := GetLanguageProviders()
	require.NotNil(t, GetProviderFrom("elixir", builtin))
	require.NotNil(t, GetProviderFrom("Elixir", builtin))
	require.NotNil(t, GetProviderFrom("dotnet", builtin))
}
