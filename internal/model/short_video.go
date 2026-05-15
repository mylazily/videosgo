package model

import (
	"time"

	"github.com/google/uuid"
)

// ShortVideo 短视频模型
type ShortVideo struct {
	ID           uuid.UUID   `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Title        string      `gorm:"type:varchar(200);not null;index;comment:标题" json:"title"`
	Description  string      `gorm:"type:text;comment:描述" json:"description"`
	CoverURL     string      `gorm:"type:varchar(500);comment:封面图 URL" json:"cover_url"`
	PreviewURL   string      `gorm:"type:varchar(500);comment:预览图 URL" json:"preview_url"`
	VideoURL     string      `gorm:"type:varchar(500);not null;comment:视频地址" json:"video_url"`
	Duration     int         `gorm:"default:0;comment:时长（秒）" json:"duration"`
	SourceFrom   string      `gorm:"type:varchar(50);comment:来源平台" json:"source_from"`
	SourceID     string      `gorm:"type:varchar(100);comment:来源平台 ID" json:"source_id"`
	ViewCount    int64       `gorm:"default:0;comment:播放量" json:"view_count"`
	LikeCount    int64       `gorm:"default:0;comment:点赞数" json:"like_count"`
	ShareCount   int64       `gorm:"default:0;comment:分享数" json:"share_count"`
	Tags         StringArray `gorm:"type:jsonb;comment:标签" json:"tags"`
	Status       string      `gorm:"type:varchar(20);default:active;comment:状态" json:"status"`
	CreatedAt    time.Time   `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time   `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (ShortVideo) TableName() string {
	return "short_videos"
}

// BeforeCreate 创建前钩子，确保 UUID 生成
func (s *ShortVideo) BeforeCreate() error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}
