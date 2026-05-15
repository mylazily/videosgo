package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/mylazily/videosgo/internal/repository"
)

// GSCService Google Search Console / Bing Sitemap 提交服务
type GSCService struct {
	siteRepo      *repository.SiteRepo
	googleAPIKey  string
	bingAPIKey    string
	client        *http.Client
}

// NewGSCService 创建 GSC 服务
func NewGSCService(siteRepo *repository.SiteRepo, googleAPIKey, bingAPIKey string) *GSCService {
	return &GSCService{
		siteRepo:     siteRepo,
		googleAPIKey: googleAPIKey,
		bingAPIKey:   bingAPIKey,
		client:       &http.Client{Timeout: 30 * time.Second},
	}
}

// SubmitToGoogle 通过 Google Indexing API 提交 sitemap
func (s *GSCService) SubmitToGoogle(domain, sitemapURL string) error {
	if s.googleAPIKey == "" {
		return fmt.Errorf("Google API Key 未配置")
	}

	// 使用 Google Indexing API 提交
	// 注意：生产环境应使用 OAuth2 服务账号认证
	apiURL := fmt.Sprintf("https://indexing.googleapis.com/v3/urlNotifications:publish")

	payload := map[string]interface{}{
		"url":  sitemapURL,
		"type": "URL_UPDATED",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("构建请求失败: %w", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.googleAPIKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("提交到 Google 失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Google API 返回错误 %d: %s", resp.StatusCode, string(respBody))
	}

	fmt.Printf("[GSC] 成功提交 sitemap 到 Google: %s (域名: %s)\n", sitemapURL, domain)
	return nil
}

// SubmitToBing 通过 Bing Webmaster API 提交 sitemap
func (s *GSCService) SubmitToBing(domain, sitemapURL string) error {
	if s.bingAPIKey == "" {
		return fmt.Errorf("Bing API Key 未配置")
	}

	// 使用 Bing Webmaster API 提交 sitemap
	apiURL := fmt.Sprintf("https://ssl.bing.com/webmaster/api.svc/pox/SubmitUrl?apikey=%s&surl=%s",
		s.bingAPIKey, sitemapURL)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("提交到 Bing 失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Bing API 返回错误 %d: %s", resp.StatusCode, string(respBody))
	}

	fmt.Printf("[GSC] 成功提交 sitemap 到 Bing: %s (域名: %s)\n", sitemapURL, domain)
	return nil
}

// SubmitAllSitemaps 提交所有域名的所有 sitemap
func (s *GSCService) SubmitAllSitemaps() {
	sites, err := s.siteRepo.GetAllActiveDomains()
	if err != nil {
		fmt.Printf("[GSC] 获取域名列表失败: %v\n", err)
		return
	}

	// 标准的 sitemap 文件列表
	sitemaps := []string{
		"/sitemap.xml",
		"/sitemap-video.xml",
		"/sitemap-tag.xml",
		"/sitemap-short.xml",
		"/sitemap-actor.xml",
	}

	totalSubmitted := 0
	totalFailed := 0

	for _, site := range sites {
		for _, sm := range sitemaps {
			sitemapURL := fmt.Sprintf("https://%s%s", site.Domain, sm)

			// 提交到 Google
			if s.googleAPIKey != "" {
				if err := s.SubmitToGoogle(site.Domain, sitemapURL); err != nil {
					fmt.Printf("[GSC] Google 提交失败 (%s): %v\n", sitemapURL, err)
					totalFailed++
				} else {
					totalSubmitted++
				}
			}

			// 提交到 Bing
			if s.bingAPIKey != "" {
				if err := s.SubmitToBing(site.Domain, sitemapURL); err != nil {
					fmt.Printf("[GSC] Bing 提交失败 (%s): %v\n", sitemapURL, err)
					totalFailed++
				} else {
					totalSubmitted++
				}
			}
		}
	}

	fmt.Printf("[GSC] Sitemap 提交完成: 成功 %d, 失败 %d\n", totalSubmitted, totalFailed)
}
