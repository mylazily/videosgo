package repository

import (
	"encoding/json"
	"strings"

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

	// Escape special characters to prevent SQL injection
	keyword = strings.ReplaceAll(keyword, "%", "\\%")
	keyword = strings.ReplaceAll(keyword, "_", "\\_")

	db := r.db.Model(&model.Video{}).
		Where("status = ? AND (title ILIKE ? ESCAPE '\\' OR actors ILIKE ? ESCAPE '\\' OR director ILIKE ? ESCAPE '\\')",
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

// GetHotPaged 分页获取热门视频
func (r *VideoRepo) GetHotPaged(page, pageSize int) ([]model.Video, int64, error) {
	var videos []model.Video
	var total int64

	offset := (page - 1) * pageSize

	err := r.db.Model(&model.Video{}).Where("status = ?", "active").Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Where("status = ?", "active").
		Order("view_count DESC").
		Offset(offset).Limit(pageSize).
		Find(&videos).Error

	return videos, total, err
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
		"cover":       video.CoverURL,
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
		Order("number ASC").
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

// ==================== 标题清洗去重相关方法 ====================

// FindByCleanTitle 根据清洗标题查找视频（精确匹配）
func (r *VideoRepo) FindByCleanTitle(cleanTitle string) (*model.Video, error) {
	var video model.Video
	err := r.db.Where("clean_title = ?", cleanTitle).First(&video).Error
	if err != nil {
		return nil, err
	}
	return &video, nil
}

// FindSimilarVideos 根据清洗标题查找相似视频（模糊匹配）
// 返回 clean_title 不为空且与给定标题可能相似的视频列表
// 注意：精确的相似度计算需要在应用层进行（因为数据库无法直接计算编辑距离）
func (r *VideoRepo) FindSimilarVideos(cleanTitle string, threshold float64) ([]model.Video, error) {
	if cleanTitle == "" {
		return nil, nil
	}

	// 策略：从数据库中获取 clean_title 不为空的活跃视频，
	// 然后在应用层做相似度过滤
	// 为了性能，限制候选集大小
	var videos []model.Video
	err := r.db.Where("clean_title != '' AND status = ?", "active").
		Limit(500).
		Find(&videos).Error
	if err != nil {
		return nil, err
	}

	return videos, nil
}

// FindByCleanTitlePrefix 根据清洗标题前缀查找候选视频
// 用于缩小模糊匹配的候选范围，提升性能
func (r *VideoRepo) FindByCleanTitlePrefix(prefix string) ([]model.Video, error) {
	if prefix == "" {
		return nil, nil
	}

	var videos []model.Video
	err := r.db.Where("clean_title LIKE ? AND clean_title != '' AND status = ?",
		prefix+"%", "active").
		Limit(100).
		Find(&videos).Error
	if err != nil {
		return nil, err
	}

	return videos, nil
}

// AppendPlayLine 向视频追加一条播放线路
// 使用 PostgreSQL jsonb_array_append 或在应用层合并后更新
func (r *VideoRepo) AppendPlayLine(videoID uuid.UUID, line model.PlayLineJSON) error {
	// 获取当前视频
	var video model.Video
	if err := r.db.Select("id, play_lines").First(&video, "id = ?", videoID).Error; err != nil {
		return err
	}

	// 追加新线路
	video.PlayLines = append(video.PlayLines, line)

	// 更新回数据库
	return r.db.Model(&model.Video{}).Where("id = ?", videoID).
		Update("play_lines", video.PlayLines).Error
}

// AppendPlayLines 批量向视频追加播放线路（去重）
func (r *VideoRepo) AppendPlayLines(videoID uuid.UUID, newLines []model.PlayLineJSON) error {
	// 获取当前视频
	var video model.Video
	if err := r.db.Select("id, play_lines").First(&video, "id = ?", videoID).Error; err != nil {
		return err
	}

	// 构建已有线路的 URL 集合，用于去重
	existingURLs := make(map[string]bool)
	for _, line := range video.PlayLines {
		existingURLs[line.M3U8URL] = true
	}

	// 追加不重复的新线路
	for _, line := range newLines {
		if !existingURLs[line.M3U8URL] {
			video.PlayLines = append(video.PlayLines, line)
			existingURLs[line.M3U8URL] = true
		}
	}

	// 更新回数据库
	return r.db.Model(&model.Video{}).Where("id = ?", videoID).
		Update("play_lines", video.PlayLines).Error
}

// UpdateDomainPool 更新域名池
func (r *VideoRepo) UpdateDomainPool(videoID uuid.UUID, domainPool []string, sharedPath string) error {
	// 将域名池序列化为 JSON
	poolJSON, err := json.Marshal(domainPool)
	if err != nil {
		return err
	}

	return r.db.Model(&model.Video{}).Where("id = ?", videoID).
		Updates(map[string]interface{}{
			"domain_pool": string(poolJSON),
			"shared_path": sharedPath,
		}).Error
}

// IncrementSourceCount 增加资源站计数
func (r *VideoRepo) IncrementSourceCount(videoID uuid.UUID) error {
	return r.db.Model(&model.Video{}).Where("id = ?", videoID).
		UpdateColumn("source_count", gorm.Expr("source_count + 1")).Error
}

// UpdateCleanTitle 更新清洗后的标题
func (r *VideoRepo) UpdateCleanTitle(videoID uuid.UUID, cleanTitle string) error {
	return r.db.Model(&model.Video{}).Where("id = ?", videoID).
		Update("clean_title", cleanTitle).Error
}
