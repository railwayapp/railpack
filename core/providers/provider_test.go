package providers

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetProviderIsCaseInsensitive(t *testing.T) {
	require.NotNil(t, GetProvider("elixir"))
	require.NotNil(t, GetProvider("Elixir"))
	require.NotNil(t, GetProvider("dotnet"))
	require.NotNil(t, GetProvider("Dotnet"))
}
