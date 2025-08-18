package document

import (
	"context"
	"fmt"
	"strings"
	"time"

	"DOC/domain"
)

// documentService 文档业务逻辑实现
// 实现 domain.DocumentUsecase 接口，提供文档的核心业务功能
type documentService struct {
	documentRepo    domain.DocumentRepository        // 文档仓储
	shareUsecase    domain.DocumentShareUsecase      // 分享子域
	permUsecase     domain.DocumentPermissionUsecase // 权限子域
	favoriteUsecase domain.DocumentFavoriteUsecase   // 收藏子域
	userRepo        domain.UserRepository            // 用户仓储（用于验证用户存在性）
}

// NewDocumentService 创建新的文档业务服务实例
// 注入所需的依赖项，包括仓储和子域服务
func NewDocumentService(
	documentRepo domain.DocumentRepository,
	shareUsecase domain.DocumentShareUsecase,
	permUsecase domain.DocumentPermissionUsecase,
	favoriteUsecase domain.DocumentFavoriteUsecase,
	userRepo domain.UserRepository,
) domain.DocumentUsecase {
	return &documentService{
		documentRepo:    documentRepo,
		shareUsecase:    shareUsecase,
		permUsecase:     permUsecase,
		favoriteUsecase: favoriteUsecase,
		userRepo:        userRepo,
	}
}

// === 文档管理方法 ===

// CreateDocument 创建新文档
// 支持创建文件和文件夹，进行必要的权限检查和数据验证
func (d *documentService) CreateDocument(ctx context.Context, userID int64, title, content string, docType domain.DocumentType, parentID, spaceID *int64, sortOrder int, isStarred bool) (*domain.Document, error) {
	// 1. 验证用户是否存在
	if _, err := d.userRepo.GetByID(ctx, userID); err != nil {
		return nil, domain.ErrUserNotFound
	}

	// 2. 验证输入参数
	if strings.TrimSpace(title) == "" {
		return nil, domain.ErrInvalidDocumentTitle
	}

	// 3. 如果指定了父文档，验证父文档的有效性
	if parentID != nil {
		parentDoc, err := d.documentRepo.GetByID(ctx, *parentID)
		if err != nil {
			return nil, domain.ErrDocumentNotFound
		}

		// 检查父文档是否为文件夹
		if !parentDoc.CanBeParent() {
			return nil, domain.ErrInvalidDocumentType
		}

		// 检查对父文档的访问权限（至少需要查看权限）
		hasAccess, err := d.CheckDocumentAccess(ctx, userID, *parentID, domain.PermissionEdit)
		if err != nil {
			return nil, err
		}
		if !hasAccess {
			return nil, domain.ErrPermissionDenied
		}
	}

	// 4. 创建文档实体
	document := &domain.Document{
		Title:     strings.TrimSpace(title),
		Content:   content,
		Type:      docType,
		Status:    domain.DocumentStatusActive,
		ParentID:  parentID,
		SpaceID:   spaceID,
		OwnerID:   userID,
		SortOrder: sortOrder,
		IsStarred: isStarred,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 5. 验证文档实体
	if err := document.Validate(); err != nil {
		return nil, err
	}

	// 6. 保存到数据库
	if err := d.documentRepo.Store(ctx, document); err != nil {
		return nil, fmt.Errorf("failed to create document: %w", err)
	}

	return document, nil
}

// GetDocument 获取文档详情
// 包含权限检查，确保用户有权限访问该文档
func (d *documentService) GetDocument(ctx context.Context, userID, documentID int64) (*domain.Document, error) {
	// 1. 获取文档
	document, err := d.documentRepo.GetByID(ctx, documentID)
	if err != nil {
		return nil, err
	}

	// 2. 验证文档操作权限
	if err := domain.ValidateDocumentOperation(userID, documentID, "view", document, domain.PermissionView); err != nil {
		if err == domain.ErrDocumentNotFound {
			return nil, err // 软删除的文档对外表现为不存在
		}
	}

	// 3. 检查访问权限
	hasAccess, err := d.CheckDocumentAccess(ctx, userID, documentID, domain.PermissionView)
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, domain.ErrPermissionDenied
	}

	return document, nil
}

