// Package main 程序入口
package main

import (
	"context"
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
	tagRepo := repository.NewTagRepo(database.DB)
	shortVideoRepo := repository.NewShortVideoRepo(database.DB)
	deviceRepo := repository.NewDeviceRepo(database.DB)
	shareRepo := repository.NewShareRepo(database.DB)
	siteRepo := repository.NewSiteRepo(database.DB)
	p2pRepo := repository.NewP2PRepo(database.DB)
	pushRepo := repository.NewPushRepo(database.DB)
	redirectRepo := repository.NewRedirectRepo(database.DB)
	tgRepo := repository.NewTGRepo(database.DB)
	xRepo := repository.NewXRepo(database.DB)
	paymentRepo := repository.NewPaymentRepo(database.DB)
	domainRotationRepo := repository.NewDomainRotationRepo(database.DB)
	adRewardRepo := repository.NewAdRewardRepo(database.DB)

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
	tagSvc := service.NewTagService(tagRepo)
	shortVideoSvc := service.NewShortVideoService(shortVideoRepo)
	recommendSvc := service.NewRecommendService(tagRepo, videoRepo)
	deviceSvc := service.NewDeviceService(deviceRepo)
	shareSvc := service.NewShareService(shareRepo, deviceRepo)
	sitemapSvc := service.NewSitemapService(database.DB)
	siteSvc := service.NewSiteService(siteRepo)
	p2pSvc := service.NewP2PService(p2pRepo)
	pushSvc := service.NewPushService(pushRepo)
	redirectSvc := service.NewRedirectService(redirectRepo)
	gscSvc := service.NewGSCService(siteRepo)
	tgSvc := service.NewTGService(tgRepo)
	xSvc := service.NewXService(xRepo)
	paymentSvc := service.NewPaymentService(paymentRepo)
	wsSvc := service.NewWSService()
	domainRotationSvc := service.NewDomainRotationService(domainRotationRepo)
	adRewardSvc := service.NewAdRewardService(adRewardRepo)

	// 8. 初始化资源站监控服务（在 Handler 之前，因为 VideoHandler 需要注入）
	stationMonitor := service.NewStationMonitor(database.RDB)
	// 从数据库加载所有启用的采集源，添加到监控列表
	if enabledSources, err := collectRepo.GetEnabled(); err == nil {
		for _, src := range enabledSources {
			stationMonitor.AddStation(src.Name, src.APIURL, "")
		}
		log.Printf("[启动] 资源站监控已加载 %d 个采集源", len(enabledSources))
	}

	// 9. 初始化 Handler 层
	healthHandler := handler.NewHealthHandler()
	videoHandler := handler.NewVideoHandler(videoSvc)
	videoHandler.SetStationMonitor(stationMonitor) // 注入资源站监控
	userHandler := handler.NewUserHandler(userSvc)
	commentHandler := handler.NewCommentHandler(commentSvc)
	danmakuHandler := handler.NewDanmakuHandler(danmakuSvc)
	rankHandler := handler.NewRankHandler(rankSvc)
	collectHandler := handler.NewCollectHandler(collectSvc)
	tagHandler := handler.NewTagHandler(tagSvc)
	shortVideoHandler := handler.NewShortVideoHandler(shortVideoSvc)
	recommendHandler := handler.NewRecommendHandler(recommendSvc)
	deviceHandler := handler.NewDeviceHandler(deviceSvc)
	shareHandler := handler.NewShareHandler(shareSvc)
	sitemapHandler := handler.NewSitemapHandler(sitemapSvc)
	siteHandler := handler.NewSiteHandler(siteSvc)
	p2pHandler := handler.NewP2PHandler(p2pSvc)
	pushHandler := handler.NewPushHandler(pushSvc)
	redirectHandler := handler.NewRedirectHandler(redirectSvc)
	tgHandler := handler.NewTGHandler(tgSvc)
	xHandler := handler.NewXHandler(xSvc)
	paymentHandler := handler.NewPaymentHandler(paymentSvc)
	wsHandler := handler.NewWSHandler(wsSvc)
	domainHandler := handler.NewDomainHandler(domainRotationSvc)
	adRewardHandler := handler.NewAdRewardHandler(adRewardSvc)
	stationHandler := handler.NewStationHandler(stationMonitor)

	// 10. 初始化路由（传入 UA 分流所需的 301 匹配函数）
	redirectFn := func(domain, path, ua string) (targetURL, ruleType string, ok bool) {
		rule := redirectSvc.MatchRule(domain, path, ua)
		if rule != nil {
			return rule.TargetURL, rule.RuleType, true
		}
		return "", "", false
	}

	r := router.Setup(cfg, jwtMgr, healthHandler, videoHandler, userHandler,
		commentHandler, danmakuHandler, rankHandler, collectHandler,
		tagHandler, shortVideoHandler, recommendHandler, deviceHandler,
		shareHandler, sitemapHandler, siteHandler, p2pHandler,
		pushHandler, redirectHandler, tgHandler, xHandler,
		paymentHandler, wsHandler, domainHandler, adRewardHandler,
		stationHandler,
		redirectFn)

	// 10. 启动采集调度器
	scheduler := collector.NewScheduler(collectRepo, worker)
	if err := scheduler.Start(); err != nil {
		log.Printf("[启动] 采集调度器启动失败: %v", err)
	} else {
		defer scheduler.Stop()
		log.Println("[启动] 采集调度器已启动")
	}

	// 10a. 启动资源站监控（1 分钟间隔）
	stationMonitor.Start(context.Background())
	defer stationMonitor.Stop()
	log.Println("[启动] 资源站监控已启动（每1分钟检查）")

	// 11. 启动定时任务

	// 11a. P2P 信令清理（每 5 分钟）
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			if err := p2pSvc.Cleanup(); err != nil {
				log.Printf("[定时任务] P2P 清理失败: %v", err)
			}
		}
	}()
	log.Println("[启动] P2P 清理定时任务已启动（每5分钟）")

	// 11b. 站群健康检查（每 10 分钟）
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		// 首次启动时延迟 30 秒执行
		time.Sleep(30 * time.Second)
		if err := siteSvc.HealthCheckAll(); err != nil {
			log.Printf("[定时任务] 站群健康检查失败: %v", err)
		}
		for range ticker.C {
			if err := siteSvc.HealthCheckAll(); err != nil {
				log.Printf("[定时任务] 站群健康检查失败: %v", err)
			}
		}
	}()
	log.Println("[启动] 站群健康检查定时任务已启动（每10分钟）")

	// 11c. Sitemap 自动提交（每 4 小时）
	go func() {
		ticker := time.NewTicker(4 * time.Hour)
		defer ticker.Stop()
		// 首次启动时延迟 2 分钟执行
		time.Sleep(2 * time.Minute)
		gscSvc.SubmitAllSitemaps()
		for range ticker.C {
			gscSvc.SubmitAllSitemaps()
		}
	}()
	log.Println("[启动] Sitemap 自动提交定时任务已启动（每4小时）")

	// 11d. Push 推送服务关闭处理
	defer pushSvc.Close()

	// 11e. 域名轮询自动检查（每分钟）
	domainRotationSvc.StartAutoRotation()
	defer domainRotationSvc.StopAutoRotation()
	log.Println("[启动] 域名轮询自动检查已启动（每分钟）")

	// 12. 启动 HTTP 服务
	addr := ":" + cfg.App.Port
	log.Printf("[启动] 服务监听 %s", addr)

	go func() {
		if err := r.Run(addr); err != nil {
			log.Fatalf("[启动] 服务启动失败: %v", err)
		}
	}()

	// 13. 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("[关闭] 正在关闭服务...")
	// 等待清理完成
	time.Sleep(2 * time.Second)
	log.Println("[关闭] 服务已关闭")
}
