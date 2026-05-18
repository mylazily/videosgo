package repository

import (
	"time"

	"github.com/google/uuid"
	"videosgo/internal/model"
	"gorm.io/gorm"
)

// PaymentRepo 支付数据仓库
type PaymentRepo struct {
	db *gorm.DB
}

// NewPaymentRepo 创建支付仓库
func NewPaymentRepo(db *gorm.DB) *PaymentRepo {
	return &PaymentRepo{db: db}
}

// ========== PaymentChannel ==========

// CreateChannel 创建支付渠道
func (r *PaymentRepo) CreateChannel(channel *model.PaymentChannel) error {
	return r.db.Create(channel).Error
}

// ListChannels 获取所有支付渠道
func (r *PaymentRepo) ListChannels() ([]model.PaymentChannel, error) {
	var channels []model.PaymentChannel
	err := r.db.Where("is_active = ?", true).Order("sort_order ASC").Find(&channels).Error
	if err != nil {
		return nil, err
	}
	return channels, nil
}

// GetChannelByID 根据 ID 获取支付渠道
func (r *PaymentRepo) GetChannelByID(id uuid.UUID) (*model.PaymentChannel, error) {
	var channel model.PaymentChannel
	err := r.db.First(&channel, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &channel, nil
}

// ========== PaymentOrder ==========

// CreateOrder 创建支付订单
func (r *PaymentRepo) CreateOrder(order *model.PaymentOrder) error {
	return r.db.Create(order).Error
}

// GetOrderByNo 根据订单号获取订单
func (r *PaymentRepo) GetOrderByNo(orderNo string) (*model.PaymentOrder, error) {
	var order model.PaymentOrder
	err := r.db.Where("order_no = ?", orderNo).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// GetOrderByID 根据 ID 获取订单
func (r *PaymentRepo) GetOrderByID(id uuid.UUID) (*model.PaymentOrder, error) {
	var order model.PaymentOrder
	err := r.db.First(&order, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// UpdateOrderStatus 更新订单状态
func (r *PaymentRepo) UpdateOrderStatus(orderNo string, status string, paymentNo string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if paymentNo != "" {
		updates["payment_no"] = paymentNo
	}
	if status == "paid" {
		now := time.Now()
		updates["paid_at"] = &now
	}
	return r.db.Model(&model.PaymentOrder{}).Where("order_no = ?", orderNo).Updates(updates).Error
}

// ========== VIPSubscription ==========

// CreateVIP 创建 VIP 订阅
func (r *PaymentRepo) CreateVIP(sub *model.VIPSubscription) error {
	return r.db.Create(sub).Error
}

// GetActiveVIP 获取活跃 VIP 订阅
func (r *PaymentRepo) GetActiveVIP(fingerprintID uuid.UUID) (*model.VIPSubscription, error) {
	var sub model.VIPSubscription
	err := r.db.Where("fingerprint_id = ? AND is_active = ? AND expires_at > ?",
		fingerprintID, true, time.Now()).
		Order("expires_at DESC").
		First(&sub).Error
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

// ExpireVIP 使 VIP 过期
func (r *PaymentRepo) ExpireVIP(id uuid.UUID) error {
	return r.db.Model(&model.VIPSubscription{}).Where("id = ?", id).Updates(map[string]interface{}{
		"is_active":  false,
		"auto_renew": false,
	}).Error
}

// UpdateVIP 更新 VIP 订阅
func (r *PaymentRepo) UpdateVIP(sub *model.VIPSubscription) error {
	return r.db.Save(sub).Error
}
