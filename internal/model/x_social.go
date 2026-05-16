package model

import (
	"time"

	"github.com/google/uuid"
)

// XAccount X.com 账号模型
type XAccount struct {
	ID                uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AccountID         string     `gorm:"type:varchar(100);uniqueIndex;not null;comment:账号 ID" json:"account_id"`
	Username          string     `gorm:"type:varchar(100);comment:用户名" json:"username"`
	AccessToken       string     `gorm:"type:text;comment:访问令牌" json:"access_token"`
	AccessTokenSecret string     `gorm:"type:varchar(500);comment:访问令牌密钥" json:"access_token_secret"`
	IsActive          bool       `gorm:"default:true;comment:是否启用" json:"is_active"`
	FollowerCount     int64      `gorm:"default:0;comment:粉丝数" json:"follower_count"`
	LastPostAt        *time.Time `gorm:"comment:最后发布时间" json:"last_post_at"`
	CreatedAt         time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (XAccount) TableName() string {
	return "x_accounts"
}

// BeforeCreate 创建前钩子，确保 UUID 生成
func (x *XAccount) BeforeCreate() error {
	if x.ID == uuid.Nil {
		x.ID = uuid.New()
	}
	return nil
}

// XPostLog X.com 发布日志模型
type XPostLog struct {
	ID             uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AccountID      uuid.UUID  `gorm:"type:uuid;index;not null;comment:账号 UUID" json:"account_id"`
	VideoID        *uuid.UUID `gorm:"type:uuid;index;comment:视频 ID" json:"video_id"`
	TweetID        string     `gorm:"type:varchar(50);comment:推文 ID" json:"tweet_id"`
	TweetText      string     `gorm:"type:text;comment:推文内容" json:"tweet_text"`
	MediaURLs      JSONB      `gorm:"type:jsonb;comment:媒体 URL 列表" json:"media_urls"`
	Hashtags       JSONB      `gorm:"type:jsonb;comment:标签列表" json:"hashtags"`
	LinkURL        string     `gorm:"type:varchar(500);comment:链接 URL" json:"link_url"`
	DomainUsed     string     `gorm:"type:varchar(200);comment:使用的域名" json:"domain_used"`
	ImpressionCount int64     `gorm:"default:0;comment:曝光数" json:"impression_count"`
	ClickCount     int64      `gorm:"default:0;comment:点击数" json:"click_count"`
	RetweetCount   int64      `gorm:"default:0;comment:转发数" json:"retweet_count"`
	LikeCount      int64      `gorm:"default:0;comment:点赞数" json:"like_count"`
	Status         string     `gorm:"type:varchar(20);default:pending;comment:状态 pending/posted/failed" json:"status"`
	ErrorMessage   string     `gorm:"type:text;comment:错误信息" json:"error_message"`
	PostedAt       *time.Time `gorm:"comment:发布时间" json:"posted_at"`
	CreatedAt      time.Time  `gorm:"autoCreateTime" json:"created_at"`
}

// TableName 指定表名
func (XPostLog) TableName() string {
	return "x_post_logs"
}

// BeforeCreate 创建前钩子，确保 UUID 生成
func (x *XPostLog) BeforeCreate() error {
	if x.ID == uuid.Nil {
		x.ID = uuid.New()
	}
	return nil
}

// XPostQueue X.com 发布队列模型
type XPostQueue struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AccountID   uuid.UUID  `gorm:"type:uuid;index;not null;comment:账号 UUID" json:"account_id"`
	VideoID     *uuid.UUID `gorm:"type:uuid;index;comment:视频 ID" json:"video_id"`
	ScheduledAt time.Time  `gorm:"index;comment:计划发布时间" json:"scheduled_at"`
	Status      string     `gorm:"type:varchar(20);default:pending;comment:状态 pending/processing/posted/failed" json:"status"`
	RetryCount  int        `gorm:"default:0;comment:重试次数" json:"retry_count"`
	MaxRetries  int        `gorm:"default:3;comment:最大重试次数" json:"max_retries"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (XPostQueue) TableName() string {
	return "x_post_queues"
}

// BeforeCreate 创建前钩子，确保 UUID 生成
func (x *XPostQueue) BeforeCreate() error {
	if x.ID == uuid.Nil {
		x.ID = uuid.New()
	}
	return nil
}
