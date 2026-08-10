package core

import (
	"bytes"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
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

func TestPrettyPrintSectionHeader(t *testing.T) {
	var output bytes.Buffer
	PrettyPrintSectionHeader(&output, "Generated railpack-plan.json")

	require.Contains(t, output.String(), "Generated railpack-plan.json")
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

func TestPrettyPrintJSON(t *testing.T) {
	originalProfile := lipgloss.ColorProfile()
	t.Cleanup(func() {
		lipgloss.SetColorProfile(originalProfile)
	})

	json := []byte(`{"key":"value","enabled":true}`)

	lipgloss.SetColorProfile(termenv.ANSI)
	var colored bytes.Buffer
	PrettyPrintJSON(&colored, json)
	require.Contains(t, colored.String(), "\x1b[")

	lipgloss.SetColorProfile(termenv.Ascii)
	var plain bytes.Buffer
	PrettyPrintJSON(&plain, json)
	require.Equal(t, string(json)+"\n", plain.String())
}
