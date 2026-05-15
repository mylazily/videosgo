// Package database 数据库连接管理
package database

import (
	"fmt"
	"log"
	"time"

	"github.com/mylazily/videosgo/internal/config"
	"github.com/mylazily/videosgo/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// DB 全局数据库实例
var DB *gorm.DB

// InitPostgres 初始化 PostgreSQL 连接
func InitPostgres(cfg *config.DatabaseConfig) error {
	var err error

	gormCfg := &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Info),
	}

	DB, err = gorm.Open(postgres.Open(cfg.DSN()), gormCfg)
	if err != nil {
		return fmt.Errorf("PostgreSQL 连接失败: %w", err)
	}

	// 获取底层 SQL DB 以配置连接池
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("获取底层 DB 失败: %w", err)
	}

	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// 自动迁移
	if err := autoMigrate(); err != nil {
		return fmt.Errorf("数据库迁移失败: %w", err)
	}

	log.Println("[数据库] PostgreSQL 连接成功")
	return nil
}

// autoMigrate 自动迁移数据库表结构
func autoMigrate() error {
	return DB.AutoMigrate(
		// 用户
		&model.User{},
		// 视频
		&model.Video{},
		&model.VideoSource{},
		&model.Episode{},
		&model.EpisodeSource{},
		&model.UserWatchHistory{},
		&model.SearchHot{},
		// 评论
		&model.Comment{},
		&model.CommentLike{},
		// 弹幕
		&model.Danmaku{},
		// 排行榜
		&model.Rank{},
		// 采集
		&model.CollectSource{},
		&model.CollectLog{},
	)
}
