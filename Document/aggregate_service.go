package document

import (
	"context"
	"time"

	"DOC/domain"
)

// documentAggregateService 文档聚合服务实现
// 作为文档聚合根的协调服务，整合文档核心操作、分享、权限、收藏等功能
// 实现 domain.DocumentAggregateService 接口
type documentAggregateService struct {
	documentUsecase domain.DocumentUsecase           // 文档核心业务
	shareUsecase    domain.DocumentShareUsecase      // 分享子域
	permUsecase     domain.DocumentPermissionUsecase // 权限子域
	favoriteUsecase domain.DocumentFavoriteUsecase   // 收藏子域
	userRepo        domain.UserRepository            // 用户仓储
}

// NewDocumentAggregateService 创建新的文档聚合服务实例
// 聚合所有相关的子域服务，提供统一的文档操作接口
func NewDocumentAggregateService(
	documentUsecase domain.DocumentUsecase,
	shareUsecase domain.DocumentShareUsecase,
	permUsecase domain.DocumentPermissionUsecase,
	favoriteUsecase domain.DocumentFavoriteUsecase,
	userRepo domain.UserRepository,
) domain.DocumentAggregateUsecase {
	return &documentAggregateService{
		documentUsecase: documentUsecase,
		shareUsecase:    shareUsecase,
		permUsecase:     permUsecase,
		favoriteUsecase: favoriteUsecase,
		userRepo:        userRepo,
	}
}

// === 文档核心操作（委托给DocumentUsecase） ===

// CreateDocument 创建文档
// 委托给文档核心业务服务处理
func (s *documentAggregateService) CreateDocument(ctx context.Context, userID int64, title, content string, docType domain.DocumentType, parentID, spaceID *int64, sortOrder int, isStarred bool) (*domain.Document, error) {
	return s.documentUsecase.CreateDocument(ctx, userID, title, content, docType, parentID, spaceID, sortOrder, isStarred)
}

// GetDocument 获取文档详情
func (s *documentAggregateService) GetDocument(ctx context.Context, userID, documentID int64) (*domain.Document, error) {
	return s.documentUsecase.GetDocument(ctx, userID, documentID)
}

// UpdateDocument 更新文档信息
func (s *documentAggregateService) UpdateDocument(ctx context.Context, userID, documentID int64, title string, docType *domain.DocumentType, parentID *int64, sortOrder *int, isStarred *bool) (*domain.Document, error) {
	return s.documentUsecase.UpdateDocument(ctx, userID, documentID, title, docType, parentID, sortOrder, isStarred)
}

// DeleteDocument 删除文档
func (s *documentAggregateService) DeleteDocument(ctx context.Context, userID, documentID int64) error {
	return s.documentUsecase.DeleteDocument(ctx, userID, documentID)
}

// RestoreDocument 恢复文档
func (s *documentAggregateService) RestoreDocument(ctx context.Context, userID, documentID int64) error {
	return s.documentUsecase.RestoreDocument(ctx, userID, documentID)
}

// UpdateDocumentContent 更新文档内容
func (s *documentAggregateService) UpdateDocumentContent(ctx context.Context, userID, documentID int64, content string) error {
	return s.documentUsecase.UpdateDocumentContent(ctx, userID, documentID, content)
}

// GetDocumentContent 获取文档内容
func (s *documentAggregateService) GetDocumentContent(ctx context.Context, userID, documentID int64) (string, error) {
	return s.documentUsecase.GetDocumentContent(ctx, userID, documentID)
}

// GetMyDocuments 获取我的文档列表
func (s *documentAggregateService) GetMyDocuments(ctx context.Context, userID int64, parentID *int64, includeDeleted bool) ([]*domain.Document, error) {
	return s.documentUsecase.GetMyDocuments(ctx, userID, parentID, includeDeleted)
}

// GetDocumentTree 获取文档树结构
func (s *documentAggregateService) GetDocumentTree(ctx context.Context, userID int64, rootID *int64) ([]*domain.Document, error) {
	return s.documentUsecase.GetDocumentTree(ctx, userID, rootID)
}

