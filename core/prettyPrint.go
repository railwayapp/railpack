package core

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/railwayapp/railpack/core/logger"
	"github.com/railwayapp/railpack/core/plan"
	"github.com/railwayapp/railpack/core/resolver"
	"github.com/railwayapp/railpack/internal/utils"
)

const (
	AnsiRed           = "1"
	AnsiYellow        = "3"
	AnsiBlue          = "4"
	AnsiMagenta       = "5"
	AnsiCyan          = "6"
	AnsiWhite         = "7"
	AnsiGray          = "8"
	AnsiBrightBlue    = "12"
	AnsiBrightCyan    = "14"
	AnsiBrightWhite   = "15"
	AnsiBrightMagenta = "13"
	AnsiDarkGray      = "238"
	AnsiMediumGray    = "245"
)

var (
	highlight = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}

	headerStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(highlight).
			MarginTop(1).
			Padding(0, 1)

	sectionHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Width(10).
				MarginLeft(2).
				MarginTop(2).
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(AnsiDarkGray)).
				BorderBottom(true)

	packageNameStyle = lipgloss.NewStyle().
				MarginLeft(2).
				Foreground(lipgloss.Color(AnsiBrightMagenta))

	versionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(AnsiBrightCyan))

	sourceStyle = lipgloss.NewStyle()

	separatorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(AnsiDarkGray)).
			Margin(0, 2)

	indentedStepHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(AnsiBrightMagenta)).
				MarginLeft(2)

	commandPrefixStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(AnsiMediumGray)).
				MarginLeft(4)

	commandStyle = lipgloss.NewStyle().
			Bold(true)

	logInfoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(AnsiBrightWhite)).
			MarginLeft(2)

	logWarnStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(AnsiYellow)).
			MarginLeft(2)

	logErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(AnsiRed)).
			Bold(true).
			MarginLeft(2)

	metadataStyle = lipgloss.NewStyle().
			MarginLeft(2)

	metadataSeparatorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(AnsiMediumGray)).
				MarginRight(1)

	metadataValueStyle = lipgloss.NewStyle().
				Bold(true)
)

type PrintOptions struct {
	Metadata bool
	Version  string
}

func PrettyPrintBuildResult(buildResult *BuildResult, options ...PrintOptions) {
	output := FormatBuildResult(buildResult, options...)
	fmt.Print(output)
}

func FormatBuildResult(br *BuildResult, options ...PrintOptions) string {
	var opts PrintOptions
	if len(options) > 0 {
		opts = options[0]
	}
	var output strings.Builder

	formatHeader(&output, opts.Version)
	formatLogs(&output, br.Logs)
	formatPackages(&output, br.ResolvedPackages)
	formatSteps(&output, br)
	formatDeploy(&output, br)
	formatMetadata(&output, br.Metadata, opts.Metadata)

	output.WriteString("\n\n")
	return output.String()
}

func formatHeader(output *strings.Builder, version string) {
	header := fmt.Sprintf("Railpack %s", version)
	output.WriteString(headerStyle.Render(header))
	output.WriteString("\n")
}

func formatLogs(output *strings.Builder, logs []logger.Msg) {
	if len(logs) == 0 {
		return
	}

	output.WriteString("\n")

	for _, log := range logs {
		msg := utils.CapitalizeFirst(log.Msg)

		switch log.Level {
		case logger.Info:
			output.WriteString(logInfoStyle.Render(fmt.Sprintf("↳ %s", msg)))
		case logger.Warn:
			output.WriteString(logWarnStyle.Render(fmt.Sprintf("⚠ %s", msg)))
		case logger.Error:
			lines := strings.Split(msg, "\n")
			for i, line := range lines {
				if i == 0 {
					output.WriteString(logErrorStyle.Render(fmt.Sprintf("✖ %s", line)))
				} else {
					output.WriteString(fmt.Sprintf("  %s", line))
				}
				if i < len(lines)-1 {
					output.WriteString("\n")
				}
			}
		default:
			output.WriteString(logInfoStyle.Render(fmt.Sprintf("• %s", msg)))
		}
		output.WriteString("\n")
	}
}

