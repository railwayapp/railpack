package buildkit

import (
	"testing"
)

func TestParsePlatform(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected BuildPlatform
		wantErr  bool
	}{
		{
			name:  "empty string returns host platform",
			input: "",
			expected: func() BuildPlatform {
				return DetermineBuildPlatformFromHost()
			}(),
			wantErr: false,
		},
		{
			name:  "linux/arm64/v8",
			input: "linux/arm64/v8",
			expected: BuildPlatform{
				OS:           "linux",
				Architecture: "arm64",
				Variant:      "v8",
			},
			wantErr: false,
		},
		{
			name:    "invalid input",
			input:   "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePlatform(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParsePlatform() expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("ParsePlatform() error = %v", err)
				return
			}
			if got != tt.expected {
				t.Errorf("ParsePlatform() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestBuildPlatformString(t *testing.T) {
	platform := BuildPlatform{
		OS:           "linux",
		Architecture: "arm64",
		Variant:      "v8",
	}
	expected := "linux/arm64/v8"

	got := platform.String()
	if got != expected {
		t.Errorf("BuildPlatform.String() = %v, want %v", got, expected)
	}
}
