package logger

import (
	"os"
	"path"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/natefinch/lumberjack"
)

var (
	logger *zap.Logger

	defaultLoggerFilename        = "kvdb.log"
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
	var tee []zapcore.Core
	if output != "" {
		productionCfg := zap.NewProductionEncoderConfig()
		productionCfg.TimeKey = "timestamp"
		productionCfg.EncodeTime = zapcore.ISO8601TimeEncoder

		file := zapcore.AddSync(
			&lumberjack.Logger{
				Filename:   path.Join(output, defaultLoggerFilename),
				MaxSize:    defaultLoggerMaxSizeMb,
				MaxBackups: defaultLoggerMaxBackupsCount,
				MaxAge:     defaultLoggerMaxAgeDays,
			})
		fileEncoder := zapcore.NewJSONEncoder(productionCfg)
		tee = append(tee, zapcore.NewCore(fileEncoder, file, level))
	}

	developmentCfg := zap.NewDevelopmentEncoderConfig()
	developmentCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	consoleEncoder := zapcore.NewConsoleEncoder(developmentCfg)
	tee = append(tee, zapcore.NewCore(
		consoleEncoder, zapcore.AddSync(os.Stdout), level))

	return zapcore.NewTee(tee...)
}
