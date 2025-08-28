package app

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/tailscale/hujson"
	"gopkg.in/yaml.v2"
)

type App struct {
	Source       string
	IsRemote     bool
	RemoteURL    string
	GitHubClient *GitHubClient
}

type GitHubClient struct {
	Owner  string
	Repo   string
	Branch string
	Token  string
	Cache  map[string]interface{}
}

type GitHubTree struct {
	SHA  string `json:"sha"`
	Tree []struct {
		Path string `json:"path"`
		Mode string `json:"mode"`
		Type string `json:"type"`
		Size int    `json:"size"`
		SHA  string `json:"sha"`
	} `json:"tree"`
	Truncated bool `json:"truncated"`
}

type GitHubContent struct {
	Type     string `json:"type"`
	Encoding string `json:"encoding"`
	Size     int    `json:"size"`
	Content  string `json:"content"`
}

func NewApp(path string) (*App, error) {
	if isGitHubURL(path) {
		return NewAppFromGitHub(path)
	}

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

	return &App{Source: source}, nil
}

// findMatches returns a list of paths matching a glob pattern, filtered by isDir
func (a *App) findMatches(pattern string, isDir bool) ([]string, error) {
	if a.IsRemote && a.GitHubClient != nil {
		return a.findMatchesGitHub(pattern, isDir)
	}

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

		if info.IsDir() == isDir {
			paths = append(paths, match)
		}
	}
	return paths, nil
}

func (a *App) findMatchesGitHub(pattern string, isDir bool) ([]string, error) {
	tree, err := a.GitHubClient.getTree()
	if err != nil {
		return nil, err
	}

	var paths []string
	for _, item := range tree.Tree {
		matched, err := doublestar.Match(pattern, item.Path)
		if err != nil {
			continue
		}

		if matched {
			if (item.Type == "tree") == isDir {
				paths = append(paths, item.Path)
			}
		}
	}
	return paths, nil
}

// FindFiles returns a list of file paths matching a glob pattern
func (a *App) FindFiles(pattern string) ([]string, error) {
	return a.findMatches(pattern, false)
}

// FindDirectories returns a list of directory paths matching a glob pattern
func (a *App) FindDirectories(pattern string) ([]string, error) {
	return a.findMatches(pattern, true)
}

