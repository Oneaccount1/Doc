package domain

import (
	"context"
	"time"
)

// DocumentType 文档类型枚举
type DocumentType string

const (
	DocumentTypeFile   DocumentType = "FILE"   // 文件
	DocumentTypeFolder DocumentType = "FOLDER" // 文件夹
)

// DocumentStatus 文档状态枚举
type DocumentStatus int

const (
	DocumentStatusActive   DocumentStatus = iota // 正常
	DocumentStatusDeleted                        // 已删除（软删除）
	DocumentStatusArchived                       // 已归档
)

// Document 文档实体（对应 CreateDocumentDto/UpdateDocumentDto）
// 这是文档聚合的根实体，包含文档的核心属性和行为
type Document struct {
	ID       int64          `json:"id" gorm:"primaryKey;autoIncrement"`
	Title    string         `json:"title" gorm:"type:varchar(255);not null;index"`
	Content  string         `json:"content" gorm:"type:longtext"`                         // JSON格式的文档内容
	Type     DocumentType   `json:"type" gorm:"type:varchar(20);not null;default:'FILE'"` // 文档类型
	Status   DocumentStatus `json:"status" gorm:"type:tinyint;not null;default:0"`        // 文档状态
	ParentID *int64         `json:"parent_id" gorm:"index"`                               // 父文件夹ID，根目录为nil
	SpaceID  *int64         `json:"space_id" gorm:"index"`                                // 所属空间ID，可选
	OwnerID  int64          `json:"owner_id" gorm:"not null;index"`                       // 文档所有者ID

	// 显示和排序
	SortOrder int  `json:"sort_order" gorm:"default:0"`     // 排序顺序
	IsStarred bool `json:"is_starred" gorm:"default:false"` // 是否星标

	// 时间字段
	CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt *time.Time `json:"deleted_at" gorm:"index"` // 软删除时间

	// 关联数据（不存储在数据库中）
	Owner    *User       `json:"owner,omitempty" gorm:"foreignKey:OwnerID"`
	Parent   *Document   `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	Children []*Document `json:"children,omitempty" gorm:"foreignKey:ParentID"`
	Space    *Space      `json:"space,omitempty" gorm:"foreignKey:SpaceID"`
}

// === 实体方法 ===

// Validate 验证文档实体
func (d *Document) Validate() error {
	if d.Title == "" {
		return ErrInvalidDocumentTitle
	}
	if d.OwnerID <= 0 {
		return ErrInvalidUser
	}
	if !d.isValidType() {
		return ErrInvalidDocumentType
	}
	return nil
}

// isValidType 验证文档类型
func (d *Document) isValidType() bool {
	return d.Type == DocumentTypeFile || d.Type == DocumentTypeFolder
}

// IsActive 检查文档是否激活
func (d *Document) IsActive() bool {
	return d.Status == DocumentStatusActive
}

// IsFolder 检查是否为文件夹
func (d *Document) IsFolder() bool {
	return d.Type == DocumentTypeFolder
}

// IsFile 检查是否为文件
func (d *Document) IsFile() bool {
	return d.Type == DocumentTypeFile
}

// IsRootDocument 检查是否为根目录文档
func (d *Document) IsRootDocument() bool {
	return d.ParentID == nil
}

// CanBeParent 检查是否可以作为父目录
func (d *Document) CanBeParent() bool {
	return d.IsFolder() && d.IsActive()
}

// SoftDelete 软删除文档
func (d *Document) SoftDelete() {
	d.Status = DocumentStatusDeleted
	now := time.Now()
	d.DeletedAt = &now
}

// Archive 归档文档
func (d *Document) Archive() {
	d.Status = DocumentStatusArchived
}

// Restore 恢复文档
func (d *Document) Restore() {
	d.Status = DocumentStatusActive
	d.DeletedAt = nil
}

// === 仓储接口 ===

// DocumentRepository 文档仓储接口
type DocumentRepository interface {
	// 文档基本操作
	Store(ctx context.Context, document *Document) error
	GetByID(ctx context.Context, id int64) (*Document, error)
	Update(ctx context.Context, document *Document) error
	Delete(ctx context.Context, id int64) error
	SoftDelete(ctx context.Context, id int64) error

	// 文档查询
	GetByOwner(ctx context.Context, ownerID int64, includeDeleted bool) ([]*Document, error)
	GetByParent(ctx context.Context, parentID *int64, ownerID int64) ([]*Document, error)
	GetBySpace(ctx context.Context, spaceID int64, ownerID int64) ([]*Document, error)
	GetDocumentTree(ctx context.Context, rootID *int64, ownerID int64) ([]*Document, error)

	// 文档搜索
	SearchDocuments(ctx context.Context, userID int64, keyword string, docType *DocumentType, limit, offset int) ([]*Document, error)
	GetStarredDocuments(ctx context.Context, userID int64) ([]*Document, error)
	GetRecentDocuments(ctx context.Context, userID int64, limit int) ([]*Document, error)

	// 文档内容操作
	UpdateContent(ctx context.Context, id int64, content string) error
	GetContent(ctx context.Context, id int64) (string, error)

	// 文档状态操作
	UpdateStatus(ctx context.Context, id int64, status DocumentStatus) error
	ToggleStar(ctx context.Context, id int64, userID int64, starred bool) error
	MoveDocument(ctx context.Context, id int64, newParentID *int64) error

	// 批量操作
	BatchDelete(ctx context.Context, ids []int64, userID int64) error
	BatchMove(ctx context.Context, ids []int64, newParentID *int64, userID int64) error
}

// === 业务逻辑接口 ===

// DocumentUsecase 文档业务逻辑接口
type DocumentUsecase interface {
	// 文档管理
	CreateDocument(ctx context.Context, userID int64, title, content string, docType DocumentType, parentID, spaceID *int64, sortOrder int, isStarred bool) (*Document, error)
	GetDocument(ctx context.Context, userID, documentID int64) (*Document, error)
	UpdateDocument(ctx context.Context, userID, documentID int64, title string, docType *DocumentType, parentID *int64, sortOrder *int, isStarred *bool) (*Document, error)
	DeleteDocument(ctx context.Context, userID, documentID int64) error
	RestoreDocument(ctx context.Context, userID, documentID int64) error

	// 文档内容管理
	UpdateDocumentContent(ctx context.Context, userID, documentID int64, content string) error
	GetDocumentContent(ctx context.Context, userID, documentID int64) (string, error)

	// 文档查询
	GetMyDocuments(ctx context.Context, userID int64, parentID *int64, includeDeleted bool) ([]*Document, error)
	GetDocumentTree(ctx context.Context, userID int64, rootID *int64) ([]*Document, error)
	SearchDocuments(ctx context.Context, userID int64, keyword string, docType *DocumentType, limit, offset int) ([]*Document, error)
	GetStarredDocuments(ctx context.Context, userID int64) ([]*Document, error)
	GetRecentDocuments(ctx context.Context, userID int64, limit int) ([]*Document, error)

	// 文档操作
	MoveDocument(ctx context.Context, userID, documentID int64, newParentID *int64) error
	ToggleStarDocument(ctx context.Context, userID, documentID int64) (bool, error)
	DuplicateDocument(ctx context.Context, userID, documentID int64, newTitle string) (*Document, error)

	// 批量操作
	BatchDeleteDocuments(ctx context.Context, userID int64, documentIDs []int64) error
	BatchMoveDocuments(ctx context.Context, userID int64, documentIDs []int64, newParentID *int64) error

	// 权限检查 (委托给权限聚合)
	CheckDocumentAccess(ctx context.Context, userID, documentID int64, permission Permission) (bool, error)
}
