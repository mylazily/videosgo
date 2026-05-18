package service

import (
	"videosgo/internal/model"
	"videosgo/internal/repository"
)

// VideoService 视频服务
type VideoService struct {
	videoRepo *repository.VideoRepository
}

// NewVideoService 创建视频服务
func NewVideoService(videoRepo *repository.VideoRepository) *VideoService {
	return &VideoService{videoRepo: videoRepo}
}

// ListVideos 获取视频列表
func (s *VideoService) ListVideos(offset, limit int) ([]model.Video, error) {
	return s.videoRepo.List(offset, limit)
}

// GetVideo 获取视频详情
func (s *VideoService) GetVideo(id string) (*model.Video, error) {
	return s.videoRepo.GetByID(id)
}
