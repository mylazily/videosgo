package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/mylazily/videosgo/internal/model"
	"gorm.io/gorm"
)

// P2PRepo P2P 信令数据仓库
type P2PRepo struct {
	db *gorm.DB
}

// NewP2PRepo 创建 P2P 信令仓库
func NewP2PRepo(db *gorm.DB) *P2PRepo {
	return &P2PRepo{db: db}
}

// CreateSignal 创建信令
func (r *P2PRepo) CreateSignal(signal *model.SignalChannel) error {
	return r.db.Create(signal).Error
}

// GetSignalsByRoom 获取房间的信令列表
func (r *P2PRepo) GetSignalsByRoom(roomID string, limit int) ([]model.SignalChannel, error) {
	var signals []model.SignalChannel
	err := r.db.Where("room_id = ?", roomID).
		Order("created_at DESC").
		Limit(limit).
		Find(&signals).Error
	return signals, err
}

// GetSignalsByRoomAndTarget 获取房间内目标节点的信令
func (r *P2PRepo) GetSignalsByRoomAndTarget(roomID, targetPeerID string, limit int) ([]model.SignalChannel, error) {
	var signals []model.SignalChannel
	err := r.db.Where("room_id = ? AND target_peer_id = ?", roomID, targetPeerID).
		Order("created_at DESC").
		Limit(limit).
		Find(&signals).Error
	return signals, err
}

// CleanupExpiredSignals 清理过期信令
func (r *P2PRepo) CleanupExpiredSignals(expireSeconds int) (int64, error) {
	cutoff := time.Now().Add(-time.Duration(expireSeconds) * time.Second)
	result := r.db.Where("created_at < ?", cutoff).Delete(&model.SignalChannel{})
	return result.RowsAffected, result.Error
}

// RegisterPeer 注册节点
func (r *P2PRepo) RegisterPeer(peer *model.PeerRegistry) error {
	return r.db.Create(peer).Error
}

// UnregisterPeer 注销节点
func (r *P2PRepo) UnregisterPeer(peerID string) error {
	now := time.Now()
	return r.db.Model(&model.PeerRegistry{}).Where("peer_id = ?", peerID).Updates(map[string]interface{}{
		"is_active":       false,
		"disconnected_at": &now,
	}).Error
}

// Heartbeat 更新节点心跳
func (r *P2PRepo) Heartbeat(peerID string, videoID string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"last_heartbeat": &now,
	}
	if videoID != "" {
		updates["current_video_id"] = videoID
	}
	return r.db.Model(&model.PeerRegistry{}).Where("peer_id = ? AND is_active = ?", peerID, true).
		Updates(updates).Error
}

// GetPeerByPeerID 根据节点 ID 获取节点
func (r *P2PRepo) GetPeerByPeerID(peerID string) (*model.PeerRegistry, error) {
	var peer model.PeerRegistry
	err := r.db.Where("peer_id = ?", peerID).First(&peer).Error
	if err != nil {
		return nil, err
	}
	return &peer, nil
}

// GetActivePeersByVideo 获取正在看同一视频的在线节点
func (r *P2PRepo) GetActivePeersByVideo(videoID string) ([]model.PeerRegistry, error) {
	var peers []model.PeerRegistry
	err := r.db.Where("current_video_id = ? AND is_active = ?", videoID, true).
		Find(&peers).Error
	return peers, err
}

// GetAllActivePeers 获取所有在线节点
func (r *P2PRepo) GetAllActivePeers() ([]model.PeerRegistry, error) {
	var peers []model.PeerRegistry
	err := r.db.Where("is_active = ?", true).Find(&peers).Error
	return peers, err
}

// CleanupInactivePeers 清理失活节点（心跳超时）
func (r *P2PRepo) CleanupInactivePeers(timeoutSeconds int) (int64, error) {
	cutoff := time.Now().Add(-time.Duration(timeoutSeconds) * time.Second)
	result := r.db.Model(&model.PeerRegistry{}).
		Where("is_active = ? AND (last_heartbeat IS NULL OR last_heartbeat < ?)", true, cutoff).
		Updates(map[string]interface{}{
			"is_active":       false,
			"disconnected_at": time.Now(),
		})
	return result.RowsAffected, result.Error
}

// LogTransfer 记录传输日志
func (r *P2PRepo) LogTransfer(log *model.TransferLog) error {
	return r.db.Create(log).Error
}
