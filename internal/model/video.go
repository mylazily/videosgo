package model

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Video 视频模型
type Video struct {
	ID          string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()" db:"id"`
	Title       string    `json:"title" gorm:"type:varchar(500);not null" db:"title"`
	SubTitle    string    `json:"sub_title" gorm:"type:varchar(500);comment:副标题" db:"sub_title"`
	Description string    `json:"description" gorm:"type:text" db:"description"`
	CoverURL    string    `json:"cover_url" gorm:"type:varchar(1000);comment:封面 URL" db:"cover_url"`
	VideoURL    string    `json:"video_url" gorm:"type:varchar(1000);comment:视频 URL" db:"video_url"`
	SourceID    string    `json:"source_id" gorm:"type:uuid;index;comment:来源 ID" db:"source_id"`
	Category    string    `json:"category" gorm:"type:varchar(100);index;comment:分类" db:"category"`
	Year        int16     `json:"year" gorm:"type:smallint;comment:年份" db:"year"`
	Area        string    `json:"area" gorm:"type:varchar(100);comment:地区" db:"area"`
	Director    string    `json:"director" gorm:"type:varchar(500);comment:导演" db:"director"`
	Actors      string    `json:"actors" gorm:"type:text;comment:演员" db:"actors"`
	Tags        []string  `json:"tags" gorm:"type:text;comment:标签" db:"tags"`
	Remarks     string    `json:"remarks" gorm:"type:varchar(200);comment:备注/状态" db:"remarks"`
	PlayLinks   []string  `json:"play_links" gorm:"type:text;comment:播放链接" db:"play_links"`
	Status      string    `json:"status" gorm:"type:varchar(20);default:active;index;comment:状态" db:"status"`
	CleanTitle  string    `json:"clean_title" gorm:"type:varchar(500);index;comment:清洗后标题(用于去重)" db:"clean_title"`
	PlayLines   PlayLinesJSON `json:"play_lines" gorm:"type:jsonb;comment:聚合播放线路" db:"play_lines"`
	SourceCount int       `json:"source_count" gorm:"default:1;comment:资源站数量" db:"source_count"`
	DomainPool  []string  `json:"domain_pool" gorm:"type:text;comment:域名池" db:"domain_pool"`
	SharedPath  string    `json:"shared_path" gorm:"type:varchar(500);comment:共享路径" db:"shared_path"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime" db:"updated_at"`
}

// TableName 指定表名
func (Video) TableName() string {
	return "videos"
}

// BeforeCreate 创建前钩子
func (v *Video) BeforeCreate() error {
	if v.ID == "" {
		v.ID = uuid.New().String()
	}
	return nil
}

// Episode 剧集模型
type Episode struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	VideoID   uuid.UUID `gorm:"type:uuid;index;not null;comment:视频 ID" json:"video_id"`
	Name      string    `gorm:"type:varchar(200);comment:集名" json:"name"`
	URL       string    `gorm:"type:text;comment:播放地址" json:"url"`
	URLType   string    `gorm:"type:varchar(20);default:m3u8;comment:URL 类型" json:"url_type"`
	SourceID  uuid.UUID `gorm:"type:uuid;index;comment:来源 ID" json:"source_id"`
	EpisodeNo int       `gorm:"default:0;comment:集号" json:"episode_no"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TableName 指定表名
func (Episode) TableName() string {
	return "episodes"
}

// BeforeCreate 创建前钩子
func (e *Episode) BeforeCreate() error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}

// PlayLineJSON 单条播放线路
type PlayLineJSON struct {
	SourceName string `json:"source_name"`
	M3U8URL    string `json:"m3u8_url"`
	Domain     string `json:"domain,omitempty"`
	Path       string `json:"path,omitempty"`
	Format     string `json:"format,omitempty"`
	Quality    string `json:"quality,omitempty"`
	Language   string `json:"language,omitempty"`
}

// PlayLinesJSON 播放线路列表
type PlayLinesJSON []PlayLineJSON

// Value 实现 driver.Valuer 接口
func (p PlayLinesJSON) Value() (interface{}, error) {
	if len(p) == 0 {
		return nil, nil
	}
	return json.Marshal(p)
}

// Scan 实现 sql.Scanner 接口
func (p *PlayLinesJSON) Scan(value interface{}) error {
	if value == nil {
		*p = nil
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("PlayLinesJSON: cannot scan %T", value)
	}

	return json.Unmarshal(bytes, p)
}
