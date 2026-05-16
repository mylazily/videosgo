package service

import (
	"context"
	"encoding/xml"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/mylazily/videosgo/internal/database"
	"github.com/mylazily/videosgo/internal/model"
)

// SitemapService Sitemap 生成服务
type SitemapService struct {
	db *gorm.DB
}

// NewSitemapService 创建 Sitemap 服务
func NewSitemapService(db *gorm.DB) *SitemapService {
	return &SitemapService{db: db}
}

// SitemapURL Sitemap URL 条目
type SitemapURL struct {
	XMLName    xml.Name `xml:"url"`
	Loc        string   `xml:"loc"`
	LastMod    string   `xml:"lastmod,omitempty"`
	ChangeFreq string   `xml:"changefreq,omitempty"`
	Priority   string   `xml:"priority,omitempty"`
}

// SitemapIndex Sitemap 索引
type SitemapIndex struct {
	XMLName xml.Name `xml:"sitemapindex"`
	Xmlns   string   `xml:"xmlns,attr"`
	Sitemaps []SitemapEntry `xml:"sitemap"`
}

// SitemapEntry Sitemap 索引条目
type SitemapEntry struct {
	Loc     string `xml:"loc"`
	LastMod string `xml:"lastmod,omitempty"`
}

// URLSet URL 集合
type URLSet struct {
	XMLName xml.Name    `xml:"urlset"`
	Xmlns   string      `xml:"xmlns,attr"`
	URLs    []SitemapURL `xml:"url"`
}

// GenerateVideoSitemap 生成视频 sitemap XML
func (s *SitemapService) GenerateVideoSitemap() (string, error) {
	cacheKey := "sitemap:video"

	// 尝试从缓存获取
	if database.RDB != nil {
		cached, err := database.RDB.Get(context.Background(), cacheKey).Result()
		if err == nil && cached != "" {
			return cached, nil
		}
	}

	var videos []model.Video
	err := s.db.Where("status = ?", "active").
		Select("id, updated_at").
		Order("updated_at DESC").
		Find(&videos).Error
	if err != nil {
		return "", fmt.Errorf("查询视频失败: %w", err)
	}

	urlset := URLSet{
		Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9",
	}

	for _, v := range videos {
		urlset.URLs = append(urlset.URLs, SitemapURL{
			Loc:        fmt.Sprintf("https://videosgo.com/video/%s", v.ID),
			LastMod:    v.UpdatedAt.Format("2006-01-02"),
			ChangeFreq: "daily",
			Priority:   "0.8",
		})
	}

	output, err := xml.MarshalIndent(urlset, "", "  ")
	if err != nil {
		return "", fmt.Errorf("生成 XML 失败: %w", err)
	}

	xmlStr := xml.Header + string(output)

	// 写入缓存（4 小时）
	if database.RDB != nil {
		database.RDB.Set(context.Background(), cacheKey, xmlStr, 4*time.Hour)
	}

	return xmlStr, nil
}

// GenerateTagSitemap 生成标签 sitemap
func (s *SitemapService) GenerateTagSitemap() (string, error) {
	cacheKey := "sitemap:tag"

	// 尝试从缓存获取
	if database.RDB != nil {
		cached, err := database.RDB.Get(context.Background(), cacheKey).Result()
		if err == nil && cached != "" {
			return cached, nil
		}
	}

	var tags []model.Tag
	err := s.db.Where("status = ?", "active").
		Select("slug, updated_at").
		Order("video_count DESC").
		Find(&tags).Error
	if err != nil {
		return "", fmt.Errorf("查询标签失败: %w", err)
	}

	urlset := URLSet{
		Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9",
	}

	for _, t := range tags {
		urlset.URLs = append(urlset.URLs, SitemapURL{
			Loc:        fmt.Sprintf("https://videosgo.com/tag/%s", t.Slug),
			LastMod:    t.UpdatedAt.Format("2006-01-02"),
			ChangeFreq: "weekly",
			Priority:   "0.6",
		})
	}

	output, err := xml.MarshalIndent(urlset, "", "  ")
	if err != nil {
		return "", fmt.Errorf("生成 XML 失败: %w", err)
	}

	xmlStr := xml.Header + string(output)

	// 写入缓存（4 小时）
	if database.RDB != nil {
		database.RDB.Set(context.Background(), cacheKey, xmlStr, 4*time.Hour)
	}

	return xmlStr, nil
}

