package service

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
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

	resp, err := client.Get("https://" + domain)
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

// GetBestDomain 获取最优域名（响应时间最快）
func (s *DomainRotationService) GetBestDomain(region string) (string, error) {
	// TODO: 从域名配置中获取所有候选域名
	// 当前实现返回默认域名
	active, err := s.repo.GetActiveDomain(region)
	if err == nil {
		return active.Domain, nil
	}
	return "default.example.com", nil
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

// GetDomainIDForCheck 获取用于检查的域名 ID（模拟）
func (s *DomainRotationService) GetDomainIDForCheck() uuid.UUID {
	// 返回一个固定的 UUID 用于检查
	return uuid.MustParse("00000000-0000-0000-0000-000000000001")
}
