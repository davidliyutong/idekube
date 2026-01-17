package config

import (
	"context"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm/logger"
)

// GormLogger wraps zap logger for GORM integration
type GormLogger struct {
	zapLogger *zap.Logger
}

// NewGormLogger creates a new GORM logger backed by zap
func NewGormLogger(zapLogger *zap.Logger) logger.Interface {
	return &GormLogger{zapLogger: zapLogger}
}

// LogMode implements logger.Interface
func (l *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	return l
}

// Info implements logger.Interface
func (l *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	l.zapLogger.Sugar().Infof(msg, data...)
}

// Warn implements logger.Interface
func (l *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	l.zapLogger.Sugar().Warnf(msg, data...)
}

// Error implements logger.Interface
func (l *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	l.zapLogger.Sugar().Errorf(msg, data...)
}

// Trace implements logger.Interface
func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()

	if err != nil {
		l.zapLogger.Error("SQL error",
			zap.String("sql", sql),
			zap.Int64("rows_affected", rows),
			zap.Duration("elapsed_ms", elapsed),
			zap.Error(err),
		)
	} else {
		l.zapLogger.Debug("SQL executed",
			zap.String("sql", sql),
			zap.Int64("rows_affected", rows),
			zap.Duration("elapsed_ms", elapsed),
		)
	}
}
