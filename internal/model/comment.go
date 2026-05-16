package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Comment 评论模型
type Comment struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	VideoID     uuid.UUID      `gorm:"type:uuid;index;not null;comment:视频 ID" json:"video_id"`
	UserID      uuid.UUID      `gorm:"type:uuid;index;not null;comment:用户 ID" json:"user_id"`
	Content     string         `gorm:"type:text;not null;comment:评论内容" json:"content"`
	ParentID    uuid.UUID      `gorm:"type:uuid;default:null;comment:父评论 ID，nil 为顶级评论" json:"parent_id"`
	RootID      uuid.UUID      `gorm:"type:uuid;index;comment:根评论 ID" json:"root_id"`
	LikeCount   int            `gorm:"default:0;comment:点赞数" json:"like_count"`
	ReplyCount  int            `gorm:"default:0;comment:回复数" json:"reply_count"`
	Status      string         `gorm:"type:varchar(20);default:approved;comment:状态" json:"status"`
	IPLocation  string         `gorm:"type:varchar(100);comment:IP归属地" json:"ip_location"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index;comment:删除时间" json:"-"`

	// 关联
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName 指定表名
func (Comment) TableName() string {
	return "comments"
}

// BeforeCreate 创建前钩子，确保 UUID 生成
func (c *Comment) BeforeCreate() error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// CommentLike 评论点赞模型
type CommentLike struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CommentID uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_comment_user;not null;comment:评论 ID" json:"comment_id"`
	UserID    uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_comment_user;not null;comment:用户 ID" json:"user_id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TableName 指定表名
func (CommentLike) TableName() string {
	return "comment_likes"
}

// BeforeCreate 创建前钩子，确保 UUID 生成
func (c *CommentLike) BeforeCreate() error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}
