package document

import (
	"context"
	"crypto/subtle"
	"strings"
	"time"

	"DOC/domain"
)

type documentShareService struct {
	shareRepo      domain.DocumentShareRepository
	documentRepo   domain.DocumentRepository
	permissionRepo domain.DocumentPermissionRepository
}

// CreateShareLink 创建文档分享链接
func (d *documentShareService) CreateShareLink(ctx context.Context, userID, documentID int64, permission domain.Permission, password string, expiresAt *time.Time, shareWithUserIDs []int64) (*domain.DocumentShare, error) {
	// 验证输入参数
	if userID <= 0 {
		return nil, domain.ErrInvalidUser
	}
	if documentID <= 0 {
		return nil, domain.ErrInvalidDocument
	}
	if !isValidSharePermission(permission) {
		return nil, domain.ErrInvalidPermission
	}

	// 验证文档存在性和活跃状态
	document, err := d.documentRepo.GetByID(ctx, documentID)
	if err != nil {
		return nil, err
	}
	if document == nil || !document.IsActive() {
		return nil, domain.ErrDocumentNotFound
	}

	// 检查用户是否有权限分享文档（必须是文档所有者或有管理权限）
	if document.OwnerID != userID {
		hasManagePermission, err := d.permissionRepo.CheckPermission(ctx, documentID, userID, domain.PermissionManage)
		if err != nil {
			return nil, err
		}
		if !hasManagePermission {
			return nil, domain.ErrPermissionDenied
		}
	}

	// 验证过期时间
	if expiresAt != nil && expiresAt.Before(time.Now()) {
		return nil, domain.ErrBadParamInput
	}

	// 确定分享类型
	shareType := domain.ShareTypePublic
	if len(shareWithUserIDs) > 0 {
		shareType = domain.ShareTypePrivate
	}

	// 创建分享实体
	share := &domain.DocumentShare{
		DocumentID: documentID,
		ShareType:  shareType,
		Permission: permission,
		Password:   strings.TrimSpace(password),
		ExpiresAt:  expiresAt,
		CreatedBy:  userID,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// 生成分享链接ID
	if err := share.GenerateLinkID(); err != nil {
		return nil, domain.ErrInternalServerError
	}

	// 验证实体
	if err := share.Validate(); err != nil {
		return nil, err
	}

	// 保存分享记录
	if err := d.shareRepo.Store(ctx, share); err != nil {
		return nil, err
	}

	// 如果是私有分享，添加指定用户
	if shareType == domain.ShareTypePrivate {
		// 过滤重复用户ID和无效用户ID
		validUserIDs := make([]int64, 0, len(shareWithUserIDs))
		seen := make(map[int64]bool)
		for _, uid := range shareWithUserIDs {
			if uid > 0 && uid != userID && !seen[uid] { // 排除分享者自己
				validUserIDs = append(validUserIDs, uid)
				seen[uid] = true
			}
		}

		// 添加分享用户
		for _, targetUserID := range validUserIDs {
			shareUser := &domain.DocumentShareUser{
				ShareID: share.ID,
				UserID:  targetUserID,
				AddedAt: time.Now(),
			}

			if err := shareUser.Validate(); err == nil {
				// 忽略重复用户的错误
				_ = d.shareRepo.AddShareUser(ctx, shareUser)
			}
		}
	}

	return share, nil
}

// UpdateShareLink 更新分享链接
func (d *documentShareService) UpdateShareLink(ctx context.Context, userID, shareID int64, permission *domain.Permission, password *string, expiresAt *time.Time) (*domain.DocumentShare, error) {
	// 验证输入参数
	if userID <= 0 {
		return nil, domain.ErrInvalidUser
	}
	if shareID <= 0 {
		return nil, domain.ErrInvalidDocument
	}

	// 获取现有分享
	share, err := d.shareRepo.GetByID(ctx, shareID)
	if err != nil {
		return nil, err
	}

	// 验证操作者权限（必须是分享创建者或文档所有者）
	document, err := d.documentRepo.GetByID(ctx, share.DocumentID)
	if err != nil {
		return nil, err
	}
	if document == nil {
		return nil, domain.ErrDocumentNotFound
	}

	if share.CreatedBy != userID && document.OwnerID != userID {
		return nil, domain.ErrPermissionDenied
	}

	// 更新字段
	updated := false
	if permission != nil && isValidSharePermission(*permission) {
		share.Permission = *permission
		updated = true
	}
	if password != nil {
		share.Password = strings.TrimSpace(*password)
		updated = true
	}
	if expiresAt != nil {
		if expiresAt.Before(time.Now()) {
			return nil, domain.ErrBadParamInput
		}
		share.ExpiresAt = expiresAt
		updated = true
	}

	if !updated {
		return share, nil // 没有需要更新的字段
	}

	// 验证更新后的实体
	if err := share.Validate(); err != nil {
		return nil, err
	}

	// 保存更新
	if err := d.shareRepo.Update(ctx, share); err != nil {
		return nil, err
	}

	return share, nil
}

// DeleteShareLink 删除分享链接
func (d *documentShareService) DeleteShareLink(ctx context.Context, userID, shareID int64) error {
	// 验证输入参数
	if userID <= 0 {
		return domain.ErrInvalidUser
	}
	if shareID <= 0 {
		return domain.ErrInvalidDocument
	}

	// 获取分享信息
	share, err := d.shareRepo.GetByID(ctx, shareID)
	if err != nil {
		return err
	}

	// 验证操作者权限（必须是分享创建者或文档所有者）
	document, err := d.documentRepo.GetByID(ctx, share.DocumentID)
	if err != nil {
		return err
	}
	if document == nil {
		return domain.ErrDocumentNotFound
	}

	if share.CreatedBy != userID && document.OwnerID != userID {
		return domain.ErrPermissionDenied
	}

	// 删除分享
	return d.shareRepo.Delete(ctx, shareID)
}

// GetSharedDocument 通过分享链接获取文档
func (d *documentShareService) GetSharedDocument(ctx context.Context, linkID, password string, accessIP string) (*domain.Document, error) {
	// 验证分享访问
	share, err := d.ValidateShareAccess(ctx, linkID, password)
	if err != nil {
		return nil, err
	}

	// 获取文档
	document, err := d.documentRepo.GetByID(ctx, share.DocumentID)
	if err != nil {
		return nil, err
	}
	if document == nil || !document.IsActive() {
		return nil, domain.ErrDocumentNotFound
	}

	// 记录访问
	if accessIP != "" {
		_ = d.RecordShareAccess(ctx, share.ID, accessIP)
	}

	return document, nil
}

// ValidateShareAccess 验证分享访问权限
func (d *documentShareService) ValidateShareAccess(ctx context.Context, linkID, password string) (*domain.DocumentShare, error) {
	// 验证输入参数
	if strings.TrimSpace(linkID) == "" {
		return nil, domain.ErrShareLinkNotFound
	}

	// 获取分享信息
	share, err := d.shareRepo.GetByLinkID(ctx, linkID)
	if err != nil {
		return nil, err
	}

	// 检查分享是否过期
	if share.IsExpired() {
		return nil, domain.ErrShareLinkExpired
	}

	// 检查密码
	if share.IsPasswordProtected() {
		if password == "" {
			return nil, domain.ErrInvalidSharePassword
		}
		// 使用常量时间比较防止时序攻击
		if subtle.ConstantTimeCompare([]byte(share.Password), []byte(password)) != 1 {
			return nil, domain.ErrInvalidSharePassword
		}
	}

	// 检查文档状态
	if share.Document != nil && !share.Document.IsActive() {
		return nil, domain.ErrDocumentNotFound
	}

	return share, nil
}

// RecordShareAccess 记录分享访问
func (d *documentShareService) RecordShareAccess(ctx context.Context, shareID int64, accessIP string) error {
	// 验证输入参数
	if shareID <= 0 {
		return domain.ErrInvalidDocument
	}

	// 记录访问统计
	return d.shareRepo.IncrementViewCount(ctx, shareID, accessIP)
}

// GetDocumentShares 获取文档的分享列表
func (d *documentShareService) GetDocumentShares(ctx context.Context, userID, documentID int64) ([]*domain.DocumentShare, error) {
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

	// 检查用户是否有权限查看分享列表（必须是文档所有者或有管理权限）
	if document.OwnerID != userID {
		hasManagePermission, err := d.permissionRepo.CheckPermission(ctx, documentID, userID, domain.PermissionManage)
		if err != nil {
			return nil, err
		}
		if !hasManagePermission {
			return nil, domain.ErrPermissionDenied
		}
	}

	// 获取分享列表
	shares, err := d.shareRepo.GetByDocument(ctx, documentID)
	if err != nil {
		return nil, err
	}

	// 过滤过期的分享
	validShares := make([]*domain.DocumentShare, 0, len(shares))
	for _, share := range shares {
		if !share.IsExpired() {
			validShares = append(validShares, share)
		}
	}

	return validShares, nil
}

// GetMySharedDocuments 获取我创建的分享文档
func (d *documentShareService) GetMySharedDocuments(ctx context.Context, userID int64) ([]*domain.DocumentShare, error) {
	// 验证输入参数
	if userID <= 0 {
		return nil, domain.ErrInvalidUser
	}

	// 获取用户创建的分享
	shares, err := d.shareRepo.GetByCreator(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 过滤过期的分享和已删除的文档
	validShares := make([]*domain.DocumentShare, 0, len(shares))
	for _, share := range shares {
		if !share.IsExpired() && share.Document != nil && share.Document.IsActive() {
			validShares = append(validShares, share)
		}
	}

	return validShares, nil
}

// GetSharedWithMeDocuments 获取分享给我的文档
func (d *documentShareService) GetSharedWithMeDocuments(ctx context.Context, userID int64) ([]*domain.DocumentShare, error) {
	// 验证输入参数
	if userID <= 0 {
		return nil, domain.ErrInvalidUser
	}

	// 获取用户可访问的分享文档
	shares, err := d.shareRepo.GetUserSharedDocuments(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 过滤过期的分享、已删除的文档和用户自己创建的分享
	validShares := make([]*domain.DocumentShare, 0, len(shares))
	for _, share := range shares {
		if !share.IsExpired() &&
			share.Document != nil &&
			share.Document.IsActive() &&
			share.CreatedBy != userID { // 排除自己创建的分享
			validShares = append(validShares, share)
		}
	}

	return validShares, nil
}

// AddShareUsers 添加私有分享用户
func (d *documentShareService) AddShareUsers(ctx context.Context, userID, shareID int64, targetUserIDs []int64) error {
	// 验证输入参数
	if userID <= 0 {
		return domain.ErrInvalidUser
	}
	if shareID <= 0 {
		return domain.ErrInvalidDocument
	}
	if len(targetUserIDs) == 0 {
		return domain.ErrInvalidBatchRequest
	}
	if len(targetUserIDs) > 50 { // 限制批量操作大小
		return domain.ErrBatchSizeExceeded
	}

	// 获取分享信息
	share, err := d.shareRepo.GetByID(ctx, shareID)
	if err != nil {
		return err
	}

	// 检查是否为私有分享
	if !share.IsPrivate() {
		return domain.ErrInvalidPermission
	}

	// 验证操作者权限（必须是分享创建者或文档所有者）
	document, err := d.documentRepo.GetByID(ctx, share.DocumentID)
	if err != nil {
		return err
	}
	if document == nil {
		return domain.ErrDocumentNotFound
	}

	if share.CreatedBy != userID && document.OwnerID != userID {
		return domain.ErrPermissionDenied
	}

	// 过滤重复用户ID和无效用户ID
	validUserIDs := make([]int64, 0, len(targetUserIDs))
	seen := make(map[int64]bool)
	for _, uid := range targetUserIDs {
		if uid > 0 && uid != userID && !seen[uid] {
			validUserIDs = append(validUserIDs, uid)
			seen[uid] = true
		}
	}

	// 添加用户到分享
	for _, targetUserID := range validUserIDs {
		shareUser := &domain.DocumentShareUser{
			ShareID: shareID,
			UserID:  targetUserID,
			AddedAt: time.Now(),
		}

		if err := shareUser.Validate(); err == nil {
			// 忽略重复用户的错误
			_ = d.shareRepo.AddShareUser(ctx, shareUser)
		}
	}

	return nil
}

// RemoveShareUsers 移除私有分享用户
func (d *documentShareService) RemoveShareUsers(ctx context.Context, userID, shareID int64, targetUserIDs []int64) error {
	// 验证输入参数
	if userID <= 0 {
		return domain.ErrInvalidUser
	}
	if shareID <= 0 {
		return domain.ErrInvalidDocument
	}
	if len(targetUserIDs) == 0 {
		return domain.ErrInvalidBatchRequest
	}

	// 获取分享信息
	share, err := d.shareRepo.GetByID(ctx, shareID)
	if err != nil {
		return err
	}

	// 检查是否为私有分享
	if !share.IsPrivate() {
		return domain.ErrInvalidPermission
	}

	// 验证操作者权限（必须是分享创建者或文档所有者）
	document, err := d.documentRepo.GetByID(ctx, share.DocumentID)
	if err != nil {
		return err
	}
	if document == nil {
		return domain.ErrDocumentNotFound
	}

	if share.CreatedBy != userID && document.OwnerID != userID {
		return domain.ErrPermissionDenied
	}

	// 移除用户
	for _, targetUserID := range targetUserIDs {
		if targetUserID > 0 {
			_ = d.shareRepo.RemoveShareUser(ctx, shareID, targetUserID)
		}
	}

	return nil
}

// GetShareUsers 获取私有分享的用户列表
func (d *documentShareService) GetShareUsers(ctx context.Context, userID, shareID int64) ([]*domain.DocumentShareUser, error) {
	// 验证输入参数
	if userID <= 0 {
		return nil, domain.ErrInvalidUser
	}
	if shareID <= 0 {
		return nil, domain.ErrInvalidDocument
	}

	// 获取分享信息
	share, err := d.shareRepo.GetByID(ctx, shareID)
	if err != nil {
		return nil, err
	}

	// 检查是否为私有分享
	if !share.IsPrivate() {
		return nil, domain.ErrInvalidPermission
	}

	// 验证操作者权限（必须是分享创建者或文档所有者）
	document, err := d.documentRepo.GetByID(ctx, share.DocumentID)
	if err != nil {
		return nil, err
	}
	if document == nil {
		return nil, domain.ErrDocumentNotFound
	}

	if share.CreatedBy != userID && document.OwnerID != userID {
		return nil, domain.ErrPermissionDenied
	}

	// 获取分享用户列表
	return d.shareRepo.GetShareUsers(ctx, shareID)
}

// GetShareStats 获取分享统计信息
func (d *documentShareService) GetShareStats(ctx context.Context, userID, shareID int64) (*domain.DocumentShare, error) {
	// 验证输入参数
	if userID <= 0 {
		return nil, domain.ErrInvalidUser
	}
	if shareID <= 0 {
		return nil, domain.ErrInvalidDocument
	}

	// 获取分享信息
	share, err := d.shareRepo.GetByID(ctx, shareID)
	if err != nil {
		return nil, err
	}

	// 验证操作者权限（必须是分享创建者或文档所有者）
	document, err := d.documentRepo.GetByID(ctx, share.DocumentID)
	if err != nil {
		return nil, err
	}
	if document == nil {
		return nil, domain.ErrDocumentNotFound
	}

	if share.CreatedBy != userID && document.OwnerID != userID {
		return nil, domain.ErrPermissionDenied
	}

	// 获取统计信息
	return d.shareRepo.GetShareStats(ctx, shareID)
}

// isValidSharePermission 验证分享权限是否有效
func isValidSharePermission(permission domain.Permission) bool {
	validPermissions := []domain.Permission{
		domain.PermissionView,
		domain.PermissionComment,
		domain.PermissionEdit,
		domain.PermissionManage,
		// 注意：通常不允许通过分享授予FULL权限
	}
	for _, p := range validPermissions {
		if permission == p {
			return true
		}
	}
	return false
}

// NewDocumentShareService 创建新的文档分享服务实例
func NewDocumentShareService(
	shareRepo domain.DocumentShareRepository,
	documentRepo domain.DocumentRepository,
	permissionRepo domain.DocumentPermissionRepository) domain.DocumentShareUsecase {
	return &documentShareService{
		shareRepo:      shareRepo,
		documentRepo:   documentRepo,
		permissionRepo: permissionRepo,
	}
}
