package model

import (
	"time"

	"github.com/google/uuid"
)

// PushSubscription 推送订阅
type PushSubscription struct {
	ID            uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	FingerprintID string     `gorm:"type:varchar(100);index;not null;comment:设备指纹 ID" json:"fingerprint_id"`
	Endpoint      string     `gorm:"type:varchar(500);not null;comment:推送端点 URL" json:"endpoint"`
	P256DHKey     string     `gorm:"type:text;comment:P-256 DH 公钥(base64)" json:"p256dh_key"`
	AuthKey       string     `gorm:"type:text;comment:认证密钥(base64)" json:"auth_key"`
	UserAgent     string     `gorm:"type:varchar(500);comment:用户代理" json:"user_agent"`
	IsActive      bool       `gorm:"default:true;index;comment:是否活跃" json:"is_active"`
	LastSentAt    *time.Time `gorm:"comment:最后发送时间" json:"last_sent_at"`
	CreatedAt     time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (PushSubscription) TableName() string {
	return "push_subscriptions"
}

// BeforeCreate 创建前钩子，确保 UUID 生成
func (p *PushSubscription) BeforeCreate() error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// PushNotification 推送通知
type PushNotification struct {
	ID             uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Title          string     `gorm:"type:varchar(200);not null;comment:通知标题" json:"title"`
	Body           string     `gorm:"type:text;comment:通知内容" json:"body"`
	Icon           string     `gorm:"type:varchar(500);comment:图标 URL" json:"icon"`
	Link           string     `gorm:"type:varchar(500);comment:点击跳转链接" json:"link"`
	Tag            string     `gorm:"type:varchar(100);comment:通知标签" json:"tag"`
	TargetType     string     `gorm:"type:varchar(50);default:all;comment:目标类型(all/video/tag)" json:"target_type"`
	TargetVideoID  string     `gorm:"type:varchar(100);comment:目标视频 ID" json:"target_video_id"`
	TotalSent      int        `gorm:"default:0;comment:总发送数" json:"total_sent"`
	TotalClicked   int        `gorm:"default:0;comment:总点击数" json:"total_clicked"`
	Status         string     `gorm:"type:varchar(20);default:pending;comment:状态(pending/sending/completed/failed)" json:"status"`
	SentAt         *time.Time `gorm:"comment:发送时间" json:"sent_at"`
	CreatedAt      time.Time  `gorm:"autoCreateTime" json:"created_at"`
}

// TableName 指定表名
func (PushNotification) TableName() string {
	return "push_notifications"
}

// BeforeCreate 创建前钩子，确保 UUID 生成
func (p *PushNotification) BeforeCreate() error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// PushClickLog 推送点击日志
type PushClickLog struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	NotificationID uuid.UUID `gorm:"type:uuid;index;not null;comment:通知 ID" json:"notification_id"`
	SubscriptionID uuid.UUID `gorm:"type:uuid;index;not null;comment:订阅 ID" json:"subscription_id"`
	ClickedAt      time.Time `gorm:"autoCreateTime;comment:点击时间" json:"clicked_at"`
}

// TableName 指定表名
func (PushClickLog) TableName() string {
	return "push_click_logs"
}

// BeforeCreate 创建前钩子，确保 UUID 生成
func (p *PushClickLog) BeforeCreate() error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}
