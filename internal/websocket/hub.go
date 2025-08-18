package websocket

import (
	"log"
	"sync"
	"time"

	"DOC/domain"
)

// Hub WebSocket 连接管理中心
// 负责管理所有 WebSocket 连接，实现房间管理和消息广播
type Hub struct {
	// 客户端连接管理
	clients   map[*Client]bool            // 所有活跃的客户端连接
	rooms     map[string]map[*Client]bool // 房间 -> 客户端映射
	userRooms map[int64]map[string]bool   // 用户 -> 房间映射

	// 消息通道
	register   chan *Client           // 客户端注册通道
	unregister chan *Client           // 客户端注销通道
	broadcast  chan *BroadcastMessage // 广播消息通道

	// 协作相关
	collaborationRepo domain.CollaborationRepository

	// 控制
	mu      sync.RWMutex
	running bool
	stopCh  chan struct{}
}

// BroadcastMessage 广播消息结构
type BroadcastMessage struct {
	RoomID  string      `json:"room_id"`
	Event   string      `json:"event"`
	Data    interface{} `json:"data"`
	Exclude *Client     `json:"-"` // 排除的客户端（通常是发送者）
}

// NewHub 创建新的 Hub 实例
func NewHub(collaborationRepo domain.CollaborationRepository) *Hub {
	return &Hub{
		clients:           make(map[*Client]bool),
		rooms:             make(map[string]map[*Client]bool),
		userRooms:         make(map[int64]map[string]bool),
		register:          make(chan *Client),
		unregister:        make(chan *Client),
		broadcast:         make(chan *BroadcastMessage),
		collaborationRepo: collaborationRepo,
		stopCh:            make(chan struct{}),
	}
}

// Start 启动 Hub
func (h *Hub) Start() {
	h.mu.Lock()
	if h.running {
		h.mu.Unlock()
		return
	}
	h.running = true
	h.mu.Unlock()

	log.Println("WebSocket Hub 启动")

	go h.run()
}

// Stop 停止 Hub
func (h *Hub) Stop() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.running {
		return
	}

	log.Println("正在停止 WebSocket Hub...")
	close(h.stopCh)
	h.running = false

	// 关闭所有客户端连接
	for client := range h.clients {
		close(client.send)
	}

	log.Println("WebSocket Hub 已停止")
}

// run Hub 主循环
func (h *Hub) run() {
	ticker := time.NewTicker(30 * time.Second) // 定期清理
	defer ticker.Stop()

	for {
		select {
		case <-h.stopCh:
			return

		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.broadcastMessage(message)

		case <-ticker.C:
			h.cleanup()
		}
	}
}

// registerClient 注册客户端
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.clients[client] = true
	log.Printf("客户端已连接: UserID=%d, SocketID=%s", client.UserID, client.ID)

	// 发送连接成功消息
	client.Send("connected", map[string]interface{}{
		"socket_id": client.ID,
		"user_id":   client.UserID,
		"timestamp": time.Now(),
	})
}

// unregisterClient 注销客户端
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[client]; ok {
		// 从所有房间中移除
		for roomID := range h.userRooms[client.UserID] {
			h.leaveRoomUnsafe(client, roomID)
		}

		// 清理用户房间映射
		delete(h.userRooms, client.UserID)

		// 移除客户端
		delete(h.clients, client)
		close(client.send)

		log.Printf("客户端已断开: UserID=%d, SocketID=%s", client.UserID, client.ID)
	}
}

