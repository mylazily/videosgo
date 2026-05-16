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

// UASplit UA/IP 智能分流中间件
// 检测逻辑：
// 1. 检查是否匹配 301 重定向规则
// 2. 检测 User-Agent 是否为搜索引擎爬虫
// 3. 正常用户不做任何处理
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

// hashUA 对 UA 字符串进行简单哈希（用于缓存 key）
func hashUA(ua string) string {
	if len(ua) > 64 {
		return ua[:64]
	}
	return ua
}
