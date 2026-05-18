package service

import (
	"sync"
	"time"

	"videosgo/internal/model"
	"videosgo/internal/repository"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// P2PService P2P 信令服务
type P2PService struct {
	repo         *repository.P2PRepo
	memoryPeers  sync.Map // 内存中的在线节点缓存 peerID -> *model.PeerRegistry
	memorySignals sync.Map // 内存中的信令缓存 roomID -> []model.SignalChannel
	roomMu       sync.Map // 保护每个 room 的并发操作 roomID -> *sync.Mutex
	logger       *zap.Logger
}

// NewP2PService 创建 P2P 信令服务
func NewP2PService(repo *repository.P2PRepo, logger *zap.Logger) *P2PService {
	return &P2PService{repo: repo, logger: logger}
}

// getRoomMutex 获取指定 room 的互斥锁
func (s *P2PService) getRoomMutex(roomID string) *sync.Mutex {
	val, _ := s.roomMu.LoadOrStore(roomID, &sync.Mutex{})
	return val.(*sync.Mutex)
}

// parseUUID 安全解析 UUID，空字符串时返回 uuid.Nil
func parseUUID(s string) uuid.UUID {
	if s == "" {
		return uuid.Nil
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil
	}
	return id
}

// RegisterPeer 注册节点
func (s *P2PService) RegisterPeer(peer *model.PeerRegistry) error {
	// 内存写入
	s.memoryPeers.Store(peer.PeerID, peer)
	// 数据库写入
	return s.repo.RegisterPeer(peer)
}

// UnregisterPeer 注销节点
func (s *P2PService) UnregisterPeer(peerID string) error {
	// 内存删除
	s.memoryPeers.Delete(peerID)
	// 数据库更新
	return s.repo.UnregisterPeer(peerID)
}

// Heartbeat 更新节点心跳
func (s *P2PService) Heartbeat(peerID, videoID string) error {
	// 更新内存缓存
	if val, ok := s.memoryPeers.Load(peerID); ok {
		if peer, ok := val.(*model.PeerRegistry); ok {
			now := time.Now()
			peer.LastHeartbeat = &now
			if videoID != "" {
				peer.CurrentVideoID = parseUUID(videoID)
			}
		}
	}
	// 数据库更新
	return s.repo.Heartbeat(peerID, videoID)
}

// OfferSignal 发送 Offer 信令
func (s *P2PService) OfferSignal(roomID, peerID, fingerprintID, targetPeerID, sdpData string) error {
	signal := &model.SignalChannel{
		RoomID:        roomID,
		PeerID:        peerID,
		FingerprintID: parseUUID(fingerprintID),
		SignalType:    "offer",
		SDPData:       sdpData,
		TargetPeerID:  targetPeerID,
		TTL:           300,
	}
	// 内存写入
	s.appendMemorySignal(roomID, *signal)
	// 数据库写入
	return s.repo.CreateSignal(signal)
}

// AnswerSignal 发送 Answer 信令
func (s *P2PService) AnswerSignal(roomID, peerID, fingerprintID, targetPeerID, sdpData string) error {
	signal := &model.SignalChannel{
		RoomID:        roomID,
		PeerID:        peerID,
		FingerprintID: parseUUID(fingerprintID),
		SignalType:    "answer",
		SDPData:       sdpData,
		TargetPeerID:  targetPeerID,
		TTL:           300,
	}
	s.appendMemorySignal(roomID, *signal)
	return s.repo.CreateSignal(signal)
}

// ExchangeICE 交换 ICE 候选
func (s *P2PService) ExchangeICE(roomID, peerID, fingerprintID, targetPeerID, iceCandidate string) error {
	signal := &model.SignalChannel{
		RoomID:        roomID,
		PeerID:        peerID,
		FingerprintID: parseUUID(fingerprintID),
		SignalType:    "ice",
		ICECandidate:  iceCandidate,
		TargetPeerID:  targetPeerID,
		TTL:           300,
	}
	s.appendMemorySignal(roomID, *signal)
	return s.repo.CreateSignal(signal)
}

// GetSignalsForPeer 获取目标节点的信令
func (s *P2PService) GetSignalsForPeer(roomID, targetPeerID string) ([]model.SignalChannel, error) {
	// 优先从内存获取
	if val, ok := s.memorySignals.Load(roomID); ok {
		signals := val.([]model.SignalChannel)
		var filtered []model.SignalChannel
		for _, sig := range signals {
			if sig.TargetPeerID == targetPeerID {
				filtered = append(filtered, sig)
			}
		}
		if len(filtered) > 0 {
			return filtered, nil
		}
	}
	// 降级到数据库
	return s.repo.GetSignalsByRoomAndTarget(roomID, targetPeerID, 50)
}

// GetPeersForVideo 获取正在看同一视频的在线节点
func (s *P2PService) GetPeersForVideo(videoID string) ([]model.PeerRegistry, error) {
	// 优先从内存获取
	var peers []model.PeerRegistry
	s.memoryPeers.Range(func(key, value interface{}) bool {
		if peer, ok := value.(*model.PeerRegistry); ok {
			if peer.IsActive && peer.CurrentVideoID.String() == videoID {
				peers = append(peers, *peer)
			}
		}
		return true
	})
	if len(peers) > 0 {
		return peers, nil
	}
	// 降级到数据库
	return s.repo.GetActivePeersByVideo(videoID)
}

// GetAllActivePeers 获取所有在线节点
func (s *P2PService) GetAllActivePeers() ([]model.PeerRegistry, error) {
	// 优先从内存获取
	var peers []model.PeerRegistry
	s.memoryPeers.Range(func(key, value interface{}) bool {
		if peer, ok := value.(*model.PeerRegistry); ok {
			if peer.IsActive {
				peers = append(peers, *peer)
			}
		}
		return true
	})
	if len(peers) > 0 {
		return peers, nil
	}
	// 降级到数据库
	return s.repo.GetAllActivePeers()
}

// LogTransfer 记录传输日志
func (s *P2PService) LogTransfer(log *model.TransferLog) error {
	return s.repo.LogTransfer(log)
}

// Cleanup 定期清理过期信令和失活节点
func (s *P2PService) Cleanup() {
	// 清理过期信令（5 分钟过期）
	deletedSignals, err := s.repo.CleanupExpiredSignals(300)
	if err == nil {
		s.logger.Info("清理了过期信令", zap.Int64("count", deletedSignals))
	}

	// 清理失活节点（心跳超时 60 秒）
	deletedPeers, err := s.repo.CleanupInactivePeers(60)
	if err == nil {
		s.logger.Info("清理了失活节点", zap.Int64("count", deletedPeers))
		// 同步清理内存缓存
		if deletedPeers > 0 {
			allPeers, _ := s.repo.GetAllActivePeers()
			activeSet := make(map[string]bool)
			for _, p := range allPeers {
				activeSet[p.PeerID] = true
			}
			s.memoryPeers.Range(func(key, value interface{}) bool {
				if peerID, ok := key.(string); ok {
					if !activeSet[peerID] {
						s.memoryPeers.Delete(key)
					}
				}
				return true
			})
		}
	}

	// 清理内存信令缓存（超过 5 分钟的）
	now := time.Now()
	s.memorySignals.Range(func(key, value interface{}) bool {
		if signals, ok := value.([]model.SignalChannel); ok {
			var valid []model.SignalChannel
			for _, sig := range signals {
				if now.Sub(sig.CreatedAt) < 5*time.Minute {
					valid = append(valid, sig)
				}
			}
			if len(valid) == 0 {
				s.memorySignals.Delete(key)
			} else {
				s.memorySignals.Store(key, valid)
			}
		}
		return true
	})
}

// appendMemorySignal 向内存信令缓存追加信令
func (s *P2PService) appendMemorySignal(roomID string, signal model.SignalChannel) {
	mu := s.getRoomMutex(roomID)
	mu.Lock()
	defer mu.Unlock()

	// 设置创建时间
	signal.CreatedAt = time.Now()

	val, _ := s.memorySignals.LoadOrStore(roomID, []model.SignalChannel{})
	signals := val.([]model.SignalChannel)
	signals = append(signals, signal)
	// 限制内存中每个房间的信令数量
	if len(signals) > 100 {
		signals = signals[len(signals)-100:]
	}
	s.memorySignals.Store(roomID, signals)
}
