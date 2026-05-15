package repository

import (
	"github.com/google/uuid"
	"github.com/mylazily/videosgo/internal/model"
	"gorm.io/gorm"
)

// DanmakuRepo 弹幕数据仓库
type DanmakuRepo struct {
	db *gorm.DB
}

// NewDanmakuRepo 创建弹幕仓库
func NewDanmakuRepo(db *gorm.DB) *DanmakuRepo {
	return &DanmakuRepo{db: db}
}

// Create 创建弹幕
func (r *DanmakuRepo) Create(danmaku *model.Danmaku) error {
	return r.db.Create(danmaku).Error
}

// BatchCreate 批量创建弹幕
func (r *DanmakuRepo) BatchCreate(danmakus []model.Danmaku) error {
	return r.db.CreateInBatches(&danmakus, 100).Error
}

// ListByEpisodeID 获取剧集的弹幕列表
func (r *DanmakuRepo) ListByEpisodeID(episodeID uuid.UUID) ([]model.Danmaku, error) {
	var danmakus []model.Danmaku
	err := r.db.Where("episode_id = ?", episodeID).
		Order("time ASC").
		Find(&danmakus).Error
	return danmakus, err
}

// ListByVideoID 获取视频的所有弹幕
func (r *DanmakuRepo) ListByVideoID(videoID uuid.UUID) ([]model.Danmaku, error) {
	var danmakus []model.Danmaku
	err := r.db.Where("video_id = ?", videoID).
		Order("episode_id ASC, time ASC").
		Find(&danmakus).Error
	return danmakus, err
}

// DeleteByEpisodeID 删除剧集的弹幕
func (r *DanmakuRepo) DeleteByEpisodeID(episodeID uuid.UUID) error {
	return r.db.Where("episode_id = ?", episodeID).Delete(&model.Danmaku{}).Error
}

// GetCountByVideoID 获取视频弹幕数
func (r *DanmakuRepo) GetCountByVideoID(videoID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&model.Danmaku{}).
		Where("video_id = ?", videoID).
		Count(&count).Error
	return count, err
}
