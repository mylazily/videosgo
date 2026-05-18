package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	ws "github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"videosgo/internal/config"
	"videosgo/internal/database"
	"videosgo/internal/handler"
	"videosgo/internal/middleware"
	"videosgo/internal/repository"
	"videosgo/internal/service"
	"videosgo/pkg/logger"
)

// Version 版本号
const Version = "1.0.0"

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	// 初始化日志
	if err := logger.Init(); err != nil {
		fmt.Printf("Failed to init logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		logger.Error("Failed to load config", zap.Error(err))
		os.Exit(1)
	}

	logger.Info("Starting videosgo server",
		zap.String("version", Version),
		zap.String("mode", cfg.App.Mode),
	)

	// 初始化数据库
	db, err := database.NewPostgres(cfg.Database)
	if err != nil {
		logger.Error("Failed to connect to database", zap.Error(err))
		os.Exit(1)
	}
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	logger.Info("Connected to PostgreSQL")

	// 初始化 Redis（可选）
	var redisClient *redis.Client
	redisClient, err = database.NewRedis(cfg.Redis)
	if err != nil {
		logger.Warn("Failed to connect to Redis, running in degraded mode", zap.Error(err))
		redisClient = nil
	} else {
		defer redisClient.Close()
		logger.Info("Connected to Redis")
	}

	// ========== 初始化 Repository ==========
	userRepo := repository.NewUserRepository(db)
	videoRepo := repository.NewVideoRepository(db)
	commentRepo := repository.NewCommentRepo(db)
	danmakuRepo := repository.NewDanmakuRepo(db)
	tagRepo := repository.NewTagRepo(db)
	collectRepo := repository.NewCollectRepo(db)
	rankRepo := repository.NewRankRepo(db)

	// ========== 初始化 Service ==========
	authService := service.NewAuthService(userRepo)
	userService := service.NewUserService(userRepo)
	videoService := service.NewVideoService(videoRepo)
	commentService := service.NewCommentService(commentRepo)
	danmakuService := service.NewDanmakuService(danmakuRepo)
	tagService := service.NewTagService(tagRepo)
	collectService := service.NewCollectService(collectRepo, videoRepo, nil)
	rankService := service.NewRankService(rankRepo)

	// P2P Service（需要 logger）
	var p2pService *service.P2PService
	wsService := service.NewWSService(nil, redisClient, nil)

	// ========== 初始化 Handler ==========
	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)
	videoHandler := handler.NewVideoHandler(videoService, commentService, danmakuService, tagService)
	healthHandler := handler.NewHealthHandler(db, redisClient, Version)
	wsHandler := handler.NewWSHandler(wsService)
	p2pHandler := handler.NewP2PHandler(p2pService)
	deviceHandler := handler.NewDeviceHandler(nil)
	rankHandler := handler.NewRankHandler(rankService)
	collectHandler := handler.NewCollectHandler(collectService)
	commentHandler := handler.NewCommentHandler(commentService)
	danmakuHandler := handler.NewDanmakuHandler(danmakuService)
	tagHandler := handler.NewTagHandler(tagService)

	// ========== 设置 Gin 模式 ==========
	if cfg.App.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// ========== 创建 Gin 引擎 ==========
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(logger.GinLogger())

	// ========== CORS 配置 ==========
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Fingerprint-ID"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// ========== 根路径和健康检查 ==========
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"name":    "videosgo API",
			"version": Version,
			"status":  "running",
		})
	})

	r.GET("/health", func(c *gin.Context) {
		health := gin.H{
			"status":  "healthy",
			"service": "videosgo",
			"version": Version,
		}

		if db != nil {
			if err := db.Raw("SELECT 1").Error; err != nil {
				health["database"] = "error"
				health["status"] = "degraded"
			} else {
				health["database"] = "ok"
			}
		}

		if redisClient != nil {
			if err := redisClient.Ping(context.Background()).Err(); err != nil {
				health["redis"] = "error"
				health["status"] = "degraded"
			} else {
				health["redis"] = "ok"
			}
		}

		status := http.StatusOK
		if health["status"] != "healthy" {
			status = http.StatusServiceUnavailable
		}
		c.JSON(status, health)
	})

	// ========== API v1 路由 ==========
	api := r.Group("/api/v1")
	{
		// 认证路由（公开）
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
		}

		// 视频路由（公开）
		videos := api.Group("/videos")
		{
			videos.GET("", videoHandler.ListVideos)
			videos.GET("/:id", videoHandler.GetVideo)
			videos.GET("/:id/episodes/:ep_id/danmaku", danmakuHandler.GetDanmakus)
			videos.GET("/:id/danmaku", danmakuHandler.GetVideoDanmakus)
			videos.GET("/:id/tags", tagHandler.GetVideoTags)
		}

		// 评论路由（公开列表，需要认证创建）
		comments := api.Group("/comments")
		{
			comments.GET("/:id/replies", commentHandler.ListReplies)
		}

		// 排行榜路由（公开）
		rank := api.Group("/rank")
		{
			rank.GET("/daily", rankHandler.GetDailyRank)
			rank.GET("/weekly", rankHandler.GetWeeklyRank)
			rank.GET("/monthly", rankHandler.GetMonthlyRank)
			rank.GET("/category/:category", rankHandler.GetCategoryRank)
		}

		// 标签路由（公开）
		tags := api.Group("/tags")
		{
			tags.GET("", tagHandler.ListTags)
			tags.GET("/:slug", tagHandler.GetTagBySlug)
			tags.GET("/:slug/videos", tagHandler.GetTagVideos)
		}

		// WebSocket 路由（需要认证）
		wsGroup := api.Group("/ws")
		wsGroup.Use(middleware.Auth(cfg.JWT.Secret))
		{
			wsGroup.GET("/danmaku/:video_id", wsHandler.HandleDanmaku)
			wsGroup.GET("/online/:video_id", wsHandler.GetOnlineCount)
		}

		// 需要认证的路由
		authRequired := api.Group("")
		authRequired.Use(middleware.Auth(cfg.JWT.Secret))
		{
			// 用户
			authRequired.GET("/user/profile", userHandler.GetProfile)
			authRequired.PUT("/user/profile", userHandler.UpdateProfile)

			// 带认证的视频操作
			authRequired.POST("/videos/:id/comments", commentHandler.CreateComment)
			authRequired.POST("/videos/:id/episodes/:ep_id/danmaku", danmakuHandler.CreateDanmaku)

			// 评论操作
			authRequired.DELETE("/comments/:id", commentHandler.DeleteComment)
			authRequired.POST("/comments/:id/like", commentHandler.LikeComment)
			authRequired.DELETE("/comments/:id/like", commentHandler.UnlikeComment)

			// 设备验证
			authRequired.GET("/device/check/:video_id", deviceHandler.CheckVideoUnlocked)

			// P2P
			authRequired.POST("/p2p/register", p2pHandler.RegisterPeer)
			authRequired.POST("/p2p/signal", p2pHandler.HandleSignal)
		}

		// 健康检查（带详细状态）
		api.GET("/health", healthHandler.Health)

		// 管理员路由（需要管理员权限）
		admin := api.Group("/admin")
		admin.Use(middleware.Auth(cfg.JWT.Secret))
		admin.Use(middleware.RequireAdmin())
		{
			// 采集源管理
			admin.POST("/collect/sources", collectHandler.CreateSource)
			admin.GET("/collect/sources", collectHandler.ListSources)
			admin.GET("/collect/sources/:id", collectHandler.GetSource)
			admin.PUT("/collect/sources/:id", collectHandler.UpdateSource)
			admin.DELETE("/collect/sources/:id", collectHandler.DeleteSource)
			admin.POST("/collect/sources/:id/trigger", collectHandler.TriggerCollect)
			admin.GET("/collect/logs", collectHandler.ListLogs)

			// 管理员视频操作
			admin.POST("/videos", videoHandler.CreateVideo)
			admin.PUT("/videos/:id", videoHandler.UpdateVideo)
			admin.DELETE("/videos/:id", videoHandler.DeleteVideo)
		}
	}

	// ========== 启动服务器 ==========
	addr := fmt.Sprintf("%s:%d", cfg.App.Host, cfg.App.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	// 优雅关闭
	go func() {
		logger.Info("Server starting", zap.String("addr", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed", zap.Error(err))
			os.Exit(1)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// 关闭超时
	timeout := 5 * time.Second
	if cfg.App.ShutdownTimeout > 0 {
		timeout = cfg.App.ShutdownTimeout
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server stopped")
}

// RequireAdmin 管理员权限中间件
func init() {
	// 注册自定义中间件
}
