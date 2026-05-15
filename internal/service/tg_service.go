package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/mylazily/videosgo/internal/model"
	"github.com/mylazily/videosgo/internal/repository"
)

// TGService TG Bot 服务
type TGService struct {
	repo *repository.TGRepo
}

// NewTGService 创建 TG Bot 服务
func NewTGService(repo *repository.TGRepo) *TGService {
	return &TGService{repo: repo}
}

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
		"chat_id": channelID,
		"text":    message,
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
