package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"videosgo/internal/config"
	"videosgo/internal/database"
	"videosgo/internal/handler"
	"videosgo/internal/logger"
	"videosgo/internal/middleware"
	"videosgo/internal/repository"
	"videosgo/internal/service"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// 版本信息 - 由 GitHub Actions 在构建时注入
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func main() {
	// 解析命令行参数
	versionFlag := flag.Bool("version", false, "显示版本信息")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("VideosGo %s (built: %s, commit: %s)\n", Version, BuildTime, GitCommit)
		os.Exit(0)
	}

	// 初始化日志
	log := logger.NewLogger()
	defer log.Sync()

	log.Info("Starting VideosGo",
		zap.String("version", Version),
		zap.String("buildTime", BuildTime),
		zap.String("gitCommit", GitCommit),
	)

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config", zap.Error(err))
	}

	log.Info("Config loaded",
		zap.String("env", cfg.App.Env),
		zap.String("port", cfg.App.Port),
	)

	// 设置 Gin 模式
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 初始化数据库
	db, err := database.NewPostgres(cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Error("Failed to close database", zap.Error(err))
		}
	}()

	// 初始化 Redis（可选）
	var redisClient *redis.Client
	redisClient, err = database.NewRedis(cfg.Redis)
	if err != nil {
		log.Warn("Failed to connect to Redis, running in degraded mode", zap.Error(err))
		redisClient = nil
	} else {
		defer redisClient.Close()
	}

	// 初始化仓库
	userRepo := repository.NewUserRepository(db)
	videoRepo := repository.NewVideoRepository(db)

	// 初始化服务
	authService := service.NewAuthService(cfg, userRepo)
	userService := service.NewUserService(userRepo)
	videoService := service.NewVideoService(videoRepo)

	// 初始化处理器
	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)
	videoHandler := handler.NewVideoHandler(videoService)
	healthHandler := handler.NewHealthHandler(db, redisClient, Version)

	// 创建路由
	r := gin.New()

	// 全局中间件
	r.Use(gin.Recovery())
	r.Use(middleware.Logger(log))
	r.Use(middleware.ErrorHandler())
	r.Use(cors.New(corsConfig(cfg)))

	// 健康检查（公开）
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"version":   Version,
			"buildTime": BuildTime,
			"gitCommit": GitCommit,
			"timestamp": time.Now().UTC(),
		})
	})
	r.GET("/api/v1/health", healthHandler.Health)
	r.GET("/api/v1/ping", healthHandler.Ping)
	
	// 版本信息端点
	r.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"version":   Version,
			"buildTime": BuildTime,
			"gitCommit": GitCommit,
		})
	})

	// API 路由组
	api := r.Group("/api/v1")
	{
		// 公开路由 - 认证
		api.POST("/auth/register", authHandler.Register)
		api.POST("/auth/login", authHandler.Login)
		api.POST("/auth/refresh", authHandler.RefreshToken)

		// 公开路由 - 视频
		api.GET("/videos", videoHandler.ListVideos)
		api.GET("/videos/:id", videoHandler.GetVideo)

		// 需要认证的路由
		authorized := api.Group("")
		authorized.Use(middleware.Auth(cfg.JWT.Secret))
		{
			authorized.GET("/user/profile", userHandler.GetProfile)
			authorized.PUT("/user/profile", userHandler.UpdateProfile)
			authorized.GET("/user/videos", videoHandler.GetUserVideos)
			authorized.POST("/videos/:id/favorite", videoHandler.FavoriteVideo)
		}
	}

	// 根路径
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service":   "videosgo",
			"version":   Version,
			"status":    "running",
			"buildTime": BuildTime,
		})
	})

	// 创建 HTTP 服务器
	srv := &http.Server{
		Addr:         ":" + cfg.App.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 优雅关闭
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	log.Info("Server started",
		zap.String("port", cfg.App.Port),
		zap.String("health", fmt.Sprintf("http://localhost:%s/health", cfg.App.Port)),
	)

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	// 使用配置的超时时间或默认值
	shutdownTimeout := 5 * time.Second
	if cfg.App.ShutdownTimeout > 0 {
		shutdownTimeout = cfg.App.ShutdownTimeout
	}

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Server forced to shutdown", zap.Error(err))
	}

	log.Info("Server exited")
}

func corsConfig(cfg *config.Config) cors.Config {
	c := cors.DefaultConfig()

	// 允许的来源
	if len(cfg.Security.CORSOrigins) == 0 {
		c.AllowOrigins = []string{
			"https://901.555554.xyz",
			"https://shipinku.pages.dev",
			"https://*.pages.dev",
			"http://localhost:3000",
			"http://localhost:5173",
			"http://localhost:8080",
		}
	} else {
		c.AllowOrigins = cfg.Security.CORSOrigins
	}

	c.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	c.AllowHeaders = []string{
		"Origin",
		"Content-Length",
		"Content-Type",
		"Authorization",
		"X-Requested-With",
		"Accept",
		"X-Real-IP",
		"X-Forwarded-For",
	}
	c.ExposeHeaders = []string{"Content-Length", "Content-Type", "X-Request-ID"}
	c.AllowCredentials = true
	c.MaxAge = 12 * time.Hour

	return c
}
