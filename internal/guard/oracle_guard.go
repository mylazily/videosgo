// Package guard Oracle ARM 服务器自动控温守护
package guard

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

// OracleGuardConfig 守护配置
type OracleGuardConfig struct {
	// CPU 阈值
	IdleCPUThreshold   float64 // 低于此值启动 lookbusy（默认 15%）
	BusyCPUThreshold   float64 // 高于此值关闭 lookbusy（默认 30%）

	// 内存阈值
	IdleMemThreshold   float64 // 内存低于此比例启动 lookbusy（默认 40%）
	MaxMemUsage        float64 // 内存高于此比例强制关闭 lookbusy（默认 70%）

	// lookbusy 参数
	LookbusyCPU        string  // lookbusy 消耗的 CPU（默认 "15"）
	LookbusyMem        string  // lookbusy 消耗的内存（默认 "6GB"）
	LookbusyNice       int     // lookbusy 的 nice 值（默认 19，最低优先级）

	// 检查间隔
	CheckInterval      time.Duration // 检查间隔（默认 5 分钟）

	// 是否启用
	Enabled            bool    // 是否启用守护（默认 true）
}

// DefaultOracleGuardConfig 默认配置（针对 4核24G ARM 优化）
func DefaultOracleGuardConfig() OracleGuardConfig {
	return OracleGuardConfig{
		IdleCPUThreshold:   15.0,  // CPU < 15% 时启动 lookbusy
		BusyCPUThreshold:   30.0,  // CPU > 30% 时关闭 lookbusy
		IdleMemThreshold:   40.0,  // 内存使用 < 40% 时启动 lookbusy
		MaxMemUsage:        70.0,  // 内存使用 > 70% 时强制关闭 lookbusy
		LookbusyCPU:        "15",  // lookbusy 吃 15% CPU
		LookbusyMem:        "6GB", // lookbusy 吃 6GB 内存
		LookbusyNice:       19,    // 最低调度优先级
		CheckInterval:      5 * time.Minute,
		Enabled:            true,
	}
}

// OracleGuard 甲骨文自动控温守护
type OracleGuard struct {
	config       OracleGuardConfig
	lookbusyCmd  *exec.Cmd
	running      bool
	stopChan     chan struct{}
	wg           sync.WaitGroup
	mu           sync.Mutex

	// 统计
	lastCPU      float64
	lastMemUsed  float64
	startCount   int
	stopCount    int
}

// NewOracleGuard 创建守护实例
func NewOracleGuard(cfg OracleGuardConfig) *OracleGuard {
	return &OracleGuard{
		config:   cfg,
		stopChan: make(chan struct{}),
	}
}

// Start 启动守护协程
func (g *OracleGuard) Start(ctx context.Context) {
	if !g.config.Enabled {
		log.Println("[Oracle守护] 已禁用，跳过启动")
		return
	}

	// 检查 lookbusy 是否可用
	if !g.isLookbusyAvailable() {
		log.Println("[Oracle守护] 警告：lookbusy 未安装，守护将仅监控不操作")
		log.Println("[Oracle守护] 安装命令：sudo apt-get install lookbusy -y")
	}

	g.mu.Lock()
	if g.running {
		g.mu.Unlock()
		return
	}
	g.running = true
	g.mu.Unlock()

	g.wg.Add(1)
	go g.run(ctx)

	log.Printf("[Oracle守护] 已启动（检查间隔：%v，CPU阈值：%.1f%%/%.1f%%）",
		g.config.CheckInterval, g.config.IdleCPUThreshold, g.config.BusyCPUThreshold)
}

// Stop 停止守护
func (g *OracleGuard) Stop() {
	g.mu.Lock()
	if !g.running {
		g.mu.Unlock()
		return
	}
	g.running = false
	close(g.stopChan)
	g.mu.Unlock()

	// 确保关闭 lookbusy
	g.stopLookbusy()

	g.wg.Wait()
	log.Println("[Oracle守护] 已停止")
}

// run 主循环
func (g *OracleGuard) run(ctx context.Context) {
	defer g.wg.Done()

	ticker := time.NewTicker(g.config.CheckInterval)
	defer ticker.Stop()

	// 首次检查延迟 30 秒
	time.Sleep(30 * time.Second)

	for {
		select {
		case <-ctx.Done():
			g.stopLookbusy()
			return
		case <-g.stopChan:
			g.stopLookbusy()
			return
		case <-ticker.C:
			g.checkAndAdjust()
		}
	}
}

