package integration_tests

import (
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

const (
	// Grace period for servers to bind after initial log lines.
	containerStartGracePeriod = 500 * time.Millisecond
	// Total timeout for the HTTP check to succeed. Instead of specifying the number of retries, we use a total timeout.
	httpCheckTimeout = 35 * time.Second
	// Time between HTTP check retries.
	httpCheckInterval = 300 * time.Millisecond
	// Timeout for a single HTTP GET request.
	httpCheckClientTimeout = 3 * time.Second
)

// HTTPCheck defines an HTTP endpoint check for a container
// Path defaults to /, Expected defaults to 200.
type HTTPCheck struct {
	InternalPort int    `json:"internalPort"`
	Path         string `json:"path"`
	Expected     int    `json:"expected"`
}

// runContainerWithHTTPCheck starts a container from the given image,
// maps the internalPort to a random host port, and checks the given path returns the expected status code.
func runContainerWithHTTPCheck(t *testing.T, imageName string, envs map[string]string, hc *HTTPCheck) error {
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
