package mise

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/log"
)

//go:embed version.txt
var Version string

// The release's SHASUMS256.txt, copied verbatim whenever version.txt changes (see
// .mise/tasks/mise-upgrade). Pinning the digests in the binary means a replaced release
// asset is rejected rather than trusted. A var so tests can pin fixture archives.
//
//go:embed SHASUMS256.txt
var pinnedChecksums string

// a var so tests can point the download at a local server
var githubReleaseBase = "https://github.com/jdx/mise/releases/download"

// http.DefaultClient has no timeout of any kind, so a stalled connection would
// hang the build indefinitely. The ceiling is generous to tolerate slow builders.
var downloadClient = &http.Client{Timeout: 5 * time.Minute}

// returns name of the mise binary based on the operating system
func getBinaryName() string {
	if runtime.GOOS == "windows" {
		return fmt.Sprintf("mise-%s.exe", Version)
	}
	return fmt.Sprintf("mise-%s", Version)
}

// returns the platform suffix mise uses in its release asset names
func getPlatform(goos, goarch string) (string, error) {
	switch {
	case goos == "linux" && goarch == "amd64":
		return "linux-x64-musl", nil
	case goos == "linux" && goarch == "arm64":
		return "linux-arm64-musl", nil
	case goos == "linux" && goarch == "arm":
		return "linux-armv7-musl", nil
	case goos == "darwin" && goarch == "amd64":
		return "macos-x64", nil
	case goos == "darwin" && goarch == "arm64":
		return "macos-arm64", nil
	case goos == "windows" && goarch == "amd64":
		return "windows-x64", nil
	case goos == "windows" && goarch == "arm64":
		return "windows-arm64", nil
	default:
		return "", fmt.Errorf("unsupported platform: %s %s", goos, goarch)
	}
}

// returns platform-specific mise github asset download name
func getAssetName(goos, goarch string) (string, error) {
	platform, err := getPlatform(goos, goarch)
	if err != nil {
		return "", err
	}

	extension := "tar.gz"
	if goos == "windows" {
		extension = "zip"
	}

	return fmt.Sprintf("mise-v%s-%s.%s", Version, platform, extension), nil
}

// returns the name of the standalone binary release asset. It is byte-identical to the
// binary inside the archive, so its pinned digest also verifies an extracted binary.
func getBinaryAssetName(goos, goarch string) (string, error) {
	platform, err := getPlatform(goos, goarch)
	if err != nil {
		return "", err
	}

	if goos == "windows" {
		return fmt.Sprintf("mise-v%s-%s.exe", Version, platform), nil
	}
	return fmt.Sprintf("mise-v%s-%s", Version, platform), nil
}

// looks up the pinned SHA-256 digest of a release asset
func pinnedChecksum(assetName string) (string, error) {
	for line := range strings.Lines(pinnedChecksums) {
		// sha256sum format: "<digest>  ./<name>", or "<digest> *<name>" in binary mode
		fields := strings.Fields(line)
		if len(fields) == 2 && strings.TrimPrefix(strings.TrimPrefix(fields[1], "*"), "./") == assetName {
			return strings.ToLower(fields[0]), nil
		}
	}

	return "", fmt.Errorf("no pinned SHA-256 checksum for %s: core/mise/SHASUMS256.txt must match mise %s", assetName, Version)
}

func verifyChecksum(path, expected string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	hash := sha256.New()
	if _, err := io.Copy(hash, f); err != nil {
		return fmt.Errorf("failed to hash %s: %w", path, err)
	}

	if actual := hex.EncodeToString(hash.Sum(nil)); actual != expected {
		return fmt.Errorf("SHA-256 mismatch for %s: expected %s, got %s", path, expected, actual)
	}

	return nil
}

// getBinaryPath returns the full path to the binary
func getBinaryPath(cacheDir string) string {
	return filepath.Join(cacheDir, getBinaryName())
}

// ensures the mise binary (at the pinned version and checksum) is installed and returns its path
func ensureInstalled(cacheDir string) (string, error) {
	binaryPath := getBinaryPath(cacheDir)

	binaryAssetName, err := getBinaryAssetName(runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return "", err
	}
	binaryChecksum, err := pinnedChecksum(binaryAssetName)
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(binaryPath); err == nil {
		// The default install dir is under /tmp, where any local user can create this file first
		if err := verifyChecksum(binaryPath, binaryChecksum); err != nil {
			return "", fmt.Errorf("refusing to run existing mise executable (delete it or set %s to a directory you control): %w", InstallDirEnv, err)
		}

		log.Debugf("Mise executable exists at %s", binaryPath)
		return binaryPath, nil
	}

	log.Debugf("Mise %s not found, installing", Version)

	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}

	if err := downloadAndInstall(cacheDir, binaryChecksum); err != nil {
		return "", fmt.Errorf("failed to download and install: %w", err)
	}

	if err := validateInstallation(cacheDir); err != nil {
		return "", fmt.Errorf("failed to validate installation: %w", err)
	}

	log.Debugf("Installed mise version: %s to %s", Version, binaryPath)

	return binaryPath, nil
}

