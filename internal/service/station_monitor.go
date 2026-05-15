package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// StationStatus 资源站状态
type StationStatus struct {
	Name          string  `json:"name"`
	PingURL       string  `json:"ping_url"`
	IsAlive       bool    `json:"is_alive"`
	Latency       int     `json:"latency_ms"`      // 延迟（毫秒）
	Speed         float64 `json:"speed_kbps"`      // 下载速度（KB/s）
	Weight        int     `json:"weight"`          // 权重（0-100，越高越优先推荐）
	FailCount     int     `json:"fail_count"`      // 连续失败次数
	LastCheckAt   string  `json:"last_check_at"`
	LastSuccessAt string  `json:"last_success_at"`
	Region        string  `json:"region"`          // 地区
}

// StationMonitor 资源站监控服务
type StationMonitor struct {
	stations      []*StationStatus
	mu            sync.RWMutex
	httpClient    *http.Client
	redisClient   *redis.Client // 可以为 nil
	checkInterval time.Duration // 检查间隔，默认 1 分钟
	speedTestSize int           // 测速下载大小（字节），默认 10KB
	stopCh        chan struct{}
	wg            sync.WaitGroup
}

const (
	// maxConcurrentChecks 最大并发检查数
	maxConcurrentChecks = 20
	// defaultCheckInterval 默认检查间隔
	defaultCheckInterval = 1 * time.Minute
	// defaultSpeedTestSize 默认测速下载大小（10KB）
	defaultSpeedTestSize = 10 * 1024
	// httpTimeout HTTP 请求超时时间
	httpTimeout = 3 * time.Second
	// maxFailCount 连续失败多少次标记为 dead
	maxFailCount = 5
	// redisStatusTTL Redis 缓存 TTL
	redisStatusTTL = 5 * time.Minute
)

// NewStationMonitor 创建监控服务
func NewStationMonitor(redisClient *redis.Client) *StationMonitor {
	return &StationMonitor{
		httpClient: &http.Client{
			Timeout: httpTimeout,
			// 不跟随重定向，避免下载到实际内容
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
			Transport: &http.Transport{
				// 禁用连接复用，每次请求独立测量延迟
				DisableKeepAlives: true,
			},
		},
		redisClient:   redisClient,
		checkInterval: defaultCheckInterval,
		speedTestSize: defaultSpeedTestSize,
		stopCh:        make(chan struct{}),
	}
}

// AddStation 添加资源站到监控列表
func (m *StationMonitor) AddStation(name, pingURL, region string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查是否已存在，避免重复添加
	for _, s := range m.stations {
		if s.Name == name {
			log.Printf("[资源站监控] 资源站已存在，跳过添加: %s", name)
			return
		}
	}

	m.stations = append(m.stations, &StationStatus{
		Name:    name,
		PingURL: pingURL,
		Region:  region,
		Weight:  50, // 初始权重
	})

	log.Printf("[资源站监控] 添加资源站: %s (%s) — %s", name, region, pingURL)
}

// RemoveStation 移除资源站
func (m *StationMonitor) RemoveStation(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, s := range m.stations {
		if s.Name == name {
			m.stations = append(m.stations[:i], m.stations[i+1:]...)
			log.Printf("[资源站监控] 移除资源站: %s", name)
			return
		}
	}
}

