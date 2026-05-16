package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// JSONB 通用 JSONB 类型
type JSONB map[string]interface{}

// Scan 实现 sql.Scanner 接口
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	switch v := value.(type) {
	case string:
		return json.Unmarshal([]byte(v), j)
	case []byte:
		return json.Unmarshal(v, j)
	}
	return fmt.Errorf("无法扫描 JSONB: %v", value)
}

// Value 实现 driver.Valuer 接口
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return "{}", nil
	}
	data, err := json.Marshal(j)
	if err != nil {
		return nil, err
	}
	return string(data), nil
}

// MarshalJSON 实现 json.Marshaler 接口
func (j JSONB) MarshalJSON() ([]byte, error) {
	if j == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(map[string]interface{}(j))
}

// UnmarshalJSON 实现 json.Unmarshaler 接口
func (j *JSONB) UnmarshalJSON(data []byte) error {
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	*j = m
	return nil
}

// TGBotConfig TG Bot 配置模型
type TGBotConfig struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	BotToken    string    `gorm:"type:text;not null;comment:Bot Token" json:"bot_token"`
	BotUsername string    `gorm:"type:varchar(100);comment:Bot 用户名" json:"bot_username"`
	WebhookURL  string    `gorm:"type:varchar(500);comment:Webhook URL" json:"webhook_url"`
	IsActive    bool      `gorm:"default:true;comment:是否启用" json:"is_active"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (TGBotConfig) TableName() string {
	return "tg_bot_configs"
}

// BeforeCreate 创建前钩子，确保 UUID 生成
func (t *TGBotConfig) BeforeCreate() error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

// TGChannel TG 频道模型
type TGChannel struct {
	ID              uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ChannelID       int64      `gorm:"not null;uniqueIndex;comment:频道 ID" json:"channel_id"`
	ChannelTitle    string     `gorm:"type:varchar(200);comment:频道标题" json:"channel_title"`
	ChannelType     string     `gorm:"type:varchar(20);default:channel;comment:频道类型 channel/group" json:"channel_type"`
	SubscriberCount int64      `gorm:"default:0;comment:订阅者数量" json:"subscriber_count"`
	IsActive        bool       `gorm:"default:true;comment:是否启用" json:"is_active"`
	LastPostAt      *time.Time `gorm:"comment:最后发布时间" json:"last_post_at"`
	CreatedAt       time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (TGChannel) TableName() string {
	return "tg_channels"
}

// BeforeCreate 创建前钩子，确保 UUID 生成
func (t *TGChannel) BeforeCreate() error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

// TGBroadcastLog TG 广播日志模型
type TGBroadcastLog struct {
	ID           uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ChannelID    uuid.UUID  `gorm:"type:uuid;index;comment:频道 UUID" json:"channel_id"`
	VideoID      *uuid.UUID `gorm:"type:uuid;index;comment:视频 ID" json:"video_id"`
	MessageID    int64      `gorm:"comment:消息 ID" json:"message_id"`
	MessageText  string     `gorm:"type:text;comment:消息文本" json:"message_text"`
	MediaURL     string     `gorm:"type:varchar(500);comment:媒体 URL" json:"media_url"`
	LinkURL      string     `gorm:"type:varchar(500);comment:链接 URL" json:"link_url"`
	PostType     string     `gorm:"type:varchar(20);default:text;comment:发布类型 text/photo/video" json:"post_type"`
	Status       string     `gorm:"type:varchar(20);default:pending;comment:状态 pending/sent/failed" json:"status"`
	ErrorMessage string     `gorm:"type:text;comment:错误信息" json:"error_message"`
	SentAt       *time.Time `gorm:"comment:发送时间" json:"sent_at"`
	CreatedAt    time.Time  `gorm:"autoCreateTime" json:"created_at"`
}

// TableName 指定表名
func (TGBroadcastLog) TableName() string {
	return "tg_broadcast_logs"
}

// BeforeCreate 创建前钩子，确保 UUID 生成
func (t *TGBroadcastLog) BeforeCreate() error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

// TGMiniAppSession TG Mini App 会话模型
type TGMiniAppSession struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TGUserID      int64     `gorm:"uniqueIndex;not null;comment:Telegram 用户 ID" json:"tg_user_id"`
	TGUsername    string    `gorm:"type:varchar(100);comment:Telegram 用户名" json:"tg_username"`
	TGLanguage    string    `gorm:"type:varchar(10);comment:语言" json:"tg_language"`
	FingerprintID *uuid.UUID `gorm:"type:uuid;index;comment:设备指纹 ID" json:"fingerprint_id"`
	SessionData   JSONB     `gorm:"type:jsonb;comment:会话数据" json:"session_data"`
	FirstOpenAt   time.Time `gorm:"autoCreateTime;comment:首次打开时间" json:"first_open_at"`
	LastOpenAt    time.Time `gorm:"autoUpdateTime;comment:最近打开时间" json:"last_open_at"`
	TotalOpens    int64     `gorm:"default:1;comment:总打开次数" json:"total_opens"`
	TotalWatchTime int64    `gorm:"default:0;comment:总观看时长（秒）" json:"total_watch_time"`
}

// TableName 指定表名
func (TGMiniAppSession) TableName() string {
	return "tg_mini_app_sessions"
}

// BeforeCreate 创建前钩子，确保 UUID 生成
func (t *TGMiniAppSession) BeforeCreate() error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}
