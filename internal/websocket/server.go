package websocket

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"DOC/domain"
	"DOC/pkg/jwt"

	"github.com/gin-gonic/gin"
)

// Server WebSocket 服务器
type Server struct {
	hub                  *Hub
	jwtManager           *jwt.JWTManager
	collaborationUsecase domain.CollaborationUsecase
}

// NewServer 创建新的 WebSocket 服务器
func NewServer(
	hub *Hub,
	jwtManager *jwt.JWTManager,
	collaborationUsecase domain.CollaborationUsecase,
) *Server {
	return &Server{
		hub:                  hub,
		jwtManager:           jwtManager,
		collaborationUsecase: collaborationUsecase,
	}
}

// Start 启动 WebSocket 服务器
func (s *Server) Start() {
	s.hub.Start()
	log.Println("WebSocket 服务器已启动")
}

// Stop 停止 WebSocket 服务器
func (s *Server) Stop() {
	s.hub.Stop()
	log.Println("WebSocket 服务器已停止")
}

// HandleWebSocket 处理 WebSocket 连接
func (s *Server) HandleWebSocket(c *gin.Context) {
	// 1. 验证用户身份
	userID, err := s.authenticateUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "认证失败",
			"message": err.Error(),
		})
		return
	}

	// 2. 升级 HTTP 连接为 WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket 升级失败: %v", err)
		return
	}

	// 3. 创建客户端并启动
	client := NewClient(s.hub, conn, userID)
	client.Start()

	log.Printf("WebSocket 连接已建立: UserID=%d, SocketID=%s", userID, client.ID)
}

// authenticateUser 验证用户身份
func (s *Server) authenticateUser(c *gin.Context) (int64, error) {
	// 方式1: 从查询参数获取 token
	token := c.Query("token")

	// 方式2: 从 Header 获取 token
	if token == "" {
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}

	// 方式3: 从 Cookie 获取 token
	if token == "" {
		if cookie, err := c.Cookie("auth_token"); err == nil {
			token = cookie
		}
	}

	if token == "" {
		return 0, fmt.Errorf("缺少认证令牌")
	}

	// 验证 JWT token
	claims, err := s.jwtManager.ValidateToken(token)
	if err != nil {
		return 0, fmt.Errorf("无效的认证令牌: %v", err)
	}

	// 从 claims 中获取用户ID
	userID := claims.UserID
	if userID <= 0 {
		return 0, fmt.Errorf("令牌中缺少有效的用户ID")
	}

	return userID, nil
}

// RegisterRoutes 注册 WebSocket 路由
func (s *Server) RegisterRoutes(router *gin.Engine) {
	// WebSocket 连接端点
	router.GET("/ws", s.HandleWebSocket)

	// WebSocket 状态和管理端点
	wsGroup := router.Group("/api/v1/ws")
	{
		wsGroup.GET("/stats", s.GetStats)
		wsGroup.GET("/rooms/:roomId/users", s.GetRoomUsers)
		wsGroup.POST("/rooms/:roomId/broadcast", s.BroadcastToRoom)
	}
}

// GetStats 获取 WebSocket 统计信息
func (s *Server) GetStats(c *gin.Context) {
	stats := s.hub.GetStats()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// GetRoomUsers 获取房间用户列表
func (s *Server) GetRoomUsers(c *gin.Context) {
	roomID := c.Param("roomId")
	if roomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "房间ID不能为空",
		})
		return
	}

	userIDs := s.hub.GetRoomUsers(roomID)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"room_id": roomID,
			"users":   userIDs,
			"count":   len(userIDs),
		},
	})
}

// BroadcastToRoom 向房间广播消息
func (s *Server) BroadcastToRoom(c *gin.Context) {
	roomID := c.Param("roomId")
	if roomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "房间ID不能为空",
		})
		return
	}

	var req struct {
		Event string      `json:"event" binding:"required"`
		Data  interface{} `json:"data"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数错误",
			"message": err.Error(),
		})
		return
	}

	// 广播消息
	s.hub.BroadcastToRoom(roomID, req.Event, req.Data)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "消息已广播",
	})
}

// GetHub 获取 Hub 实例（供其他服务使用）
func (s *Server) GetHub() *Hub {
	return s.hub
}

// 实现 domain.CollaborationService 接口
func (s *Server) BroadcastToRoomService(ctx context.Context, roomID string, event string, data interface{}) error {
	s.hub.BroadcastToRoom(roomID, event, data)
	return nil
}

func (s *Server) SendToUser(ctx context.Context, userID int64, event string, data interface{}) error {
	// 查找用户的所有连接并发送消息
	// 这里需要扩展 Hub 来支持按用户ID发送消息
	// 暂时通过广播到用户所在的所有房间来实现
	return nil
}

func (s *Server) CreateRoom(ctx context.Context, roomID string) error {
	// WebSocket 房间是动态创建的，无需特殊处理
	return nil
}

func (s *Server) JoinRoom(ctx context.Context, roomID string, userID int64, socketID string) error {
	// 这个操作由客户端主动发起，服务端响应
	return nil
}

func (s *Server) LeaveRoom(ctx context.Context, roomID string, userID int64) error {
	// 这个操作由客户端主动发起，服务端响应
	return nil
}

func (s *Server) GetRoomUsersService(ctx context.Context, roomID string) ([]int64, error) {
	return s.hub.GetRoomUsers(roomID), nil
}

func (s *Server) TransformOperation(ctx context.Context, operation *domain.CollaborationOperation, concurrentOps []*domain.CollaborationOperation) (*domain.CollaborationOperation, error) {
	// 操作转换逻辑，这里需要实现 OT (Operational Transformation) 算法
	// 暂时返回原操作
	return operation, nil
}

func (s *Server) HealthCheck(ctx context.Context) error {
	// 检查 Hub 是否正常运行
	if !s.hub.running {
		return fmt.Errorf("WebSocket Hub 未运行")
	}
	return nil
}
