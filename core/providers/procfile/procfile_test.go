package procfile

import (
	"os"
	"path/filepath"
	"testing"

	testingUtils "github.com/railwayapp/railpack/core/testing"
	"github.com/stretchr/testify/require"
)

func TestProcfile(t *testing.T) {
	ctx := testingUtils.CreateGenerateContext(t, "../../../examples/ruby-vanilla")
	provider := ProcfileProvider{}

	_, err := provider.Plan(ctx)
	require.NoError(t, err)

	require.Equal(t, "ruby app.rb", ctx.Deploy.StartCmd)
}

func planProcfile(t *testing.T, contents string) (string, error) {
	t.Helper()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Procfile"), []byte(contents), 0o644))

	ctx := testingUtils.CreateGenerateContext(t, dir)
	_, err := (&ProcfileProvider{}).Plan(ctx)
	return ctx.Deploy.StartCmd, err
}

func TestProcfileCommandWithColonAndSpace(t *testing.T) {
	startCmd, err := planProcfile(t, "web: python -c \"print('listening: 8080')\"\n")
	require.NoError(t, err)
	require.Equal(t, "python -c \"print('listening: 8080')\"", startCmd)
}

func TestProcfileCommandStartingWithQuote(t *testing.T) {
	startCmd, err := planProcfile(t, "web: \"./bin/server\" --port $PORT\n")
	require.NoError(t, err)
	require.Equal(t, "\"./bin/server\" --port $PORT", startCmd)
}

func TestProcfileIgnoresCommentsAndBlankLines(t *testing.T) {
	startCmd, err := planProcfile(t, "# processes\n\nweb: node server.js\n")
	require.NoError(t, err)
	require.Equal(t, "node server.js", startCmd)
}

func TestProcfileFallbackUsesFirstProcess(t *testing.T) {
	// map iteration order is random, so repeat to catch a nondeterministic pick
	for range 20 {
		startCmd, err := planProcfile(t, "clock: ./clock.sh\nqueue: ./queue.sh\nurgent: ./urgent.sh\n")
		require.NoError(t, err)
		require.Equal(t, "./clock.sh", startCmd)
	}
}