// CheckAll 检查所有资源站（核心探活函数）
// 对每个资源站：
//  1. 发送 HTTP HEAD 请求，3 秒超时
//  2. 如果 HEAD 不支持，发送 GET 请求（只下载前 10KB）
//  3. 计算延迟和下载速度
//  4. 根据结果更新权重
//  5. 将结果写入 Redis（key: station:{name}:status，TTL: 5 分钟）
//  6. 并发检查（goroutine pool，最大 20 并发）
func (m *StationMonitor) CheckAll(ctx context.Context) {
	m.mu.RLock()
	stations := make([]*StationStatus, len(m.stations))
	copy(stations, m.stations)
	m.mu.RUnlock()

	if len(stations) == 0 {
		return
	}

	log.Printf("[资源站监控] 开始检查 %d 个资源站...", len(stations))

	// 使用 semaphore 控制并发
	sem := make(chan struct{}, maxConcurrentChecks)
	var wg sync.WaitGroup

	for _, station := range stations {
		wg.Add(1)
		sem <- struct{}{} // 获取信号量
		go func(s *StationStatus) {
			defer wg.Done()
			defer func() { <-sem }() // 释放信号量
			m.checkSingle(ctx, s)
		}(station)
	}

	wg.Wait()

	// 统计结果
	m.mu.RLock()
	aliveCount := 0
	for _, s := range m.stations {
		if s.IsAlive {
			aliveCount++
		}
	}
	m.mu.RUnlock()

	log.Printf("[资源站监控] 检查完成: 存活 %d / 总计 %d", aliveCount, len(stations))
}

// checkSingle 检查单个资源站
func (m *StationMonitor) checkSingle(ctx context.Context, station *StationStatus) {
	now := time.Now().Format(time.RFC3339)

	// 记录检查时间
	station.LastCheckAt = now

	// 先尝试 HEAD 请求
	latency, speed, err := m.probeStation(station.PingURL)

	if err != nil {
		// HEAD 失败，尝试 GET 请求（只下载前 N 字节）
		latency, speed, err = m.probeStationGET(station.PingURL)
	}

	if err != nil {
		// 探活失败
		station.IsAlive = false
		station.Latency = 0
		station.Speed = 0
		station.Weight = 0
		station.FailCount++

		log.Printf("[资源站监控] %s 探活失败 (连续失败 %d 次): %v", station.Name, station.FailCount, err)

		// 连续失败超过阈值，标记为 dead
		if station.FailCount >= maxFailCount {
			log.Printf("[资源站监控] %s 连续失败 %d 次，标记为 dead", station.Name, station.FailCount)
		}
	} else {
		// 探活成功
		station.IsAlive = true
		station.Latency = latency
		station.Speed = speed
		station.Weight = calculateWeight(latency, speed, station.FailCount)
		station.FailCount = 0 // 恢复成功，归零失败计数
		station.LastSuccessAt = now

		log.Printf("[资源站监控] %s 存活 — 延迟: %dms, 速度: %.1fKB/s, 权重: %d",
			station.Name, latency, speed, station.Weight)
	}

	// 写入 Redis 缓存
	m.cacheStationStatus(ctx, station)
}

// probeStation 使用 HEAD 请求探测资源站
func (m *StationMonitor) probeStation(url string) (latency int, speed float64, err error) {
	start := time.Now()
	resp, err := m.httpClient.Head(url)
	latency = int(time.Since(start).Milliseconds())

	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	// 状态码 2xx/3xx 视为存活
	if resp.StatusCode >= 400 {
		return 0, 0, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	// HEAD 请求成功，尝试测速
	speed, _ = measureSpeed(m.httpClient, url, m.speedTestSize)

	return latency, speed, nil
}

// probeStationGET 使用 GET 请求探测资源站（HEAD 不支持时的降级方案）
func (m *StationMonitor) probeStationGET(url string) (latency int, speed float64, err error) {
	// 创建一个限制读取的 GET 客户端
	start := time.Now()
	resp, err := m.httpClient.Get(url)
	latency = int(time.Since(start).Milliseconds())

	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return 0, 0, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	// 读取前 N 字节用于测速
	speed, _ = measureSpeedFromReader(resp.Body, m.speedTestSize, time.Since(start))

	return latency, speed, nil
}

// cacheStationStatus 将资源站状态写入 Redis
func (m *StationMonitor) cacheStationStatus(ctx context.Context, station *StationStatus) {
	if m.redisClient == nil {
		return
	}

	key := fmt.Sprintf("station:%s:status", station.Name)
	data, err := json.Marshal(station)
	if err != nil {
		log.Printf("[资源站监控] 序列化状态失败 (%s): %v", station.Name, err)
		return
	}

	if err := m.redisClient.Set(ctx, key, data, redisStatusTTL).Err(); err != nil {
		log.Printf("[资源站监控] 写入 Redis 失败 (%s): %v", station.Name, err)
	}
}

// GetStatus 获取所有资源站状态
func (m *StationMonitor) GetStatus() []*StationStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*StationStatus, len(m.stations))
	copy(result, m.stations)
	return result
}