// checkAndAdjust 检查系统状态并调整
func (g *OracleGuard) checkAndAdjust() {
	// 获取 CPU 使用率（最近 3 秒平均）
	cpuPercents, err := cpu.Percent(3*time.Second, false)
	if err != nil || len(cpuPercents) == 0 {
		log.Printf("[Oracle守护] 获取 CPU 失败: %v", err)
		return
	}
	currentCPU := cpuPercents[0]

	// 获取内存使用率
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		log.Printf("[Oracle守护] 获取内存失败: %v", err)
		return
	}
	currentMemUsed := memInfo.UsedPercent

	// 更新统计
	g.mu.Lock()
	g.lastCPU = currentCPU
	g.lastMemUsed = currentMemUsed
	g.mu.Unlock()

	// 决策逻辑
	shouldRun := false
	reason := ""

	// 1. 内存紧张时强制关闭
	if currentMemUsed > g.config.MaxMemUsage {
		shouldRun = false
		reason = fmt.Sprintf("内存紧张 (%.1f%% > %.1f%%)", currentMemUsed, g.config.MaxMemUsage)
	} else if currentCPU > g.config.BusyCPUThreshold {
		// 2. CPU 繁忙时关闭，释放资源给业务
		shouldRun = false
		reason = fmt.Sprintf("业务繁忙 (CPU %.1f%% > %.1f%%)", currentCPU, g.config.BusyCPUThreshold)
	} else if currentCPU < g.config.IdleCPUThreshold && currentMemUsed < g.config.IdleMemThreshold {
		// 3. CPU 和内存都空闲时启动
		shouldRun = true
		reason = fmt.Sprintf("系统空闲 (CPU %.1f%% < %.1f%%, 内存 %.1f%% < %.1f%%)",
			currentCPU, g.config.IdleCPUThreshold, currentMemUsed, g.config.IdleMemThreshold)
	}

	// 执行调整
	isRunning := g.isLookbusyRunning()
	if shouldRun && !isRunning {
		g.startLookbusy(reason)
	} else if !shouldRun && isRunning {
		g.stopLookbusyWithReason(reason)
	} else {
		// 状态不变，仅记录
		log.Printf("[Oracle守护] 状态保持: CPU=%.1f%%, 内存=%.1f%%, lookbusy=%v",
			currentCPU, currentMemUsed, isRunning)
	}
}

// isLookbusyAvailable 检查 lookbusy 是否可用
func (g *OracleGuard) isLookbusyAvailable() bool {
	_, err := exec.LookPath("lookbusy")
	return err == nil
}

// isLookbusyRunning 检查 lookbusy 是否在运行
func (g *OracleGuard) isLookbusyRunning() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.lookbusyCmd != nil && g.lookbusyCmd.Process != nil
}

// startLookbusy 启动 lookbusy
func (g *OracleGuard) startLookbusy(reason string) {
	if !g.isLookbusyAvailable() {
		log.Printf("[Oracle守护] 无法启动 lookbusy: 未安装")
		return
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	// 防止重复启动
	if g.lookbusyCmd != nil && g.lookbusyCmd.Process != nil {
		return
	}

	log.Printf("[Oracle守护] 🚀 启动 lookbusy: %s", reason)

	// 构建命令：nice -n 19 lookbusy -c 15 -m 6GB
	args := []string{
		"-n", fmt.Sprintf("%d", g.config.LookbusyNice),
		"lookbusy",
		"-c", g.config.LookbusyCPU,
		"-m", g.config.LookbusyMem,
	}

	g.lookbusyCmd = exec.Command("nice", args...)
	g.lookbusyCmd.Stdout = nil
	g.lookbusyCmd.Stderr = nil

	if err := g.lookbusyCmd.Start(); err != nil {
		log.Printf("[Oracle守护] ❌ 启动 lookbusy 失败: %v", err)
		g.lookbusyCmd = nil
		return
	}

	g.startCount++
	log.Printf("[Oracle守护] ✅ lookbusy 已启动 (PID: %d, CPU: %s, 内存: %s, Nice: %d)",
		g.lookbusyCmd.Process.Pid, g.config.LookbusyCPU, g.config.LookbusyMem, g.config.LookbusyNice)
}

// stopLookbusy 停止 lookbusy
func (g *OracleGuard) stopLookbusy() {
	g.stopLookbusyWithReason("服务停止")
}

// stopLookbusyWithReason 停止 lookbusy 并记录原因
func (g *OracleGuard) stopLookbusyWithReason(reason string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.lookbusyCmd == nil || g.lookbusyCmd.Process == nil {
		return
	}

	log.Printf("[Oracle守护] 🛑 关闭 lookbusy: %s", reason)

	// 发送 SIGTERM
	if err := g.lookbusyCmd.Process.Signal(nil); err == nil {
		_ = g.lookbusyCmd.Process.Kill()
		_ = g.lookbusyCmd.Wait()
	}

	g.lookbusyCmd = nil
	g.stopCount++
	log.Println("[Oracle守护] ✅ lookbusy 已关闭，资源已释放")
}

// GetStats 获取统计信息
func (g *OracleGuard) GetStats() map[string]interface{} {
	g.mu.Lock()
	defer g.mu.Unlock()

	return map[string]interface{}{
		"enabled":       g.config.Enabled,
		"running":       g.running,
		"lookbusy_running": g.lookbusyCmd != nil && g.lookbusyCmd.Process != nil,
		"last_cpu":      g.lastCPU,
		"last_mem_used": g.lastMemUsed,
		"start_count":   g.startCount,
		"stop_count":    g.stopCount,
		"config":        g.config,
		"goroutines":    runtime.NumGoroutine(),
	}
}
