package repository

import (
	"time"

	"github.com/google/uuid"
	"videosgo/internal/model"
	"gorm.io/gorm"
)

// XRepo X.com 数据仓库
type XRepo struct {
	db *gorm.DB
}

// NewXRepo 创建 X.com 仓库
func NewXRepo(db *gorm.DB) *XRepo {
	return &XRepo{db: db}
}

// ========== XAccount ==========

// CreateAccount 创建账号
func (r *XRepo) CreateAccount(account *model.XAccount) error {
	return r.db.Create(account).Error
}

// GetAccountByID 根据 UUID 获取账号
func (r *XRepo) GetAccountByID(id uuid.UUID) (*model.XAccount, error) {
	var account model.XAccount
	err := r.db.First(&account, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

// ListAccounts 获取所有账号
func (r *XRepo) ListAccounts() ([]model.XAccount, error) {
	var accounts []model.XAccount
	err := r.db.Find(&accounts).Error
	if err != nil {
		return nil, err
	}
	return accounts, nil
}

// ListActiveAccounts 获取所有活跃账号
func (r *XRepo) ListActiveAccounts() ([]model.XAccount, error) {
	var accounts []model.XAccount
	err := r.db.Where("is_active = ?", true).Find(&accounts).Error
	if err != nil {
		return nil, err
	}
	return accounts, nil
}

// UpdateAccount 更新账号
func (r *XRepo) UpdateAccount(account *model.XAccount) error {
	return r.db.Save(account).Error
}

// ========== XPostLog ==========

// CreatePostLog 创建发布日志
func (r *XRepo) CreatePostLog(log *model.XPostLog) error {
	return r.db.Create(log).Error
}

// ListPostLogs 获取发布日志列表
func (r *XRepo) ListPostLogs(limit int) ([]model.XPostLog, error) {
	var logs []model.XPostLog
	err := r.db.Order("created_at DESC").Limit(limit).Find(&logs).Error
	if err != nil {
		return nil, err
	}
	return logs, nil
}

// MarkAsPosted 标记为已发布
func (r *XRepo) MarkAsPosted(id uuid.UUID, tweetID string) error {
	now := time.Now()
	return r.db.Model(&model.XPostLog{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":    "posted",
		"tweet_id":  tweetID,
		"posted_at": &now,
	}).Error
}

// MarkAsFailed 标记为发布失败
func (r *XRepo) MarkAsFailed(id uuid.UUID, errMsg string) error {
	return r.db.Model(&model.XPostLog{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":        "failed",
		"error_message": errMsg,
	}).Error
}

// ========== XPostQueue ==========

// CreatePostQueue 创建发布队列项
func (r *XRepo) CreatePostQueue(queue *model.XPostQueue) error {
	return r.db.Create(queue).Error
}

// GetPendingPosts 获取待发布队列
func (r *XRepo) GetPendingPosts() ([]model.XPostQueue, error) {
	var queues []model.XPostQueue
	err := r.db.Where("status = ? AND scheduled_at <= ?", "pending", time.Now()).
		Order("scheduled_at ASC").
		Find(&queues).Error
	if err != nil {
		return nil, err
	}
	return queues, nil
}

// MarkQueueAsProcessing 标记队列为处理中
func (r *XRepo) MarkQueueAsProcessing(id uuid.UUID) error {
	return r.db.Model(&model.XPostQueue{}).Where("id = ?", id).Update("status", "processing").Error
}

// MarkQueueAsPosted 标记队列为已发布
func (r *XRepo) MarkQueueAsPosted(id uuid.UUID) error {
	return r.db.Model(&model.XPostQueue{}).Where("id = ?", id).Update("status", "posted").Error
}

// MarkQueueAsFailed 标记队列为失败
func (r *XRepo) MarkQueueAsFailed(id uuid.UUID) error {
	// 先增加重试计数
	err := r.db.Model(&model.XPostQueue{}).
		Where("id = ?", id).
		Update("retry_count", gorm.Expr("retry_count + 1")).Error
	if err != nil {
		return err
	}

	// 如果重试次数超过最大值，标记为失败
	return r.db.Model(&model.XPostQueue{}).
		Where("id = ? AND retry_count >= max_retries", id).
		Update("status", "failed").Error
}
