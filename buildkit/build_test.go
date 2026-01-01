package buildkit

import (
	"path/filepath"
	"testing"
)

func TestGetImageName(t *testing.T) {
	tests := []struct {
		name     string
		appDir   string
		expected string
	}{
		{
			name:     "lowercase directory name",
			appDir:   "/path/to/myapp",
			expected: "myapp",
		},
		{
			name:     "uppercase directory name is lowercased",
			appDir:   "/path/to/SteelMC",
			expected: "steelmc",
		},
		{
			name:     "mixed case directory name is lowercased",
			appDir:   "/path/to/MyApp",
			expected: "myapp",
		},
		{
			name:     "directory name with hyphens",
			appDir:   "/path/to/my-app",
			expected: "my-app",
		},
		{
			name:     "empty path returns fallback",
			appDir:   "",
			expected: "railpack-app",
		},
		{
			name:     "path ending with separator",
			appDir:   "/path/to/myapp" + string(filepath.Separator),
			expected: "railpack-app",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getImageName(tt.appDir)
			if got != tt.expected {
				t.Errorf("getImageName(%q) = %q, want %q", tt.appDir, got, tt.expected)
			}
		})
	}
}
