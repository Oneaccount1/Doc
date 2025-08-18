package domain

import (
	"context"
	"time"
)

// DocumentFavorite 文档收藏实体（用户收藏的共享文档）
type DocumentFavorite struct {
	ID          int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	DocumentID  int64     `json:"document_id" gorm:"not null;index"`
	UserID      int64     `json:"user_id" gorm:"not null;index"`
	CustomTitle string    `json:"custom_title" gorm:"type:varchar(255)"` // 用户自定义标题
	CreatedAt   time.Time `json:"created_at"`                            // 收藏时间

	// 关联数据
	Document *Document `json:"document,omitempty" gorm:"foreignKey:DocumentID"`
	User     *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// === 实体方法 ===

// Validate 验证文档收藏
func (df *DocumentFavorite) Validate() error {
	if df.DocumentID <= 0 {
		return ErrInvalidDocument
	}
	if df.UserID <= 0 {
		return ErrInvalidUser
	}
	return nil
}

// SetCustomTitle 设置自定义标题
func (df *DocumentFavorite) SetCustomTitle(title string) {
	df.CustomTitle = title
}

// === 仓储接口 ===

// DocumentFavoriteRepository 文档收藏仓储接口
type DocumentFavoriteRepository interface {
	// 收藏基本操作
	Store(ctx context.Context, favorite *DocumentFavorite) error
	GetByID(ctx context.Context, id int64) (*DocumentFavorite, error)
	Update(ctx context.Context, favorite *DocumentFavorite) error
	Delete(ctx context.Context, id int64) error
	DeleteByDocumentAndUser(ctx context.Context, documentID, userID int64) error

	// 收藏查询
	GetByUser(ctx context.Context, userID int64) ([]*DocumentFavorite, error)
	GetByDocument(ctx context.Context, documentID int64) ([]*DocumentFavorite, error)
	IsFavorite(ctx context.Context, documentID, userID int64) (bool, error)

	// 收藏操作
	ToggleFavorite(ctx context.Context, documentID, userID int64) (bool, error) // 返回是否已收藏
	SetCustomTitle(ctx context.Context, documentID, userID int64, customTitle string) error
}

// === 业务逻辑接口 ===

// DocumentFavoriteUsecase 文档收藏业务逻辑接口
type DocumentFavoriteUsecase interface {
	// 收藏管理
	ToggleFavorite(ctx context.Context, userID, documentID int64) (bool, error)
	SetCustomTitle(ctx context.Context, userID, documentID int64, customTitle string) error
	RemoveFavorite(ctx context.Context, userID, documentID int64) error

	// 收藏查询
	GetMyFavorites(ctx context.Context, userID int64) ([]*DocumentFavorite, error)
	IsFavorite(ctx context.Context, userID, documentID int64) (bool, error)

	// 收藏操作
	HandleFavoriteAction(ctx context.Context, userID, documentID int64, action string) error // action: "favorite", "unfavorite", "remove"
}
