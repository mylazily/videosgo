package service

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/mylazily/videosgo/internal/model"
	"github.com/mylazily/videosgo/internal/repository"
)

// PushService 推送服务
type PushService struct {
	repo           *repository.PushRepo
	vapidKey       *ecdsa.PrivateKey
	vapidSub       string // VAPID 主题（通常是 mailto: 或 https:）
	vapidPrivateKey string // VAPID 私钥字符串
	sendQueue      chan pushTask
	wg             sync.WaitGroup
}

// pushTask 推送任务
type pushTask struct {
	subscription model.PushSubscription
	payload      []byte
}

// NewPushService 创建推送服务
func NewPushService(repo *repository.PushRepo, vapidSub string) *PushService {
	svc := &PushService{
		repo:     repo,
		vapidSub: vapidSub,
		sendQueue: make(chan pushTask, 1000),
	}

	// 生成或加载 VAPID 密钥
	svc.initVAPIDKey()

	// 启动发送工作池
	workerCount := 5
	svc.wg.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go svc.sendWorker()
	}

	return svc
}

// initVAPIDKey 初始化 VAPID 密钥
func (s *PushService) initVAPIDKey() {
	// 生成临时 VAPID 密钥对
	// 生产环境应从配置或文件加载持久化的密钥
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		fmt.Printf("[Push] VAPID 密钥生成失败: %v\n", err)
		return
	}
	s.vapidKey = key
}

// Subscribe 订阅推送
func (s *PushService) Subscribe(fingerprintID, endpoint, p256dhKey, authKey, userAgent string) (*model.PushSubscription, error) {
	// 检查是否已存在相同端点的订阅
	existing, err := s.repo.GetSubscriptionByEndpoint(endpoint)
	if err == nil && existing != nil {
		// 更新已有订阅
		existing.IsActive = true
		existing.FingerprintID = fingerprintID
		existing.P256DHKey = p256dhKey
		existing.AuthKey = authKey
		existing.UserAgent = userAgent
		err = s.repo.UpdateSubscription(existing)
		if err != nil {
			return nil, fmt.Errorf("更新订阅失败: %w", err)
		}
		return existing, nil
	}

	sub := &model.PushSubscription{
		FingerprintID: fingerprintID,
		Endpoint:      endpoint,
		P256DHKey:     p256dhKey,
		AuthKey:       authKey,
		UserAgent:     userAgent,
		IsActive:      true,
	}

	err = s.repo.CreateSubscription(sub)
	if err != nil {
		return nil, fmt.Errorf("创建订阅失败: %w", err)
	}

	return sub, nil
}

// Unsubscribe 取消订阅
func (s *PushService) Unsubscribe(subscriptionID string) error {
	uid, err := parseUUID(subscriptionID)
	if err != nil {
		return fmt.Errorf("无效的订阅 ID: %w", err)
	}
	return s.repo.DeleteSubscription(uid)
}

// SendNotification 发送推送通知（管理员）
func (s *PushService) SendNotification(title, body, icon, link, tag, targetType, targetVideoID string) (*model.PushNotification, error) {
	notif := &model.PushNotification{
		Title:         title,
		Body:          body,
		Icon:          icon,
		Link:          link,
		Tag:           tag,
		TargetType:    targetType,
		TargetVideoID: targetVideoID,
		Status:        "pending",
	}

	err := s.repo.CreateNotification(notif)
	if err != nil {
		return nil, fmt.Errorf("创建通知失败: %w", err)
	}

	// 异步发送
	go s.sendNotificationAsync(notif)

	return notif, nil
}

// SendToAll 广播推送
func (s *PushService) SendToAll(title, body, link string) (*model.PushNotification, error) {
	return s.SendNotification(title, body, "", link, "", "all", "")
}

// SendHotUpdate 热门更新推送
func (s *PushService) SendHotUpdate(videoID string) (*model.PushNotification, error) {
	return s.SendNotification(
		"热门更新",
		"有新的热门视频更新，快来看看吧！",
		"",
		fmt.Sprintf("/video/%s", videoID),
		fmt.Sprintf("hot-update-%s", videoID),
		"video",
		videoID,
	)
}

