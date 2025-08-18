package domain

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"
)

// ShareType 分享类型枚举
// 表示文档被分享的方式：公开链接或私有（指定用户）
type ShareType int

const (
	ShareTypePublic  ShareType = iota // 公开链接
	ShareTypePrivate                  // 私有分享（指定用户）
)

// DocumentShare 文档分享实体
// 对应分享创建/查询接口的核心领域模型，承载分享链路的权限、有效期、统计数据
type DocumentShare struct {
	ID         int64      `json:"id" gorm:"primaryKey;autoIncrement"`
	DocumentID int64      `json:"document_id" gorm:"not null;index"`
	ShareType  ShareType  `json:"share_type" gorm:"type:tinyint;not null;default:0"` // 分享类型
	LinkID     string     `json:"link_id" gorm:"type:varchar(100);uniqueIndex"`      // 分享链接ID（对外暴露）
	Permission Permission `json:"permission" gorm:"type:varchar(20);not null"`       // 分享权限（VIEW/COMMENT/EDIT/MANAGE/FULL）
	Password   string     `json:"password" gorm:"type:varchar(255)"`                 // 访问密码（可选）
	ExpiresAt  *time.Time `json:"expires_at" gorm:"index"`                           // 过期时间（可选）
	CreatedBy  int64      `json:"created_by" gorm:"not null;index"`                  // 创建者ID
	CreatedAt  time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time  `json:"updated_at" gorm:"autoUpdateTime"`

	// 使用统计（便于风控与分析）
	ViewCount    int        `json:"view_count" gorm:"default:0"`            // 查看次数
	LastAccessAt *time.Time `json:"last_access_at"`                         // 最后访问时间
	LastAccessIP string     `json:"last_access_ip" gorm:"type:varchar(45)"` // 最后访问IP

	// 关联数据（读取场景使用，不参与持久化约束）
	Document *Document `json:"document,omitempty" gorm:"foreignKey:DocumentID"`
	Creator  *User     `json:"creator,omitempty" gorm:"foreignKey:CreatedBy"`

	// 指定分享用户（私有分享场景）
	SharedUsers []*DocumentShareUser `json:"shared_users,omitempty" gorm:"foreignKey:ShareID"`
}

