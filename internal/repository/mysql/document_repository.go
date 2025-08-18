package mysql

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"DOC/domain"
)

// documentRepository MySQL文档仓储实现
// 实现 domain.DocumentRepository 接口，负责文档数据的持久化操作
type documentRepository struct {
	db *gorm.DB
}

// NewDocumentRepository 创建新的文档仓储实例
func NewDocumentRepository(db *gorm.DB) domain.DocumentRepository {
	return &documentRepository{
		db: db,
	}
}

// Store 保存文档
func (d *documentRepository) Store(ctx context.Context, document *domain.Document) error {
	if err := d.db.WithContext(ctx).Create(document).Error; err != nil {
		return err
	}
	return nil
}

// GetByID 根据ID获取文档
func (d *documentRepository) GetByID(ctx context.Context, id int64) (*domain.Document, error) {
	var document domain.Document
	if err := d.db.WithContext(ctx).Where("id = ?", id).First(&document).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrDocumentNotFound
		}
		return nil, err
	}
	return &document, nil
}

// Update 更新文档
func (d *documentRepository) Update(ctx context.Context, document *domain.Document) error {
	document.UpdatedAt = time.Now()
	if err := d.db.WithContext(ctx).Save(document).Error; err != nil {
		return err
	}
	return nil
}

// Delete 硬删除文档
func (d *documentRepository) Delete(ctx context.Context, id int64) error {
	if err := d.db.WithContext(ctx).Delete(&domain.Document{}, id).Error; err != nil {
		return err
	}
	return nil
}

// SoftDelete 软删除文档
func (d *documentRepository) SoftDelete(ctx context.Context, id int64) error {
	// 先获取文档
	doc, err := d.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// 使用领域对象的软删除方法
	// todo， 软删除使用领域对象的删除方法
	doc.SoftDelete()

	// 更新到数据库
	return d.Update(ctx, doc)
}

// GetByOwner 根据所有者ID获取文档列表
func (d *documentRepository) GetByOwner(ctx context.Context, ownerID int64, includeDeleted bool) ([]*domain.Document, error) {
	var documents []*domain.Document
	query := d.db.WithContext(ctx).Where("owner_id = ?", ownerID)

	if !includeDeleted {
		query = query.Where("status != ?", domain.DocumentStatusDeleted)
	}

	if err := query.Order("created_at DESC").Find(&documents).Error; err != nil {
		return nil, err
	}
	return documents, nil
}

// GetByParent 根据父文档ID获取子文档列表
func (d *documentRepository) GetByParent(ctx context.Context, parentID *int64, ownerID int64) ([]*domain.Document, error) {
	var documents []*domain.Document
	query := d.db.WithContext(ctx).Where("owner_id = ? AND status != ?", ownerID, domain.DocumentStatusDeleted)

	if parentID == nil {
		query = query.Where("parent_id IS NULL")
	} else {
		query = query.Where("parent_id = ?", *parentID)
	}

	if err := query.Order("sort_order ASC, created_at DESC").Find(&documents).Error; err != nil {
		return nil, err
	}
	return documents, nil
}

// GetBySpace 根据空间ID获取文档列表
func (d *documentRepository) GetBySpace(ctx context.Context, spaceID int64, ownerID int64) ([]*domain.Document, error) {
	var documents []*domain.Document
	if err := d.db.WithContext(ctx).
		Where("space_id = ? AND owner_id = ? AND status != ?", spaceID, ownerID, domain.DocumentStatusDeleted).
		Order("sort_order ASC, created_at DESC").
		Find(&documents).Error; err != nil {
		return nil, err
	}
	return documents, nil
}

// GetDocumentTree 获取文档树结构
func (d *documentRepository) GetDocumentTree(ctx context.Context, rootID *int64, ownerID int64) ([]*domain.Document, error) {
	var documents []*domain.Document

	// 使用递归CTE查询获取完整的文档树
	sql := `
		WITH RECURSIVE document_tree AS (
			-- 基础查询：获取根节点
			SELECT * FROM documents 
			WHERE owner_id = ? AND status != ? AND parent_id ` +
		map[bool]string{true: "IS NULL", false: "= ?"}[rootID == nil] + `
			
			UNION ALL
			
			-- 递归查询：获取子节点
			SELECT d.* FROM documents d
			INNER JOIN document_tree dt ON d.parent_id = dt.id
			WHERE d.owner_id = ? AND d.status != ?
		)
		SELECT * FROM document_tree
		ORDER BY parent_id ASC, sort_order ASC, created_at DESC
	`

	var args []interface{}
	args = append(args, ownerID, domain.DocumentStatusDeleted)
	if rootID != nil {
		args = append(args, *rootID)
	}
	args = append(args, ownerID, domain.DocumentStatusDeleted)

	if err := d.db.WithContext(ctx).Raw(sql, args...).Scan(&documents).Error; err != nil {
		return nil, err
	}
	return documents, nil
}

