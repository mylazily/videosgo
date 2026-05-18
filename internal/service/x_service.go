package service

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"videosgo/internal/model"
	"videosgo/internal/repository"
)

// XService X.com 自动发布服务
type XService struct {
	repo *repository.XRepo
}

// NewXService 创建 X.com 服务
func NewXService(repo *repository.XRepo) *XService {
	return &XService{repo: repo}
}

// CreatePost 创建推文
func (s *XService) CreatePost(accountID, videoID uuid.UUID, text string, hashtags []string) (*model.XPostLog, error) {
	// 获取账号信息
	account, err := s.repo.GetAccountByID(accountID)
	if err != nil {
		return nil, fmt.Errorf("获取账号失败: %w", err)
	}

	// 构建标签 JSON
	hashtagJSON := make(map[string]interface{})
	if len(hashtags) > 0 {
		hashtagJSON["tags"] = hashtags
	}

	postLog := &model.XPostLog{
		AccountID: accountID,
		VideoID:   &videoID,
		TweetText: text,
		Hashtags:  hashtagJSON,
		Status:    "pending",
	}

	// 异步发布
	go s.publishPostAsync(postLog, account, hashtags)

	if err := s.repo.CreatePostLog(postLog); err != nil {
		return nil, err
	}
	return postLog, nil
}

// publishPostAsync 异步发布推文
func (s *XService) publishPostAsync(postLog *model.XPostLog, account *model.XAccount, hashtags []string) {
	log.Printf("[X] 正在发布推文到账号 %s...", account.Username)

	// 模拟发布延迟
	time.Sleep(1 * time.Second)

	// 模拟生成推文 ID
	tweetID := fmt.Sprintf("%d", time.Now().UnixNano())

	// 更新状态为已发布
	_ = s.repo.MarkAsPosted(postLog.ID, tweetID)

	// 更新账号最后发布时间
	now := time.Now()
	account.LastPostAt = &now
	_ = s.repo.UpdateAccount(account)

	log.Printf("[X] 推文发布成功，TweetID: %s", tweetID)
}

// ProcessQueue 处理发布队列
func (s *XService) ProcessQueue() (int, error) {
	pendingPosts, err := s.repo.GetPendingPosts()
	if err != nil {
		return 0, fmt.Errorf("获取待发布队列失败: %w", err)
	}

	count := 0
	for _, post := range pendingPosts {
		// 标记为处理中
		_ = s.repo.MarkQueueAsProcessing(post.ID)

		// 获取账号信息
		account, err := s.repo.GetAccountByID(post.AccountID)
		if err != nil {
			log.Printf("[X] 获取账号失败: %v", err)
			_ = s.repo.MarkQueueAsFailed(post.ID)
			continue
		}

		// 生成推文内容
		text := fmt.Sprintf("新视频推荐！快来观看 #短视频 #推荐 %s", time.Now().Format("2006-01-02"))

		postLog := &model.XPostLog{
			AccountID: post.AccountID,
			VideoID:   post.VideoID,
			TweetText: text,
			Status:    "pending",
		}

		// 创建发布日志
		_ = s.repo.CreatePostLog(postLog)

		// 异步发布
		go s.publishPostAsync(postLog, account, nil)

		// 标记队列为已发布
		_ = s.repo.MarkQueueAsPosted(post.ID)
		count++
	}

	return count, nil
}

// GenerateTweetContent 根据视频信息生成推文文案
func (s *XService) GenerateTweetContent(title, category, tags string) string {
	text := fmt.Sprintf("🎬 %s", title)
	if category != "" {
		text = fmt.Sprintf("%s | %s", text, category)
	}
	if tags != "" {
		// 将逗号分隔的标签转换为 # 标签
		tagList := strings.Split(tags, ",")
		for _, tag := range tagList {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				text = fmt.Sprintf("%s #%s", text, tag)
			}
		}
	}
	// 截断到 280 字符（X 推文限制）
	if len(text) > 280 {
		text = text[:277] + "..."
	}
	return text
}

// GetAccounts 获取账号列表
func (s *XService) GetAccounts() ([]model.XAccount, error) {
	return s.repo.ListAccounts()
}

// GetPostLogs 获取发布记录
func (s *XService) GetPostLogs(limit int) ([]model.XPostLog, error) {
	return s.repo.ListPostLogs(limit)
}

// AddToQueue 添加到发布队列
func (s *XService) AddToQueue(accountID, videoID uuid.UUID, scheduledAt time.Time) error {
	queue := &model.XPostQueue{
		AccountID:   accountID,
		VideoID:     &videoID,
		ScheduledAt: scheduledAt,
		Status:      "pending",
		MaxRetries:  3,
	}
	return s.repo.CreatePostQueue(queue)
}