// findGlob finds paths matching a glob pattern
func (a *App) findGlob(pattern string) ([]string, error) {
	if a.IsRemote && a.GitHubClient != nil {
		tree, err := a.GitHubClient.getTree()
		if err != nil {
			return nil, err
		}

		var matches []string
		for _, item := range tree.Tree {
			matched, err := doublestar.Match(pattern, item.Path)
			if err != nil {
				continue
			}
			if matched {
				matches = append(matches, item.Path)
			}
		}
		return matches, nil
	}

	matches, err := doublestar.Glob(os.DirFS(a.Source), pattern)

	if err != nil {
		return nil, err
	}

	return matches, nil
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

// ReadFile reads the contents of a file
func (a *App) ReadFile(name string) (string, error) {
	if a.IsRemote && a.GitHubClient != nil {
		content, err := a.GitHubClient.getFileContent(name)
		if err != nil {
			return "", fmt.Errorf("error reading %s: %w", name, err)
		}
		return strings.ReplaceAll(content, "\r\n", "\n"), nil
	}

	path := filepath.Join(a.Source, name)
	data, err := os.ReadFile(path)
	if err != nil {
		relativePath, _ := a.stripSourcePath(path)
		return "", fmt.Errorf("error reading %s: %w", relativePath, err)
	}

	return strings.ReplaceAll(string(data), "\r\n", "\n"), nil
}

// ReadJSON reads and parses a JSON file
func (a *App) ReadJSON(name string, v interface{}) error {
	data, err := a.ReadFile(name)
	if err != nil {
		return err
	}

	jsonBytes, err := standardizeJSON([]byte(data))
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
func (a *App) ReadYAML(name string, v interface{}) error {
	data, err := a.ReadFile(name)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal([]byte(data), v); err != nil {
		return fmt.Errorf("error reading %s as YAML: %w", name, err)
	}

	return nil
}

// ReadTOML reads and parses a TOML file
func (a *App) ReadTOML(name string, v interface{}) error {
	data, err := a.ReadFile(name)
	if err != nil {
		return err
	}

	return toml.Unmarshal([]byte(data), v)
}

// IsFileExecutable checks if a path is an executable file
func (a *App) IsFileExecutable(name string) bool {
	if a.IsRemote && a.GitHubClient != nil {
		tree, err := a.GitHubClient.getTree()
		if err != nil {
			return false
		}

		for _, item := range tree.Tree {
			if item.Path == name && item.Type == "blob" {
				return item.Mode == "100755"
			}
		}
		return false
	}

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

// StripSourcePath converts an absolute path to a path relative to the app source directory
func (a *App) stripSourcePath(absPath string) (string, error) {
	rel, err := filepath.Rel(a.Source, absPath)
	if err != nil {
		return "", errors.New("failed to parse source path")
	}
	return rel, nil
}

func standardizeJSON(b []byte) ([]byte, error) {
	ast, err := hujson.Parse(b)
	if err != nil {
		return b, err
	}
	ast.Standardize()
	return ast.Pack(), nil
}

func isGitHubURL(path string) bool {
	if strings.HasPrefix(path, "github.com/") ||
		strings.HasPrefix(path, "https://github.com/") ||
		strings.HasPrefix(path, "http://github.com/") {
		return true
	}
	return false
}

func NewAppFromGitHub(githubURL string) (*App, error) {
	if !strings.HasPrefix(githubURL, "https://") && !strings.HasPrefix(githubURL, "http://") {
		githubURL = "https://" + githubURL
	}

	parsedURL, err := url.Parse(githubURL)
	if err != nil {
		return nil, fmt.Errorf("invalid GitHub URL: %w", err)
	}

	if parsedURL.Host != "github.com" {
		return nil, fmt.Errorf("not a GitHub URL: %s", githubURL)
	}

	pathParts := strings.Split(strings.TrimPrefix(parsedURL.Path, "/"), "/")
	if len(pathParts) < 2 {
		return nil, fmt.Errorf("invalid GitHub repository URL format")
	}

	owner := pathParts[0]
	repo := strings.TrimSuffix(pathParts[1], ".git")
	branch := "main"

	if len(pathParts) >= 4 && pathParts[2] == "tree" {
		branch = pathParts[3]
	}

	client := &GitHubClient{
		Owner:  owner,
		Repo:   repo,
		Branch: branch,
		Token:  os.Getenv("GITHUB_TOKEN"),
		Cache:  make(map[string]interface{}),
	}

	return &App{
		Source:       fmt.Sprintf("github.com/%s/%s", owner, repo),
		IsRemote:     true,
		RemoteURL:    githubURL,
		GitHubClient: client,
	}, nil
}

func (g *GitHubClient) makeRequest(endpoint string) ([]byte, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/%s", g.Owner, g.Repo, endpoint)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if g.Token != "" {
		req.Header.Set("Authorization", "token "+g.Token)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error: %s (status %d)\n%s", url, resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}

func (g *GitHubClient) getTree() (*GitHubTree, error) {
	if cached, ok := g.Cache["tree"]; ok {
		return cached.(*GitHubTree), nil
	}

	data, err := g.makeRequest(fmt.Sprintf("git/trees/%s?recursive=1", g.Branch))
	if err != nil {
		data, err = g.makeRequest("git/trees/master?recursive=1")
		if err != nil {
			return nil, fmt.Errorf("failed to get repository tree: %w", err)
		}
	}

	var tree GitHubTree
	if err := json.Unmarshal(data, &tree); err != nil {
		return nil, err
	}

	g.Cache["tree"] = &tree
	return &tree, nil
}

func (g *GitHubClient) getFileContent(path string) (string, error) {
	cacheKey := "file:" + path
	if cached, ok := g.Cache[cacheKey]; ok {
		return cached.(string), nil
	}

	data, err := g.makeRequest(fmt.Sprintf("contents/%s?ref=%s", path, g.Branch))
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			data, err = g.makeRequest(fmt.Sprintf("contents/%s?ref=master", path))
			if err != nil {
				return "", err
			}
		} else {
			return "", err
		}
	}

	var content GitHubContent
	if err := json.Unmarshal(data, &content); err != nil {
		return "", err
	}

	if content.Type != "file" {
		return "", fmt.Errorf("%s is not a file", path)
	}

	if content.Encoding == "base64" {
		decoded, err := base64.StdEncoding.DecodeString(content.Content)
		if err != nil {
			return "", err
		}
		result := string(decoded)
		g.Cache[cacheKey] = result
		return result, nil
	}

	g.Cache[cacheKey] = content.Content
	return content.Content, nil
}
