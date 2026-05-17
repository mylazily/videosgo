package repository

import (
	"database/sql"
	"videosgo/internal/model"
)

// UserRepository 用户仓库
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository 创建用户仓库
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create 创建用户
func (r *UserRepository) Create(user *model.User) error {
	query := `
		INSERT INTO users (id, username, email, password_hash, avatar, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.Exec(query,
		user.ID, user.Username, user.Email, user.Password,
		user.Avatar, user.Status, user.CreatedAt, user.UpdatedAt,
	)
	return err
}

// GetByEmail 根据邮箱获取用户
func (r *UserRepository) GetByEmail(email string) (*model.User, error) {
	query := `
		SELECT id, username, email, password_hash, avatar, status, created_at, updated_at
		FROM users WHERE email = $1
	`
	user := &model.User{}
	err := r.db.QueryRow(query, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.Password,
		&user.Avatar, &user.Status, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return user, err
}

// GetByID 根据 ID 获取用户
func (r *UserRepository) GetByID(id string) (*model.User, error) {
	query := `
		SELECT id, username, email, password_hash, avatar, status, created_at, updated_at
		FROM users WHERE id = $1
	`
	user := &model.User{}
	err := r.db.QueryRow(query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.Password,
		&user.Avatar, &user.Status, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return user, err
}

// Update 更新用户
func (r *UserRepository) Update(user *model.User) error {
	query := `
		UPDATE users SET username = $1, avatar = $2, updated_at = $3
		WHERE id = $4
	`
	_, err := r.db.Exec(query, user.Username, user.Avatar, user.UpdatedAt, user.ID)
	return err
}
