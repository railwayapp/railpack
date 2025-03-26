package utils

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseSemver(t *testing.T) {
	tests := []struct {
		name        string
		version     string
		want        *Semver
		wantErr     bool
		errorPrefix string
	}{
		{
			name:    "basic semver",
			version: "1.2.3",
			want:    &Semver{Major: 1, Minor: 2, Patch: 3},
			wantErr: false,
		},
		{
			name:    "with v prefix",
			version: "v1.2.3",
			want:    &Semver{Major: 1, Minor: 2, Patch: 3},
			wantErr: false,
		},
		{
			name:        "with other prefix",
			version:     "ruby-2.3.4",
			want:        nil,
			wantErr:     true,
			errorPrefix: "invalid major version",
		},
		{
			name:    "with alpha suffix",
			version: "1.2.3-alpha",
			want:    &Semver{Major: 1, Minor: 2, Patch: 3},
			wantErr: false,
		},
		{
			name:    "with beta suffix",
			version: "1.2.3-beta",
			want:    &Semver{Major: 1, Minor: 2, Patch: 3},
			wantErr: false,
		},
		{
			name:        "too few parts",
			version:     "1.2",
			want:        nil,
			wantErr:     true,
			errorPrefix: "invalid semver",
		},
		{
			name:        "invalid major",
			version:     "a.2.3",
			want:        nil,
			wantErr:     true,
			errorPrefix: "invalid major version",
		},
		{
			name:        "invalid minor",
			version:     "1.b.3",
			want:        nil,
			wantErr:     true,
			errorPrefix: "invalid minor version",
		},
		{
			name:        "invalid patch",
			version:     "1.2.c",
			want:        nil,
			wantErr:     true,
			errorPrefix: "invalid patch version",
		},
		{
			name:        "invalid",
			version:     "1-23",
			want:        nil,
			wantErr:     true,
			errorPrefix: "invalid semver",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSemver(tt.version)

			// Check error cases
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSemver() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// If we expect an error, check the error message prefix
			if tt.wantErr && err != nil {
				if !strings.HasPrefix(err.Error(), tt.errorPrefix) {
					t.Errorf("ParseSemver() error = %v, wantErrPrefix %v", err, tt.errorPrefix)
				}
				return
			}

			// Check valid result cases
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseSemver() = %v, want %v", got, tt.want)
			}
		})
	}
}
