package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"videosgo/internal/service"
)

// SitemapHandler Sitemap 处理器
type SitemapHandler struct {
	svc *service.SitemapService
}

// NewSitemapHandler 创建 Sitemap 处理器
func NewSitemapHandler(svc *service.SitemapService) *SitemapHandler {
	return &SitemapHandler{svc: svc}
}

// GetSitemapIndex 获取 sitemap 索引
// GET /sitemap.xml
func (h *SitemapHandler) GetSitemapIndex(c *gin.Context) {
	xml, err := h.svc.GetAllSitemaps()
	if err != nil {
		c.String(http.StatusInternalServerError, "生成 sitemap 失败")
		return
	}
	c.Data(http.StatusOK, "application/xml; charset=utf-8", []byte(xml))
}

// GetVideoSitemap 获取视频 sitemap
// GET /sitemap-video.xml
func (h *SitemapHandler) GetVideoSitemap(c *gin.Context) {
	xml, err := h.svc.GenerateVideoSitemap()
	if err != nil {
		c.String(http.StatusInternalServerError, "生成视频 sitemap 失败")
		return
	}
	c.Data(http.StatusOK, "application/xml; charset=utf-8", []byte(xml))
}

// GetTagSitemap 获取标签 sitemap
// GET /sitemap-tag.xml
func (h *SitemapHandler) GetTagSitemap(c *gin.Context) {
	xml, err := h.svc.GenerateTagSitemap()
	if err != nil {
		c.String(http.StatusInternalServerError, "生成标签 sitemap 失败")
		return
	}
	c.Data(http.StatusOK, "application/xml; charset=utf-8", []byte(xml))
}

// GetShortVideoSitemap 获取短视频 sitemap
// GET /sitemap-short.xml
func (h *SitemapHandler) GetShortVideoSitemap(c *gin.Context) {
	xml, err := h.svc.GenerateShortVideoSitemap()
	if err != nil {
		c.String(http.StatusInternalServerError, "生成短视频 sitemap 失败")
		return
	}
	c.Data(http.StatusOK, "application/xml; charset=utf-8", []byte(xml))
}

// GetActorSitemap 获取演员索引 sitemap
// GET /sitemap-actor.xml
func (h *SitemapHandler) GetActorSitemap(c *gin.Context) {
	xml, err := h.svc.GenerateActorSitemap()
	if err != nil {
		c.String(http.StatusInternalServerError, "生成演员 sitemap 失败")
		return
	}
	c.Data(http.StatusOK, "application/xml; charset=utf-8", []byte(xml))
}

// GetRobotsTxt 获取动态 robots.txt
// GET /robots.txt
func (h *SitemapHandler) GetRobotsTxt(c *gin.Context) {
	c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte(h.svc.GenerateRobotsTxt()))
}
