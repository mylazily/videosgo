package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"videosgo/internal/service"

	"github.com/gin-gonic/gin"
)

// TGHandler Telegram 处理器
type TGHandler struct {
	tgService *service.TGService
}

// NewTGHandler 创建 TG 处理器
func NewTGHandler(tgService *service.TGService) *TGHandler {
	return &TGHandler{tgService: tgService}
}

// Webhook 处理 Webhook
func (h *TGHandler) Webhook(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "read body failed"})
		return
	}

	var update service.TGUpdate
	if err := json.Unmarshal(body, &update); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}

	if err := h.tgService.HandleUpdate(&update); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}
