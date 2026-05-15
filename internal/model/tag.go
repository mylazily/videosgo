package model

import (
	"time"

	"github.com/google/uuid"
)

// Tag 标签模型
type Tag struct {
	ID          uuid.UUID    `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name        string       `gorm:"type:varchar(100);not null;uniqueIndex;comment:标签名称" json:"name"`
	Slug        string       `gorm:"type:varchar(100);not null;uniqueIndex;comment:URL 友好标识" json:"slug"`
	Description string       `gorm:"type:varchar(500);comment:标签描述" json:"description"`
	Icon        string       `gorm:"type:varchar(200);comment:标签图标" json:"icon"`
	Color       string       `gorm:"type:varchar(20);comment:标签颜色" json:"color"`
	SortOrder   int          `gorm:"default:0;comment:排序权重" json:"sort_order"`
	VideoCount  int64        `gorm:"default:0;comment:关联视频数" json:"video_count"`
	Status      string       `gorm:"type:varchar(20);default:active;comment:状态" json:"status"`
	ParentID    *uuid.UUID   `gorm:"type:uuid;index;comment:父标签 ID" json:"parent_id"`
	CreatedAt   time.Time    `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time    `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (Tag) TableName() string {
	return "tags"
}

// BeforeCreate 创建前钩子，确保 UUID 生成
func (t *Tag) BeforeCreate() error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

// VideoTag 视频标签关联模型
type VideoTag struct {
	ID     uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	VideoID uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_video_tag;not null;comment:视频 ID" json:"video_id"`
	TagID   uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_video_tag;not null;comment:标签 ID" json:"tag_id"`

	// 关联
	Tag  Tag   `gorm:"foreignKey:TagID" json:"tag,omitempty"`
	Video Video `gorm:"foreignKey:VideoID" json:"video,omitempty"`
}

// TableName 指定表名
func (VideoTag) TableName() string {
	return "video_tags"
}

// BeforeCreate 创建前钩子，确保 UUID 生成
func (v *VideoTag) BeforeCreate() error {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	return nil
}
