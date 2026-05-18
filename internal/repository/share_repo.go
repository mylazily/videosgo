package repository

import (
	"github.com/google/uuid"
	"videosgo/internal/model"
	"gorm.io/gorm"
)

// ShareRepo 分享数据仓库
type ShareRepo struct {
	db *gorm.DB
}

// NewShareRepo 创建分享仓库
func NewShareRepo(db *gorm.DB) *ShareRepo {
	return &ShareRepo{db: db}
}

// CreateShareLink 创建分享链接
func (r *ShareRepo) CreateShareLink(link *model.ShareLink) error {
	return r.db.Create(link).Error
}

// GetByCode 通过分享码查询
func (r *ShareRepo) GetByCode(code string) (*model.ShareLink, error) {
	var link model.ShareLink
	err := r.db.Preload("Video").Where("share_code = ?", code).First(&link).Error
	if err != nil {
		return nil, err
	}
	return &link, nil
}

// GetByID 根据 ID 获取分享链接
func (r *ShareRepo) GetByID(id uuid.UUID) (*model.ShareLink, error) {
	var link model.ShareLink
	err := r.db.Preload("Video").First(&link, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &link, nil
}

// RecordClick 记录分享点击
func (r *ShareRepo) RecordClick(click *model.ShareClick) error {
	return r.db.Create(click).Error
}

// IncrementUnlock 增加解锁次数
func (r *ShareRepo) IncrementUnlock(shareLinkID uuid.UUID) error {
	return r.db.Model(&model.ShareLink{}).Where("id = ?", shareLinkID).
		UpdateColumn("unlock_count", gorm.Expr("unlock_count + 1")).Error
}

// IncrementClick 增加点击次数
func (r *ShareRepo) IncrementClick(shareLinkID uuid.UUID) error {
	return r.db.Model(&model.ShareLink{}).Where("id = ?", shareLinkID).
		UpdateColumn("click_count", gorm.Expr("click_count + 1")).Error
}

// GetByCreator 获取创建者的所有分享链接
func (r *ShareRepo) GetByCreator(fingerprintID uuid.UUID, page, pageSize int) ([]model.ShareLink, int64, error) {
	var links []model.ShareLink
	var total int64

	db := r.db.Model(&model.ShareLink{}).Where("creator_fingerprint_id = ?", fingerprintID)
	db.Count(&total)

	err := db.Offset((page - 1) * pageSize).Limit(pageSize).
		Order("created_at DESC").
		Find(&links).Error
	return links, total, err
}

// HasClicked 检查设备是否已点击过该分享链接
func (r *ShareRepo) HasClicked(shareLinkID, fingerprintID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&model.ShareClick{}).
		Where("share_link_id = ? AND fingerprint_id = ?", shareLinkID, fingerprintID).
		Count(&count).Error
	return count > 0, err
}
