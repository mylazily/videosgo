package service

import (
	"fmt"

	"github.com/mylazily/videosgo/internal/model"
	"github.com/mylazily/videosgo/internal/repository"
)

// RedirectService 301 重定向引擎服务
type RedirectService struct {
	repo *repository.RedirectRepo
}

// NewRedirectService 创建重定向引擎服务
func NewRedirectService(repo *repository.RedirectRepo) *RedirectService {
	return &RedirectService{repo: repo}
}

// MatchAndRedirect 匹配并记录重定向
func (s *RedirectService) MatchAndRedirect(domain, path, ua, ip string) (*model.RedirectRule, error) {
	rule, err := s.repo.MatchRule(domain, path, ua, ip)
	if err != nil {
		return nil, err
	}

	// 增加命中次数（异步）
	go func() {
		_ = s.repo.IncrementHitCount(rule.ID)
		// 记录命中日志
		hitLog := &model.RedirectHitLog{
			RuleID:       rule.ID,
			SourceDomain: domain,
			SourcePath:   path,
			TargetURL:    rule.TargetURL,
			IPAddress:    ip,
			UserAgent:    ua,
		}
		_ = s.repo.LogHit(hitLog)
	}()

	return rule, nil
}

// CreateRule 创建重定向规则
func (s *RedirectService) CreateRule(rule *model.RedirectRule) error {
	return s.repo.Create(rule)
}

// UpdateRule 更新重定向规则
func (s *RedirectService) UpdateRule(rule *model.RedirectRule) error {
	return s.repo.Update(rule)
}

// DeleteRule 删除重定向规则
func (s *RedirectService) DeleteRule(id string) error {
	uid, err := parseUUID(id)
	if err != nil {
		return fmt.Errorf("无效的规则 ID: %w", err)
	}
	return s.repo.Delete(uid)
}

// GetRule 获取规则详情
func (s *RedirectService) GetRule(id string) (*model.RedirectRule, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return nil, fmt.Errorf("无效的规则 ID: %w", err)
	}
	return s.repo.GetByID(uid)
}

// ListRules 获取规则列表
func (s *RedirectService) ListRules(page, pageSize int) ([]model.RedirectRule, int64, error) {
	return s.repo.List(page, pageSize, false)
}

// ToggleRule 切换规则启用状态
func (s *RedirectService) ToggleRule(id string, active bool) error {
	uid, err := parseUUID(id)
	if err != nil {
		return fmt.Errorf("无效的规则 ID: %w", err)
	}
	return s.repo.ToggleRule(uid, active)
}

// GetHitLogs 获取命中日志
func (s *RedirectService) GetHitLogs(ruleID string, limit int) ([]model.RedirectHitLog, error) {
	uid, err := parseUUID(ruleID)
	if err != nil {
		return nil, fmt.Errorf("无效的规则 ID: %w", err)
	}
	return s.repo.ListHitLogs(uid, limit)
}