// SearchDocuments 搜索文档
func (d *documentRepository) SearchDocuments(ctx context.Context, userID int64, keyword string, docType *domain.DocumentType, limit, offset int) ([]*domain.Document, error) {
	var documents []*domain.Document
	query := d.db.WithContext(ctx).
		Where("owner_id = ? AND status != ?", userID, domain.DocumentStatusDeleted)

	if keyword != "" {
		query = query.Where("title LIKE ? OR content LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	if docType != nil {
		query = query.Where("type = ?", *docType)
	}

	if err := query.
		Order("updated_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&documents).Error; err != nil {
		return nil, err
	}
	return documents, nil
}

// GetStarredDocuments 获取用户星标文档
func (d *documentRepository) GetStarredDocuments(ctx context.Context, userID int64) ([]*domain.Document, error) {
	var documents []*domain.Document
	if err := d.db.WithContext(ctx).
		Where("owner_id = ? AND is_starred = ? AND status != ?", userID, true, domain.DocumentStatusDeleted).
		Order("updated_at DESC").
		Find(&documents).Error; err != nil {
		return nil, err
	}
	return documents, nil
}

// GetRecentDocuments 获取用户最近访问的文档
func (d *documentRepository) GetRecentDocuments(ctx context.Context, userID int64, limit int) ([]*domain.Document, error) {
	var documents []*domain.Document
	if err := d.db.WithContext(ctx).
		Where("owner_id = ? AND status != ?", userID, domain.DocumentStatusDeleted).
		Order("updated_at DESC").
		Limit(limit).
		Find(&documents).Error; err != nil {
		return nil, err
	}
	return documents, nil
}

// UpdateContent 更新文档内容
func (d *documentRepository) UpdateContent(ctx context.Context, id int64, content string) error {
	if err := d.db.WithContext(ctx).
		Model(&domain.Document{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"content":    content,
			"updated_at": time.Now(),
		}).Error; err != nil {
		return err
	}
	return nil
}

// GetContent 获取文档内容
func (d *documentRepository) GetContent(ctx context.Context, id int64) (string, error) {
	var content string
	if err := d.db.WithContext(ctx).
		Model(&domain.Document{}).
		Where("id = ?", id).
		Select("content").
		Scan(&content).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", domain.ErrDocumentNotFound
		}
		return "", err
	}
	return content, nil
}

// UpdateStatus 更新文档状态
func (d *documentRepository) UpdateStatus(ctx context.Context, id int64, status domain.DocumentStatus) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}

	// 如果是删除状态，设置删除时间
	if status == domain.DocumentStatusDeleted {
		updates["deleted_at"] = time.Now()
	} else if status == domain.DocumentStatusActive {
		// 如果是恢复状态，清除删除时间
		updates["deleted_at"] = nil
	}

	if err := d.db.WithContext(ctx).
		Model(&domain.Document{}).
		Where("id = ?", id).
		Updates(updates).Error; err != nil {
		return err
	}
	return nil
}

// ToggleStar 切换文档星标状态
func (d *documentRepository) ToggleStar(ctx context.Context, id int64, userID int64, starred bool) error {
	if err := d.db.WithContext(ctx).
		Model(&domain.Document{}).
		Where("id = ? AND owner_id = ?", id, userID).
		Updates(map[string]interface{}{
			"is_starred": starred,
			"updated_at": time.Now(),
		}).Error; err != nil {
		return err
	}
	return nil
}

// MoveDocument 移动文档到新的父目录
func (d *documentRepository) MoveDocument(ctx context.Context, id int64, newParentID *int64) error {
	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}

	if newParentID == nil {
		updates["parent_id"] = nil
	} else {
		updates["parent_id"] = *newParentID
	}

	if err := d.db.WithContext(ctx).
		Model(&domain.Document{}).
		Where("id = ?", id).
		Updates(updates).Error; err != nil {
		return err
	}
	return nil
}

// BatchDelete 批量软删除文档
func (d *documentRepository) BatchDelete(ctx context.Context, ids []int64, userID int64) error {
	if len(ids) == 0 {
		return nil
	}

	now := time.Now()
	if err := d.db.WithContext(ctx).
		Model(&domain.Document{}).
		Where("id IN ? AND owner_id = ?", ids, userID).
		Updates(map[string]interface{}{
			"status":     domain.DocumentStatusDeleted,
			"deleted_at": &now,
			"updated_at": now,
		}).Error; err != nil {
		return err
	}
	return nil
}

// BatchMove 批量移动文档
func (d *documentRepository) BatchMove(ctx context.Context, ids []int64, newParentID *int64, userID int64) error {
	if len(ids) == 0 {
		return nil
	}

	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}

	if newParentID == nil {
		updates["parent_id"] = nil
	} else {
		updates["parent_id"] = *newParentID
	}

	if err := d.db.WithContext(ctx).
		Model(&domain.Document{}).
		Where("id IN ? AND owner_id = ?", ids, userID).
		Updates(updates).Error; err != nil {
		return err
	}
	return nil
}
