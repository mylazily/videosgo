package service

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/mylazily/videosgo/internal/model"
	"github.com/mylazily/videosgo/internal/repository"
)

// DomainRotationService 域名轮询服务
type DomainRotationService struct {
	repo    *repository.DomainRotationRepo
	mu      sync.RWMutex
	stopCh  chan struct{}
	running bool
}

// NewDomainRotationService 创建域名轮询服务
func NewDomainRotationService(repo *repository.DomainRotationRepo) *DomainRotationService {
	return &DomainRotationService{
		repo:   repo,
		stopCh: make(chan struct{}),
	}
}

// CheckDomain 检查域名可用性
func (s *DomainRotationService) CheckDomain(domain, region string) (bool, int, string, error) {
	start := time.Now()
	client := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	// Ensure domain has protocol
	checkURL := domain
	if !strings.HasPrefix(strings.ToLower(checkURL), "http") {
		checkURL = "https://" + checkURL
	}
	resp, err := client.Get(checkURL)
	responseTimeMs := int(time.Since(start).Milliseconds())

	if err != nil {
		return false, responseTimeMs, fmt.Sprintf("连接失败: %v", err), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		return false, responseTimeMs, fmt.Sprintf("服务器错误: HTTP %d", resp.StatusCode), nil
	}

	if resp.StatusCode >= 400 {
		return false, responseTimeMs, fmt.Sprintf("客户端错误: HTTP %d", resp.StatusCode), nil
	}

	return true, responseTimeMs, "", nil
}

// RotateDomain 自动轮询切换域名
func (s *DomainRotationService) RotateDomain(region string) error {
	// 获取当前活跃域名
	current, err := s.repo.GetActiveDomain(region)
	if err != nil {
		return fmt.Errorf("获取当前活跃域名失败: %w", err)
	}

	// 检查当前域名可用性
	accessible, responseTimeMs, errorType, err := s.CheckDomain(current.Domain, region)
	if err != nil {
		log.Printf("[域名轮询] 检查域名 %s 失败: %v", current.Domain, err)
	}

	// 记录可用性
	_ = s.repo.UpdateAvailability(current.ID, region, accessible, responseTimeMs, errorType)

	// 如果当前域名不可用，尝试切换
	if !accessible {
		log.Printf("[域名轮询] 域名 %s 不可用，尝试切换...", current.Domain)
		bestDomain, err := s.GetBestDomain(region)
		if err != nil {
			return fmt.Errorf("获取最优域名失败: %w", err)
		}

		if bestDomain != current.Domain {
			// 执行切换
			err = s.repo.SetActiveDomain(bestDomain, region, "auto")
			if err != nil {
				return fmt.Errorf("设置活跃域名失败: %w", err)
			}
			// 记录切换事件
			_ = s.repo.LogSwitch(current.Domain, bestDomain, "自动切换: 原域名不可用")
			log.Printf("[域名轮询] 已切换域名: %s -> %s", current.Domain, bestDomain)
		}
	}

	return nil
}

// GetBestDomain 获取最优域名（响应时间最快且可用）
func (s *DomainRotationService) GetBestDomain(region string) (string, error) {
	// 获取当前活跃域名作为候选
	current, err := s.repo.GetActiveDomain(region)
	if err == nil && current != nil {
		// 检查当前域名是否可用
		accessible, _, _, _ := s.CheckDomain(current.Domain, region)
		if accessible {
			return current.Domain, nil
		}
	}

	// 当前域名不可用，尝试从域名池中获取其他域名
	type DomainCandidate struct {
		Domain       string
		ResponseTime int
		IsAccessible bool
	}

	// 尝试获取域名列表
	actives, availabilities, err := s.repo.GetDomainList()
	if err != nil {
		return "", fmt.Errorf("获取域名列表失败: %w", err)
	}

	// 构建域名到最新可用性的映射
	latestAvail := make(map[string]*model.DomainAvailability)
	for i := range availabilities {
		a := &availabilities[i]
		key := a.DomainID.String()
		if _, exists := latestAvail[key]; !exists {
			latestAvail[key] = a
		}
	}

	// 检查活跃域名
	var candidates []DomainCandidate
	for _, active := range actives {
		candidates = append(candidates, DomainCandidate{
			Domain: active.Domain,
		})
	}

	// 如果没有候选域名，返回错误
	if len(candidates) == 0 {
		return "", fmt.Errorf("没有可用的候选域名")
	}

	// 并发检查所有候选域名
	var wg sync.WaitGroup
	results := make(chan DomainCandidate, len(candidates))

	for _, candidate := range candidates {
		wg.Add(1)
		go func(domain string) {
			defer wg.Done()
			accessible, responseTime, _, _ := s.CheckDomain(domain, region)
			results <- DomainCandidate{
				Domain:       domain,
				ResponseTime: responseTime,
				IsAccessible: accessible,
			}
		}(candidate.Domain)
	}

	wg.Wait()
	close(results)

	// 找到响应时间最快的可用域名
	var bestDomain string
	bestResponseTime := int(1<<63 - 1) // max int

	for result := range results {
		if result.IsAccessible && result.ResponseTime < bestResponseTime {
			bestResponseTime = result.ResponseTime
			bestDomain = result.Domain
		}
	}

	if bestDomain == "" {
		return "", fmt.Errorf("所有域名均不可用")
	}

	return bestDomain, nil
}

// StartAutoRotation 启动自动轮询
func (s *DomainRotationService) StartAutoRotation() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.stopCh = make(chan struct{})
	s.mu.Unlock()

	log.Println("[域名轮询] 自动轮询已启动（每分钟检查一次）")

	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		// 首次延迟 10 秒执行
		time.Sleep(10 * time.Second)
		_ = s.RotateDomain("")

		for {
			select {
			case <-ticker.C:
				_ = s.RotateDomain("")
			case <-s.stopCh:
				log.Println("[域名轮询] 自动轮询已停止")
				return
			}
		}
	}()
}

// StopAutoRotation 停止自动轮询
func (s *DomainRotationService) StopAutoRotation() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running {
		close(s.stopCh)
		s.running = false
	}
}

// GetActiveDomain 获取当前活跃域名
func (s *DomainRotationService) GetActiveDomain(region string) (*string, error) {
	domain, err := s.repo.GetActiveDomain(region)
	if err != nil {
		return nil, err
	}
	return &domain.Domain, nil
}

// SwitchDomain 手动切换域名
func (s *DomainRotationService) SwitchDomain(domain, region, reason string) error {
	// 获取当前域名
	current, err := s.repo.GetActiveDomain(region)
	if err != nil {
		return fmt.Errorf("获取当前域名失败: %w", err)
	}

	// 设置新域名
	err = s.repo.SetActiveDomain(domain, region, "manual")
	if err != nil {
		return fmt.Errorf("设置域名失败: %w", err)
	}

	// 记录切换事件
	_ = s.repo.LogSwitch(current.Domain, domain, reason)
	log.Printf("[域名轮询] 手动切换域名: %s -> %s (原因: %s)", current.Domain, domain, reason)

	return nil
}

// GetDomainList 获取域名列表及状态
func (s *DomainRotationService) GetDomainList() (interface{}, error) {
	actives, availabilities, err := s.repo.GetDomainList()
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"active":        actives,
		"availabilities": availabilities,
	}, nil
}

// GetSwitchHistory 获取切换历史
func (s *DomainRotationService) GetSwitchHistory(limit int) (interface{}, error) {
	return s.repo.GetSwitchHistory(limit)
}
// GetSwitchHistory 获取切换历史