func formatPackages(output *strings.Builder, packages map[string]*resolver.ResolvedPackage) {
	if len(packages) == 0 {
		return
	}

	output.WriteString(sectionHeaderStyle.MarginTop(1).Render("Packages"))
	output.WriteString("\n")

	nameWidth, versionWidth := 1, 1
	for _, pkg := range packages {
		nameWidth = max(nameWidth, len(pkg.Name))
		if pkg.ResolvedVersion != nil {
			versionWidth = max(versionWidth, len(*pkg.ResolvedVersion))
		}
	}

	localPackageNameStyle := packageNameStyle.Width(nameWidth).MaxWidth(20)
	localVersionStyle := versionStyle.Width(versionWidth).MaxWidth(20)
	separator := separatorStyle.Render("│")

	for _, pkg := range packages {
		name := localPackageNameStyle.Render(pkg.Name)

		version := "-"
		if pkg.ResolvedVersion != nil {
			version = *pkg.ResolvedVersion
		}
		version = localVersionStyle.Render(version)
		source := sourceStyle.Render(formatSource(pkg))
		output.WriteString(fmt.Sprintf("%s%s%s%s%s", name, separator, version, separator, source))
		output.WriteString("\n")
	}
}

func formatSteps(output *strings.Builder, br *BuildResult) {
	stepsToPrint := getStepsToPrint(br)
	if len(stepsToPrint) == 0 {
		return
	}

	output.WriteString(sectionHeaderStyle.MarginTop(1).Render("Steps"))
	output.WriteString("\n")

	for i, step := range stepsToPrint {
		commands := getCommandsToPrint(step.Commands)
		if len(commands) == 0 {
			continue
		}

		currentStepStyle := indentedStepHeaderStyle
		if i > 0 {
			currentStepStyle = currentStepStyle.MarginTop(1)
		}

		output.WriteString(currentStepStyle.Render(fmt.Sprintf("▸ %s", step.Name)))
		output.WriteString("\n")

		for _, cmd := range commands {
			cmdText := cmd.Cmd
			if cmd.CustomName != "" {
				cmdText = cmd.CustomName
			}
			output.WriteString(fmt.Sprintf("%s %s", commandPrefixStyle.Render("$"), commandStyle.Render(cmdText)))
			output.WriteString("\n")
		}
	}
}

func formatDeploy(output *strings.Builder, br *BuildResult) {
	if br.Plan != nil && br.Plan.Deploy.StartCmd != "" {
		output.WriteString(sectionHeaderStyle.MarginTop(1).Render("Deploy"))
		output.WriteString("\n")
		output.WriteString(fmt.Sprintf("%s %s", commandPrefixStyle.Render("$"), commandStyle.Render(br.Plan.Deploy.StartCmd)))
	}
}

func formatMetadata(output *strings.Builder, metadata map[string]string, showMetadata bool) {
	if !showMetadata || metadata == nil || len(metadata) == 0 {
		return
	}

	output.WriteString(sectionHeaderStyle.MarginTop(2).Render("Metadata"))
	output.WriteString("\n")

	separator := metadataSeparatorStyle.Render(":")

	for key, value := range metadata {
		output.WriteString(metadataStyle.Render(fmt.Sprintf("%s%s%s", key, separator, metadataValueStyle.Render(value))))
		output.WriteString("\n")
	}
}

func getStepsToPrint(br *BuildResult) []*plan.Step {
	execSteps := []*plan.Step{}
	if br.Plan == nil {
		return execSteps
	}

	for _, step := range br.Plan.Steps {
		if !strings.HasPrefix(step.Name, "packages") && step.Commands != nil {
			commands := getCommandsToPrint(step.Commands)
			if len(commands) > 0 {
				execSteps = append(execSteps, &step)
			}
		}
	}
	return execSteps
}

// TODO we should really just mark a command as skippable on the ExecCommand instead
var skippableCommands = []string{
	"mkdir -p /app/node_modules/.cache",
	"mkdir -p config deps _build",
}

func getCommandsToPrint(commands []plan.Command) []plan.ExecCommand {
	if commands == nil {
		return []plan.ExecCommand{}
	}

	execCommands := []plan.ExecCommand{}
	for _, cmd := range commands {
		if execCmd, ok := cmd.(plan.ExecCommand); ok {
			if !isSkippableCommand(execCmd.Cmd) {
				execCommands = append(execCommands, execCmd)
			}
		}
	}
	return execCommands
}

func isSkippableCommand(cmd string) bool {
	for _, skippable := range skippableCommands {
		if cmd == skippable {
			return true
		}
	}
	return false
}

func formatSource(pkg *resolver.ResolvedPackage) string {
	if pkg.RequestedVersion != nil {
		return fmt.Sprintf("%s (%s)", pkg.Source, *pkg.RequestedVersion)
	}
	return pkg.Source
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
