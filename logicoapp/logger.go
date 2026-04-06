package logicoapp

import (
	"context"

	general_logger "github.com/gnemade360/go-logger/logicologger/general-logger"
	logicologger "github.com/gnemade360/go-logger/logicologger"
)

type Logger interface {
	Info(msg string, args ...any)
	Error(msg string, args ...any)
	Debug(msg string, args ...any)
}

func newDefaultLogger(name string) Logger {
	return general_logger.NewGeneralLogger(context.Background(), name, nil)
}

func NewNoopLogger() logicologger.Logger {
	return logicologger.NewNoopLogger()
}
