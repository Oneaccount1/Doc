package document

import (
	"context"
	"strings"
	"time"

	"DOC/domain"
)

type documentFavoriteService struct {
	favoriteRepo domain.DocumentFavoriteRepository
	documentRepo domain.DocumentRepository
}

// ToggleFavorite 切换文档收藏状态
func (d *documentFavoriteService) ToggleFavorite(ctx context.Context, userID, documentID int64) (bool, error) {
	// 验证输入参数
	if userID <= 0 {
		return false, domain.ErrInvalidUser
	}
	if documentID <= 0 {
		return false, domain.ErrInvalidDocument
	}

	// 验证文档是否存在且用户有访问权限
	document, err := d.documentRepo.GetByID(ctx, documentID)
	if err != nil {
		return false, err
	}
	if document == nil {
		return false, domain.ErrDocumentNotFound
	}

	// 检查文档是否处于活跃状态
	if !document.IsActive() {
		return false, domain.ErrDocumentNotFound // 对于非活跃文档，表现为不存在
	}

	// 执行切换操作
	isFavorited, err := d.favoriteRepo.ToggleFavorite(ctx, documentID, userID)
	if err != nil {
		return false, err
	}

	return isFavorited, nil
}

// SetCustomTitle 设置收藏文档的自定义标题
func (d *documentFavoriteService) SetCustomTitle(ctx context.Context, userID, documentID int64, customTitle string) error {
	// 验证输入参数
	if userID <= 0 {
		return domain.ErrInvalidUser
	}
	if documentID <= 0 {
		return domain.ErrInvalidDocument
	}

	// 验证自定义标题
	customTitle = strings.TrimSpace(customTitle)
	if len(customTitle) > 255 {
		return domain.ErrInvalidDocumentTitle
	}

	// 检查是否已收藏
	isFavorite, err := d.favoriteRepo.IsFavorite(ctx, documentID, userID)
	if err != nil {
		return err
	}
	if !isFavorite {
		return domain.ErrNotFound
	}

	// 设置自定义标题
	return d.favoriteRepo.SetCustomTitle(ctx, documentID, userID, customTitle)
}

// RemoveFavorite 移除文档收藏
func (d *documentFavoriteService) RemoveFavorite(ctx context.Context, userID, documentID int64) error {
	// 验证输入参数
	if userID <= 0 {
		return domain.ErrInvalidUser
	}
	if documentID <= 0 {
		return domain.ErrInvalidDocument
	}

	// 检查是否已收藏
	isFavorite, err := d.favoriteRepo.IsFavorite(ctx, documentID, userID)
	if err != nil {
		return err
	}
	if !isFavorite {
		return domain.ErrNotFound
	}

	// 删除收藏记录
	return d.favoriteRepo.DeleteByDocumentAndUser(ctx, documentID, userID)
}

// GetMyFavorites 获取用户的收藏文档列表
func (d *documentFavoriteService) GetMyFavorites(ctx context.Context, userID int64) ([]*domain.DocumentFavorite, error) {
	// 验证输入参数
	if userID <= 0 {
		return nil, domain.ErrInvalidUser
	}

	// 获取收藏列表
	favorites, err := d.favoriteRepo.GetByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 过滤掉已删除或不可访问的文档
	validFavorites := make([]*domain.DocumentFavorite, 0, len(favorites))
	for _, favorite := range favorites {
		if favorite.Document != nil && favorite.Document.IsActive() {
			validFavorites = append(validFavorites, favorite)
		}
	}

	return validFavorites, nil
}

// IsFavorite 检查文档是否已被用户收藏
func (d *documentFavoriteService) IsFavorite(ctx context.Context, userID, documentID int64) (bool, error) {
	// 验证输入参数
	if userID <= 0 {
		return false, domain.ErrInvalidUser
	}
	if documentID <= 0 {
		return false, domain.ErrInvalidDocument
	}

	// 检查收藏状态
	return d.favoriteRepo.IsFavorite(ctx, documentID, userID)
}

// HandleFavoriteAction 处理收藏相关操作
func (d *documentFavoriteService) HandleFavoriteAction(ctx context.Context, userID, documentID int64, action string) error {
	// 验证输入参数
	if userID <= 0 {
		return domain.ErrInvalidUser
	}
	if documentID <= 0 {
		return domain.ErrInvalidDocument
	}

	action = strings.ToLower(strings.TrimSpace(action))

	switch action {
	case "favorite":
		// 添加收藏
		isFavorite, err := d.favoriteRepo.IsFavorite(ctx, documentID, userID)
		if err != nil {
			return err
		}
		if isFavorite {
			return nil // 已经收藏，无需重复操作
		}

		// 验证文档存在性
		document, err := d.documentRepo.GetByID(ctx, documentID)
		if err != nil {
			return err
		}
		if document == nil || !document.IsActive() {
			return domain.ErrDocumentNotFound
		}

		// 创建收藏记录
		favorite := &domain.DocumentFavorite{
			DocumentID: documentID,
			UserID:     userID,
			CreatedAt:  time.Now(),
		}

		if err := favorite.Validate(); err != nil {
			return err
		}

		return d.favoriteRepo.Store(ctx, favorite)

	case "unfavorite", "remove":
		// 取消收藏
		return d.RemoveFavorite(ctx, userID, documentID)

	default:
		return domain.ErrInvalidBatchRequest // 使用现有的错误类型表示无效操作
	}
}

// NewDocumentFavoriteService 创建新的文档收藏服务实例
func NewDocumentFavoriteService(
	favoriteRepo domain.DocumentFavoriteRepository,
	documentRepo domain.DocumentRepository) domain.DocumentFavoriteUsecase {
	return &documentFavoriteService{
		favoriteRepo: favoriteRepo,
		documentRepo: documentRepo,
	}
}
