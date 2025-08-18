package mysql

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"DOC/domain"
)

// documentPermissionRepository MySQL文档权限仓储实现
// 实现 domain.DocumentPermissionRepository 接口
type documentPermissionRepository struct {
	db *gorm.DB
}

// NewDocumentPermissionRepository 创建新的文档权限仓储实例
func NewDocumentPermissionRepository(db *gorm.DB) domain.DocumentPermissionRepository {
	return &documentPermissionRepository{
		db: db,
	}
}

// Store 保存文档权限
func (d *documentPermissionRepository) Store(ctx context.Context, permission *domain.DocumentPermission) error {
	if err := d.db.WithContext(ctx).Create(permission).Error; err != nil {
		return err
	}
	return nil
}

// GetByID 根据ID获取文档权限
func (d *documentPermissionRepository) GetByID(ctx context.Context, id int64) (*domain.DocumentPermission, error) {
	var permission domain.DocumentPermission
	if err := d.db.WithContext(ctx).Where("id = ?", id).First(&permission).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrPermissionNotFound
		}
		return nil, err
	}
	return &permission, nil
}

// Update 更新文档权限
func (d *documentPermissionRepository) Update(ctx context.Context, permission *domain.DocumentPermission) error {
	permission.UpdatedAt = time.Now()
	if err := d.db.WithContext(ctx).Save(permission).Error; err != nil {
		return err
	}
	return nil
}

// Delete 删除文档权限
func (d *documentPermissionRepository) Delete(ctx context.Context, id int64) error {
	if err := d.db.WithContext(ctx).Delete(&domain.DocumentPermission{}, id).Error; err != nil {
		return err
	}
	return nil
}

// GetByDocument 根据文档ID获取所有权限
func (d *documentPermissionRepository) GetByDocument(ctx context.Context, documentID int64) ([]*domain.DocumentPermission, error) {
	var permissions []*domain.DocumentPermission
	if err := d.db.WithContext(ctx).
		Where("document_id = ?", documentID).
		Preload("User").
		Preload("GrantedByUser").
		Order("granted_at DESC").
		Find(&permissions).Error; err != nil {
		return nil, err
	}
	return permissions, nil
}

// GetByUser 根据用户ID获取所有权限
func (d *documentPermissionRepository) GetByUser(ctx context.Context, userID int64) ([]*domain.DocumentPermission, error) {
	var permissions []*domain.DocumentPermission
	if err := d.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Preload("Document").
		Preload("GrantedByUser").
		Order("granted_at DESC").
		Find(&permissions).Error; err != nil {
		return nil, err
	}
	return permissions, nil
}

// GetUserPermission 获取用户对特定文档的权限
func (d *documentPermissionRepository) GetUserPermission(ctx context.Context, documentID, userID int64) (*domain.DocumentPermission, error) {
	var permission domain.DocumentPermission
	if err := d.db.WithContext(ctx).
		Where("document_id = ? AND user_id = ?", documentID, userID).
		First(&permission).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrPermissionNotFound
		}
		return nil, err
	}
	return &permission, nil
}

// CheckPermission 检查用户是否具有特定权限
func (d *documentPermissionRepository) CheckPermission(ctx context.Context, documentID, userID int64, permission domain.Permission) (bool, error) {
	// 首先检查用户是否是文档所有者
	var doc domain.Document
	if err := d.db.WithContext(ctx).Where("id = ?", documentID).First(&doc).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, domain.ErrDocumentNotFound
		}
		return false, err
	}

	// 如果是文档所有者，拥有所有权限
	if doc.OwnerID == userID {
		return true, nil
	}

	// 检查用户的权限授权
	var userPermission domain.DocumentPermission
	if err := d.db.WithContext(ctx).
		Where("document_id = ? AND user_id = ?", documentID, userID).
		First(&userPermission).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil // 无权限
		}
		return false, err
	}

	// 根据权限类型检查
	switch permission {
	case domain.PermissionView:
		return userPermission.CanView(), nil
	case domain.PermissionComment:
		return userPermission.CanComment(), nil
	case domain.PermissionEdit:
		return userPermission.CanEdit(), nil
	case domain.PermissionManage:
		return userPermission.CanManage(), nil
	case domain.PermissionFull:
		return userPermission.CanFullControl(), nil
	default:
		return false, domain.ErrInvalidPermission
	}
}