// SearchDocuments 搜索文档
func (s *documentAggregateService) SearchDocuments(ctx context.Context, userID int64, keyword string, docType *domain.DocumentType, limit, offset int) ([]*domain.DocumentSearchResult, error) {
	// 1. 获取基础搜索结果
	documents, err := s.documentUsecase.SearchDocuments(ctx, userID, keyword, docType, limit, offset)
	if err != nil {
		return nil, err
	}

	// 2. 转换为搜索结果对象，并附加额外信息
	results := make([]*domain.DocumentSearchResult, 0, len(documents))
	for _, doc := range documents {
		// 检查是否收藏
		isFavorite, _ := s.favoriteUsecase.IsFavorite(ctx, doc.ID, userID)

		// 获取用户权限
		permission := domain.PermissionView
		if doc.OwnerID == userID {
			permission = domain.PermissionFull
		} else if perms, err := s.permUsecase.GetDocumentPermissions(ctx, userID, doc.ID); err == nil && len(perms) > 0 {
			permission = perms[0].Permission
		}

		// 获取所有者信息
		var owner *domain.User
		if ownerUser, err := s.userRepo.GetByID(ctx, doc.OwnerID); err == nil {
			owner = ownerUser
		}

		result := &domain.DocumentSearchResult{
			ID:         doc.ID,
			Title:      doc.Title,
			Type:       doc.Type,
			IsStarred:  doc.IsStarred,
			IsFavorite: isFavorite,
			UpdatedAt:  doc.UpdatedAt,
			Owner:      owner,
			Permission: permission,
			MatchScore: 1.0, // 简单实现，实际应该基于搜索算法计算
		}
		results = append(results, result)
	}

	return results, nil
}

// GetRecentDocuments 获取最近访问的文档
func (s *documentAggregateService) GetRecentDocuments(ctx context.Context, userID int64, limit int) ([]*domain.Document, error) {
	return s.documentUsecase.GetRecentDocuments(ctx, userID, limit)
}

// MoveDocument 移动文档
func (s *documentAggregateService) MoveDocument(ctx context.Context, userID, documentID int64, newParentID *int64) error {
	return s.documentUsecase.MoveDocument(ctx, userID, documentID, newParentID)
}

// DuplicateDocument 复制文档
func (s *documentAggregateService) DuplicateDocument(ctx context.Context, userID, documentID int64, newTitle string) (*domain.Document, error) {
	return s.documentUsecase.DuplicateDocument(ctx, userID, documentID, newTitle)
}

// BatchDeleteDocuments 批量删除文档
func (s *documentAggregateService) BatchDeleteDocuments(ctx context.Context, userID int64, documentIDs []int64) error {
	return s.documentUsecase.BatchDeleteDocuments(ctx, userID, documentIDs)
}

// BatchMoveDocuments 批量移动文档
func (s *documentAggregateService) BatchMoveDocuments(ctx context.Context, userID int64, documentIDs []int64, newParentID *int64) error {
	return s.documentUsecase.BatchMoveDocuments(ctx, userID, documentIDs, newParentID)
}

// === 文档分享操作（委托给DocumentShareUsecase） ===

// ShareDocument 分享文档
func (s *documentAggregateService) ShareDocument(ctx context.Context, userID, documentID int64, permission domain.Permission, password string, expiresAt *time.Time, shareWithUserIDs []int64) (*domain.DocumentShare, error) {
	return s.shareUsecase.CreateShareLink(ctx, userID, documentID, permission, password, expiresAt, shareWithUserIDs)
}

// UpdateShareLink 更新分享链接
func (s *documentAggregateService) UpdateShareLink(ctx context.Context, userID, shareID int64, permission *domain.Permission, password *string, expiresAt *time.Time) (*domain.DocumentShare, error) {
	return s.shareUsecase.UpdateShareLink(ctx, userID, shareID, permission, password, expiresAt)
}

// DeleteShareLink 删除分享链接
func (s *documentAggregateService) DeleteShareLink(ctx context.Context, userID, shareID int64) error {
	return s.shareUsecase.DeleteShareLink(ctx, userID, shareID)
}

// GetDocumentShares 获取文档的分享列表
func (s *documentAggregateService) GetDocumentShares(ctx context.Context, userID, documentID int64) ([]*domain.DocumentShare, error) {
	return s.shareUsecase.GetDocumentShares(ctx, userID, documentID)
}

