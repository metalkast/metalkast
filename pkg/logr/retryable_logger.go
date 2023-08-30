package logr

import (
	"github.com/go-logr/logr"
	"github.com/hashicorp/go-retryablehttp"
)

var _ retryablehttp.LeveledLogger = &RetryableLogger{}

type RetryableLogger struct {
	log logr.Logger
}

// Debug implements retryablehttp.LeveledLogger.
func (l *RetryableLogger) Debug(msg string, keysAndValues ...interface{}) {
	l.log.V(2).Info(msg, keysAndValues...)
}

// Error implements retryablehttp.LeveledLogger.
func (l *RetryableLogger) Error(msg string, keysAndValues ...interface{}) {
	var err error
	for i, k := range keysAndValues {
		if k == "error" && len(keysAndValues) > i+1 {
			err, _ = keysAndValues[i+1].(error)
		}
	}
	l.log.Error(err, msg, keysAndValues...)
}

// Info implements retryablehttp.LeveledLogger.
func (l *RetryableLogger) Info(msg string, keysAndValues ...interface{}) {
	l.log.Info(msg, keysAndValues...)
}

// Warn implements retryablehttp.LeveledLogger.
func (l *RetryableLogger) Warn(msg string, keysAndValues ...interface{}) {
	l.log.V(1).Info(msg, keysAndValues...)
}

func NewFromLogger(log logr.Logger) *RetryableLogger {
	return &RetryableLogger{
		log: log,
	}
}
