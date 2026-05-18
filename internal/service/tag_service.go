package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"videosgo/internal/database"
	"videosgo/internal/model"
	"videosgo/internal/repository"
)

// TagService 标签服务
type TagService struct {
	repo *repository.TagRepo
}

// NewTagService 创建标签服务
func NewTagService(repo *repository.TagRepo) *TagService {
	return &TagService{repo: repo}
}

// GetTagBySlug 根据 Slug 获取标签（带 Redis 缓存）
func (s *TagService) GetTagBySlug(slug string) (*model.Tag, error) {
	cacheKey := fmt.Sprintf("tag:slug:%s", slug)

	// 尝试从缓存获取
	if database.RDB != nil {
		cached, err := database.RDB.Get(context.Background(), cacheKey).Result()
		if err == nil && cached != "" {
			var tag model.Tag
			if err := json.Unmarshal([]byte(cached), &tag); err == nil {
				return &tag, nil
			}
		}
	}

	// 从数据库获取
	tag, err := s.repo.GetBySlug(slug)
	if err != nil {
		return nil, err
	}

	// 写入缓存（10 分钟）
	if database.RDB != nil {
		data, _ := json.Marshal(tag)
		database.RDB.Set(context.Background(), cacheKey, data, 10*time.Minute)
	}

	return tag, nil
}

// ListTags 获取标签列表（带 Redis 缓存）
func (s *TagService) ListTags(page, pageSize int) ([]model.Tag, int64, error) {
	cacheKey := fmt.Sprintf("tag:list:%d:%d", page, pageSize)

	// 尝试从缓存获取
	if database.RDB != nil {
		cached, err := database.RDB.Get(context.Background(), cacheKey).Result()
		if err == nil && cached != "" {
			var result struct {
				List  []model.Tag `json:"list"`
				Total int64       `json:"total"`
			}
			if err := json.Unmarshal([]byte(cached), &result); err == nil {
				return result.List, result.Total, nil
			}
		}
	}

	// 从数据库获取
	tags, total, err := s.repo.ListTags(page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	// 写入缓存（10 分钟）
	if database.RDB != nil {
		data, _ := json.Marshal(map[string]interface{}{
			"list":  tags,
			"total": total,
		})
		database.RDB.Set(context.Background(), cacheKey, data, 10*time.Minute)
	}

	return tags, total, nil
}

// GetTrendingTags 获取热门标签（带 Redis 缓存）
func (s *TagService) GetTrendingTags(limit int) ([]model.Tag, error) {
	cacheKey := fmt.Sprintf("tag:trending:%d", limit)

	// 尝试从缓存获取
	if database.RDB != nil {
		cached, err := database.RDB.Get(context.Background(), cacheKey).Result()
		if err == nil && cached != "" {
			var tags []model.Tag
			if err := json.Unmarshal([]byte(cached), &tags); err == nil {
				return tags, nil
			}
		}
	}

	// 从数据库获取
	tags, err := s.repo.GetTrendingTags(limit)
	if err != nil {
		return nil, err
	}

	// 写入缓存（10 分钟）
	if database.RDB != nil {
		data, _ := json.Marshal(tags)
		database.RDB.Set(context.Background(), cacheKey, data, 10*time.Minute)
	}

	return tags, nil
}

// GetVideoTags 获取视频的所有标签
func (s *TagService) GetVideoTags(videoID uuid.UUID) ([]model.Tag, error) {
	return s.repo.GetVideoTags(videoID)
}

// GetVideosByTag 获取标签下的视频（分页）
func (s *TagService) GetVideosByTag(tagID uuid.UUID, page, pageSize int) ([]model.Video, int64, error) {
	return s.repo.GetVideosByTag(tagID, page, pageSize)
}

// SearchTags 搜索标签
func (s *TagService) SearchTags(keyword string) ([]model.Tag, error) {
	return s.repo.SearchTags(keyword)
}

// SyncVideoTags 同步视频标签（增删关联）
func (s *TagService) SyncVideoTags(videoID uuid.UUID, tagNames []string) error {
	// 获取当前标签
	currentTags, err := s.repo.GetVideoTags(videoID)
	if err != nil {
		return fmt.Errorf("获取当前标签失败: %w", err)
	}

	// 自动创建不存在的标签
	tags, err := s.repo.AutoCreateTags(tagNames)
	if err != nil {
		return fmt.Errorf("自动创建标签失败: %w", err)
	}

	// 构建新标签名称集合
	newTagNames := make(map[string]bool)
	for _, name := range tagNames {
		newTagNames[name] = true
	}

	// 移除不再需要的标签关联
	for _, currentTag := range currentTags {
		if !newTagNames[currentTag.Name] {
			if err := s.repo.RemoveVideoTag(videoID, currentTag.ID); err != nil {
				fmt.Printf("[Tag] 移除视频标签关联失败: %v\n", err)
			}
			if err := s.repo.DecrementVideoCount(currentTag.ID); err != nil {
				fmt.Printf("[Tag] 减少标签视频计数失败: %v\n", err)
			}
		}
	}

	// 添加新的标签关联
	for _, tag := range tags {
		// 检查是否已存在关联
		exists := false
		for _, currentTag := range currentTags {
			if currentTag.ID == tag.ID {
				exists = true
				break
			}
		}
		if !exists {
			if err := s.repo.AddVideoTag(videoID, tag.ID); err != nil {
				fmt.Printf("[Tag] 添加视频标签关联失败: %v\n", err)
			}
			if err := s.repo.IncrementVideoCount(tag.ID); err != nil {
				fmt.Printf("[Tag] 增加标签视频计数失败: %v\n", err)
			}
		}
	}

	return nil
}

// ClearTagCache 清除标签缓存
func (s *TagService) ClearTagCache(slug string) {
	if database.RDB == nil {
		return
	}
	ctx := context.Background()
	database.RDB.Del(ctx, fmt.Sprintf("tag:slug:%s", slug))
	// 清除列表缓存
	iter := database.RDB.Scan(ctx, 0, "tag:list:*", 100).Iterator()
	for iter.Next(ctx) {
		database.RDB.Del(ctx, iter.Val())
	}
	// 清除热门标签缓存
	iter = database.RDB.Scan(ctx, 0, "tag:trending:*", 100).Iterator()
	for iter.Next(ctx) {
		database.RDB.Del(ctx, iter.Val())
	}
}
