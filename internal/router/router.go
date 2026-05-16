// Package router 路由注册
package router

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mylazily/videosgo/internal/config"
	"github.com/mylazily/videosgo/internal/handler"
	"github.com/mylazily/videosgo/internal/middleware"
	jwtpkg "github.com/mylazily/videosgo/pkg/jwt"
)

// Setup 设置路由
func Setup(
	cfg *config.Config,
	jwtMgr *jwtpkg.JWTManager,
	healthH *handler.HealthHandler,
	videoH *handler.VideoHandler,
	userH *handler.UserHandler,
	commentH *handler.CommentHandler,
	danmakuH *handler.DanmakuHandler,
	rankH *handler.RankHandler,
	collectH *handler.CollectHandler,
	tagH *handler.TagHandler,
	shortVideoH *handler.ShortVideoHandler,
	recommendH *handler.RecommendHandler,
	deviceH *handler.DeviceHandler,
	shareH *handler.ShareHandler,
	sitemapH *handler.SitemapHandler,
	// 新增模块
	siteH *handler.SiteHandler,
	p2pH *handler.P2PHandler,
	pushH *handler.PushHandler,
	redirectH *handler.RedirectHandler,
	// 功能扩展模块
	tgH *handler.TGHandler,
	xH *handler.XHandler,
	paymentH *handler.PaymentHandler,
	wsH *handler.WSHandler,
	domainH *handler.DomainHandler,
	adRewardH *handler.AdRewardHandler,
	// 资源站监控
	stationH *handler.StationHandler,
	// UA 分流中间件回调
	redirectFn func(domain, path, ua, ip string) (targetURL string, found bool),
) *gin.Engine {
	r := gin.New()

	// UA/IP 智能分流中间件（注册在最前面）
	r.Use(middleware.UASplit(redirectFn))

	// 全局中间件
	r.Use(middleware.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.Logging())
	r.Use(middleware.CORS(cfg.Security.CORSOrigins()))
		r.Use(middleware.Security(&middleware.SecurityConfig{
			UAFilterEnabled: cfg.Security.UAFilterEnabled,
			WAFEnabled:     cfg.Security.WAFEnabled,
			WhitelistPaths:  cfg.Security.UAWhitelist(),
		}))

	// 限流中间件
	limiter := middleware.NewRateLimiter(cfg.Security.RateLimitPerMinute, time.Minute)
	r.Use(middleware.RateLimit(limiter))

	// XOR 加密中间件（可选）
	if cfg.Crypto.Key != "" {
		r.Use(middleware.Crypto(cfg.Crypto.Key))
	}

	api := r.Group("/api/v1")

	// ========== 公开路由 ==========
	{
		// 健康检查
		api.GET("/ping", healthH.Ping)
		api.GET("/health", healthH.Health)

		// 视频公开接口
		api.GET("/videos", videoH.ListVideos)
		api.GET("/videos/latest", videoH.GetLatest)
		api.GET("/videos/hot", videoH.GetHot)
		api.GET("/videos/random", videoH.GetRandom)
		api.GET("/videos/:id", videoH.GetVideo)
		api.GET("/videos/:id/episodes", videoH.GetEpisodes)
		api.GET("/videos/:id/tags", tagH.GetVideoTags)
		api.GET("/videos/:id/related", recommendH.GetRelatedVideos)
		api.GET("/categories", videoH.GetCategories)
		api.GET("/search", videoH.SearchVideos)
		api.GET("/search/hot", videoH.GetSearchHot)

		// 标签
		api.GET("/tags", tagH.ListTags)
		api.GET("/tags/:slug", tagH.GetTagBySlug)
		api.GET("/tags/:slug/videos", tagH.GetTagVideos)

		// 短视频
		api.GET("/shorts", shortVideoH.ListShorts)
		api.GET("/shorts/random", shortVideoH.GetRandom)
		api.GET("/shorts/:id", shortVideoH.GetShort)
		api.POST("/shorts/:id/view", shortVideoH.IncrementView)
		api.POST("/shorts/:id/like", shortVideoH.IncrementLike)

		// 推荐
		api.GET("/recommendations", recommendH.GetPersonalizedRecommendations)

		// 设备指纹
		api.POST("/device/register", deviceH.RegisterDevice)
		api.GET("/device/profile", deviceH.GetDeviceProfile)
		api.POST("/device/unlock", deviceH.UnlockVideo)
		api.GET("/device/check/:videoId", deviceH.CheckVideoUnlocked)

		// 分享裂变
		api.POST("/share/create", shareH.CreateShareLink)
		api.GET("/share/:code", shareH.GetShareLink)
		api.POST("/share/:code/click", shareH.RecordShareClick)

		// 排行榜
		api.GET("/rank/daily", rankH.GetDailyRank)
		api.GET("/rank/weekly", rankH.GetWeeklyRank)
		api.GET("/rank/monthly", rankH.GetMonthlyRank)
		api.GET("/rank/category/:category", rankH.GetCategoryRank)

		// 视频评论（公开查看）
		api.GET("/videos/:id/comments", commentH.ListComments)
		api.GET("/comments/:id/replies", commentH.ListReplies)

		// 弹幕（公开查看）
		api.GET("/videos/:id/episodes/:ep_id/danmaku", danmakuH.GetDanmakus)
		api.GET("/videos/:id/danmaku", danmakuH.GetVideoDanmakus)

		// 认证
		api.POST("/auth/register", userH.Register)
		api.POST("/auth/login", userH.Login)

		// P2P 信令（公开接口）
		api.POST("/p2p/register", p2pH.RegisterPeer)
		api.POST("/p2p/heartbeat", p2pH.Heartbeat)
		api.DELETE("/p2p/unregister", p2pH.UnregisterPeer)
		api.POST("/p2p/signal/offer", p2pH.OfferSignal)
		api.POST("/p2p/signal/answer", p2pH.AnswerSignal)
		api.POST("/p2p/signal/ice", p2pH.ExchangeICE)
		api.GET("/p2p/peers/:videoId", p2pH.GetVideoPeers)

		// Push 推送（公开接口）
		api.POST("/push/subscribe", pushH.Subscribe)
		api.DELETE("/push/subscribe", pushH.Unsubscribe)
		api.GET("/push/stats", pushH.GetStats)

		// TG Bot（公开接口）
		api.POST("/tg/webhook", tgH.Webhook)
		api.GET("/tg/channels", tgH.ListChannels)
		api.POST("/tg/miniapp/session", tgH.RegisterMiniAppSession)
		api.GET("/tg/miniapp/stats", tgH.GetMiniAppStats)

		// X.com（公开接口）
		api.GET("/x/accounts", xH.ListAccounts)
		api.GET("/x/posts", xH.ListPosts)

		// 支付（公开接口）
		api.GET("/payment/channels", paymentH.ListChannels)
		api.GET("/payment/vip/status", paymentH.GetVIPStatus)
		api.POST("/payment/verify", paymentH.VerifyVIP)

		// 域名轮询（公开接口）
		api.GET("/domain/active", domainH.GetActiveDomain)
		api.GET("/domain/list", domainH.GetDomainList)
		api.GET("/domains/healthy", domainH.GetHealthyDomains)

		// 广告金币（公开接口）
		api.GET("/reward/tasks", adRewardH.ListTasks)
		api.GET("/reward/balance", adRewardH.GetBalance)
		api.GET("/reward/history", adRewardH.GetHistory)
		api.GET("/reward/dashboard", adRewardH.GetDashboard)

		// 资源站监控（公开接口）
		api.GET("/stations/status", stationH.GetStatus)
		api.GET("/stations/best", stationH.GetBest)
		api.GET("/stations/alive", stationH.GetAliveStations)

		// WebSocket 弹幕
		api.GET("/ws/danmaku/:videoId", wsH.HandleDanmaku)
		api.GET("/ws/online/:videoId", wsH.GetOnlineCount)
	}

	// ========== 需要认证的路由 ==========
	auth := api.Group("")
	auth.Use(middleware.Auth(jwtMgr))
	{
		// 用户
		auth.GET("/user/profile", userH.GetProfile)
		auth.PUT("/user/profile", userH.UpdateProfile)
		auth.POST("/user/password", userH.ChangePassword)
		auth.POST("/auth/refresh", userH.RefreshToken)

		// 观看历史
		auth.GET("/user/history", videoH.GetWatchHistory)
		auth.POST("/videos/:id/watch", videoH.RecordWatch)

		// 评论
		auth.POST("/videos/:id/comments", commentH.CreateComment)
		auth.DELETE("/comments/:id", commentH.DeleteComment)
		auth.POST("/comments/:id/like", commentH.LikeComment)
		auth.DELETE("/comments/:id/like", commentH.UnlikeComment)

		// 弹幕
		auth.POST("/videos/:id/episodes/:ep_id/danmaku", danmakuH.CreateDanmaku)
	}

	// ========== 管理员路由 ==========
	admin := api.Group("/admin")
	admin.Use(middleware.Auth(jwtMgr))
	admin.Use(middleware.AdminRequired())
	{
		// 用户管理
		admin.GET("/users", userH.ListUsers)
		admin.DELETE("/users/:id", userH.DeleteUser)

		// 采集源管理
		admin.POST("/collect/sources", collectH.CreateSource)
		admin.PUT("/collect/sources/:id", collectH.UpdateSource)
		admin.DELETE("/collect/sources/:id", collectH.DeleteSource)
		admin.GET("/collect/sources", collectH.ListSources)
		admin.GET("/collect/sources/:id", collectH.GetSource)
		admin.POST("/collect/sources/:id/trigger", collectH.TriggerCollect)
		admin.GET("/collect/logs", collectH.ListLogs)

		// 站群管理
		admin.GET("/sites", siteH.ListSites)
		admin.POST("/sites", siteH.CreateSite)
		admin.PUT("/sites/:id", siteH.UpdateSite)
		admin.DELETE("/sites/:id", siteH.DeleteSite)
		admin.POST("/sites/:id/health-check", siteH.HealthCheck)
		admin.GET("/sites/audit", siteH.GetLinkAudit)

		// 301 重定向规则管理
		admin.GET("/redirects", redirectH.ListRules)
		admin.POST("/redirects", redirectH.CreateRule)
		admin.PUT("/redirects/:id", redirectH.UpdateRule)
		admin.DELETE("/redirects/:id", redirectH.DeleteRule)
		admin.GET("/redirects/:id/logs", redirectH.GetHitLogs)

		// 推送管理
		admin.POST("/push/send", pushH.SendNotification)

		// TG Bot 管理
		admin.POST("/tg/broadcast", tgH.Broadcast)

		// X.com 管理
		admin.POST("/x/post", xH.CreatePost)
		admin.POST("/x/process-queue", xH.ProcessQueue)

		// 支付管理
		admin.POST("/payment/create", paymentH.CreateOrder)
		admin.GET("/payment/:orderNo", paymentH.GetOrder)

		// 域名管理
		admin.POST("/domain/switch", domainH.SwitchDomain)
		admin.GET("/domain/history", domainH.GetSwitchHistory)

		// 广告金币管理
		admin.POST("/reward/complete", adRewardH.CompleteTask)
		admin.POST("/reward/unlock", adRewardH.UnlockVideo)
		admin.POST("/reward/checkin", adRewardH.DailyCheckIn)

		// 资源站监控管理
		admin.POST("/stations/check", stationH.TriggerCheck)
	}

	// ========== SEO 路由 ==========
	{
		r.GET("/sitemap.xml", sitemapH.GetSitemapIndex)
		r.GET("/sitemap-video.xml", sitemapH.GetVideoSitemap)
		r.GET("/sitemap-tag.xml", sitemapH.GetTagSitemap)
		r.GET("/sitemap-short.xml", sitemapH.GetShortVideoSitemap)
		r.GET("/sitemap-actor.xml", sitemapH.GetActorSitemap)
		r.GET("/robots.txt", sitemapH.GetRobotsTxt)
	}

	return r
}
