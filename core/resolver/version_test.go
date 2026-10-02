package resolver

import (
	"testing"

	"github.com/Masterminds/semver/v3"
)

func TestResolveToFuzzyVersion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple major version", "14", "14"},
		{"major.minor version", "18.2", "18.2"},
		{"major.minor.patch version", "18.3.2", "18.3.2"},
		{"x notation", "14.x", "14"},
		{"x notation with minor", "14.2.x", "14.2"},
		{"range notation", ">=22 <23", "22"},
		{"range notation major only", ">= 22", "22"},
		{"range notation with minor", ">=22.1 <23", "22"},
		{"range notation with patch", ">=22.1.3 <23", "22"},
		{"range notation simple", ">=20.0.0", "20"},
		{"range notation major.minor", ">=20.4", "20"},
		{"caret notation", "^14.3.2", "14"},
		{"caret notation minor", "^14.3", "14"},
		{"caret notation minor", "^20.0.0", "20"},
		{"caret notation major", "^14", "14"},
		{"tilde notation", "~14.3.2", "14.3.2"},
		{"v prefix", "v14.3.2", "14.3.2"},
		{"empty string", "", "latest"},
		{"star wildcard", "*", "latest"},
		{"whitespace", "  14.3  ", "14.3"},
		{"multiple x", "14.x.x", "14"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolveToFuzzyVersion(tt.input)
			if result != tt.expected {
				t.Errorf("resolveToFuzzyVersion(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestRangeConstraint(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		wantConstraint bool
		// versions that must / must not satisfy the parsed constraint
		satisfied    []string
		notSatisfied []string
	}{
		{name: "exact version keeps fuzzy path", input: "18.3.2", wantConstraint: false},
		{name: "x notation keeps fuzzy path", input: "14.x", wantConstraint: false},
		{name: "caret keeps fuzzy path", input: "^20.0.0", wantConstraint: false},
		{name: "lower bound only keeps fuzzy path", input: ">=20.0.0", wantConstraint: false},
		{name: "latest keeps fuzzy path", input: "", wantConstraint: false},
		{
			name:           "bounded range",
			input:          ">=1.1.0 <1.3.0",
			wantConstraint: true,
			satisfied:      []string{"1.1.0", "1.2.9"},
			notSatisfied:   []string{"1.0.9", "1.3.0", "1.4.0"},
		},
		{
			name:           "upper bound only",
			input:          "<2",
			wantConstraint: true,
			satisfied:      []string{"1.9.9"},
			notSatisfied:   []string{"2.0.0"},
		},
		{
			name:           "disjunction",
			input:          "20 || 22",
			wantConstraint: true,
			satisfied:      []string{"20.1.0", "22.0.0"},
			notSatisfied:   []string{"21.0.0", "23.0.0"},
		},
		{
			name:           "tilde pins minor",
			input:          "~1.2.3",
			wantConstraint: true,
			satisfied:      []string{"1.2.3", "1.2.9"},
			notSatisfied:   []string{"1.2.2", "1.3.0"},
		},
		{
			name:           "hyphen range",
			input:          "1.2 - 1.4",
			wantConstraint: true,
			satisfied:      []string{"1.3.0"},
			notSatisfied:   []string{"1.1.0", "1.5.0"},
		},
		{name: "unparseable range falls back", input: "<not-a-version", wantConstraint: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint, ok := rangeConstraint(tt.input)
			if ok != tt.wantConstraint {
				t.Fatalf("rangeConstraint(%q) ok = %v, want %v", tt.input, ok, tt.wantConstraint)
			}
			if !ok {
				return
			}

			for _, v := range tt.satisfied {
				if !constraint.Check(semver.MustParse(v)) {
					t.Errorf("%q should satisfy %q", v, tt.input)
				}
			}
			for _, v := range tt.notSatisfied {
				if constraint.Check(semver.MustParse(v)) {
					t.Errorf("%q should not satisfy %q", v, tt.input)
				}
			}
		})
	}
}

func TestNewestMatching(t *testing.T) {
	versions := []string{"1.0.0", "1.1.0", "1.2.0", "1.2.9", "1.3.0", "1.4.0"}

	t.Run("picks the newest version within the range", func(t *testing.T) {
		constraint, _ := rangeConstraint(">=1.1.0 <1.3.0")
		if got := newestMatching(versions, constraint, nil); got != "1.2.9" {
			t.Errorf("got %q, want 1.2.9", got)
		}
	})

	t.Run("honors the availability filter", func(t *testing.T) {
		constraint, _ := rangeConstraint(">=1.1.0 <1.3.0")
		isAvailable := func(v string) bool { return v != "1.2.9" }
		if got := newestMatching(versions, constraint, isAvailable); got != "1.2.0" {
			t.Errorf("got %q, want 1.2.0", got)
		}
	})

	t.Run("skips versions mise reports that are not semver", func(t *testing.T) {
		constraint, _ := rangeConstraint("<2")
		if got := newestMatching([]string{"1.0.0", "nightly"}, constraint, nil); got != "1.0.0" {
			t.Errorf("got %q, want 1.0.0", got)
		}
	})

	t.Run("returns empty when nothing matches", func(t *testing.T) {
		constraint, _ := rangeConstraint("<0.5")
		if got := newestMatching(versions, constraint, nil); got != "" {
			t.Errorf("got %q, want empty", got)
		}
	})
}
