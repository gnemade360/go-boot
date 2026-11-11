package logicoapp

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	logicologger "github.com/gnemade360/go-logger/logicologger"
	general_logger "github.com/gnemade360/go-logger/logicologger/general-logger"
)

type Logger interface {
	Info(msg string, args ...any)
	Error(msg string, args ...any)
	Debug(msg string, args ...any)
}

const defaultLogLevel = logicologger.ERROR | logicologger.WARN | logicologger.INFO

var levelNames = map[logicologger.LogLevel]string{
	logicologger.ERROR: "ERROR",
	logicologger.WARN:  "WARN",
	logicologger.INFO:  "INFO",
	logicologger.DEBUG: "DEBUG",
}

func newDefaultLogger(name string) Logger {
	gl := general_logger.NewGeneralLogger(
		context.Background(),
		name,
		nil,
		general_logger.WithLogLevel(defaultLogLevel),
		general_logger.WithWriter(unclosableWriter{os.Stdout}),
		general_logger.WithFormatter(lineFormatter{}),
	)
	return &printfLogger{target: gl}
}

func NewNoopLogger() logicologger.Logger {
	return logicologger.NewNoopLogger()
}

// printfLogger formats printf-style arguments itself because go-logger v0.1.0
// stores args as event data instead of applying them to the message.
type printfLogger struct {
	target Logger
}

func (l *printfLogger) Info(msg string, args ...any)  { l.target.Info(format(msg, args)) }
func (l *printfLogger) Error(msg string, args ...any) { l.target.Error(format(msg, args)) }
func (l *printfLogger) Debug(msg string, args ...any) { l.target.Debug(format(msg, args)) }

func format(msg string, args []any) string {
	if len(args) == 0 {
		return msg
	}
	return fmt.Sprintf(msg, args...)
}

type lineFormatter struct{}

func (lineFormatter) Format(event logicologger.LogEvent) ([]byte, error) {
	line := fmt.Sprintf("%s %s %s: %s\n",
		event.Time.Format(time.RFC3339), levelNames[event.Level], event.Prefix, event.Message)
	return []byte(line), nil
}

// unclosableWriter keeps go-logger from closing os.Stdout: its writer closes
// all targets on Stop and from a GC finalizer.
type unclosableWriter struct {
	io.Writer
}

func (unclosableWriter) Close() error { return nil }
