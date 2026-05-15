// Package logger 结构化日志模块
package logger

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// LogLevel 日志级别
type LogLevel int

const (
	// DebugLevel 调试级别
	DebugLevel LogLevel = iota
	// InfoLevel 信息级别
	InfoLevel
	// WarnLevel 警告级别
	WarnLevel
	// ErrorLevel 错误级别
	ErrorLevel
	// FatalLevel 致命级别
	FatalLevel
)

// String 返回日志级别字符串
func (l LogLevel) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	case FatalLevel:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// ParseLevel 解析日志级别字符串
func ParseLevel(s string) LogLevel {
	switch strings.ToUpper(s) {
	case "DEBUG":
		return DebugLevel
	case "INFO":
		return InfoLevel
	case "WARN", "WARNING":
		return WarnLevel
	case "ERROR":
		return ErrorLevel
	case "FATAL":
		return FatalLevel
	default:
		return InfoLevel
	}
}

// Fields 日志字段
type Fields map[string]interface{}

// Entry 日志条目
type Entry struct {
	Logger  *Logger
	Level   LogLevel
	Time    time.Time
	Message string
	Fields  Fields
	Caller  string
}

// Logger 日志记录器
type Logger struct {
	level      LogLevel
	output     io.Writer
	formatter  Formatter
	mu         sync.RWMutex
	fields     Fields
	callerSkip int
}

// Formatter 日志格式化接口
type Formatter interface {
	Format(entry *Entry) ([]byte, error)
}

// TextFormatter 文本格式化器
type TextFormatter struct {
	DisableColors bool
	FullTimestamp bool
}

// Format 格式化日志条目为文本
func (f *TextFormatter) Format(entry *Entry) ([]byte, error) {
	timestamp := entry.Time.Format("2006-01-02 15:04:05.000")
	level := entry.Level.String()
	
	var sb strings.Builder
	sb.WriteString("[")
	sb.WriteString(timestamp)
	sb.WriteString("]")
	sb.WriteString("[")
	sb.WriteString(level)
	sb.WriteString("]")
	
	if entry.Caller != "" {
		sb.WriteString("[")
		sb.WriteString(entry.Caller)
		sb.WriteString("]")
	}
	
	sb.WriteString(" ")
	sb.WriteString(entry.Message)
	
	if len(entry.Fields) > 0 {
		for k, v := range entry.Fields {
			sb.WriteString(fmt.Sprintf(" %s=%v", k, v))
		}
	}
	
	sb.WriteString("\n")
	return []byte(sb.String()), nil
}

// JSONFormatter JSON 格式化器
type JSONFormatter struct {
	PrettyPrint bool
}

// Format 格式化日志条目为 JSON
func (f *JSONFormatter) Format(entry *Entry) ([]byte, error) {
	data := map[string]interface{}{
		"timestamp": entry.Time.Format(time.RFC3339Nano),
		"level":     entry.Level.String(),
		"message":   entry.Message,
	}
	
	if entry.Caller != "" {
		data["caller"] = entry.Caller
	}
	
	for k, v := range entry.Fields {
		data[k] = v
	}
	
	return nil, nil // 简化实现，实际使用 json.Marshal
}

var (
	// 全局默认日志实例
	std = New()
)

// New 创建新的日志记录器
func New() *Logger {
	return &Logger{
		level:      InfoLevel,
		output:     os.Stdout,
		formatter:  &TextFormatter{FullTimestamp: true},
		fields:     make(Fields),
		callerSkip: 3,
	}
}

// SetLevel 设置日志级别
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// SetOutput 设置输出目标
func (l *Logger) SetOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.output = w
}

// SetFormatter 设置格式化器
func (l *Logger) SetFormatter(f Formatter) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.formatter = f
}

// WithField 添加字段
func (l *Logger) WithField(key string, value interface{}) *Logger {
	return l.WithFields(Fields{key: value})
}

// WithFields 添加多个字段
func (l *Logger) WithFields(fields Fields) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	newFields := make(Fields, len(l.fields)+len(fields))
	for k, v := range l.fields {
		newFields[k] = v
	}
	for k, v := range fields {
		newFields[k] = v
	}
	
	return &Logger{
		level:      l.level,
		output:     l.output,
		formatter:  l.formatter,
		fields:     newFields,
		callerSkip: l.callerSkip,
	}
}

// WithError 添加错误字段
func (l *Logger) WithError(err error) *Logger {
	return l.WithField("error", err)
}

// getCaller 获取调用者信息
func (l *Logger) getCaller() string {
	_, file, line, ok := runtime.Caller(l.callerSkip)
	if !ok {
		return ""
	}
	return fmt.Sprintf("%s:%d", filepath.Base(file), line)
}