// DocumentShareUser 文档分享用户关联
// 仅在 ShareTypePrivate 时生效，表示可访问该分享的具体用户集合
type DocumentShareUser struct {
	ID      int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	ShareID int64     `json:"share_id" gorm:"not null;index"`
	UserID  int64     `json:"user_id" gorm:"not null;index"`
	AddedAt time.Time `json:"added_at" gorm:"autoCreateTime"`

	// 关联数据（读取场景使用）
	Share *DocumentShare `json:"share,omitempty" gorm:"foreignKey:ShareID"`
	User  *User          `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// === 行为方法（领域规则） ===

// Validate 校验分享实体的业务规则
func (ds *DocumentShare) Validate() error {
	if ds.DocumentID <= 0 {
		return ErrInvalidDocument
	}
	if ds.CreatedBy <= 0 {
		return ErrInvalidUser
	}
	if !ds.isValidPermission() {
		return ErrInvalidPermission
	}
	return nil
}

// isValidPermission 校验权限枚举合法性
func (ds *DocumentShare) isValidPermission() bool {
	valid := []Permission{PermissionView, PermissionComment, PermissionEdit, PermissionManage, PermissionFull}
	for _, p := range valid {
		if ds.Permission == p {
			return true
		}
	}
	return false
}

// IsExpired 分享是否过期
func (ds *DocumentShare) IsExpired() bool {
	if ds.ExpiresAt == nil {
		return false
	}
	return ds.ExpiresAt.Before(time.Now())
}

// IsPublic 是否公开分享
func (ds *DocumentShare) IsPublic() bool { return ds.ShareType == ShareTypePublic }

// IsPrivate 是否私有分享
func (ds *DocumentShare) IsPrivate() bool { return ds.ShareType == ShareTypePrivate }

// IsPasswordProtected 是否设置密码保护
func (ds *DocumentShare) IsPasswordProtected() bool { return ds.Password != "" }

// IncrementViewCount 增加访问统计并记录最近访问
func (ds *DocumentShare) IncrementViewCount(accessIP string) {
	ds.ViewCount++
	now := time.Now()
	ds.LastAccessAt = &now
	ds.LastAccessIP = accessIP
}

// GenerateLinkID 生成对外分享链接ID（无语义的随机串）
func (ds *DocumentShare) GenerateLinkID() error {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return err
	}
	ds.LinkID = hex.EncodeToString(bytes)
	return nil
}

// Validate 校验分享用户关联
func (dsu *DocumentShareUser) Validate() error {
	if dsu.ShareID <= 0 {
		return ErrInvalidDocument
	}
	if dsu.UserID <= 0 {
		return ErrInvalidUser
	}
	return nil
}

// === 仓储接口（持久化端口） ===

// DocumentShareRepository 分享仓储接口
// 持久化 DocumentShare 及其关联用户的数据访问契约
type DocumentShareRepository interface {
	// 基本 CRUD
	Store(ctx context.Context, share *DocumentShare) error
	GetByID(ctx context.Context, id int64) (*DocumentShare, error)
	GetByLinkID(ctx context.Context, linkID string) (*DocumentShare, error)
	Update(ctx context.Context, share *DocumentShare) error
	Delete(ctx context.Context, id int64) error

	// 查询
	GetByDocument(ctx context.Context, documentID int64) ([]*DocumentShare, error)
	GetByCreator(ctx context.Context, creatorID int64) ([]*DocumentShare, error)
	GetUserSharedDocuments(ctx context.Context, userID int64) ([]*DocumentShare, error)

	// 关联用户管理（私有分享）
	AddShareUser(ctx context.Context, shareUser *DocumentShareUser) error
	RemoveShareUser(ctx context.Context, shareID, userID int64) error
	GetShareUsers(ctx context.Context, shareID int64) ([]*DocumentShareUser, error)

	// 统计
	IncrementViewCount(ctx context.Context, shareID int64, accessIP string) error
	GetShareStats(ctx context.Context, shareID int64) (*DocumentShare, error)

	// 清理
	CleanupExpiredShares(ctx context.Context) error
}

// === 用例接口（应用服务端口） ===

// DocumentShareUsecase 分享用例接口
// 面向应用层的分享编排与权限映射逻辑
type DocumentShareUsecase interface {
	// 分享管理
	CreateShareLink(ctx context.Context, userID, documentID int64, permission Permission, password string, expiresAt *time.Time, shareWithUserIDs []int64) (*DocumentShare, error)
	UpdateShareLink(ctx context.Context, userID, shareID int64, permission *Permission, password *string, expiresAt *time.Time) (*DocumentShare, error)
	DeleteShareLink(ctx context.Context, userID, shareID int64) error

	// 访问与鉴权
	GetSharedDocument(ctx context.Context, linkID, password string, accessIP string) (*Document, error)
	ValidateShareAccess(ctx context.Context, linkID, password string) (*DocumentShare, error)
	RecordShareAccess(ctx context.Context, shareID int64, accessIP string) error

	// 查询
	GetDocumentShares(ctx context.Context, userID, documentID int64) ([]*DocumentShare, error)
	GetMySharedDocuments(ctx context.Context, userID int64) ([]*DocumentShare, error)
	GetSharedWithMeDocuments(ctx context.Context, userID int64) ([]*DocumentShare, error)

	// 私有分享的用户管理
	AddShareUsers(ctx context.Context, userID, shareID int64, targetUserIDs []int64) error
	RemoveShareUsers(ctx context.Context, userID, shareID int64, targetUserIDs []int64) error
	GetShareUsers(ctx context.Context, userID, shareID int64) ([]*DocumentShareUser, error)

	// 统计
	GetShareStats(ctx context.Context, userID, shareID int64) (*DocumentShare, error)
}