// UpdateDocument 更新文档信息
// 支持更新标题、类型、父目录、排序和星标状态
func (d *documentService) UpdateDocument(ctx context.Context, userID, documentID int64, title string, docType *domain.DocumentType, parentID *int64, sortOrder *int, isStarred *bool) (*domain.Document, error) {
	// 1. 获取现有文档
	document, err := d.documentRepo.GetByID(ctx, documentID)
	if err != nil {
		return nil, err
	}

	// 2. 检查编辑权限
	hasAccess, err := d.CheckDocumentAccess(ctx, userID, documentID, domain.PermissionEdit)
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, domain.ErrPermissionDenied
	}

	// 3. 验证文档操作
	if err := domain.ValidateDocumentOperation(userID, documentID, "edit", document, domain.PermissionEdit); err != nil {
		return nil, err
	}

	// 4. 更新字段
	needsUpdate := false

	if strings.TrimSpace(title) != "" && title != document.Title {
		document.Title = strings.TrimSpace(title)
		needsUpdate = true
	}

	if docType != nil && *docType != document.Type {
		document.Type = *docType
		needsUpdate = true
	}

	if parentID != nil && (document.ParentID == nil || *parentID != *document.ParentID) {
		// 验证层级关系（防止循环引用）
		allDocs, err := d.documentRepo.GetByOwner(ctx, userID, false)
		if err != nil {
			return nil, err
		}
		if err := domain.ValidateDocumentHierarchy(document, parentID, allDocs); err != nil {
			return nil, err
		}
		document.ParentID = parentID
		needsUpdate = true
	}

	if sortOrder != nil && *sortOrder != document.SortOrder {
		document.SortOrder = *sortOrder
		needsUpdate = true
	}

	if isStarred != nil && *isStarred != document.IsStarred {
		document.IsStarred = *isStarred
		needsUpdate = true
	}

	// 5. 如果有变更，执行更新
	if needsUpdate {
		if err := d.documentRepo.Update(ctx, document); err != nil {
			return nil, fmt.Errorf("failed to update document: %w", err)
		}
	}

	return document, nil
}

// DeleteDocument 删除文档（软删除）
// 只有文档所有者或具有管理权限的用户才能删除文档
func (d *documentService) DeleteDocument(ctx context.Context, userID, documentID int64) error {
	// 1. 获取文档
	document, err := d.documentRepo.GetByID(ctx, documentID)
	if err != nil {
		return err
	}

	// 2. 验证删除权限
	if err := domain.ValidateDocumentOperation(userID, documentID, "delete", document, domain.PermissionManage); err != nil {
		return err
	}

	// 3. 执行软删除
	return d.documentRepo.SoftDelete(ctx, documentID)
}

// RestoreDocument 恢复已删除的文档
// 只有文档所有者才能恢复文档
func (d *documentService) RestoreDocument(ctx context.Context, userID, documentID int64) error {
	// 1. 获取文档（包括已删除的）
	document, err := d.documentRepo.GetByID(ctx, documentID)
	if err != nil {
		return err
	}

	// 2. 验证恢复权限（只有所有者可以恢复）
	if err := domain.ValidateDocumentOperation(userID, documentID, "restore", document, domain.PermissionFull); err != nil {
		return err
	}

	// 3. 恢复文档
	document.Restore()
	return d.documentRepo.Update(ctx, document)
}

// === 文档内容管理方法 ===

// UpdateDocumentContent 更新文档内容
func (d *documentService) UpdateDocumentContent(ctx context.Context, userID, documentID int64, content string) error {
	// 1. 检查文档访问权限
	hasAccess, err := d.CheckDocumentAccess(ctx, userID, documentID, domain.PermissionEdit)
	if err != nil {
		return err
	}
	if !hasAccess {
		return domain.ErrPermissionDenied
	}

	// 2. 更新内容
	return d.documentRepo.UpdateContent(ctx, documentID, content)
}

// GetDocumentContent 获取文档内容
func (d *documentService) GetDocumentContent(ctx context.Context, userID, documentID int64) (string, error) {
	// 1. 检查文档访问权限
	hasAccess, err := d.CheckDocumentAccess(ctx, userID, documentID, domain.PermissionView)
	if err != nil {
		return "", err
	}
	if !hasAccess {
		return "", domain.ErrPermissionDenied
	}

	// 2. 获取内容
	return d.documentRepo.GetContent(ctx, documentID)
}

// === 文档查询方法 ===

// GetMyDocuments 获取用户的文档列表
func (d *documentService) GetMyDocuments(ctx context.Context, userID int64, parentID *int64, includeDeleted bool) ([]*domain.Document, error) {
	return d.documentRepo.GetByParent(ctx, parentID, userID)
}

// GetDocumentTree 获取文档树结构
func (d *documentService) GetDocumentTree(ctx context.Context, userID int64, rootID *int64) ([]*domain.Document, error) {
	return d.documentRepo.GetDocumentTree(ctx, rootID, userID)
}

// SearchDocuments 搜索文档
func (d *documentService) SearchDocuments(ctx context.Context, userID int64, keyword string, docType *domain.DocumentType, limit, offset int) ([]*domain.Document, error) {
	if strings.TrimSpace(keyword) == "" {
		return []*domain.Document{}, nil
	}

	return d.documentRepo.SearchDocuments(ctx, userID, keyword, docType, limit, offset)
}

// GetStarredDocuments 获取星标文档
func (d *documentService) GetStarredDocuments(ctx context.Context, userID int64) ([]*domain.Document, error) {
	return d.documentRepo.GetStarredDocuments(ctx, userID)
}