func downloadAndInstall(cacheDir, binaryChecksum string) error {
	assetName, err := getAssetName(runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return err
	}
	archiveChecksum, err := pinnedChecksum(assetName)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/v%s/%s", githubReleaseBase, Version, assetName)
	binaryPath := getBinaryPath(cacheDir)

	log.Debugf("Downloading mise from %s", url)

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "mise-install")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	archivePath := filepath.Join(tempDir, assetName)
	if err := downloadArchive(url, archivePath); err != nil {
		return err
	}

	// checked before extraction so a tampered archive never reaches the decompressor
	if err := verifyChecksum(archivePath, archiveChecksum); err != nil {
		return err
	}

	if runtime.GOOS == "windows" {
		err = extractZip(archivePath, binaryPath, binaryChecksum)
	} else {
		err = extractTarGz(archivePath, binaryPath, binaryChecksum)
	}
	if err != nil {
		return fmt.Errorf("failed to extract archive: %w", err)
	}

	if runtime.GOOS != "windows" {
		if err := os.Chmod(binaryPath, 0755); err != nil {
			return fmt.Errorf("failed to set executable permissions: %w", err)
		}
	}

	return nil
}

// downloads url to archivePath. Transient failures are typed so the caller can
// decide whether to retry; this deliberately does not retry on its own.
func downloadArchive(url, archivePath string) error {
	resp, err := downloadClient.Get(url)
	if err != nil {
		// transport failures (dial timeouts, resets) are never the app's fault
		return &TemporaryError{URL: url, Err: err}
	}
	defer func() { _ = resp.Body.Close() }()

	if err := checkDownloadStatus(url, resp); err != nil {
		return err
	}

	f, err := os.Create(archivePath)
	if err != nil {
		return fmt.Errorf("failed to create archive file: %w", err)
	}
	defer func() { _ = f.Close() }()

	if _, err := io.Copy(f, resp.Body); err != nil {
		// the connection dropped mid-body, leaving a truncated archive
		return &TemporaryError{URL: url, Err: fmt.Errorf("failed to save archive: %w", err)}
	}

	return nil
}

// classifies the response status. Rate limits and server errors are transient;
// anything else (notably a 404 for a mise version that does not exist) is not.
func checkDownloadStatus(url string, resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	statusErr := fmt.Errorf("unexpected status %s", resp.Status)

	if resp.StatusCode == http.StatusRequestTimeout || resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return &TemporaryError{URL: url, Err: statusErr}
	}

	return fmt.Errorf("failed to download mise from %s: %w", url, statusErr)
}

func extractTarGz(archivePath, binaryPath, binaryChecksum string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer func() { _ = gzr.Close() }()

	tr := tar.NewReader(gzr)
	binaryPathInArchive := "mise/bin/mise"
	found := false

	writeAndMove, cleanup, err := createAtomicWriter(binaryPath, binaryChecksum)
	if err != nil {
		return err
	}
	defer cleanup()

	return writeAndMove(func(tempFile *os.File) error {
		for {
			header, err := tr.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}

			if header.Name == binaryPathInArchive {
				if _, err := io.Copy(tempFile, tr); err != nil {
					return err
				}
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("binary not found in archive at %s", binaryPathInArchive)
		}

		return nil
	})
}

func extractZip(archivePath, binaryPath, binaryChecksum string) error {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer func() { _ = r.Close() }()

	writeAndMove, cleanup, err := createAtomicWriter(binaryPath, binaryChecksum)
	if err != nil {
		return err
	}
	defer cleanup()

	// the Windows archive also ships mise/bin/mise-shim.exe, so match the full path
	binaryPathInArchive := "mise/bin/mise.exe"
	for _, f := range r.File {
		if f.Name == binaryPathInArchive {
			rc, err := f.Open()
			if err != nil {
				return err
			}

			err = writeAndMove(func(tempFile *os.File) error {
				_, err := io.Copy(tempFile, rc)
				_ = rc.Close()
				return err
			})

			return err
		}
	}

	return fmt.Errorf("binary not found in archive at %s", binaryPathInArchive)
}

func validateInstallation(cacheDir string) error {
	binaryPath := getBinaryPath(cacheDir)
	cmd := exec.Command(binaryPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to run version check: %w", err)
	}

	versionOutput := string(output)
	if !strings.Contains(versionOutput, Version) {
		return fmt.Errorf("mise version mismatch: expected %s, got %s", Version, strings.TrimSpace(versionOutput))
	}

	return nil
}

// creates a temporary file and returns a function to atomically write content to the final
// destination. The content must match checksum before it is made executable and moved into place.
func createAtomicWriter(targetPath, checksum string) (writeAndMove func(write func(tempFile *os.File) error) error, cleanup func(), err error) {
	tempFile, err := os.CreateTemp(filepath.Dir(targetPath), "mise-temp-*")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	tempPath := tempFile.Name()

	success := false
	cleanup = func() {
		_ = tempFile.Close()
		if !success {
			_ = os.Remove(tempPath)
		}
	}

	writeAndMove = func(write func(tempFile *os.File) error) error {
		if err := write(tempFile); err != nil {
			return err
		}

		if err := tempFile.Close(); err != nil {
			return fmt.Errorf("failed to close temp file: %w", err)
		}

		if err := verifyChecksum(tempPath, checksum); err != nil {
			return err
		}

		if runtime.GOOS != "windows" {
			if err := os.Chmod(tempPath, 0755); err != nil {
				return fmt.Errorf("failed to set executable permissions: %w", err)
			}
		}

		if err := os.Rename(tempPath, targetPath); err != nil {
			return fmt.Errorf("failed to move temp file to target: %w", err)
		}

		success = true
		return nil
	}

	return writeAndMove, cleanup, nil
}
