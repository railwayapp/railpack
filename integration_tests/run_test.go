package integration_tests

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/railwayapp/railpack/buildkit"
	"github.com/railwayapp/railpack/core"
	"github.com/railwayapp/railpack/core/app"
	"github.com/stretchr/testify/require"
)

var buildkitCacheImport = flag.String("buildkit-cache-import", "", "BuildKit cache import configuration")
var buildkitCacheExport = flag.String("buildkit-cache-export", "", "BuildKit cache export configuration")

type TestCase struct {
	ExpectedOutput string            `json:"expectedOutput"`
	Platform       string            `json:"platform"`
	Envs           map[string]string `json:"envs"`
	ConfigFilePath string            `json:"configFile"`
	JustBuild      bool              `json:"justBuild"`
	ShouldFail     bool              `json:"shouldFail"`
}

func TestExamplesIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	wd, err := os.Getwd()
	require.NoError(t, err)

	examplesDir := filepath.Join(filepath.Dir(wd), "examples")
	entries, err := os.ReadDir(examplesDir)
	require.NoError(t, err)

	for _, entry := range entries {
		entry := entry // capture for parallel execution
		if !entry.IsDir() {
			continue
		}

		testConfigPath := filepath.Join(examplesDir, entry.Name(), "test.json")
		if _, err := os.Stat(testConfigPath); os.IsNotExist(err) {
			continue
		}

		testConfigBytes, err := os.ReadFile(testConfigPath)
		require.NoError(t, err)

		var testCases []TestCase
		err = json.Unmarshal(testConfigBytes, &testCases)
		require.NoError(t, err)

		for i, testCase := range testCases {
			testCase := testCase // capture for parallel execution
			i := i

			testName := fmt.Sprintf("%s/case-%d", entry.Name(), i)
			t.Run(testName, func(t *testing.T) {
				t.Parallel()

				examplePath := filepath.Join(examplesDir, entry.Name())
				userApp, err := app.NewApp(examplePath)
				if err != nil {
					t.Fatalf("failed to create app: %v", err)
				}

				env := app.NewEnvironment(&testCase.Envs)
				buildResult := core.GenerateBuildPlan(userApp, env, &core.GenerateBuildPlanOptions{
					ConfigFilePath: testCase.ConfigFilePath,
				})

				// Handle case where we expect the build to fail
				if testCase.ShouldFail {
					if buildResult.Success {
						t.Fatalf("expected build to fail, but it succeeded")
					}
					// Test passes - build failed as expected
					return
				}

				if !buildResult.Success {
					t.Fatalf("failed to generate build plan: %v", buildResult.Logs)
				}
				if buildResult == nil {
					t.Fatal("build result is nil")
				}

				imageName := fmt.Sprintf("railpack-test-%s-%s",
					strings.ToLower(strings.ReplaceAll(testName, "/", "-")),
					strings.ToLower(uuid.New().String()))

				if err := buildkit.BuildWithBuildkitClient(examplePath, buildResult.Plan, buildkit.BuildWithBuildkitClientOptions{
					ImageName:   imageName,
					Platform:    testCase.Platform,
					ImportCache: *buildkitCacheImport,
					ExportCache: *buildkitCacheExport,
					Secrets:     testCase.Envs,
					CacheKey:    imageName,
					// Pass through GITHUB_TOKEN if it exists, this avoids mise timeouts during build
					// this can easily occur since we run all integration tests in parallel via GHA
					GitHubToken: os.Getenv("GITHUB_TOKEN"),
				}); err != nil {
					t.Fatalf("failed to build image: %v", err)
				}

				if testCase.JustBuild {
					return
				}

				if err := runContainerWithTimeout(t, imageName, testCase.ExpectedOutput, testCase.Envs, testCase.Platform); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}

func cmdDoneChan(cmd *exec.Cmd) chan error {
	ch := make(chan error, 1)
	go func() { ch <- cmd.Wait() }()
	return ch
}

func runContainerWithTimeout(t *testing.T, imageName, expectedOutput string, envs map[string]string, platformStr string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Generate a unique container name so we can reference it later for cleanup
	containerName := fmt.Sprintf("railpack-test-%s", uuid.New().String())

	// Build docker run command with environment variables
	args := []string{"run", "--rm", "--name", containerName}

	// Add platform specification if provided
	if platformStr != "" {
		args = append(args, "--platform", platformStr)
	}

	for key, value := range envs {
		args = append(args, "-e", fmt.Sprintf("%s=%s", key, value))
	}
	args = append(args, imageName)

	cmd := exec.CommandContext(ctx, "docker", args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %v", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start container: %v", err)
	}

	// Ensure cleanup on function exit
	defer func() {
		// Stop the container if it's still running
		stopCmd := exec.Command("docker", "stop", containerName)
		_ = stopCmd.Run()
		// Remove the container if it still exists
		rmCmd := exec.Command("docker", "rm", "-f", containerName)
		_ = rmCmd.Run()
	}()

	var output, errOutput strings.Builder
	done := make(chan error, 1)
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			output.WriteString(line + "\n")
			if strings.Contains(line, expectedOutput) {
				done <- nil
				return
			}
		}
		if err := scanner.Err(); err != nil {
			done <- fmt.Errorf("error reading stdout: %v", err)
			return
		}
		done <- fmt.Errorf("container output:\n%s\nErrors:\n%s", output.String(), errOutput.String())
	}()

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			errOutput.WriteString(scanner.Text() + "\n")
		}
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("container timed out after 2 minutes")
	case err := <-done:
		if err != nil {
			require.Contains(t, output.String(), expectedOutput, "container output did not contain expected string")
			return err
		}
		return nil
	case err := <-cmdDoneChan(cmd):
		if err != nil && !strings.Contains(err.Error(), "signal: killed") {
			return fmt.Errorf("container failed: %v", err)
		}
		require.Contains(t, output.String(), expectedOutput, "container output did not contain expected string")
		return nil
	}
}
