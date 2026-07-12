package buildkit

import (
	"maps"
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

func TestExtractCacheType(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedType  string
		expectedAttrs map[string]string
	}{
		{
			name:         "registry cache type is extracted",
			input:        "type=registry,ref=localhost:5555/cache",
			expectedType: "registry",
			expectedAttrs: map[string]string{
				"ref": "localhost:5555/cache",
			},
		},
		{
			name:         "missing type is not defaulted",
			input:        "scope=my-cache,mode=max",
			expectedType: "",
			expectedAttrs: map[string]string{
				"scope": "my-cache",
				"mode":  "max",
			},
		},
		{
			name:          "empty attrs",
			input:         "",
			expectedType:  "",
			expectedAttrs: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotType, gotAttrs := extractCacheType(parseKeyValue(tt.input))
			if gotType != tt.expectedType {
				t.Errorf("extractCacheType(%q) type = %q, want %q", tt.input, gotType, tt.expectedType)
			}
			if !maps.Equal(gotAttrs, tt.expectedAttrs) {
				t.Errorf("extractCacheType(%q) attrs = %v, want %v", tt.input, gotAttrs, tt.expectedAttrs)
			}
		})
	}
}
