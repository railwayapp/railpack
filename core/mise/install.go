package mise

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/charmbracelet/log"
)

const (
	miseVersion       = "2025.8.13"
	githubReleaseBase = "https://github.com/jdx/mise/releases/download"
)

// getBinaryName returns the name of the binary based on the operating system
func getBinaryName() string {
	if runtime.GOOS == "windows" {
		return fmt.Sprintf("mise-%s.exe", miseVersion)
	}
	return fmt.Sprintf("mise-%s", miseVersion)
}

// getAssetName returns the platform-specific asset name
func getAssetName() (string, error) {
	var platform string

	switch {
	case runtime.GOOS == "linux" && runtime.GOARCH == "amd64":
		platform = "linux-x64"
	case runtime.GOOS == "linux" && runtime.GOARCH == "arm64":
		platform = "linux-arm64"
	case runtime.GOOS == "linux" && runtime.GOARCH == "arm":
		platform = "linux-armv7"
	case runtime.GOOS == "darwin" && runtime.GOARCH == "amd64":
		platform = "macos-x64"
	case runtime.GOOS == "darwin" && runtime.GOARCH == "arm64":
		platform = "macos-arm64"
	case runtime.GOOS == "windows" && runtime.GOARCH == "amd64":
		platform = "windows-x64"
	case runtime.GOOS == "windows" && runtime.GOARCH == "arm64":
		platform = "windows-arm64"
	default:
		return "", fmt.Errorf("unsupported platform: %s %s", runtime.GOOS, runtime.GOARCH)
	}

	extension := "tar.gz"
	if runtime.GOOS == "windows" {
		extension = "zip"
	}

	return fmt.Sprintf("mise-v%s-%s.%s", miseVersion, platform, extension), nil
}

// getBinaryPath returns the full path to the binary
func getBinaryPath(cacheDir string) string {
	return filepath.Join(cacheDir, getBinaryName())
}

// ensureInstalled ensures the mise binary is installed and returns its path
func ensureInstalled(cacheDir string) (string, error) {
	binaryPath := getBinaryPath(cacheDir)

	if _, err := os.Stat(binaryPath); err == nil {
		log.Debugf("Mise executable exists at %s", binaryPath)
		return binaryPath, nil
	}

	log.Debugf("Mise %s not found, installing", miseVersion)

	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}

	if err := downloadAndInstall(cacheDir); err != nil {
		return "", fmt.Errorf("failed to download and install: %w", err)
	}

	if err := validateInstallation(cacheDir); err != nil {
		return "", fmt.Errorf("failed to validate installation: %w", err)
	}

	log.Debugf("Installed mise version: %s to %s", miseVersion, binaryPath)

	return binaryPath, nil
}

func downloadAndInstall(cacheDir string) error {
	assetName, err := getAssetName()
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/v%s/%s", githubReleaseBase, miseVersion, assetName)
	binaryPath := getBinaryPath(cacheDir)

	log.Debugf("Downloading mise from %s", url)

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download mise: %w", err)
	}
	defer resp.Body.Close()

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "mise-install")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	archivePath := filepath.Join(tempDir, assetName)
	f, err := os.Create(archivePath)
	if err != nil {
		return fmt.Errorf("failed to create archive file: %w", err)
	}

	if _, err := io.Copy(f, resp.Body); err != nil {
		f.Close()
		return fmt.Errorf("failed to save archive: %w", err)
	}
	f.Close()

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

func extractTarGz(archivePath, binaryPath string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzr.Close()

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
	defer r.Close()

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
				rc.Close()
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
	if !strings.Contains(versionOutput, miseVersion) {
		return fmt.Errorf("mise version mismatch: expected %s, got %s", miseVersion, strings.TrimSpace(versionOutput))
	}

	return nil
}

// createAtomicWriter creates a temporary file and returns a function to atomically write content to the final destination
func createAtomicWriter(targetPath string) (writeAndMove func(write func(tempFile *os.File) error) error, cleanup func(), err error) {
	tempFile, err := os.CreateTemp(filepath.Dir(targetPath), "mise-temp-*")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	tempPath := tempFile.Name()

	success := false
	cleanup = func() {
		tempFile.Close()
		if !success {
			os.Remove(tempPath)
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
