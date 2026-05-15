package service

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	// 写超时
	writeWait = 10 * time.Second
	// 读超时（心跳间隔）
	pongWait = 30 * time.Second
	// 心跳发送间隔
	pingPeriod = (pongWait * 9) / 10
	// 最大消息大小
	maxMessageSize = 512
)

// DanmakuMessage 弹幕消息
type DanmakuMessage struct {
	Type     string `json:"type"`     // 消息类型: danmaku/online_count/system
	VideoID  string `json:"video_id"` // 视频 ID
	Content  string `json:"content"`  // 弹幕内容
	User     string `json:"user"`     // 用户标识
	Color    string `json:"color"`    // 弹幕颜色
	Time     int64  `json:"time"`     // 时间戳
	Online   int    `json:"online"`   // 在线人数
}

// Connection WebSocket 连接
type Connection struct {
	VideoID string
	Conn    *websocket.Conn
	Send    chan []byte
	User    string // 用户标识
}

// Hub WebSocket 连接管理中心
type Hub struct {
	// 视频房间映射 videoID -> []*Connection
	rooms map[string][]*Connection
	mu    sync.RWMutex
}

// NewHub 创建 Hub
func NewHub() *Hub {
	return &Hub{
		rooms: make(map[string][]*Connection),
	}
}

// Register 注册连接
func (h *Hub) Register(conn *Connection) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.rooms[conn.VideoID] = append(h.rooms[conn.VideoID], conn)
	log.Printf("[WS] 新连接加入房间 %s，当前房间人数: %d", conn.VideoID, len(h.rooms[conn.VideoID]))
}

// Unregister 注销连接
func (h *Hub) Unregister(conn *Connection) {
	h.mu.Lock()
	defer h.mu.Unlock()

	conns, ok := h.rooms[conn.VideoID]
	if !ok {
		return
	}

	for i, c := range conns {
		if c == conn {
			h.rooms[conn.VideoID] = append(conns[:i], conns[i+1:]...)
			close(conn.Send)
			break
		}
	}

	if len(h.rooms[conn.VideoID]) == 0 {
		delete(h.rooms, conn.VideoID)
	}

	log.Printf("[WS] 连接离开房间 %s，当前房间人数: %d", conn.VideoID, len(h.rooms[conn.VideoID]))
}

// BroadcastDanmaku 广播弹幕到指定视频房间
func (h *Hub) BroadcastDanmaku(videoID string, danmaku *DanmakuMessage) {
	data, err := json.Marshal(danmaku)
	if err != nil {
		log.Printf("[WS] 序列化弹幕失败: %v", err)
		return
	}

	h.mu.RLock()
	conns, ok := h.rooms[videoID]
	h.mu.RUnlock()

	if !ok {
		return
	}

	for _, conn := range conns {
		select {
		case conn.Send <- data:
		default:
			// 发送缓冲区满，关闭连接
			go func(c *Connection) {
				h.Unregister(c)
				c.Conn.Close()
			}(conn)
		}
	}
}

// GetOnlineCount 获取指定视频的在线人数
func (h *Hub) GetOnlineCount(videoID string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.rooms[videoID])
}

// GetTotalOnline 获取总在线人数
func (h *Hub) GetTotalOnline() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	total := 0
	for _, conns := range h.rooms {
		total += len(conns)
	}
	return total
}

// ReadPump 读取泵
func (c *Connection) ReadPump(hub *Hub) {
	defer func() {
		hub.Unregister(c)
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}

		// 解析弹幕消息
		var danmaku DanmakuMessage
		if err := json.Unmarshal(message, &danmaku); err != nil {
			continue
		}

		// 设置弹幕元信息
		danmaku.Type = "danmaku"
		danmaku.VideoID = c.VideoID
		danmaku.User = c.User
		danmaku.Time = time.Now().UnixMilli()

		// 广播弹幕
		hub.BroadcastDanmaku(c.VideoID, &danmaku)
	}
}

// WritePump 写入泵
func (c *Connection) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// 通道关闭
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// 批量发送缓冲区中的消息
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte("\n"))
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			// 发送心跳
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// WSService WebSocket 服务
type WSService struct {
	hub *Hub
}

// NewWSService 创建 WebSocket 服务
func NewWSService() *WSService {
	return &WSService{
		hub: NewHub(),
	}
}

// GetHub 获取 Hub
func (s *WSService) GetHub() *Hub {
	return s.hub
}

// HandleConnection 处理 WebSocket 连接
func (s *WSService) HandleConnection(conn *websocket.Conn, videoID string, user string) {
	connection := &Connection{
		VideoID: videoID,
		Conn:    conn,
		Send:    make(chan []byte, 256),
		User:    user,
	}

	s.hub.Register(connection)

	// 发送在线人数
	onlineMsg := &DanmakuMessage{
		Type:    "online_count",
		VideoID: videoID,
		Online:  s.hub.GetOnlineCount(videoID),
	}
	data, _ := json.Marshal(onlineMsg)
	connection.Send <- data

	go connection.WritePump()
	go connection.ReadPump(s.hub)
}

// GetOnlineCount 获取在线人数
func (s *WSService) GetOnlineCount(videoID string) int {
	return s.hub.GetOnlineCount(videoID)
}

// BroadcastSystemMessage 广播系统消息
func (s *WSService) BroadcastSystemMessage(videoID, message string) {
	msg := &DanmakuMessage{
		Type:    "system",
		VideoID: videoID,
		Content: message,
		Time:    time.Now().UnixMilli(),
	}
	s.hub.BroadcastDanmaku(videoID, msg)
}

// ValidateVideoID 验证视频 ID 格式
func ValidateVideoID(videoID string) bool {
	_, err := uuid.Parse(videoID)
	return err == nil
}
