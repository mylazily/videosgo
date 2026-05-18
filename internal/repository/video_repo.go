package repository

import (
	"strings"
	"videosgo/internal/model"

	"gorm.io/gorm"
)

// VideoRepository 视频仓库
type VideoRepository struct {
	db *gorm.DB
}

// NewVideoRepository 创建视频仓库
func NewVideoRepository(db *gorm.DB) *VideoRepository {
	return &VideoRepository{db: db}
}

// List 获取视频列表
func (r *VideoRepository) List(offset, limit int) ([]model.Video, error) {
	var videos []model.Video
	err := r.db.Where("status = ?", "active").
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&videos).Error
	return videos, err
}

// GetByID 根据 ID 获取视频
func (r *VideoRepository) GetByID(id string) (*model.Video, error) {
	var video model.Video
	err := r.db.Where("id = ?", id).First(&video).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	return &video, nil
}

// GetRandom 获取随机视频列表
func (r *VideoRepository) GetRandom(limit int) ([]model.Video, error) {
	var videos []model.Video
	err := r.db.Where("status = ?", "active").
		Order("RANDOM()").
		Limit(limit).
		Find(&videos).Error
	return videos, err
}

// GetHot 获取热门视频列表
func (r *VideoRepository) GetHot(limit int) ([]model.Video, error) {
	var videos []model.Video
	err := r.db.Where("status = ?", "active").
		Order("created_at DESC").
		Limit(limit).
		Find(&videos).Error
	return videos, err
}

// Create 创建视频
func (r *VideoRepository) Create(video *model.Video) error {
	return r.db.Create(video).Error
}

// Update 更新视频
func (r *VideoRepository) Update(video *model.Video) error {
	return r.db.Save(video).Error
}

// Delete 删除视频
func (r *VideoRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&model.Video{}).Error
}

// Count 统计视频数量
func (r *VideoRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&model.Video{}).Where("status = ?", "active").Count(&count).Error
	return count, err
}

// Search 搜索视频
func (r *VideoRepository) Search(keyword string, offset, limit int) ([]model.Video, error) {
	var videos []model.Video
	// 转义 SQL LIKE 特殊字符，防止注入
	escapedKeyword := strings.ReplaceAll(keyword, "%", "\\%")
	escapedKeyword = strings.ReplaceAll(escapedKeyword, "_", "\\_")
	err := r.db.Where("status = ? AND (title ILIKE ? ESCAPE '\\\\' OR clean_title ILIKE ? ESCAPE '\\\\')",
		"active", "%"+escapedKeyword+"%", "%"+escapedKeyword+"%").
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&videos).Error
	return videos, err
}
