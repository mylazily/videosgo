package repository

import (
	"database/sql"
	"videosgo/internal/model"
)

// VideoRepository 视频仓库
type VideoRepository struct {
	db *sql.DB
}

// NewVideoRepository 创建视频仓库
func NewVideoRepository(db *sql.DB) *VideoRepository {
	return &VideoRepository{db: db}
}

// List 获取视频列表
func (r *VideoRepository) List(offset, limit int) ([]model.Video, error) {
	query := `
		SELECT id, title, description, cover_url, video_url, source_id, category, status, created_at, updated_at
		FROM videos WHERE status = 1 ORDER BY created_at DESC LIMIT $1 OFFSET $2
	`
	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var videos []model.Video
	for rows.Next() {
		var v model.Video
		err := rows.Scan(
			&v.ID, &v.Title, &v.Description, &v.CoverURL, &v.VideoURL,
			&v.SourceID, &v.Category, &v.Status, &v.CreatedAt, &v.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		videos = append(videos, v)
	}
	return videos, nil
}

// GetByID 根据 ID 获取视频
func (r *VideoRepository) GetByID(id string) (*model.Video, error) {
	query := `
		SELECT id, title, description, cover_url, video_url, source_id, category, status, created_at, updated_at
		FROM videos WHERE id = $1
	`
	video := &model.Video{}
	err := r.db.QueryRow(query, id).Scan(
		&video.ID, &video.Title, &video.Description, &video.CoverURL, &video.VideoURL,
		&video.SourceID, &video.Category, &video.Status, &video.CreatedAt, &video.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return video, err
}
