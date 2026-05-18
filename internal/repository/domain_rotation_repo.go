package repository

import (
	"github.com/google/uuid"
	"videosgo/internal/model"
	"gorm.io/gorm"
)

// DomainRotationRepo 域名轮询数据仓库
type DomainRotationRepo struct {
	db *gorm.DB
}

// NewDomainRotationRepo 创建域名轮询仓库
func NewDomainRotationRepo(db *gorm.DB) *DomainRotationRepo {
	return &DomainRotationRepo{db: db}
}

// UpdateAvailability 更新域名可用性
func (r *DomainRotationRepo) UpdateAvailability(domainID uuid.UUID, region string, isAccessible bool, responseTimeMs int, errorType string) error {
	record := &model.DomainAvailability{
		DomainID:       domainID,
		Region:         region,
		IsAccessible:   isAccessible,
		ResponseTimeMs: responseTimeMs,
		ErrorType:      errorType,
	}
	return r.db.Create(record).Error
}

// GetActiveDomain 获取当前活跃域名
func (r *DomainRotationRepo) GetActiveDomain(region string) (*model.ActiveDomain, error) {
	var domain model.ActiveDomain
	query := r.db.Where("1 = 1")
	if region != "" {
		query = query.Where("region = ? OR region = ''", region)
	}
	err := query.Order("activated_at DESC").First(&domain).Error
	if err != nil {
		return nil, err
	}
	return &domain, nil
}

// SetActiveDomain 设置活跃域名（使用事务确保原子性）
func (r *DomainRotationRepo) SetActiveDomain(domain, region, activatedBy string) error {
	// 使用事务确保删除和创建的原子性
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 先停用所有活跃域名
		if err := tx.Where("1 = 1").Delete(&model.ActiveDomain{}).Error; err != nil {
			return err
		}
		// 创建新的活跃域名
		active := &model.ActiveDomain{
			Domain:      domain,
			Region:      region,
			ActivatedBy: activatedBy,
		}
		return tx.Create(active).Error
	})
}

// LogSwitch 记录域名切换事件
func (r *DomainRotationRepo) LogSwitch(fromDomain, toDomain, reason string) error {
	event := &model.DomainSwitchEvent{
		FromDomain:   fromDomain,
		ToDomain:     toDomain,
		SwitchReason: reason,
	}
	return r.db.Create(event).Error
}

// GetSwitchHistory 获取域名切换历史
func (r *DomainRotationRepo) GetSwitchHistory(limit int) ([]model.DomainSwitchEvent, error) {
	var events []model.DomainSwitchEvent
	err := r.db.Order("switched_at DESC").Limit(limit).Find(&events).Error
	if err != nil {
		return nil, err
	}
	return events, nil
}

// GetLatestAvailability 获取域名最新可用性
func (r *DomainRotationRepo) GetLatestAvailability(domainID uuid.UUID, region string) (*model.DomainAvailability, error) {
	var avail model.DomainAvailability
	query := r.db.Where("domain_id = ?", domainID)
	if region != "" {
		query = query.Where("region = ?", region)
	}
	err := query.Order("checked_at DESC").First(&avail).Error
	if err != nil {
		return nil, err
	}
	return &avail, nil
}

// GetDomainList 获取所有域名（从 active_domains 和 domain_availabilities 聚合）
func (r *DomainRotationRepo) GetDomainList() ([]model.ActiveDomain, []model.DomainAvailability, error) {
	var actives []model.ActiveDomain
	err := r.db.Find(&actives).Error
	if err != nil {
		return nil, nil, err
	}
	var availabilities []model.DomainAvailability
	err = r.db.Order("checked_at DESC").Limit(100).Find(&availabilities).Error
	if err != nil {
		return nil, nil, err
	}
	return actives, availabilities, nil
}
