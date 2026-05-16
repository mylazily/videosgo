package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mylazily/videosgo/internal/config"
	"github.com/mylazily/videosgo/internal/database"
	"github.com/mylazily/videosgo/internal/model"
	"github.com/mylazily/videosgo/internal/repository"
)

// TGService TG Bot 服务
type TGService struct {
	repo       *repository.TGRepo
	cfg        *config.TGConfig
	collectSvc *CollectService
}

// NewTGService 创建 TG Bot 服务
func NewTGService(repo *repository.TGRepo, cfg *config.TGConfig, collectSvc *CollectService) *TGService {
	return &TGService{
		repo:       repo,
		cfg:        cfg,
		collectSvc: collectSvc,
	}
}

// getBotToken 获取 Bot Token，优先使用配置，回退到数据库
func (s *TGService) getBotToken() string {
	if s.cfg != nil && s.cfg.BotToken != "" {
		return s.cfg.BotToken
	}
	// 回退到数据库
	config, err := s.repo.GetBotConfig()
	if err != nil {
		log.Printf("[TG] 获取 Bot 配置失败: %v", err)
		return ""
	}
	return config.BotToken
}

// isAdmin 检查 TG 用户 ID 是否为管理员
func (s *TGService) isAdmin(tgUserID int64) bool {
	// 如果 admin_user_ids 为空，所有用户都是管理员（方便初始设置）
	if s.cfg == nil || s.cfg.AdminUserIDs == "" {
		return true
	}
	ids := strings.Split(s.cfg.AdminUserIDs, ",")
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		parsed, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			continue
		}
		if parsed == tgUserID {
			return true
		}
	}
	return false
}

