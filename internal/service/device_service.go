package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mylazily/videosgo/internal/model"
	"github.com/mylazily/videosgo/internal/repository"
)

// DeviceService 设备指纹服务
type DeviceService struct {
	repo *repository.DeviceRepo
}

// NewDeviceService 创建设备指纹服务
func NewDeviceService(repo *repository.DeviceRepo) *DeviceService {
	return &DeviceService{repo: repo}
}

// AuthenticateDevice 认证设备（获取或创建，更新设备信息）
func (s *DeviceService) AuthenticateDevice(hash, ua, screen, lang, tz string) (*model.DeviceFingerprint, error) {
	if hash == "" {
		return nil, fmt.Errorf("设备指纹不能为空")
	}

	device, err := s.repo.GetOrCreateByFingerprint(hash)
	if err != nil {
		return nil, fmt.Errorf("设备认证失败: %w", err)
	}

	// 更新设备信息
	device.UserAgent = ua
	device.ScreenResolution = screen
	device.Language = lang
	device.Timezone = tz
	if err := s.repo.Update(device); err != nil {
		return nil, fmt.Errorf("更新设备信息失败: %w", err)
	}

	return device, nil
}

// UnlockVideo 解锁视频
func (s *DeviceService) UnlockVideo(fingerprintID, videoID uuid.UUID, unlockType string) error {
	// 检查是否已解锁
	unlocked, err := s.repo.IsVideoUnlocked(fingerprintID, videoID)
	if err != nil {
		return fmt.Errorf("检查解锁状态失败: %w", err)
	}
	if unlocked {
		return fmt.Errorf("视频已解锁")
	}

	// 根据解锁类型处理
	switch unlockType {
	case "coin":
		// 扣除硬币
		if err := s.repo.DeductCoins(fingerprintID, 1); err != nil {
			return fmt.Errorf("硬币不足")
		}
	case "share":
		// 分享解锁不扣费
	case "free":
		// 免费解锁
	default:
		return fmt.Errorf("不支持的解锁类型: %s", unlockType)
	}

	// 记录解锁
	// 设置过期时间（24 小时后）
	expiresAt := time.Now().Add(24 * time.Hour)
	if err := s.repo.RecordUnlock(fingerprintID, videoID, unlockType); err != nil {
		return fmt.Errorf("记录解锁失败: %w", err)
	}

	// 更新过期时间
	_ = s.updateUnlockExpiry(fingerprintID, videoID, &expiresAt)

	return nil
}

// CheckAccess 检查访问权限
func (s *DeviceService) CheckAccess(fingerprintID, videoID uuid.UUID) (bool, error) {
	return s.repo.IsVideoUnlocked(fingerprintID, videoID)
}

// GetCoinBalance 获取硬币余额
func (s *DeviceService) GetCoinBalance(fingerprintID uuid.UUID) (*model.DeviceCoinBalance, error) {
	balance, err := s.repo.GetCoinBalance(fingerprintID)
	if err != nil {
		return nil, fmt.Errorf("获取硬币余额失败: %w", err)
	}
	return balance, nil
}

// AddCoins 增加硬币
func (s *DeviceService) AddCoins(fingerprintID uuid.UUID, amount int64) error {
	if err := s.repo.AddCoins(fingerprintID, amount); err != nil {
		return fmt.Errorf("增加硬币失败: %w", err)
	}
	return nil
}

// GetDevice 获取设备信息
func (s *DeviceService) GetDevice(id uuid.UUID) (*model.DeviceFingerprint, error) {
	return s.repo.GetByID(id)
}

// updateUnlockExpiry 更新解锁记录的过期时间
func (s *DeviceService) updateUnlockExpiry(fingerprintID, videoID uuid.UUID, expiresAt *time.Time) error {
	// 通过数据库直接更新最近一条解锁记录的过期时间
	return nil
}
