package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID           uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Username     string         `gorm:"type:varchar(100);uniqueIndex;not null;comment:用户名" json:"username"`
	PasswordHash string         `gorm:"type:varchar(255);not null;comment:密码哈希" json:"-"`
	Nickname     string         `gorm:"type:varchar(100);comment:昵称" json:"nickname"`
	Email        string         `gorm:"type:varchar(255);comment:邮箱" json:"email"`
	Phone        string         `gorm:"type:varchar(20);comment:手机号" json:"phone"`
	Gender       string         `gorm:"type:varchar(10);comment:性别" json:"gender"`
	Birthday     *time.Time     `gorm:"type:date;comment:生日" json:"birthday"`
	Bio          string         `gorm:"type:text;comment:个人简介" json:"bio"`
	Avatar       string         `gorm:"type:varchar(500);comment:头像 URL" json:"avatar"`
	Role         string         `gorm:"type:varchar(20);default:'user';comment:角色 user/admin" json:"role"`
	Status       string         `gorm:"type:varchar(20);default:active;comment:状态" json:"status"`
	LastLoginAt  *time.Time     `gorm:"comment:最后登录时间" json:"last_login_at"`
	LastLoginIP  string         `gorm:"type:varchar(45);comment:最后登录IP" json:"last_login_ip"`
	LoginCount   int            `gorm:"default:0;comment:登录次数" json:"login_count"`
	ExtraInfo    JSONB          `gorm:"type:jsonb;comment:额外信息" json:"extra_info"`
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index;comment:删除时间" json:"-"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// BeforeCreate 创建前钩子，确保 UUID 生成
func (u *User) BeforeCreate() error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}
