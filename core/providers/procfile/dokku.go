package procfile

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"
)

// ProcfileEntry represents a process type and command extracted from a Procfile.
type ProcfileEntry struct {
	Name    string
	Command string
}

var (
	// Matches process names with letters, digits, underscores, and dashes, followed by colon.
	procfileLineRegex = regexp.MustCompile(`^([A-Za-z0-9_-]+)\s*:\s*(.+)$`)

	// Matches inline comments preceded by whitespace.
	commentRegex             = regexp.MustCompile(`^(.*)\s#.+$`)
	forwardslashCommentRegex = regexp.MustCompile(`^(.*)\s//.+$`)
)

// ParseProcfile parses a Procfile using Dokku's procfile-util parser algorithm:
// https://github.com/dokku/procfile-util/blob/master/pkg/procfile/procfile.go
func ParseProcfile(text string) ([]ProcfileEntry, error) {
	var entries []ProcfileEntry
	names := make(map[string]bool)
	scanner := bufio.NewScanner(strings.NewReader(text))
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if len(line) == 0 {
			continue
		}

		if strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}

		params := procfileLineRegex.FindStringSubmatch(line)
		if len(params) != 3 {
			return nil, fmt.Errorf("invalid line in Procfile, line %d", lineNumber)
		}

		name := params[1]
		cmd := params[2]

		if len(name) > 63 {
			return nil, fmt.Errorf("process name over 63 characters, line %d", lineNumber)
		}

		if names[name] {
			return nil, fmt.Errorf("process names must be unique, line %d", lineNumber)
		}
		names[name] = true

		if commentParams := commentRegex.FindStringSubmatch(cmd); len(commentParams) == 2 {
			cmd = commentParams[1]
		} else if forwardslashParams := forwardslashCommentRegex.FindStringSubmatch(cmd); len(forwardslashParams) == 2 {
			cmd = forwardslashParams[1]
		}

		cmd = strings.TrimSpace(cmd)
		if strings.HasPrefix(cmd, "#") || strings.HasPrefix(cmd, "//") {
			return nil, fmt.Errorf("comment specified in place of command, line %d", lineNumber)
		}

		if len(cmd) == 0 {
			return nil, fmt.Errorf("no command specified, line %d", lineNumber)
		}

		entries = append(entries, ProcfileEntry{Name: name, Command: cmd})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading Procfile: %w", err)
	}

	return entries, nil
}
