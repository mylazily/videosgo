package repository

import (
	"github.com/google/uuid"
	"github.com/mylazily/videosgo/internal/model"
	"gorm.io/gorm"
)

// VideoRepo 视频数据仓库
type VideoRepo struct {
	db *gorm.DB
}

// NewVideoRepo 创建视频仓库
func NewVideoRepo(db *gorm.DB) *VideoRepo {
	return &VideoRepo{db: db}
}

// Create 创建视频
func (r *VideoRepo) Create(video *model.Video) error {
	return r.db.Create(video).Error
}

// Update 更新视频
func (r *VideoRepo) Update(video *model.Video) error {
	return r.db.Save(video).Error
}

// Delete 删除视频
func (r *VideoRepo) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.Video{}, "id = ?", id).Error
}

// GetByID 根据 ID 获取视频（含剧集）
func (r *VideoRepo) GetByID(id uuid.UUID) (*model.Video, error) {
	var video model.Video
	err := r.db.Preload("Episodes").First(&video, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &video, nil
}

// GetByTitle 根据标题获取视频
func (r *VideoRepo) GetByTitle(title string) (*model.Video, error) {
	var video model.Video
	err := r.db.Where("title = ?", title).First(&video).Error
	if err != nil {
		return nil, err
	}
	return &video, nil
}

// List 获取视频列表（分页）
func (r *VideoRepo) List(page, pageSize int, category string) ([]model.Video, int64, error) {
	var videos []model.Video
	var total int64

	db := r.db.Model(&model.Video{}).Where("status = ?", "active")

	if category != "" {
		db = db.Where("category = ?", category)
	}

	db.Count(&total)

	err := db.Offset((page - 1) * pageSize).Limit(pageSize).
		Order("created_at DESC").
		Find(&videos).Error
	return videos, total, err
}

// Search 搜索视频
func (r *VideoRepo) Search(keyword string, page, pageSize int) ([]model.Video, int64, error) {
	var videos []model.Video
	var total int64

	db := r.db.Model(&model.Video{}).
		Where("status = ? AND (title LIKE ? OR actors LIKE ? OR director LIKE ?)",
			"active", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")

	db.Count(&total)

	err := db.Offset((page - 1) * pageSize).Limit(pageSize).
		Order("created_at DESC").
		Find(&videos).Error
	return videos, total, err
}

// GetCategories 获取所有分类
func (r *VideoRepo) GetCategories() ([]string, error) {
	var categories []string
	err := r.db.Model(&model.Video{}).
		Distinct("category").
		Where("category != '' AND status = ?", "active").
		Pluck("category", &categories).Error
	return categories, err
}

// IncrementViewCount 增加播放量
func (r *VideoRepo) IncrementViewCount(id uuid.UUID) error {
	return r.db.Model(&model.Video{}).Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

// GetRandom 获取随机推荐视频
func (r *VideoRepo) GetRandom(limit int) ([]model.Video, error) {
	var videos []model.Video
	err := r.db.Where("status = ?", "active").
		Order("RANDOM()").
		Limit(limit).
		Find(&videos).Error
	return videos, err
}

// GetLatest 获取最新视频
func (r *VideoRepo) GetLatest(limit int) ([]model.Video, error) {
	var videos []model.Video
	err := r.db.Where("status = ?", "active").
		Order("created_at DESC").
		Limit(limit).
		Find(&videos).Error
	return videos, err
}

// GetHot 获取热门视频（按播放量）
func (r *VideoRepo) GetHot(limit int) ([]model.Video, error) {
	var videos []model.Video
	err := r.db.Where("status = ?", "active").
		Order("view_count DESC").
		Limit(limit).
		Find(&videos).Error
	return videos, err
}

// GetByCategory 根据分类获取视频
func (r *VideoRepo) GetByCategory(category string, page, pageSize int) ([]model.Video, int64, error) {
	return r.List(page, pageSize, category)
}

// UpsertByTitle 根据标题创建或更新视频
func (r *VideoRepo) UpsertByTitle(video *model.Video) error {
	var existing model.Video
	err := r.db.Where("title = ?", video.Title).First(&existing).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return r.db.Create(video).Error
		}
		return err
	}
	// 更新已有视频
	return r.db.Model(&existing).Updates(map[string]interface{}{
		"cover":       video.Cover,
		"description": video.Description,
		"category":    video.Category,
		"year":        video.Year,
		"area":        video.Area,
		"director":    video.Director,
		"actors":      video.Actors,
		"tags":        video.Tags,
		"remarks":     video.Remarks,
		"play_links":  video.PlayLinks,
		"source_id":   video.SourceID,
	}).Error
}

// CreateEpisode 创建剧集
func (r *VideoRepo) CreateEpisode(episode *model.Episode) error {
	return r.db.Create(episode).Error
}

// GetEpisodesByVideoID 获取视频的剧集列表
func (r *VideoRepo) GetEpisodesByVideoID(videoID uuid.UUID) ([]model.Episode, error) {
	var episodes []model.Episode
	err := r.db.Where("video_id = ?", videoID).
		Order("ep_index ASC").
		Find(&episodes).Error
	return episodes, err
}

// SaveWatchHistory 保存观看历史
func (r *VideoRepo) SaveWatchHistory(history *model.UserWatchHistory) error {
	return r.db.Save(history).Error
}

// GetWatchHistory 获取用户观看历史
func (r *VideoRepo) GetWatchHistory(userID uuid.UUID, page, pageSize int) ([]model.UserWatchHistory, int64, error) {
	var histories []model.UserWatchHistory
	var total int64

	db := r.db.Model(&model.UserWatchHistory{}).Where("user_id = ?", userID)
	db.Count(&total)

	err := db.Preload("Video").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Order("updated_at DESC").
		Find(&histories).Error
	return histories, total, err
}

// UpsertSearchHot 更新热搜
func (r *VideoRepo) UpsertSearchHot(keyword string) error {
	var hot model.SearchHot
	err := r.db.Where("keyword = ?", keyword).First(&hot).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return r.db.Create(&model.SearchHot{Keyword: keyword, Count: 1}).Error
		}
		return err
	}
	return r.db.Model(&hot).UpdateColumn("count", gorm.Expr("count + 1")).Error
}

// GetSearchHot 获取热搜列表
func (r *VideoRepo) GetSearchHot(limit int) ([]model.SearchHot, error) {
	var hots []model.SearchHot
	err := r.db.Order("count DESC").Limit(limit).Find(&hots).Error
	return hots, err
}
