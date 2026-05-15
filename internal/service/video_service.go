package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mylazily/videosgo/internal/database"
	"github.com/mylazily/videosgo/internal/model"
	"github.com/mylazily/videosgo/internal/repository"
)

// VideoService 视频服务
type VideoService struct {
	repo *repository.VideoRepo
}

// NewVideoService 创建视频服务
func NewVideoService(repo *repository.VideoRepo) *VideoService {
	return &VideoService{repo: repo}
}

// GetVideo 获取视频详情（带 Redis 缓存）
func (s *VideoService) GetVideo(id uint) (*model.Video, error) {
	cacheKey := fmt.Sprintf("video:detail:%d", id)

	// 尝试从缓存获取
	if database.RDB != nil {
		cached, err := database.RDB.Get(context.Background(), cacheKey).Result()
		if err == nil && cached != "" {
			var video model.Video
			if err := json.Unmarshal([]byte(cached), &video); err == nil {
				return &video, nil
			}
		}
	}

	// 从数据库获取
	video, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// 写入缓存（10 分钟）
	if database.RDB != nil {
		data, _ := json.Marshal(video)
		database.RDB.Set(context.Background(), cacheKey, data, 10*time.Minute)
	}

	return video, nil
}

// ListVideos 获取视频列表（带 Redis 缓存）
func (s *VideoService) ListVideos(page, pageSize int, category string) ([]model.Video, int64, error) {
	cacheKey := fmt.Sprintf("video:list:%s:%d:%d", category, page, pageSize)

	// 尝试从缓存获取
	if database.RDB != nil {
		cached, err := database.RDB.Get(context.Background(), cacheKey).Result()
		if err == nil && cached != "" {
			var result struct {
				List  []model.Video `json:"list"`
				Total int64         `json:"total"`
			}
			if err := json.Unmarshal([]byte(cached), &result); err == nil {
				return result.List, result.Total, nil
			}
		}
	}

	// 从数据库获取
	videos, total, err := s.repo.List(page, pageSize, category)
	if err != nil {
		return nil, 0, err
	}

	// 写入缓存（5 分钟）
	if database.RDB != nil {
		data, _ := json.Marshal(map[string]interface{}{
			"list":  videos,
			"total": total,
		})
		database.RDB.Set(context.Background(), cacheKey, data, 5*time.Minute)
	}

	return videos, total, nil
}

// SearchVideos 搜索视频
func (s *VideoService) SearchVideos(keyword string, page, pageSize int) ([]model.Video, int64, error) {
	// 记录搜索关键词到热搜
	_ = s.repo.UpsertSearchHot(keyword)

	return s.repo.Search(keyword, page, pageSize)
}

// GetCategories 获取分类列表（带 Redis 缓存）
func (s *VideoService) GetCategories() ([]string, error) {
	cacheKey := "video:categories"

	// 尝试从缓存获取
	if database.RDB != nil {
		cached, err := database.RDB.Get(context.Background(), cacheKey).Result()
		if err == nil && cached != "" {
			var categories []string
			if err := json.Unmarshal([]byte(cached), &categories); err == nil {
				return categories, nil
			}
		}
	}

	// 从数据库获取
	categories, err := s.repo.GetCategories()
	if err != nil {
		return nil, err
	}

	// 写入缓存（30 分钟）
	if database.RDB != nil {
		data, _ := json.Marshal(categories)
		database.RDB.Set(context.Background(), cacheKey, data, 30*time.Minute)
	}

	return categories, nil
}

// GetRandom 获取随机推荐
func (s *VideoService) GetRandom(limit int) ([]model.Video, error) {
	return s.repo.GetRandom(limit)
}

// GetLatest 获取最新视频
func (s *VideoService) GetLatest(limit int) ([]model.Video, error) {
	return s.repo.GetLatest(limit)
}

// GetHot 获取热门视频
func (s *VideoService) GetHot(limit int) ([]model.Video, error) {
	return s.repo.GetHot(limit)
}

// GetEpisodes 获取视频剧集
func (s *VideoService) GetEpisodes(videoID uint) ([]model.Episode, error) {
	return s.repo.GetEpisodesByVideoID(videoID)
}

// RecordWatch 记录观看历史和增加播放量
func (s *VideoService) RecordWatch(userID, videoID uint, progress, duration float64) error {
	// 增加播放量
	_ = s.repo.IncrementViewCount(videoID)

	// 保存观看历史
	if userID > 0 {
		history := &model.UserWatchHistory{
			UserID:   userID,
			VideoID:  videoID,
			Progress: progress,
			Duration: duration,
		}
		return s.repo.SaveWatchHistory(history)
	}
	return nil
}

// GetWatchHistory 获取观看历史
func (s *VideoService) GetWatchHistory(userID uint, page, pageSize int) ([]model.UserWatchHistory, int64, error) {
	return s.repo.GetWatchHistory(userID, page, pageSize)
}

// GetSearchHot 获取热搜列表（使用 Redis ZSET）
func (s *VideoService) GetSearchHot(limit int) ([]model.SearchHot, error) {
	// 优先从 Redis ZSET 获取
	if database.RDB != nil {
		zsetKey := "search:hot"
		result, err := database.RDB.ZRevRangeWithScores(context.Background(), zsetKey, 0, int64(limit-1)).Result()
		if err == nil && len(result) > 0 {
			hots := make([]model.SearchHot, 0, len(result))
			for _, z := range result {
				hots = append(hots, model.SearchHot{
					Keyword: z.Member.(string),
					Count:   int64(z.Score),
				})
			}
			return hots, nil
		}
	}

	// 降级到数据库
	return s.repo.GetSearchHot(limit)
}

// ClearVideoCache 清除视频缓存
func (s *VideoService) ClearVideoCache(videoID uint) {
	if database.RDB == nil {
		return
	}
	ctx := context.Background()
	// 清除详情缓存
	database.RDB.Del(ctx, fmt.Sprintf("video:detail:%d", videoID))
	// 清除列表缓存（使用模糊匹配）
	iter := database.RDB.Scan(ctx, 0, "video:list:*", 100).Iterator()
	for iter.Next(ctx) {
		database.RDB.Del(ctx, iter.Val())
	}
}
