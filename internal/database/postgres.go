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

// ClosePostgres 关闭 PostgreSQL 连接
func ClosePostgres() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
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
		// 标签
		&model.Tag{},
		&model.VideoTag{},
		// 短视频
		&model.ShortVideo{},
		// 设备指纹
		&model.DeviceFingerprint{},
		&model.DeviceUnlockRecord{},
		&model.DeviceCoinBalance{},
		// 分享裂变
		&model.ShareLink{},
		&model.ShareClick{},
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
		// 站群管理
		&model.SiteDomain{},
		&model.SiteHealthLog{},
		&model.DomainLinkAudit{},
		// P2P 信令
		&model.SignalChannel{},
		&model.PeerRegistry{},
		&model.TransferLog{},
		// Push 推送
		&model.PushSubscription{},
		&model.PushNotification{},
		&model.PushClickLog{},
		// 301 重定向
		&model.RedirectRule{},
		&model.RedirectHitLog{},
		// TG Bot
		&model.TGBotConfig{},
		&model.TGChannel{},
		&model.TGBroadcastLog{},
		&model.TGMiniAppSession{},
		// X.com 自动发布
		&model.XAccount{},
		&model.XPostLog{},
		&model.XPostQueue{},
		// 支付网关
		&model.PaymentChannel{},
		&model.PaymentOrder{},
		&model.VIPSubscription{},
		// 域名轮询
		&model.DomainAvailability{},
		&model.DomainSwitchEvent{},
		&model.ActiveDomain{},
		// 广告金币系统
		&model.AdTask{},
		&model.CoinTransaction{},
		&model.DailyTaskCompletion{},
	)
}
