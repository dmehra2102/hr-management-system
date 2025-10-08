package logger

import (
	"context"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

type Logger struct {
	logger *logrus.Logger
	entry  *logrus.Entry
}

func NewLogger(level, format string) *Logger {
	logger := logrus.New()

	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		logLevel = logrus.InfoLevel
	}
	logger.SetLevel(logLevel)

	switch strings.ToLower(format) {
	case "json":
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "timestamp",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "message",
				logrus.FieldKeyFunc:  "caller",
			},
		})
	default:
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
		})
	}

	logger.SetOutput(os.Stdout)
	logger.SetReportCaller(true)

	return &Logger{
		logger: logger,
		entry:  logger.WithFields(logrus.Fields{}),
	}
}

func (l *Logger) WithFields(fields map[string]any) *Logger {
	return &Logger{
		logger: l.logger,
		entry:  l.entry.WithFields(fields),
	}
}

func (l *Logger) WithField(key string, value any) *Logger {
	return &Logger{
		logger: l.logger,
		entry:  l.entry.WithField(key, value),
	}
}

func (l *Logger) WithContext(ctx context.Context) *Logger {
	return &Logger{
		logger: l.logger,
		entry:  l.entry.WithContext(ctx),
	}
}

func (l *Logger) WithError(err error) *Logger {
	return &Logger{
		logger: l.logger,
		entry:  l.entry.WithError(err),
	}
}

func (l *Logger) Debug(msg string, keysAndValues ...any) {
	l.entry.WithFields(l.parseKeysAndValues(keysAndValues...)).Debug(msg)
}

func (l *Logger) Info(msg string, keysAndValues ...any) {
	l.entry.WithFields(l.parseKeysAndValues(keysAndValues...)).Info(msg)
}

func (l *Logger) Warn(msg string, keysAndValues ...any) {
	l.entry.WithFields(l.parseKeysAndValues(keysAndValues...)).Warn(msg)
}

func (l *Logger) Error(msg string, keysAndValues ...any) {
	l.entry.WithFields(l.parseKeysAndValues(keysAndValues...)).Error(msg)
}

func (l *Logger) Fatal(msg string, keysAndValues ...any) {
	l.entry.WithFields(l.parseKeysAndValues(keysAndValues...)).Fatal(msg)
}

func (l *Logger) Panic(msg string, keysAndValues ...any) {
	l.entry.WithFields(l.parseKeysAndValues(keysAndValues...)).Panic(msg)
}

func (l *Logger) GetLogrusLogger() *logrus.Logger {
	return l.logger
}

func (l *Logger) GetLogrusEntry() *logrus.Entry {
	return l.entry
}

func (l *Logger) parseKeysAndValues(keysAndValues ...any) logrus.Fields {
	fields := logrus.Fields{}

	if len(keysAndValues)%2 != 0 {
		keysAndValues = append(keysAndValues, "MISSING_VALUE")
	}

	for i := 0; i < len(keysAndValues); i += 2 {
		key, ok := keysAndValues[i].(string)
		if !ok {
			key = "INVALID_KEY"
		}
		value := keysAndValues[i+1]
		fields[key] = value
	}

	return fields
}

func (l *Logger) RequestLogger(requestID, method, path string) *Logger {
	return l.WithFields(map[string]any{
		"request_id": requestID,
		"method":     method,
		"path":       path,
		"component":  "request",
	})
}

func (l *Logger) ServiceLogger(service string) *Logger {
	return l.WithFields(map[string]any{
		"service":   service,
		"component": "service",
	})
}

func (l *Logger) RepositoryLogger(repository string) *Logger {
	return l.WithFields(map[string]any{
		"repository": repository,
		"component":  "repository",
	})
}

func (l *Logger) HandlerLogger(handler string) *Logger {
	return l.WithFields(map[string]any{
		"handler":   handler,
		"component": "handler",
	})
}
