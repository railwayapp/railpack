// helper utilities to run the mise tool on the host
// this is distinct from the mise step builder which generates mise commands to be run inside the container
// for this reason, the commands here are heavily sandboxed from the host environment to avoid picking up host configs

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
	InstallDir                = "/tmp/railpack/mise"
	TestInstallDir            = "/tmp/railpack/mise-test"
	IdiomaticVersionFileTools = "python,node,ruby,elixir,go,java,yarn"
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

	// without the GITHUB_TOKEN, mise will 403 us
	githubToken := os.Getenv("GITHUB_TOKEN")

	return &Mise{
		binaryPath:  binaryPath,
		cacheDir:    cacheDir,
		githubToken: githubToken,
	}, nil
}

// gets the latest version of a package matching the version constraint
func (m *Mise) GetLatestVersion(pkg, version string) (string, error) {
	_, unlock, err := m.createAndLock(pkg)
	if err != nil {
		return "", err
	}
	defer unlock()

	semverVersion := utils.ExtractSemverVersion(version)
	query := fmt.Sprintf("%s@%s", pkg, semverVersion)

	// Try with extracted semver version first
	output, err := m.runCmdWithEnv([]string{"MISE_NO_CONFIG=1", "MISE_PARANOID=1"}, "latest", query)

	// If semver extraction fails, try with original version
	// https://github.com/railwayapp/railpack/issues/203
	if (err != nil || strings.TrimSpace(output) == "") && semverVersion != version {
		query = fmt.Sprintf("%s@%s", pkg, version)
		output, err = m.runCmdWithEnv([]string{"MISE_NO_CONFIG=1", "MISE_PARANOID=1"}, "latest", query)
	}

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

	// Try with extracted semver version first
	semverVersion := utils.ExtractSemverVersion(version)
	query := fmt.Sprintf("%s@%s", pkg, semverVersion)
	output, err := m.runCmdWithEnv([]string{"MISE_NO_CONFIG=1", "MISE_PARANOID=1"}, "ls-remote", query)

	// If semver extraction fails, try with original version
	// https://github.com/railwayapp/railpack/issues/203
	if (err != nil || strings.TrimSpace(output) == "") && semverVersion != version {
		query = fmt.Sprintf("%s@%s", pkg, version)
		output, err = m.runCmdWithEnv([]string{"MISE_NO_CONFIG=1", "MISE_PARANOID=1"}, "ls-remote", query)
	}

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

// returns the JSON output of 'mise list --current --json' for the app
func (m *Mise) GetCurrentList(appDir string) (string, error) {
	// MISE_TRUSTED_CONFIG_PATHS allows mise to use configs in the app directory without a trust warning
	trustedConfigEnv := fmt.Sprintf("MISE_TRUSTED_CONFIG_PATHS=%s", appDir)

	// MISE_CEILING_PATHS prevents mise from searching parent directories, isolating it to the app directory
	// This eliminates the risk of local configuration (when running on a dev machine, for instance) polluting the mise
	// configuration (and therefore packages) that are bundled into the image.

	// We set the ceiling to the parent dir so mise can still read configs in appDir itself
	// since MISE_CEILING_PATHS prevents reading the root mise.toml settings
	ceilingPathsEnv := fmt.Sprintf("MISE_CEILING_PATHS=%s", filepath.Dir(appDir))

	// eliminates the need to have custom .python-version, etc parsing logic for each provider
	enabledIdiomaticEnv := fmt.Sprintf("MISE_IDIOMATIC_VERSION_FILE_ENABLE_TOOLS=%s", IdiomaticVersionFileTools)

	return m.runCmdWithEnv([]string{
		trustedConfigEnv,
		ceilingPathsEnv,
		enabledIdiomaticEnv,
		// MISE_PARANOID enables stricter security validation
		"MISE_PARANOID=1",
	}, "--cd", appDir, "list", "--current", "--json")
}

// runCmdWithEnv runs a mise command with additional environment variables
func (m *Mise) runCmdWithEnv(extraEnv []string, args ...string) (string, error) {
	cacheDir := filepath.Join(m.cacheDir, "cache")
	dataDir := filepath.Join(m.cacheDir, "data")
	stateDir := filepath.Join(m.cacheDir, "state")
	systemDir := filepath.Join(m.cacheDir, "system")

	cmd := exec.Command(m.binaryPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// https://github.com/jdx/mise/blob/main/src/dirs.rs
	// MISE_SYSTEM_DIR ensures any local config on the host does not interfere with mise commands
	cmd.Env = append(cmd.Env,
		fmt.Sprintf("HOME=%s", m.cacheDir),
		fmt.Sprintf("MISE_CACHE_DIR=%s", cacheDir),
		fmt.Sprintf("MISE_DATA_DIR=%s", dataDir),
		fmt.Sprintf("MISE_STATE_DIR=%s", stateDir),
		fmt.Sprintf("MISE_SYSTEM_DIR=%s", systemDir),
		// TODO doesn't HTTP timeout apply to fetch remote versions too?
		"MISE_HTTP_TIMEOUT=60s",
		"MISE_FETCH_REMOTE_VERSIONS_TIMEOUT=60s",
		// allows for a 2m outage on mise (10ms base backoff retry)
		"MISE_HTTP_RETRIES=5",
		fmt.Sprintf("PATH=%s", os.Getenv("PATH")),
	)

	if m.githubToken != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("GITHUB_TOKEN=%s", m.githubToken))
	}

	if len(extraEnv) > 0 {
		cmd.Env = append(cmd.Env, extraEnv...)
	}

	log.Debugf("Running mise command with env: %v", cmd.Env)

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
	Tools map[string]MisePackage `toml:"tools"`
}

// used by the container mise logic, but uses the package structs defined in this file
func GenerateMiseToml(packages map[string]string) (string, error) {
	config := MiseConfig{
		Tools: make(map[string]MisePackage),
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

// lock ensuring mise does not work on the same package concurrently
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
