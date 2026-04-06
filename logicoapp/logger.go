package logicoapp

import "log"

type Logger interface {
	Info(msg string, args ...any)
	Error(msg string, args ...any)
	Debug(msg string, args ...any)
}

type defaultLogger struct{}

func (defaultLogger) Info(msg string, args ...any)  { log.Printf("[INFO]  "+msg, args...) }
func (defaultLogger) Error(msg string, args ...any) { log.Printf("[ERROR] "+msg, args...) }
func (defaultLogger) Debug(msg string, args ...any) { log.Printf("[DEBUG] "+msg, args...) }
