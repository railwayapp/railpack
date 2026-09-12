package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/charmbracelet/log"
	"github.com/moby/patternmatcher"
	"github.com/railwayapp/railpack/internal/utils"
	"gopkg.in/yaml.v2"
)

var ErrNoFileFound = errors.New("unable to find a matching file")

type App struct {
	Source    string
	globCache map[string][]string

	// excludeMatcher filters out paths that will not exist in the build
	// context. Nil until SetExcludePatterns is called, which means an App
	// created on its own sees the source directory unfiltered.
	excludeMatcher *patternmatcher.PatternMatcher
}

func NewApp(path string) (*App, error) {
	var source string

	if filepath.IsAbs(path) {
		source = path
	} else {
		currentDir, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		source, err = filepath.Abs(filepath.Join(currentDir, path))
		if err != nil {
			return nil, errors.New("failed to read app source directory")
		}
	}

	if _, err := os.Stat(source); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("directory %s does not exist", source)
		}
		return nil, fmt.Errorf("failed to check directory %s: %w", source, err)
	}

	return &App{
		Source:    source,
		globCache: make(map[string][]string),
	}, nil
}

// SetExcludePatterns compiles the patterns that keep files out of the build
// context, so that path discovery cannot return a file BuildKit will never
// load.
//
// The planner reads the source directory directly, while BuildKit only ever
// sees the filtered context. Without this, a provider can glob up a file that
// is excluded and emit a COPY for it; the build then fails late with
// "failed to compute cache key: ... not found".
//
// Patterns come from .dockerignore and railpack.json's `exclude`. Negations
// are honoured, since this is the same matcher BuildKit builds the context
// with. Passing no patterns clears the filter.
func (a *App) SetExcludePatterns(patterns []string) error {
	if len(patterns) == 0 {
		a.excludeMatcher = nil
		return nil
	}

	matcher, err := patternmatcher.New(patterns)
	if err != nil {
		return fmt.Errorf("failed to compile exclude patterns: %w", err)
	}

	a.excludeMatcher = matcher
	return nil
}

// isExcluded reports whether a path is kept out of the build context.
func (a *App) isExcluded(path string) bool {
	if a.excludeMatcher == nil {
		return false
	}

	excluded, err := a.excludeMatcher.MatchesOrParentMatches(path)
	if err != nil {
		// Keep the path rather than silently dropping it; a build that copies
		// one file too many is easier to debug than one missing a file.
		log.Warnf("Failed to match %q against exclude patterns: %s", path, err)
		return false
	}

	if excluded {
		log.Debugf("Skipping %q, excluded from the build context", path)
	}

	return excluded
}

// findMatches returns a list of paths matching a glob pattern, filtered by isDir
func (a *App) findMatches(pattern string, isDir bool) ([]string, error) {
	matches, err := a.findGlob(pattern)

	if err != nil {
		return nil, err
	}

	var paths []string
	for _, match := range matches {
		fullPath := filepath.Join(a.Source, match)

		info, err := os.Stat(fullPath)
		if err != nil {
			continue
		}

		if info.IsDir() != isDir {
			continue
		}

		if a.isExcluded(match) {
			continue
		}

		paths = append(paths, match)
	}
	return paths, nil
}

// returns a list of file paths matching a glob pattern
func (a *App) FindFiles(pattern string) ([]string, error) {
	return a.findMatches(pattern, false)
}

// FindDirectories returns a list of directory paths matching a glob pattern
func (a *App) FindDirectories(pattern string) ([]string, error) {
	return a.findMatches(pattern, true)
}

// findGlob finds paths matching a glob pattern, with caching
func (a *App) findGlob(pattern string) ([]string, error) {
	if cached, ok := a.globCache[pattern]; ok {
		return cached, nil
	}

	matches, err := doublestar.Glob(os.DirFS(a.Source), pattern)
	if err != nil {
		return nil, err
	}

	a.globCache[pattern] = matches
	return matches, nil
}

// Check if a relative file exists in the app's source directory
func (a *App) HasFile(path string) bool {
	fullPath := filepath.Join(a.Source, path)

	_, err := os.Stat(fullPath)
	return !os.IsNotExist(err)
}

// HasMatch checks if a path matching a glob exists (files or directories)
func (a *App) HasMatch(pattern string) bool {
	files, err := a.FindFiles(pattern)
	if err != nil {
		return false
	}

	dirs, err := a.FindDirectories(pattern)
	if err != nil {
		return false
	}

	return len(files) > 0 || len(dirs) > 0
}

func (a *App) FindFilesWithContent(pattern string, regex *regexp.Regexp) []string {
	files, err := a.FindFiles(pattern)
	if err != nil {
		return nil
	}

	var matches []string
	for _, file := range files {
		content, err := a.ReadFile(file)
		if err != nil {
			continue
		}

		if regex.MatchString(content) {
			matches = append(matches, file)
		}
	}

	return matches
}

// reads the contents of the first file that exists within the application source directory
// helpful for reading config from multiple possible locations (something.js, something.ts, etc)
func (a *App) ReadFirstFileOf(names ...string) (string, string, error) {
	for _, name := range names {
		if !a.HasFile(name) {
			continue
		}

		contents, err := a.ReadFile(name)
		if err != nil {
			return "", "", err
		}

		return name, contents, nil
	}

	return "", "", ErrNoFileFound
}

// ReadFile reads the contents of a file within the application source directory
func (a *App) ReadFile(name string) (string, error) {
	path := filepath.Join(a.Source, name)
	data, err := os.ReadFile(path)
	if err != nil {
		relativePath, _ := a.stripSourcePath(path)
		return "", fmt.Errorf("error reading %s: %w", relativePath, err)
	}

	return strings.ReplaceAll(string(data), "\r\n", "\n"), nil
}

// ReadJSON reads and parses a JSON file
func (a *App) ReadJSON(name string, v any) error {
	data, err := a.ReadFile(name)
	if err != nil {
		return err
	}

	jsonBytes, err := utils.StandardizeJSON([]byte(data))
	if err != nil {
		return err
	}

	data = string(jsonBytes)

	if err := json.Unmarshal([]byte(data), v); err != nil {
		relativePath, _ := a.stripSourcePath(filepath.Join(a.Source, name))
		return fmt.Errorf("error reading %s as JSON: %w", relativePath, err)
	}

	return nil
}

// ReadYAML reads and parses a YAML file
func (a *App) ReadYAML(name string, v any) error {
	data, err := a.ReadFile(name)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal([]byte(data), v); err != nil {
		return fmt.Errorf("error reading %s as YAML: %w", name, err)
	}

	return nil
}

func (a *App) ReadTOML(name string, v any) error {
	data, err := a.ReadFile(name)
	if err != nil {
		return err
	}

	return toml.Unmarshal([]byte(data), v)
}

// checks if a path is an executable file
func (a *App) IsFileExecutable(name string) bool {
	path := filepath.Join(a.Source, name)
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	if !info.Mode().IsRegular() {
		return false
	}

	// Check executable bit
	return info.Mode()&0111 != 0
}

// converts an absolute path to a path relative to the app source directory
func (a *App) stripSourcePath(absPath string) (string, error) {
	rel, err := filepath.Rel(a.Source, absPath)
	if err != nil {
		return "", errors.New("failed to parse source path")
	}
	return rel, nil
}
