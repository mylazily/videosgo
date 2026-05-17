package repository

import (
	"database/sql"
)

// TGRepository Telegram 仓库
type TGRepository struct {
	db *sql.DB
}

// NewTGRepository 创建 TG 仓库
func NewTGRepository(db *sql.DB) *TGRepository {
	return &TGRepository{db: db}
}
