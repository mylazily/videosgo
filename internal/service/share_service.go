package service

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
	"videosgo/internal/model"
	"videosgo/internal/repository"
)

// ShareService 分享裂变服务
type ShareService struct {
	repo         *repository.ShareRepo
	deviceRepo   *repository.DeviceRepo
}

// NewShareService 创建分享裂变服务
func NewShareService(
	repo *repository.ShareRepo,
	deviceRepo *repository.DeviceRepo,
) *ShareService {
	return &ShareService{
		repo:       repo,
		deviceRepo: deviceRepo,
	}
}

// GenerateShareLink 生成分享链接
func (s *ShareService) GenerateShareLink(fingerprintID, videoID uuid.UUID) (*model.ShareLink, error) {
	// 生成唯一 8 位分享码
	shareCode, err := s.generateUniqueCode()
	if err != nil {
		return nil, fmt.Errorf("生成分享码失败: %w", err)
	}

	link := &model.ShareLink{
		VideoID:              videoID,
		CreatorFingerprintID: fingerprintID,
		ShareCode:            shareCode,
		MaxUnlocks:           10,
		RewardType:           "coin",
		RewardAmount:         1,
		Status:               "active",
	}

	if err := s.repo.CreateShareLink(link); err != nil {
		return nil, fmt.Errorf("创建分享链接失败: %w", err)
	}

	return link, nil
}

// ProcessShareClick 处理分享点击
func (s *ShareService) ProcessShareClick(code string, fingerprintID uuid.UUID, ip, ua string) (*model.ShareLink, error) {
	// 获取分享链接
	link, err := s.repo.GetByCode(code)
	if err != nil {
		return nil, fmt.Errorf("分享链接不存在")
	}

	// 检查链接状态
	if link.Status != "active" {
		return nil, fmt.Errorf("分享链接已失效")
	}

	// 检查是否过期
	if link.ExpiresAt != nil && link.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("分享链接已过期")
	}

	// 检查是否已达到最大解锁次数
	if link.UnlockCount >= int64(link.MaxUnlocks) {
		return nil, fmt.Errorf("分享链接已达最大解锁次数")
	}

	// 记录点击
	click := &model.ShareClick{
		ShareLinkID:   link.ID,
		FingerprintID: fingerprintID,
		IPAddress:     ip,
		UserAgent:     ua,
	}
	if err := s.repo.RecordClick(click); err != nil {
		return nil, fmt.Errorf("记录点击失败: %w", err)
	}

	// 增加点击次数
	_ = s.repo.IncrementClick(link.ID)

	return link, nil
}

// CheckAndReward 检查并发放奖励
func (s *ShareService) CheckAndReward(shareLinkID, fingerprintID uuid.UUID) (bool, error) {
	// 检查是否已经点击过
	clicked, err := s.repo.HasClicked(shareLinkID, fingerprintID)
	if err != nil {
		return false, fmt.Errorf("检查点击记录失败: %w", err)
	}
	if !clicked {
		return false, fmt.Errorf("请先点击分享链接")
	}

	// 获取分享链接
	link, err := s.repo.GetByID(shareLinkID)
	if err != nil {
		return false, fmt.Errorf("分享链接不存在")
	}

	// 不能给自己的链接领奖
	if link.CreatorFingerprintID == fingerprintID {
		return false, fmt.Errorf("不能给自己的分享链接领奖")
	}

	// 增加解锁次数
	if err := s.repo.IncrementUnlock(shareLinkID); err != nil {
		return false, fmt.Errorf("更新解锁次数失败: %w", err)
	}

	// 给创建者发放奖励
	if link.RewardType == "coin" && link.RewardAmount > 0 {
		if err := s.deviceRepo.AddCoins(link.CreatorFingerprintID, link.RewardAmount); err != nil {
			return false, fmt.Errorf("发放奖励失败: %w", err)
		}
	}

	// 给点击者解锁视频
	if err := s.deviceRepo.AddCoins(fingerprintID, 1); err != nil {
		fmt.Printf("[Share] 给点击者发放奖励失败: %v\n", err)
	}

	return true, nil
}

// GetShareLink 获取分享链接信息
func (s *ShareService) GetShareLink(code string) (*model.ShareLink, error) {
	return s.repo.GetByCode(code)
}

// generateUniqueCode 生成唯一的 8 位分享码
func (s *ShareService) generateUniqueCode() (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const codeLength = 8

	for i := 0; i < 10; i++ {
		code := make([]byte, codeLength)
		for j := 0; j < codeLength; j++ {
			n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
			if err != nil {
				return "", err
			}
			code[j] = charset[n.Int64()]
		}

		codeStr := string(code)
		// 检查是否唯一
		_, err := s.repo.GetByCode(codeStr)
		if err != nil {
			return codeStr, nil
		}
	}

	return "", fmt.Errorf("生成唯一分享码失败")
}
