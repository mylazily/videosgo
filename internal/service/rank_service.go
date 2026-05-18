package service

import (
	"fmt"

	"github.com/google/uuid"
	"videosgo/internal/model"
	"videosgo/internal/repository"
)

// RankService 排行榜服务
type RankService struct {
	repo *repository.RankRepo
}

// NewRankService 创建排行榜服务
func NewRankService(repo *repository.RankRepo) *RankService {
	return &RankService{repo: repo}
}

// GetDailyRank 获取日排行榜
func (s *RankService) GetDailyRank(limit int) ([]model.Rank, error) {
	return s.repo.GetByType("daily", limit)
}

// GetWeeklyRank 获取周排行榜
func (s *RankService) GetWeeklyRank(limit int) ([]model.Rank, error) {
	return s.repo.GetByType("weekly", limit)
}

// GetMonthlyRank 获取月排行榜
func (s *RankService) GetMonthlyRank(limit int) ([]model.Rank, error) {
	return s.repo.GetByType("monthly", limit)
}

// GetCategoryRank 获取分类排行榜
func (s *RankService) GetCategoryRank(category, rankType string, limit int) ([]model.Rank, error) {
	return s.repo.GetTopByCategory(category, rankType, limit)
}

// IncrementScore 增加视频热度
func (s *RankService) IncrementScore(videoID string, score int) error {
	parsedID, err := uuid.Parse(videoID)
	if err != nil {
		return fmt.Errorf("invalid UUID: %s", videoID)
	}
	return s.repo.IncrementScore(parsedID, score)
}
