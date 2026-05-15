package repository

import (
	"github.com/google/uuid"
	"github.com/mylazily/videosgo/internal/model"
	"gorm.io/gorm"
)

// UserRepo 用户数据仓库
type UserRepo struct {
	db *gorm.DB
}

// NewUserRepo 创建用户仓库
func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db: db}
}

// Create 创建用户
func (r *UserRepo) Create(user *model.User) error {
	return r.db.Create(user).Error
}

// Update 更新用户
func (r *UserRepo) Update(user *model.User) error {
	return r.db.Save(user).Error
}

// Delete 删除用户
func (r *UserRepo) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.User{}, "id = ?", id).Error
}

// GetByID 根据 ID 获取用户
func (r *UserRepo) GetByID(id uuid.UUID) (*model.User, error) {
	var user model.User
	err := r.db.First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByUsername 根据用户名获取用户
func (r *UserRepo) GetByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// List 获取用户列表
func (r *UserRepo) List(page, pageSize int) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	db := r.db.Model(&model.User{})
	db.Count(&total)

	err := db.Offset((page - 1) * pageSize).Limit(pageSize).
		Order("created_at DESC").
		Find(&users).Error
	return users, total, err
}

// ExistsByUsername 检查用户名是否存在
func (r *UserRepo) ExistsByUsername(username string) (bool, error) {
	var count int64
	err := r.db.Model(&model.User{}).Where("username = ?", username).Count(&count).Error
	return count > 0, err
}
