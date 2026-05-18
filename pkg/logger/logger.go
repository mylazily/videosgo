package logger

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// L 全局日志实例
var L *zap.Logger

// Init 初始化日志
func Init() error {
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.CallerKey = "caller"
	config.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	config.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	logger, err := config.Build()
	if err != nil {
		return err
	}
	L = logger
	return nil
}

// Sync 刷新日志缓冲区
func Sync() {
	if L != nil {
		_ = L.Sync()
	}
}

// Info 信息日志
func Info(msg string, fields ...zap.Field) {
	if L != nil {
		L.Info(msg, fields...)
	}
}

// Error 错误日志
func Error(msg string, fields ...zap.Field) {
	if L != nil {
		L.Error(msg, fields...)
	}
}

// Warn 警告日志
func Warn(msg string, fields ...zap.Field) {
	if L != nil {
		L.Warn(msg, fields...)
	}
}

// Debug 调试日志
func Debug(msg string, fields ...zap.Field) {
	if L != nil {
		L.Debug(msg, fields...)
	}
}

// Infof 格式化信息日志
func Infof(template string, args ...interface{}) {
	if L != nil {
		L.Sugar().Infof(template, args...)
	}
}

// Errorf 格式化错误日志
func Errorf(template string, args ...interface{}) {
	if L != nil {
		L.Sugar().Errorf(template, args...)
	}
}

// Warnf 格式化警告日志
func Warnf(template string, args ...interface{}) {
	if L != nil {
		L.Sugar().Warnf(template, args...)
	}
}

// Debugf 格式化调试日志
func Debugf(template string, args ...interface{}) {
	if L != nil {
		L.Sugar().Debugf(template, args...)
	}
}

// GinLogger Gin 日志中间件
func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		if L != nil {
			if len(query) > 0 {
				path = path + "?" + query
			}
			L.Info("HTTP",
				zap.Int("status", status),
				zap.String("method", c.Request.Method),
				zap.String("path", path),
				zap.Duration("latency", latency),
				zap.String("ip", c.ClientIP()),
			)
		}
	}
}

