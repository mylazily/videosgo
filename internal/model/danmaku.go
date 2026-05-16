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
	Time      string    `gorm:"type:decimal(10,3);not null;comment:弹幕出现时间（秒）" json:"time"`
	Type      string    `gorm:"type:varchar(20);default:'scroll';comment:弹幕类型 scroll/top/bottom" json:"type"`
	Color     string    `gorm:"type:varchar(20);default:#FFFFFF;comment:弹幕颜色" json:"color"`
	Content   string    `gorm:"type:varchar(500);not null;comment:弹幕内容" json:"content"`
	Status    bool      `gorm:"default:true;comment:状态" json:"status"`
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
