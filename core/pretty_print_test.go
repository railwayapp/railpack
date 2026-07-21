package core

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPrettyPrintStyles(t *testing.T) {
	highlighted := FormatHighlight("docker run -it shell-script")
	output := boxStyle.Render("Successfully built image in 1.08s\n\nRun:\n" + highlighted)

	require.Contains(t, output, "Successfully built image in 1.08s")
	require.Contains(t, output, "docker run -it shell-script")
	require.NotContains(t, output, "✓")
}
