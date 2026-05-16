package middleware

import (
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/mylazily/videosgo/internal/logger"
	"github.com/mylazily/videosgo/pkg/response"
)

// SecurityConfig 安全配置
type SecurityConfig struct {
	UAFilterEnabled  bool
	WAFEnabled       bool
	WhitelistPaths   []string
	MaxRequestSize   int64
	BlockUserAgents  []string
}

// DefaultSecurityConfig 返回默认安全配置
func DefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		UAFilterEnabled: false,
		WAFEnabled:      true,
		WhitelistPaths:  []string{"/api/v1/health", "/api/v1/ping"},
		MaxRequestSize:  10 * 1024 * 1024, // 10MB
		BlockUserAgents: []string{},
	}
}

// 移动端 UA 正则表达式 - 预编译以提高性能
var mobileUAPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)Android`),
	regexp.MustCompile(`(?i)iPhone`),
	regexp.MustCompile(`(?i)iPad`),
	regexp.MustCompile(`(?i)iPod`),
	regexp.MustCompile(`(?i)Mobile`),
	regexp.MustCompile(`(?i)BlackBerry`),
	regexp.MustCompile(`(?i)Windows Phone`),
	regexp.MustCompile(`(?i)webOS`),
	regexp.MustCompile(`(?i)Opera Mini`),
	regexp.MustCompile(`(?i)IEMobile`),
}

// WAF 规则：检测恶意请求 - 预编译以提高性能
var wafPatterns = []*regexp.Regexp{
	// SQL 注入
	regexp.MustCompile(`(?i)(union\s+select|select\s+.+\s+from|insert\s+into|delete\s+from|drop\s+table|update\s+.+\s+set)`),
	// XSS 攻击
	regexp.MustCompile(`(?i)(<script|javascript:|on\w+\s*=|eval\(|alert\()`),
	// 路径遍历
	regexp.MustCompile(`\.\./|\.\.\\`),
	// 常见攻击工具特征
	regexp.MustCompile(`(?i)(nikto|nmap|sqlmap|masscan|burpsuite|dirbuster)`),
	// 命令注入
	regexp.MustCompile(`(?i)(\||;|\$\(|\` + "`" + `|\|\||&&)`),
}

// 缓存编译后的白名单路径
var whitelistCache sync.Map

// Security 安全中间件（UA 过滤 + WAF 规则）
func Security(config *SecurityConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultSecurityConfig()
	}

	whitelist := make(map[string]bool)
	for _, path := range config.WhitelistPaths {
		whitelist[path] = true
	}

	return func(c *gin.Context) {
		path := c.Request.URL.Path
		userAgent := c.GetHeader("User-Agent")
		clientIP := c.ClientIP()
		query := c.Request.URL.RawQuery

		// 检查请求大小
		if config.MaxRequestSize > 0 && c.Request.ContentLength > config.MaxRequestSize {
			logger.Warnf("[Security] 请求过大: %s, size: %d, IP: %s", path, c.Request.ContentLength, clientIP)
			response.Error(c, http.StatusRequestEntityTooLarge, "请求体过大")
			c.Abort()
			return
		}

		// WAF 规则检查
		if config.WAFEnabled {
			if blocked, reason := checkWAF(path, query, userAgent); blocked {
				logSecurityEvent(clientIP, "WAF_BLOCK", path, userAgent, reason)
				response.Error(c, http.StatusForbidden, "请求被安全策略拦截")
				c.Abort()
				return
			}
		}

		// UA 过滤（仅允许移动端）
		if config.UAFilterEnabled {
			if !isWhitelisted(path, whitelist) {
				if !isMobileUA(userAgent) {
					logSecurityEvent(clientIP, "UA_BLOCK", path, userAgent, "非移动端访问")
					response.Error(c, http.StatusForbidden, "仅支持移动端访问")
					c.Abort()
					return
				}
			}
		}

		// 检查是否是需要阻止的 User-Agent
		for _, blockedUA := range config.BlockUserAgents {
			if strings.Contains(userAgent, blockedUA) {
				logSecurityEvent(clientIP, "UA_BLOCK", path, userAgent, "黑名单 User-Agent")
				response.Error(c, http.StatusForbidden, "请求被拒绝")
				c.Abort()
				return
			}
		}

		// 安全响应头
		setSecurityHeaders(c)

		c.Next()
	}
}

// checkWAF 检查 WAF 规则
func checkWAF(path, query, userAgent string) (bool, string) {
	checkStr := path + " " + query + " " + userAgent
	
	for i, pattern := range wafPatterns {
		if pattern.MatchString(checkStr) {
			switch i {
			case 0:
				return true, "SQL 注入检测"
			case 1:
				return true, "XSS 攻击检测"
			case 2:
				return true, "路径遍历检测"
			case 3:
				return true, "攻击工具特征检测"
			case 4:
				return true, "命令注入检测"
			}
		}
	}
	
	return false, ""
}

// isMobileUA 检查是否为移动端 UA
func isMobileUA(ua string) bool {
	if ua == "" {
		return false
	}
	for _, pattern := range mobileUAPatterns {
		if pattern.MatchString(ua) {
			return true
		}
	}
	return false
}

// isWhitelisted 检查路径是否在白名单中
func isWhitelisted(path string, whitelist map[string]bool) bool {
	// 检查缓存
	if val, ok := whitelistCache.Load(path); ok {
		return val.(bool)
	}

	// 精确匹配
	if whitelist[path] {
		whitelistCache.Store(path, true)
		return true
	}

	// 支持前缀匹配
	for wp := range whitelist {
		if strings.HasSuffix(wp, "*") && strings.HasPrefix(path, wp[:len(wp)-1]) {
			whitelistCache.Store(path, true)
			return true
		}
	}

	whitelistCache.Store(path, false)
	return false
}

// setSecurityHeaders 设置安全响应头
func setSecurityHeaders(c *gin.Context) {
	path := c.Request.URL.Path

	// 防止 MIME 类型嗅探
	c.Header("X-Content-Type-Options", "nosniff")

	// 防止点击劫持
	c.Header("X-Frame-Options", "DENY")

	// XSS 保护
	c.Header("X-XSS-Protection", "1; mode=block")

	// 引用策略
	c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

	// 内容安全策略（允许内联 JSON-LD 用于 SEO）
	c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; media-src 'self' https: blob:; connect-src 'self' https: wss:; frame-ancestors 'none'")

	// HSTS (仅在 HTTPS 环境下启用)
	// c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

	// 缓存策略：区分 API 和静态资源
	if strings.HasPrefix(path, "/api/") {
		// API 响应：不缓存
		c.Header("Cache-Control", "no-store, no-cache, must-revalidate")
		c.Header("Pragma", "no-cache")
	} else {
		// 静态资源和页面：允许浏览器缓存
		c.Header("Cache-Control", "public, max-age=3600")
	}
}

// logSecurityEvent 记录安全事件
func logSecurityEvent(ip, eventType, path, ua, reason string) {
	logger.Warnf("[Security] 安全事件: type=%s, ip=%s, path=%s, reason=%s, ua=%s", 
		eventType, ip, path, reason, ua)
}

// ClearWhitelistCache 清除白名单缓存
func ClearWhitelistCache() {
	whitelistCache.Range(func(key, value interface{}) bool {
		whitelistCache.Delete(key)
		return true
	})
}