// GetAliveStations 获取存活的资源站（按权重排序）
func (m *StationMonitor) GetAliveStations() []*StationStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var alive []*StationStatus
	for _, s := range m.stations {
		if s.IsAlive {
			alive = append(alive, s)
		}
	}

	// 按权重降序排序
	sort.Slice(alive, func(i, j int) bool {
		return alive[i].Weight > alive[j].Weight
	})

	return alive
}

// GetBestStation 获取最优资源站（权重最高）
func (m *StationMonitor) GetBestStation() *StationStatus {
	alive := m.GetAliveStations()
	if len(alive) == 0 {
		return nil
	}
	return alive[0]
}

// GetStationStatus 获取单个资源站状态
func (m *StationMonitor) GetStationStatus(name string) *StationStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, s := range m.stations {
		if s.Name == name {
			// 返回副本，避免外部修改
			copy := *s
			return &copy
		}
	}
	return nil
}

// Start 启动定时监控
func (m *StationMonitor) Start(ctx context.Context) {
	log.Printf("[资源站监控] 启动定时监控（间隔: %v）", m.checkInterval)

	// 启动时立即执行一次检查
	m.CheckAll(ctx)

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		ticker := time.NewTicker(m.checkInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				m.CheckAll(ctx)
			case <-m.stopCh:
				log.Println("[资源站监控] 定时监控已停止")
				return
			}
		}
	}()
}

// Stop 停止监控
func (m *StationMonitor) Stop() {
	log.Println("[资源站监控] 正在停止...")
	close(m.stopCh)
	m.wg.Wait()
	log.Println("[资源站监控] 已停止")
}

// calculateWeight 根据延迟和速度计算权重
// 权重规则：
//   - 延迟 < 500ms 且速度 > 200KB/s → weight = 100
//   - 延迟 < 1000ms 且速度 > 100KB/s → weight = 80
//   - 延迟 < 2000ms → weight = 50
//   - 延迟 > 2000ms → weight = 20
//   - 失败 → weight = 0, failCount++
//   - 连续失败 5 次 → 标记为 dead，从推荐列表移除
//   - 恢复成功 → failCount 归零，重新纳入推荐
func calculateWeight(latency int, speed float64, failCount int) int {
	// 如果有历史失败记录，适当降低权重
	failPenalty := 0
	if failCount > 0 {
		failPenalty = failCount * 5
	}

	var weight int

	switch {
	case latency < 500 && speed > 200:
		weight = 100
	case latency < 1000 && speed > 100:
		weight = 80
	case latency < 2000:
		weight = 50
	default:
		weight = 20
	}

	// 扣除失败惩罚
	weight -= failPenalty
	if weight < 0 {
		weight = 0
	}

	return weight
}

// measureSpeed 测量下载速度（下载前 N 字节计算速率）
func measureSpeed(client *http.Client, url string, size int) (speed float64, err error) {
	start := time.Now()

	// 创建一个带超时的 context
	ctx, cancel := context.WithTimeout(context.Background(), httpTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return 0, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	return measureSpeedFromReader(resp.Body, size, time.Since(start))
}

// measureSpeedFromReader 从 Reader 测量下载速度
func measureSpeedFromReader(reader io.Reader, size int, elapsed time.Duration) (speed float64, err error) {
	buf := make([]byte, size)
	n, err := io.ReadFull(reader, buf)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		return 0, err
	}

	// 实际读取的字节数
	if n == 0 {
		return 0, nil
	}

	// 总耗时（包含连接时间）
	totalElapsed := elapsed
	if totalElapsed == 0 {
		totalElapsed = time.Millisecond
	}

	// 计算速度: (字节数 / 秒) / 1024 = KB/s
	bytesPerSec := float64(n) / totalElapsed.Seconds()
	speed = bytesPerSec / 1024.0

	return speed, nil
}