// log 写入日志
func (l *Logger) log(level LogLevel, msg string) {
	l.mu.RLock()
	if level < l.level {
		l.mu.RUnlock()
		return
	}
	output := l.output
	formatter := l.formatter
	fields := l.fields
	l.mu.RUnlock()
	
	entry := &Entry{
		Logger:  l,
		Level:   level,
		Time:    time.Now(),
		Message: msg,
		Fields:  fields,
		Caller:  l.getCaller(),
	}
	
	if level >= ErrorLevel {
		entry.Caller = l.getCaller()
	}
	
	data, err := formatter.Format(entry)
	if err != nil {
		log.Printf("日志格式化失败: %v", err)
		return
	}
	
	if data == nil {
		// JSONFormatter 返回 nil，使用默认格式
		data = []byte(fmt.Sprintf("[%s][%s] %s\n", entry.Time.Format("2006-01-02 15:04:05"), level.String(), msg))
	}
	
	output.Write(data)
	
	if level == FatalLevel {
		os.Exit(1)
	}
}

// Debug 调试日志
func (l *Logger) Debug(args ...interface{}) {
	l.log(DebugLevel, fmt.Sprint(args...))
}

// Debugf 格式化调试日志
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.log(DebugLevel, fmt.Sprintf(format, args...))
}

// Info 信息日志
func (l *Logger) Info(args ...interface{}) {
	l.log(InfoLevel, fmt.Sprint(args...))
}

// Infof 格式化信息日志
func (l *Logger) Infof(format string, args ...interface{}) {
	l.log(InfoLevel, fmt.Sprintf(format, args...))
}

// Warn 警告日志
func (l *Logger) Warn(args ...interface{}) {
	l.log(WarnLevel, fmt.Sprint(args...))
}

// Warnf 格式化警告日志
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.log(WarnLevel, fmt.Sprintf(format, args...))
}

// Error 错误日志
func (l *Logger) Error(args ...interface{}) {
	l.log(ErrorLevel, fmt.Sprint(args...))
}

// Errorf 格式化错误日志
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.log(ErrorLevel, fmt.Sprintf(format, args...))
}

// Fatal 致命日志
func (l *Logger) Fatal(args ...interface{}) {
	l.log(FatalLevel, fmt.Sprint(args...))
}

// Fatalf 格式化致命日志
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.log(FatalLevel, fmt.Sprintf(format, args...))
}

// 全局函数

// SetLevel 设置全局日志级别
func SetLevel(level LogLevel) {
	std.SetLevel(level)
}

// SetOutput 设置全局输出目标
func SetOutput(w io.Writer) {
	std.SetOutput(w)
}

// WithField 添加字段到全局日志
func WithField(key string, value interface{}) *Logger {
	return std.WithField(key, value)
}

// WithFields 添加多个字段到全局日志
func WithFields(fields Fields) *Logger {
	return std.WithFields(fields)
}

// WithError 添加错误字段到全局日志
func WithError(err error) *Logger {
	return std.WithError(err)
}

// Debug 全局调试日志
func Debug(args ...interface{}) {
	std.Debug(args...)
}

// Debugf 全局格式化调试日志
func Debugf(format string, args ...interface{}) {
	std.Debugf(format, args...)
}

// Info 全局信息日志
func Info(args ...interface{}) {
	std.Info(args...)
}

// Infof 全局格式化信息日志
func Infof(format string, args ...interface{}) {
	std.Infof(format, args...)
}

// Warn 全局警告日志
func Warn(args ...interface{}) {
	std.Warn(args...)
}

// Warnf 全局格式化警告日志
func Warnf(format string, args ...interface{}) {
	std.Warnf(format, args...)
}

// Error 全局错误日志
func Error(args ...interface{}) {
	std.Error(args...)
}

// Errorf 全局格式化错误日志
func Errorf(format string, args ...interface{}) {
	std.Errorf(format, args...)
}

// Fatal 全局致命日志
func Fatal(args ...interface{}) {
	std.Fatal(args...)
}

// Fatalf 全局格式化致命日志
func Fatalf(format string, args ...interface{}) {
	std.Fatalf(format, args...)
}

// InitLogger 初始化日志配置
func InitLogger(level string, outputPath ...string) {
	std.SetLevel(ParseLevel(level))
	
	if len(outputPath) > 0 && outputPath[0] != "" {
		// 同时输出到文件和控制台
		file, err := os.OpenFile(outputPath[0], os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			std.SetOutput(io.MultiWriter(os.Stdout, file))
		} else {
			Errorf("无法打开日志文件: %v", err)
		}
	}
}

// ContextKey 上下文键类型
type ContextKey string

const (
	// RequestIDKey 请求 ID 键
	RequestIDKey ContextKey = "request_id"
	// UserIDKey 用户 ID 键
	UserIDKey ContextKey = "user_id"
)

// FromContext 从上下文创建带字段的日志
func FromContext(ctx context.Context) *Logger {
	logger := std
	
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok && requestID != "" {
		logger = logger.WithField("request_id", requestID)
	}
	
	if userID, ok := ctx.Value(UserIDKey).(string); ok && userID != "" {
		logger = logger.WithField("user_id", userID)
	}
	
	return logger
}