// sendReply 发送回复消息到聊天
func (s *TGService) sendReply(chatID int64, text string, parseMode string) error {
	token := s.getBotToken()
	if token == "" {
		return fmt.Errorf("Bot Token 未配置")
	}

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
	payload := map[string]interface{}{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": parseMode,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("序列化请求失败: %w", err)
	}

	resp, err := http.Post(apiURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("发送消息失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	if ok, _ := result["ok"].(bool); !ok {
		desc, _ := result["description"].(string)
		return fmt.Errorf("TG API 错误: %s", desc)
	}

	return nil
}

// sendReplyHTML 发送 HTML 格式的回复消息
func (s *TGService) sendReplyHTML(chatID int64, text string) error {
	return s.sendReply(chatID, text, "HTML")
}

// ProcessUpdate 处理 Telegram Webhook 更新
func (s *TGService) ProcessUpdate(update map[string]interface{}) {
	// 提取消息
	message, ok := update["message"].(map[string]interface{})
	if !ok {
		return
	}

	// 提取聊天 ID
	chat, ok := message["chat"].(map[string]interface{})
	if !ok {
		return
	}
	chatID, ok := chat["id"].(float64)
	if !ok {
		return
	}

	// 提取发送者信息
	from, _ := message["from"].(map[string]interface{})

	// 提取文本
	text, _ := message["text"].(string)
	if text == "" {
		return
	}

	// 解析命令
	text = strings.TrimSpace(text)
	if !strings.HasPrefix(text, "/") {
		return // 忽略非命令消息
	}

	// 分离命令和参数
	parts := strings.SplitN(text, " ", 2)
	cmd := strings.ToLower(parts[0])
	args := ""
	if len(parts) > 1 {
		args = strings.TrimSpace(parts[1])
	}

	// 去除 @botname 后缀（例如 /start@mybot）
	if idx := strings.Index(cmd, "@"); idx > 0 {
		cmd = cmd[:idx]
	}

	log.Printf("[TG] 收到命令: %s from chat=%d user=%v", cmd, int64(chatID), from)

	// 路由命令
	switch cmd {
	case "/start":
		s.handleStart(int64(chatID), from)
	case "/help":
		s.handleHelp(int64(chatID))
	case "/addsource":
		s.handleAddSource(int64(chatID), args)
	case "/sources", "/listsources":
		s.handleListSources(int64(chatID))
	case "/delsource":
		s.handleDelSource(int64(chatID), args)
	case "/toggle":
		s.handleToggleSource(int64(chatID), args)
	case "/collect":
		s.handleTriggerCollect(int64(chatID), args)
	case "/logs":
		s.handleCollectLogs(int64(chatID), args)
	case "/stats":
		s.handleStats(int64(chatID))
	case "/broadcast":
		s.handleBroadcast(int64(chatID), args)
	case "/domains":
		s.handleDomains(int64(chatID))
	case "/health":
		s.handleHealth(int64(chatID))
	default:
		s.sendReplyHTML(int64(chatID), "未知命令，请发送 /help 查看可用命令列表。")
	}
}

// handleStart 处理 /start 命令
func (s *TGService) handleStart(chatID int64, from map[string]interface{}) {
	username := ""
	if from != nil {
		if name, ok := from["username"].(string); ok {
			username = name
		}
	}

	// 如果 admin_user_ids 为空，提示设置
	adminHint := ""
	if s.cfg != nil && s.cfg.AdminUserIDs == "" {
		adminHint = "\n\n⚠️ <b>注意：</b>TG_ADMIN_USER_IDS 未配置，当前所有用户都有管理员权限。\n请在 .env 中设置 TG_ADMIN_USER_IDS 以限制管理员访问。"
		if from != nil {
			if userID, ok := from["id"].(float64); ok {
				adminHint += fmt.Sprintf("\n你的 TG User ID: <code>%d</code>", int64(userID))
			}
		}
	}

	text := fmt.Sprintf("👋 你好%s！\n\n我是 <b>VideoSGO 管理机器人</b>\n用于管理视频采集系统。\n\n发送 /help 查看所有可用命令。%s",
		func() string {
			if username != "" {
				return ", @" + username
			}
			return ""
		}(),
		adminHint,
	)

	_ = s.sendReplyHTML(chatID, text)
}

// handleHelp 处理 /help 命令
func (s *TGService) handleHelp(chatID int64) {
	text := `<b>📋 命令列表</b>

<b>基础命令：</b>
/start - 欢迎信息
/help - 显示此帮助

<b>采集源管理：</b>
/addsource &lt;名称&gt; &lt;API地址&gt; [API密钥] - 添加采集源
/sources - 列出所有采集源
/delsource &lt;ID前8位&gt; - 删除采集源
/toggle &lt;ID前8位&gt; - 启用/禁用采集源
/collect &lt;ID前8位&gt; [full|incremental] - 触发采集
/logs [页码] - 查看采集日志

<b>系统管理：</b>
/stats - 系统统计信息
/broadcast &lt;消息&gt; - 广播消息到所有频道
/domains - 域名轮换状态
/health - 系统健康检查`

	_ = s.sendReplyHTML(chatID, text)
}

// handleAddSource 处理 /addsource 命令
func (s *TGService) handleAddSource(chatID int64, args string) {
	if !s.isAdmin(chatID) {
		_ = s.sendReplyHTML(chatID, "⛔ 权限不足，仅管理员可执行此操作。")
		return
	}

	parts := strings.Fields(args)
	if len(parts) < 2 {
		_ = s.sendReplyHTML(chatID, "用法: /addsource <名称> <API地址> [API密钥]\n\n示例:\n/addsource mysite https://api.example.com/api.php mykey123")
		return
	}

	name := parts[0]
	apiURL := parts[1]
	apiKey := ""
	if len(parts) >= 3 {
		apiKey = parts[2]
	}

	source := &model.CollectSource{
		Name:       name,
		APIURL:     apiURL,
		APIKey:     apiKey,
		Interval:   60,
		MaxPages:   10,
		Timeout:    30,
		RetryCount: 3,
		Enabled:    true,
		Status:     "active",
	}

	if err := s.collectSvc.CreateSource(source); err != nil {
		_ = s.sendReplyHTML(chatID, fmt.Sprintf("❌ 添加采集源失败: %s", err.Error()))
		return
	}

	shortID := source.ID.String()[:8]
	_ = s.sendReplyHTML(chatID, fmt.Sprintf("✅ 采集源添加成功！\n\n<b>名称:</b> %s\n<b>API:</b> %s\n<b>ID:</b> <code>%s</code>\n<b>状态:</b> 已启用", name, apiURL, shortID))
}

// handleListSources 处理 /sources 命令
func (s *TGService) handleListSources(chatID int64) {
	if !s.isAdmin(chatID) {
		_ = s.sendReplyHTML(chatID, "⛔ 权限不足，仅管理员可执行此操作。")
		return
	}

	sources, total, err := s.collectSvc.ListSources(1, 50)
	if err != nil {
		_ = s.sendReplyHTML(chatID, fmt.Sprintf("❌ 获取采集源列表失败: %s", err.Error()))
		return
	}

	if total == 0 {
		_ = s.sendReplyHTML(chatID, "暂无采集源。使用 /addsource 添加。")
		return
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<b>📡 采集源列表</b> (共 %d 个)\n\n", total))

	for i, src := range sources {
		shortID := src.ID.String()[:8]
		status := "🟢 启用"
		if !src.Enabled {
			status = "🔴 禁用"
		}

		lastCollect := "从未"
		if src.LastCollect != nil {
			lastCollect = src.LastCollect.Format("01-02 15:04")
		}

		sb.WriteString(fmt.Sprintf("<b>%d.</b> <code>%s</code> %s\n", i+1, shortID, status))
		sb.WriteString(fmt.Sprintf("   名称: %s\n", src.Name))
		sb.WriteString(fmt.Sprintf("   API: %s\n", truncateStr(src.APIURL, 50)))
		sb.WriteString(fmt.Sprintf("   上次采集: %s | 已采集: %d\n\n", lastCollect, src.TotalCollected))
	}

	_ = s.sendReplyHTML(chatID, sb.String())
}

// handleDelSource 处理 /delsource 命令
func (s *TGService) handleDelSource(chatID int64, args string) {
	if !s.isAdmin(chatID) {
		_ = s.sendReplyHTML(chatID, "⛔ 权限不足，仅管理员可执行此操作。")
		return
	}

	args = strings.TrimSpace(args)
	if args == "" {
		_ = s.sendReplyHTML(chatID, "用法: /delsource <ID前8位>\n\n使用 /sources 查看采集源列表获取 ID。")
		return
	}

	// 查找匹配的采集源（通过前8位 ID 匹配）
	sources, _, err := s.collectSvc.ListSources(1, 50)
	if err != nil {
		_ = s.sendReplyHTML(chatID, fmt.Sprintf("❌ 获取采集源列表失败: %s", err.Error()))
		return
	}

	var foundID string
	var foundName string
	for _, src := range sources {
		if strings.HasPrefix(src.ID.String(), args) {
			foundID = src.ID.String()
			foundName = src.Name
			break
		}
	}

	if foundID == "" {
		_ = s.sendReplyHTML(chatID, fmt.Sprintf("❌ 未找到匹配的采集源: %s", args))
		return
	}

	if err := s.collectSvc.DeleteSource(foundID); err != nil {
		_ = s.sendReplyHTML(chatID, fmt.Sprintf("❌ 删除采集源失败: %s", err.Error()))
		return
	}

	_ = s.sendReplyHTML(chatID, fmt.Sprintf("✅ 采集源已删除\n\n<b>名称:</b> %s\n<b>ID:</b> <code>%s</code>", foundName, args))
}

// handleToggleSource 处理 /toggle 命令
func (s *TGService) handleToggleSource(chatID int64, args string) {
	if !s.isAdmin(chatID) {
		_ = s.sendReplyHTML(chatID, "⛔ 权限不足，仅管理员可执行此操作。")
		return
	}

	args = strings.TrimSpace(args)
	if args == "" {
		_ = s.sendReplyHTML(chatID, "用法: /toggle <ID前8位>\n\n使用 /sources 查看采集源列表获取 ID。")
		return
	}

	// 查找匹配的采集源
	sources, _, err := s.collectSvc.ListSources(1, 50)
	if err != nil {
		_ = s.sendReplyHTML(chatID, fmt.Sprintf("❌ 获取采集源列表失败: %s", err.Error()))
		return
	}

	var found *model.CollectSource
	for i := range sources {
		if strings.HasPrefix(sources[i].ID.String(), args) {
			found = &sources[i]
			break
		}
	}

	if found == nil {
		_ = s.sendReplyHTML(chatID, fmt.Sprintf("❌ 未找到匹配的采集源: %s", args))
		return
	}

	// 切换状态
	found.Enabled = !found.Enabled
	if err := s.collectSvc.UpdateSource(found); err != nil {
		_ = s.sendReplyHTML(chatID, fmt.Sprintf("❌ 更新采集源失败: %s", err.Error()))
		return
	}

	newStatus := "已启用 🟢"
	if !found.Enabled {
		newStatus = "已禁用 🔴"
	}

	_ = s.sendReplyHTML(chatID, fmt.Sprintf("✅ 采集源状态已更新\n\n<b>名称:</b> %s\n<b>新状态:</b> %s", found.Name, newStatus))
}

// handleTriggerCollect 处理 /collect 命令
func (s *TGService) handleTriggerCollect(chatID int64, args string) {
	if !s.isAdmin(chatID) {
		_ = s.sendReplyHTML(chatID, "⛔ 权限不足，仅管理员可执行此操作。")
		return
	}

	parts := strings.Fields(args)
	if len(parts) < 1 {
		_ = s.sendReplyHTML(chatID, "用法: /collect <ID前8位> [full|incremental]\n\n示例:\n/collect abc12345 full\n/collect abc12345")
		return
	}

	sourceShortID := parts[0]
	collectType := "incremental"
	if len(parts) >= 2 {
		collectType = strings.ToLower(parts[1])
		if collectType != "full" && collectType != "incremental" {
			collectType = "incremental"
		}
	}

	// 查找匹配的采集源
	sources, _, err := s.collectSvc.ListSources(1, 50)
	if err != nil {
		_ = s.sendReplyHTML(chatID, fmt.Sprintf("❌ 获取采集源列表失败: %s", err.Error()))
		return
	}

	var foundID string
	var foundName string
	for _, src := range sources {
		if strings.HasPrefix(src.ID.String(), sourceShortID) {
			foundID = src.ID.String()
			foundName = src.Name
			break
		}
	}

	if foundID == "" {
		_ = s.sendReplyHTML(chatID, fmt.Sprintf("❌ 未找到匹配的采集源: %s", sourceShortID))
		return
	}

	if err := s.collectSvc.TriggerCollect(foundID, collectType); err != nil {
		_ = s.sendReplyHTML(chatID, fmt.Sprintf("❌ 触发采集失败: %s", err.Error()))
		return
	}

	typeLabel := "增量"
	if collectType == "full" {
		typeLabel = "全量"
	}

	_ = s.sendReplyHTML(chatID, fmt.Sprintf("🚀 采集任务已提交\n\n<b>采集源:</b> %s\n<b>类型:</b> %s采集\n\n采集将在后台异步执行，使用 /logs 查看进度。", foundName, typeLabel))
}

// handleCollectLogs 处理 /logs 命令
func (s *TGService) handleCollectLogs(chatID int64, args string) {
	if !s.isAdmin(chatID) {
		_ = s.sendReplyHTML(chatID, "⛔ 权限不足，仅管理员可执行此操作。")
		return
	}

	page := 1
	args = strings.TrimSpace(args)
	if args != "" {
		if p, err := strconv.Atoi(args); err == nil && p > 0 {
			page = p
		}
	}

	logs, total, err := s.collectSvc.ListLogs(page, 10)
	if err != nil {
		_ = s.sendReplyHTML(chatID, fmt.Sprintf("❌ 获取采集日志失败: %s", err.Error()))
		return
	}

	if total == 0 {
		_ = s.sendReplyHTML(chatID, "暂无采集日志。")
		return
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<b>📝 采集日志</b> (第 %d 页，共 %d 条)\n\n", page, total))

	for _, l := range logs {
		statusIcon := "✅"
		if l.Status == "running" {
			statusIcon = "⏳"
		} else if l.Status == "failed" {
			statusIcon = "❌"
		}

		sb.WriteString(fmt.Sprintf("%s <b>%s</b> - %s\n", statusIcon, l.SourceName, l.Type))
		sb.WriteString(fmt.Sprintf("   总数:%d 新增:%d 更新:%d 失败:%d 耗时:%ds\n", l.TotalCount, l.NewCount, l.UpdateCount, l.FailCount, l.Duration))
		sb.WriteString(fmt.Sprintf("   时间: %s\n", l.CreatedAt.Format("01-02 15:04:05")))
		if l.ErrorMessage != "" {
			sb.WriteString(fmt.Sprintf("   错误: %s\n", truncateStr(l.ErrorMessage, 80)))
		}
		sb.WriteString("\n")
	}

	_ = s.sendReplyHTML(chatID, sb.String())
}

// handleStats 处理 /stats 命令
func (s *TGService) handleStats(chatID int64) {
	if !s.isAdmin(chatID) {
		_ = s.sendReplyHTML(chatID, "⛔ 权限不足，仅管理员可执行此操作。")
		return
	}

	var videoCount int64
	var sourceCount int64
	var tagCount int64

	db := database.DB
	db.Model(&model.Video{}).Count(&videoCount)
	db.Model(&model.CollectSource{}).Count(&sourceCount)
	db.Model(&model.Tag{}).Count(&tagCount)

	// MiniApp 统计
	sessions, watchTime, err := s.GetMiniAppStats()
	miniAppInfo := "N/A"
	if err == nil {
		hours := watchTime / 3600
		minutes := (watchTime % 3600) / 60
		miniAppInfo = fmt.Sprintf("%d 次会话, 观看 %d小时%d分钟", sessions, hours, minutes)
	}

	text := fmt.Sprintf(`<b>📊 系统统计</b>

<b>视频:</b> %d
<b>采集源:</b> %d
<b>标签:</b> %d
<b>MiniApp:</b> %s
<b>运行时间:</b> %s`, videoCount, sourceCount, tagCount, miniAppInfo, getUptime())

	_ = s.sendReplyHTML(chatID, text)
}

// handleBroadcast 处理 /broadcast 命令
func (s *TGService) handleBroadcast(chatID int64, args string) {
	if !s.isAdmin(chatID) {
		_ = s.sendReplyHTML(chatID, "⛔ 权限不足，仅管理员可执行此操作。")
		return
	}

	args = strings.TrimSpace(args)
	if args == "" {
		_ = s.sendReplyHTML(chatID, "用法: /broadcast <消息内容>\n\n将消息广播到所有活跃的 TG 频道。")
		return
	}

	channels, err := s.repo.ListActiveChannels()
	if err != nil {
		_ = s.sendReplyHTML(chatID, fmt.Sprintf("❌ 获取频道列表失败: %s", err.Error()))
		return
	}

	if len(channels) == 0 {
		_ = s.sendReplyHTML(chatID, "⚠️ 没有活跃的 TG 频道。")
		return
	}

	// 异步广播
	go func() {
		successCount := 0
		failCount := 0
		for _, ch := range channels {
			_, err := s.SendMessage(ch.ChannelID, args, "", "")
			if err != nil {
				failCount++
				log.Printf("[TG] 广播到频道 %s 失败: %v", ch.ChannelTitle, err)
			} else {
				successCount++
			}
		}
		resultText := fmt.Sprintf("📢 广播完成\n\n<b>成功:</b> %d 个频道\n<b>失败:</b> %d 个频道", successCount, failCount)
		_ = s.sendReplyHTML(chatID, resultText)
	}()

	_ = s.sendReplyHTML(chatID, fmt.Sprintf("📢 正在向 %d 个频道广播消息...", len(channels)))
}

// handleDomains 处理 /domains 命令
func (s *TGService) handleDomains(chatID int64) {
	if !s.isAdmin(chatID) {
		_ = s.sendReplyHTML(chatID, "⛔ 权限不足，仅管理员可执行此操作。")
		return
	}

	db := database.DB

	var actives []model.ActiveDomain
	if err := db.Find(&actives).Error; err != nil {
		_ = s.sendReplyHTML(chatID, fmt.Sprintf("❌ 获取域名信息失败: %s", err.Error()))
		return
	}

	if len(actives) == 0 {
		_ = s.sendReplyHTML(chatID, "暂无活跃域名。")
		return
	}

	var sb strings.Builder
	sb.WriteString("<b>🌐 域名轮换状态</b>\n\n")

	for _, ad := range actives {
		sb.WriteString(fmt.Sprintf("🟢 <b>%s</b>\n", ad.Domain))
		sb.WriteString(fmt.Sprintf("   区域: %s\n", ad.Region))
		sb.WriteString(fmt.Sprintf("   激活时间: %s\n", ad.ActivatedAt.Format("2006-01-02 15:04:05")))
		if ad.ActivatedBy != "" {
			sb.WriteString(fmt.Sprintf("   操作人: %s\n", ad.ActivatedBy))
		}
		sb.WriteString("\n")
	}

	_ = s.sendReplyHTML(chatID, sb.String())
}

// handleHealth 处理 /health 命令
func (s *TGService) handleHealth(chatID int64) {
	if !s.isAdmin(chatID) {
		_ = s.sendReplyHTML(chatID, "⛔ 权限不足，仅管理员可执行此操作。")
		return
	}

	var sb strings.Builder
	sb.WriteString("<b>🏥 系统健康检查</b>\n\n")

	// 检查数据库
	db := database.DB
	sqlDB, err := db.DB()
	if err != nil {
		sb.WriteString("❌ <b>PostgreSQL:</b> 连接异常\n")
	} else {
		if err := sqlDB.Ping(); err != nil {
			sb.WriteString(fmt.Sprintf("❌ <b>PostgreSQL:</b> %s\n", err.Error()))
		} else {
			stats := sqlDB.Stats()
			sb.WriteString(fmt.Sprintf("✅ <b>PostgreSQL:</b> 正常 (空闲:%d/打开:%d)\n", stats.Idle, stats.OpenConnections))
		}
	}

	// 检查 Redis
	rdb := database.RDB
	if rdb == nil {
		sb.WriteString("⚠️ <b>Redis:</b> 未初始化\n")
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := rdb.Ping(ctx).Err(); err != nil {
			sb.WriteString(fmt.Sprintf("❌ <b>Redis:</b> %s\n", err.Error()))
		} else {
			sb.WriteString("✅ <b>Redis:</b> 正常\n")
		}
	}

	// 检查 Bot Token
	token := s.getBotToken()
	if token == "" {
		sb.WriteString("❌ <b>TG Bot:</b> Token 未配置\n")
	} else {
		sb.WriteString("✅ <b>TG Bot:</b> Token 已配置\n")
	}

	sb.WriteString(fmt.Sprintf("⏱️ <b>运行时间:</b> %s", getUptime()))

	_ = s.sendReplyHTML(chatID, sb.String())
}

// ========== 保留原有方法 ==========

// SendMessage 发送消息到 TG 频道
func (s *TGService) SendMessage(channelID int64, text, mediaURL, linkURL string) (*model.TGBroadcastLog, error) {
	// 获取 Bot 配置
	config, err := s.repo.GetBotConfig()
	if err != nil {
		return nil, fmt.Errorf("获取 Bot 配置失败: %w", err)
	}

	// 创建广播日志
	broadcastLog := &model.TGBroadcastLog{
		ChannelID:   uuid.Nil, // 稍后关联
		MessageText: text,
		MediaURL:    mediaURL,
		LinkURL:     linkURL,
		PostType:    "text",
		Status:      "pending",
	}

	// 构建消息
	message := text
	if linkURL != "" {
		message = fmt.Sprintf("%s\n\n%s", text, linkURL)
	}

	// 调用 TG Bot API
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", config.BotToken)
	payload := map[string]interface{}{
		"chat_id":    channelID,
		"text":       message,
		"parse_mode": "HTML",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	resp, err := http.Post(apiURL, "application/json", bytes.NewReader(body))
	if err != nil {
		broadcastLog.Status = "failed"
		broadcastLog.ErrorMessage = err.Error()
		_ = s.repo.CreateBroadcast(broadcastLog)
		return broadcastLog, fmt.Errorf("发送消息失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		broadcastLog.Status = "failed"
		broadcastLog.ErrorMessage = "解析响应失败"
		_ = s.repo.CreateBroadcast(broadcastLog)
		return broadcastLog, fmt.Errorf("解析响应失败: %w", err)
	}

	// 检查是否成功
	if ok, _ := result["ok"].(bool); !ok {
		desc, _ := result["description"].(string)
		broadcastLog.Status = "failed"
		broadcastLog.ErrorMessage = desc
		_ = s.repo.CreateBroadcast(broadcastLog)
		return broadcastLog, fmt.Errorf("TG API 错误: %s", desc)
	}

	// 提取 message_id
	var messageID int64
	if resultObj, ok := result["result"].(map[string]interface{}); ok {
		if msgID, ok := resultObj["message_id"].(float64); ok {
			messageID = int64(msgID)
		}
	}

	broadcastLog.MessageID = messageID
	broadcastLog.Status = "sent"
	_ = s.repo.CreateBroadcast(broadcastLog)

	return broadcastLog, nil
}

// BroadcastToAllChannels 群发消息到所有活跃频道
func (s *TGService) BroadcastToAllChannels(videoID uuid.UUID, text string) {
	channels, err := s.repo.ListActiveChannels()
	if err != nil {
		log.Printf("[TG] 获取活跃频道失败: %v", err)
		return
	}

	for _, channel := range channels {
		go func(ch model.TGChannel) {
			log.Printf("[TG] 正在发送到频道 %s (%d)...", ch.ChannelTitle, ch.ChannelID)
			_, err := s.SendMessage(ch.ChannelID, text, "", "")
			if err != nil {
				log.Printf("[TG] 发送到频道 %s 失败: %v", ch.ChannelTitle, err)
			} else {
				log.Printf("[TG] 发送到频道 %s 成功", ch.ChannelTitle)
				// 更新频道最后发布时间
				now := time.Now()
				ch.LastPostAt = &now
				_ = s.repo.UpdateChannel(&ch)
			}
		}(channel)
	}
}

// RegisterMiniAppSession 注册 Mini App 会话
func (s *TGService) RegisterMiniAppSession(tgUserID int64, username, lang string, fingerprintID *uuid.UUID) (*model.TGMiniAppSession, error) {
	session := &model.TGMiniAppSession{
		TGUserID:      tgUserID,
		TGUsername:    username,
		TGLanguage:    lang,
		FingerprintID: fingerprintID,
	}
	err := s.repo.UpsertMiniAppSession(session)
	if err != nil {
		return nil, fmt.Errorf("注册 Mini App 会话失败: %w", err)
	}
	return s.repo.GetMiniAppSession(tgUserID)
}

// GetChannels 获取频道列表
func (s *TGService) GetChannels() ([]model.TGChannel, error) {
	return s.repo.ListChannels()
}

// GetBroadcasts 获取广播日志
func (s *TGService) GetBroadcasts(limit int) ([]model.TGBroadcastLog, error) {
	return s.repo.ListBroadcasts(limit)
}

// GetMiniAppStats 获取 Mini App 统计
func (s *TGService) GetMiniAppStats() (totalSessions int64, totalWatchTime int64, err error) {
	return s.repo.GetMiniAppStats()
}

// ========== 工具函数 ==========

// startTime 记录服务启动时间，用于计算运行时间
var startTime = time.Now()

// getUptime 获取服务运行时间
func getUptime() string {
	d := time.Since(startTime)
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	if days > 0 {
		return fmt.Sprintf("%d天%d小时%d分钟", days, hours, minutes)
	}
	if hours > 0 {
		return fmt.Sprintf("%d小时%d分钟", hours, minutes)
	}
	return fmt.Sprintf("%d分钟", minutes)
}

// truncateStr 截断字符串
func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
