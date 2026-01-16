package logger

import "go.uber.org/zap"

// Logger wraps zap.Logger to provide a simple logging interface
type Logger struct {
	zap *zap.Logger
}

// NewLogger creates a new logger from a zap logger
func NewLogger(zapLogger *zap.Logger) *Logger {
	return &Logger{zap: zapLogger}
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, fields ...interface{}) {
	if len(fields) > 0 {
		l.zap.Sugar().Debugw(msg, fields...)
	} else {
		l.zap.Sugar().Debug(msg)
	}
}

// Debugf logs a debug message with formatting
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.zap.Sugar().Debugf(format, args...)
}

// Info logs an info message
func (l *Logger) Info(msg string, fields ...interface{}) {
	if len(fields) > 0 {
		l.zap.Sugar().Infow(msg, fields...)
	} else {
		l.zap.Sugar().Info(msg)
	}
}

// Infof logs an info message with formatting
func (l *Logger) Infof(format string, args ...interface{}) {
	l.zap.Sugar().Infof(format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, fields ...interface{}) {
	if len(fields) > 0 {
		l.zap.Sugar().Warnw(msg, fields...)
	} else {
		l.zap.Sugar().Warn(msg)
	}
}

// Warnf logs a warning message with formatting
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.zap.Sugar().Warnf(format, args...)
}

// Error logs an error message
func (l *Logger) Error(msg string, fields ...interface{}) {
	if len(fields) > 0 {
		l.zap.Sugar().Errorw(msg, fields...)
	} else {
		l.zap.Sugar().Error(msg)
	}
}

// Errorf logs an error message with formatting
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.zap.Sugar().Errorf(format, args...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(msg string, fields ...interface{}) {
	if len(fields) > 0 {
		l.zap.Sugar().Fatalw(msg, fields...)
	} else {
		l.zap.Sugar().Fatal(msg)
	}
}

// Fatalf logs a fatal message with formatting and exits
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.zap.Sugar().Fatalf(format, args...)
}

// WithField returns a logger with an additional field
func (l *Logger) WithField(key string, value interface{}) *Logger {
	return &Logger{zap: l.zap.With(zap.Any(key, value))}
}

// WithFields returns a logger with additional fields
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}
	return &Logger{zap: l.zap.With(zapFields...)}
}
