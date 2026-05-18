package service

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
	"videosgo/internal/model"
	"videosgo/internal/repository"
)

// PaymentService 支付服务
type PaymentService struct {
	repo *repository.PaymentRepo
}

// NewPaymentService 创建支付服务
func NewPaymentService(repo *repository.PaymentRepo) *PaymentService {
	return &PaymentService{repo: repo}
}

// GenerateOrderNo 生成唯一订单号（线程安全）
func (s *PaymentService) GenerateOrderNo() string {
	timestamp := time.Now().Format("20060102150405")
	// 使用 crypto/rand 保证并发安全
	n, _ := rand.Int(rand.Reader, big.NewInt(1000000))
	random := n.Int64()
	return fmt.Sprintf("PAY%s%06d", timestamp, random)
}

// CreateOrder 创建支付订单
func (s *PaymentService) CreateOrder(fingerprintID *uuid.UUID, channelID uuid.UUID, productType, productID, productName string, amount float64) (*model.PaymentOrder, error) {
	// 获取渠道信息
	channel, err := s.repo.GetChannelByID(channelID)
	if err != nil {
		return nil, fmt.Errorf("获取支付渠道失败: %w", err)
	}

	// 计算手续费
	feeAmount := amount * channel.FeeRate

	order := &model.PaymentOrder{
		OrderNo:       s.GenerateOrderNo(),
		FingerprintID: fingerprintID,
		ChannelID:     channelID,
		ProductType: productType,
		ProductID:   productID,
		ProductName: productName,
		Amount:      amount,
		FeeAmount:   feeAmount,
		Status:      "pending",
		ExpiresAt:   time.Now().Add(30 * time.Minute),
	}

	if err := s.repo.CreateOrder(order); err != nil {
		return nil, fmt.Errorf("创建订单失败: %w", err)
	}

	return order, nil
}

// ProcessPayment 处理支付（模拟实现）
func (s *PaymentService) ProcessPayment(orderNo string, channelType string) (*model.PaymentOrder, error) {
	order, err := s.repo.GetOrderByNo(orderNo)
	if err != nil {
		return nil, fmt.Errorf("订单不存在: %w", err)
	}

	if order.Status != "pending" {
		return nil, fmt.Errorf("订单状态异常: %s", order.Status)
	}

	// 检查是否过期
	if time.Now().After(order.ExpiresAt) {
		_ = s.repo.UpdateOrderStatus(orderNo, "expired", "")
		return nil, fmt.Errorf("订单已过期")
	}

	// 模拟支付处理
	// TODO: 对接真实支付网关（支付宝、微信支付等）
	paymentNo := fmt.Sprintf("SIM%s%d", channelType, time.Now().UnixNano())

	// 模拟支付成功（90% 成功率）
	if rand.Float64() > 0.1 {
		// 支付成功
		err = s.repo.UpdateOrderStatus(orderNo, "paid", paymentNo)
		if err != nil {
			return nil, fmt.Errorf("更新订单状态失败: %w", err)
		}

		// 如果是 VIP 产品，自动创建 VIP 订阅
		if order.ProductType == "vip" {
			planDuration := s.getPlanDuration(order.ProductID)
			sub := &model.VIPSubscription{
				FingerprintID: order.FingerprintID,
				PlanType:      order.ProductID,
				StartAt:       time.Now(),
				ExpiresAt:     time.Now().Add(planDuration),
				IsActive:      true,
				AutoRenew:     false,
			}
			if err := s.repo.CreateVIP(sub); err != nil {
				// VIP 创建失败是严重问题，记录日志并回滚订单状态
				fmt.Printf("[Payment] VIP 创建失败（订单 %s）: %v\n", orderNo, err)
				_ = s.repo.UpdateOrderStatus(orderNo, "failed", "VIP创建失败")
				return order, fmt.Errorf("支付成功但 VIP 激活失败，请联系客服")
			}
		}

		// 重新获取订单
		return s.repo.GetOrderByNo(orderNo)
	}

	// 支付失败
	if err := s.repo.UpdateOrderStatus(orderNo, "failed", ""); err != nil {
		fmt.Printf("[Payment] 更新订单状态失败: %v\n", err)
	}
	return order, nil
}

// getPlanDuration 根据套餐类型获取时长
func (s *PaymentService) getPlanDuration(planType string) time.Duration {
	switch planType {
	case "monthly":
		return 30 * 24 * time.Hour
	case "quarterly":
		return 90 * 24 * time.Hour
	case "yearly":
		return 365 * 24 * time.Hour
	default:
		return 30 * 24 * time.Hour
	}
}

// VerifyVIPAccess 验证 VIP 权限
func (s *PaymentService) VerifyVIPAccess(fingerprintID uuid.UUID) (*model.VIPSubscription, error) {
	sub, err := s.repo.GetActiveVIP(fingerprintID)
	if err != nil {
		return nil, fmt.Errorf("无活跃 VIP 订阅")
	}
	return sub, nil
}

// GetOrder 获取订单
func (s *PaymentService) GetOrder(orderNo string) (*model.PaymentOrder, error) {
	return s.repo.GetOrderByNo(orderNo)
}

// GetChannels 获取支付渠道列表
func (s *PaymentService) GetChannels() ([]model.PaymentChannel, error) {
	return s.repo.ListChannels()
}

// GetVIPStatus 获取 VIP 状态
func (s *PaymentService) GetVIPStatus(fingerprintID uuid.UUID) (*model.VIPSubscription, error) {
	return s.repo.GetActiveVIP(fingerprintID)
}
