// this provider is unique: it used solely to extract a start command
package procfile

import (
	"regexp"
	"strings"

	"github.com/railwayapp/railpack/core/generate"
)

type ProcfileProvider struct{}

func (p *ProcfileProvider) Name() string {
	return "procfile"
}

// process type names are limited to letters, digits, underscores and dashes; everything after the colon is the command
var procfileLineRegex = regexp.MustCompile(`^([A-Za-z0-9_-]+):\s*(.+)$`)

type procfileEntry struct {
	processType string
	command     string
}

func (p *ProcfileProvider) Plan(ctx *generate.GenerateContext) (bool, error) {
	contents, err := ctx.App.ReadFile("Procfile")
	if err != nil {
		return false, nil
	}

	// A Procfile is not YAML: commands can contain ": " or start with a quote, so parse it line by line
	entries := parseProcfile(contents)
	if len(entries) == 0 {
		return false, nil
	}

	webCommand := ""
	workerCommand := ""
	for _, entry := range entries {
		switch entry.processType {
		case "web":
			webCommand = entry.command
		case "worker":
			workerCommand = entry.command
		}
	}

	if webCommand != "" {
		ctx.Logger.LogInfo("Found web command in Procfile")
		ctx.Deploy.StartCmd = webCommand
		return false, nil
	}

	if workerCommand != "" {
		ctx.Logger.LogInfo("Found worker command in Procfile")
		ctx.Deploy.StartCmd = workerCommand
		return false, nil
	}

	// use the first process in file order so the result is deterministic
	ctx.Logger.LogInfo("Found %s command in Procfile", entries[0].processType)
	ctx.Deploy.StartCmd = entries[0].command

	return false, nil
}

// Returns a slice rather than a map to preserve file order for deterministic
// fallback process selection (Go map iteration order is randomized).
func parseProcfile(contents string) []procfileEntry {
	entries := []procfileEntry{}

	for _, line := range strings.Split(contents, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		matches := procfileLineRegex.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		entries = append(entries, procfileEntry{processType: matches[1], command: matches[2]})
	}

	return entries
}
