package repository

import (
	"github.com/mylazily/videosgo/internal/model"
	"gorm.io/gorm"
)

// RankRepo 排行榜数据仓库
type RankRepo struct {
	db *gorm.DB
}

// NewRankRepo 创建排行榜仓库
func NewRankRepo(db *gorm.DB) *RankRepo {
	return &RankRepo{db: db}
}

// GetByType 根据类型获取排行榜
func (r *RankRepo) GetByType(rankType string, limit int) ([]model.Rank, error) {
	var ranks []model.Rank
	err := r.db.Preload("Video").
		Where("type = ?", rankType).
		Order("score DESC").
		Limit(limit).
		Find(&ranks).Error
	return ranks, err
}

// Upsert 创建或更新排行榜记录
func (r *RankRepo) Upsert(rank *model.Rank) error {
	var existing model.Rank
	err := r.db.Where("video_id = ? AND type = ?", rank.VideoID, rank.Type).First(&existing).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return r.db.Create(rank).Error
		}
		return err
	}
	return r.db.Model(&existing).Update("score", gorm.Expr("score + ?", rank.Score)).Error
}

// IncrementScore 增加视频热度分数
func (r *RankRepo) IncrementScore(videoID uint, score int) error {
	return r.db.Model(&model.Rank{}).
		Where("video_id = ?", videoID).
		UpdateColumn("score", gorm.Expr("score + ?", score)).Error
}

// DeleteByType 删除指定类型的排行榜
func (r *RankRepo) DeleteByType(rankType string) error {
	return r.db.Where("type = ?", rankType).Delete(&model.Rank{}).Error
}

// GetTopByCategory 获取分类排行榜
func (r *RankRepo) GetTopByCategory(category, rankType string, limit int) ([]model.Rank, error) {
	var ranks []model.Rank
	err := r.db.Preload("Video").
		Joins("JOIN videos ON videos.id = ranks.video_id").
		Where("ranks.type = ? AND videos.category = ?", rankType, category).
		Order("ranks.score DESC").
		Limit(limit).
		Find(&ranks).Error
	return ranks, err
}
