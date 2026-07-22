package logger

import "fmt"

type Level string

const (
	Info        Level = "info"
	Warn        Level = "warn"
	Deprecation Level = "deprecation"
	Error       Level = "error"
)

type Msg struct {
	Level Level
	Msg   string
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
