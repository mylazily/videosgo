package database

import (
	"context"
	"fmt"
	"log"

	"github.com/mylazily/videosgo/internal/config"
	"github.com/redis/go-redis/v9"
)

// RDB 全局 Redis 客户端
var RDB *redis.Client

// InitRedis 初始化 Redis 连接
func InitRedis(cfg *config.RedisConfig) error {
	RDB = redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5)
	defer cancel()

	// 测试连接
	_, err := RDB.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("Redis 连接失败: %w", err)
	}

	log.Println("[数据库] Redis 连接成功")
	return nil
}

// CloseRedis 关闭 Redis 连接
func CloseRedis() error {
	if RDB != nil {
		return RDB.Close()
	}
	return nil
}
