package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mylazily/videosgo/internal/model"
	"github.com/mylazily/videosgo/internal/repository"
)

// AdRewardService 广告金币服务
type AdRewardService struct {
	repo *repository.AdRewardRepo
}

// NewAdRewardService 创建广告金币服务
func NewAdRewardService(repo *repository.AdRewardRepo) *AdRewardService {
	return &AdRewardService{repo: repo}
}

// CompleteTask 完成任务发放奖励
func (s *AdRewardService) CompleteTask(fingerprintID, taskID uuid.UUID) (*model.CoinTransaction, error) {
	// 获取任务信息
	task, err := s.repo.GetTaskByID(taskID)
	if err != nil {
		return nil, fmt.Errorf("任务不存在: %w", err)
	}

	if !task.IsActive {
		return nil, fmt.Errorf("任务已禁用")
	}

	// 检查今日完成次数
	count, err := s.repo.GetTaskCompletionCount(fingerprintID, taskID)
	if err != nil {
		return nil, fmt.Errorf("查询完成次数失败: %w", err)
	}
	if count >= task.MaxDaily {
		return nil, fmt.Errorf("今日已完成该任务 %d 次，已达上限", count)
	}

	// 确保余额记录存在
	if err := s.repo.EnsureCoinBalance(fingerprintID); err != nil {
		return nil, fmt.Errorf("初始化余额失败: %w", err)
	}

	// 发放奖励
	if err := s.repo.AddCoins(fingerprintID, task.RewardCoins); err != nil {
		return nil, fmt.Errorf("发放奖励失败: %w", err)
	}

	// 获取当前余额
	balance, err := s.repo.GetBalance(fingerprintID)
	if err != nil {
		return nil, fmt.Errorf("获取余额失败: %w", err)
	}

	// 记录交易
	tx := &model.CoinTransaction{
		FingerprintID:   fingerprintID,
		Amount:          task.RewardCoins,
		BalanceAfter:    balance,
		TransactionType: "reward",
		ReferenceID:     taskID.String(),
		Description:     fmt.Sprintf("完成任务: %s", task.TaskName),
	}
	_ = s.repo.CreateTransaction(fingerprintID, task.RewardCoins, balance, "reward", taskID.String(), tx.Description)

	// 记录任务完成
	_ = s.repo.RecordTaskCompletion(fingerprintID, taskID)

	return tx, nil
}

// WatchAdReward 看广告奖励
func (s *AdRewardService) WatchAdReward(fingerprintID uuid.UUID) (*model.CoinTransaction, error) {
	// 查找看广告任务
	tasks, err := s.repo.ListActiveTasks()
	if err != nil {
		return nil, fmt.Errorf("获取任务列表失败: %w", err)
	}

	var adTask *model.AdTask
	for _, t := range tasks {
		if t.TaskType == "watch_ad" {
			adTask = &t
			break
		}
	}

	if adTask == nil {
		return nil, fmt.Errorf("看广告任务未配置")
	}

	return s.CompleteTask(fingerprintID, adTask.ID)
}

// ShareReward 分享奖励
func (s *AdRewardService) ShareReward(fingerprintID uuid.UUID) (*model.CoinTransaction, error) {
	tasks, err := s.repo.ListActiveTasks()
	if err != nil {
		return nil, fmt.Errorf("获取任务列表失败: %w", err)
	}

	var shareTask *model.AdTask
	for _, t := range tasks {
		if t.TaskType == "share" {
			shareTask = &t
			break
		}
	}

	if shareTask == nil {
		return nil, fmt.Errorf("分享任务未配置")
	}

	return s.CompleteTask(fingerprintID, shareTask.ID)
}

// DailyCheckIn 每日签到
func (s *AdRewardService) DailyCheckIn(fingerprintID uuid.UUID) (*model.CoinTransaction, error) {
	tasks, err := s.repo.ListActiveTasks()
	if err != nil {
		return nil, fmt.Errorf("获取任务列表失败: %w", err)
	}

	var checkinTask *model.AdTask
	for _, t := range tasks {
		if t.TaskType == "checkin" {
			checkinTask = &t
			break
		}
	}

	if checkinTask == nil {
		return nil, fmt.Errorf("签到任务未配置")
	}

	return s.CompleteTask(fingerprintID, checkinTask.ID)
}

