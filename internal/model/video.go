package model

import (
	"time"
)

// Video 视频模型
type Video struct {
	ID          string    `json:"id" db:"id"`
	Title       string    `json:"title" db:"title"`
	Description string    `json:"description" db:"description"`
	CoverURL    string    `json:"cover_url" db:"cover_url"`
	VideoURL    string    `json:"video_url" db:"video_url"`
	SourceID    string    `json:"source_id" db:"source_id"`
	Category    string    `json:"category" db:"category"`
	Status      int       `json:"status" db:"status"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}
