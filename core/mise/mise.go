package mise

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/alexflint/go-filemutex"
	"github.com/charmbracelet/log"
	"github.com/railwayapp/railpack/internal/utils"
)

const (
	InstallDir     = "/tmp/railpack/mise"
	TestInstallDir = "/tmp/railpack/mise-test"
)

type Mise struct {
	binaryPath  string
	cacheDir    string
	githubToken string
}

const (
	ErrMiseGetLatestVersion = "failed to resolve version %s of %s"
)

func New(cacheDir string) (*Mise, error) {
	binaryPath, err := ensureInstalled(cacheDir)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure mise is installed: %w", err)
	}

	githubToken := os.Getenv("GITHUB_TOKEN")

	return &Mise{
		binaryPath:  binaryPath,
		cacheDir:    cacheDir,
		githubToken: githubToken,
	}, nil
}

// GetLatestVersion gets the latest version of a package matching the version constraint
func (m *Mise) GetLatestVersion(pkg, version string) (string, error) {
	_, unlock, err := m.createAndLock(pkg)
	if err != nil {
		return "", err
	}
	defer unlock()

	query := fmt.Sprintf("%s@%s", pkg, utils.ExtractSemverVersion(version))
	output, err := m.runCmd("latest", query)
	if err != nil {
		if strings.Contains(err.Error(), "not found in mise tool registry") {
			return "", fmt.Errorf("package `%s` not available in Mise. Try installing as apt package instead", pkg)
		}

		return "", err
	}

	latestVersion := strings.TrimSpace(output)
	if latestVersion == "" {
		return "", fmt.Errorf(ErrMiseGetLatestVersion, version, pkg)
	}

	return latestVersion, nil
}

func (m *Mise) GetAllVersions(pkg, version string) ([]string, error) {
	_, unlock, err := m.createAndLock(pkg)
	if err != nil {
		return nil, err
	}
	defer unlock()

	query := fmt.Sprintf("%s@%s", pkg, utils.ExtractSemverVersion(version))
	output, err := m.runCmd("ls-remote", query)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	var versions []string
	for _, line := range lines {
		version := strings.TrimSpace(line)
		if version == "" || strings.Contains(version, "RC") {
			continue
		}
		versions = append(versions, version)
	}

	if len(versions) == 0 {
		return nil, fmt.Errorf(ErrMiseGetLatestVersion, version, pkg)
	}

	return versions, nil
}

// runCmd runs a mise command with the given arguments
func (m *Mise) runCmd(args ...string) (string, error) {
	cacheDir := filepath.Join(m.cacheDir, "cache")
	dataDir := filepath.Join(m.cacheDir, "data")

	cmd := exec.Command(m.binaryPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	cmd.Env = append(cmd.Env,
		fmt.Sprintf("MISE_CACHE_DIR=%s", cacheDir),
		fmt.Sprintf("MISE_DATA_DIR=%s", dataDir),
		"MISE_HTTP_TIMEOUT=120s",
		"MISE_FETCH_REMOTE_VERSIONS_TIMEOUT=120s",
		fmt.Sprintf("PATH=%s", os.Getenv("PATH")),
	)

	if m.githubToken != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("GITHUB_TOKEN=%s", m.githubToken))
	}

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to run mise command '%s': %w\n%s\n\n%s",
			strings.Join(append([]string{m.binaryPath}, args...), " "),
			err,
			stdout.String(),
			stderr.String())
	}

	return stdout.String(), nil
}

// MisePackage represents a single mise package configuration
type MisePackage struct {
	Version string `toml:"version"`
}

// MiseConfig represents the overall mise configuration
type MiseConfig struct {
	Tools    map[string]MisePackage `toml:"tools"`
	Settings map[string]interface{} `toml:"settings,omitempty"`
}

func GenerateMiseToml(packages map[string]string) (string, error) {
	config := MiseConfig{
		Tools: make(map[string]MisePackage),
		Settings: map[string]interface{}{
			"experimental":        true,
			"auto_install":        false,
			"legacy_version_file": false,
		},
	}

	for name, version := range packages {
		config.Tools[name] = MisePackage{Version: version}
	}

	buf := bytes.NewBuffer(nil)
	if err := toml.NewEncoder(buf).Encode(config); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// createAndLock creates a file mutex and locks it, returning the mutex and an unlock function
func (m *Mise) createAndLock(pkg string) (*filemutex.FileMutex, func(), error) {
	fileLockPath := filepath.Join(m.cacheDir, fmt.Sprintf("lock-%s", strings.ReplaceAll(pkg, "/", "-")))
	mu, err := filemutex.New(fileLockPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create mutex: %w", err)
	}

	if err := mu.Lock(); err != nil {
		return nil, nil, fmt.Errorf("failed to acquire lock: %w", err)
	}

	unlock := func() {
		if err := mu.Unlock(); err != nil {
			log.Printf("failed to release lock: %v", err)
		}
	}

	return mu, unlock, nil
}
