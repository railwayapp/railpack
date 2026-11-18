package cli

import (
	"testing"

	"github.com/railwayapp/railpack/core/app"
	"github.com/stretchr/testify/require"
)

func TestGetSecretsHashIsStable(t *testing.T) {
	env := app.NewEnvironment(&map[string]string{"A": "1", "B": "2", "C": "3", "D": "4", "E": "5"})
	want := getSecretsHash(env)
	for range 20 {
		require.Equal(t, want, getSecretsHash(env))
	}
}
