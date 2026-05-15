// Package main 程序入口
package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mylazily/videosgo/internal/collector"
	"github.com/mylazily/videosgo/internal/config"
	"github.com/mylazily/videosgo/internal/database"
	"github.com/mylazily/videosgo/internal/handler"
	"github.com/mylazily/videosgo/internal/repository"
	"github.com/mylazily/videosgo/internal/router"
	"github.com/mylazily/videosgo/internal/service"
	jwtpkg "github.com/mylazily/videosgo/pkg/jwt"
)

func main() {
	// 1. 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("[启动] 加载配置失败: %v", err)
	}
	log.Printf("[启动] 环境: %s", cfg.App.Env)

	// 2. 初始化数据库
	if err := database.InitPostgres(&cfg.Database); err != nil {
		log.Fatalf("[启动] PostgreSQL 初始化失败: %v", err)
	}

	// 3. 初始化 Redis
	if err := database.InitRedis(&cfg.Redis); err != nil {
		log.Printf("[启动] Redis 初始化失败（将降级运行）: %v", err)
	}

	// 4. 初始化 JWT 管理器
	jwtMgr := jwtpkg.NewJWTManager(cfg.JWT.Secret, cfg.JWT.ExpireHours)

	// 5. 初始化 Repository 层
	collectRepo := repository.NewCollectRepo(database.DB)
	commentRepo := repository.NewCommentRepo(database.DB)
	danmakuRepo := repository.NewDanmakuRepo(database.DB)
	rankRepo := repository.NewRankRepo(database.DB)
	userRepo := repository.NewUserRepo(database.DB)
	videoRepo := repository.NewVideoRepo(database.DB)

	// 6. 初始化采集器
	probeTimeout, _ := time.ParseDuration(cfg.Collector.ProbeTimeout)
	if probeTimeout == 0 {
		probeTimeout = 2 * time.Second
	}

	maccmsClient := collector.NewMacCMSClient(30 * time.Second)
	parser := collector.NewParser()
	probe := collector.NewProbe(probeTimeout, cfg.Collector.Workers)
	worker := collector.NewWorker(maccmsClient, parser, probe, database.DB, cfg.Collector.Workers, cfg.Collector.RetryMax)

	// 7. 初始化 Service 层
	collectSvc := service.NewCollectService(collectRepo, videoRepo, worker)
	commentSvc := service.NewCommentService(commentRepo)
	danmakuSvc := service.NewDanmakuService(danmakuRepo)
	rankSvc := service.NewRankService(rankRepo)
	userSvc := service.NewUserService(userRepo, jwtMgr)
	videoSvc := service.NewVideoService(videoRepo)

	// 8. 初始化 Handler 层
	healthHandler := handler.NewHealthHandler()
	videoHandler := handler.NewVideoHandler(videoSvc)
	userHandler := handler.NewUserHandler(userSvc)
	commentHandler := handler.NewCommentHandler(commentSvc)
	danmakuHandler := handler.NewDanmakuHandler(danmakuSvc)
	rankHandler := handler.NewRankHandler(rankSvc)
	collectHandler := handler.NewCollectHandler(collectSvc)

	// 9. 初始化路由
	r := router.Setup(cfg, jwtMgr, healthHandler, videoHandler, userHandler,
		commentHandler, danmakuHandler, rankHandler, collectHandler)

	// 10. 启动采集调度器
	scheduler := collector.NewScheduler(collectRepo, worker)
	if err := scheduler.Start(); err != nil {
		log.Printf("[启动] 采集调度器启动失败: %v", err)
	} else {
		defer scheduler.Stop()
		log.Println("[启动] 采集调度器已启动")
	}

	// 11. 启动 HTTP 服务
	addr := ":" + cfg.App.Port
	log.Printf("[启动] 服务监听 %s", addr)

	go func() {
		if err := r.Run(addr); err != nil {
			log.Fatalf("[启动] 服务启动失败: %v", err)
		}
	}()

	// 12. 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("[关闭] 正在关闭服务...")
	// 等待清理完成
	time.Sleep(2 * time.Second)
	log.Println("[关闭] 服务已关闭")
}
