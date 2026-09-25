package resolver

import (
	"strings"

	"github.com/Masterminds/semver/v3"
)

func resolveToFuzzyVersion(version string) string {
	// Remove any whitespace
	version = strings.TrimSpace(version)

	// Handle empty string and "*" cases
	if version == "" || version == "*" {
		return "latest"
	}

	// Handle range notation (e.g. ">=22 <23" or ">= 22" or ">=20.0.0")
	if strings.Contains(version, ">=") || strings.Contains(version, "<") {
		parts := strings.Fields(version)
		for i, part := range parts {
			if after, ok := strings.CutPrefix(part, ">="); ok {
				// Version number is either after the >= in this part, or in the next part
				v := after
				if v == "" && i+1 < len(parts) {
					v = parts[i+1]
				}
				return strings.Split(strings.TrimSpace(v), ".")[0]
			}
		}
	}

	// Handle caret notation by only keeping major version
	if after, ok := strings.CutPrefix(version, "^"); ok {
		version = after
		parts := strings.Split(version, ".")
		return parts[0]
	}

	// Remove any prefix characters (~, v)
	version = strings.TrimPrefix(version, "~")
	version = strings.TrimPrefix(version, "v")

	// Replace .x with empty string (e.g. "14.x" -> "14")
	version = strings.ReplaceAll(version, ".x", "")

	// Remove any trailing dots
	version = strings.TrimRight(version, ".")

	return version
}

// rangeConstraint parses the version strings whose meaning the fuzzy prefix form cannot
// carry: an upper bound, a disjunction, a hyphen range, or a tilde range.
//
// mise only understands prefix queries (`bun@1`), so ">=1.1.0 <1.3.0" degrades to "1"
// and mise happily installs 1.4.x. For these we enumerate the available versions and
// pick the match ourselves. Anything the prefix form already expresses faithfully
// (exact versions, "14.x", "^20", ">=22") keeps the cheaper mise-side resolution.
func rangeConstraint(version string) (*semver.Constraints, bool) {
	version = strings.TrimSpace(version)
	if !needsConstraintResolution(version) {
		return nil, false
	}

	constraint, err := semver.NewConstraint(version)
	if err != nil {
		// Not a constraint we understand; leave it to the existing fuzzy path so behavior
		// is unchanged for anything exotic.
		return nil, false
	}

	return constraint, true
}

func needsConstraintResolution(version string) bool {
	if strings.HasPrefix(version, "~") {
		return true
	}

	return strings.Contains(version, "<") ||
		strings.Contains(version, "||") ||
		strings.Contains(version, " - ")
}

// newestMatching returns the highest version in versions that satisfies constraint and
// passes isAvailable. versions is ascending, so we walk it backwards.
func newestMatching(versions []string, constraint *semver.Constraints, isAvailable func(string) bool) string {
	for i := len(versions) - 1; i >= 0; i-- {
		parsed, err := semver.NewVersion(versions[i])
		if err != nil {
			continue
		}

		if !constraint.Check(parsed) {
			continue
		}

		if isAvailable != nil && !isAvailable(versions[i]) {
			continue
		}

		return versions[i]
	}

	return ""
}
