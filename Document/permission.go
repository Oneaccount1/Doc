package document

import (
	"context"
	"time"

	"DOC/domain"
)

type documentPermissionService struct {
	permissionRepo domain.DocumentPermissionRepository
	documentRepo   domain.DocumentRepository
}

// GrantPermission 授予用户文档权限
func (d *documentPermissionService) GrantPermission(ctx context.Context, userID, documentID, targetUserID int64, permission domain.Permission) error {
	// 验证输入参数
	if userID <= 0 || targetUserID <= 0 {
		return domain.ErrInvalidUser
	}
	if documentID <= 0 {
		return domain.ErrInvalidDocument
	}
	if !isValidPermission(permission) {
		return domain.ErrInvalidPermission
	}

	// 验证文档存在性
	document, err := d.documentRepo.GetByID(ctx, documentID)
	if err != nil {
		return err
	}
	if document == nil || !document.IsActive() {
		return domain.ErrDocumentNotFound
	}

	// 检查授权者是否有权限进行授权（必须是文档所有者或有管理权限）
	if document.OwnerID != userID {
		hasManagePermission, err := d.permissionRepo.CheckPermission(ctx, documentID, userID, domain.PermissionManage)
		if err != nil {
			return err
		}
		if !hasManagePermission {
			return domain.ErrPermissionDenied
		}
	}

	// 防止给文档所有者授权（所有者天然拥有所有权限）
	if targetUserID == document.OwnerID {
		return domain.ErrConflict // 文档所有者不需要额外授权
	}

	// 检查是否已存在权限
	existingPermission, err := d.permissionRepo.GetUserPermission(ctx, documentID, targetUserID)
	if err != nil && err != domain.ErrPermissionNotFound {
		return err
	}

	if existingPermission != nil {
		// 更新现有权限
		existingPermission.Permission = permission
		existingPermission.GrantedBy = userID
		existingPermission.UpdatedAt = time.Now()
		return d.permissionRepo.Update(ctx, existingPermission)
	}

	// 创建新权限
	newPermission := &domain.DocumentPermission{
		DocumentID: documentID,
		UserID:     targetUserID,
		Permission: permission,
		GrantedBy:  userID,
		GrantedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := newPermission.Validate(); err != nil {
		return err
	}

	return d.permissionRepo.Store(ctx, newPermission)
}

// RevokePermission 撤销用户文档权限
func (d *documentPermissionService) RevokePermission(ctx context.Context, userID, documentID, targetUserID int64) error {
	// 验证输入参数
	if userID <= 0 || targetUserID <= 0 {
		return domain.ErrInvalidUser
	}
	if documentID <= 0 {
		return domain.ErrInvalidDocument
	}

	// 验证文档存在性
	document, err := d.documentRepo.GetByID(ctx, documentID)
	if err != nil {
		return err
	}
	if document == nil {
		return domain.ErrDocumentNotFound
	}

	// 检查撤销者是否有权限进行撤销（必须是文档所有者或有管理权限）
	if document.OwnerID != userID {
		hasManagePermission, err := d.permissionRepo.CheckPermission(ctx, documentID, userID, domain.PermissionManage)
		if err != nil {
			return err
		}
		if !hasManagePermission {
			return domain.ErrPermissionDenied
		}
	}

	// 防止撤销文档所有者的权限
	if targetUserID == document.OwnerID {
		return domain.ErrPermissionDenied
	}

	// 获取现有权限
	existingPermission, err := d.permissionRepo.GetUserPermission(ctx, documentID, targetUserID)
	if err != nil {
		if err == domain.ErrPermissionNotFound {
			return nil // 权限不存在，无需撤销
		}
		return err
	}

	// 删除权限记录
	return d.permissionRepo.Delete(ctx, existingPermission.ID)
}

// UpdatePermission 更新用户文档权限
func (d *documentPermissionService) UpdatePermission(ctx context.Context, userID, documentID, targetUserID int64, permission domain.Permission) error {
	// 验证输入参数
	if userID <= 0 || targetUserID <= 0 {
		return domain.ErrInvalidUser
	}
	if documentID <= 0 {
		return domain.ErrInvalidDocument
	}
	if !isValidPermission(permission) {
		return domain.ErrInvalidPermission
	}

	// 验证文档存在性
	document, err := d.documentRepo.GetByID(ctx, documentID)
	if err != nil {
		return err
	}
	if document == nil || !document.IsActive() {
		return domain.ErrDocumentNotFound
	}

	// 检查操作者是否有权限进行更新
	if document.OwnerID != userID {
		hasManagePermission, err := d.permissionRepo.CheckPermission(ctx, documentID, userID, domain.PermissionManage)
		if err != nil {
			return err
		}
		if !hasManagePermission {
			return domain.ErrPermissionDenied
		}
	}

	// 防止更新文档所有者的权限
	if targetUserID == document.OwnerID {
		return domain.ErrPermissionDenied
	}

	// 获取现有权限
	existingPermission, err := d.permissionRepo.GetUserPermission(ctx, documentID, targetUserID)
	if err != nil {
		return err
	}

	// 更新权限
	existingPermission.Permission = permission
	existingPermission.GrantedBy = userID
	existingPermission.UpdatedAt = time.Now()

	if err := existingPermission.Validate(); err != nil {
		return err
	}

	return d.permissionRepo.Update(ctx, existingPermission)
}

// GetDocumentPermissions 获取文档的所有权限列表
func (d *documentPermissionService) GetDocumentPermissions(ctx context.Context, userID, documentID int64) ([]*domain.DocumentPermission, error) {
	// 验证输入参数
	if userID <= 0 {
		return nil, domain.ErrInvalidUser
	}
	if documentID <= 0 {
		return nil, domain.ErrInvalidDocument
	}

	// 验证文档存在性
	document, err := d.documentRepo.GetByID(ctx, documentID)
	if err != nil {
		return nil, err
	}
	if document == nil {
		return nil, domain.ErrDocumentNotFound
	}

	// 检查查询者是否有权限查看权限列表（必须是文档所有者或有管理权限）
	if document.OwnerID != userID {
		hasManagePermission, err := d.permissionRepo.CheckPermission(ctx, documentID, userID, domain.PermissionManage)
		if err != nil {
			return nil, err
		}
		if !hasManagePermission {
			return nil, domain.ErrPermissionDenied
		}
	}

	// 获取权限列表
	return d.permissionRepo.GetByDocument(ctx, documentID)
}

// GetUserPermission 获取用户对特定文档的权限
func (d *documentPermissionService) GetUserPermission(ctx context.Context, documentID, userID int64) (*domain.DocumentPermission, error) {
	// 验证输入参数
	if userID <= 0 {
		return nil, domain.ErrInvalidUser
	}
	if documentID <= 0 {
		return nil, domain.ErrInvalidDocument
	}

	// 验证文档存在性
	document, err := d.documentRepo.GetByID(ctx, documentID)
	if err != nil {
		return nil, err
	}
	if document == nil {
		return nil, domain.ErrDocumentNotFound
	}

	// 如果是文档所有者，返回虚拟的完全权限
	if document.OwnerID == userID {
		return &domain.DocumentPermission{
			DocumentID: documentID,
			UserID:     userID,
			Permission: domain.PermissionFull,
			GrantedBy:  userID,
			GrantedAt:  document.CreatedAt,
			UpdatedAt:  document.UpdatedAt,
		}, nil
	}

	// 获取用户权限
	return d.permissionRepo.GetUserPermission(ctx, documentID, userID)
}

// GetUserDocumentsWithPermission 获取用户有特定权限的文档列表
func (d *documentPermissionService) GetUserDocumentsWithPermission(ctx context.Context, userID int64, permission domain.Permission) ([]*domain.Document, error) {
	// 验证输入参数
	if userID <= 0 {
		return nil, domain.ErrInvalidUser
	}
	if !isValidPermission(permission) {
		return nil, domain.ErrInvalidPermission
	}

	// 委托给仓储层处理
	return d.permissionRepo.GetUserDocumentsWithPermission(ctx, userID, permission)
}

// CheckPermission 检查用户是否具有特定权限
func (d *documentPermissionService) CheckPermission(ctx context.Context, documentID, userID int64, permission domain.Permission) (bool, error) {
	// 验证输入参数
	if userID <= 0 {
		return false, domain.ErrInvalidUser
	}
	if documentID <= 0 {
		return false, domain.ErrInvalidDocument
	}
	if !isValidPermission(permission) {
		return false, domain.ErrInvalidPermission
	}

	// 委托给仓储层处理
	return d.permissionRepo.CheckPermission(ctx, documentID, userID, permission)
}

// CanAccessDocument 检查用户是否可以访问文档，并返回最高权限级别
func (d *documentPermissionService) CanAccessDocument(ctx context.Context, documentID, userID int64) (bool, domain.Permission, error) {
	// 验证输入参数
	if userID <= 0 {
		return false, "", domain.ErrInvalidUser
	}
	if documentID <= 0 {
		return false, "", domain.ErrInvalidDocument
	}

	// 验证文档存在性
	document, err := d.documentRepo.GetByID(ctx, documentID)
	if err != nil {
		return false, "", err
	}
	if document == nil || !document.IsActive() {
		return false, "", domain.ErrDocumentNotFound
	}

	// 如果是文档所有者，拥有完全权限
	if document.OwnerID == userID {
		return true, domain.PermissionFull, nil
	}

	// 检查用户的权限
	userPermission, err := d.permissionRepo.GetUserPermission(ctx, documentID, userID)
	if err != nil {
		if err == domain.ErrPermissionNotFound {
			return false, "", nil // 无权限访问
		}
		return false, "", err
	}

	return true, userPermission.Permission, nil
}

// BatchGrantPermission 批量授权
func (d *documentPermissionService) BatchGrantPermission(ctx context.Context, userID, documentID int64, targetUserIDs []int64, permission domain.Permission) error {
	// 验证输入参数
	if userID <= 0 {
		return domain.ErrInvalidUser
	}
	if documentID <= 0 {
		return domain.ErrInvalidDocument
	}
	if len(targetUserIDs) == 0 {
		return domain.ErrInvalidBatchRequest
	}
	if len(targetUserIDs) > 100 { // 限制批量操作大小
		return domain.ErrBatchSizeExceeded
	}
	if !isValidPermission(permission) {
		return domain.ErrInvalidPermission
	}

	// 验证文档存在性
	document, err := d.documentRepo.GetByID(ctx, documentID)
	if err != nil {
		return err
	}
	if document == nil || !document.IsActive() {
		return domain.ErrDocumentNotFound
	}

	// 检查授权者是否有权限进行批量授权
	if document.OwnerID != userID {
		hasManagePermission, err := d.permissionRepo.CheckPermission(ctx, documentID, userID, domain.PermissionManage)
		if err != nil {
			return err
		}
		if !hasManagePermission {
			return domain.ErrPermissionDenied
		}
	}

	// 过滤掉文档所有者和重复的用户ID
	validUserIDs := make([]int64, 0, len(targetUserIDs))
	seen := make(map[int64]bool)
	for _, uid := range targetUserIDs {
		if uid > 0 && uid != document.OwnerID && !seen[uid] {
			validUserIDs = append(validUserIDs, uid)
			seen[uid] = true
		}
	}

	if len(validUserIDs) == 0 {
		return nil // 没有有效的用户ID
	}

	// 执行批量授权
	return d.permissionRepo.BatchGrantPermission(ctx, documentID, validUserIDs, permission, userID)
}

// BatchRevokePermission 批量撤销权限
func (d *documentPermissionService) BatchRevokePermission(ctx context.Context, userID, documentID int64, targetUserIDs []int64) error {
	// 验证输入参数
	if userID <= 0 {
		return domain.ErrInvalidUser
	}
	if documentID <= 0 {
		return domain.ErrInvalidDocument
	}
	if len(targetUserIDs) == 0 {
		return domain.ErrInvalidBatchRequest
	}
	if len(targetUserIDs) > 100 { // 限制批量操作大小
		return domain.ErrBatchSizeExceeded
	}

	// 验证文档存在性
	document, err := d.documentRepo.GetByID(ctx, documentID)
	if err != nil {
		return err
	}
	if document == nil {
		return domain.ErrDocumentNotFound
	}

	// 检查撤销者是否有权限进行批量撤销
	if document.OwnerID != userID {
		hasManagePermission, err := d.permissionRepo.CheckPermission(ctx, documentID, userID, domain.PermissionManage)
		if err != nil {
			return err
		}
		if !hasManagePermission {
			return domain.ErrPermissionDenied
		}
	}

	// 过滤掉文档所有者和无效的用户ID
	validUserIDs := make([]int64, 0, len(targetUserIDs))
	seen := make(map[int64]bool)
	for _, uid := range targetUserIDs {
		if uid > 0 && uid != document.OwnerID && !seen[uid] {
			validUserIDs = append(validUserIDs, uid)
			seen[uid] = true
		}
	}

	if len(validUserIDs) == 0 {
		return nil // 没有有效的用户ID
	}

	// 执行批量撤销
	return d.permissionRepo.BatchRevokePermission(ctx, documentID, validUserIDs)
}

// isValidPermission 验证权限是否有效
func isValidPermission(permission domain.Permission) bool {
	validPermissions := []domain.Permission{
		domain.PermissionView,
		domain.PermissionComment,
		domain.PermissionEdit,
		domain.PermissionManage,
		domain.PermissionFull,
	}
	for _, p := range validPermissions {
		if permission == p {
			return true
		}
	}
	return false
}

// NewDocumentPermissionService 创建新的文档权限服务实例
func NewDocumentPermissionService(
	permissionRepo domain.DocumentPermissionRepository,
	documentRepo domain.DocumentRepository) domain.DocumentPermissionUsecase {
	return &documentPermissionService{
		permissionRepo: permissionRepo,
		documentRepo:   documentRepo,
	}
}
