package collector

import (
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/mylazily/videosgo/internal/model"
	"github.com/mylazily/videosgo/internal/repository"
)

// Scheduler 采集调度器
type Scheduler struct {
	repo    *repository.CollectRepo
	worker  *Worker
	sources map[uuid.UUID]*sourceSchedule
	mu      sync.RWMutex
	stopCh  chan struct{}
	wg      sync.WaitGroup
}

// sourceSchedule 采集源调度信息
type sourceSchedule struct {
	source   *model.CollectSource
	interval time.Duration
	ticker   *time.Ticker
}

// NewScheduler 创建调度器
func NewScheduler(repo *repository.CollectRepo, worker *Worker) *Scheduler {
	return &Scheduler{
		repo:    repo,
		worker:  worker,
		sources: make(map[uuid.UUID]*sourceSchedule),
		stopCh:  make(chan struct{}),
	}
}

// Start 启动调度器
func (s *Scheduler) Start() error {
	// 获取所有启用的采集源
	sources, err := s.repo.GetEnabled()
	if err != nil {
		return err
	}

	log.Printf("[调度器] 启动，发现 %d 个启用的采集源", len(sources))

	for _, source := range sources {
		s.addSource(&source)
	}

	// 启动时立即执行一次全量采集
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.runInitialFullCollect()
	}()

	return nil
}

// Stop 停止调度器
func (s *Scheduler) Stop() {
	log.Println("[调度器] 正在停止...")
	close(s.stopCh)

	s.mu.Lock()
	defer s.mu.Unlock()

	for id, ss := range s.sources {
		ss.ticker.Stop()
		delete(s.sources, id)
	}

	s.wg.Wait()
	log.Println("[调度器] 已停止")
}

// AddSource 动态添加采集源
func (s *Scheduler) AddSource(source *model.CollectSource) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.addSource(source)
}

// RemoveSource 动态移除采集源
func (s *Scheduler) RemoveSource(sourceID uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if ss, ok := s.sources[sourceID]; ok {
		ss.ticker.Stop()
		delete(s.sources, sourceID)
		log.Printf("[调度器] 移除采集源: %s (ID: %s)", ss.source.Name, sourceID)
	}
}

// addSource 添加采集源到调度器
func (s *Scheduler) addSource(source *model.CollectSource) {
	interval := time.Duration(source.Interval) * time.Minute
	if interval < time.Minute {
		interval = time.Minute
	}

	ss := &sourceSchedule{
		source:   source,
		interval: interval,
		ticker:   time.NewTicker(interval),
	}

	s.sources[source.ID] = ss

	// 启动定时采集 goroutine
	s.wg.Add(1)
	go func(src *model.CollectSource, ticker *time.Ticker) {
		defer s.wg.Done()
		for {
			select {
			case <-ticker.C:
				s.runCollect(src, true) // 定时采集为增量
			case <-s.stopCh:
				return
			}
		}
	}(source, ss.ticker)

	log.Printf("[调度器] 添加采集源: %s (间隔: %v)", source.Name, interval)
}

// runInitialFullCollect 启动时执行全量采集
func (s *Scheduler) runInitialFullCollect() {
	s.mu.RLock()
	sources := make([]*model.CollectSource, 0, len(s.sources))
	for _, ss := range s.sources {
		sources = append(sources, ss.source)
	}
	s.mu.RUnlock()

	for _, source := range sources {
		s.runCollect(source, false) // 全量采集
	}
}

// runCollect 执行采集任务
func (s *Scheduler) runCollect(source *model.CollectSource, incremental bool) {
	log.Printf("[调度器] 开始采集: %s (类型: %s)", source.Name, map[bool]string{true: "增量", false: "全量"}[incremental])

	// 创建采集日志
	collectType := "full"
	if incremental {
		collectType = "incremental"
	}

	collectLog := &model.CollectLog{
		SourceID:   source.ID,
		SourceName: source.Name,
		Type:       collectType,
		Status:     "running",
	}
	_ = s.repo.CreateLog(collectLog)

	startTime := time.Now()

	result, err := s.worker.Collect(source, incremental)

	duration := int(time.Since(startTime).Seconds())
	collectLog.Duration = duration
	collectLog.TotalCount = result.Total
	collectLog.NewCount = result.New
	collectLog.UpdateCount = result.Updated
	collectLog.FailCount = result.Failed

	if err != nil {
		collectLog.Status = "failed"
		collectLog.ErrorMessage = err.Error()
		log.Printf("[调度器] 采集失败: %s - %v", source.Name, err)
	} else {
		collectLog.Status = "success"
		log.Printf("[调度器] 采集完成: %s (新增=%d, 更新=%d, 失败=%d)",
			source.Name, result.New, result.Updated, result.Failed)
	}

	_ = s.repo.UpdateLog(collectLog)
	_ = s.repo.UpdateLastCollect(source.ID)
}
