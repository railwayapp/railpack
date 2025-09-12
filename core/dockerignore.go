package main

import (
    "fmt"
    "io"
    "os"
    "path/filepath"
    
    "github.com/moby/buildkit/client/llb"
    "github.com/moby/patternmatcher"
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

