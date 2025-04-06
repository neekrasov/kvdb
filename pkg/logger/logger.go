package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/natefinch/lumberjack"
)

var (
	logger *zap.Logger

	defaultLoggerFilename        = "/var/log/kvdb/server.log"
	defaultLoggerMaxSizeMb       = 10
	defaultLoggerMaxBackupsCount = 3
	defaultLoggerMaxAgeDays      = 7
)

// MockLogger - mocks logger
func MockLogger() {
	logger = zap.NewNop()
}

// InitLogger - initializes logger with level
func InitLogger(level, ouput string) {
	Init(getCore(getAtomicLevel(level), ouput))
}

// Init - initializes new logger
func Init(core zapcore.Core, options ...zap.Option) {
	logger = zap.New(core, options...)
}

// Debug - used for debug logging
func Debug(msg string, fields ...zap.Field) {
	logger.Debug(msg, fields...)
}

// Info - used for info logging
func Info(msg string, fields ...zap.Field) {
	logger.Info(msg, fields...)
}

// Warn - used for warn logging
func Warn(msg string, fields ...zap.Field) {
	logger.Warn(msg, fields...)
}

// Error - used for error logging
func Error(msg string, fields ...zap.Field) {
	logger.Error(msg, fields...)
}

// Fatal - used for fatal logging
func Fatal(msg string, fields ...zap.Field) {
	logger.Fatal(msg, fields...)
}

// WithOptions - applies options
func WithOptions(opts ...zap.Option) *zap.Logger {
	return logger.WithOptions(opts...)
}

func getAtomicLevel(logLevel string) zap.AtomicLevel {
	var level zapcore.Level
	if err := level.Set(logLevel); err != nil {
		Fatal("failed to set log level: ", zap.Error(err))
	}

	return zap.NewAtomicLevelAt(level)
}

func getCore(level zap.AtomicLevel, output string) zapcore.Core {
	if output == "" {
		output = defaultLoggerFilename
	}

	stdout := zapcore.AddSync(os.Stdout)
	file := zapcore.AddSync(&lumberjack.Logger{
		Filename:   output,
		MaxSize:    defaultLoggerMaxSizeMb,
		MaxBackups: defaultLoggerMaxBackupsCount,
		MaxAge:     defaultLoggerMaxAgeDays,
	})

	productionCfg := zap.NewProductionEncoderConfig()
	productionCfg.TimeKey = "timestamp"
	productionCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	developmentCfg := zap.NewDevelopmentEncoderConfig()
	developmentCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder

	consoleEncoder := zapcore.NewConsoleEncoder(developmentCfg)
	fileEncoder := zapcore.NewJSONEncoder(productionCfg)

	return zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, stdout, level),
		zapcore.NewCore(fileEncoder, file, level),
	)
}