// GetMySharedDocuments 获取我分享的文档列表
func (s *documentAggregateService) GetMySharedDocuments(ctx context.Context, userID int64) ([]*domain.DocumentShare, error) {
	return s.shareUsecase.GetMySharedDocuments(ctx, userID)
}

// GetSharedDocument 通过分享链接获取文档
func (s *documentAggregateService) GetSharedDocument(ctx context.Context, linkID, password string, accessIP string) (*domain.DocumentAccessInfo, error) {
	//return s.shareUsecase.GetSharedDocument(ctx, linkID, password, accessIP)
	return nil, nil
}

// todo梳理整个文档模块

// ValidateShareAccess 验证分享访问权限
func (s *documentAggregateService) ValidateShareAccess(ctx context.Context, linkID, password string) (*domain.DocumentShare, error) {
	return s.shareUsecase.ValidateShareAccess(ctx, linkID, password)
}

// === 文档权限操作（委托给DocumentPermissionUsecase） ===

// GrantDocumentPermission 授予文档权限
func (s *documentAggregateService) GrantDocumentPermission(ctx context.Context, userID, documentID, targetUserID int64, permission domain.Permission) error {
	return s.permUsecase.GrantPermission(ctx, userID, documentID, targetUserID, permission)
}

// RevokeDocumentPermission 撤销文档权限
func (s *documentAggregateService) RevokeDocumentPermission(ctx context.Context, userID, documentID, targetUserID int64) error {
	return s.permUsecase.RevokePermission(ctx, userID, documentID, targetUserID)
}

// UpdateDocumentPermission 更新文档权限
func (s *documentAggregateService) UpdateDocumentPermission(ctx context.Context, userID, documentID, targetUserID int64, permission domain.Permission) error {
	return s.permUsecase.UpdatePermission(ctx, userID, documentID, targetUserID, permission)
}

// CheckDocumentPermission 检查文档权限
func (s *documentAggregateService) CheckDocumentPermission(ctx context.Context, userID, documentID int64, permission domain.Permission) (bool, error) {
	return s.permUsecase.CheckPermission(ctx, documentID, userID, permission)
}

// GetDocumentPermissions 获取文档权限列表
func (s *documentAggregateService) GetDocumentPermissions(ctx context.Context, userID, documentID int64) ([]*domain.DocumentPermission, error) {
	return s.permUsecase.GetDocumentPermissions(ctx, userID, documentID)
}

// BatchGrantPermission 批量授予权限
func (s *documentAggregateService) BatchGrantPermission(ctx context.Context, userID, documentID int64, targetUserIDs []int64, permission domain.Permission) error {
	return s.permUsecase.BatchGrantPermission(ctx, userID, documentID, targetUserIDs, permission)
}

// BatchRevokePermission 批量撤销权限
func (s *documentAggregateService) BatchRevokePermission(ctx context.Context, userID, documentID int64, targetUserIDs []int64) error {
	return s.permUsecase.BatchRevokePermission(ctx, userID, documentID, targetUserIDs)
}

// === 文档收藏操作（委托给DocumentFavoriteUsecase） ===

// ToggleDocumentFavorite 切换文档收藏状态
func (s *documentAggregateService) ToggleDocumentFavorite(ctx context.Context, userID, documentID int64) (bool, error) {
	return s.favoriteUsecase.ToggleFavorite(ctx, userID, documentID)
}

// SetFavoriteCustomTitle 设置收藏的自定义标题
func (s *documentAggregateService) SetFavoriteCustomTitle(ctx context.Context, userID, documentID int64, customTitle string) error {
	return s.favoriteUsecase.SetCustomTitle(ctx, userID, documentID, customTitle)
}

// RemoveDocumentFavorite 移除文档收藏
func (s *documentAggregateService) RemoveDocumentFavorite(ctx context.Context, userID, documentID int64) error {
	return s.favoriteUsecase.RemoveFavorite(ctx, userID, documentID)
}

// GetFavoriteDocuments 获取收藏的文档列表
func (s *documentAggregateService) GetFavoriteDocuments(ctx context.Context, userID int64) ([]*domain.DocumentFavorite, error) {
	return s.favoriteUsecase.GetMyFavorites(ctx, userID)
}