// GetStationStatusMap 获取所有资源站状态映射（用于视频 API 增强）
// 返回 map[sourceName]*StationStatus，方便快速查找
func (m *StationMonitor) GetStationStatusMap() map[string]*StationStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]*StationStatus, len(m.stations))
	for _, s := range m.stations {
		result[s.Name] = &StationStatus{
			Name:    s.Name,
			IsAlive: s.IsAlive,
			Latency: s.Latency,
			Speed:   s.Speed,
			Weight:  s.Weight,
		}
	}
	return result
}

// GetStationStatusCompact 获取资源站状态（精简版，用于嵌入视频响应）
type StationStatusCompact struct {
	IsAlive   bool    `json:"is_alive"`
	Latency   int     `json:"latency_ms"`
	Speed     float64 `json:"speed_kbps"`
	Weight    int     `json:"weight"`
}

// GetStationStatusCompactMap 获取精简版状态映射
func (m *StationMonitor) GetStationStatusCompactMap() map[string]StationStatusCompact {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]StationStatusCompact, len(m.stations))
	for _, s := range m.stations {
		result[s.Name] = StationStatusCompact{
			IsAlive: s.IsAlive,
			Latency: s.Latency,
			Speed:   s.Speed,
			Weight:  s.Weight,
		}
	}
	return result
}

// LoadStationsFromRedis 从 Redis 加载资源站状态（启动时恢复）
func (m *StationMonitor) LoadStationsFromRedis(ctx context.Context) {
	if m.redisClient == nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, station := range m.stations {
		key := fmt.Sprintf("station:%s:status", station.Name)
		data, err := m.redisClient.Get(ctx, key).Bytes()
		if err != nil {
			continue
		}

		var cached StationStatus
		if err := json.Unmarshal(data, &cached); err != nil {
			continue
		}

		// 恢复状态
		station.IsAlive = cached.IsAlive
		station.Latency = cached.Latency
		station.Speed = cached.Speed
		station.Weight = cached.Weight
		station.FailCount = cached.FailCount
		station.LastCheckAt = cached.LastCheckAt
		station.LastSuccessAt = cached.LastSuccessAt

		log.Printf("[资源站监控] 从 Redis 恢复状态: %s (alive=%v, weight=%d)",
			station.Name, station.IsAlive, station.Weight)
	}
}

// StationCheckResult 单个资源站检查结果（用于手动触发检查的响应）
type StationCheckResult struct {
	Name    string  `json:"name"`
	IsAlive bool    `json:"is_alive"`
	Latency int     `json:"latency_ms"`
	Speed   float64 `json:"speed_kbps"`
	Weight  int     `json:"weight"`
	Error   string  `json:"error,omitempty"`
}

// CheckAllAndReturnResults 检查所有资源站并返回结果（用于手动触发 API）
func (m *StationMonitor) CheckAllAndReturnResults(ctx context.Context) []StationCheckResult {
	m.CheckAll(ctx)

	statusMap := m.GetStationStatusMap()
	results := make([]StationCheckResult, 0, len(statusMap))

	for name, status := range statusMap {
		result := StationCheckResult{
			Name:    name,
			IsAlive: status.IsAlive,
			Latency: status.Latency,
			Speed:   status.Speed,
			Weight:  status.Weight,
		}
		if !status.IsAlive {
			result.Error = "连接失败"
		}
		results = append(results, result)
	}

	// 按权重降序排序
	sort.Slice(results, func(i, j int) bool {
		return results[i].Weight > results[j].Weight
	})

	return results
}

// Close 实现 io.Closer 接口
func (m *StationMonitor) Close() error {
	m.Stop()
	return nil
}
