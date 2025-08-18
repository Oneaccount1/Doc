package domain

import (
	"context"
	"time"
)

// DocumentAggregateUsecase 文档聚合服务接口
// 作为文档聚合根的协调服务，整合文档的核心操作、分享、权限、收藏等功能
type DocumentAggregateUsecase interface {
	// === 文档核心操作（委托给DocumentUsecase） ===
	CreateDocument(ctx context.Context, userID int64, title, content string, docType DocumentType, parentID, spaceID *int64, sortOrder int, isStarred bool) (*Document, error)
	GetDocument(ctx context.Context, userID, documentID int64) (*Document, error)
	UpdateDocument(ctx context.Context, userID, documentID int64, title string, docType *DocumentType, parentID *int64, sortOrder *int, isStarred *bool) (*Document, error)
	DeleteDocument(ctx context.Context, userID, documentID int64) error
	RestoreDocument(ctx context.Context, userID, documentID int64) error

	// 文档内容操作
	UpdateDocumentContent(ctx context.Context, userID, documentID int64, content string) error
	GetDocumentContent(ctx context.Context, userID, documentID int64) (string, error)

	// 文档查询与搜索
	GetMyDocuments(ctx context.Context, userID int64, parentID *int64, includeDeleted bool) ([]*Document, error)
	GetDocumentTree(ctx context.Context, userID int64, rootID *int64) ([]*Document, error)
	SearchDocuments(ctx context.Context, userID int64, keyword string, docType *DocumentType, limit, offset int) ([]*DocumentSearchResult, error)
	GetRecentDocuments(ctx context.Context, userID int64, limit int) ([]*Document, error)

	// 文档操作
	MoveDocument(ctx context.Context, userID, documentID int64, newParentID *int64) error
	DuplicateDocument(ctx context.Context, userID, documentID int64, newTitle string) (*Document, error)

	// 批量操作
	BatchDeleteDocuments(ctx context.Context, userID int64, documentIDs []int64) error
	BatchMoveDocuments(ctx context.Context, userID int64, documentIDs []int64, newParentID *int64) error

	// === 文档分享操作（委托给DocumentShareUsecase） ===
	ShareDocument(ctx context.Context, userID, documentID int64, permission Permission, password string, expiresAt *time.Time, shareWithUserIDs []int64) (*DocumentShare, error)
	UpdateShareLink(ctx context.Context, userID, shareID int64, permission *Permission, password *string, expiresAt *time.Time) (*DocumentShare, error)
	DeleteShareLink(ctx context.Context, userID, shareID int64) error
	GetDocumentShares(ctx context.Context, userID, documentID int64) ([]*DocumentShare, error)
	GetMySharedDocuments(ctx context.Context, userID int64) ([]*DocumentShare, error)

	// 分享访问
	GetSharedDocument(ctx context.Context, linkID, password string, accessIP string) (*DocumentAccessInfo, error)
	ValidateShareAccess(ctx context.Context, linkID, password string) (*DocumentShare, error)

	// === 文档权限操作（委托给DocumentPermissionUsecase） ===
	GrantDocumentPermission(ctx context.Context, userID, documentID, targetUserID int64, permission Permission) error
	RevokeDocumentPermission(ctx context.Context, userID, documentID, targetUserID int64) error
	UpdateDocumentPermission(ctx context.Context, userID, documentID, targetUserID int64, permission Permission) error
	CheckDocumentPermission(ctx context.Context, userID, documentID int64, permission Permission) (bool, error)
	GetDocumentPermissions(ctx context.Context, userID, documentID int64) ([]*DocumentPermission, error)

	// 批量权限操作
	BatchGrantPermission(ctx context.Context, userID, documentID int64, targetUserIDs []int64, permission Permission) error
	BatchRevokePermission(ctx context.Context, userID, documentID int64, targetUserIDs []int64) error

	// === 文档收藏操作（委托给DocumentFavoriteUsecase） ===
	ToggleDocumentFavorite(ctx context.Context, userID, documentID int64) (bool, error)
	SetFavoriteCustomTitle(ctx context.Context, userID, documentID int64, customTitle string) error
	RemoveDocumentFavorite(ctx context.Context, userID, documentID int64) error
	GetFavoriteDocuments(ctx context.Context, userID int64) ([]*DocumentFavorite, error)
	IsFavoriteDocument(ctx context.Context, userID, documentID int64) (bool, error)

	// === 聚合根级别的复合操作 ===
	GetDocumentWithAccessInfo(ctx context.Context, userID, documentID int64) (*DocumentAccessInfo, error)
	GetDocumentFullInfo(ctx context.Context, userID, documentID int64) (*DocumentFullInfo, error)
	CheckDocumentAccess(ctx context.Context, userID, documentID int64, requiredPermission Permission) (*DocumentAccessInfo, error)
}

