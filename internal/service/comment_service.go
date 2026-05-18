package service

import (
	"fmt"

	"github.com/google/uuid"
	"videosgo/internal/model"
	"videosgo/internal/repository"
)

// CommentService 评论服务
type CommentService struct {
	repo *repository.CommentRepo
}

// NewCommentService 创建评论服务
func NewCommentService(repo *repository.CommentRepo) *CommentService {
	return &CommentService{repo: repo}
}

// CreateComment 创建评论
func (s *CommentService) CreateComment(comment *model.Comment) error {
	if comment.Content == "" {
		return fmt.Errorf("评论内容不能为空")
	}
	if comment.VideoID == uuid.Nil {
		return fmt.Errorf("视频 ID 不能为空")
	}
	return s.repo.Create(comment)
}

// DeleteComment 删除评论
func (s *CommentService) DeleteComment(id, userID uuid.UUID, isAdmin bool) error {
	comment, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("评论不存在")
	}
	// 只有评论作者或管理员可以删除
	if comment.UserID != userID && !isAdmin {
		return fmt.Errorf("无权删除此评论")
	}
	return s.repo.Delete(id)
}

// GetComment 获取评论详情
func (s *CommentService) GetComment(id uuid.UUID) (*model.Comment, error) {
	return s.repo.GetByID(id)
}

// ListComments 获取视频评论列表
func (s *CommentService) ListComments(videoID uuid.UUID, page, pageSize int) ([]model.Comment, int64, error) {
	return s.repo.ListByVideoID(videoID, page, pageSize)
}

// ListReplies 获取回复列表
func (s *CommentService) ListReplies(parentID uuid.UUID) ([]model.Comment, error) {
	return s.repo.ListReplies(parentID)
}

// LikeComment 点赞评论
func (s *CommentService) LikeComment(commentID, userID uuid.UUID) error {
	// 检查是否已点赞
	liked, err := s.repo.IsLiked(commentID, userID)
	if err != nil {
		return fmt.Errorf("检查点赞状态失败: %w", err)
	}
	if liked {
		return fmt.Errorf("已经点赞过了")
	}

	// 创建点赞记录
	if err := s.repo.CreateLike(&model.CommentLike{
		CommentID: commentID,
		UserID:    userID,
	}); err != nil {
		return fmt.Errorf("点赞失败: %w", err)
	}

	// 增加点赞数
	return s.repo.IncrementLikeCount(commentID)
}

// UnlikeComment 取消点赞
func (s *CommentService) UnlikeComment(commentID, userID uuid.UUID) error {
	// 检查是否已点赞
	liked, err := s.repo.IsLiked(commentID, userID)
	if err != nil {
		return fmt.Errorf("检查点赞状态失败: %w", err)
	}
	if !liked {
		return fmt.Errorf("尚未点赞")
	}

	// 删除点赞记录
	if err := s.repo.DeleteLike(commentID, userID); err != nil {
		return fmt.Errorf("取消点赞失败: %w", err)
	}

	// 减少点赞数
	return s.repo.DecrementLikeCount(commentID)
}

// GetCommentCount 获取视频评论数
func (s *CommentService) GetCommentCount(videoID uuid.UUID) (int64, error) {
	return s.repo.GetCountByVideoID(videoID)
}
