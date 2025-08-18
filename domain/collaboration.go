package domain

import (
	"context"
	"time"
)

// CollaborationSessionStatus 协作会话状态枚举
type CollaborationSessionStatus int

const (
	CollaborationSessionStatusActive   CollaborationSessionStatus = iota // 活跃
	CollaborationSessionStatusInactive                                   // 非活跃
	CollaborationSessionStatusClosed                                     // 已关闭
)

// CollaborationSession 协作会话实体
// 管理文档的实时协作会话
type CollaborationSession struct {
	ID         int64                      `json:"id" gorm:"primaryKey;autoIncrement"`
	DocumentID int64                      `json:"document_id" gorm:"not null;index"`
	RoomID     string                     `json:"room_id" gorm:"type:varchar(100);uniqueIndex;not null"`
	Status     CollaborationSessionStatus `json:"status" gorm:"type:tinyint;default:0;index"`
	MaxUsers   int                        `json:"max_users" gorm:"default:10"`
	CreatedAt  time.Time                  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time                  `json:"updated_at" gorm:"autoUpdateTime"`
	ClosedAt   *time.Time                 `json:"closed_at"`

	// 关联数据
	Document     *Document                 `json:"document,omitempty" gorm:"-"`
	Participants []*CollaborationUser      `json:"participants,omitempty" gorm:"-"`
	Operations   []*CollaborationOperation `json:"operations,omitempty" gorm:"-"`
}

// CollaborationUser 协作用户实体
// 记录参与协作的用户信息
type CollaborationUser struct {
	ID        int64      `json:"id" gorm:"primaryKey;autoIncrement"`
	SessionID int64      `json:"session_id" gorm:"not null;index"`
	UserID    int64      `json:"user_id" gorm:"not null;index"`
	SocketID  string     `json:"socket_id" gorm:"type:varchar(100);index"`
	CursorPos int        `json:"cursor_pos" gorm:"default:0"`
	IsActive  bool       `json:"is_active" gorm:"default:true"`
	JoinedAt  time.Time  `json:"joined_at" gorm:"autoCreateTime"`
	LeftAt    *time.Time `json:"left_at"`

	// 关联数据
	User    *User                 `json:"user,omitempty" gorm:"-"`
	Session *CollaborationSession `json:"session,omitempty" gorm:"-"`
}

// CollaborationOperationType 协作操作类型枚举
type CollaborationOperationType string

const (
	CollaborationOperationTypeInsert CollaborationOperationType = "insert" // 插入
	CollaborationOperationTypeDelete CollaborationOperationType = "delete" // 删除
	CollaborationOperationTypeFormat CollaborationOperationType = "format" // 格式化
	CollaborationOperationTypeCursor CollaborationOperationType = "cursor" // 光标移动
)

// CollaborationOperation 协作操作实体
// 记录协作过程中的所有操作，用于操作转换和冲突解决
type CollaborationOperation struct {
	ID        int64                      `json:"id" gorm:"primaryKey;autoIncrement"`
	SessionID int64                      `json:"session_id" gorm:"not null;index"`
	UserID    int64                      `json:"user_id" gorm:"not null;index"`
	Type      CollaborationOperationType `json:"type" gorm:"type:varchar(20);not null"`
	Position  int                        `json:"position" gorm:"not null"`
	Content   string                     `json:"content" gorm:"type:text"`
	Length    int                        `json:"length" gorm:"default:0"`
	Metadata  string                     `json:"metadata" gorm:"type:json"` // JSON格式的额外数据
	Timestamp time.Time                  `json:"timestamp" gorm:"autoCreateTime"`

	// 关联数据
	User    *User                 `json:"user,omitempty" gorm:"-"`
	Session *CollaborationSession `json:"session,omitempty" gorm:"-"`
}

// Validate 验证协作会话
func (cs *CollaborationSession) Validate() error {
	if cs.DocumentID <= 0 {
		return ErrDocumentIDRequired
	}
	if cs.RoomID == "" {
		return ErrInvalidDocument
	}
	if cs.MaxUsers <= 0 {
		cs.MaxUsers = 10 // 设置默认值
	}
	return nil
}

// IsActive 检查会话是否活跃
func (cs *CollaborationSession) IsActive() bool {
	return cs.Status == CollaborationSessionStatusActive
}

// IsClosed 检查会话是否已关闭
func (cs *CollaborationSession) IsClosed() bool {
	return cs.Status == CollaborationSessionStatusClosed
}

// Close 关闭协作会话
func (cs *CollaborationSession) Close() {
	cs.Status = CollaborationSessionStatusClosed
	now := time.Now()
	cs.ClosedAt = &now
	cs.UpdatedAt = now
}

// Activate 激活协作会话
func (cs *CollaborationSession) Activate() {
	cs.Status = CollaborationSessionStatusActive
	cs.UpdatedAt = time.Now()
}

// Deactivate 停用协作会话
func (cs *CollaborationSession) Deactivate() {
	cs.Status = CollaborationSessionStatusInactive
	cs.UpdatedAt = time.Now()
}

