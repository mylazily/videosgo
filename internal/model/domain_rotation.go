package model

import (
	"time"

	"github.com/google/uuid"
)

// DomainAvailability 域名可用性记录模型
type DomainAvailability struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	DomainID       uuid.UUID `gorm:"type:uuid;index;not null;comment:域名 ID" json:"domain_id"`
	Region         string    `gorm:"type:varchar(50);index;comment:区域" json:"region"`
	IsAccessible   bool      `gorm:"default:true;comment:是否可访问" json:"is_accessible"`
	ResponseTimeMs int       `gorm:"default:0;comment:响应时间（毫秒）" json:"response_time_ms"`
	ErrorType      string    `gorm:"type:varchar(100);comment:错误类型" json:"error_type"`
	CheckedAt      time.Time `gorm:"autoCreateTime;comment:检查时间" json:"checked_at"`
}

// TableName 指定表名
func (DomainAvailability) TableName() string {
	return "domain_availabilities"
}

// BeforeCreate 创建前钩子，确保 UUID 生成
func (d *DomainAvailability) BeforeCreate() error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}

// DomainSwitchEvent 域名切换事件模型
type DomainSwitchEvent struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	FromDomain   string    `gorm:"type:varchar(200);not null;comment:原域名" json:"from_domain"`
	ToDomain     string    `gorm:"type:varchar(200);not null;comment:新域名" json:"to_domain"`
	SwitchReason string    `gorm:"type:varchar(200);comment:切换原因" json:"switch_reason"`
	AffectedUsers int64    `gorm:"default:0;comment:影响用户数" json:"affected_users"`
	SwitchedAt   time.Time `gorm:"autoCreateTime;comment:切换时间" json:"switched_at"`
}

// TableName 指定表名
func (DomainSwitchEvent) TableName() string {
	return "domain_switch_events"
}

// BeforeCreate 创建前钩子，确保 UUID 生成
func (d *DomainSwitchEvent) BeforeCreate() error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}

// ActiveDomain 活跃域名模型
type ActiveDomain struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Domain      string    `gorm:"type:varchar(200);uniqueIndex;not null;comment:域名" json:"domain"`
	Region      string    `gorm:"type:varchar(50);index;comment:区域" json:"region"`
	ActivatedAt time.Time `gorm:"autoCreateTime;comment:激活时间" json:"activated_at"`
	ActivatedBy string    `gorm:"type:varchar(100);comment:激活人" json:"activated_by"`
}

// TableName 指定表名
func (ActiveDomain) TableName() string {
	return "active_domains"
}

// BeforeCreate 创建前钩子，确保 UUID 生成
func (a *ActiveDomain) BeforeCreate() error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}
