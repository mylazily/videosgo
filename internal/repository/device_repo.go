package repository

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/mylazily/videosgo/internal/model"
	"gorm.io/gorm"
)

// DeviceRepo 设备指纹数据仓库
type DeviceRepo struct {
	db *gorm.DB
}

// NewDeviceRepo 创建设备指纹仓库
func NewDeviceRepo(db *gorm.DB) *DeviceRepo {
	return &DeviceRepo{db: db}
}

// GetOrCreateByFingerprint 获取或创建设备
func (r *DeviceRepo) GetOrCreateByFingerprint(hash string) (*model.DeviceFingerprint, error) {
	var device model.DeviceFingerprint
	err := r.db.Where("fingerprint_hash = ?", hash).First(&device).Error
	if err == gorm.ErrRecordNotFound {
		device = model.DeviceFingerprint{
			FingerprintHash: hash,
		}
		if err := r.db.Create(&device).Error; err != nil {
			return nil, err
		}
		// 创建硬币余额记录
		if err := r.db.Create(&model.DeviceCoinBalance{
			FingerprintID: device.ID,
		}).Error; err != nil {
			// 硬币余额创建失败，记录日志但不影响设备创建成功
			fmt.Printf("[DeviceRepo] 创建设备硬币余额失败: %v\n", err)
		}
		return &device, nil
	}
	if err != nil {
		return nil, err
	}
	return &device, nil
}

// GetByID 根据 ID 获取设备
func (r *DeviceRepo) GetByID(id uuid.UUID) (*model.DeviceFingerprint, error) {
	var device model.DeviceFingerprint
	err := r.db.First(&device, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &device, nil
}

// Update 更新设备信息
func (r *DeviceRepo) Update(device *model.DeviceFingerprint) error {
	return r.db.Save(device).Error
}

// RecordUnlock 记录解锁
func (r *DeviceRepo) RecordUnlock(fingerprintID, videoID uuid.UUID, unlockType string) error {
	record := &model.DeviceUnlockRecord{
		FingerprintID: fingerprintID,
		VideoID:       videoID,
		UnlockType:    unlockType,
	}
	return r.db.Create(record).Error
}

// IsVideoUnlocked 检查视频是否已解锁
func (r *DeviceRepo) IsVideoUnlocked(fingerprintID, videoID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&model.DeviceUnlockRecord{}).
		Where("fingerprint_id = ? AND video_id = ?", fingerprintID, videoID).
		Count(&count).Error
	return count > 0, err
}

// GetCoinBalance 获取硬币余额
func (r *DeviceRepo) GetCoinBalance(fingerprintID uuid.UUID) (*model.DeviceCoinBalance, error) {
	var balance model.DeviceCoinBalance
	err := r.db.Where("fingerprint_id = ?", fingerprintID).First(&balance).Error
	if err != nil {
		return nil, err
	}
	return &balance, nil
}

// AddCoins 增加硬币
func (r *DeviceRepo) AddCoins(fingerprintID uuid.UUID, amount int64) error {
	return r.db.Model(&model.DeviceCoinBalance{}).
		Where("fingerprint_id = ?", fingerprintID).
		Updates(map[string]interface{}{
			"balance":      gorm.Expr("balance + ?", amount),
			"total_earned": gorm.Expr("total_earned + ?", amount),
		}).Error
}

// DeductCoins 扣除硬币
func (r *DeviceRepo) DeductCoins(fingerprintID uuid.UUID, amount int64) error {
	// 先检查余额是否足够
	balance, err := r.GetCoinBalance(fingerprintID)
	if err != nil {
		return err
	}
	if balance.Balance < amount {
		return gorm.ErrRecordNotFound // 使用此错误表示余额不足
	}

	return r.db.Model(&model.DeviceCoinBalance{}).
		Where("fingerprint_id = ? AND balance >= ?", fingerprintID, amount).
		Updates(map[string]interface{}{
			"balance":     gorm.Expr("balance - ?", amount),
			"total_spent": gorm.Expr("total_spent + ?", amount),
		}).Error
}

// EnsureCoinBalance 确保硬币余额记录存在
func (r *DeviceRepo) EnsureCoinBalance(fingerprintID uuid.UUID) error {
	var count int64
	if err := r.db.Model(&model.DeviceCoinBalance{}).
		Where("fingerprint_id = ?", fingerprintID).
		Count(&count).Error; err != nil {
		return fmt.Errorf("查询硬币余额失败: %w", err)
	}
	if count == 0 {
		return r.db.Create(&model.DeviceCoinBalance{
			FingerprintID: fingerprintID,
		}).Error
	}
	return nil
}

// UpdateUnlockExpiry 更新解锁记录的过期时间
func (r *DeviceRepo) UpdateUnlockExpiry(fingerprintID uuid.UUID, expiresAt interface{}) error {
	return r.db.Model(&model.DeviceCoinBalance{}).
		Where("fingerprint_id = ?", fingerprintID).
		Update("unlock_expiry", expiresAt).Error
}
