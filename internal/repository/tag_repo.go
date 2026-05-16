package repository

import (
	"strings"

	"github.com/google/uuid"
	"github.com/mylazily/videosgo/internal/model"
	"gorm.io/gorm"
)

// TagRepo 标签数据仓库
type TagRepo struct {
	db *gorm.DB
}

// NewTagRepo 创建标签仓库
func NewTagRepo(db *gorm.DB) *TagRepo {
	return &TagRepo{db: db}
}

// Create 创建标签
func (r *TagRepo) Create(tag *model.Tag) error {
	return r.db.Create(tag).Error
}

// Update 更新标签
func (r *TagRepo) Update(tag *model.Tag) error {
	return r.db.Save(tag).Error
}

// Delete 删除标签
func (r *TagRepo) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.Tag{}, "id = ?", id).Error
}

// GetByID 根据 ID 获取标签
func (r *TagRepo) GetByID(id uuid.UUID) (*model.Tag, error) {
	var tag model.Tag
	err := r.db.First(&tag, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

// GetBySlug 根据 Slug 获取标签
func (r *TagRepo) GetBySlug(slug string) (*model.Tag, error) {
	var tag model.Tag
	err := r.db.Where("slug = ?", slug).First(&tag).Error
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

// ListTags 获取标签列表（按视频数降序，分页）
func (r *TagRepo) ListTags(page, pageSize int) ([]model.Tag, int64, error) {
	var tags []model.Tag
	var total int64

	db := r.db.Model(&model.Tag{}).Where("status = ?", "active")
	db.Count(&total)

	err := db.Offset((page - 1) * pageSize).Limit(pageSize).
		Order("video_count DESC, sort_order ASC").
		Find(&tags).Error
	return tags, total, err
}

// GetTrendingTags 获取热门标签（按近期关联视频数排序）
func (r *TagRepo) GetTrendingTags(limit int) ([]model.Tag, error) {
	var tags []model.Tag
	// 使用子查询方式，避免字段映射问题
	// 先获取所有活跃标签
	err := r.db.Where("status = ?", "active").
		Order("video_count DESC").
		Limit(limit).
		Find(&tags).Error
	return tags, err
}

// AddVideoTag 添加视频标签关联
func (r *TagRepo) AddVideoTag(videoID, tagID uuid.UUID) error {
	vt := &model.VideoTag{
		VideoID: videoID,
		TagID:   tagID,
	}
	return r.db.Create(vt).Error
}

// RemoveVideoTag 移除视频标签关联
func (r *TagRepo) RemoveVideoTag(videoID, tagID uuid.UUID) error {
	return r.db.Where("video_id = ? AND tag_id = ?", videoID, tagID).
		Delete(&model.VideoTag{}).Error
}

// GetVideoTags 获取视频的所有标签
func (r *TagRepo) GetVideoTags(videoID uuid.UUID) ([]model.Tag, error) {
	var tags []model.Tag
	err := r.db.Joins("JOIN video_tags ON video_tags.tag_id = tags.id").
		Where("video_tags.video_id = ?", videoID).
		Find(&tags).Error
	return tags, err
}

// GetVideosByTag 获取标签下的视频（分页）
func (r *TagRepo) GetVideosByTag(tagID uuid.UUID, page, pageSize int) ([]model.Video, int64, error) {
	var videos []model.Video
	var total int64

	db := r.db.Model(&model.Video{}).
		Joins("JOIN video_tags ON video_tags.video_id = videos.id").
		Where("video_tags.tag_id = ? AND videos.status = ?", tagID, "active")
	db.Count(&total)

	err := db.Offset((page - 1) * pageSize).Limit(pageSize).
		Order("videos.created_at DESC").
		Find(&videos).Error
	return videos, total, err
}

// SearchTags 模糊搜索标签
func (r *TagRepo) SearchTags(keyword string) ([]model.Tag, error) {
	var tags []model.Tag
	// 转义 SQL LIKE 特殊字符
	keyword = strings.ReplaceAll(keyword, "%", "\\%")
	keyword = strings.ReplaceAll(keyword, "_", "\\_")

	err := r.db.Where("status = ? AND (name LIKE ? ESCAPE '\\' OR slug LIKE ? ESCAPE '\\')",
		"active", "%"+keyword+"%", "%"+keyword+"%").
		Order("video_count DESC").
		Limit(20).
		Find(&tags).Error
	return tags, err
}

// AutoCreateTags 批量自动创建标签（不存在则创建）
func (r *TagRepo) AutoCreateTags(names []string) ([]model.Tag, error) {
	var tags []model.Tag
	for _, name := range names {
		if name == "" {
			continue
		}
		var existing model.Tag
		err := r.db.Where("name = ?", name).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			// 生成 slug（简单处理，实际可使用更完善的 slug 生成逻辑）
			slug := generateSlug(name)
			newTag := &model.Tag{
				Name:   name,
				Slug:   slug,
				Status: "active",
			}
			if err := r.db.Create(newTag).Error; err != nil {
				continue
			}
			tags = append(tags, *newTag)
		} else if err == nil {
			tags = append(tags, existing)
		}
	}
	return tags, nil
}

// GetByName 根据名称获取标签
func (r *TagRepo) GetByName(name string) (*model.Tag, error) {
	var tag model.Tag
	err := r.db.Where("name = ?", name).First(&tag).Error
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

// IncrementVideoCount 增加标签的视频计数
func (r *TagRepo) IncrementVideoCount(tagID uuid.UUID) error {
	return r.db.Model(&model.Tag{}).Where("id = ?", tagID).
		UpdateColumn("video_count", gorm.Expr("video_count + 1")).Error
}

// DecrementVideoCount 减少标签的视频计数
func (r *TagRepo) DecrementVideoCount(tagID uuid.UUID) error {
	return r.db.Model(&model.Tag{}).Where("id = ?", tagID).
		UpdateColumn("video_count", gorm.Expr("GREATEST(video_count - 1, 0)")).Error
}

// generateSlug 根据名称生成 URL 友好的 slug
func generateSlug(name string) string {
	// 简单的 slug 生成，实际项目中可使用更完善的实现
	slug := name
	// 这里可以做更多处理，如去除特殊字符、转小写等
	return slug
}
