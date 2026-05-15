package model

import (
	"time"

	"github.com/google/uuid"
)

// SiteDomain 站群域名模型
type SiteDomain struct {
	ID               uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Domain           string     `gorm:"type:varchar(200);uniqueIndex;not null;comment:域名" json:"domain"`
	Cluster          string     `gorm:"type:varchar(1);not null;default:'A';index;comment:集群(A/B)" json:"cluster"`
	Role             string     `gorm:"type:varchar(50);default:primary;comment:角色(primary/secondary/cdn)" json:"role"`
	CloudflareZoneID string     `gorm:"type:varchar(100);comment:Cloudflare Zone ID" json:"cloudflare_zone_id"`
	SSLStatus        string     `gorm:"type:varchar(20);default:unknown;comment:SSL状态(valid/invalid/expired/unknown)" json:"ssl_status"`
	IsActive         bool       `gorm:"default:true;index;comment:是否启用" json:"is_active"`
	HealthStatus     string     `gorm:"type:varchar(20);default:unknown;comment:健康状态(healthy/unhealthy/unknown)" json:"health_status"`
	LastHealthCheck  *time.Time `gorm:"comment:上次健康检查时间" json:"last_health_check"`
	GoogleRank       int        `gorm:"default:0;comment:Google PR值" json:"google_rank"`
	DailyTraffic     int64      `gorm:"default:0;comment:日流量" json:"daily_traffic"`
	MonthlyTraffic   int64      `gorm:"default:0;comment:月流量" json:"monthly_traffic"`
	RedirectTarget   string     `gorm:"type:varchar(500);comment:重定向目标" json:"redirect_target"`
	RedirectEnabled  bool       `gorm:"default:false;comment:是否启用重定向" json:"redirect_enabled"`
	SEOScore         int        `gorm:"default:0;comment:SEO评分(0-100)" json:"seo_score"`
	Notes            string     `gorm:"type:text;comment:备注" json:"notes"`
	CreatedAt        time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (SiteDomain) TableName() string {
	return "site_domains"
}

// BeforeCreate 创建前钩子，确保 UUID 生成
func (s *SiteDomain) BeforeCreate() error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// SiteHealthLog 站点健康检查日志
type SiteHealthLog struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	DomainID       uuid.UUID `gorm:"type:uuid;index;not null;comment:域名 ID" json:"domain_id"`
	StatusCode     int       `gorm:"default:0;comment:HTTP 状态码" json:"status_code"`
	ResponseTimeMs int       `gorm:"default:0;comment:响应时间(毫秒)" json:"response_time_ms"`
	IsSSLValid     bool      `gorm:"default:false;comment:SSL 是否有效" json:"is_ssl_valid"`
	ErrorMessage   string    `gorm:"type:text;comment:错误信息" json:"error_message"`
	CheckedAt      time.Time `gorm:"autoCreateTime;index;comment:检查时间" json:"checked_at"`
}

// TableName 指定表名
func (SiteHealthLog) TableName() string {
	return "site_health_logs"
}

// BeforeCreate 创建前钩子，确保 UUID 生成
func (s *SiteHealthLog) BeforeCreate() error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// DomainLinkAudit 域名交叉链接审计记录
type DomainLinkAudit struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SourceDomain string    `gorm:"type:varchar(200);index;not null;comment:来源域名" json:"source_domain"`
	TargetDomain string    `gorm:"type:varchar(200);index;not null;comment:目标域名" json:"target_domain"`
	LinkType     string    `gorm:"type:varchar(50);comment:链接类型(internal/external/cross_cluster)" json:"link_type"`
	DetectedAt   time.Time `gorm:"autoCreateTime;index;comment:检测时间" json:"detected_at"`
	IsResolved   bool      `gorm:"default:false;index;comment:是否已处理" json:"is_resolved"`
}

// TableName 指定表名
func (DomainLinkAudit) TableName() string {
	return "domain_link_audits"
}

// BeforeCreate 创建前钩子，确保 UUID 生成
func (d *DomainLinkAudit) BeforeCreate() error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}
