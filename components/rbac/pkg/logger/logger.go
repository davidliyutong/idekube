package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	zap *zap.SugaredLogger
}

func NewLogger() *Logger {
	// Create encoder config
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	// Create core with console output
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(os.Stdout),
		zapcore.InfoLevel,
	)

	// Create logger
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	sugar := logger.Sugar()

	return &Logger{
		zap: sugar,
	}
}

func (l *Logger) Info(v ...interface{}) {
	l.zap.Info(v...)
}

func (l *Logger) Infof(format string, v ...interface{}) {
	l.zap.Infof(format, v...)
}

func (l *Logger) Debug(v ...interface{}) {
	l.zap.Debug(v...)
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	l.zap.Debugf(format, v...)
}

func (l *Logger) Warn(v ...interface{}) {
	l.zap.Warn(v...)
}

func (l *Logger) Warnf(format string, v ...interface{}) {
	l.zap.Warnf(format, v...)
}

func (l *Logger) Error(v ...interface{}) {
	l.zap.Error(v...)
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	l.zap.Errorf(format, v...)
}

func (l *Logger) Fatal(v ...interface{}) {
	l.zap.Fatal(v...)
}

func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.zap.Fatalf(format, v...)
}

