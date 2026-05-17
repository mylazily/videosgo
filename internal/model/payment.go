package model

import (
	"time"

	"github.com/google/uuid"
)

// PaymentChannel 支付渠道模型
type PaymentChannel struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ChannelName string    `gorm:"type:varchar(100);not null;comment:渠道名称" json:"channel_name"`
	ChannelType string    `gorm:"type:varchar(50);not null;comment:渠道类型 alipay/wechat/crypto" json:"channel_type"`
	Config      JSONB     `gorm:"type:jsonb;comment:渠道配置" json:"-"`
	IsActive    bool      `gorm:"default:true;comment:是否启用" json:"is_active"`
	MinAmount   float64   `gorm:"default:0.01;comment:最小金额" json:"min_amount"`
	MaxAmount   float64   `gorm:"default:99999;comment:最大金额" json:"max_amount"`
	FeeRate     float64   `gorm:"default:0;comment:费率" json:"fee_rate"`
	SortOrder   int       `gorm:"default:0;comment:排序" json:"sort_order"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (PaymentChannel) TableName() string {
	return "payment_channels"
}

// BeforeCreate 创建前钩子，确保 UUID 生成
func (p *PaymentChannel) BeforeCreate() error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// PaymentOrder 支付订单模型
type PaymentOrder struct {
	ID            uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	OrderNo       string     `gorm:"type:varchar(64);uniqueIndex;not null;comment:订单号" json:"order_no"`
	FingerprintID *uuid.UUID `gorm:"type:uuid;index;comment:设备指纹 ID" json:"fingerprint_id"`
	ChannelID     uuid.UUID  `gorm:"type:uuid;index;comment:支付渠道 ID" json:"channel_id"`
	ProductType   string     `gorm:"type:varchar(50);not null;comment:产品类型 vip/coin" json:"product_type"`
	ProductID     string     `gorm:"type:varchar(100);comment:产品 ID" json:"product_id"`
	ProductName   string     `gorm:"type:varchar(200);comment:产品名称" json:"product_name"`
	Amount        float64    `gorm:"not null;comment:金额" json:"amount"`
	FeeAmount     float64    `gorm:"default:0;comment:手续费" json:"fee_amount"`
	Status        string     `gorm:"type:varchar(20);default:pending;comment:状态 pending/paid/failed/expired/refunded" json:"status"`
	PaymentNo     string     `gorm:"type:varchar(100);comment:第三方支付单号" json:"payment_no"`
	PaidAt        *time.Time `gorm:"comment:支付时间" json:"paid_at"`
	ExpiresAt     time.Time  `gorm:"comment:过期时间" json:"expires_at"`
	CreatedAt     time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (PaymentOrder) TableName() string {
	return "payment_orders"
}

// BeforeCreate 创建前钩子，确保 UUID 生成
func (p *PaymentOrder) BeforeCreate() error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// VIPSubscription VIP 订阅模型
type VIPSubscription struct {
	ID            uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	FingerprintID *uuid.UUID `gorm:"type:uuid;index;comment:设备指纹 ID" json:"fingerprint_id"`
	PlanType      string     `gorm:"type:varchar(50);not null;comment:套餐类型 monthly/quarterly/yearly" json:"plan_type"`
	StartAt       time.Time  `gorm:"comment:开始时间" json:"start_at"`
	ExpiresAt     time.Time  `gorm:"index;comment:过期时间" json:"expires_at"`
	IsActive      bool       `gorm:"default:true;comment:是否生效" json:"is_active"`
	AutoRenew     bool       `gorm:"default:false;comment:是否自动续费" json:"auto_renew"`
	CreatedAt     time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (VIPSubscription) TableName() string {
	return "vip_subscriptions"
}

// BeforeCreate 创建前钩子，确保 UUID 生成
func (v *VIPSubscription) BeforeCreate() error {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	return nil
}
