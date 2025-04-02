package utils

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

func RemoveDuplicates[T comparable](sliceList []T) []T {
	allKeys := make(map[T]bool)
	list := []T{}
	for _, item := range sliceList {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

// MergeStringSlicePointers combines multiple string slice pointers, deduplicates values, and sorts them
func MergeStringSlicePointers(slices ...*[]string) *[]string {
	if len(slices) == 0 {
		return nil
	}

	var allStrings []string
	for _, slice := range slices {
		if slice != nil {
			allStrings = append(allStrings, *slice...)
		}
	}

	if len(allStrings) == 0 {
		return nil
	}

	// Deduplicate and sort
	seen := make(map[string]bool)
	var uniqueStrings []string
	for _, s := range allStrings {
		if !seen[s] {
			seen[s] = true
			uniqueStrings = append(uniqueStrings, s)
		}
	}
	sort.Strings(uniqueStrings)
	return &uniqueStrings
}

// CapitalizeFirst converts the first character of a string to uppercase.
// The rest of the string remains unchanged.
// Examples:
//   - "hello" -> "Hello"
//   - "world" -> "World"
//   - "" -> ""
//   - "already Capitalized" -> "Already Capitalized"
func CapitalizeFirst(s string) string {
	if s == "" {
		return ""
	}

	runes := []rune(s)
	runes[0] = []rune(strings.ToUpper(string(runes[0])))[0]
	return string(runes)
}

// ParsePackageWithVersion parses a slice of package specifications in the format "name@version"
// and returns a map of package names to their versions.
// If a package has no version specified (no @ symbol), it defaults to "latest".
// Examples:
//   - ["node@14.2"] -> {"node": "14.2"}
//   - ["python"] -> {"python": "latest"}
//   - ["ruby@3.0.0", "go"] -> {"ruby": "3.0.0", "go": "latest"}
//   - ["node@^14.3", "python@>=3.9"] -> {"node": "^14.3", "python": ">=3.9"}
func ParsePackageWithVersion(versions []string) map[string]string {
	parsedVersions := make(map[string]string)

	for _, version := range versions {
		parts := strings.Split(version, "@")
		if len(parts) == 1 {
			parsedVersions[parts[0]] = "latest"
		} else {
			parsedVersions[parts[0]] = parts[1]
		}
	}

	return parsedVersions
}

// ExtractSemverVersion extracts the first version number found in a string.
// It supports full semver (major.minor.patch) as well as partial versions (major or major.minor).
// The version can appear anywhere in the string and can have prefixes or suffixes.
// Examples:
//   - "1.2.3" -> "1.2.3"
//   - "v1.2.3" -> "1.2.3"
//   - "python-3.10.7" -> "3.10.7"
//   - "version1.2.3-beta" -> "1.2.3"
//   - "requires node 14.2" -> "14.2"
//   - "python 3" -> "3"
//
// Returns an empty string if no version number is found.
func ExtractSemverVersion(version string) string {
	semverRe := regexp.MustCompile(`(\d+(?:\.\d+)?(?:\.\d+)?)`)
	matches := semverRe.FindStringSubmatch(strings.TrimSpace(version))
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// ParseSemver parses a semantic version string and returns a Semver struct.
func ParseSemver(version string) (*Semver, error) {
	// Remove potential prefixes like "v" or "ruby-"
	version = strings.TrimPrefix(version, "v")

	// Split by dots
	parts := strings.Split(version, ".")

	// Parse major version
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid major version: %s", parts[0])
	}

	minor := 0
	patch := 0
	if len(parts) > 1 {
		// Parse minor version
		minorVer, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("invalid minor version: %s", parts[1])
		}
		minor = minorVer
	}

	if len(parts) > 2 {
		// Parse patch version - take care of potential suffixes like "-alpha", "-beta", etc.
		patchStr := parts[2]
		patchParts := strings.Split(patchStr, "-")
		patchVer, err := strconv.Atoi(patchParts[0])
		if err != nil {
			return nil, fmt.Errorf("invalid patch version: %s", patchParts[0])
		}
		patch = patchVer
	}

	return &Semver{
		Major: major,
		Minor: minor,
		Patch: patch,
	}, nil
}

type Semver struct {
	Major int
	Minor int
	Patch int
}
