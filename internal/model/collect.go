package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CollectSource 采集源配置
type CollectSource struct {
	ID             uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name           string         `gorm:"type:varchar(100);not null;comment:采集源名称" json:"name"`
	Category       string         `gorm:"type:varchar(100);comment:采集分类" json:"category"`
	APIURL         string         `gorm:"type:varchar(2048);not null;comment:MacCMS API 地址" json:"api_url"`
	APIKey         string         `gorm:"type:varchar(200);comment:API 密钥" json:"api_key"`
	APIParam       string         `gorm:"type:text;comment:API 参数" json:"api_param"`
	Headers        JSONB          `gorm:"type:jsonb;comment:请求头" json:"headers"`
	Interval       int            `gorm:"default:60;comment:采集间隔（分钟）" json:"interval"`
	MaxPages       int            `gorm:"default:10;comment:最大采集页数" json:"max_pages"`
	Timeout        int            `gorm:"default:30;comment:超时时间（秒）" json:"timeout"`
	RetryCount     int            `gorm:"default:3;comment:重试次数" json:"retry_count"`
	Enabled        bool           `gorm:"default:true;comment:是否启用" json:"enabled"`
	LastCollect    *time.Time     `gorm:"comment:上次采集时间" json:"last_collect"`
	LastSync       *time.Time     `gorm:"comment:上次同步时间" json:"last_sync"`
	LastError      string         `gorm:"type:text;comment:上次错误信息" json:"last_error"`
	TotalCollected int            `gorm:"default:0;comment:总采集数" json:"total_collected"`
	TotalNew       int            `gorm:"default:0;comment:新增数" json:"total_new"`
	Status         string         `gorm:"type:varchar(20);default:active;comment:状态" json:"status"`
	CreatedAt      time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index;comment:删除时间" json:"-"`
}

// TableName 指定表名
func (CollectSource) TableName() string {
	return "collect_sources"
}

// BeforeCreate 创建前钩子，确保 UUID 生成
func (c *CollectSource) BeforeCreate() error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// CollectLog 采集日志
type CollectLog struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SourceID     uuid.UUID `gorm:"type:uuid;index;not null;comment:采集源 ID" json:"source_id"`
	SourceName   string    `gorm:"type:varchar(100);comment:采集源名称" json:"source_name"`
	Type         string    `gorm:"type:varchar(20);comment:采集类型 full/incremental" json:"type"`
	TotalCount   int       `gorm:"default:0;comment:总采集数" json:"total_count"`
	NewCount     int       `gorm:"default:0;comment:新增数" json:"new_count"`
	UpdateCount  int       `gorm:"default:0;comment:更新数" json:"update_count"`
	FailCount    int       `gorm:"default:0;comment:失败数" json:"fail_count"`
	Duration     int       `gorm:"default:0;comment:耗时（秒）" json:"duration"`
	Status       string    `gorm:"type:varchar(20);default:running;comment:状态 running/success/failed" json:"status"`
	ErrorMessage string    `gorm:"type:text;comment:错误信息" json:"error_message"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TableName 指定表名
func (CollectLog) TableName() string {
	return "collect_logs"
}

// BeforeCreate 创建前钩子，确保 UUID 生成
func (c *CollectLog) BeforeCreate() error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}
