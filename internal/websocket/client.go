package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	// 写入等待时间
	writeWait = 10 * time.Second

	// 读取等待时间
	pongWait = 60 * time.Second

	// ping 发送间隔，必须小于 pongWait
	pingPeriod = (pongWait * 9) / 10

	// 最大消息大小
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// 在生产环境中应该检查 Origin
		return true
	},
}

// Client WebSocket 客户端连接
type Client struct {
	// 基本信息
	ID     string `json:"id"`      // 连接唯一标识
	UserID int64  `json:"user_id"` // 用户ID

	// WebSocket 连接
	conn *websocket.Conn

	// 消息发送通道
	send chan []byte

	// Hub 引用
	hub *Hub

	// 当前房间
	CurrentRoom string `json:"current_room"`

	// 连接时间
	ConnectedAt time.Time `json:"connected_at"`
}

// Message WebSocket 消息结构
type Message struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

// NewClient 创建新的客户端连接
func NewClient(hub *Hub, conn *websocket.Conn, userID int64) *Client {
	return &Client{
		ID:          uuid.New().String(),
		UserID:      userID,
		conn:        conn,
		send:        make(chan []byte, 256),
		hub:         hub,
		ConnectedAt: time.Now(),
	}
}

// Start 启动客户端连接处理
func (c *Client) Start() {
	// 注册到 Hub
	c.hub.register <- c

	// 启动读写协程
	go c.writePump()
	go c.readPump()
}

// readPump 处理从 WebSocket 连接读取消息
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, messageBytes, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket 读取错误: %v", err)
			}
			break
		}

		// 解析消息
		var msg Message
		if err := json.Unmarshal(messageBytes, &msg); err != nil {
			log.Printf("消息解析失败: %v", err)
			continue
		}

		// 处理消息
		c.handleMessage(&msg)
	}
}

// writePump 处理向 WebSocket 连接写入消息
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub 关闭了发送通道
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// 批量发送队列中的其他消息
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage 处理接收到的消息
func (c *Client) handleMessage(msg *Message) {
	switch msg.Event {
	case "join_room":
		c.handleJoinRoom(msg.Data)
	case "leave_room":
		c.handleLeaveRoom(msg.Data)
	case "collaboration_operation":
		c.handleCollaborationOperation(msg.Data)
	case "cursor_update":
		c.handleCursorUpdate(msg.Data)
	default:
		log.Printf("未知消息类型: %s", msg.Event)
	}
}

// handleJoinRoom 处理加入房间
func (c *Client) handleJoinRoom(data interface{}) {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		c.SendError("invalid_data", "加入房间数据格式错误")
		return
	}

	roomID, ok := dataMap["room_id"].(string)
	if !ok || roomID == "" {
		c.SendError("invalid_room_id", "房间ID无效")
		return
	}

	// 如果已在其他房间，先离开
	if c.CurrentRoom != "" && c.CurrentRoom != roomID {
		c.hub.LeaveRoom(c, c.CurrentRoom)
	}

	// 加入新房间
	if err := c.hub.JoinRoom(c, roomID); err != nil {
		c.SendError("join_room_failed", err.Error())
		return
	}
}

// handleLeaveRoom 处理离开房间
func (c *Client) handleLeaveRoom(data interface{}) {
	if c.CurrentRoom == "" {
		return
	}

	c.hub.LeaveRoom(c, c.CurrentRoom)
}

// handleCollaborationOperation 处理协作操作
func (c *Client) handleCollaborationOperation(data interface{}) {
	if c.CurrentRoom == "" {
		c.SendError("not_in_room", "未加入任何房间")
		return
	}

	// 广播操作给房间内其他用户
	c.hub.BroadcastToRoom(c.CurrentRoom, "collaboration_operation", map[string]interface{}{
		"user_id":   c.UserID,
		"socket_id": c.ID,
		"operation": data,
		"timestamp": time.Now(),
	})
}

// handleCursorUpdate 处理光标更新
func (c *Client) handleCursorUpdate(data interface{}) {
	if c.CurrentRoom == "" {
		return
	}

	// 广播光标位置给房间内其他用户
	c.hub.BroadcastToRoom(c.CurrentRoom, "cursor_update", map[string]interface{}{
		"user_id":   c.UserID,
		"socket_id": c.ID,
		"cursor":    data,
		"timestamp": time.Now(),
	})
}

// Send 发送消息给客户端
func (c *Client) Send(event string, data interface{}) {
	message := Message{
		Event: event,
		Data:  data,
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		log.Printf("消息序列化失败: %v", err)
		return
	}

	select {
	case c.send <- messageBytes:
	default:
		// 发送队列已满，关闭连接
		close(c.send)
	}
}

// SendError 发送错误消息
func (c *Client) SendError(code, message string) {
	c.Send("error", map[string]interface{}{
		"code":      code,
		"message":   message,
		"timestamp": time.Now(),
	})
}

// Close 关闭客户端连接
func (c *Client) Close() {
	if c.CurrentRoom != "" {
		c.hub.LeaveRoom(c, c.CurrentRoom)
	}
	c.conn.Close()
}
