package mise

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	_ "embed"
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

// returns platform-specific mise github asset download name
func getAssetName(goos, goarch string) (string, error) {
	var platform string

	switch {
	case goos == "linux" && goarch == "amd64":
		platform = "linux-x64-musl"
	case goos == "linux" && goarch == "arm64":
		platform = "linux-arm64-musl"
	case goos == "linux" && goarch == "arm":
		platform = "linux-armv7-musl"
	case goos == "darwin" && goarch == "amd64":
		platform = "macos-x64"
	case goos == "darwin" && goarch == "arm64":
		platform = "macos-arm64"
	case goos == "windows" && goarch == "amd64":
		platform = "windows-x64"
	case goos == "windows" && goarch == "arm64":
		platform = "windows-arm64"
	default:
		return "", fmt.Errorf("unsupported platform: %s %s", goos, goarch)
	}

	extension := "tar.gz"
	if goos == "windows" {
		extension = "zip"
	}

	return fmt.Sprintf("mise-v%s-%s.%s", Version, platform, extension), nil
}

// getBinaryPath returns the full path to the binary
func getBinaryPath(cacheDir string) string {
	return filepath.Join(cacheDir, getBinaryName())
}

// ensures the mise binary (at the pinned version) is installed and returns its path
func ensureInstalled(cacheDir string) (string, error) {
	binaryPath := getBinaryPath(cacheDir)

	if _, err := os.Stat(binaryPath); err == nil {
		log.Debugf("Mise executable exists at %s", binaryPath)
		return binaryPath, nil
	}

	log.Debugf("Mise %s not found, installing", Version)

	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}

	if err := downloadAndInstall(cacheDir); err != nil {
		return "", fmt.Errorf("failed to download and install: %w", err)
	}

	if err := validateInstallation(cacheDir); err != nil {
		return "", fmt.Errorf("failed to validate installation: %w", err)
	}

	log.Debugf("Installed mise version: %s to %s", Version, binaryPath)

	return binaryPath, nil
}

func downloadAndInstall(cacheDir string) error {
	assetName, err := getAssetName(runtime.GOOS, runtime.GOARCH)
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

	if runtime.GOOS == "windows" {
		err = extractZip(archivePath, binaryPath)
	} else {
		err = extractTarGz(archivePath, binaryPath)
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

func extractTarGz(archivePath, binaryPath string) error {
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

	writeAndMove, cleanup, err := createAtomicWriter(binaryPath)
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

func extractZip(archivePath, binaryPath string) error {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer func() { _ = r.Close() }()

	writeAndMove, cleanup, err := createAtomicWriter(binaryPath)
	if err != nil {
		return err
	}
	defer cleanup()

	binaryName := getBinaryName()
	for _, f := range r.File {
		if strings.HasSuffix(f.Name, binaryName) {
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

	return fmt.Errorf("binary not found in archive")
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

// creates a temporary file and returns a function to atomically write content to the final destination
func createAtomicWriter(targetPath string) (writeAndMove func(write func(tempFile *os.File) error) error, cleanup func(), err error) {
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