// GenerateShortVideoSitemap 生成短视频 sitemap
func (s *SitemapService) GenerateShortVideoSitemap() (string, error) {
	cacheKey := "sitemap:short"

	// 尝试从缓存获取
	if database.RDB != nil {
		cached, err := database.RDB.Get(context.Background(), cacheKey).Result()
		if err == nil && cached != "" {
			return cached, nil
		}
	}

	var shorts []model.ShortVideo
	err := s.db.Where("status = ?", "active").
		Select("id, updated_at").
		Order("updated_at DESC").
		Find(&shorts).Error
	if err != nil {
		return "", fmt.Errorf("查询短视频失败: %w", err)
	}

	urlset := URLSet{
		Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9",
	}

	for _, sv := range shorts {
		urlset.URLs = append(urlset.URLs, SitemapURL{
			Loc:        fmt.Sprintf("https://videosgo.com/short/%s", sv.ID),
			LastMod:    sv.UpdatedAt.Format("2006-01-02"),
			ChangeFreq: "daily",
			Priority:   "0.7",
		})
	}

	output, err := xml.MarshalIndent(urlset, "", "  ")
	if err != nil {
		return "", fmt.Errorf("生成 XML 失败: %w", err)
	}

	xmlStr := xml.Header + string(output)

	// 写入缓存（4 小时）
	if database.RDB != nil {
		database.RDB.Set(context.Background(), cacheKey, xmlStr, 4*time.Hour)
	}

	return xmlStr, nil
}

// GenerateActorSitemap 生成演员索引 sitemap
func (s *SitemapService) GenerateActorSitemap() (string, error) {
	cacheKey := "sitemap:actor"

	// 尝试从缓存获取
	if database.RDB != nil {
		cached, err := database.RDB.Get(context.Background(), cacheKey).Result()
		if err == nil && cached != "" {
			return cached, nil
		}
	}

	// 从视频中提取不重复的演员
	type ActorResult struct {
		Actors    string
		UpdatedAt time.Time
	}
	var actors []ActorResult
	err := s.db.Table("videos").
		Where("status = ? AND actors != ''", "active").
		Select("actors, updated_at").
		Find(&actors).Error
	if err != nil {
		return "", fmt.Errorf("查询演员失败: %w", err)
	}

	// 去重
	seen := make(map[string]bool)
	urlset := URLSet{
		Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9",
	}

	for _, a := range actors {
		actorList := strings.Split(a.Actors, ",")
		for _, actor := range actorList {
			actor = strings.TrimSpace(actor)
			if actor == "" || seen[actor] {
				continue
			}
			seen[actor] = true
			urlset.URLs = append(urlset.URLs, SitemapURL{
				Loc:        fmt.Sprintf("https://videosgo.com/actor/%s", actor),
				LastMod:    a.UpdatedAt.Format("2006-01-02"),
				ChangeFreq: "weekly",
				Priority:   "0.5",
			})
		}
	}

	output, err := xml.MarshalIndent(urlset, "", "  ")
	if err != nil {
		return "", fmt.Errorf("生成 XML 失败: %w", err)
	}

	xmlStr := xml.Header + string(output)

	// 写入缓存（4 小时）
	if database.RDB != nil {
		database.RDB.Set(context.Background(), cacheKey, xmlStr, 4*time.Hour)
	}

	return xmlStr, nil
}

// GetAllSitemaps 返回所有 sitemap URL 列表
func (s *SitemapService) GetAllSitemaps() (string, error) {
	cacheKey := "sitemap:index"

	// 尝试从缓存获取
	if database.RDB != nil {
		cached, err := database.RDB.Get(context.Background(), cacheKey).Result()
		if err == nil && cached != "" {
			return cached, nil
		}
	}

	now := time.Now().Format("2006-01-02")

	index := SitemapIndex{
		Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9",
		Sitemaps: []SitemapEntry{
			{Loc: "https://videosgo.com/sitemap-video.xml", LastMod: now},
			{Loc: "https://videosgo.com/sitemap-tag.xml", LastMod: now},
			{Loc: "https://videosgo.com/sitemap-short.xml", LastMod: now},
			{Loc: "https://videosgo.com/sitemap-actor.xml", LastMod: now},
		},
	}

	output, err := xml.MarshalIndent(index, "", "  ")
	if err != nil {
		return "", fmt.Errorf("生成 XML 失败: %w", err)
	}

	xmlStr := xml.Header + string(output)

	// 写入缓存（4 小时）
	if database.RDB != nil {
		database.RDB.Set(context.Background(), cacheKey, xmlStr, 4*time.Hour)
	}

	return xmlStr, nil
}

// GenerateRobotsTxt 生成动态 robots.txt
func (s *SitemapService) GenerateRobotsTxt() string {
	return `User-agent: *
Allow: /
Disallow: /api/
Disallow: /admin/

Sitemap: https://videosgo.com/sitemap.xml
Sitemap: https://videosgo.com/sitemap-video.xml
Sitemap: https://videosgo.com/sitemap-tag.xml
Sitemap: https://videosgo.com/sitemap-short.xml
Sitemap: https://videosgo.com/sitemap-actor.xml
`
}
