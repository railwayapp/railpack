package plan

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/moby/patternmatcher/ignorefile"
)

// CheckAndParseDockerignore checks if a .dockerignore file exists and parses it
func CheckAndParseDockerignore(repoPath string) ([]string, []string, error) {
	dockerignorePath := filepath.Join(repoPath, ".dockerignore")

	// 1. Check if .dockerignore exists
	file, err := os.Open(dockerignorePath)
	if err != nil {
		if os.IsNotExist(err) {
			// No .dockerignore file exists
			return nil, nil, nil
		}
		return nil, nil, fmt.Errorf("error opening .dockerignore: %w", err)
	}
	defer file.Close()

	// 2. Read and parse the .dockerignore file
	patterns, err := ignorefile.ReadAll(file)
	if err != nil {
		return nil, nil, fmt.Errorf("error parsing .dockerignore: %w", err)
	}

	// 3. Separate exclude and include patterns
	excludePatterns, includePatterns := separatePatterns(patterns)

	return excludePatterns, includePatterns, nil
}

// separatePatterns separates patterns into exclude and include lists
// Include patterns are those starting with '!' (negation)
func separatePatterns(patterns []string) (excludes []string, includes []string) {
	for _, pattern := range patterns {
		if len(pattern) > 0 && pattern[0] == '!' {
			// Remove the '!' prefix for include patterns
			includes = append(includes, pattern[1:])
		} else {
			excludes = append(excludes, pattern)
		}
	}
	return excludes, includes
}

// DockerignoreContext holds parsed dockerignore information with caching
type DockerignoreContext struct {
	parsed   bool
	excludes []string
	includes []string
	repoPath string
}

// NewDockerignoreContext creates a new DockerignoreContext for the given repository path
func NewDockerignoreContext(repoPath string) *DockerignoreContext {
	return &DockerignoreContext{
		repoPath: repoPath,
	}
}

// Parse parses the .dockerignore file and caches the results
func (d *DockerignoreContext) Parse() ([]string, []string, error) {
	if !d.parsed {
		excludes, includes, err := CheckAndParseDockerignore(d.repoPath)
		if err != nil {
			return nil, nil, err
		}

		d.excludes = excludes
		d.includes = includes
		d.parsed = true
	}

	return d.excludes, d.includes, nil
}

// ParseWithLogging parses the .dockerignore file, caches the results, and logs when found
func (d *DockerignoreContext) ParseWithLogging(logger interface{ LogInfo(string, ...interface{}) }) ([]string, []string, error) {
	excludes, includes, err := d.Parse()
	if err != nil {
		return nil, nil, err
	}

	if excludes != nil || includes != nil {
		logger.LogInfo("Found .dockerignore file, applying filters")
	}

	return excludes, includes, nil
}

// HasDockerignoreFile checks if a .dockerignore file exists in the given directory
func HasDockerignoreFile(repoPath string) bool {
	dockerignorePath := filepath.Join(repoPath, ".dockerignore")
	_, err := os.Stat(dockerignorePath)
	return !os.IsNotExist(err)
}
