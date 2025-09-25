package mise

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/railwayapp/railpack/core/logger"
	"github.com/stretchr/testify/require"
)

type testContext struct {
	appSource string
	logger    *logger.Logger
}

func (t *testContext) GetAppSource() string {
	return t.appSource
}

func (t *testContext) GetLogger() *logger.Logger {
	return t.logger
}

func CreateTestContext(t *testing.T, appDir string) MiseAppContext {
	absAppDir, err := filepath.Abs(appDir)
	if err != nil {
		t.Fatalf("failed to get absolute path: %v", err)
	}

	logger := logger.NewLogger()
	return &testContext{
		appSource: absAppDir,
		logger:    logger,
	}
}

func TestMistGetLatestVersion(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mise-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	mise, err := New(tempDir)
	if err != nil {
		t.Fatalf("failed to create mise: %v", err)
	}

	tests := []struct {
		name       string
		runtime    string
		version    string
		wantPrefix string
		wantErr    bool
	}{
		{
			name:       "node latest version",
			runtime:    "node",
			version:    "22",
			wantPrefix: "22",
		},
		{
			name:    "bun latest version",
			runtime: "bun",
			version: "latest",
		},
		{
			name:    "non-existent latest version",
			runtime: "node",
			version: "999",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := mise.GetLatestVersion(tt.runtime, tt.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLatestVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if tt.wantPrefix != "" && !strings.HasPrefix(got, tt.wantPrefix) {
					t.Errorf("GetLatestVersion() got = %v, want prefix %v", got, tt.wantPrefix)
				}
				if got == "" {
					t.Error("GetLatestVersion() got empty version")
				}
			}
		})
	}
}

func TestMiseGetAllVersions(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mise-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	mise, err := New(tempDir)
	if err != nil {
		t.Fatalf("failed to create mise: %v", err)
	}

	tests := []struct {
		name     string
		runtime  string
		version  string
		versions []string
		wantErr  bool
	}{
		{
			name:     "node all versions",
			runtime:  "node",
			version:  "18.20",
			versions: []string{"18.20.0", "18.20.1", "18.20.2", "18.20.3", "18.20.4", "18.20.5", "18.20.6", "18.20.7", "18.20.8"},
		},
		{
			name:     "bun all versions",
			runtime:  "bun",
			version:  "0.8",
			versions: []string{"0.8.0", "0.8.1"},
		},
		{
			name:     "php all versions",
			runtime:  "php",
			version:  "7.4.2",
			versions: []string{"7.4.2", "7.4.20", "7.4.21", "7.4.22", "7.4.23", "7.4.24", "7.4.25", "7.4.26", "7.4.27", "7.4.28", "7.4.29"},
		},
		{
			name:    "non-existent all versions",
			runtime: "node",
			version: "999",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := mise.GetAllVersions(tt.runtime, tt.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAllVersions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				require.Equal(t, tt.versions, got)
			}
		})
	}
}

func TestGetPackageVersions(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mise-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	mise, err := New(tempDir)
	if err != nil {
		t.Fatalf("failed to create mise: %v", err)
	}

	ctx := CreateTestContext(t, "../../examples/python-uv-tool-versions")

	packages, err := mise.GetPackageVersions(ctx)
	require.NoError(t, err)

	// Expected packages from the example
	expected := map[string]struct{}{
		"python": {},
		"uv":     {},
	}

	// Ensure ONLY expected packages are present
	require.Len(t, packages, len(expected), "unexpected number of packages returned")
	for name := range packages {
		_, ok := expected[name]
		require.True(t, ok, "unexpected package found: %s", name)
	}

	// The python-uv-tool-versions example should have python and uv defined
	require.Contains(t, packages, "python")
	require.Contains(t, packages, "uv")

	// Verify versions are not empty
	require.NotEmpty(t, packages["python"].Version)
	require.NotEmpty(t, packages["uv"].Version)

	// Verify python version starts with "3.9" (as defined in .tool-versions)
	require.True(t, strings.HasPrefix(packages["python"].Version, "3.9"))

	// Verify uv version starts with "0.7" (as defined in .tool-versions)
	require.True(t, strings.HasPrefix(packages["uv"].Version, "0.7"))

	// Verify source types are set
	require.NotEmpty(t, packages["python"].Source)
	require.NotEmpty(t, packages["uv"].Source)
}

func TestGetPackageVersionsWithNoToolVersions(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mise-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	mise, err := New(tempDir)
	if err != nil {
		t.Fatalf("failed to create mise: %v", err)
	}

	// Use an example directory that doesn't have a .tool-versions file
	ctx := CreateTestContext(t, "../../examples/node-tanstack-start")

	packages, err := mise.GetPackageVersions(ctx)
	require.NoError(t, err)

	// Should return empty map for directory with no .tool-versions
	require.Empty(t, packages)
}
