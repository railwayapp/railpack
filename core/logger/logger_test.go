package logger

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLogSuggestion(t *testing.T) {
	t.Run("without docs path", func(t *testing.T) {
		l := NewLogger()
		l.LogSuggestion("try including `...`")

		require.Len(t, l.Logs, 1)
		require.Equal(t, Suggestion, l.Logs[0].Level)
		require.Equal(t, "try including `...`", l.Logs[0].Msg)
		require.Empty(t, l.Logs[0].DocsPath)
	})

	t.Run("with docs path", func(t *testing.T) {
		l := NewLogger()
		l.LogSuggestion("try including `...`", "/guides/installing-packages")

		require.Len(t, l.Logs, 1)
		require.Equal(t, Suggestion, l.Logs[0].Level)
		require.Equal(t, "try including `...`", l.Logs[0].Msg)
		require.Equal(t, "/guides/installing-packages", l.Logs[0].DocsPath)
	})
}

func TestDocsURL(t *testing.T) {
	require.Equal(t, "https://railpack.com", DocsURL(""))
	require.Equal(t, "https://railpack.com/guides/installing-packages", DocsURL("/guides/installing-packages"))
	require.Equal(t, "https://railpack.com/guides/installing-packages", DocsURL("guides/installing-packages"))
}
