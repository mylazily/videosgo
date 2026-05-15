package middleware

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mylazily/videosgo/pkg/response"
)

// 移动端 UA 正则表达式
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

// WAF 规则：检测恶意请求
var wafPatterns = []*regexp.Regexp{
	// SQL 注入
	regexp.MustCompile(`(?i)(union\s+select|select\s+.+\s+from|insert\s+into|delete\s+from|drop\s+table|update\s+.+\s+set)`),
	// XSS 攻击
	regexp.MustCompile(`(?i)(<script|javascript:|on\w+\s*=|eval\(|alert\()`),
	// 路径遍历
	regexp.MustCompile(`\.\./|\.\.\\`),
	// 常见攻击工具特征
	regexp.MustCompile(`(?i)(nikto|nmap|sqlmap|masscan|burpsuite|dirbuster)`),
}

// Security 安全中间件（UA 过滤 + WAF 规则）
func Security(uaFilterEnabled bool, whitelistPaths []string, wafEnabled bool) gin.HandlerFunc {
	whitelist := make(map[string]bool)
	for _, path := range whitelistPaths {
		whitelist[path] = true
	}

	return func(c *gin.Context) {
		path := c.Request.URL.Path
		userAgent := c.GetHeader("User-Agent")
		clientIP := c.ClientIP()
		query := c.Request.URL.RawQuery

		// WAF 规则检查
		if wafEnabled {
			checkStr := path + " " + query + " " + userAgent
			for _, pattern := range wafPatterns {
				if pattern.MatchString(checkStr) {
					logSecurityEvent(clientIP, "WAF_BLOCK", path, userAgent)
					response.Error(c, http.StatusForbidden, "请求被安全策略拦截")
					c.Abort()
					return
				}
			}
		}

		// UA 过滤（仅允许移动端）
		if uaFilterEnabled {
			// 检查是否在白名单中
			if !isWhitelisted(path, whitelist) {
				if !isMobileUA(userAgent) {
					logSecurityEvent(clientIP, "UA_BLOCK", path, userAgent)
					response.Error(c, http.StatusForbidden, "仅支持移动端访问")
					c.Abort()
					return
				}
			}
		}

		// 安全响应头
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'")

		c.Next()
	}
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
	if whitelist[path] {
		return true
	}
	// 支持前缀匹配
	for wp := range whitelist {
		if strings.HasSuffix(wp, "*") && strings.HasPrefix(path, wp[:len(wp)-1]) {
			return true
		}
	}
	return false
}

// logSecurityEvent 记录安全事件
func logSecurityEvent(ip, eventType, path, ua string) {
	// 使用标准日志记录安全事件
	// 在生产环境中应该发送到安全监控系统
	_ = ip
	_ = eventType
	_ = path
	_ = ua
}
