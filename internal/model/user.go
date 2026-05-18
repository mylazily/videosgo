package model

import (
	"time"
)

// User 用户模型
type User struct {
	ID        string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()" db:"id"`
	Username  string    `json:"username" gorm:"uniqueIndex;size:100;not null" db:"username"`
	Password  string    `json:"-" gorm:"column:password_hash;size:255;not null" db:"password_hash"`
	Nickname  string    `json:"nickname" gorm:"size:100;not null;default:''" db:"nickname"`
	Avatar    string    `json:"avatar" gorm:"column:avatar_url;size:1024;default:''" db:"avatar_url"`
	Email     string    `json:"email" gorm:"size:255;default:''" db:"email"`
	Phone     string    `json:"phone,omitempty" gorm:"size:50;default:''" db:"phone"`
	Gender    int       `json:"gender" gorm:"default:0" db:"gender"`
	Bio       string    `json:"bio" gorm:"size:500;default:''" db:"bio"`
	Role      string    `json:"role" gorm:"type:user_role;not null;default:'user'" db:"role"`
	Status    string    `json:"status" gorm:"type:user_status;not null;default:'active'" db:"status"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime" db:"updated_at"`
}

func (User) TableName() string {
	return "users"
}

// UserResponse 用户响应（不包含敏感信息）
type UserResponse struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Nickname  string    `json:"nickname"`
	Avatar    string    `json:"avatar"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// ToResponse 转换为响应格式
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		Username:  u.Username,
		Nickname:  u.Nickname,
		Avatar:    u.Avatar,
		Email:     u.Email,
		Role:      u.Role,
		Status:    u.Status,
		CreatedAt: u.CreatedAt,
	}
}
