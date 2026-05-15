package model

import (
	"time"

	"github.com/google/uuid"
)

// Rank 排行榜模型
type Rank struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	VideoID   uuid.UUID `gorm:"type:uuid;index;not null;comment:视频 ID" json:"video_id"`
	Type      string    `gorm:"type:varchar(20);index;not null;comment:排行榜类型 daily/weekly/monthly" json:"type"`
	Score     int       `gorm:"default:0;comment:热度分数" json:"score"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// 关联
	Video Video `gorm:"foreignKey:VideoID" json:"video,omitempty"`
}

// TableName 指定表名
func (Rank) TableName() string {
	return "ranks"
}

// BeforeCreate 创建前钩子，确保 UUID 生成
func (r *Rank) BeforeCreate() error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}
