// Package config 基于 Viper 的环境变量配置管理
package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config 全局配置结构
type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Crypto   CryptoConfig   `mapstructure:"crypto"`
	Collector CollectorConfig `mapstructure:"collector"`
	Security SecurityConfig `mapstructure:"security"`
	Log      LogConfig      `mapstructure:"log"`
}

// AppConfig 应用配置
type AppConfig struct {
	Env    string `mapstructure:"env"`
	Port   string `mapstructure:"port"`
	Secret string `mapstructure:"secret"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host         string `mapstructure:"host"`
	Port         string `mapstructure:"port"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	Name         string `mapstructure:"name"`
	SSLMode      string `mapstructure:"sslmode"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
}

// DSN 返回 PostgreSQL 连接字符串
func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode,
	)
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// Addr 返回 Redis 地址
func (r *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%s", r.Host, r.Port)
}

// JWTConfig JWT 配置
type JWTConfig struct {
	Secret      string `mapstructure:"secret"`
	ExpireHours int    `mapstructure:"expire_hours"`
}

// CryptoConfig 加密配置
type CryptoConfig struct {
	Key string `mapstructure:"key"`
}

// CollectorConfig 采集器配置
type CollectorConfig struct {
	Workers      int    `mapstructure:"workers"`
	RetryMax     int    `mapstructure:"retry_max"`
	ProbeTimeout string `mapstructure:"probe_timeout"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	CORSAllowedOrigins string `mapstructure:"cors_allowed_origins"`
	UAWhitelistPaths   string `mapstructure:"ua_whitelist_paths"`
	UAFilterEnabled    bool   `mapstructure:"ua_filter_enabled"`
	WAFEnabled         bool   `mapstructure:"waf_enabled"`
	RateLimitPerMinute int    `mapstructure:"rate_limit_per_minute"`
}

// CORSOrigins 返回 CORS 允许的域名列表
func (s *SecurityConfig) CORSOrigins() []string {
	if s.CORSAllowedOrigins == "" {
		return []string{"*"}
	}
	return strings.Split(s.CORSAllowedOrigins, ",")
}

// UAWhitelist 返回 UA 白名单路径列表
func (s *SecurityConfig) UAWhitelist() []string {
	if s.UAWhitelistPaths == "" {
		return []string{}
	}
	return strings.Split(s.UAWhitelistPaths, ",")
}

// LogConfig 日志配置
type LogConfig struct {
	Level string `mapstructure:"level"`
}

var GlobalConfig *Config

// Load 加载配置
func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	// 设置默认值
	setDefaults()

	// 读取配置文件（如果存在）
	_ = viper.ReadInConfig()

	config := &Config{}
	if err := viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("配置解析失败: %w", err)
	}

	GlobalConfig = config
	return config, nil
}

// setDefaults 设置默认值
func setDefaults() {
	// 应用
	viper.SetDefault("APP_ENV", "development")
	viper.SetDefault("APP_PORT", "8080")
	viper.SetDefault("APP_SECRET", "videosgo-default-secret")

	// 数据库
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_USER", "postgres")
	viper.SetDefault("DB_PASSWORD", "postgres")
	viper.SetDefault("DB_NAME", "videosgo")
	viper.SetDefault("DB_SSLMODE", "disable")
	viper.SetDefault("DB_MAX_IDLE_CONNS", 10)
	viper.SetDefault("DB_MAX_OPEN_CONNS", 100)

	// Redis
	viper.SetDefault("REDIS_HOST", "localhost")
	viper.SetDefault("REDIS_PORT", "6379")
	viper.SetDefault("REDIS_PASSWORD", "")
	viper.SetDefault("REDIS_DB", 0)

	// JWT
	viper.SetDefault("JWT_SECRET", "videosgo-jwt-secret")
	viper.SetDefault("JWT_EXPIRE_HOURS", 72)

	// 加密
	viper.SetDefault("CRYPTO_KEY", "videosgo-crypto-key")

	// 采集器
	viper.SetDefault("COLLECTOR_WORKERS", 10)
	viper.SetDefault("COLLECTOR_RETRY_MAX", 3)
	viper.SetDefault("COLLECTOR_PROBE_TIMEOUT", "2s")

	// 安全
	viper.SetDefault("CORS_ALLOWED_ORIGINS", "*")
	viper.SetDefault("UA_WHITELIST_PATHS", "/api/v1/health,/api/v1/ping")
	viper.SetDefault("UA_FILTER_ENABLED", false)
	viper.SetDefault("WAF_ENABLED", true)
	viper.SetDefault("RATE_LIMIT_PER_MINUTE", 60)

	// 日志
	viper.SetDefault("LOG_LEVEL", "info")

	// 环境变量绑定
	viper.BindEnv("app.env", "APP_ENV")
	viper.BindEnv("app.port", "APP_PORT")
	viper.BindEnv("app.secret", "APP_SECRET")

	viper.BindEnv("database.host", "DB_HOST")
	viper.BindEnv("database.port", "DB_PORT")
	viper.BindEnv("database.user", "DB_USER")
	viper.BindEnv("database.password", "DB_PASSWORD")
	viper.BindEnv("database.name", "DB_NAME")
	viper.BindEnv("database.sslmode", "DB_SSLMODE")
	viper.BindEnv("database.max_idle_conns", "DB_MAX_IDLE_CONNS")
	viper.BindEnv("database.max_open_conns", "DB_MAX_OPEN_CONNS")

	viper.BindEnv("redis.host", "REDIS_HOST")
	viper.BindEnv("redis.port", "REDIS_PORT")
	viper.BindEnv("redis.password", "REDIS_PASSWORD")
	viper.BindEnv("redis.db", "REDIS_DB")

	viper.BindEnv("jwt.secret", "JWT_SECRET")
	viper.BindEnv("jwt.expire_hours", "JWT_EXPIRE_HOURS")

	viper.BindEnv("crypto.key", "CRYPTO_KEY")

	viper.BindEnv("collector.workers", "COLLECTOR_WORKERS")
	viper.BindEnv("collector.retry_max", "COLLECTOR_RETRY_MAX")
	viper.BindEnv("collector.probe_timeout", "COLLECTOR_PROBE_TIMEOUT")

	viper.BindEnv("security.cors_allowed_origins", "CORS_ALLOWED_ORIGINS")
	viper.BindEnv("security.ua_whitelist_paths", "UA_WHITELIST_PATHS")
	viper.BindEnv("security.ua_filter_enabled", "UA_FILTER_ENABLED")
	viper.BindEnv("security.waf_enabled", "WAF_ENABLED")
	viper.BindEnv("security.rate_limit_per_minute", "RATE_LIMIT_PER_MINUTE")

	viper.BindEnv("log.level", "LOG_LEVEL")
}
