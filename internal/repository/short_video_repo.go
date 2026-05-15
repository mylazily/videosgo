package repository

import (
	"github.com/google/uuid"
	"github.com/mylazily/videosgo/internal/model"
	"gorm.io/gorm"
)

// ShortVideoRepo 短视频数据仓库
type ShortVideoRepo struct {
	db *gorm.DB
}

// NewShortVideoRepo 创建短视频仓库
func NewShortVideoRepo(db *gorm.DB) *ShortVideoRepo {
	return &ShortVideoRepo{db: db}
}

// Create 创建短视频
func (r *ShortVideoRepo) Create(sv *model.ShortVideo) error {
	return r.db.Create(sv).Error
}

// Update 更新短视频
func (r *ShortVideoRepo) Update(sv *model.ShortVideo) error {
	return r.db.Save(sv).Error
}

// Delete 删除短视频
func (r *ShortVideoRepo) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.ShortVideo{}, "id = ?", id).Error
}

// GetByID 根据 ID 获取短视频
func (r *ShortVideoRepo) GetByID(id uuid.UUID) (*model.ShortVideo, error) {
	var sv model.ShortVideo
	err := r.db.Where("id = ? AND status = ?", id, "active").First(&sv).Error
	if err != nil {
		return nil, err
	}
	return &sv, nil
}

// List 获取短视频列表（分页，支持排序）
func (r *ShortVideoRepo) List(page, pageSize int, sortBy string) ([]model.ShortVideo, int64, error) {
	var shorts []model.ShortVideo
	var total int64

	db := r.db.Model(&model.ShortVideo{}).Where("status = ?", "active")
	db.Count(&total)

	// 根据 sortBy 选择排序方式
	switch sortBy {
	case "popular":
		db = db.Order("view_count DESC")
	case "latest":
		db = db.Order("created_at DESC")
	default:
		db = db.Order("created_at DESC")
	}

	err := db.Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&shorts).Error
	return shorts, total, err
}

// GetRandom 随机获取短视频（用于"发现"页）
func (r *ShortVideoRepo) GetRandom(limit int) ([]model.ShortVideo, error) {
	var shorts []model.ShortVideo
	err := r.db.Where("status = ?", "active").
		Order("RANDOM()").
		Limit(limit).
		Find(&shorts).Error
	return shorts, err
}

// IncrementView 增加播放量
func (r *ShortVideoRepo) IncrementView(id uuid.UUID) error {
	return r.db.Model(&model.ShortVideo{}).Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

// IncrementLike 增加点赞数
func (r *ShortVideoRepo) IncrementLike(id uuid.UUID) error {
	return r.db.Model(&model.ShortVideo{}).Where("id = ?", id).
		UpdateColumn("like_count", gorm.Expr("like_count + 1")).Error
}

// IncrementShare 增加分享数
func (r *ShortVideoRepo) IncrementShare(id uuid.UUID) error {
	return r.db.Model(&model.ShortVideo{}).Where("id = ?", id).
		UpdateColumn("share_count", gorm.Expr("share_count + 1")).Error
}

// Search 搜索短视频
func (r *ShortVideoRepo) Search(keyword string, page, pageSize int) ([]model.ShortVideo, int64, error) {
	var shorts []model.ShortVideo
	var total int64

	db := r.db.Model(&model.ShortVideo{}).
		Where("status = ? AND (title LIKE ? OR description LIKE ?)",
			"active", "%"+keyword+"%", "%"+keyword+"%")
	db.Count(&total)

	err := db.Offset((page - 1) * pageSize).Limit(pageSize).
		Order("created_at DESC").
		Find(&shorts).Error
	return shorts, total, err
}

// GetHot 获取热门短视频（按播放量排序）
func (r *ShortVideoRepo) GetHot(limit int) ([]model.ShortVideo, error) {
	var shorts []model.ShortVideo
	err := r.db.Where("status = ?", "active").
		Order("view_count DESC").
		Limit(limit).
		Find(&shorts).Error
	return shorts, err
}

// GetLatest 获取最新短视频
func (r *ShortVideoRepo) GetLatest(limit int) ([]model.ShortVideo, error) {
	var shorts []model.ShortVideo
	err := r.db.Where("status = ?", "active").
		Order("created_at DESC").
		Limit(limit).
		Find(&shorts).Error
	return shorts, err
}
