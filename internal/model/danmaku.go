package model

import "time"

// Danmaku 弹幕
type Danmaku struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	VideoID   uint      `gorm:"index;not null;comment:视频 ID" json:"video_id"`
	EpisodeID uint      `gorm:"index;comment:集数 ID" json:"episode_id"`
	UserID    uint      `gorm:"index;comment:用户 ID" json:"user_id"`
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
