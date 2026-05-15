package model

import "time"

// Comment 评论
type Comment struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	VideoID   uint      `gorm:"index;not null;comment:视频 ID" json:"video_id"`
	UserID    uint      `gorm:"index;not null;comment:用户 ID" json:"user_id"`
	Content   string    `gorm:"type:text;not null;comment:评论内容" json:"content"`
	ParentID  uint      `gorm:"default:0;comment:父评论 ID，0 为顶级评论" json:"parent_id"`
	LikeCount int       `gorm:"default:0;comment:点赞数" json:"like_count"`
	Status    string    `gorm:"type:varchar(20);default:active;comment:状态" json:"status"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// 关联
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName 指定表名
func (Comment) TableName() string {
	return "comments"
}

// CommentLike 评论点赞
type CommentLike struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CommentID uint      `gorm:"uniqueIndex:idx_comment_user;not null;comment:评论 ID" json:"comment_id"`
	UserID    uint      `gorm:"uniqueIndex:idx_comment_user;not null;comment:用户 ID" json:"user_id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TableName 指定表名
func (CommentLike) TableName() string {
	return "comment_likes"
}
