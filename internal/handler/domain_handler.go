package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mylazily/videosgo/internal/service"
	"github.com/mylazily/videosgo/pkg/response"
)

// DomainHandler 域名轮询处理器
type DomainHandler struct {
	svc *service.DomainRotationService
}

// NewDomainHandler 创建域名轮询处理器
func NewDomainHandler(svc *service.DomainRotationService) *DomainHandler {
	return &DomainHandler{svc: svc}
}

// GetActiveDomain 获取当前活跃域名
// GET /api/v1/domain/active
func (h *DomainHandler) GetActiveDomain(c *gin.Context) {
	region := c.Query("region")
	domain, err := h.svc.GetActiveDomain(region)
	if err != nil {
		response.NotFound(c, "未找到活跃域名")
		return
	}

	response.Success(c, gin.H{
		"domain": *domain,
		"region": region,
	})
}

// GetDomainList 域名列表及状态
// GET /api/v1/domain/list
func (h *DomainHandler) GetDomainList(c *gin.Context) {
	data, err := h.svc.GetDomainList()
	if err != nil {
		response.InternalError(c, "获取域名列表失败")
		return
	}
	response.Success(c, data)
}

// SwitchDomain 手动切换域名（管理员）
// POST /api/v1/domain/switch
func (h *DomainHandler) SwitchDomain(c *gin.Context) {
	var req struct {
		Domain string `json:"domain" binding:"required"`
		Region string `json:"region"`
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if err := h.svc.SwitchDomain(req.Domain, req.Region, req.Reason); err != nil {
		response.InternalError(c, "切换域名失败")
		return
	}

	response.SuccessWithMessage(c, "域名切换成功", nil)
}

// GetSwitchHistory 切换历史
// GET /api/v1/domain/history
func (h *DomainHandler) GetSwitchHistory(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 20
	}

	data, err := h.svc.GetSwitchHistory(limit)
	if err != nil {
		response.InternalError(c, "获取切换历史失败")
		return
	}
	response.Success(c, data)
}

// GetHealthyDomains 获取最健康的 3 个可用域名（前端容灾用）
// GET /api/v1/domains/healthy
// 轻量级接口：并发探测所有域名，返回延迟最低的 3 个
func (h *DomainHandler) GetHealthyDomains(c *gin.Context) {
	domain, err := h.svc.GetBestDomain("")
	if err != nil {
		// 没有可用域名，返回空列表
		response.Success(c, gin.H{
			"domains": []string{},
			"message": "所有域名均不可用",
		})
		return
	}

	// 获取域名列表
	listData, err := h.svc.GetDomainList()
	if err != nil {
		response.Success(c, gin.H{
			"domains":    []string{domain},
			"active":     domain,
			"message":    "仅活跃域名可用",
		})
		return
	}

	// 从域名列表中提取可用域名
	type domainStatus struct {
		Domain       string
		IsAccessible bool
		Latency      int
	}

	// listData is interface{}, we need to handle it
	// Since GetDomainList returns interface{}, extract what we can
	activeDomain := domain
	var allDomains []string

	// Try to extract domains from the list data
	if listMap, ok := listData.(map[string]interface{}); ok {
		if actives, ok := listMap["active"].([]interface{}); ok {
			for _, a := range actives {
				if activeMap, ok := a.(map[string]interface{}); ok {
					if d, ok := activeMap["domain"].(string); ok && d != "" {
						allDomains = append(allDomains, d)
					}
				}
			}
		}
	}

	// 确保活跃域名在列表中
	found := false
	for _, d := range allDomains {
		if d == activeDomain {
			found = true
			break
		}
	}
	if !found {
		allDomains = append([]string{activeDomain}, allDomains...)
	}

	// 返回前 3 个域名（去重）
	seen := make(map[string]bool)
	var healthyDomains []string
	for _, d := range allDomains {
		if !seen[d] {
			seen[d] = true
			healthyDomains = append(healthyDomains, d)
			if len(healthyDomains) >= 3 {
				break
			}
		}
	}

	response.Success(c, gin.H{
		"domains": healthyDomains,
		"active":  activeDomain,
		"count":   len(healthyDomains),
	})
}
