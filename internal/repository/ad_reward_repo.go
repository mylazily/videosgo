package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/mylazily/videosgo/internal/model"
	"gorm.io/gorm"
)

// AdRewardRepo 广告金币数据仓库
type AdRewardRepo struct {
	db *gorm.DB
}

// NewAdRewardRepo 创建广告金币仓库
func NewAdRewardRepo(db *gorm.DB) *AdRewardRepo {
	return &AdRewardRepo{db: db}
}

// ========== AdTask ==========

// ListActiveTasks 获取所有活跃任务
func (r *AdRewardRepo) ListActiveTasks() ([]model.AdTask, error) {
	var tasks []model.AdTask
	err := r.db.Where("is_active = ?", true).Order("sort_order ASC").Find(&tasks).Error
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

// GetTaskByID 根据 ID 获取任务
func (r *AdRewardRepo) GetTaskByID(id uuid.UUID) (*model.AdTask, error) {
	var task model.AdTask
	err := r.db.First(&task, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// ========== CoinTransaction ==========

// CreateTransaction 创建金币交易记录
func (r *AdRewardRepo) CreateTransaction(fingerprintID uuid.UUID, amount int64, balanceAfter int64, txType string, refID string, desc string) error {
	tx := &model.CoinTransaction{
		FingerprintID:   fingerprintID,
		Amount:          amount,
		BalanceAfter:    balanceAfter,
		TransactionType: txType,
		ReferenceID:     refID,
		Description:     desc,
	}
	return r.db.Create(tx).Error
}

// GetBalance 查询余额（通过设备指纹关联）
func (r *AdRewardRepo) GetBalance(fingerprintID uuid.UUID) (int64, error) {
	var balance model.DeviceCoinBalance
	err := r.db.Where("fingerprint_id = ?", fingerprintID).First(&balance).Error
	if err != nil {
		return 0, err
	}
	return balance.Balance, nil
}

// GetTransactionHistory 获取交易历史
func (r *AdRewardRepo) GetTransactionHistory(fingerprintID uuid.UUID, limit int) ([]model.CoinTransaction, error) {
	var txs []model.CoinTransaction
	err := r.db.Where("fingerprint_id = ?", fingerprintID).
		Order("created_at DESC").
		Limit(limit).
		Find(&txs).Error
	if err != nil {
		return nil, err
	}
	return txs, nil
}

// ========== DailyTaskCompletion ==========

// RecordTaskCompletion 记录任务完成
func (r *AdRewardRepo) RecordTaskCompletion(fingerprintID, taskID uuid.UUID) error {
	// 检查今日是否已有记录
	today := time.Now().Truncate(24 * time.Hour)
	var existing model.DailyTaskCompletion
	err := r.db.Where("fingerprint_id = ? AND task_id = ? AND completed_at >= ?",
		fingerprintID, taskID, today).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		// 创建新记录
		completion := &model.DailyTaskCompletion{
			FingerprintID:   fingerprintID,
			TaskID:          taskID,
			CompletionCount: 1,
			RewardGiven:     true,
		}
		return r.db.Create(completion).Error
	}
	if err != nil {
		return err
	}
	// 更新已有记录
	return r.db.Model(&existing).Updates(map[string]interface{}{
		"completion_count": gorm.Expr("completion_count + 1"),
		"reward_given":     true,
	}).Error
}

// GetDailyCompletions 获取今日完成情况
func (r *AdRewardRepo) GetDailyCompletions(fingerprintID uuid.UUID) ([]model.DailyTaskCompletion, error) {
	today := time.Now().Truncate(24 * time.Hour)
	var completions []model.DailyTaskCompletion
	err := r.db.Where("fingerprint_id = ? AND completed_at >= ?", fingerprintID, today).
		Find(&completions).Error
	if err != nil {
		return nil, err
	}
	return completions, nil
}

// GetTaskCompletionCount 获取今日某任务完成次数
func (r *AdRewardRepo) GetTaskCompletionCount(fingerprintID, taskID uuid.UUID) (int, error) {
	today := time.Now().Truncate(24 * time.Hour)
	var count int64
	err := r.db.Model(&model.DailyTaskCompletion{}).
		Where("fingerprint_id = ? AND task_id = ? AND completed_at >= ?",
			fingerprintID, taskID, today).
		Count(&count).Error
	return int(count), err
}

// AddCoins 增加金币（直接操作 device_coin_balances 表）
func (r *AdRewardRepo) AddCoins(fingerprintID uuid.UUID, amount int64) error {
	return r.db.Model(&model.DeviceCoinBalance{}).
		Where("fingerprint_id = ?", fingerprintID).
		Updates(map[string]interface{}{
			"balance":      gorm.Expr("balance + ?", amount),
			"total_earned": gorm.Expr("total_earned + ?", amount),
		}).Error
}

// DeductCoins 扣除金币
func (r *AdRewardRepo) DeductCoins(fingerprintID uuid.UUID, amount int64) error {
	balance, err := r.GetBalance(fingerprintID)
	if err != nil {
		return err
	}
	if balance < amount {
		return gorm.ErrRecordNotFound
	}
	return r.db.Model(&model.DeviceCoinBalance{}).
		Where("fingerprint_id = ? AND balance >= ?", fingerprintID, amount).
		Updates(map[string]interface{}{
			"balance":     gorm.Expr("balance - ?", amount),
			"total_spent": gorm.Expr("total_spent + ?", amount),
		}).Error
}

// EnsureCoinBalance 确保金币余额记录存在
func (r *AdRewardRepo) EnsureCoinBalance(fingerprintID uuid.UUID) error {
	var count int64
	r.db.Model(&model.DeviceCoinBalance{}).
		Where("fingerprint_id = ?", fingerprintID).
		Count(&count)
	if count == 0 {
		return r.db.Create(&model.DeviceCoinBalance{
			FingerprintID: fingerprintID,
		}).Error
	}
	return nil
}
