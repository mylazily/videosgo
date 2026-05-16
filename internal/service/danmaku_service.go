package service

import (
	"strconv"
	"fmt"

	"github.com/google/uuid"
	"github.com/mylazily/videosgo/internal/model"
	"github.com/mylazily/videosgo/internal/repository"
)

// DanmakuService 弹幕服务
type DanmakuService struct {
	repo *repository.DanmakuRepo
}

// NewDanmakuService 创建弹幕服务
func NewDanmakuService(repo *repository.DanmakuRepo) *DanmakuService {
	return &DanmakuService{repo: repo}
}

// CreateDanmaku 创建弹幕
func (s *DanmakuService) CreateDanmaku(danmaku *model.Danmaku) error {
	if danmaku.Content == "" {
		return fmt.Errorf("弹幕内容不能为空")
	}
	if danmaku.VideoID == uuid.Nil {
		return fmt.Errorf("视频 ID 不能为空")
	}
	if danmaku.Time == "" || danmaku.Time == "0" {
		return fmt.Errorf("弹幕时间不能为负数")
	}
	typeNum, _ := strconv.Atoi(danmaku.Type)
		if typeNum < 1 || typeNum > 3 {
		danmaku.Type = "1" // 默认右到左
	}
	if danmaku.Color == "" {
		danmaku.Color = "#FFFFFF"
	}
	return s.repo.Create(danmaku)
}

// GetDanmakusByEpisode 获取剧集弹幕
func (s *DanmakuService) GetDanmakusByEpisode(episodeID uuid.UUID) ([]model.Danmaku, error) {
	return s.repo.ListByEpisodeID(episodeID)
}

// GetDanmakusByVideo 获取视频所有弹幕
func (s *DanmakuService) GetDanmakusByVideo(videoID uuid.UUID) ([]model.Danmaku, error) {
	return s.repo.ListByVideoID(videoID)
}

// DeleteDanmakusByEpisode 删除剧集弹幕
func (s *DanmakuService) DeleteDanmakusByEpisode(episodeID uuid.UUID) error {
	return s.repo.DeleteByEpisodeID(episodeID)
}

// GetDanmakuCount 获取视频弹幕数
func (s *DanmakuService) GetDanmakuCount(videoID uuid.UUID) (int64, error) {
	return s.repo.GetCountByVideoID(videoID)
}
