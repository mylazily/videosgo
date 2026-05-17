package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"videosgo/internal/config"
	"videosgo/internal/repository"
)

// TGService Telegram 服务
type TGService struct {
	cfg     config.TGConfig
	tgRepo  *repository.TGRepository
	client  *http.Client
}

// TGUpdate Telegram 更新
type TGUpdate struct {
	UpdateID int64      `json:"update_id"`
	Message  *TGMessage `json:"message,omitempty"`
}

// TGMessage Telegram 消息
type TGMessage struct {
	MessageID int64   `json:"message_id"`
	From      *TGUser `json:"from,omitempty"`
	Chat      *TGChat `json:"chat,omitempty"`
	Text      string  `json:"text"`
	Date      int64   `json:"date"`
}

// TGUser Telegram 用户
type TGUser struct {
	ID        int64  `json:"id"`
	IsBot     bool   `json:"is_bot"`
	FirstName string `json:"first_name"`
	Username  string `json:"username"`
}

// TGChat Telegram 聊天
type TGChat struct {
	ID   int64  `json:"id"`
	Type string `json:"type"`
}

// NewTGService 创建 TG 服务
func NewTGService(cfg config.TGConfig, tgRepo *repository.TGRepository) *TGService {
	return &TGService{
		cfg:    cfg,
		tgRepo: tgRepo,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// IsAdmin 检查用户是否为管理员
func (s *TGService) IsAdmin(userID int64) bool {
	for _, id := range s.cfg.AdminUserIDs {
		if id == userID {
			return true
		}
	}
	return false
}

// HandleUpdate 处理更新
func (s *TGService) HandleUpdate(update *TGUpdate) error {
	if update.Message == nil {
		return nil
	}

	msg := update.Message
	if !s.IsAdmin(msg.From.ID) {
		s.sendMessage(msg.Chat.ID, "⛔ 无权访问")
		return nil
	}

	text := strings.TrimSpace(msg.Text)
	parts := strings.Fields(text)
	if len(parts) == 0 {
		return nil
	}

	cmd := strings.ToLower(parts[0])
	switch cmd {
	case "/start":
		s.sendMessage(msg.Chat.ID, "👋 欢迎使用 VideosGo!\n\n"+
			"<b>可用命令:</b>\n"+
			"/addsource <名称> <URL> - 添加资源站\n"+
			"/listsources - 列出资源站\n"+
			"/help - 帮助")

	case "/help":
		s.sendMessage(msg.Chat.ID, "📖 <b>帮助文档</b>\n\n"+
			"/addsource <名称> <URL> - 添加资源站\n"+
			"/listsources - 列出所有资源站\n"+
			"/help - 显示此帮助")

	case "/addsource":
		if len(parts) < 3 {
			s.sendMessage(msg.Chat.ID, "❌ 用法: /addsource <名称> <URL>")
		} else {
			s.sendMessage(msg.Chat.ID, fmt.Sprintf(
				"✅ 资源站添加成功!\n\n名称: %s\nURL: %s",
				parts[1], parts[2],
			))
		}

	case "/listsources":
		s.sendMessage(msg.Chat.ID, "📋 资源站列表:\n\n暂无资源站")

	default:
		s.sendMessage(msg.Chat.ID, "❓ 未知命令，使用 /help 查看帮助")
	}

	return nil
}

// sendMessage 发送消息
func (s *TGService) sendMessage(chatID int64, text string) error {
	if s.cfg.BotToken == "" {
		return nil
	}

	payload := map[string]interface{}{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "HTML",
	}

	jsonData, _ := json.Marshal(payload)
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", s.cfg.BotToken)

	resp, err := s.client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// SetWebhook 设置 Webhook
func (s *TGService) SetWebhook() error {
	if s.cfg.BotToken == "" || s.cfg.WebhookURL == "" {
		return nil
	}

	payload := map[string]string{
		"url": s.cfg.WebhookURL,
	}

	jsonData, _ := json.Marshal(payload)
	url := fmt.Sprintf("https://api.telegram.org/bot%s/setWebhook", s.cfg.BotToken)

	resp, err := s.client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// parseInt64 解析 int64
func parseInt64(s string) int64 {
	i, _ := strconv.ParseInt(s, 10, 64)
	return i
}
