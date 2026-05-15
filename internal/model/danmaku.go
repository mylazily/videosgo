package model

import (
	"time"

	"github.com/google/uuid"
)

// Danmaku 弹幕模型
type Danmaku struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	VideoID   uuid.UUID `gorm:"type:uuid;index;not null;comment:视频 ID" json:"video_id"`
	EpisodeID uuid.UUID `gorm:"type:uuid;index;comment:集数 ID" json:"episode_id"`
	UserID    uuid.UUID `gorm:"type:uuid;index;comment:用户 ID" json:"user_id"`
	Time      float64   `gorm:"not null;comment:弹幕出现时间（秒）" json:"time"`
	Type      int       `gorm:"default:1;comment:弹幕类型 1右到左 2顶部 3底部" json:"type"`
	Color     string    `gorm:"type:varchar(20);default:#FFFFFF;comment:弹幕颜色" json:"color"`
	Content   string    `gorm:"type:varchar(500);not null;comment:弹幕内容" json:"content"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TableName 指定表名
func (Danmaku) TableName() string {
	return "danmakus"
}

// BeforeCreate 创建前钩子，确保 UUID 生成
func (d *Danmaku) BeforeCreate() error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}