// UnlockVideoWithCoins 金币解锁视频
func (s *AdRewardService) UnlockVideoWithCoins(fingerprintID, videoID uuid.UUID, cost int64) (*model.CoinTransaction, error) {
	// 确保余额记录存在
	if err := s.repo.EnsureCoinBalance(fingerprintID); err != nil {
		return nil, fmt.Errorf("初始化余额失败: %w", err)
	}

	// 扣除金币
	if err := s.repo.DeductCoins(fingerprintID, cost); err != nil {
		return nil, fmt.Errorf("金币不足")
	}

	// 获取当前余额
	balance, err := s.repo.GetBalance(fingerprintID)
	if err != nil {
		return nil, fmt.Errorf("获取余额失败: %w", err)
	}

	// 记录交易
	tx := &model.CoinTransaction{
		FingerprintID:   fingerprintID,
		Amount:          -cost,
		BalanceAfter:    balance,
		TransactionType: "unlock",
		ReferenceID:     videoID.String(),
		Description:     fmt.Sprintf("解锁视频，花费 %d 金币", cost),
	}
	_ = s.repo.CreateTransaction(fingerprintID, -cost, balance, "unlock", videoID.String(), tx.Description)

	return tx, nil
}

// GetRewardDashboard 获取奖励面板数据
func (s *AdRewardService) GetRewardDashboard(fingerprintID uuid.UUID) (map[string]interface{}, error) {
	// 确保余额记录存在
	_ = s.repo.EnsureCoinBalance(fingerprintID)

	// 获取余额
	balance, err := s.repo.GetBalance(fingerprintID)
	if err != nil {
		return nil, fmt.Errorf("获取余额失败: %w", err)
	}

	// 获取任务列表
	tasks, err := s.repo.ListActiveTasks()
	if err != nil {
		return nil, fmt.Errorf("获取任务列表失败: %w", err)
	}

	// 获取今日完成情况
	completions, err := s.repo.GetDailyCompletions(fingerprintID)
	if err != nil {
		return nil, fmt.Errorf("获取完成情况失败: %w", err)
	}

	// 构建任务完成映射
	completionMap := make(map[string]int)
	for _, c := range completions {
		completionMap[c.TaskID.String()] = c.CompletionCount
	}

	// 构建任务详情
	type TaskDetail struct {
		TaskID          string `json:"task_id"`
		TaskName        string `json:"task_name"`
		TaskType        string `json:"task_type"`
		RewardCoins     int64  `json:"reward_coins"`
		MaxDaily        int    `json:"max_daily"`
		DailyCompleted  int    `json:"daily_completed"`
		CanComplete     bool   `json:"can_complete"`
	}

	var taskDetails []TaskDetail
	for _, task := range tasks {
		dailyCompleted := completionMap[task.ID.String()]
		taskDetails = append(taskDetails, TaskDetail{
			TaskID:         task.ID.String(),
			TaskName:       task.TaskName,
			TaskType:       task.TaskType,
			RewardCoins:    task.RewardCoins,
			MaxDaily:       task.MaxDaily,
			DailyCompleted: dailyCompleted,
			CanComplete:    dailyCompleted < task.MaxDaily,
		})
	}

	// 获取最近交易记录
	history, err := s.repo.GetTransactionHistory(fingerprintID, 10)
	if err != nil {
		history = nil
	}

	return map[string]interface{}{
		"balance":          balance,
		"tasks":            taskDetails,
		"today_completions": len(completions),
		"recent_history":   history,
		"updated_at":       time.Now(),
	}, nil
}

// GetTasks 获取任务列表
func (s *AdRewardService) GetTasks() ([]model.AdTask, error) {
	return s.repo.ListActiveTasks()
}

// GetBalance 获取金币余额
func (s *AdRewardService) GetBalance(fingerprintID uuid.UUID) (int64, error) {
	_ = s.repo.EnsureCoinBalance(fingerprintID)
	return s.repo.GetBalance(fingerprintID)
}

// GetTransactionHistory 获取交易历史
func (s *AdRewardService) GetTransactionHistory(fingerprintID uuid.UUID, limit int) ([]model.CoinTransaction, error) {
	return s.repo.GetTransactionHistory(fingerprintID, limit)
}
