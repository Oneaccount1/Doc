package mysql

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"DOC/domain"
)

// documentFavoriteRepository MySQL文档收藏仓储实现
// 实现 domain.DocumentFavoriteRepository 接口
type documentFavoriteRepository struct {
	db *gorm.DB
}

// NewDocumentFavoriteRepository 创建新的文档收藏仓储实例
func NewDocumentFavoriteRepository(db *gorm.DB) domain.DocumentFavoriteRepository {
	return &documentFavoriteRepository{db: db}
}

// Store 保存文档收藏
func (d *documentFavoriteRepository) Store(ctx context.Context, favorite *domain.DocumentFavorite) error {
	if err := d.db.WithContext(ctx).Create(favorite).Error; err != nil {
		return err
	}
	return nil
}

// GetByID 根据ID获取文档收藏
func (d *documentFavoriteRepository) GetByID(ctx context.Context, id int64) (*domain.DocumentFavorite, error) {
	var favorite domain.DocumentFavorite
	if err := d.db.WithContext(ctx).Where("id = ?", id).First(&favorite).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &favorite, nil
}

// Update 更新文档收藏
func (d *documentFavoriteRepository) Update(ctx context.Context, favorite *domain.DocumentFavorite) error {
	if err := d.db.WithContext(ctx).Save(favorite).Error; err != nil {
		return err
	}
	return nil
}

// Delete 删除文档收藏
func (d *documentFavoriteRepository) Delete(ctx context.Context, id int64) error {
	if err := d.db.WithContext(ctx).Delete(&domain.DocumentFavorite{}, id).Error; err != nil {
		return err
	}
	return nil
}

// DeleteByDocumentAndUser 根据文档ID和用户ID删除收藏
func (d *documentFavoriteRepository) DeleteByDocumentAndUser(ctx context.Context, documentID, userID int64) error {
	if err := d.db.WithContext(ctx).
		Where("document_id = ? AND user_id = ?", documentID, userID).
		Delete(&domain.DocumentFavorite{}).Error; err != nil {
		return err
	}
	return nil
}

// GetByUser 根据用户ID获取收藏列表
func (d *documentFavoriteRepository) GetByUser(ctx context.Context, userID int64) ([]*domain.DocumentFavorite, error) {
	var favorites []*domain.DocumentFavorite
	if err := d.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Preload("Document").
		Order("added_at DESC").
		Find(&favorites).Error; err != nil {
		return nil, err
	}
	return favorites, nil
}

// GetByDocument 根据文档ID获取收藏列表
func (d *documentFavoriteRepository) GetByDocument(ctx context.Context, documentID int64) ([]*domain.DocumentFavorite, error) {
	var favorites []*domain.DocumentFavorite
	if err := d.db.WithContext(ctx).
		Where("document_id = ?", documentID).
		Preload("User").
		Order("added_at DESC").
		Find(&favorites).Error; err != nil {
		return nil, err
	}
	return favorites, nil
}

// IsFavorite 检查是否已收藏
func (d *documentFavoriteRepository) IsFavorite(ctx context.Context, documentID, userID int64) (bool, error) {
	var count int64
	if err := d.db.WithContext(ctx).
		Model(&domain.DocumentFavorite{}).
		Where("document_id = ? AND user_id = ?", documentID, userID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// ToggleFavorite 切换收藏状态
func (d *documentFavoriteRepository) ToggleFavorite(ctx context.Context, documentID, userID int64) (bool, error) {
	// 检查是否已收藏
	isFavorite, err := d.IsFavorite(ctx, documentID, userID)
	if err != nil {
		return false, err
	}

	if isFavorite {
		// 已收藏，则取消收藏
		if err := d.DeleteByDocumentAndUser(ctx, documentID, userID); err != nil {
			return false, err
		}
		return false, nil
	} else {
		// 未收藏，则添加收藏
		favorite := &domain.DocumentFavorite{
			DocumentID: documentID,
			UserID:     userID,
			CreatedAt:  time.Now(),
		}
		if err := d.Store(ctx, favorite); err != nil {
			return false, err
		}
		return true, nil
	}
}

// SetCustomTitle 设置自定义标题
func (d *documentFavoriteRepository) SetCustomTitle(ctx context.Context, documentID, userID int64, customTitle string) error {
	if err := d.db.WithContext(ctx).
		Model(&domain.DocumentFavorite{}).
		Where("document_id = ? AND user_id = ?", documentID, userID).
		Update("custom_title", customTitle).Error; err != nil {
		return err
	}
	return nil
}