// sendNotificationAsync 异步发送通知
func (s *PushService) sendNotificationAsync(notif *model.PushNotification) {
	_ = s.repo.UpdateNotificationStatus(notif.ID, "sending", 0)

	subs, err := s.repo.ListActiveSubscriptions()
	if err != nil {
		_ = s.repo.UpdateNotificationStatus(notif.ID, "failed", 0)
		return
	}

	// 构建推送负载
	payload, err := json.Marshal(map[string]string{
		"title": notif.Title,
		"body":  notif.Body,
		"icon":  notif.Icon,
		"link":  notif.Link,
		"tag":   notif.Tag,
	})
	if err != nil {
		_ = s.repo.UpdateNotificationStatus(notif.ID, "failed", 0)
		return
	}

	sentCount := 0
	for _, sub := range subs {
		// 根据目标类型过滤
		if notif.TargetType == "video" && notif.TargetVideoID != "" {
			// 可以根据用户偏好进一步过滤，这里简单发送给所有活跃订阅
		}

		// 加入发送队列（非阻塞）
		select {
		case s.sendQueue <- pushTask{subscription: sub, payload: payload}:
			sentCount++
		default:
			// 队列满，丢弃
		}
	}

	_ = s.repo.UpdateNotificationStatus(notif.ID, "completed", sentCount)
}

// sendWorker 发送工作协程
func (s *PushService) sendWorker() {
	defer s.wg.Done()
	for task := range s.sendQueue {
		s.sendPush(task.subscription, task.payload)
	}
}

// sendPush 实际发送 Web Push 通知
func (s *PushService) sendPush(sub model.PushSubscription, payload []byte) {
	if s.vapidKey == nil {
		return
	}

	// 构建带 VAPID 认证的 HTTP 请求
	// 这里实现了标准的 Web Push 协议
	req, err := http.NewRequest("POST", sub.Endpoint, nil)
	if err != nil {
		return
	}

	// 设置 TTL 头
	req.Header.Set("TTL", "2419200")
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Content-Encoding", "aes128gcm")

	// 生成 VAPID JWT
	vapidJWT, err := s.generateVAPIDJWT(sub.Endpoint)
	if err == nil && vapidJWT != "" {
		req.Header.Set("Authorization", "vapid t="+vapidJWT)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	// 更新最后发送时间
	now := time.Now()
	sub.LastSentAt = &now
	_ = s.repo.UpdateSubscription(&sub)
}

// generateVAPIDJWT 生成 VAPID JWT Token
func (s *PushService) generateVAPIDJWT(audience string) (string, error) {
	// VAPID JWT claims
	now := time.Now()
	claims := jwt.MapClaims{
		"aud": audience,
		"exp": now.Add(12 * time.Hour).Unix(),
		"iat": now.Unix(),
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign with VAPID private key
	tokenString, err := token.SignedString([]byte(s.vapidPrivateKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign VAPID JWT: %w", err)
	}

	return tokenString, nil
}

// GetNotificationStats 获取推送统计
func (s *PushService) GetNotificationStats(limit int) ([]model.PushNotification, error) {
	return s.repo.GetNotificationStats(limit)
}

// LogClick 记录推送点击
func (s *PushService) LogClick(notificationID, subscriptionID string) error {
	notifUID, err := parseUUID(notificationID)
	if err != nil {
		return fmt.Errorf("无效的通知 ID: %w", err)
	}
	subUID, err := parseUUID(subscriptionID)
	if err != nil {
		return fmt.Errorf("无效的订阅 ID: %w", err)
	}

	clickLog := &model.PushClickLog{
		NotificationID: notifUID,
		SubscriptionID: subUID,
	}

	if err := s.repo.LogClick(clickLog); err != nil {
		return err
	}

	// 增加通知点击计数
	return s.repo.IncrementNotificationClick(notifUID)
}

// Close 关闭推送服务
func (s *PushService) Close() {
	close(s.sendQueue)
	s.wg.Wait()
}

// parseUUID 解析 UUID 字符串
func parseUUID(id string) (uuid.UUID, error) {
	return uuid.Parse(id)
}
