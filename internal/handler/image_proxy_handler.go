// Package handler 图片代理处理器
// 让 Cloudflare 缓存外链图片，不占用自己服务器带宽
package handler

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mylazily/videosgo/pkg/response"
)

// ImageProxyHandler 图片代理处理器
type ImageProxyHandler struct {
	client *http.Client
}

// NewImageProxyHandler 创建图片代理处理器
func NewImageProxyHandler() *ImageProxyHandler {
	return &ImageProxyHandler{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// ProxyImage 代理图片请求
// GET /api/v1/image/proxy?url=https://example.com/image.jpg
func (h *ImageProxyHandler) ProxyImage(c *gin.Context) {
	imageURL := c.Query("url")
	if imageURL == "" {
		response.BadRequest(c, "缺少图片URL")
		return
	}

	// 验证URL格式
	parsedURL, err := url.Parse(imageURL)
	if err != nil || !strings.HasPrefix(parsedURL.Scheme, "http") {
		response.BadRequest(c, "无效的图片URL")
		return
	}

	// 请求目标图片
	req, err := http.NewRequest("GET", imageURL, nil)
	if err != nil {
		response.InternalError(c, "请求失败")
		return
	}

	// 设置请求头，模拟浏览器
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Referer", parsedURL.Scheme+"://"+parsedURL.Host)

	resp, err := h.client.Do(req)
	if err != nil {
		response.InternalError(c, "获取图片失败")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		response.InternalError(c, "图片获取失败")
		return
	}

	// 设置响应头 - 让CF缓存1年
	c.Header("Cache-Control", "public, max-age=31536000, immutable")
	c.Header("Access-Control-Allow-Origin", "*")

	// 透传Content-Type
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/jpeg"
	}
	c.Header("Content-Type", contentType)

	// 透传图片数据
	c.Status(http.StatusOK)
	io.Copy(c.Writer, resp.Body)
}

// ProxyImageWithCache 带缓存键的图片代理（用于不同尺寸）
// GET /api/v1/image/proxy/:cache_key?url=https://example.com/image.jpg
func (h *ImageProxyHandler) ProxyImageWithCache(c *gin.Context) {
	// 复用普通代理逻辑，cache_key仅用于URL区分，让CF分别缓存
	h.ProxyImage(c)
}
