package logger

import (
	"fmt"
	"net/url"
	"strings"
)

type Level string

const (
	Info        Level = "info"
	Warn        Level = "warn"
	Deprecation Level = "deprecation"
	Suggestion  Level = "suggestion"
	Error       Level = "error"

	DocsBaseURL = "https://railpack.com"
)

type Msg struct {
	Level    Level
	Msg      string
	DocsPath string // optional Railpack-relative path or absolute URL
}

type Logger struct {
	Logs []Msg
}

func NewLogger() *Logger {
	return &Logger{
		Logs: []Msg{},
	}
}

func (l *Logger) LogInfo(format string, args ...any) {
	l.log(Info, format, args...)
}

func (l *Logger) LogWarn(format string, args ...any) {
	l.log(Warn, format, args...)
}

func (l *Logger) LogDeprecation(format string, args ...any) {
	l.log(Deprecation, format, args...)
}

// LogSuggestion records a helpful config suggestion with an optional docs link.
// Relative paths resolve against railpack.com; absolute URLs are used unchanged.
func (l *Logger) LogSuggestion(msg string, docsPath ...string) {
	path := ""
	if len(docsPath) > 0 {
		path = docsPath[0]
	}
	l.Logs = append(l.Logs, Msg{
		Level:    Suggestion,
		Msg:      msg,
		DocsPath: path,
	})
}

func (l *Logger) LogError(format string, args ...any) {
	l.log(Error, format, args...)
}

func (l *Logger) log(level Level, format string, args ...any) {
	msg := format
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	}
	l.Logs = append(l.Logs, Msg{
		Level: level,
		Msg:   msg,
	})
}

// DocsURL resolves Railpack-relative docs paths while preserving absolute URLs.
func DocsURL(docsPath string) string {
	if docsPath == "" {
		return DocsBaseURL
	}

	parsedURL, err := url.Parse(docsPath)
	if err == nil && parsedURL.IsAbs() {
		return docsPath
	}

	if !strings.HasPrefix(docsPath, "/") {
		docsPath = "/" + docsPath
	}
	return DocsBaseURL + docsPath
}
