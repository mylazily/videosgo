package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"videosgo/internal/service"
	"videosgo/pkg/response"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源（生产环境应限制）
	},
}

// WSHandler WebSocket 处理器
type WSHandler struct {
	svc *service.WSService
}

// NewWSHandler 创建 WebSocket 处理器
func NewWSHandler(svc *service.WSService) *WSHandler {
	return &WSHandler{svc: svc}
}

// HandleDanmaku WebSocket 弹幕连接
// GET /api/v1/ws/danmaku/:videoId
func (h *WSHandler) HandleDanmaku(c *gin.Context) {
	videoID := c.Param("videoId")

	// 验证视频 ID 格式
	if !service.ValidateVideoID(videoID) {
		response.BadRequest(c, "无效的视频 ID")
		return
	}

	// 升级为 WebSocket 连接
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[WS] WebSocket 升级失败: %v", err)
		return
	}

	// 获取用户标识
	user := c.Query("user")
	if user == "" {
		user = "anonymous"
	}

	// 处理连接
	h.svc.HandleConnection(conn, videoID, user)
}

// GetOnlineCount 获取在线人数
// GET /api/v1/ws/online/:videoId
func (h *WSHandler) GetOnlineCount(c *gin.Context) {
	videoID := c.Param("videoId")
	count := h.svc.GetOnlineCount(videoID)
	response.Success(c, gin.H{
		"video_id": videoID,
		"online":   count,
	})
}
