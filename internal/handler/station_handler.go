package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"videosgo/internal/service"
	"videosgo/pkg/response"
)

// StationHandler 资源站监控 Handler
type StationHandler struct {
	monitor *service.StationMonitor
}

// NewStationHandler 构造函数
func NewStationHandler(monitor *service.StationMonitor) *StationHandler {
	return &StationHandler{monitor: monitor}
}

// GetStatus GET /api/v1/stations/status — 获取所有资源站状态
// 返回: {stations: [{name, is_alive, latency_ms, speed_kbps, weight, region}]}
func (h *StationHandler) GetStatus(c *gin.Context) {
	stations := h.monitor.GetStatus()

	// 构建精简响应
	list := make([]gin.H, 0, len(stations))
	for _, s := range stations {
		list = append(list, gin.H{
			"name":       s.Name,
			"ping_url":   s.PingURL,
			"is_alive":   s.IsAlive,
			"latency_ms": s.Latency,
			"speed_kbps": s.Speed,
			"weight":     s.Weight,
			"fail_count": s.FailCount,
			"region":     s.Region,
			"last_check_at":   s.LastCheckAt,
			"last_success_at": s.LastSuccessAt,
		})
	}

	response.Success(c, gin.H{
		"total":    len(list),
		"stations": list,
	})
}

// GetBest GET /api/v1/stations/best — 获取最优资源站
// 返回: {station: {name, latency_ms, speed_kbps, weight}}
func (h *StationHandler) GetBest(c *gin.Context) {
	best := h.monitor.GetBestStation()
	if best == nil {
		response.Error(c, http.StatusServiceUnavailable, "当前没有可用的资源站")
		return
	}

	response.Success(c, gin.H{
		"station": gin.H{
			"name":       best.Name,
			"latency_ms": best.Latency,
			"speed_kbps": best.Speed,
			"weight":     best.Weight,
			"region":     best.Region,
		},
	})
}

// TriggerCheck POST /api/v1/stations/check — 手动触发检查（管理员）
// 返回: {message: "检查完成", results: [...]}
func (h *StationHandler) TriggerCheck(c *gin.Context) {
	results := h.monitor.CheckAllAndReturnResults(c.Request.Context())

	response.SuccessWithMessage(c, "检查完成", gin.H{
		"total":   len(results),
		"results": results,
	})
}

// GetAliveStations GET /api/v1/stations/alive — 获取存活的资源站（按权重排序）
func (h *StationHandler) GetAliveStations(c *gin.Context) {
	alive := h.monitor.GetAliveStations()

	list := make([]gin.H, 0, len(alive))
	for _, s := range alive {
		list = append(list, gin.H{
			"name":       s.Name,
			"latency_ms": s.Latency,
			"speed_kbps": s.Speed,
			"weight":     s.Weight,
			"region":     s.Region,
		})
	}

	response.Success(c, gin.H{
		"total":    len(list),
		"stations": list,
	})
}
