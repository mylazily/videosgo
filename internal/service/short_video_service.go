package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mylazily/videosgo/internal/database"
	"github.com/mylazily/videosgo/internal/model"
	"github.com/mylazily/videosgo/internal/repository"
)

// ShortVideoService 短视频服务
type ShortVideoService struct {
	repo *repository.ShortVideoRepo
}

// NewShortVideoService 创建短视频服务
func NewShortVideoService(repo *repository.ShortVideoRepo) *ShortVideoService {
	return &ShortVideoService{repo: repo}
}

// GetShortVideo 获取短视频详情（带 Redis 缓存）
func (s *ShortVideoService) GetShortVideo(id uuid.UUID) (*model.ShortVideo, error) {
	cacheKey := fmt.Sprintf("short:detail:%s", id)

	// 尝试从缓存获取
	if database.RDB != nil {
		cached, err := database.RDB.Get(context.Background(), cacheKey).Result()
		if err == nil && cached != "" {
			var sv model.ShortVideo
			if err := json.Unmarshal([]byte(cached), &sv); err == nil {
				return &sv, nil
			}
		}
	}

	// 从数据库获取
	sv, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// 写入缓存（10 分钟）
	if database.RDB != nil {
		data, _ := json.Marshal(sv)
		database.RDB.Set(context.Background(), cacheKey, data, 10*time.Minute)
	}

	return sv, nil
}

// ListShortVideos 获取短视频列表（带 Redis 缓存）
func (s *ShortVideoService) ListShortVideos(page, pageSize int, sortBy string) ([]model.ShortVideo, int64, error) {
	cacheKey := fmt.Sprintf("short:list:%s:%d:%d", sortBy, page, pageSize)

	// 尝试从缓存获取
	if database.RDB != nil {
		cached, err := database.RDB.Get(context.Background(), cacheKey).Result()
		if err == nil && cached != "" {
			var result struct {
				List  []model.ShortVideo `json:"list"`
				Total int64              `json:"total"`
			}
			if err := json.Unmarshal([]byte(cached), &result); err == nil {
				return result.List, result.Total, nil
			}
		}
	}

	// 从数据库获取
	shorts, total, err := s.repo.List(page, pageSize, sortBy)
	if err != nil {
		return nil, 0, err
	}

	// 写入缓存（5 分钟）
	if database.RDB != nil {
		data, _ := json.Marshal(map[string]interface{}{
			"list":  shorts,
			"total": total,
		})
		database.RDB.Set(context.Background(), cacheKey, data, 5*time.Minute)
	}

	return shorts, total, nil
}

// GetRandom 获取随机短视频
func (s *ShortVideoService) GetRandom(limit int) ([]model.ShortVideo, error) {
	return s.repo.GetRandom(limit)
}

// IncrementView 增加播放量
func (s *ShortVideoService) IncrementView(id uuid.UUID) error {
	return s.repo.IncrementView(id)
}

// IncrementLike 增加点赞数
func (s *ShortVideoService) IncrementLike(id uuid.UUID) error {
	return s.repo.IncrementLike(id)
}

// SearchShortVideos 搜索短视频
func (s *ShortVideoService) SearchShortVideos(keyword string, page, pageSize int) ([]model.ShortVideo, int64, error) {
	return s.repo.Search(keyword, page, pageSize)
}

// GetHot 获取热门短视频（带 Redis 缓存）
func (s *ShortVideoService) GetHot(limit int) ([]model.ShortVideo, error) {
	cacheKey := fmt.Sprintf("short:hot:%d", limit)

	// 尝试从缓存获取
	if database.RDB != nil {
		cached, err := database.RDB.Get(context.Background(), cacheKey).Result()
		if err == nil && cached != "" {
			var shorts []model.ShortVideo
			if err := json.Unmarshal([]byte(cached), &shorts); err == nil {
				return shorts, nil
			}
		}
	}

	// 从数据库获取
	shorts, err := s.repo.GetHot(limit)
	if err != nil {
		return nil, err
	}

	// 写入缓存（5 分钟）
	if database.RDB != nil {
		data, _ := json.Marshal(shorts)
		database.RDB.Set(context.Background(), cacheKey, data, 5*time.Minute)
	}

	return shorts, nil
}

// ClearShortVideoCache 清除短视频缓存
func (s *ShortVideoService) ClearShortVideoCache(id uuid.UUID) {
	if database.RDB == nil {
		return
	}
	ctx := context.Background()
	database.RDB.Del(ctx, fmt.Sprintf("short:detail:%s", id))
	// 清除列表缓存
	iter := database.RDB.Scan(ctx, 0, "short:list:*", 100).Iterator()
	for iter.Next(ctx) {
		database.RDB.Del(ctx, iter.Val())
	}
	// 清除热门缓存
	iter = database.RDB.Scan(ctx, 0, "short:hot:*", 100).Iterator()
	for iter.Next(ctx) {
		database.RDB.Del(ctx, iter.Val())
	}
}
