package handler

import (
	"github.com/google/uuid"
	"github.com/gin-gonic/gin"
	"videosgo/internal/model"
	"videosgo/internal/service"
	"videosgo/pkg/response"
)

// P2PHandler P2P 信令处理器
type P2PHandler struct {
	svc *service.P2PService
}

// NewP2PHandler 创建 P2P 信令处理器
func NewP2PHandler(svc *service.P2PService) *P2PHandler {
	return &P2PHandler{svc: svc}
}

// RegisterPeer 注册节点
// POST /api/v1/p2p/register
func (h *P2PHandler) RegisterPeer(c *gin.Context) {
	var req struct {
		PeerID        string `json:"peer_id" binding:"required"`
		FingerprintID string `json:"fingerprint_id"`
		IPAddress     string `json:"ip_address"`
		Region        string `json:"region"`
		VideoID       string `json:"video_id"`
		BandwidthScore int   `json:"bandwidth_score"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	peer := &model.PeerRegistry{
		PeerID:         req.PeerID,
		FingerprintID:  parseUUID(req.FingerprintID),
		IPAddress:      req.IPAddress,
		Region:         req.Region,
		IsActive:       true,
		CurrentVideoID: parseUUID(req.VideoID),
		BandwidthScore: req.BandwidthScore,
	}

	if err := h.svc.RegisterPeer(peer); err != nil {
		response.InternalError(c, "注册节点失败")
		return
	}

	response.Success(c, peer)
}

// Heartbeat 心跳
// POST /api/v1/p2p/heartbeat
func (h *P2PHandler) Heartbeat(c *gin.Context) {
	var req struct {
		PeerID  string `json:"peer_id" binding:"required"`
		VideoID string `json:"video_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	if err := h.svc.Heartbeat(req.PeerID, req.VideoID); err != nil {
		response.InternalError(c, "心跳更新失败")
		return
	}

	response.SuccessWithMessage(c, "心跳成功", nil)
}

// UnregisterPeer 注销节点
// DELETE /api/v1/p2p/unregister
func (h *P2PHandler) UnregisterPeer(c *gin.Context) {
	var req struct {
		PeerID string `json:"peer_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	if err := h.svc.UnregisterPeer(req.PeerID); err != nil {
		response.InternalError(c, "注销节点失败")
		return
	}

	response.SuccessWithMessage(c, "注销成功", nil)
}

// OfferSignal 发送 Offer 信令
// POST /api/v1/p2p/signal/offer
func (h *P2PHandler) OfferSignal(c *gin.Context) {
	var req struct {
		RoomID        string `json:"room_id" binding:"required"`
		PeerID        string `json:"peer_id" binding:"required"`
		FingerprintID string `json:"fingerprint_id"`
		TargetPeerID  string `json:"target_peer_id" binding:"required"`
		SDPData       string `json:"sdp_data" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	if err := h.svc.OfferSignal(req.RoomID, req.PeerID, req.FingerprintID, req.TargetPeerID, req.SDPData); err != nil {
		response.InternalError(c, "发送 Offer 失败")
		return
	}

	response.SuccessWithMessage(c, "Offer 已发送", nil)
}

// AnswerSignal 发送 Answer 信令
// POST /api/v1/p2p/signal/answer
func (h *P2PHandler) AnswerSignal(c *gin.Context) {
	var req struct {
		RoomID        string `json:"room_id" binding:"required"`
		PeerID        string `json:"peer_id" binding:"required"`
		FingerprintID string `json:"fingerprint_id"`
		TargetPeerID  string `json:"target_peer_id" binding:"required"`
		SDPData       string `json:"sdp_data" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	if err := h.svc.AnswerSignal(req.RoomID, req.PeerID, req.FingerprintID, req.TargetPeerID, req.SDPData); err != nil {
		response.InternalError(c, "发送 Answer 失败")
		return
	}

	response.SuccessWithMessage(c, "Answer 已发送", nil)
}

// ExchangeICE 交换 ICE 候选
// POST /api/v1/p2p/signal/ice
func (h *P2PHandler) ExchangeICE(c *gin.Context) {
	var req struct {
		RoomID        string `json:"room_id" binding:"required"`
		PeerID        string `json:"peer_id" binding:"required"`
		FingerprintID string `json:"fingerprint_id"`
		TargetPeerID  string `json:"target_peer_id" binding:"required"`
		ICECandidate  string `json:"ice_candidate" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	if err := h.svc.ExchangeICE(req.RoomID, req.PeerID, req.FingerprintID, req.TargetPeerID, req.ICECandidate); err != nil {
		response.InternalError(c, "ICE 交换失败")
		return
	}

	response.SuccessWithMessage(c, "ICE 候选已交换", nil)
}

// GetVideoPeers 获取视频的在线节点
// GET /api/v1/p2p/peers/:video_id
func (h *P2PHandler) GetVideoPeers(c *gin.Context) {
	videoID := c.Param("video_id")
	if videoID == "" {
		response.BadRequest(c, "缺少视频 ID")
		return
	}

	peers, err := h.svc.GetPeersForVideo(videoID)
	if err != nil {
		response.InternalError(c, "获取节点列表失败")
		return
	}

	response.Success(c, peers)
}

// parseUUID 安全解析 UUID，空字符串时返回 uuid.Nil
func parseUUID(s string) uuid.UUID {
	if s == "" {
		return uuid.Nil
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil
	}
	return id
}
