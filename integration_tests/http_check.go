package integration_tests

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

const (
	// Grace period for servers to bind after initial log lines.
	containerStartGracePeriod = 2 * time.Second
	// Total timeout for the HTTP check to succeed. Instead of specifying the number of retries, we use a total timeout.
	httpCheckTimeout = 35 * time.Second
	// Time between HTTP check retries.
	httpCheckInterval = 300 * time.Millisecond
	// Timeout for a single HTTP GET request.
	httpCheckClientTimeout = 3 * time.Second
	// Network name suffix when docker-compose creates the network
	composeNetworkSuffix = "_default"
)

// ComposeConfig holds information about docker-compose services for an example
type ComposeConfig struct {
	ProjectName       string
	ExamplePath       string
	NetworkName       string
	DockerComposePath string
}

// detectAndStartCompose checks if a docker-compose.yml exists in the example directory
// and starts the services with a unique network name. Returns nil if no compose file exists or is empty.
func detectAndStartCompose(examplePath string, t interface {
	Logf(format string, args ...interface{})
}) (*ComposeConfig, error) {
	composeFile := filepath.Join(examplePath, "docker-compose.yml")

	// Check if docker-compose.yml exists
	fileInfo, err := os.Stat(composeFile)
	if os.IsNotExist(err) {
		return nil, nil
	}

	// Skip empty files (e.g., test files that verify dockerignore works)
	if fileInfo.Size() == 0 {
		return nil, nil
	}

	// Generate a unique project name and network name
	projectName := fmt.Sprintf("railpack-test-%s", strings.ToLower(uuid.New().String()))
	networkName := projectName + composeNetworkSuffix

	// Start docker-compose services and wait for them to be ready
	cmd := exec.Command("docker", "compose",
		"-f", composeFile,
		"--project-name", projectName,
		"up", "-d", "--wait")

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to start docker-compose services: %v: %s", err, string(out))
	}
	t.Logf("Started docker-compose services with network: %s", networkName)

	return &ComposeConfig{
		ProjectName:       projectName,
		ExamplePath:       examplePath,
		NetworkName:       networkName,
		DockerComposePath: composeFile,
	}, nil
}

// stopAndCleanupCompose stops and removes docker-compose services
func stopAndCleanupCompose(config *ComposeConfig, logger interface {
	Logf(format string, args ...interface{})
}) error {
	if config == nil {
		return nil
	}

	cmd := exec.Command("docker", "compose",
		"-f", config.DockerComposePath,
		"--project-name", config.ProjectName,
		"down", "--volumes")

	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to stop docker-compose services: %v: %s", err, string(out))
	}

	return nil
}

// HTTPCheck defines an HTTP endpoint check for a container
// Path defaults to /, Expected defaults to 200.
type HTTPCheck struct {
	InternalPort int    `json:"internalPort"`
	Path         string `json:"path"`
	Expected     int    `json:"expected"`
}

// runContainerWithHTTPCheck starts a container from the given image,
// maps the internalPort to a random host port, and checks the given path returns the expected status code.
// If networkName is provided, the container will be connected to that network.
func runContainerWithHTTPCheck(t *testing.T, imageName string, envs map[string]string, hc *HTTPCheck, networkName string) error {
	if networkName != "" {
		t.Logf("Using network: %s", networkName)
	}
	if hc.Path == "" {
		hc.Path = "/"
	}
	if hc.Expected == 0 {
		hc.Expected = 200
	}
	if hc.InternalPort == 0 {
		return fmt.Errorf("httpCheck.internalPort must be specified")
	}

	hostPort, err := pickFreePort()
	if err != nil {
		return fmt.Errorf("allocate host port: %w", err)
	}

	// at this point, the container has already been built, so GITHUB_TOKEN should not be needed
	containerName := uniqueContainerName(uuid.New().String())
	// Run detached without --rm so we can fetch logs even if it exits.
	// Cleanup is handled explicitly in the deferred stopContainer call below.
	args := []string{"run", "-d", "--name", containerName, "-p", fmt.Sprintf("127.0.0.1:%d:%d", hostPort, hc.InternalPort)}

	if networkName != "" {
		args = append(args, "--network", networkName)
	}

	for k, v := range envs { // test-provided envs
		args = append(args, "-e", fmt.Sprintf("%s=%s", k, v))
	}
	args = append(args, imageName)

	if out, runErr := exec.Command("docker", args...).CombinedOutput(); runErr != nil {
		return fmt.Errorf("docker run failed: %v: %s", runErr, string(out))
	}
	t.Logf("Started container %s from image %s (mapped host %d -> %d)", containerName, imageName, hostPort, hc.InternalPort)

	defer func() {
		logs := fetchContainerLogs(containerName)
		t.Logf("Container logs:\n%s", logs)
		stopContainer(containerName)
	}()

	// Small grace period: some servers bind after initial log lines
	time.Sleep(containerStartGracePeriod)
	url := fmt.Sprintf("http://127.0.0.1:%d%s", hostPort, hc.Path)
	if err := waitForHTTPStatus(url, hc.Expected); err != nil {
		return fmt.Errorf("http check failed: %w", err)
	}

	t.Logf("HTTP check passed (%s => %d)", hc.Path, hc.Expected)
	return nil
}

func uniqueContainerName(parts ...string) string {
	return fmt.Sprintf("railpack-test-%s", strings.Join(parts, "-"))
}

func pickFreePort() (int, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

func waitForHTTPStatus(url string, expected int) error {
	deadline := time.Now().Add(httpCheckTimeout)
	client := &http.Client{Timeout: httpCheckClientTimeout}

	var lastErr error
	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == expected {
				return nil
			}
			lastErr = fmt.Errorf("http status %d != %d", resp.StatusCode, expected)
		} else {
			lastErr = err
		}

		time.Sleep(httpCheckInterval)
	}

	if lastErr == nil {
		// if we haven't returned, and the HTTP client failed, then we've timed out
		lastErr = fmt.Errorf("timeout waiting for %s", url)
	}

	return lastErr
}

func fetchContainerLogs(name string) string {
	out, _ := exec.Command("docker", "logs", name).CombinedOutput()
	return string(out)
}

func stopContainer(name string) { _ = exec.Command("docker", "rm", "-f", name).Run() }
