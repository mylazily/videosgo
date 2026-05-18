package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// L 全局日志实例
var L *zap.Logger

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

	// 赋值给全局变量
	L = logger

	return logger
}

// Debugf 格式化调试日志
func Debugf(template string, args ...interface{}) {
	if L != nil {
		L.Sugar().Debugf(template, args...)
	}
}

// Infof 格式化信息日志
func Infof(template string, args ...interface{}) {
	if L != nil {
		L.Sugar().Infof(template, args...)
	}
}

// Warnf 格式化警告日志
func Warnf(template string, args ...interface{}) {
	if L != nil {
		L.Sugar().Warnf(template, args...)
	}
}

// Errorf 格式化错误日志
func Errorf(template string, args ...interface{}) {
	if L != nil {
		L.Sugar().Errorf(template, args...)
	}
}

// Fatalf 格式化致命日志
func Fatalf(template string, args ...interface{}) {
	if L != nil {
		L.Sugar().Fatalf(template, args...)
	}
}
