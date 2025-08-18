package domain

import (
	"context"
	"time"
)

// Permission 权限类型枚举
// 描述用户对文档的能力边界，用于分享、协作、编辑等多处判定
type Permission string

const (
	PermissionView    Permission = "VIEW"    // 查看
	PermissionComment Permission = "COMMENT" // 评论
	PermissionEdit    Permission = "EDIT"    // 编辑
	PermissionManage  Permission = "MANAGE"  // 管理
	PermissionFull    Permission = "FULL"    // 完全控制
)

// DocumentPermission 文档权限实体
// 表达“用户-文档-权限”的直接授权关系
type DocumentPermission struct {
	ID         int64      `json:"id" gorm:"primaryKey;autoIncrement"`
	DocumentID int64      `json:"document_id" gorm:"not null;index"`
	UserID     int64      `json:"user_id" gorm:"not null;index"`
	Permission Permission `json:"permission" gorm:"type:varchar(20);not null"`
	GrantedBy  int64      `json:"granted_by" gorm:"not null"`
	GrantedAt  time.Time  `json:"granted_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time  `json:"updated_at" gorm:"autoUpdateTime"`

	// 关联数据（读取场景使用）
	Document      *Document `json:"document,omitempty" gorm:"foreignKey:DocumentID"`
	User          *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
	GrantedByUser *User     `json:"granted_by_user,omitempty" gorm:"foreignKey:GrantedBy"`
}

// === 行为方法（领域规则） ===

// Validate 校验权限实体（最小合法授权）
func (dp *DocumentPermission) Validate() error {
	if dp.DocumentID <= 0 {
		return ErrInvalidDocument
	}
	if dp.UserID <= 0 {
		return ErrInvalidUser
	}
	if dp.GrantedBy <= 0 {
		return ErrInvalidUser
	}
	if !dp.isValidPermission() {
		return ErrInvalidPermission
	}
	return nil
}

// isValidPermission 枚举合法性
func (dp *DocumentPermission) isValidPermission() bool {
	valid := []Permission{PermissionView, PermissionComment, PermissionEdit, PermissionManage, PermissionFull}
	for _, p := range valid {
		if dp.Permission == p {
			return true
		}
	}
	return false
}

// CanView 是否可查看
func (dp *DocumentPermission) CanView() bool { return true }

// CanComment 是否可评论
func (dp *DocumentPermission) CanComment() bool {
	return dp.Permission == PermissionComment || dp.Permission == PermissionEdit || dp.Permission == PermissionManage || dp.Permission == PermissionFull
}

// CanEdit 是否可编辑
func (dp *DocumentPermission) CanEdit() bool {
	return dp.Permission == PermissionEdit || dp.Permission == PermissionManage || dp.Permission == PermissionFull
}

// CanManage 是否可管理
func (dp *DocumentPermission) CanManage() bool {
	return dp.Permission == PermissionManage || dp.Permission == PermissionFull
}

// CanFullControl 是否完全控制
func (dp *DocumentPermission) CanFullControl() bool { return dp.Permission == PermissionFull }

// === 仓储接口（持久化端口） ===

// DocumentPermissionRepository 权限仓储接口
// 对权限授权的持久化与查询
type DocumentPermissionRepository interface {
	// 基本 CRUD
	Store(ctx context.Context, permission *DocumentPermission) error
	GetByID(ctx context.Context, id int64) (*DocumentPermission, error)
	Update(ctx context.Context, permission *DocumentPermission) error
	Delete(ctx context.Context, id int64) error

	// 查询
	GetByDocument(ctx context.Context, documentID int64) ([]*DocumentPermission, error)
	GetByUser(ctx context.Context, userID int64) ([]*DocumentPermission, error)
	GetUserPermission(ctx context.Context, documentID, userID int64) (*DocumentPermission, error)

	// 权限计算/检查
	CheckPermission(ctx context.Context, documentID, userID int64, permission Permission) (bool, error)
	GetUserDocumentsWithPermission(ctx context.Context, userID int64, permission Permission) ([]*Document, error)

	// 批量授权
	BatchGrantPermission(ctx context.Context, documentID int64, userIDs []int64, permission Permission, grantedBy int64) error
	BatchRevokePermission(ctx context.Context, documentID int64, userIDs []int64) error
}

// === 用例接口（应用服务端口） ===

// DocumentPermissionUsecase 权限用例接口
// 面向应用层，封装权限的授权、撤销、查询与检查
type DocumentPermissionUsecase interface {
	// 授权管理
	GrantPermission(ctx context.Context, userID, documentID, targetUserID int64, permission Permission) error
	RevokePermission(ctx context.Context, userID, documentID, targetUserID int64) error
	UpdatePermission(ctx context.Context, userID, documentID, targetUserID int64, permission Permission) error

	// 查询
	GetDocumentPermissions(ctx context.Context, userID, documentID int64) ([]*DocumentPermission, error)
	GetUserPermission(ctx context.Context, documentID, userID int64) (*DocumentPermission, error)
	GetUserDocumentsWithPermission(ctx context.Context, userID int64, permission Permission) ([]*Document, error)

	// 检查
	CheckPermission(ctx context.Context, documentID, userID int64, permission Permission) (bool, error)
	CanAccessDocument(ctx context.Context, documentID, userID int64) (bool, Permission, error)

	// 批量授权
	BatchGrantPermission(ctx context.Context, userID, documentID int64, targetUserIDs []int64, permission Permission) error
	BatchRevokePermission(ctx context.Context, userID, documentID int64, targetUserIDs []int64) error
}
