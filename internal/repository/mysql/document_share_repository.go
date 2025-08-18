package mysql

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"DOC/domain"
)

// documentShareRepository MySQL文档分享仓储实现
// 实现 domain.DocumentShareRepository 接口，负责文档分享数据的持久化操作
type documentShareRepository struct {
	db *gorm.DB
}

// NewDocumentShareRepository 创建新的文档分享仓储实例
func NewDocumentShareRepository(db *gorm.DB) domain.DocumentShareRepository {
	return &documentShareRepository{db: db}
}

// Store 保存文档分享
func (d *documentShareRepository) Store(ctx context.Context, share *domain.DocumentShare) error {
	if err := d.db.WithContext(ctx).Create(share).Error; err != nil {
		return err
	}
	return nil
}

// GetByID 根据ID获取文档分享
func (d *documentShareRepository) GetByID(ctx context.Context, id int64) (*domain.DocumentShare, error) {
	var share domain.DocumentShare
	if err := d.db.WithContext(ctx).Where("id = ?", id).First(&share).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrShareLinkNotFound
		}
		return nil, err
	}
	return &share, nil
}

// GetByLinkID 根据分享链接ID获取文档分享
func (d *documentShareRepository) GetByLinkID(ctx context.Context, linkID string) (*domain.DocumentShare, error) {
	var share domain.DocumentShare
	if err := d.db.WithContext(ctx).
		Where("link_id = ?", linkID).
		Preload("Document").
		Preload("Creator").
		First(&share).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrShareLinkNotFound
		}
		return nil, err
	}
	return &share, nil
}

// Update 更新文档分享
func (d *documentShareRepository) Update(ctx context.Context, share *domain.DocumentShare) error {
	share.UpdatedAt = time.Now()
	if err := d.db.WithContext(ctx).Save(share).Error; err != nil {
		return err
	}
	return nil
}

// Delete 删除文档分享
func (d *documentShareRepository) Delete(ctx context.Context, id int64) error {
	// 开启事务，同时删除分享和关联的用户
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 先删除关联的分享用户
		if err := tx.Where("share_id = ?", id).Delete(&domain.DocumentShareUser{}).Error; err != nil {
			return err
		}
		// 再删除分享记录
		if err := tx.Delete(&domain.DocumentShare{}, id).Error; err != nil {
			return err
		}
		return nil
	})
}

// GetByDocument 根据文档ID获取所有分享
func (d *documentShareRepository) GetByDocument(ctx context.Context, documentID int64) ([]*domain.DocumentShare, error) {
	var shares []*domain.DocumentShare
	if err := d.db.WithContext(ctx).
		Where("document_id = ?", documentID).
		Preload("Creator").
		Preload("SharedUsers").
		Preload("SharedUsers.User").
		Order("created_at DESC").
		Find(&shares).Error; err != nil {
		return nil, err
	}
	return shares, nil
}

// GetByCreator 根据创建者ID获取分享列表
func (d *documentShareRepository) GetByCreator(ctx context.Context, creatorID int64) ([]*domain.DocumentShare, error) {
	var shares []*domain.DocumentShare
	if err := d.db.WithContext(ctx).
		Where("created_by = ?", creatorID).
		Preload("Document").
		Preload("SharedUsers").
		Preload("SharedUsers.User").
		Order("created_at DESC").
		Find(&shares).Error; err != nil {
		return nil, err
	}
	return shares, nil
}

// GetUserSharedDocuments 获取用户可访问的共享文档（包括公开和私有分享）
func (d *documentShareRepository) GetUserSharedDocuments(ctx context.Context, userID int64) ([]*domain.DocumentShare, error) {
	var shares []*domain.DocumentShare

	// 查询用户可访问的分享：公开分享 + 私有分享中指定用户
	if err := d.db.WithContext(ctx).
		Table("document_shares ds").
		Select("DISTINCT ds.*").
		Joins("LEFT JOIN document_share_users dsu ON ds.id = dsu.share_id").
		Where("ds.share_type = ? OR (ds.share_type = ? AND dsu.user_id = ?)",
			domain.ShareTypePublic, domain.ShareTypePrivate, userID).
		Preload("Document").
		Preload("Creator").
		Order("ds.created_at DESC").
		Find(&shares).Error; err != nil {
		return nil, err
	}
	return shares, nil
}

