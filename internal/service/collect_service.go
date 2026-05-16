// Package service 业务逻辑层
package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mylazily/videosgo/internal/collector"
	"github.com/mylazily/videosgo/internal/model"
	"github.com/mylazily/videosgo/internal/repository"
)

// CollectService 采集源服务
type CollectService struct {
	repo     *repository.CollectRepo
	videoRepo *repository.VideoRepo
	worker   *collector.Worker
}

// NewCollectService 创建采集源服务
func NewCollectService(repo *repository.CollectRepo, videoRepo *repository.VideoRepo, worker *collector.Worker) *CollectService {
	return &CollectService{
		repo:      repo,
		videoRepo: videoRepo,
		worker:    worker,
	}
}

// CreateSource 创建采集源
func (s *CollectService) CreateSource(source *model.CollectSource) error {
	if source.Name == "" {
		return fmt.Errorf("采集源名称不能为空")
	}
	if source.APIURL == "" {
		return fmt.Errorf("API 地址不能为空")
	}
	if source.Interval <= 0 {
		source.Interval = 60
	}
	return s.repo.Create(source)
}

// UpdateSource 更新采集源
func (s *CollectService) UpdateSource(source *model.CollectSource) error {
	existing, err := s.repo.GetByID(source.ID)
	if err != nil {
		return fmt.Errorf("采集源不存在")
	}
	source.CreatedAt = existing.CreatedAt
	return s.repo.Update(source)
}

// DeleteSource 删除采集源
func (s *CollectService) DeleteSource(id string) error {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid UUID: %s", id)
	}
	return s.repo.Delete(parsedID)
}

// GetSource 获取采集源详情
func (s *CollectService) GetSource(id string) (*model.CollectSource, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid UUID: %s", id)
	}
	return s.repo.GetByID(parsedID)
}

// ListSources 获取采集源列表
func (s *CollectService) ListSources(page, pageSize int) ([]model.CollectSource, int64, error) {
	return s.repo.List(page, pageSize)
}

// TriggerCollect 手动触发采集
func (s *CollectService) TriggerCollect(sourceID string, collectType string) error {
	parsedID, err := uuid.Parse(sourceID)
	if err != nil {
		return fmt.Errorf("invalid UUID: %s", sourceID)
	}
	source, err := s.repo.GetByID(parsedID)
	if err != nil {
		return fmt.Errorf("采集源不存在")
	}
	if !source.Enabled {
		return fmt.Errorf("采集源已禁用")
	}

	// 创建采集日志
	collectLog := &model.CollectLog{
		SourceID:   source.ID,
		SourceName: source.Name,
		Type:       collectType,
		Status:     "running",
	}
	if err := s.repo.CreateLog(collectLog); err != nil {
		return fmt.Errorf("创建采集日志失败: %w", err)
	}

	// 异步执行采集
	go func() {
		startTime := time.Now()

		result, err := s.worker.Collect(source, collectType == "incremental")

		duration := int(time.Since(startTime).Seconds())
		collectLog.Duration = duration
		collectLog.TotalCount = result.Total
		collectLog.NewCount = result.New
		collectLog.UpdateCount = result.Updated
		collectLog.FailCount = result.Failed

		if err != nil {
			collectLog.Status = "failed"
			collectLog.ErrorMessage = err.Error()
		} else {
			collectLog.Status = "success"
		}

		_ = s.repo.UpdateLog(collectLog)
		_ = s.repo.UpdateLastCollect(parsedID)
	}()

	return nil
}

// ListLogs 获取采集日志列表
func (s *CollectService) ListLogs(page, pageSize int) ([]model.CollectLog, int64, error) {
	return s.repo.ListLogs(page, pageSize)
}
