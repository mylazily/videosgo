package model

import (
	"time"

	"github.com/google/uuid"
)

// AdTask 广告任务模型
type AdTask struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TaskName       string    `gorm:"type:varchar(100);not null;comment:任务名称" json:"task_name"`
	TaskType       string    `gorm:"type:varchar(50);not null;comment:任务类型 watch_ad/share/checkin/invite" json:"task_type"`
	RewardCoins    int64     `gorm:"default:0;comment:奖励金币" json:"reward_coins"`
	RewardType     string    `gorm:"type:varchar(50);default:coin;comment:奖励类型 coin/vip" json:"reward_type"`
	RewardValue    int64     `gorm:"default:0;comment:奖励数值" json:"reward_value"`
	MaxDaily       int       `gorm:"default:1;comment:每日最大完成次数" json:"max_daily"`
	DurationSeconds int      `gorm:"default:0;comment:观看时长（秒）" json:"duration_seconds"`
	IsActive       bool      `gorm:"default:true;comment:是否启用" json:"is_active"`
	SortOrder      int       `gorm:"default:0;comment:排序" json:"sort_order"`
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (AdTask) TableName() string {
	return "ad_tasks"
}

// BeforeCreate 创建前钩子，确保 UUID 生成
func (a *AdTask) BeforeCreate() error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

// CoinTransaction 金币交易记录模型
type CoinTransaction struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	FingerprintID  uuid.UUID `gorm:"type:uuid;index;not null;comment:设备指纹 ID" json:"fingerprint_id"`
	Amount         int64     `gorm:"not null;comment:金额（正为收入，负为支出）" json:"amount"`
	BalanceAfter   int64     `gorm:"not null;comment:交易后余额" json:"balance_after"`
	TransactionType string   `gorm:"type:varchar(50);not null;comment:交易类型 reward/unlock/checkin/refund" json:"transaction_type"`
	ReferenceID    string    `gorm:"type:varchar(100);comment:关联 ID" json:"reference_id"`
	Description    string    `gorm:"type:varchar(200);comment:描述" json:"description"`
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TableName 指定表名
func (CoinTransaction) TableName() string {
	return "coin_transactions"
}

// BeforeCreate 创建前钩子，确保 UUID 生成
func (c *CoinTransaction) BeforeCreate() error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// DailyTaskCompletion 每日任务完成记录模型
type DailyTaskCompletion struct {
	ID              uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	FingerprintID   uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_fp_task_date;not null;comment:设备指纹 ID" json:"fingerprint_id"`
	TaskID          uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_fp_task_date;not null;comment:任务 ID" json:"task_id"`
	CompletedAt     time.Time `gorm:"autoCreateTime;comment:完成时间" json:"completed_at"`
	CompletionCount int       `gorm:"default:1;comment:完成次数" json:"completion_count"`
	RewardGiven     int       `gorm:"default:0;comment:已发放奖励数量" json:"reward_given"`
}

// TableName 指定表名
func (DailyTaskCompletion) TableName() string {
	return "daily_task_completions"
}

// BeforeCreate 创建前钩子，确保 UUID 生成
func (d *DailyTaskCompletion) BeforeCreate() error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}