// AddShareUser 添加分享用户（私有分享）
func (d *documentShareRepository) AddShareUser(ctx context.Context, shareUser *domain.DocumentShareUser) error {
	// 检查用户是否已存在于该分享中
	var count int64
	if err := d.db.WithContext(ctx).
		Model(&domain.DocumentShareUser{}).
		Where("share_id = ? AND user_id = ?", shareUser.ShareID, shareUser.UserID).
		Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		return domain.ErrConflict // 用户已存在
	}

	if err := d.db.WithContext(ctx).Create(shareUser).Error; err != nil {
		return err
	}
	return nil
}

// RemoveShareUser 移除分享用户
func (d *documentShareRepository) RemoveShareUser(ctx context.Context, shareID, userID int64) error {
	if err := d.db.WithContext(ctx).
		Where("share_id = ? AND user_id = ?", shareID, userID).
		Delete(&domain.DocumentShareUser{}).Error; err != nil {
		return err
	}
	return nil
}

// GetShareUsers 获取分享的用户列表
func (d *documentShareRepository) GetShareUsers(ctx context.Context, shareID int64) ([]*domain.DocumentShareUser, error) {
	var shareUsers []*domain.DocumentShareUser
	if err := d.db.WithContext(ctx).
		Where("share_id = ?", shareID).
		Preload("User").
		Order("added_at DESC").
		Find(&shareUsers).Error; err != nil {
		return nil, err
	}
	return shareUsers, nil
}

// IncrementViewCount 增加访问量并记录访问信息
func (d *documentShareRepository) IncrementViewCount(ctx context.Context, shareID int64, accessIP string) error {
	now := time.Now()
	if err := d.db.WithContext(ctx).
		Model(&domain.DocumentShare{}).
		Where("id = ?", shareID).
		Updates(map[string]interface{}{
			"view_count":     gorm.Expr("view_count + 1"),
			"last_access_at": &now,
			"last_access_ip": accessIP,
			"updated_at":     now,
		}).Error; err != nil {
		return err
	}
	return nil
}

// GetShareStats 获取分享统计信息
func (d *documentShareRepository) GetShareStats(ctx context.Context, shareID int64) (*domain.DocumentShare, error) {
	var share domain.DocumentShare
	if err := d.db.WithContext(ctx).
		Select("id, view_count, last_access_at, last_access_ip, created_at").
		Where("id = ?", shareID).
		First(&share).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrShareLinkNotFound
		}
		return nil, err
	}
	return &share, nil
}

// CleanupExpiredShares 清理过期的分享链接
func (d *documentShareRepository) CleanupExpiredShares(ctx context.Context) error {
	// 开启事务，同时清理分享和关联的用户
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 获取过期的分享 ID
		var expiredShareIDs []int64
		if err := tx.Model(&domain.DocumentShare{}).
			Select("id").
			Where("expires_at IS NOT NULL AND expires_at < ?", time.Now()).
			Pluck("id", &expiredShareIDs).Error; err != nil {
			return err
		}

		if len(expiredShareIDs) == 0 {
			return nil // 没有过期的分享
		}

		// 删除过期分享的关联用户
		if err := tx.Where("share_id IN ?", expiredShareIDs).
			Delete(&domain.DocumentShareUser{}).Error; err != nil {
			return err
		}

		// 删除过期的分享记录
		if err := tx.Where("id IN ?", expiredShareIDs).
			Delete(&domain.DocumentShare{}).Error; err != nil {
			return err
		}

		return nil
	})
}
