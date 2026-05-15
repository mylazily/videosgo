package model

import (
	"time"

	"github.com/google/uuid"
)

// ShareLink 分享链接模型
type ShareLink struct {
	ID                    uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	VideoID               uuid.UUID  `gorm:"type:uuid;index;not null;comment:视频 ID" json:"video_id"`
	CreatorFingerprintID  uuid.UUID  `gorm:"type:uuid;index;not null;comment:创建者设备指纹 ID" json:"creator_fingerprint_id"`
	ShareCode             string     `gorm:"type:varchar(8);uniqueIndex;not null;comment:分享码" json:"share_code"`
	ClickCount            int64      `gorm:"default:0;comment:点击次数" json:"click_count"`
	UnlockCount           int64      `gorm:"default:0;comment:解锁次数" json:"unlock_count"`
	MaxUnlocks            int        `gorm:"default:10;comment:最大解锁次数" json:"max_unlocks"`
	RewardType            string     `gorm:"type:varchar(20);default:coin;comment:奖励类型" json:"reward_type"`
	RewardAmount          int64      `gorm:"default:1;comment:奖励数量" json:"reward_amount"`
	Status                string     `gorm:"type:varchar(20);default:active;comment:状态" json:"status"`
	ExpiresAt             *time.Time `gorm:"comment:过期时间" json:"expires_at"`
	CreatedAt             time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt             time.Time  `gorm:"autoUpdateTime" json:"updated_at"`

	// 关联
	Video Video `gorm:"foreignKey:VideoID" json:"video,omitempty"`
}

// TableName 指定表名
func (ShareLink) TableName() string {
	return "share_links"
}

// BeforeCreate 创建前钩子，确保 UUID 生成
func (s *ShareLink) BeforeCreate() error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// ShareClick 分享点击记录模型
type ShareClick struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ShareLinkID uuid.UUID `gorm:"type:uuid;index;not null;comment:分享链接 ID" json:"share_link_id"`
	FingerprintID uuid.UUID `gorm:"type:uuid;index;comment:点击者设备指纹 ID" json:"fingerprint_id"`
	IPAddress   string    `gorm:"type:varchar(45);comment:IP 地址" json:"ip_address"`
	UserAgent   string    `gorm:"type:varchar(500);comment:用户代理" json:"user_agent"`
	ClickedAt   time.Time `gorm:"autoCreateTime;comment:点击时间" json:"clicked_at"`
}

// TableName 指定表名
func (ShareClick) TableName() string {
	return "share_clicks"
}

// BeforeCreate 创建前钩子，确保 UUID 生成
func (s *ShareClick) BeforeCreate() error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}
