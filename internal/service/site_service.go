package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"videosgo/internal/database"
	"videosgo/internal/model"
	"videosgo/internal/repository"
)

// SiteService 站群管理服务
type SiteService struct {
	repo *repository.SiteRepo
}

// NewSiteService 创建站群管理服务
func NewSiteService(repo *repository.SiteRepo) *SiteService {
	return &SiteService{repo: repo}
}

// CreateSite 创建域名
func (s *SiteService) CreateSite(site *model.SiteDomain) error {
	return s.repo.Create(site)
}

// UpdateSite 更新域名
func (s *SiteService) UpdateSite(site *model.SiteDomain) error {
	// 清除缓存
	s.clearSiteCache()
	return s.repo.Update(site)
}

// DeleteSite 删除域名
func (s *SiteService) DeleteSite(id string) error {
	uid, err := parseUUID(id)
	if err != nil {
		return fmt.Errorf("无效的域名 ID: %w", err)
	}
	s.clearSiteCache()
	return s.repo.Delete(uid)
}

// GetSite 获取域名详情
func (s *SiteService) GetSite(id string) (*model.SiteDomain, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return nil, fmt.Errorf("无效的域名 ID: %w", err)
	}
	return s.repo.GetByID(uid)
}

// GetSiteByDomain 根据域名获取
func (s *SiteService) GetSiteByDomain(domain string) (*model.SiteDomain, error) {
	// 尝试从缓存获取
	cacheKey := fmt.Sprintf("site:domain:%s", domain)
	if database.RDB != nil {
		cached, err := database.RDB.Get(context.Background(), cacheKey).Result()
		if err == nil && cached != "" {
			var site model.SiteDomain
			if err := json.Unmarshal([]byte(cached), &site); err == nil {
				return &site, nil
			}
		}
	}

	site, err := s.repo.GetByDomain(domain)
	if err != nil {
		return nil, err
	}

	// 写入缓存（30 分钟）
	if database.RDB != nil {
		data, _ := json.Marshal(site)
		database.RDB.Set(context.Background(), cacheKey, data, 30*time.Minute)
	}

	return site, nil
}

// ListSites 获取域名列表（带缓存）
func (s *SiteService) ListSites(page, pageSize int, cluster string) ([]model.SiteDomain, int64, error) {
	cacheKey := fmt.Sprintf("site:list:%s:%d:%d", cluster, page, pageSize)

	// 尝试从缓存获取
	if database.RDB != nil {
		cached, err := database.RDB.Get(context.Background(), cacheKey).Result()
		if err == nil && cached != "" {
			var result struct {
				List  []model.SiteDomain `json:"list"`
				Total int64              `json:"total"`
			}
			if err := json.Unmarshal([]byte(cached), &result); err == nil {
				return result.List, result.Total, nil
			}
		}
	}

	sites, total, err := s.repo.List(page, pageSize, cluster, false)
	if err != nil {
		return nil, 0, err
	}

	// 写入缓存（10 分钟）
	if database.RDB != nil {
		data, _ := json.Marshal(map[string]interface{}{
			"list":  sites,
			"total": total,
		})
		database.RDB.Set(context.Background(), cacheKey, data, 10*time.Minute)
	}

	return sites, total, nil
}

// HealthCheckAll 批量健康检查所有活跃域名
func (s *SiteService) HealthCheckAll() {
	sites, err := s.repo.GetAllActiveDomains()
	if err != nil {
		return
	}

	var wg sync.WaitGroup
	// 限制并发数
	semaphore := make(chan struct{}, 10)

	for i := range sites {
		wg.Add(1)
		semaphore <- struct{}{}
		go func(site model.SiteDomain) {
			defer wg.Done()
			defer func() { <-semaphore }()

			s.healthCheckSingle(site)
		}(sites[i])
	}

	wg.Wait()
}

// healthCheckSingle 对单个域名执行健康检查
func (s *SiteService) healthCheckSingle(site model.SiteDomain) {
	client := &http.Client{Timeout: 10 * time.Second}
	url := fmt.Sprintf("https://%s", site.Domain)

	startTime := time.Now()
	resp, err := client.Get(url)
	responseTimeMs := int(time.Since(startTime).Milliseconds())

	healthLog := &model.SiteHealthLog{
		DomainID:       site.ID,
		ResponseTimeMs: responseTimeMs,
		CheckedAt:      time.Now(),
	}

	if err != nil {
		healthLog.ErrorMessage = err.Error()
		healthLog.StatusCode = 0
		healthLog.IsSSLValid = false
		_ = s.repo.UpdateHealthStatus(site.ID, "unhealthy", responseTimeMs)
	} else {
		defer resp.Body.Close()
		healthLog.StatusCode = resp.StatusCode
		healthLog.IsSSLValid = resp.TLS != nil

		if resp.StatusCode >= 200 && resp.StatusCode < 400 {
			_ = s.repo.UpdateHealthStatus(site.ID, "healthy", responseTimeMs)
		} else {
			_ = s.repo.UpdateHealthStatus(site.ID, "unhealthy", responseTimeMs)
		}
	}

	_ = s.repo.CreateHealthLog(healthLog)
}

// HealthCheckSingle 对单个域名执行手动健康检查
func (s *SiteService) HealthCheckSingle(id string) error {
	uid, err := parseUUID(id)
	if err != nil {
		return fmt.Errorf("无效的域名 ID: %w", err)
	}

	site, err := s.repo.GetByID(uid)
	if err != nil {
		return fmt.Errorf("域名不存在: %w", err)
	}

	s.healthCheckSingle(*site)
	return nil
}

// GetRedirectTarget 获取 301 重定向目标
func (s *SiteService) GetRedirectTarget(domain, path, ua, ip string) (string, bool) {
	site, err := s.GetSiteByDomain(domain)
	if err != nil || !site.RedirectEnabled || site.RedirectTarget == "" {
		return "", false
	}

	// 构建完整重定向 URL
	target := site.RedirectTarget
	if !strings.HasSuffix(target, "/") && !strings.HasPrefix(path, "/") {
		target += "/"
	}
	target += path

	return target, true
}

// IncrementTraffic 增加域名流量计数
func (s *SiteService) IncrementTraffic(domain string) {
	site, err := s.GetSiteByDomain(domain)
	if err != nil {
		return
	}
	_ = s.repo.IncrementTraffic(site.ID)
}

// GetCrossClusterLinks 获取交叉链接审计
func (s *SiteService) GetCrossClusterLinks() ([]model.DomainLinkAudit, error) {
	return s.repo.CheckCrossClusterLinks()
}

// clearSiteCache 清除站点相关缓存
func (s *SiteService) clearSiteCache() {
	if database.RDB == nil {
		return
	}
	ctx := context.Background()
	iter := database.RDB.Scan(ctx, 0, "site:*", 100).Iterator()
	for iter.Next(ctx) {
		database.RDB.Del(ctx, iter.Val())
	}
}
