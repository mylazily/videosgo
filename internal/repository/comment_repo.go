package repository

import (
	"github.com/mylazily/videosgo/internal/model"
	"gorm.io/gorm"
)

// CommentRepo 评论数据仓库
type CommentRepo struct {
	db *gorm.DB
}

// NewCommentRepo 创建评论仓库
func NewCommentRepo(db *gorm.DB) *CommentRepo {
	return &CommentRepo{db: db}
}

// Create 创建评论
func (r *CommentRepo) Create(comment *model.Comment) error {
	return r.db.Create(comment).Error
}

// Delete 删除评论
func (r *CommentRepo) Delete(id uint) error {
	return r.db.Delete(&model.Comment{}, id).Error
}

// GetByID 根据 ID 获取评论
func (r *CommentRepo) GetByID(id uint) (*model.Comment, error) {
	var comment model.Comment
	err := r.db.Preload("User").First(&comment, id).Error
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

// ListByVideoID 获取视频的评论列表
func (r *CommentRepo) ListByVideoID(videoID uint, page, pageSize int) ([]model.Comment, int64, error) {
	var comments []model.Comment
	var total int64

	db := r.db.Model(&model.Comment{}).Where("video_id = ? AND status = ?", videoID, "active")
	db.Count(&total)

	err := db.Preload("User").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Order("created_at DESC").
		Find(&comments).Error
	return comments, total, err
}

// ListReplies 获取回复列表
func (r *CommentRepo) ListReplies(parentID uint) ([]model.Comment, error) {
	var comments []model.Comment
	err := r.db.Preload("User").
		Where("parent_id = ? AND status = ?", parentID, "active").
		Order("created_at ASC").
		Find(&comments).Error
	return comments, err
}

// IncrementLikeCount 增加点赞数
func (r *CommentRepo) IncrementLikeCount(id uint) error {
	return r.db.Model(&model.Comment{}).Where("id = ?", id).
		UpdateColumn("like_count", gorm.Expr("like_count + 1")).Error
}

// DecrementLikeCount 减少点赞数
func (r *CommentRepo) DecrementLikeCount(id uint) error {
	return r.db.Model(&model.Comment{}).Where("id = ?", id).
		UpdateColumn("like_count", gorm.Expr("GREATEST(like_count - 1, 0)")).Error
}

// CreateLike 创建评论点赞
func (r *CommentRepo) CreateLike(like *model.CommentLike) error {
	return r.db.Create(like).Error
}

// DeleteLike 删除评论点赞
func (r *CommentRepo) DeleteLike(commentID, userID uint) error {
	return r.db.Where("comment_id = ? AND user_id = ?", commentID, userID).
		Delete(&model.CommentLike{}).Error
}

// IsLiked 检查用户是否已点赞
func (r *CommentRepo) IsLiked(commentID, userID uint) (bool, error) {
	var count int64
	err := r.db.Model(&model.CommentLike{}).
		Where("comment_id = ? AND user_id = ?", commentID, userID).
		Count(&count).Error
	return count > 0, err
}

// GetCountByVideoID 获取视频评论数
func (r *CommentRepo) GetCountByVideoID(videoID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.Comment{}).
		Where("video_id = ? AND status = ?", videoID, "active").
		Count(&count).Error
	return count, err
}
