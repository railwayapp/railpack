package mise

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
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

func TestPinnedChecksumsCoverEverySupportedPlatform(t *testing.T) {
	platforms := [][2]string{
		{"linux", "amd64"}, {"linux", "arm64"}, {"linux", "arm"},
		{"darwin", "amd64"}, {"darwin", "arm64"},
		{"windows", "amd64"}, {"windows", "arm64"},
	}

	for _, p := range platforms {
		archive, err := getAssetName(p[0], p[1])
		require.NoError(t, err)
		binary, err := getBinaryAssetName(p[0], p[1])
		require.NoError(t, err)

		// fails when version.txt is bumped without refreshing SHASUMS256.txt
		for _, asset := range []string{archive, binary} {
			checksum, err := pinnedChecksum(asset)
			require.NoError(t, err)
			require.Regexp(t, "^[0-9a-f]{64}$", checksum)
		}
	}
}

func TestEnsureInstalledRejectsArchiveNotMatchingPinnedChecksum(t *testing.T) {
	serveRelease(t, []byte("not the pinned archive"))

	cacheDir := t.TempDir()
	_, err := ensureInstalled(cacheDir)
	require.ErrorContains(t, err, "SHA-256 mismatch")
	requireEmptyDir(t, cacheDir)
}

func TestEnsureInstalledRejectsExtractedBinaryNotMatchingPinnedChecksum(t *testing.T) {
	archive := buildReleaseArchive(t, []byte("tampered binary"))
	pinChecksums(t, archive, []byte("pinned binary"))
	serveRelease(t, archive)

	cacheDir := t.TempDir()
	_, err := ensureInstalled(cacheDir)
	require.ErrorContains(t, err, "SHA-256 mismatch")
	requireEmptyDir(t, cacheDir)
}

func TestEnsureInstalledInstallsBinaryMatchingPinnedChecksum(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the fixture binary is a shell script")
	}

	// validateInstallation runs `<binary> --version` and expects Version in the output
	binary := fmt.Appendf(nil, "#!/bin/sh\necho %s\n", Version)
	archive := buildReleaseArchive(t, binary)
	pinChecksums(t, archive, binary)
	serveRelease(t, archive)

	binaryPath, err := ensureInstalled(t.TempDir())
	require.NoError(t, err)

	info, err := os.Stat(binaryPath)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0755), info.Mode().Perm())
}

func TestEnsureInstalledRefusesExistingBinaryNotMatchingPinnedChecksum(t *testing.T) {
	pinChecksums(t, nil, []byte("pinned binary"))
	serveRelease(t, nil)

	cacheDir := t.TempDir()
	require.NoError(t, os.WriteFile(getBinaryPath(cacheDir), []byte("planted binary"), 0755))

	_, err := ensureInstalled(cacheDir)
	require.ErrorContains(t, err, "refusing to run existing mise executable")
}

func TestEnsureInstalledReusesExistingBinaryMatchingPinnedChecksum(t *testing.T) {
	binary := []byte("pinned binary")
	pinChecksums(t, nil, binary)
	// any download attempt would fail against a closed server
	server := httptest.NewServer(http.NotFoundHandler())
	server.Close()
	setReleaseBase(t, server.URL)

	cacheDir := t.TempDir()
	require.NoError(t, os.WriteFile(getBinaryPath(cacheDir), binary, 0755))

	binaryPath, err := ensureInstalled(cacheDir)
	require.NoError(t, err)
	require.Equal(t, getBinaryPath(cacheDir), binaryPath)
}

func TestResolveInstallDirPrefersExplicitDirThenEnvironmentThenDefault(t *testing.T) {
	baseDir := t.TempDir()
	envDir := filepath.Join(baseDir, "from-env")
	explicitDir := filepath.Join(baseDir, "explicit")

	t.Setenv(InstallDirEnv, "")
	dir, err := resolveInstallDir("")
	require.NoError(t, err)
	require.Equal(t, InstallDir, dir)

	t.Setenv(InstallDirEnv, envDir)
	dir, err = resolveInstallDir("")
	require.NoError(t, err)
	require.Equal(t, envDir, dir)

	dir, err = resolveInstallDir(explicitDir)
	require.NoError(t, err)
	require.Equal(t, explicitDir, dir)
}

// builds a release archive in the format getAssetName names for this platform
func buildReleaseArchive(t *testing.T, binary []byte) []byte {
	t.Helper()
	var buf bytes.Buffer

	if runtime.GOOS == "windows" {
		zw := zip.NewWriter(&buf)
		w, err := zw.Create("mise/bin/mise.exe")
		require.NoError(t, err)
		_, err = w.Write(binary)
		require.NoError(t, err)
		require.NoError(t, zw.Close())
		return buf.Bytes()
	}

	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)
	require.NoError(t, tw.WriteHeader(&tar.Header{Name: "mise/bin/mise", Mode: 0755, Size: int64(len(binary)), Typeflag: tar.TypeReg}))
	_, err := tw.Write(binary)
	require.NoError(t, err)
	require.NoError(t, tw.Close())
	require.NoError(t, gw.Close())
	return buf.Bytes()
}

// replaces the embedded SHASUMS256.txt with digests for this platform's fixture assets
func pinChecksums(t *testing.T, archive, binary []byte) {
	t.Helper()
	archiveName, err := getAssetName(runtime.GOOS, runtime.GOARCH)
	require.NoError(t, err)
	binaryName, err := getBinaryAssetName(runtime.GOOS, runtime.GOARCH)
	require.NoError(t, err)

	original := pinnedChecksums
	pinnedChecksums = fmt.Sprintf("%x  ./%s\n%x  ./%s\n", sha256.Sum256(archive), archiveName, sha256.Sum256(binary), binaryName)
	t.Cleanup(func() { pinnedChecksums = original })
}

// serves archive as every release asset
func serveRelease(t *testing.T, archive []byte) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(archive)
	}))
	t.Cleanup(server.Close)
	setReleaseBase(t, server.URL)
}

func setReleaseBase(t *testing.T, url string) {
	t.Helper()
	original := githubReleaseBase
	githubReleaseBase = url
	t.Cleanup(func() { githubReleaseBase = original })
}

func requireEmptyDir(t *testing.T, dir string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	require.Empty(t, entries)
}

func TestIsTemporaryUnwrapsWrappedErrors(t *testing.T) {
	temporary := &TemporaryError{URL: "https://example.com", Err: fmt.Errorf("i/o timeout")}

	require.True(t, IsTemporary(fmt.Errorf("failed to ensure mise is installed: %w", temporary)))
	require.False(t, IsTemporary(fmt.Errorf("binary not found in archive")))
	require.False(t, IsTemporary(nil))
}