// GetRecentDocuments 获取最近访问的文档
func (d *documentService) GetRecentDocuments(ctx context.Context, userID int64, limit int) ([]*domain.Document, error) {
	return d.documentRepo.GetRecentDocuments(ctx, userID, limit)
}

// === 文档操作方法 ===

// MoveDocument 移动文档到新的父目录
func (d *documentService) MoveDocument(ctx context.Context, userID, documentID int64, newParentID *int64) error {
	// 1. 获取文档
	document, err := d.documentRepo.GetByID(ctx, documentID)
	if err != nil {
		return err
	}

	// 2. 验证移动权限
	if err := domain.ValidateDocumentOperation(userID, documentID, "move", document, domain.PermissionManage); err != nil {
		return err
	}

	// 3. 验证层级关系
	allDocs, err := d.documentRepo.GetByOwner(ctx, userID, false)
	if err != nil {
		return err
	}
	if err := domain.ValidateDocumentHierarchy(document, newParentID, allDocs); err != nil {
		return err
	}

	// 4. 执行移动
	return d.documentRepo.MoveDocument(ctx, documentID, newParentID)
}

// ToggleStarDocument 切换文档星标状态
func (d *documentService) ToggleStarDocument(ctx context.Context, userID, documentID int64) (bool, error) {
	// 1. 检查访问权限
	hasAccess, err := d.CheckDocumentAccess(ctx, userID, documentID, domain.PermissionView)
	if err != nil {
		return false, err
	}
	if !hasAccess {
		return false, domain.ErrPermissionDenied
	}

	// 2. 获取当前状态
	document, err := d.documentRepo.GetByID(ctx, documentID)
	if err != nil {
		return false, err
	}

	// 3. 切换星标状态
	newStarred := !document.IsStarred
	if err := d.documentRepo.ToggleStar(ctx, documentID, userID, newStarred); err != nil {
		return false, err
	}

	return newStarred, nil
}

// DuplicateDocument 复制文档
func (d *documentService) DuplicateDocument(ctx context.Context, userID, documentID int64, newTitle string) (*domain.Document, error) {
	// 1. 获取原文档
	originalDoc, err := d.documentRepo.GetByID(ctx, documentID)
	if err != nil {
		return nil, err
	}

	// 2. 检查访问权限
	hasAccess, err := d.CheckDocumentAccess(ctx, userID, documentID, domain.PermissionView)
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, domain.ErrPermissionDenied
	}

	// 3. 获取原文档内容
	content, err := d.documentRepo.GetContent(ctx, documentID)
	if err != nil {
		return nil, err
	}

	// 4. 创建新文档
	if strings.TrimSpace(newTitle) == "" {
		newTitle = originalDoc.Title + " - 副本"
	}

	return d.CreateDocument(ctx, userID, newTitle, content, originalDoc.Type, originalDoc.ParentID, originalDoc.SpaceID, 0, false)
}

// === 批量操作方法 ===

// BatchDeleteDocuments 批量删除文档
func (d *documentService) BatchDeleteDocuments(ctx context.Context, userID int64, documentIDs []int64) error {
	// 1. 验证批量操作参数
	if err := domain.ValidateBatchOperation(userID, documentIDs, "delete"); err != nil {
		return err
	}

	// 2. 逐个验证权限并删除
	for _, docID := range documentIDs {
		if err := d.DeleteDocument(ctx, userID, docID); err != nil {
			return fmt.Errorf("failed to delete document %d: %w", docID, err)
		}
	}

	return nil
}

// BatchMoveDocuments 批量移动文档
func (d *documentService) BatchMoveDocuments(ctx context.Context, userID int64, documentIDs []int64, newParentID *int64) error {
	// 1. 验证批量操作参数
	if err := domain.ValidateBatchOperation(userID, documentIDs, "move"); err != nil {
		return err
	}

	// 2. 逐个验证权限并移动
	for _, docID := range documentIDs {
		if err := d.MoveDocument(ctx, userID, docID, newParentID); err != nil {
			return fmt.Errorf("failed to move document %d: %w", docID, err)
		}
	}

	return nil
}

// === 权限检查方法 ===

// CheckDocumentAccess 检查文档访问权限
// 委托给权限子域进行处理
func (d *documentService) CheckDocumentAccess(ctx context.Context, userID, documentID int64, permission domain.Permission) (bool, error) {
	// 1. 获取文档信息
	document, err := d.documentRepo.GetByID(ctx, documentID)
	if err != nil {
		return false, err
	}

	// 2. 检查是否为文档所有者
	if userID == document.OwnerID {
		return true, nil
	}

	// 3. 委托给权限子域检查
	return d.permUsecase.CheckPermission(ctx, documentID, userID, permission)
}