// IsFavoriteDocument 检查是否收藏了文档
func (s *documentAggregateService) IsFavoriteDocument(ctx context.Context, userID, documentID int64) (bool, error) {
	return s.favoriteUsecase.IsFavorite(ctx, userID, documentID)
}

// === 聚合根级别的复合操作 ===

// GetDocumentWithAccessInfo 获取文档及其访问信息
// 这是聚合根特有的复合操作，整合多个子域的信息
func (s *documentAggregateService) GetDocumentWithAccessInfo(ctx context.Context, userID, documentID int64) (*domain.DocumentAccessInfo, error) {
	// 1. 获取文档基本信息
	document, err := s.documentUsecase.GetDocument(ctx, userID, documentID)
	if err != nil {
		return nil, err
	}

	// 2. 获取用户信息
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 3. 获取权限信息
	permissions, err := s.permUsecase.GetDocumentPermissions(ctx, userID, documentID)
	if err != nil {
		return nil, err
	}

	// 4. 构建访问信息
	return domain.BuildDocumentAccessInfo(user, document, permissions), nil
}

// GetDocumentFullInfo 获取文档完整信息
// 包含文档本身及其相关的分享、权限、收藏状态等完整信息
func (s *documentAggregateService) GetDocumentFullInfo(ctx context.Context, userID, documentID int64) (*domain.DocumentFullInfo, error) {
	// 1. 获取基本访问信息
	accessInfo, err := s.GetDocumentWithAccessInfo(ctx, userID, documentID)
	if err != nil {
		return nil, err
	}

	// 2. 获取文档所有者信息
	owner, err := s.userRepo.GetByID(ctx, accessInfo.Document.OwnerID)
	if err != nil {
		return nil, err
	}

	// 3. 获取分享信息（如果用户有权限查看）
	var shares []*domain.DocumentShare
	if accessInfo.CanManage {
		shares, _ = s.shareUsecase.GetDocumentShares(ctx, userID, documentID)
	}

	// 4. 获取权限列表（如果用户有权限查看）
	var permissions []*domain.DocumentPermission
	if accessInfo.CanManage {
		permissions, _ = s.permUsecase.GetDocumentPermissions(ctx, userID, documentID)
	}

	// 5. 获取收藏状态
	isFavorite, _ := s.favoriteUsecase.IsFavorite(ctx, userID, documentID)
	var favoriteInfo *domain.DocumentFavorite
	if isFavorite {
		favorites, err := s.favoriteUsecase.GetMyFavorites(ctx, userID)
		if err == nil {
			for _, fav := range favorites {
				if fav.DocumentID == documentID {
					favoriteInfo = fav
					break
				}
			}
		}
	}

	// 6. 获取子文档（如果是文件夹）
	var children []*domain.Document
	if accessInfo.Document.IsFolder() {
		children, _ = s.documentUsecase.GetMyDocuments(ctx, userID, &documentID, false)
	}

	// 7. 获取面包屑导航
	allDocs, _ := s.documentUsecase.GetMyDocuments(ctx, userID, nil, false)
	breadcrumb := domain.GetDocumentBreadcrumb(accessInfo.Document, allDocs)

	return &domain.DocumentFullInfo{
		Document:     accessInfo.Document,
		Owner:        owner,
		AccessInfo:   accessInfo,
		Shares:       shares,
		Permissions:  permissions,
		IsFavorite:   isFavorite,
		FavoriteInfo: favoriteInfo,
		Children:     children,
		Breadcrumb:   breadcrumb,
	}, nil
}

// CheckDocumentAccess 检查文档访问权限
// 返回详细的访问信息，包括具体的权限级别
func (s *documentAggregateService) CheckDocumentAccess(ctx context.Context, userID, documentID int64, requiredPermission domain.Permission) (*domain.DocumentAccessInfo, error) {
	// 1. 获取基本访问信息
	accessInfo, err := s.GetDocumentWithAccessInfo(ctx, userID, documentID)
	if err != nil {
		return nil, err
	}

	// 2. 检查权限是否足够
	if !domain.IsPermissionSufficient(accessInfo.Permission, requiredPermission) {
		return nil, domain.ErrPermissionDenied
	}

	return accessInfo, nil
}
