package plan

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestUnmarshalInput(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected Layer
		wantErr  bool
	}{
		{
			name:     "JSON step input",
			input:    []byte(`{"step": "build", "include": ["src"]}`),
			expected: NewStepLayer("build", Filter{Include: []string{"src"}}),
			wantErr:  false,
		},
		{
			name:     "JSON image input",
			input:    []byte(`{"image": "golang:1.21", "exclude": ["tmp"]}`),
			expected: NewImageLayer("golang:1.21", Filter{Exclude: []string{"tmp"}}),
			wantErr:  false,
		},
		{
			name:     "JSON local input",
			input:    []byte(`{"local": true, "include": ["."]}`),
			expected: NewLocalLayer("."),
			wantErr:  false,
		},
		{
			name:     "String local input with dot",
			input:    []byte(`"."`),
			expected: NewLocalLayer("."),
			wantErr:  false,
		},
		{
			name:     "String spread input",
			input:    []byte(`"..."`),
			expected: Layer{Spread: true},
			wantErr:  false,
		},
		{
			name:     "String step input",
			input:    []byte(`"$build"`),
			expected: NewStepLayer("build"),
			wantErr:  false,
		},
		{
			name:    "Invalid string input",
			input:   []byte(`"invalid"`),
			wantErr: true,
		},
		{
			name:    "Invalid JSON input",
			input:   []byte(`{"invalid": json}`),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got Layer
			err := json.Unmarshal(tt.input, &got)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalInput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if diff := cmp.Diff(tt.expected, got); diff != "" {
					t.Errorf("UnmarshalInput() mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}
