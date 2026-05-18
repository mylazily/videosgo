package service

import (
	"github.com/google/uuid"
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

// GetVideoByUUID 获取视频详情（通过 UUID）
func (s *VideoService) GetVideoByUUID(id uuid.UUID) (*model.Video, error) {
	return s.videoRepo.GetByID(id.String())
}

// GetRandomVideos 获取随机视频列表
func (s *VideoService) GetRandomVideos(limit int) ([]model.Video, error) {
	return s.videoRepo.GetRandom(limit)
}

// GetHotVideos 获取热门视频列表
func (s *VideoService) GetHotVideos(limit int) ([]model.Video, error) {
	return s.videoRepo.GetHot(limit)
}

// SearchVideos 搜索视频
func (s *VideoService) SearchVideos(keyword string, offset, limit int) ([]model.Video, error) {
	return s.videoRepo.Search(keyword, offset, limit)
}

// CreateVideo 创建视频
func (s *VideoService) CreateVideo(video *model.Video) error {
	return s.videoRepo.Create(video)
}

// UpdateVideo 更新视频
func (s *VideoService) UpdateVideo(video *model.Video) error {
	return s.videoRepo.Update(video)
}

// DeleteVideo 删除视频
func (s *VideoService) DeleteVideo(id string) error {
	return s.videoRepo.Delete(id)
}

// CountVideos 统计视频数量
func (s *VideoService) CountVideos() (int64, error) {
	return s.videoRepo.Count()
}
