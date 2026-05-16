package model

import (
	"time"

	"github.com/google/uuid"
)

// RedirectRule 301 重定向规则
type RedirectRule struct {
	ID           uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SourceDomain string     `gorm:"type:varchar(200);index;not null;comment:来源域名" json:"source_domain"`
	SourcePath   string     `gorm:"type:varchar(500);index;comment:来源路径(支持通配符)" json:"source_path"`
	TargetURL    string     `gorm:"type:varchar(500);not null;comment:目标 URL" json:"target_url"`
	RuleType     string     `gorm:"type:varchar(20);default:exact;comment:规则类型(exact/prefix/regex/wildcard)" json:"rule_type"`
	Priority     int        `gorm:"default:0;index;comment:优先级(越大越优先)" json:"priority"`
	IsActive     bool       `gorm:"default:true;index;comment:是否启用" json:"is_active"`
	HitCount     int64      `gorm:"default:0;comment:命中次数" json:"hit_count"`
	Conditions   JSONB      `gorm:"type:jsonb;comment:附加条件(JSONB)" json:"conditions"`
	ExpiresAt    *time.Time `gorm:"comment:过期时间" json:"expires_at"`
	Notes        string     `gorm:"type:text;comment:备注" json:"notes"`
	CreatedAt    time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (RedirectRule) TableName() string {
	return "redirect_rules"
}

// BeforeCreate 创建前钩子，确保 UUID 生成
func (r *RedirectRule) BeforeCreate() error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

// RedirectHitLog 重定向命中日志
type RedirectHitLog struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	RuleID       uuid.UUID `gorm:"type:uuid;index;not null;comment:规则 ID" json:"rule_id"`
	SourceDomain string    `gorm:"type:varchar(200);index;comment:来源域名" json:"source_domain"`
	SourcePath   string    `gorm:"type:varchar(500);comment:来源路径" json:"source_path"`
	TargetURL    string    `gorm:"type:varchar(500);comment:目标 URL" json:"target_url"`
	IPAddress    string    `gorm:"type:varchar(50);index;comment:IP 地址" json:"ip_address"`
	UserAgent    string    `gorm:"type:varchar(500);comment:用户代理" json:"user_agent"`
	HitAt        time.Time `gorm:"autoCreateTime;index;comment:命中时间" json:"hit_at"`
}

// TableName 指定表名
func (RedirectHitLog) TableName() string {
	return "redirect_hit_logs"
}

// BeforeCreate 创建前钩子，确保 UUID 生成
func (r *RedirectHitLog) BeforeCreate() error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}
