package buildkit

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseCacheImports(t *testing.T) {
	t.Parallel()

	t.Run("empty opts", func(t *testing.T) {
		imports, err := parseCacheImports(nil)
		require.NoError(t, err)
		require.Empty(t, imports)
	})

	t.Run("cache-imports JSON", func(t *testing.T) {
		imports, err := parseCacheImports(map[string]string{
			keyCacheImports: `[{"Type":"registry","Attrs":{"ref":"host.docker.internal:7890/node-bun:cache"}}]`,
		})
		require.NoError(t, err)
		require.Len(t, imports, 1)
		require.Equal(t, "registry", imports[0].Type)
		require.Equal(t, "host.docker.internal:7890/node-bun:cache", imports[0].Attrs["ref"])
	})

	t.Run("multiple cache-imports", func(t *testing.T) {
		imports, err := parseCacheImports(map[string]string{
			keyCacheImports: `[{"Type":"registry","Attrs":{"ref":"a:cache"}},{"Type":"gha","Attrs":{"scope":"my-scope"}}]`,
		})
		require.NoError(t, err)
		require.Len(t, imports, 2)
		require.Equal(t, "registry", imports[0].Type)
		require.Equal(t, "a:cache", imports[0].Attrs["ref"])
		require.Equal(t, "gha", imports[1].Type)
		require.Equal(t, "my-scope", imports[1].Attrs["scope"])
	})

	t.Run("invalid cache-imports JSON", func(t *testing.T) {
		_, err := parseCacheImports(map[string]string{
			keyCacheImports: `not-json`,
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), keyCacheImports)
	})
}