// Validate 验证协作用户
func (cu *CollaborationUser) Validate() error {
	if cu.SessionID <= 0 {
		return ErrInvalidDocument
	}
	if cu.UserID <= 0 {
		return ErrInvalidUser
	}
	return nil
}

// Leave 用户离开协作
func (cu *CollaborationUser) Leave() {
	cu.IsActive = false
	now := time.Now()
	cu.LeftAt = &now
}

// UpdateCursor 更新光标位置
func (cu *CollaborationUser) UpdateCursor(position int) {
	cu.CursorPos = position
}

// Validate 验证协作操作
func (co *CollaborationOperation) Validate() error {
	if co.SessionID <= 0 {
		return ErrInvalidDocument
	}
	if co.UserID <= 0 {
		return ErrInvalidUser
	}
	if co.Position < 0 {
		return ErrInvalidDocument
	}
	return nil
}

// IsInsert 检查是否为插入操作
func (co *CollaborationOperation) IsInsert() bool {
	return co.Type == CollaborationOperationTypeInsert
}

// IsDelete 检查是否为删除操作
func (co *CollaborationOperation) IsDelete() bool {
	return co.Type == CollaborationOperationTypeDelete
}

// CollaborationRepository 协作仓储接口
type CollaborationRepository interface {
	// 会话管理
	StoreSession(ctx context.Context, session *CollaborationSession) error
	GetSessionByID(ctx context.Context, id int64) (*CollaborationSession, error)
	GetSessionByRoomID(ctx context.Context, roomID string) (*CollaborationSession, error)
	GetSessionByDocumentID(ctx context.Context, documentID int64) (*CollaborationSession, error)
	UpdateSession(ctx context.Context, session *CollaborationSession) error
	DeleteSession(ctx context.Context, id int64) error

	// 用户管理
	StoreUser(ctx context.Context, user *CollaborationUser) error
	GetUsersBySessionID(ctx context.Context, sessionID int64) ([]*CollaborationUser, error)
	GetActiveUsersBySessionID(ctx context.Context, sessionID int64) ([]*CollaborationUser, error)
	UpdateUser(ctx context.Context, user *CollaborationUser) error
	RemoveUser(ctx context.Context, sessionID, userID int64) error

	// 操作管理
	StoreOperation(ctx context.Context, operation *CollaborationOperation) error
	GetOperationsBySessionID(ctx context.Context, sessionID int64, offset, limit int) ([]*CollaborationOperation, error)
	GetOperationsAfterTimestamp(ctx context.Context, sessionID int64, timestamp time.Time) ([]*CollaborationOperation, error)

	// 清理操作
	CleanupInactiveSessions(ctx context.Context, inactiveThreshold time.Duration) error
	CleanupOldOperations(ctx context.Context, olderThan time.Time) error
}

// CollaborationUsecase 协作业务逻辑接口
type CollaborationUsecase interface {
	// 会话管理
	CreateSession(ctx context.Context, documentID int64, userID int64) (*CollaborationSession, error)
	GetSession(ctx context.Context, roomID string) (*CollaborationSession, error)
	CloseSession(ctx context.Context, roomID string, userID int64) error

	// 用户管理
	JoinSession(ctx context.Context, roomID string, userID int64, socketID string) (*CollaborationUser, error)
	LeaveSession(ctx context.Context, roomID string, userID int64) error
	GetSessionParticipants(ctx context.Context, roomID string) ([]*CollaborationUser, error)
	UpdateUserCursor(ctx context.Context, roomID string, userID int64, position int) error

	// 操作管理
	ApplyOperation(ctx context.Context, roomID string, userID int64, operation *CollaborationOperation) error
	GetOperations(ctx context.Context, roomID string, afterTimestamp *time.Time) ([]*CollaborationOperation, error)
	SyncDocument(ctx context.Context, roomID string, userID int64, content string) error

	// 权限检查
	CheckCollaborationPermission(ctx context.Context, userID int64, documentID int64) (bool, error)

	// 清理操作
	CleanupInactiveSessions(ctx context.Context) error
}

// CollaborationService 协作服务接口
// 定义实时协作的核心服务接口，由基础设施层实现
type CollaborationService interface {
	// 实时通信
	BroadcastToRoom(ctx context.Context, roomID string, event string, data interface{}) error
	SendToUser(ctx context.Context, userID int64, event string, data interface{}) error

	// 房间管理
	CreateRoom(ctx context.Context, roomID string) error
	JoinRoom(ctx context.Context, roomID string, userID int64, socketID string) error
	LeaveRoom(ctx context.Context, roomID string, userID int64) error
	GetRoomUsers(ctx context.Context, roomID string) ([]int64, error)

	// 操作转换
	TransformOperation(ctx context.Context, operation *CollaborationOperation, concurrentOps []*CollaborationOperation) (*CollaborationOperation, error)

	// 健康检查
	HealthCheck(ctx context.Context) error
}
