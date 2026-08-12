package mise

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDownloadArchiveClassifiesStatusCodes(t *testing.T) {
	tests := []struct {
		status      int
		isTemporary bool
	}{
		{status: http.StatusOK},
		{status: http.StatusNotFound},
		{status: http.StatusForbidden},
		{status: http.StatusRequestTimeout, isTemporary: true},
		{status: http.StatusTooManyRequests, isTemporary: true},
		{status: http.StatusInternalServerError, isTemporary: true},
		{status: http.StatusServiceUnavailable, isTemporary: true},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("status-%d", tt.status), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
			}))
			defer server.Close()

			err := downloadArchive(server.URL, filepath.Join(t.TempDir(), "mise.tar.gz"))

			if tt.status == http.StatusOK {
				require.NoError(t, err)
				return
			}

			require.Error(t, err)
			require.Equal(t, tt.isTemporary, IsTemporary(err))
		})
	}
}

func TestDownloadArchiveWritesTheResponseBody(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		_, _ = w.Write([]byte("archive contents"))
	}))
	defer server.Close()

	archivePath := filepath.Join(t.TempDir(), "mise.tar.gz")
	require.NoError(t, downloadArchive(server.URL, archivePath))

	// the download is attempted once; retries are the caller's decision
	require.EqualValues(t, 1, requests.Load())

	contents, err := os.ReadFile(archivePath)
	require.NoError(t, err)
	require.Equal(t, "archive contents", string(contents))
}

func TestDownloadArchiveTransportFailureIsTemporary(t *testing.T) {
	// closing the server up front guarantees a connection error rather than a response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	url := server.URL
	server.Close()

	err := downloadArchive(url, filepath.Join(t.TempDir(), "mise.tar.gz"))
	require.True(t, IsTemporary(err))
}

// the install path wraps the download error several times before a caller sees it
func TestEnsureInstalledKeepsTemporaryFailuresDetectable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	originalBase := githubReleaseBase
	githubReleaseBase = server.URL
	t.Cleanup(func() { githubReleaseBase = originalBase })

	_, err := ensureInstalled(filepath.Join(t.TempDir(), "mise"))
	require.Error(t, err)
	require.True(t, IsTemporary(err))
}

func TestIsTemporaryUnwrapsWrappedErrors(t *testing.T) {
	temporary := &TemporaryError{URL: "https://example.com", Err: fmt.Errorf("i/o timeout")}

	require.True(t, IsTemporary(fmt.Errorf("failed to ensure mise is installed: %w", temporary)))
	require.False(t, IsTemporary(fmt.Errorf("binary not found in archive")))
	require.False(t, IsTemporary(nil))
}
