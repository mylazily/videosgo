// Package repository 数据访问层
package repository

import (
	"github.com/google/uuid"
	"videosgo/internal/model"
	"gorm.io/gorm"
)

// CollectRepo 采集源数据仓库
type CollectRepo struct {
	db *gorm.DB
}

// NewCollectRepo 创建采集源仓库
func NewCollectRepo(db *gorm.DB) *CollectRepo {
	return &CollectRepo{db: db}
}

// Create 创建采集源
func (r *CollectRepo) Create(source *model.CollectSource) error {
	return r.db.Create(source).Error
}

// Update 更新采集源
func (r *CollectRepo) Update(source *model.CollectSource) error {
	return r.db.Save(source).Error
}

// Delete 删除采集源
func (r *CollectRepo) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.CollectSource{}, "id = ?", id).Error
}

// GetByID 根据 ID 获取采集源
func (r *CollectRepo) GetByID(id uuid.UUID) (*model.CollectSource, error) {
	var source model.CollectSource
	err := r.db.First(&source, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &source, nil
}

// List 获取采集源列表
func (r *CollectRepo) List(page, pageSize int) ([]model.CollectSource, int64, error) {
	var sources []model.CollectSource
	var total int64

	db := r.db.Model(&model.CollectSource{})
	db.Count(&total)

	err := db.Offset((page - 1) * pageSize).Limit(pageSize).
		Order("created_at DESC").
		Find(&sources).Error
	return sources, total, err
}

// GetEnabled 获取所有启用的采集源
func (r *CollectRepo) GetEnabled() ([]model.CollectSource, error) {
	var sources []model.CollectSource
	err := r.db.Where("enabled = ?", true).Find(&sources).Error
	return sources, err
}

// UpdateLastCollect 更新上次采集时间
func (r *CollectRepo) UpdateLastCollect(id uuid.UUID) error {
	return r.db.Model(&model.CollectSource{}).Where("id = ?", id).
		Update("last_collect", gorm.Expr("NOW()")).Error
}

// CreateLog 创建采集日志
func (r *CollectRepo) CreateLog(log *model.CollectLog) error {
	return r.db.Create(log).Error
}

// UpdateLog 更新采集日志
func (r *CollectRepo) UpdateLog(log *model.CollectLog) error {
	return r.db.Save(log).Error
}

// ListLogs 获取采集日志列表
func (r *CollectRepo) ListLogs(page, pageSize int) ([]model.CollectLog, int64, error) {
	var logs []model.CollectLog
	var total int64

	db := r.db.Model(&model.CollectLog{})
	db.Count(&total)

	err := db.Offset((page - 1) * pageSize).Limit(pageSize).
		Order("created_at DESC").
		Find(&logs).Error
	return logs, total, err
}

// GetLogsBySourceID 根据采集源 ID 获取日志
func (r *CollectRepo) GetLogsBySourceID(sourceID uuid.UUID, page, pageSize int) ([]model.CollectLog, int64, error) {
	var logs []model.CollectLog
	var total int64

	db := r.db.Model(&model.CollectLog{}).Where("source_id = ?", sourceID)
	db.Count(&total)

	err := db.Offset((page - 1) * pageSize).Limit(pageSize).
		Order("created_at DESC").
		Find(&logs).Error
	return logs, total, err
}
