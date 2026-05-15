package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/mylazily/videosgo/internal/model"
	"gorm.io/gorm"
)

// SiteRepo 站群域名数据仓库
type SiteRepo struct {
	db *gorm.DB
}

// NewSiteRepo 创建站群域名仓库
func NewSiteRepo(db *gorm.DB) *SiteRepo {
	return &SiteRepo{db: db}
}

// Create 创建域名
func (r *SiteRepo) Create(site *model.SiteDomain) error {
	return r.db.Create(site).Error
}

// Update 更新域名
func (r *SiteRepo) Update(site *model.SiteDomain) error {
	return r.db.Save(site).Error
}

// Delete 删除域名
func (r *SiteRepo) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.SiteDomain{}, "id = ?", id).Error
}

// GetByID 根据 ID 获取域名
func (r *SiteRepo) GetByID(id uuid.UUID) (*model.SiteDomain, error) {
	var site model.SiteDomain
	err := r.db.First(&site, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &site, nil
}

// GetByDomain 根据域名获取
func (r *SiteRepo) GetByDomain(domain string) (*model.SiteDomain, error) {
	var site model.SiteDomain
	err := r.db.Where("domain = ?", domain).First(&site).Error
	if err != nil {
		return nil, err
	}
	return &site, nil
}

// List 获取域名列表（分页）
func (r *SiteRepo) List(page, pageSize int, cluster string, activeOnly bool) ([]model.SiteDomain, int64, error) {
	var sites []model.SiteDomain
	var total int64

	db := r.db.Model(&model.SiteDomain{})
	if cluster != "" {
		db = db.Where("cluster = ?", cluster)
	}
	if activeOnly {
		db = db.Where("is_active = ?", true)
	}

	db.Count(&total)

	err := db.Offset((page - 1) * pageSize).Limit(pageSize).
		Order("created_at DESC").
		Find(&sites).Error
	return sites, total, err
}

// ListByCluster 按 A/B 队筛选域名
func (r *SiteRepo) ListByCluster(cluster string) ([]model.SiteDomain, error) {
	var sites []model.SiteDomain
	err := r.db.Where("cluster = ? AND is_active = ?", cluster, true).
		Find(&sites).Error
	return sites, err
}

// UpdateHealthStatus 更新健康状态
func (r *SiteRepo) UpdateHealthStatus(domainID uuid.UUID, status string, responseTimeMs int) error {
	now := time.Now()
	return r.db.Model(&model.SiteDomain{}).Where("id = ?", domainID).Updates(map[string]interface{}{
		"health_status":      status,
		"last_health_check":  &now,
	}).Error
}

// IncrementTraffic 增加流量计数
func (r *SiteRepo) IncrementTraffic(domainID uuid.UUID) error {
	return r.db.Model(&model.SiteDomain{}).Where("id = ?", domainID).Updates(map[string]interface{}{
		"daily_traffic":   gorm.Expr("daily_traffic + 1"),
		"monthly_traffic": gorm.Expr("monthly_traffic + 1"),
	}).Error
}

// CreateHealthLog 创建健康检查日志
func (r *SiteRepo) CreateHealthLog(log *model.SiteHealthLog) error {
	return r.db.Create(log).Error
}

// ListHealthLogs 获取健康检查日志
func (r *SiteRepo) ListHealthLogs(domainID uuid.UUID, limit int) ([]model.SiteHealthLog, error) {
	var logs []model.SiteHealthLog
	err := r.db.Where("domain_id = ?", domainID).
		Order("checked_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

// CreateLinkAudit 创建交叉链接审计记录
func (r *SiteRepo) CreateLinkAudit(audit *model.DomainLinkAudit) error {
	return r.db.Create(audit).Error
}

// ResolveLinkAudit 标记审计记录已处理
func (r *SiteRepo) ResolveLinkAudit(id uuid.UUID) error {
	return r.db.Model(&model.DomainLinkAudit{}).Where("id = ?", id).
		Update("is_resolved", true).Error
}

// CheckCrossClusterLinks 检测 A/B 队交叉链接
func (r *SiteRepo) CheckCrossClusterLinks() ([]model.DomainLinkAudit, error) {
	var audits []model.DomainLinkAudit

	// 查询所有未处理的交叉链接审计
	err := r.db.Where("is_resolved = ? AND link_type = ?", false, "cross_cluster").
		Order("detected_at DESC").
		Find(&audits).Error
	return audits, err
}

// GetAllActiveDomains 获取所有活跃域名
func (r *SiteRepo) GetAllActiveDomains() ([]model.SiteDomain, error) {
	var sites []model.SiteDomain
	err := r.db.Where("is_active = ?", true).Find(&sites).Error
	return sites, err
}
