package integration_tests

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/containerd/platforms"
	"github.com/google/uuid"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/railwayapp/railpack/buildkit"
	"github.com/railwayapp/railpack/core"
	"github.com/railwayapp/railpack/core/app"
	"github.com/stretchr/testify/require"
)

type PlatformTestCase struct {
	ExpectedOutput string            `json:"expectedOutput"`
	Platform       string            `json:"platform"`
	Envs           map[string]string `json:"envs"`
	ConfigFilePath string            `json:"configFile"`
	JustBuild      bool              `json:"justBuild"`
	ShouldFail     bool              `json:"shouldFail"`
}

func TestPlatformIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	wd, err := os.Getwd()
	require.NoError(t, err)

	examplesDir := filepath.Join(filepath.Dir(wd), "examples")
	platformExampleDir := filepath.Join(examplesDir, "shell-platform-arch")

	// Check if the platform example exists
	if _, err := os.Stat(platformExampleDir); os.IsNotExist(err) {
		t.Skip("platform example directory does not exist")
	}

	testConfigPath := filepath.Join(platformExampleDir, "test.json")
	if _, err := os.Stat(testConfigPath); os.IsNotExist(err) {
		t.Skip("platform test config does not exist")
	}

	testConfigBytes, err := os.ReadFile(testConfigPath)
	require.NoError(t, err)

	var testCases []PlatformTestCase
	err = json.Unmarshal(testConfigBytes, &testCases)
	require.NoError(t, err)

	for i, testCase := range testCases {
		testCase := testCase // capture for parallel execution
		i := i

		testName := fmt.Sprintf("platform-test-case-%d", i)
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			userApp, err := app.NewApp(platformExampleDir)
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

			imageName := fmt.Sprintf("railpack-platform-test-%s-%s",
				strings.ToLower(strings.ReplaceAll(testName, "/", "-")),
				strings.ToLower(uuid.New().String()))

			// Parse the platform string using our helper function
			platform, err := buildkit.ParsePlatformWithDefaults(testCase.Platform)
			if err != nil {
				t.Fatalf("failed to parse platform %s: %v", testCase.Platform, err)
			}

			if err := buildkit.BuildWithBuildkitClient(platformExampleDir, buildResult.Plan, buildkit.BuildWithBuildkitClientOptions{
				ImageName:   imageName,
				Platform:    platform,
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

			if err := runPlatformContainerWithTimeout(t, imageName, testCase.ExpectedOutput, testCase.Envs, platform); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func runPlatformContainerWithTimeout(t *testing.T, imageName, expectedOutput string, envs map[string]string, platform specs.Platform) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Generate a unique container name so we can reference it later for cleanup
	containerName := fmt.Sprintf("railpack-platform-test-%s", uuid.New().String())

	// Build docker run command with environment variables and platform
	args := []string{"run", "--rm", "--name", containerName}

	// Add platform specification
	platformStr := platforms.Format(platform)
	args = append(args, "--platform", platformStr)

	// Add environment variables
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

	// Read output
	outputChan := make(chan string, 1)
	go func() {
		var output strings.Builder
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			output.WriteString(line + "\n")
		}
		outputChan <- output.String()
	}()

	// Read stderr
	stderrChan := make(chan string, 1)
	go func() {
		var stderrOutput strings.Builder
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			stderrOutput.WriteString(line + "\n")
		}
		stderrChan <- stderrOutput.String()
	}()

	// Wait for command to complete
	done := cmdDoneChan(cmd)
	select {
	case err := <-done:
		if err != nil {
			stderrOutput := <-stderrChan
			return fmt.Errorf("container failed: %v\nstderr: %s", err, stderrOutput)
		}
	case <-ctx.Done():
		_ = cmd.Process.Kill()
		return fmt.Errorf("container timed out")
	}

	// Get the output
	output := <-outputChan
	stderrOutput := <-stderrChan

	// Check if expected output is in the actual output
	if !strings.Contains(output, expectedOutput) {
		return fmt.Errorf("expected output '%s' not found in container output:\n%s\nstderr:\n%s", expectedOutput, output, stderrOutput)
	}

	return nil
}