// === 聚合根数据传输对象 ===

// DocumentAccessInfo 文档访问信息
// 对应API中的DocumentAccessResponseDto，用于检查用户对文档的访问权限
type DocumentAccessInfo struct {
	User       *User      `json:"user"`
	Document   *Document  `json:"document"`
	IsOwner    bool       `json:"isOwner"`
	Permission Permission `json:"permission"`
	CanEdit    bool       `json:"canEdit"`
	CanShare   bool       `json:"canShare"`
	CanManage  bool       `json:"canManage"`
}

// DocumentFullInfo 文档完整信息
// 包含文档本身及其相关的分享、权限、收藏状态等完整信息
type DocumentFullInfo struct {
	Document     *Document             `json:"document"`
	Owner        *User                 `json:"owner"`
	AccessInfo   *DocumentAccessInfo   `json:"accessInfo"`
	Shares       []*DocumentShare      `json:"shares,omitempty"`
	Permissions  []*DocumentPermission `json:"permissions,omitempty"`
	IsFavorite   bool                  `json:"isFavorite"`
	FavoriteInfo *DocumentFavorite     `json:"favoriteInfo,omitempty"`
	Children     []*Document           `json:"children,omitempty"`   // 如果是文件夹
	Breadcrumb   []*Document           `json:"breadcrumb,omitempty"` // 面包屑导航
}

// DocumentSearchResult 文档搜索结果
// 对应API中的DocumentSearchItemDto
type DocumentSearchResult struct {
	ID         int64        `json:"id"`
	Title      string       `json:"title"`
	Type       DocumentType `json:"type"`
	IsStarred  bool         `json:"is_starred"`
	IsFavorite bool         `json:"is_favorite"`
	UpdatedAt  time.Time    `json:"updated_at"`
	Owner      *User        `json:"owner,omitempty"`
	Permission Permission   `json:"permission"`
	ParentPath string       `json:"parent_path,omitempty"` // 父路径，用于搜索结果显示
	MatchScore float64      `json:"match_score,omitempty"` // 匹配分数
}

// DocumentOperationResult 文档操作结果
// 用于返回操作结果和相关的业务信息
type DocumentOperationResult struct {
	Success   bool        `json:"success"`
	Document  *Document   `json:"document,omitempty"`
	Message   string      `json:"message,omitempty"`
	ErrorCode string      `json:"error_code,omitempty"`
	Metadata  interface{} `json:"metadata,omitempty"` // 额外的元数据
}

// === 聚合根业务规则与方法 ===

// CanUserAccessDocument 检查用户是否可以访问文档
// 这是聚合根级别的业务规则，需要协调权限检查
func CanUserAccessDocument(userID, documentID int64, requiredPermission Permission, ownerID int64, permissions []*DocumentPermission) bool {
	// 1. 检查是否为文档所有者
	if userID == ownerID {
		return true
	}

	// 2. 检查显式权限
	for _, perm := range permissions {
		if perm.UserID == userID {
			return IsPermissionSufficient(perm.Permission, requiredPermission)
		}
	}

	return false
}

// IsPermissionSufficient 检查权限是否足够
func IsPermissionSufficient(userPermission, requiredPermission Permission) bool {
	permissionLevels := map[Permission]int{
		PermissionView:    1,
		PermissionComment: 2,
		PermissionEdit:    3,
		PermissionManage:  4,
		PermissionFull:    5,
	}

	userLevel, userExists := permissionLevels[userPermission]
	requiredLevel, requiredExists := permissionLevels[requiredPermission]

	if !userExists || !requiredExists {
		return false
	}

	return userLevel >= requiredLevel
}

// ValidateDocumentOperation 验证文档操作的业务规则
func ValidateDocumentOperation(userID, documentID int64, operation string, document *Document, permission Permission) error {
	if document == nil {
		return ErrDocumentNotFound
	}

	// 检查文档状态
	if !document.IsActive() && operation != "restore" {
		return ErrDocumentNotFound // 软删除的文档对外表现为不存在
	}

	// 基于操作类型检查权限
	switch operation {
	case "read", "view":
		if !IsPermissionSufficient(permission, PermissionView) {
			return ErrPermissionDenied
		}
	case "edit", "update_content":
		if !IsPermissionSufficient(permission, PermissionEdit) {
			return ErrPermissionDenied
		}
	case "delete", "move", "duplicate":
		if userID != document.OwnerID && !IsPermissionSufficient(permission, PermissionManage) {
			return ErrPermissionDenied
		}
	case "share", "grant_permission":
		if userID != document.OwnerID && !IsPermissionSufficient(permission, PermissionManage) {
			return ErrPermissionDenied
		}
	case "restore":
		if userID != document.OwnerID {
			return ErrPermissionDenied
		}
	default:
		return ErrInvalidBatchRequest
	}

	return nil
}

