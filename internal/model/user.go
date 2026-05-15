package model

import (
	"time"

	"github.com/google/uuid"
)

// User 用户模型
type User struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Username  string    `gorm:"type:varchar(50);uniqueIndex;not null;comment:用户名" json:"username"`
	Password  string    `gorm:"type:varchar(200);not null;comment:密码哈希" json:"-"`
	Avatar    string    `gorm:"type:varchar(500);comment:头像 URL" json:"avatar"`
	IsAdmin   bool      `gorm:"default:false;comment:是否管理员" json:"is_admin"`
	Status    string    `gorm:"type:varchar(20);default:active;comment:状态" json:"status"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
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
