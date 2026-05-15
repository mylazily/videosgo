package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// StringArray 字符串数组类型（兼容 PostgreSQL text[] 和 JSON 序列化）
type StringArray []string

// Scan 实现 sql.Scanner 接口
func (s *StringArray) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}

	switch v := value.(type) {
	case string:
		return json.Unmarshal([]byte(v), s)
	case []byte:
		return json.Unmarshal(v, s)
	}

	return fmt.Errorf("无法扫描 StringArray: %v", value)
}

// Value 实现 driver.Valuer 接口
func (s StringArray) Value() (driver.Value, error) {
	if s == nil {
		return "[]", nil
	}
	data, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	return string(data), nil
}

// MarshalJSON 实现 json.Marshaler 接口
func (s StringArray) MarshalJSON() ([]byte, error) {
	if s == nil {
		return []byte("[]"), nil
	}
	return json.Marshal([]string(s))
}

// UnmarshalJSON 实现 json.Unmarshaler 接口
func (s *StringArray) UnmarshalJSON(data []byte) error {
	var arr []string
	if err := json.Unmarshal(data, &arr); err != nil {
		return err
	}
	*s = arr
	return nil
}

// Video 视频模型
type Video struct {
	ID          uuid.UUID   `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Title       string      `gorm:"type:varchar(200);not null;index;comment:标题" json:"title"`
	SubTitle    string      `gorm:"type:varchar(200);comment:副标题" json:"sub_title"`
	Cover       string      `gorm:"type:varchar(500);comment:封面图" json:"cover"`
	Description string      `gorm:"type:text;comment:简介" json:"description"`
	CategoryID  int         `gorm:"index;default:0;comment:分类 ID" json:"category_id"`
	Category    string      `gorm:"type:varchar(50);index;comment:分类名称" json:"category"`
	Year        string      `gorm:"type:varchar(10);comment:年份" json:"year"`
	Area        string      `gorm:"type:varchar(50);comment:地区" json:"area"`
	Director    string      `gorm:"type:varchar(200);comment:导演" json:"director"`
	Actors      string      `gorm:"type:text;comment:演员" json:"actors"`
	Tags        StringArray `gorm:"type:jsonb;comment:标签" json:"tags"`
	Remarks     string      `gorm:"type:varchar(100);comment:备注" json:"remarks"`
	PlayLinks   StringArray `gorm:"type:jsonb;comment:播放链接 JSONB" json:"play_links"`
	Status      string      `gorm:"type:varchar(20);default:active;comment:状态" json:"status"`
	SourceID    uuid.UUID   `gorm:"type:uuid;index;comment:来源采集源 ID" json:"source_id"`
	ViewCount   int64       `gorm:"default:0;comment:播放量" json:"view_count"`
	LikeCount   int64       `gorm:"default:0;comment:点赞数" json:"like_count"`
	Score       float64     `gorm:"default:0;comment:评分" json:"score"`
	CreatedAt   time.Time   `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time   `gorm:"autoUpdateTime" json:"updated_at"`

	// 关联
	Episodes []Episode `gorm:"foreignKey:VideoID" json:"episodes,omitempty"`
}

// TableName 指定表名
func (Video) TableName() string {
	return "videos"
}

// BeforeCreate 创建前钩子，确保 UUID 生成
func (v *Video) BeforeCreate() error {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	return nil
}

// VideoSource 视频来源
type VideoSource struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	VideoID   uuid.UUID `gorm:"type:uuid;index;not null;comment:视频 ID" json:"video_id"`
	SourceID  uuid.UUID `gorm:"type:uuid;index;not null;comment:采集源 ID" json:"source_id"`
	SourceURL string    `gorm:"type:varchar(500);comment:来源 URL" json:"source_url"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TableName 指定表名
func (VideoSource) TableName() string {
	return "video_sources"
}

// BeforeCreate 创建前钩子，确保 UUID 生成
func (v *VideoSource) BeforeCreate() error {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	return nil
}

// Episode 剧集
type Episode struct {
	ID       uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	VideoID  uuid.UUID `gorm:"type:uuid;index;not null;comment:视频 ID" json:"video_id"`
	Name     string    `gorm:"type:varchar(100);comment:集名" json:"name"`
	EpIndex  int       `gorm:"default:0;comment:集序号" json:"ep_index"`
	URL      string    `gorm:"type:varchar(500);comment:播放地址" json:"url"`
	URLType  string    `gorm:"type:varchar(20);default:m3u8;comment:播放类型" json:"url_type"`
	SourceID uuid.UUID `gorm:"type:uuid;comment:来源采集源 ID" json:"source_id"`
}

// TableName 指定表名
func (Episode) TableName() string {
	return "episodes"
}

// BeforeCreate 创建前钩子，确保 UUID 生成
func (e *Episode) BeforeCreate() error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}

// EpisodeSource 剧集来源（多源聚合）
type EpisodeSource struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	EpisodeID  uuid.UUID `gorm:"type:uuid;index;not null;comment:剧集 ID" json:"episode_id"`
	SourceID   uuid.UUID `gorm:"type:uuid;index;not null;comment:采集源 ID" json:"source_id"`
	SourceName string    `gorm:"type:varchar(100);comment:来源名称" json:"source_name"`
	URL        string    `gorm:"type:varchar(500);comment:播放地址" json:"url"`
	Alive      bool      `gorm:"default:true;comment:链接是否存活" json:"alive"`
}

// TableName 指定表名
func (EpisodeSource) TableName() string {
	return "episode_sources"
}

// BeforeCreate 创建前钩子，确保 UUID 生成
func (e *EpisodeSource) BeforeCreate() error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}

// UserWatchHistory 用户观看历史
type UserWatchHistory struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_user_video;not null;comment:用户 ID" json:"user_id"`
	VideoID   uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_user_video;not null;comment:视频 ID" json:"video_id"`
	Progress  float64   `gorm:"default:0;comment:观看进度（秒）" json:"progress"`
	Duration  float64   `gorm:"default:0;comment:总时长（秒）" json:"duration"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// 关联
	Video Video `gorm:"foreignKey:VideoID" json:"video,omitempty"`
}

// TableName 指定表名
func (UserWatchHistory) TableName() string {
	return "user_watch_histories"
}

// BeforeCreate 创建前钩子，确保 UUID 生成
func (u *UserWatchHistory) BeforeCreate() error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// SearchHot 热搜
type SearchHot struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Keyword   string    `gorm:"type:varchar(100);uniqueIndex;not null;comment:搜索关键词" json:"keyword"`
	Count     int64     `gorm:"default:0;comment:搜索次数" json:"count"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (SearchHot) TableName() string {
	return "search_hots"
}

// BeforeCreate 创建前钩子，确保 UUID 生成
func (s *SearchHot) BeforeCreate() error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}
