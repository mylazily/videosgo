// Package config 基于 Viper 的环境变量配置管理
package config

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

// Config 全局配置结构
type Config struct {
	App       AppConfig       `mapstructure:"app"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Redis     RedisConfig     `mapstructure:"redis"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	Crypto    CryptoConfig    `mapstructure:"crypto"`
	Collector CollectorConfig `mapstructure:"collector"`
	Security  SecurityConfig  `mapstructure:"security"`
	Log       LogConfig       `mapstructure:"log"`
	TG        TGConfig        `mapstructure:"tg"`
	Push      PushConfig      `mapstructure:"push"`
	Payment   PaymentConfig   `mapstructure:"payment"`
	Domain    DomainConfig    `mapstructure:"domain"`
}

// AppConfig 应用配置
type AppConfig struct {
	Env    string `mapstructure:"env"`
	Port   string `mapstructure:"port"`
	Secret string `mapstructure:"secret"`
}

// Validate 验证应用配置
func (c *AppConfig) Validate() error {
	if c.Port == "" {
		return fmt.Errorf("应用端口不能为空")
	}
	// 验证端口是否为数字
	if _, err := strconv.Atoi(c.Port); err != nil {
		return fmt.Errorf("应用端口必须是数字: %s", c.Port)
	}
	if c.Secret == "" {
		return fmt.Errorf("应用密钥不能为空")
	}
	if len(c.Secret) < 16 {
		return fmt.Errorf("应用密钥长度不能少于 16 位")
	}
	return nil
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

// Validate 验证数据库配置
func (d *DatabaseConfig) Validate() error {
	if d.Host == "" {
		return fmt.Errorf("数据库主机不能为空")
	}
	if d.Port == "" {
		return fmt.Errorf("数据库端口不能为空")
	}
	if d.User == "" {
		return fmt.Errorf("数据库用户不能为空")
	}
	if d.Password == "" {
		return fmt.Errorf("数据库密码不能为空")
	}
	if d.Name == "" {
		return fmt.Errorf("数据库名称不能为空")
	}
	if d.SSLMode == "" {
		d.SSLMode = "disable"
	}
	if d.MaxIdleConns <= 0 {
		d.MaxIdleConns = 10
	}
	if d.MaxOpenConns <= 0 {
		d.MaxOpenConns = 100
	}
	if d.MaxIdleConns > d.MaxOpenConns {
		return fmt.Errorf("最大空闲连接数不能大于最大打开连接数")
	}
	return nil
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

// Validate 验证 Redis 配置
func (r *RedisConfig) Validate() error {
	if r.Host == "" {
		return fmt.Errorf("Redis 主机不能为空")
	}
	if r.Port == "" {
		return fmt.Errorf("Redis 端口不能为空")
	}
	if r.DB < 0 || r.DB > 15 {
		return fmt.Errorf("Redis 数据库索引必须在 0-15 之间")
	}
	return nil
}

// JWTConfig JWT 配置
type JWTConfig struct {
	Secret      string `mapstructure:"secret"`
	ExpireHours int    `mapstructure:"expire_hours"`
}

// Validate 验证 JWT 配置
func (j *JWTConfig) Validate() error {
	if j.Secret == "" {
		return fmt.Errorf("JWT 密钥不能为空")
	}
	if len(j.Secret) < 32 {
		return fmt.Errorf("JWT 密钥长度不能少于 32 位，建议使用随机生成的强密钥")
	}
	if j.ExpireHours <= 0 {
		j.ExpireHours = 72
	}
	if j.ExpireHours > 720 {
		return fmt.Errorf("JWT 过期时间不能超过 720 小时（30 天）")
	}
	return nil
}

// CryptoConfig 加密配置
type CryptoConfig struct {
	Key string `mapstructure:"key"`
}

// Validate 验证加密配置
func (c *CryptoConfig) Validate() error {
	if c.Key == "" {
		return fmt.Errorf("加密密钥不能为空")
	}
	if len(c.Key) < 16 {
		return fmt.Errorf("加密密钥长度不能少于 16 位")
	}
	return nil
}

// CollectorConfig 采集器配置
type CollectorConfig struct {
	Workers      int    `mapstructure:"workers"`
	RetryMax     int    `mapstructure:"retry_max"`
	ProbeTimeout string `mapstructure:"probe_timeout"`
}

// Validate 验证采集器配置
func (c *CollectorConfig) Validate() error {
	if c.Workers <= 0 {
		c.Workers = 10
	}
	if c.Workers > 100 {
		return fmt.Errorf("采集 Worker 数量不能超过 100")
	}
	if c.RetryMax < 0 {
		c.RetryMax = 3
	}
	if c.RetryMax > 10 {
		return fmt.Errorf("重试次数不能超过 10")
	}
	if c.ProbeTimeout == "" {
		c.ProbeTimeout = "2s"
	}
	return nil
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

// Validate 验证安全配置
func (s *SecurityConfig) Validate() error {
	if s.RateLimitPerMinute <= 0 {
		s.RateLimitPerMinute = 60
	}
	if s.RateLimitPerMinute > 10000 {
		return fmt.Errorf("速率限制每分钟请求数不能超过 10000")
	}
	return nil
}

// LogConfig 日志配置
type LogConfig struct {
	Level  string `mapstructure:"level"`
	Output string `mapstructure:"output"`
}

// Validate 验证日志配置
func (l *LogConfig) Validate() error {
	validLevels := []string{"debug", "info", "warn", "error", "fatal"}
	levelLower := strings.ToLower(l.Level)
	isValid := false
	for _, level := range validLevels {
		if levelLower == level {
			isValid = true
			break
		}
	}
	if !isValid {
		l.Level = "info"
	}
	return nil
}

// TGConfig Telegram Bot 配置
type TGConfig struct {
	BotToken     string `mapstructure:"bot_token"`
	WebhookURL   string `mapstructure:"webhook_url"`
	AdminUserIDs string `mapstructure:"admin_user_ids"` // 逗号分隔的 TG 用户 ID
}

// PushConfig Web Push 配置
type PushConfig struct {
	VAPIDPublicKey  string `mapstructure:"vapid_public_key"`
	VAPIDPrivateKey string `mapstructure:"vapid_private_key"`
	VAPIDSubject    string `mapstructure:"vapid_subject"`
}

// PaymentConfig 支付配置
type PaymentConfig struct {
	DefaultChannel string `mapstructure:"default_channel"`
	OrderExpireSec int    `mapstructure:"order_expire_sec"`
}

// DomainConfig 域名轮换配置
type DomainConfig struct {
	CheckIntervalSec int    `mapstructure:"check_interval_sec"`
	DefaultRegion    string `mapstructure:"default_region"`
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

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	GlobalConfig = config
	return config, nil
}

// Validate 验证所有配置
func (c *Config) Validate() error {
	if err := c.App.Validate(); err != nil {
		return fmt.Errorf("应用配置错误: %w", err)
	}
	if err := c.Database.Validate(); err != nil {
		return fmt.Errorf("数据库配置错误: %w", err)
	}
	if err := c.Redis.Validate(); err != nil {
		return fmt.Errorf("Redis 配置错误: %w", err)
	}
	if err := c.JWT.Validate(); err != nil {
		return fmt.Errorf("JWT 配置错误: %w", err)
	}
	if err := c.Crypto.Validate(); err != nil {
		return fmt.Errorf("加密配置错误: %w", err)
	}
	if err := c.Collector.Validate(); err != nil {
		return fmt.Errorf("采集器配置错误: %w", err)
	}
	if err := c.Security.Validate(); err != nil {
		return fmt.Errorf("安全配置错误: %w", err)
	}
	if err := c.Log.Validate(); err != nil {
		return fmt.Errorf("日志配置错误: %w", err)
	}
	return nil
}

// IsDevelopment 检查是否为开发环境
func (c *Config) IsDevelopment() bool {
	return strings.ToLower(c.App.Env) == "development" || strings.ToLower(c.App.Env) == "dev"
}

// IsProduction 检查是否为生产环境
func (c *Config) IsProduction() bool {
	return strings.ToLower(c.App.Env) == "production" || strings.ToLower(c.App.Env) == "prod"
}

// setDefaults 设置默认值
func setDefaults() {
	// 应用
	viper.SetDefault("APP_ENV", "development")
	viper.SetDefault("APP_PORT", "8080")
	viper.SetDefault("APP_SECRET", "videosgo-default-secret-change-in-production")

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
	viper.SetDefault("JWT_SECRET", "videosgo-jwt-secret-change-in-production-min-32-chars")
	viper.SetDefault("JWT_EXPIRE_HOURS", 72)

	// 加密
	viper.SetDefault("CRYPTO_KEY", "videosgo-crypto-key-change-in-production")

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
	viper.SetDefault("LOG_OUTPUT", "")

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
	viper.BindEnv("log.output", "LOG_OUTPUT")

	// TG Bot
	viper.BindEnv("tg.bot_token", "TG_BOT_TOKEN")
	viper.BindEnv("tg.webhook_url", "TG_WEBHOOK_URL")
	viper.BindEnv("tg.admin_user_ids", "TG_ADMIN_USER_IDS")
}
