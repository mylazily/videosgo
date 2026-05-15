package middleware

import (
	"context"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mylazily/videosgo/internal/database"
)

// 已知的搜索引擎爬虫 User-Agent 关键词
var botKeywords = []string{
	"Googlebot",
	"Baiduspider",
	"Slurp",        // Yahoo
	"bingbot",
	"YandexBot",
	"DuckDuckBot",
	"Baiduspider",
	"Sogou",
	"360Spider",
	"Bytespider",   // 字节跳动
	"facebookexternalhit",
	"Twitterbot",
	"linkedinbot",
	"AhrefsBot",
	"SemrushBot",
	"MJ12bot",
	"DotBot",
}

// 敏感地区 IP 段（示例，实际应从配置或 GeoIP 数据库加载）
var sensitiveRegions = []string{
	// 可以根据实际需求配置
}

// UASplit UA/IP 智能分流中间件
// 检测逻辑：
// 1. 检查是否匹配 301 重定向规则
// 2. 检测 User-Agent 是否为搜索引擎爬虫
// 3. 检测 IP 地区是否为敏感地区
// 4. 正常用户不做任何处理
func UASplit(redirectFn func(domain, path, ua, ip string) (targetURL string, found bool)) gin.HandlerFunc {
	return func(c *gin.Context) {
		ua := c.GetHeader("User-Agent")
		ip := c.ClientIP()
		host := c.Request.Host
		path := c.Request.URL.Path

		// 1. 检查 301 重定向规则
		if redirectFn != nil {
			targetURL, found := redirectFn(host, path, ua, ip)
			if found && targetURL != "" {
				c.Redirect(301, targetURL)
				c.Abort()
				return
			}
		}

		// 2. 检测 User-Agent 是否为搜索引擎爬虫
		if isBotUA(ua) {
			c.Set("is_bot", true)
			c.Set("bot_type", detectBotType(ua))
		}

		// 3. 检测 IP 地区
		if isSensitiveIP(ip) {
			c.Set("is_sensitive_region", true)
		}

		c.Next()
	}
}

// isBotUA 检测 User-Agent 是否为搜索引擎爬虫
func isBotUA(ua string) bool {
	if ua == "" {
		return false
	}

	uaLower := strings.ToLower(ua)

	// 先检查 Redis 缓存
	if database.RDB != nil {
		cacheKey := "ua:bot:cached:" + hashUA(ua)
		cached, err := database.RDB.Get(context.Background(), cacheKey).Result()
		if err == nil && cached == "1" {
			return true
		}
	}

	// 关键词匹配
	isBot := false
	for _, keyword := range botKeywords {
		if strings.Contains(uaLower, strings.ToLower(keyword)) {
			isBot = true
			break
		}
	}

	// 写入 Redis 缓存（24 小时过期）
	if isBot && database.RDB != nil {
		cacheKey := "ua:bot:cached:" + hashUA(ua)
		database.RDB.Set(context.Background(), cacheKey, "1", 24*time.Hour)
	}

	return isBot
}

// detectBotType 检测爬虫类型
func detectBotType(ua string) string {
	uaLower := strings.ToLower(ua)
	switch {
	case strings.Contains(uaLower, "googlebot"):
		return "google"
	case strings.Contains(uaLower, "baiduspider"):
		return "baidu"
	case strings.Contains(uaLower, "bingbot"):
		return "bing"
	case strings.Contains(uaLower, "yandexbot"):
		return "yandex"
	case strings.Contains(uaLower, "sogou"):
		return "sogou"
	case strings.Contains(uaLower, "360spider"):
		return "360"
	case strings.Contains(uaLower, "bytespider"):
		return "bytedance"
	case strings.Contains(uaLower, "duckduckbot"):
		return "duckduckgo"
	case strings.Contains(uaLower, "slurp"):
		return "yahoo"
	default:
		return "unknown"
	}
}

// isSensitiveIP 检测 IP 是否属于敏感地区
func isSensitiveIP(ip string) bool {
	if ip == "" || len(sensitiveRegions) == 0 {
		return false
	}

	// 先检查 Redis 缓存
	if database.RDB != nil {
		cacheKey := "ip:sensitive:" + ip
		cached, err := database.RDB.Get(context.Background(), cacheKey).Result()
		if err == nil {
			return cached == "1"
		}
	}

	// 简单的前缀匹配（生产环境应使用 GeoIP 数据库）
	isSensitive := false
	for _, region := range sensitiveRegions {
		if strings.HasPrefix(ip, region) {
			isSensitive = true
			break
		}
	}

	// 写入 Redis 缓存（24 小时过期）
	if database.RDB != nil {
		cacheKey := "ip:sensitive:" + ip
		val := "0"
		if isSensitive {
			val = "1"
		}
		database.RDB.Set(context.Background(), cacheKey, val, 24*time.Hour)
	}

	return isSensitive
}

// hashUA 对 UA 字符串进行简单哈希（用于缓存 key）
func hashUA(ua string) string {
	if len(ua) > 64 {
		return ua[:64]
	}
	return ua
}
