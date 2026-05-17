package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger 创建新的日志实例
func NewLogger() *zap.Logger {
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.CallerKey = "caller"
	config.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	logger, err := config.Build()
	if err != nil {
		panic(err)
	}

	return logger
}