// GetUserDocumentsWithPermission 获取用户有特定权限的文档列表
func (d *documentPermissionRepository) GetUserDocumentsWithPermission(ctx context.Context, userID int64, permission domain.Permission) ([]*domain.Document, error) {
	var documents []*domain.Document

	// 构建权限条件：根据请求的权限确定需要的最低权限级别
	var permissionConditions []domain.Permission
	switch permission {
	case domain.PermissionView:
		// VIEW权限：所有权限级别都包含VIEW
		permissionConditions = []domain.Permission{
			domain.PermissionView, domain.PermissionComment,
			domain.PermissionEdit, domain.PermissionManage, domain.PermissionFull,
		}
	case domain.PermissionComment:
		permissionConditions = []domain.Permission{
			domain.PermissionComment, domain.PermissionEdit,
			domain.PermissionManage, domain.PermissionFull,
		}
	case domain.PermissionEdit:
		permissionConditions = []domain.Permission{
			domain.PermissionEdit, domain.PermissionManage, domain.PermissionFull,
		}
	case domain.PermissionManage:
		permissionConditions = []domain.Permission{
			domain.PermissionManage, domain.PermissionFull,
		}
	case domain.PermissionFull:
		permissionConditions = []domain.Permission{domain.PermissionFull}
	default:
		return nil, domain.ErrInvalidPermission
	}

	// 查询用户作为所有者的文档 + 有权限的文档
	if err := d.db.WithContext(ctx).
		Table("documents d").
		Select("DISTINCT d.*").
		Joins("LEFT JOIN document_permissions dp ON d.id = dp.document_id").
		Where("(d.owner_id = ? OR (dp.user_id = ? AND dp.permission IN ?)) AND d.status != ?",
			userID, userID, permissionConditions, domain.DocumentStatusDeleted).
		Order("d.updated_at DESC").
		Find(&documents).Error; err != nil {
		return nil, err
	}

	return documents, nil
}

// BatchGrantPermission 批量授权
func (d *documentPermissionRepository) BatchGrantPermission(ctx context.Context, documentID int64, userIDs []int64, permission domain.Permission, grantedBy int64) error {
	if len(userIDs) == 0 {
		return nil
	}

	// 开启事务
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, userID := range userIDs {
			// 检查是否已存在权限
			var existing domain.DocumentPermission
			err := tx.Where("document_id = ? AND user_id = ?", documentID, userID).First(&existing).Error

			if err == nil {
				// 更新现有权限
				existing.Permission = permission
				existing.GrantedBy = grantedBy
				existing.UpdatedAt = time.Now()
				if err := tx.Save(&existing).Error; err != nil {
					return err
				}
			} else if errors.Is(err, gorm.ErrRecordNotFound) {
				// 创建新权限
				newPermission := &domain.DocumentPermission{
					DocumentID: documentID,
					UserID:     userID,
					Permission: permission,
					GrantedBy:  grantedBy,
					GrantedAt:  time.Now(),
					UpdatedAt:  time.Now(),
				}
				if err := tx.Create(newPermission).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}
		return nil
	})
}

// BatchRevokePermission 批量撤销权限
func (d *documentPermissionRepository) BatchRevokePermission(ctx context.Context, documentID int64, userIDs []int64) error {
	if len(userIDs) == 0 {
		return nil
	}

	if err := d.db.WithContext(ctx).
		Where("document_id = ? AND user_id IN ?", documentID, userIDs).
		Delete(&domain.DocumentPermission{}).Error; err != nil {
		return err
	}
	return nil
}
