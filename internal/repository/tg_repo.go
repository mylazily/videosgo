package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/mylazily/videosgo/internal/model"
	"gorm.io/gorm"
)

// TGRepo TG Bot 数据仓库
type TGRepo struct {
	db *gorm.DB
}

// NewTGRepo 创建 TG Bot 仓库
func NewTGRepo(db *gorm.DB) *TGRepo {
	return &TGRepo{db: db}
}

// ========== TGBotConfig ==========

// GetBotConfig 获取 Bot 配置
func (r *TGRepo) GetBotConfig() (*model.TGBotConfig, error) {
	var config model.TGBotConfig
	err := r.db.Where("is_active = ?", true).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// CreateBotConfig 创建 Bot 配置
func (r *TGRepo) CreateBotConfig(config *model.TGBotConfig) error {
	return r.db.Create(config).Error
}

// UpdateBotConfig 更新 Bot 配置
func (r *TGRepo) UpdateBotConfig(config *model.TGBotConfig) error {
	return r.db.Save(config).Error
}

// ========== TGChannel ==========

// CreateChannel 创建频道
func (r *TGRepo) CreateChannel(channel *model.TGChannel) error {
	return r.db.Create(channel).Error
}

// GetChannelByID 根据 UUID 获取频道
func (r *TGRepo) GetChannelByID(id uuid.UUID) (*model.TGChannel, error) {
	var channel model.TGChannel
	err := r.db.First(&channel, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &channel, nil
}

// GetChannelByChannelID 根据 TG ChannelID 获取频道
func (r *TGRepo) GetChannelByChannelID(channelID int64) (*model.TGChannel, error) {
	var channel model.TGChannel
	err := r.db.First(&channel, "channel_id = ?", channelID).Error
	if err != nil {
		return nil, err
	}
	return &channel, nil
}

// ListActiveChannels 获取所有活跃频道
func (r *TGRepo) ListActiveChannels() ([]model.TGChannel, error) {
	var channels []model.TGChannel
	err := r.db.Where("is_active = ?", true).Find(&channels).Error
	if err != nil {
		return nil, err
	}
	return channels, nil
}

// ListChannels 获取所有频道
func (r *TGRepo) ListChannels() ([]model.TGChannel, error) {
	var channels []model.TGChannel
	err := r.db.Find(&channels).Error
	if err != nil {
		return nil, err
	}
	return channels, nil
}

// UpdateChannel 更新频道
func (r *TGRepo) UpdateChannel(channel *model.TGChannel) error {
	return r.db.Save(channel).Error
}

// ========== TGBroadcastLog ==========

// CreateBroadcast 创建广播日志
func (r *TGRepo) CreateBroadcast(log *model.TGBroadcastLog) error {
	return r.db.Create(log).Error
}

// ListBroadcasts 获取广播日志列表
func (r *TGRepo) ListBroadcasts(limit int) ([]model.TGBroadcastLog, error) {
	var logs []model.TGBroadcastLog
	err := r.db.Order("created_at DESC").Limit(limit).Find(&logs).Error
	if err != nil {
		return nil, err
	}
	return logs, nil
}

// UpdateBroadcastStatus 更新广播状态
func (r *TGRepo) UpdateBroadcastStatus(id uuid.UUID, status, errMsg string, messageID int64) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if errMsg != "" {
		updates["error_message"] = errMsg
	}
	if messageID > 0 {
		updates["message_id"] = messageID
	}
	if status == "sent" {
		now := time.Now()
		updates["sent_at"] = &now
	}
	return r.db.Model(&model.TGBroadcastLog{}).Where("id = ?", id).Updates(updates).Error
}

// ========== TGMiniAppSession ==========

// UpsertMiniAppSession 创建或更新 Mini App 会话
func (r *TGRepo) UpsertMiniAppSession(session *model.TGMiniAppSession) error {
	var existing model.TGMiniAppSession
	err := r.db.Where("tg_user_id = ?", session.TGUserID).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return r.db.Create(session).Error
	}
	if err != nil {
		return err
	}
	// 更新现有记录
	return r.db.Model(&existing).Updates(map[string]interface{}{
		"tg_username":     session.TGUsername,
		"tg_language":     session.TGLanguage,
		"fingerprint_id":  session.FingerprintID,
		"session_data":    session.SessionData,
		"last_open_at":    time.Now(),
		"total_opens":     gorm.Expr("total_opens + 1"),
	}).Error
}

// IncrementWatchTime 增加观看时长
func (r *TGRepo) IncrementWatchTime(tgUserID int64, seconds int64) error {
	return r.db.Model(&model.TGMiniAppSession{}).
		Where("tg_user_id = ?", tgUserID).
		Update("total_watch_time", gorm.Expr("total_watch_time + ?", seconds)).Error
}

// GetMiniAppStats 获取 Mini App 统计
func (r *TGRepo) GetMiniAppStats() (totalSessions int64, totalWatchTime int64, err error) {
	err = r.db.Model(&model.TGMiniAppSession{}).Count(&totalSessions).Error
	if err != nil {
		return
	}
	err = r.db.Model(&model.TGMiniAppSession{}).
		Select("COALESCE(SUM(total_watch_time), 0)").
		Scan(&totalWatchTime).Error
	return
}

// GetMiniAppSession 获取 Mini App 会话
func (r *TGRepo) GetMiniAppSession(tgUserID int64) (*model.TGMiniAppSession, error) {
	var session model.TGMiniAppSession
	err := r.db.Where("tg_user_id = ?", tgUserID).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}