// ValidateDocumentHierarchy 验证文档层级关系
func ValidateDocumentHierarchy(document *Document, newParentID *int64, allDocuments []*Document) error {
	if newParentID == nil {
		return nil // 移动到根目录
	}

	// 查找新父文档
	var newParent *Document
	for _, doc := range allDocuments {
		if doc.ID == *newParentID {
			newParent = doc
			break
		}
	}

	if newParent == nil {
		return ErrDocumentNotFound
	}

	// 检查新父文档是否为文件夹
	if !newParent.CanBeParent() {
		return ErrInvalidDocumentType
	}

	// 检查是否会形成循环引用
	if wouldCreateCycle(document.ID, *newParentID, allDocuments) {
		return ErrConflict
	}

	return nil
}

// wouldCreateCycle 检查移动操作是否会创建循环引用
func wouldCreateCycle(documentID, newParentID int64, allDocuments []*Document) bool {
	// 从新父文档开始，向上遍历父文档链
	currentID := newParentID
	visited := make(map[int64]bool)

	for currentID != 0 {
		if visited[currentID] {
			return true // 检测到循环
		}

		if currentID == documentID {
			return true // 将要移动的文档在父链中
		}

		visited[currentID] = true

		// 查找当前文档的父文档
		var found bool
		for _, doc := range allDocuments {
			if doc.ID == currentID {
				if doc.ParentID == nil {
					currentID = 0 // 到达根目录
				} else {
					currentID = *doc.ParentID
				}
				found = true
				break
			}
		}

		if !found {
			break // 文档不存在，停止检查
		}
	}

	return false
}

// ValidateBatchOperation 验证批量操作
func ValidateBatchOperation(userID int64, documentIDs []int64, operation string) error {
	if len(documentIDs) == 0 {
		return ErrInvalidBatchRequest
	}

	// 检查批量操作大小限制
	const maxBatchSize = 100
	if len(documentIDs) > maxBatchSize {
		return ErrBatchSizeExceeded
	}

	// 检查是否有重复ID
	seen := make(map[int64]bool)
	for _, id := range documentIDs {
		if id <= 0 {
			return ErrInvalidDocument
		}
		if seen[id] {
			return ErrInvalidBatchRequest // 重复的文档ID
		}
		seen[id] = true
	}

	return nil
}

// GetDocumentBreadcrumb 获取文档面包屑导航
func GetDocumentBreadcrumb(document *Document, allDocuments []*Document) []*Document {
	if document == nil {
		return nil
	}

	var breadcrumb []*Document
	current := document
	visited := make(map[int64]bool)

	// 向上遍历父文档链，构建面包屑
	for current != nil && current.ParentID != nil {
		if visited[current.ID] {
			break // 检测到循环，停止
		}
		visited[current.ID] = true

		// 查找父文档
		var parent *Document
		for _, doc := range allDocuments {
			if doc.ID == *current.ParentID {
				parent = doc
				break
			}
		}

		if parent != nil {
			breadcrumb = append([]*Document{parent}, breadcrumb...)
			current = parent
		} else {
			break
		}
	}

	return breadcrumb
}

// BuildDocumentAccessInfo 构建文档访问信息
func BuildDocumentAccessInfo(user *User, document *Document, permissions []*DocumentPermission) *DocumentAccessInfo {
	if user == nil || document == nil {
		return nil
	}

	isOwner := user.ID == document.OwnerID
	permission := PermissionView // 默认权限

	if isOwner {
		permission = PermissionFull
	} else {
		// 查找用户的权限
		for _, perm := range permissions {
			if perm.UserID == user.ID && perm.DocumentID == document.ID {
				permission = perm.Permission
				break
			}
		}
	}

	return &DocumentAccessInfo{
		User:       user,
		Document:   document,
		IsOwner:    isOwner,
		Permission: permission,
		CanEdit:    IsPermissionSufficient(permission, PermissionEdit),
		CanShare:   IsPermissionSufficient(permission, PermissionManage),
		CanManage:  IsPermissionSufficient(permission, PermissionManage),
	}
}
