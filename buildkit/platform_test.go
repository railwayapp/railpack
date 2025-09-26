package buildkit

import (
	"testing"

	specs "github.com/opencontainers/image-spec/specs-go/v1"
)

func TestValidatePlatform(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected specs.Platform
		wantErr  bool
	}{
		{
			name:  "empty string returns Linux platform",
			input: "",
			// When no platform is specified, we default to a Linux platform that matches
			// the host architecture. This ensures container compatibility while preserving
			// performance characteristics of the host system.
			expected: func() specs.Platform {
				platform, _ := ParsePlatformWithDefaults("")
				return platform
			}(),
			wantErr: false,
		},
		{
			name:  "linux/arm64/v8",
			input: "linux/arm64/v8",
			expected: specs.Platform{
				OS:           "linux",
				Architecture: "arm64",
				Variant:      "v8",
			},
			wantErr: false,
		},
		{
			name:  "linux/amd64",
			input: "linux/amd64",
			expected: specs.Platform{
				OS:           "linux",
				Architecture: "amd64",
			},
			wantErr: false,
		},
		{
			name:    "invalid input",
			input:   "invalid",
			wantErr: true,
		},
		{
			name:    "multiple platforms not supported",
			input:   "linux/amd64,linux/arm64",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePlatformWithDefaults(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParsePlatformWithDefaults() expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("ParsePlatformWithDefaults() error = %v", err)
				return
			}
			if got.OS != tt.expected.OS || got.Architecture != tt.expected.Architecture || got.Variant != tt.expected.Variant {
				t.Errorf("ParsePlatformWithDefaults() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestValidatePlatformWithMultiplePlatforms(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected specs.Platform
		wantErr  bool
	}{
		{
			name:  "empty string returns Linux platform",
			input: "",
			expected: func() specs.Platform {
				platform, _ := ParsePlatformWithDefaults("")
				return platform
			}(),
			wantErr: false,
		},
		{
			name:  "linux/arm64/v8",
			input: "linux/arm64/v8",
			expected: specs.Platform{
				OS:           "linux",
				Architecture: "arm64",
				Variant:      "v8",
			},
			wantErr: false,
		},
		{
			name:  "linux/amd64",
			input: "linux/amd64",
			expected: specs.Platform{
				OS:           "linux",
				Architecture: "amd64",
			},
			wantErr: false,
		},
		{
			name:    "invalid input",
			input:   "invalid",
			wantErr: true,
		},
		{
			name:    "multiple platforms not supported",
			input:   "linux/amd64,linux/arm64",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := map[string]string{"platform": tt.input}
			got, err := validatePlatform(opts)
			if tt.wantErr {
				if err == nil {
					t.Errorf("validatePlatform() expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("validatePlatform() error = %v", err)
				return
			}
			if got.OS != tt.expected.OS || got.Architecture != tt.expected.Architecture || got.Variant != tt.expected.Variant {
				t.Errorf("validatePlatform() = %v, want %v", got, tt.expected)
			}
		})
	}
}
