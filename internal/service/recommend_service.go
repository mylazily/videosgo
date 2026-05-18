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

// RecommendService 推荐服务
type RecommendService struct {
	tagRepo   *repository.TagRepo
	videoRepo *repository.VideoRepo
}

// NewRecommendService 创建推荐服务
func NewRecommendService(
	tagRepo *repository.TagRepo,
	videoRepo *repository.VideoRepo,
) *RecommendService {
	return &RecommendService{
		tagRepo:   tagRepo,
		videoRepo: videoRepo,
	}
}

// GetRelatedVideos 基于标签的关联推荐
func (s *RecommendService) GetRelatedVideos(videoID uuid.UUID, limit int) ([]model.Video, error) {
	cacheKey := fmt.Sprintf("recommend:related:%s:%d", videoID, limit)

	// 尝试从缓存获取
	if database.RDB != nil {
		cached, err := database.RDB.Get(context.Background(), cacheKey).Result()
		if err == nil && cached != "" {
			var videos []model.Video
			if err := json.Unmarshal([]byte(cached), &videos); err == nil {
				return videos, nil
			}
		}
	}

	// 获取视频的标签
	tags, err := s.tagRepo.GetVideoTags(videoID)
	if err != nil || len(tags) == 0 {
		// 如果没有标签，返回随机视频
		return s.videoRepo.GetRandom(limit)
	}

	// 收集所有标签 ID
	tagIDs := make([]uuid.UUID, 0, len(tags))
	for _, tag := range tags {
		tagIDs = append(tagIDs, tag.ID)
	}

	// 通过标签查找关联视频（排除当前视频）
	var videos []model.Video
	err = database.DB.Table("videos").
		Select("DISTINCT videos.*").
		Joins("JOIN video_tags ON video_tags.video_id = videos.id").
		Where("video_tags.tag_id IN ? AND videos.id != ? AND videos.status = ?", tagIDs, videoID, "active").
		Order("videos.view_count DESC").
		Limit(limit).
		Find(&videos).Error

	if err != nil {
		return nil, err
	}

	// 如果关联视频不足，用随机视频补充
	if len(videos) < limit {
		remaining := limit - len(videos)
		existingIDs := make(map[uuid.UUID]bool)
		for _, v := range videos {
			existingIDs[v.ID] = true
		}

		randomVideos, _ := s.videoRepo.GetRandom(remaining * 2)
		for _, rv := range randomVideos {
			if len(videos) >= limit {
				break
			}
			if !existingIDs[rv.ID] {
				videos = append(videos, rv)
				existingIDs[rv.ID] = true
			}
		}
	}

	// 写入缓存（30 分钟）
	if database.RDB != nil {
		data, _ := json.Marshal(videos)
		database.RDB.Set(context.Background(), cacheKey, data, 30*time.Minute)
	}

	return videos, nil
}

// GetPersonalizedRecommendations 个性化推荐
func (s *RecommendService) GetPersonalizedRecommendations(fingerprintID string, limit int) ([]model.Video, error) {
	cacheKey := fmt.Sprintf("recommend:personal:%s:%d", fingerprintID, limit)

	// 尝试从缓存获取
	if database.RDB != nil {
		cached, err := database.RDB.Get(context.Background(), cacheKey).Result()
		if err == nil && cached != "" {
			var videos []model.Video
			if err := json.Unmarshal([]byte(cached), &videos); err == nil {
				return videos, nil
			}
		}
	}

	// 从 Redis 获取用户偏好标签
	prefKey := fmt.Sprintf("user:pref:%s", fingerprintID)
	var videos []model.Video

	if database.RDB != nil {
		// 获取用户偏好的标签 ID 列表
		tagIDs, err := database.RDB.SMembers(context.Background(), prefKey).Result()
		if err == nil && len(tagIDs) > 0 {
			// 基于偏好标签推荐
			uuids := make([]uuid.UUID, 0, len(tagIDs))
			for _, tid := range tagIDs {
				id, err := uuid.Parse(tid)
				if err == nil {
					uuids = append(uuids, id)
				}
			}

			if len(uuids) > 0 {
				err = database.DB.Table("videos").
					Select("DISTINCT videos.*").
					Joins("JOIN video_tags ON video_tags.video_id = videos.id").
					Where("video_tags.tag_id IN ? AND videos.status = ?", uuids, "active").
					Order("videos.created_at DESC").
					Limit(limit).
					Find(&videos).Error
				if err == nil && len(videos) > 0 {
					// 写入缓存（30 分钟）
					data, _ := json.Marshal(videos)
					database.RDB.Set(context.Background(), cacheKey, data, 30*time.Minute)
					return videos, nil
				}
			}
		}
	}

	// 降级：返回热门视频
	hotVideos, err := s.videoRepo.GetHot(limit)
	if err != nil {
		// 最终降级：返回随机视频
		return s.videoRepo.GetRandom(limit)
	}

	// 写入缓存（30 分钟）
	if database.RDB != nil {
		data, _ := json.Marshal(hotVideos)
		database.RDB.Set(context.Background(), cacheKey, data, 30*time.Minute)
	}

	return hotVideos, nil
}

// RecordUserPreference 记录用户偏好
func (s *RecommendService) RecordUserPreference(fingerprintID string, videoID uuid.UUID) error {
	if database.RDB == nil {
		return nil
	}

	// 获取视频的标签
	tags, err := s.tagRepo.GetVideoTags(videoID)
	if err != nil || len(tags) == 0 {
		return nil
	}

	ctx := context.Background()
	prefKey := fmt.Sprintf("user:pref:%s", fingerprintID)

	// 将视频标签添加到用户偏好集合中
	for _, tag := range tags {
		database.RDB.SAdd(ctx, prefKey, tag.ID.String())
	}

	// 设置过期时间（7 天）
	database.RDB.Expire(ctx, prefKey, 7*24*time.Hour)

	return nil
}