// JoinRoom 加入房间
func (h *Hub) JoinRoom(client *Client, roomID string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 初始化房间
	if h.rooms[roomID] == nil {
		h.rooms[roomID] = make(map[*Client]bool)
	}

	// 初始化用户房间映射
	if h.userRooms[client.UserID] == nil {
		h.userRooms[client.UserID] = make(map[string]bool)
	}

	// 添加到房间
	h.rooms[roomID][client] = true
	h.userRooms[client.UserID][roomID] = true
	client.CurrentRoom = roomID

	log.Printf("用户 %d 加入房间 %s", client.UserID, roomID)

	// 通知房间内其他用户
	h.broadcastToRoomUnsafe(roomID, "user_joined", map[string]interface{}{
		"user_id":   client.UserID,
		"socket_id": client.ID,
		"room_id":   roomID,
		"timestamp": time.Now(),
	}, client)

	// 发送房间信息给新用户
	roomUsers := h.getRoomUsersUnsafe(roomID)
	client.Send("room_joined", map[string]interface{}{
		"room_id":   roomID,
		"users":     roomUsers,
		"timestamp": time.Now(),
	})

	return nil
}

// LeaveRoom 离开房间
func (h *Hub) LeaveRoom(client *Client, roomID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.leaveRoomUnsafe(client, roomID)
}

// leaveRoomUnsafe 离开房间（不加锁版本）
func (h *Hub) leaveRoomUnsafe(client *Client, roomID string) {
	if room, exists := h.rooms[roomID]; exists {
		if _, inRoom := room[client]; inRoom {
			delete(room, client)
			delete(h.userRooms[client.UserID], roomID)

			// 如果房间为空，删除房间
			if len(room) == 0 {
				delete(h.rooms, roomID)
			} else {
				// 通知房间内其他用户
				h.broadcastToRoomUnsafe(roomID, "user_left", map[string]interface{}{
					"user_id":   client.UserID,
					"socket_id": client.ID,
					"room_id":   roomID,
					"timestamp": time.Now(),
				}, nil)
			}

			log.Printf("用户 %d 离开房间 %s", client.UserID, roomID)
		}
	}

	if client.CurrentRoom == roomID {
		client.CurrentRoom = ""
	}
}

// BroadcastToRoom 向房间广播消息
func (h *Hub) BroadcastToRoom(roomID, event string, data interface{}) {
	message := &BroadcastMessage{
		RoomID: roomID,
		Event:  event,
		Data:   data,
	}

	select {
	case h.broadcast <- message:
	default:
		log.Printf("广播队列已满，丢弃消息: room=%s, event=%s", roomID, event)
	}
}

// broadcastMessage 处理广播消息
func (h *Hub) broadcastMessage(message *BroadcastMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	h.broadcastToRoomUnsafe(message.RoomID, message.Event, message.Data, message.Exclude)
}

// broadcastToRoomUnsafe 向房间广播消息（不加锁版本）
func (h *Hub) broadcastToRoomUnsafe(roomID, event string, data interface{}, exclude *Client) {
	room, exists := h.rooms[roomID]
	if !exists {
		return
	}

	for client := range room {
		if exclude != nil && client == exclude {
			continue
		}

		client.Send(event, data)
	}
}

// getRoomUsersUnsafe 获取房间用户列表（不加锁版本）
func (h *Hub) getRoomUsersUnsafe(roomID string) []map[string]interface{} {
	room, exists := h.rooms[roomID]
	if !exists {
		return []map[string]interface{}{}
	}

	users := make([]map[string]interface{}, 0, len(room))
	for client := range room {
		users = append(users, map[string]interface{}{
			"user_id":   client.UserID,
			"socket_id": client.ID,
		})
	}

	return users
}

// GetRoomUsers 获取房间用户列表
func (h *Hub) GetRoomUsers(roomID string) []int64 {
	h.mu.RLock()
	defer h.mu.RUnlock()

	room, exists := h.rooms[roomID]
	if !exists {
		return []int64{}
	}

	userIDs := make([]int64, 0, len(room))
	for client := range room {
		userIDs = append(userIDs, client.UserID)
	}

	return userIDs
}

// cleanup 定期清理
func (h *Hub) cleanup() {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 清理空房间
	for roomID, room := range h.rooms {
		if len(room) == 0 {
			delete(h.rooms, roomID)
		}
	}
}

// GetStats 获取统计信息
func (h *Hub) GetStats() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return map[string]interface{}{
		"total_clients": len(h.clients),
		"total_rooms":   len(h.rooms),
		"timestamp":     time.Now(),
	}
}
