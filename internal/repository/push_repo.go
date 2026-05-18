package repository

import (
	"github.com/google/uuid"
	"videosgo/internal/model"
	"gorm.io/gorm"
)

// PushRepo 推送数据仓库
type PushRepo struct {
	db *gorm.DB
}

// NewPushRepo 创建推送仓库
func NewPushRepo(db *gorm.DB) *PushRepo {
	return &PushRepo{db: db}
}

// CreateSubscription 创建推送订阅
func (r *PushRepo) CreateSubscription(sub *model.PushSubscription) error {
	return r.db.Create(sub).Error
}

// GetSubscription 根据 ID 获取订阅
func (r *PushRepo) GetSubscription(id uuid.UUID) (*model.PushSubscription, error) {
	var sub model.PushSubscription
	err := r.db.First(&sub, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

// GetSubscriptionByEndpoint 根据端点获取订阅
func (r *PushRepo) GetSubscriptionByEndpoint(endpoint string) (*model.PushSubscription, error) {
	var sub model.PushSubscription
	err := r.db.Where("endpoint = ?", endpoint).First(&sub).Error
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

// DeleteSubscription 删除订阅
func (r *PushRepo) DeleteSubscription(id uuid.UUID) error {
	return r.db.Delete(&model.PushSubscription{}, "id = ?", id).Error
}

// UpdateSubscription 更新订阅
func (r *PushRepo) UpdateSubscription(sub *model.PushSubscription) error {
	return r.db.Save(sub).Error
}

// ListActiveSubscriptions 获取所有活跃订阅
func (r *PushRepo) ListActiveSubscriptions() ([]model.PushSubscription, error) {
	var subs []model.PushSubscription
	err := r.db.Where("is_active = ?", true).Find(&subs).Error
	return subs, err
}

// CreateNotification 创建推送通知
func (r *PushRepo) CreateNotification(notif *model.PushNotification) error {
	return r.db.Create(notif).Error
}

// GetNotification 根据 ID 获取通知
func (r *PushRepo) GetNotification(id uuid.UUID) (*model.PushNotification, error) {
	var notif model.PushNotification
	err := r.db.First(&notif, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &notif, nil
}

// UpdateNotificationStatus 更新通知状态
func (r *PushRepo) UpdateNotificationStatus(id uuid.UUID, status string, totalSent int) error {
	return r.db.Model(&model.PushNotification{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":     status,
		"total_sent": totalSent,
	}).Error
}

// IncrementNotificationClick 增加通知点击数
func (r *PushRepo) IncrementNotificationClick(id uuid.UUID) error {
	return r.db.Model(&model.PushNotification{}).Where("id = ?", id).
		UpdateColumn("total_clicked", gorm.Expr("total_clicked + 1")).Error
}

// LogClick 记录点击日志
func (r *PushRepo) LogClick(log *model.PushClickLog) error {
	return r.db.Create(log).Error
}

// GetNotificationStats 获取通知统计
func (r *PushRepo) GetNotificationStats(limit int) ([]model.PushNotification, error) {
	var notifs []model.PushNotification
	err := r.db.Order("created_at DESC").Limit(limit).Find(&notifs).Error
	return notifs, err
}
