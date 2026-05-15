package model

import (
	"time"

	"github.com/google/uuid"
)

// SignalChannel P2P 信令通道
type SignalChannel struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	RoomID       string    `gorm:"type:varchar(100);index;not null;comment:房间 ID" json:"room_id"`
	PeerID       string    `gorm:"type:varchar(100);index;not null;comment:发送方节点 ID" json:"peer_id"`
	FingerprintID string   `gorm:"type:uuid;index;comment:设备指纹 ID" json:"fingerprint_id"`
	SignalType   string    `gorm:"type:varchar(20);not null;comment:信令类型(offer/answer/ice)" json:"signal_type"`
	SDPData      string    `gorm:"type:text;comment:SDP 数据" json:"sdp_data"`
	ICECandidate string    `gorm:"type:text;comment:ICE 候选" json:"ice_candidate"`
	TargetPeerID string    `gorm:"type:varchar(100);index;comment:目标节点 ID" json:"target_peer_id"`
	TTL          int       `gorm:"default:300;comment:存活时间(秒)" json:"ttl"`
	CreatedAt    time.Time `gorm:"autoCreateTime;index;comment:创建时间" json:"created_at"`
}

// TableName 指定表名
func (SignalChannel) TableName() string {
	return "signal_channels"
}

// BeforeCreate 创建前钩子，确保 UUID 生成
func (s *SignalChannel) BeforeCreate() error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// PeerRegistry P2P 节点注册表
type PeerRegistry struct {
	ID              uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	PeerID          string     `gorm:"type:varchar(100);uniqueIndex;not null;comment:节点 ID" json:"peer_id"`
	FingerprintID   string     `gorm:"type:uuid;index;comment:设备指纹 ID" json:"fingerprint_id"`
	IPAddress       string     `gorm:"type:varchar(50);comment:IP 地址" json:"ip_address"`
	Region          string     `gorm:"type:varchar(50);comment:地区" json:"region"`
	IsActive        bool       `gorm:"default:true;index;comment:是否在线" json:"is_active"`
	CurrentVideoID  string     `gorm:"type:varchar(100);index;comment:当前观看的视频 ID" json:"current_video_id"`
	BandwidthScore  int        `gorm:"default:0;comment:带宽评分" json:"bandwidth_score"`
	LastHeartbeat   *time.Time `gorm:"comment:最后心跳时间" json:"last_heartbeat"`
	ConnectedAt     time.Time  `gorm:"autoCreateTime;comment:连接时间" json:"connected_at"`
	DisconnectedAt  *time.Time `gorm:"comment:断开时间" json:"disconnected_at"`
}

// TableName 指定表名
func (PeerRegistry) TableName() string {
	return "peer_registries"
}

// BeforeCreate 创建前钩子，确保 UUID 生成
func (p *PeerRegistry) BeforeCreate() error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// TransferLog P2P 传输日志
type TransferLog struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SourcePeerID   string    `gorm:"type:varchar(100);index;not null;comment:源节点 ID" json:"source_peer_id"`
	TargetPeerID   string    `gorm:"type:varchar(100);index;not null;comment:目标节点 ID" json:"target_peer_id"`
	VideoID        string    `gorm:"type:varchar(100);index;comment:视频 ID" json:"video_id"`
	DataType       string    `gorm:"type:varchar(50);comment:数据类型(chunk/metadata)" json:"data_type"`
	DataSizeKB     int64     `gorm:"default:0;comment:数据大小(KB)" json:"data_size_kb"`
	TransferTimeMs int       `gorm:"default:0;comment:传输时间(毫秒)" json:"transfer_time_ms"`
	CreatedAt      time.Time `gorm:"autoCreateTime;index;comment:创建时间" json:"created_at"`
}

// TableName 指定表名
func (TransferLog) TableName() string {
	return "transfer_logs"
}

// BeforeCreate 创建前钩子，确保 UUID 生成
func (t *TransferLog) BeforeCreate() error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}
