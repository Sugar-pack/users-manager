// Package logging is a local wrapper around logrus package
// if, suddenly, logrus will become deprecated module
// it will be easy to move to another solution
package logging

import (
	"context"
	"os"

	log "github.com/sirupsen/logrus"
)

// Fields is a wrapper around logrus.Field
type Fields log.Fields

// logWrapper is a wrapper around *logrus.Entry
type logWrapper struct {
	*log.Entry
}

// Logger interface
type Logger interface {
	Trace(...interface{})
	Debug(...interface{})
	Info(...interface{})
	Warn(...interface{})
	Error(...interface{})
	Fatal(...interface{})
	WithError(err error) Logger
	WithField(key string, value interface{}) Logger
	WithFields(fields Fields) Logger
}

// GetLogger is a Logger getter with default settings
func GetLogger() Logger {
	logger := log.New()
	logger.SetFormatter(&log.TextFormatter{})
	logger.SetLevel(log.TraceLevel)
	logger.SetOutput(os.Stdout)
	logger.AddHook(GetFileLineHook())
	l := &logWrapper{logger.WithFields(nil)}

	return l
}

func (lw *logWrapper) WithError(err error) Logger {
	return &logWrapper{lw.Entry.WithError(err)}
}

func (lw *logWrapper) WithField(key string, value interface{}) Logger {
	return &logWrapper{lw.Entry.WithField(key, value)}
}

func (lw *logWrapper) WithFields(fields Fields) Logger {
	return &logWrapper{lw.Entry.WithFields(log.Fields(fields))}
}

type logCtx struct{}

// WithContext puts logger to the context
func WithContext(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, logCtx{}, logger)
}

// FromContext extracts Logger from context and returns itself
// otherwise, creates default one logger
func FromContext(ctx context.Context) Logger {
	logger, ok := ctx.Value(logCtx{}).(Logger)
	if !ok {
		return GetLogger()
	}
	return logger
}
