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
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/railwayapp/railpack/buildkit"
	"github.com/railwayapp/railpack/core"
	"github.com/railwayapp/railpack/core/app"
	"github.com/railwayapp/railpack/internal/utils"
	"github.com/stretchr/testify/require"
)

var buildkitCacheImport = flag.String("buildkit-cache-import", "", "BuildKit cache import configuration")
var buildkitCacheExport = flag.String("buildkit-cache-export", "", "BuildKit cache export configuration")

type StringOrArray []string

func (s *StringOrArray) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		*s = []string{str}
		return nil
	}

	var arr []string
	if err := json.Unmarshal(data, &arr); err != nil {
		return err
	}

	*s = arr
	return nil
}

type TestCase struct {
	Platform string `json:"platform"`
	// can be a single string, or an array of strings for multiple expected outputs
	// matches against the entire output of the container if it cannot be found in a single line
	ExpectedOutput StringOrArray     `json:"expectedOutput"`
	Envs           map[string]string `json:"envs"`
	ConfigFilePath string            `json:"configFile"`
	JustBuild      bool              `json:"justBuild"`
	ShouldFail     bool              `json:"shouldFail"`
	HTTPCheck      *HTTPCheck        `json:"httpCheck"`
	StderrAllowed  bool              `json:"stderrAllowed"`
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

		// read in the test assertions in each example and generate a unique test case for each entry in the JSON array
		testConfigPath := filepath.Join(examplesDir, entry.Name(), "test.json")
		if _, err := os.Stat(testConfigPath); os.IsNotExist(err) {
			continue
		}

		testConfigBytes, err := os.ReadFile(testConfigPath)
		require.NoError(t, err)

		// allow json5/hujson in the test.json file
		testConfigBytes, err = utils.StandardizeJSON(testConfigBytes)
		require.NoError(t, err)

		// extract test.json from the example directory, and fail if unexpected fields are present
		var testCases []TestCase
		decoder := json.NewDecoder(strings.NewReader(string(testConfigBytes)))
		decoder.DisallowUnknownFields()
		err = decoder.Decode(&testCases)
		require.NoError(t, err)

		// Validate test case configuration
		for i, testCase := range testCases {
			// Check if both httpCheck and expectedOutput are specified in the same test case
			if testCase.HTTPCheck != nil && len(testCase.ExpectedOutput) > 0 {
				t.Fatalf("%s case-%d: cannot have both httpCheck and expectedOutput in the same test case", entry.Name(), i)
			}

			// Check if justBuild is used alongside other test cases
			if testCase.JustBuild && len(testCases) > 1 {
				t.Fatalf("%s: justBuild can only be used alone (no other test cases allowed in the same file)", entry.Name())
			}
		}

		examplePath := filepath.Join(examplesDir, entry.Name())

		// each entry in the tests.json array is an individual test case which runs it's own container
		for i, testCase := range testCases {
			testCase := testCase // capture for parallel execution
			i := i

			testName := fmt.Sprintf("%s/case-%d", entry.Name(), i)
			t.Run(testName, func(t *testing.T) {
				t.Parallel()

				fmt.Printf("\033[32mRunning: examples/%s\033[0m\n", entry.Name())

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

				// strictly for debugging when attempting to reproduce and compare a build locally
				core.PrettyPrintBuildResult(buildResult)

				// generate a completely random, but readable, image name for the example project
				imageName := uniqueContainerName(
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

				// only need to calculate the size of the image *once*, not for every test case
				// this is for the image benchmarking system
				if i == 0 {
					if sizeBytes, err := getImageSize(imageName); err == nil {
						t.Logf("image size for %s: %s", entry.Name(), formatBytes(sizeBytes))
						folderSizes, err := getFolderSizes(imageName)
						if err != nil {
							t.Logf("warning: failed to get folder sizes: %v", err)
						}
						if err := writeImageSize(examplePath, entry.Name(), sizeBytes, folderSizes); err != nil {
							t.Logf("warning: failed to write size.json: %v", err)
						}
					} else {
						t.Logf("warning: failed to get image size: %v", err)
					}
				}

				if testCase.JustBuild {
					return
				}

				// Start docker-compose services for this test case if they exist
				composeConfig, err := detectAndStartCompose(examplePath, t)
				if err != nil {
					t.Fatalf("failed to start compose for %s: %v", entry.Name(), err)
				}

				if composeConfig != nil {
					t.Cleanup(func() {
						if err := stopAndCleanupCompose(composeConfig, t); err != nil {
							t.Errorf("failed to cleanup docker-compose for %s: %v", entry.Name(), err)
						}
					})
				}

				networkName := ""
				if composeConfig != nil {
					networkName = composeConfig.NetworkName
				}

				if testCase.HTTPCheck != nil {
					if err := runContainerWithHTTPCheck(t, imageName, testCase.Envs, testCase.HTTPCheck, networkName); err != nil {
						t.Fatal(err)
					}
					return
				}

				if err := runContainerWithTimeout(t, imageName, testCase.ExpectedOutput, testCase.Envs, testCase.Platform, networkName, testCase.StderrAllowed); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}

func getImageSize(imageName string) (int64, error) {
	out, err := exec.Command("docker", "image", "inspect", imageName, "--format", "{{.Size}}").Output()
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(strings.TrimSpace(string(out)), 10, 64)
}

// returns the byte size of each root-level directory in the
// image. This is strictly informational for developers to understand what is
// contributing to image size
func getFolderSizes(imageName string) (map[string]int64, error) {
	out, err := exec.Command("docker", "run", "--rm", "--entrypoint", "sh", imageName, "-c", "du -sb /* 2>/dev/null || true").Output()
	if err != nil {
		return nil, err
	}

	skip := map[string]bool{"/proc": true, "/sys": true, "/dev": true}
	sizes := map[string]int64{}
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 {
			continue
		}
		path := strings.TrimSpace(parts[1])
		if skip[path] {
			continue
		}
		size, err := strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 64)
		if err != nil {
			continue
		}
		sizes[path] = size
	}
	return sizes, nil
}

func formatBytes(bytes int64) string {
	const mb = 1024 * 1024
	return fmt.Sprintf("%.1f MB", float64(bytes)/mb)
}

// Writes image size metadata to size.json in the example directory so CI can
// collect and track image sizes over time via benchmark-action.
func writeImageSize(examplePath string, name string, sizeBytes int64, folderSizes map[string]int64) error {
	data := map[string]any{
		// not required, but makes it easier to identify which example a size.json was tied to
		"name":      name,
		"size":      sizeBytes,
		"sizeHuman": formatBytes(sizeBytes),
		"folders":   folderSizes,
	}
	out, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(examplePath, "size.json"), out, 0644)
}

// wait until cmd subprocess exits
func cmdDoneChan(cmd *exec.Cmd) chan error {
	ch := make(chan error, 1)
	go func() { ch <- cmd.Wait() }()
	return ch
}

// this is run for each test case generated by test.json
// If networkName is provided, the container will be connected to that network.
func runContainerWithTimeout(t *testing.T, imageName string, expectedOutputs []string, envs map[string]string, platformStr string, networkName string, stderrAllowed bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Generate a unique container name so we can reference it later for cleanup
	containerName := uniqueContainerName(uuid.New().String())

	// Build docker run command with environment variables
	args := []string{"run", "--rm", "--name", containerName}

	// Add platform specification if provided
	if platformStr != "" {
		args = append(args, "--platform", platformStr)
	}

	// Add network if provided
	if networkName != "" {
		args = append(args, "--network", networkName)
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

	// read container output and look for expected output
	var output, stdErrOutput strings.Builder
	done := make(chan error, 1)
	go func() {
		foundOutputs := make([]bool, len(expectedOutputs))

		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			output.WriteString(line + "\n")

			// Check if this line contains any of the expected outputs
			for i, expected := range expectedOutputs {
				if !foundOutputs[i] && strings.Contains(line, expected) {
					foundOutputs[i] = true
				}
			}

			// Check if all expected outputs have been found
			if !slices.Contains(foundOutputs, false) {
				done <- nil
				return
			}
		}
		if err := scanner.Err(); err != nil {
			done <- fmt.Errorf("error reading stdout: %v", err)
			return
		}
		done <- fmt.Errorf("container stdout output:\n%s\nStderr output:\n%s", output.String(), stdErrOutput.String())
	}()

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			stdErrOutput.WriteString(scanner.Text() + "\n")
		}
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("container timed out after 2 minutes")
	case err := <-done:
		if err != nil {
			return fmt.Errorf("%v\ncontainer output did not contain expected string:\n%v", err, expectedOutputs)
		}
		// if err == nil, then we found the expected output
		// check that stderr is empty (unless explicitly allowed)
		if !stderrAllowed {
			stderr := stdErrOutput.String()
			if stderr != "" {
				return fmt.Errorf("expected stderr to be empty, but got:\n%s", stderr)
			}
		}
		return nil
	case err := <-cmdDoneChan(cmd):
		if err != nil && !strings.Contains(err.Error(), "signal: killed") {
			return fmt.Errorf("container failed: %v", err)
		}

		return fmt.Errorf("container output did not contain expected string:\n%v", expectedOutputs)
	}
}
