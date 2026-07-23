package core

import (
	"testing"

	"github.com/railwayapp/railpack/core/logger"
	"github.com/stretchr/testify/require"
)

func TestPrettyPrintStyles(t *testing.T) {
	highlighted := FormatHighlight("docker run -it shell-script")
	output := boxStyle.Render("Successfully built image in 1.08s\n\nRun:\n" + highlighted)

	require.Contains(t, output, "Successfully built image in 1.08s")
	require.Contains(t, output, "docker run -it shell-script")
	require.NotContains(t, output, "✓")
}

func TestPrettyPrintDeprecationLog(t *testing.T) {
	buildResult := &BuildResult{
		Logs: []logger.Msg{
			{
				Level: logger.Deprecation,
				Msg:   "old behavior will change",
			},
		},
	}

	output := FormatBuildResult(buildResult)

	require.Contains(t, output, "⚑ Deprecated: Old behavior will change")
}

func TestPrettyPrintSuggestionLog(t *testing.T) {
	buildResult := &BuildResult{
		Logs: []logger.Msg{
			{
				Level:    logger.Suggestion,
				Msg:      "include `...` in `buildAptPackages`",
				DocsPath: "/guides/installing-packages",
			},
		},
	}

	output := FormatBuildResult(buildResult)

	require.Contains(t, output, "→ Include `...` in `buildAptPackages`")
	require.Contains(t, output, "https://railpack.com/guides/installing-packages")
}
