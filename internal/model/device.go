package model

import (
	"time"

	"github.com/google/uuid"
)

// DeviceFingerprint 设备指纹模型
type DeviceFingerprint struct {
	ID               uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	FingerprintHash  string    `gorm:"type:varchar(64);uniqueIndex;not null;comment:指纹哈希" json:"fingerprint_hash"`
	UserAgent        string    `gorm:"type:varchar(500);comment:用户代理" json:"user_agent"`
	ScreenResolution string    `gorm:"type:varchar(20);comment:屏幕分辨率" json:"screen_resolution"`
	Language         string    `gorm:"type:varchar(10);comment:语言" json:"language"`
	Timezone         string    `gorm:"type:varchar(50);comment:时区" json:"timezone"`
	FirstSeenAt      time.Time `gorm:"autoCreateTime;comment:首次访问时间" json:"first_seen_at"`
	LastSeenAt       time.Time `gorm:"autoUpdateTime;comment:最近访问时间" json:"last_seen_at"`
	IsBanned         bool      `gorm:"default:false;comment:是否封禁" json:"is_banned"`
}

// TableName 指定表名
func (DeviceFingerprint) TableName() string {
	return "device_fingerprints"
}

// BeforeCreate 创建前钩子，确保 UUID 生成
func (d *DeviceFingerprint) BeforeCreate() error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}

// DeviceUnlockRecord 设备解锁记录模型
type DeviceUnlockRecord struct {
	ID            uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	FingerprintID uuid.UUID  `gorm:"type:uuid;index;not null;comment:设备指纹 ID" json:"fingerprint_id"`
	VideoID       uuid.UUID  `gorm:"type:uuid;index;not null;comment:视频 ID" json:"video_id"`
	UnlockType    string     `gorm:"type:varchar(20);not null;comment:解锁类型 coin/share/free" json:"unlock_type"`
	UnlockedAt    time.Time  `gorm:"autoCreateTime;comment:解锁时间" json:"unlocked_at"`
	ExpiresAt     *time.Time `gorm:"comment:过期时间" json:"expires_at"`
}

// TableName 指定表名
func (DeviceUnlockRecord) TableName() string {
	return "device_unlock_records"
}

// BeforeCreate 创建前钩子，确保 UUID 生成
func (d *DeviceUnlockRecord) BeforeCreate() error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}

// DeviceCoinBalance 设备硬币余额模型
type DeviceCoinBalance struct {
	FingerprintID uuid.UUID `gorm:"type:uuid;primary_key;comment:设备指纹 ID" json:"fingerprint_id"`
	Balance       int64     `gorm:"default:0;comment:当前余额" json:"balance"`
	TotalEarned   int64     `gorm:"default:0;comment:累计获得" json:"total_earned"`
	TotalSpent    int64     `gorm:"default:0;comment:累计消费" json:"total_spent"`
}

// TableName 指定表名
func (DeviceCoinBalance) TableName() string {
	return "device_coin_balances"
}
