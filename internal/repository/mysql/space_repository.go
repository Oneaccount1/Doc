package mysql

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"DOC/domain"
)

type spaceRepository struct {
	db *gorm.DB
}

func (s *spaceRepository) Store(ctx context.Context, space *domain.Space) error {
	return s.db.WithContext(ctx).Create(space).Error
}

func (s *spaceRepository) GetByID(ctx context.Context, id int64) (*domain.Space, error) {
	var space domain.Space
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&space).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrSpaceNotFound
		}
		return nil, err
	}
	return &space, nil
}

func (s *spaceRepository) GetByName(ctx context.Context, name string, orgID *int64) (*domain.Space, error) {
	var space domain.Space
	query := s.db.WithContext(ctx).Where("name = ?  ", name)

	if orgID != nil {
		query = query.Where("organization_id = ?", *orgID)
	} else {
		query = query.Where("organization_id IS NULL")
	}

	if err := query.First(&space).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrSpaceNotFound
		}
		return nil, err
	}
	return &space, nil
}

func (s *spaceRepository) Update(ctx context.Context, space *domain.Space) error {
	return s.db.WithContext(ctx).Save(space).Error
}

func (s *spaceRepository) Delete(ctx context.Context, id int64) error {
	return s.db.WithContext(ctx).Model(&domain.Space{}).
		Where("id = ?", id).
		Update("status = ?", domain.SpaceStatusDeleted).Error
}

func (s *spaceRepository) GetUserSpaces(ctx context.Context, userID int64) ([]*domain.Space, error) {
	var spaces []*domain.Space

	err := s.db.WithContext(ctx).
		Table("spaces s").
		Select("s.*").
		Joins("LEFT JOIN space_members sm ON s.id = sm.space_id").
		Where("(s.created_by = ? OR sm.user_id = ?)", userID, userID).
		Group("s.id").
		Order("s.created_at DESC").
		Find(&spaces).Error

	if err != nil {
		return nil, err
	}

	return spaces, nil
}

func (s *spaceRepository) GetOrganizationSpaces(ctx context.Context, orgID int64) ([]*domain.Space, error) {
	var spaces []*domain.Space

	err := s.db.WithContext(ctx).
		Where("organization_id = ? ", orgID).
		Order("created_at DESC").
		Find(&spaces).Error

	if err != nil {
		return nil, err
	}

	return spaces, nil
}

func (s *spaceRepository) GetPublicSpaces(ctx context.Context, limit, offset int) ([]*domain.Space, error) {
	var spaces []*domain.Space

	err := s.db.WithContext(ctx).
		Where("is_public = ? ", true).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&spaces).Error

	if err != nil {
		return nil, err
	}

	return spaces, nil
}

func (s *spaceRepository) SearchSpaces(ctx context.Context, keyword string, userID int64, limit, offset int) ([]*domain.Space, error) {
	var spaces []*domain.Space

	query := s.db.WithContext(ctx).
		Table("spaces s").
		Select("s.*").
		Joins("LEFT JOIN space_members sm ON s.id = sm.space_id").
		Where("s.status = ?", domain.SpaceStatusActive)

	if keyword != "" {
		query = query.Where("s.name LIKE ? OR s.description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 只返回用户有权限访问的空间
	query = query.Where("s.is_public = ? OR s.created_by = ? OR sm.user_id = ?", true, userID, userID)

	err := query.Group("s.id").
		Order("s.created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&spaces).Error

	if err != nil {
		return nil, err
	}

	return spaces, nil
}

func (s *spaceRepository) AddMember(ctx context.Context, member *domain.SpaceMember) error {
	return s.db.WithContext(ctx).Create(member).Error
}

func (s *spaceRepository) GetMember(ctx context.Context, spaceID, userID int64) (*domain.SpaceMember, error) {
	var member domain.SpaceMember
	if err := s.db.WithContext(ctx).Where("space_id = ? AND user_id = ?", spaceID, userID).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return &member, nil
}

func (s *spaceRepository) GetMembers(ctx context.Context, spaceID int64) ([]*domain.SpaceMember, error) {
	var members []*domain.SpaceMember

	err := s.db.WithContext(ctx).
		Preload("User").
		Preload("AddedByUser").
		Where("space_id = ?", spaceID).
		Order("added_at ASC").
		Find(&members).Error

	if err != nil {
		return nil, err
	}

	return members, nil
}

func (s *spaceRepository) UpdateMemberRole(ctx context.Context, spaceID, userID int64, role domain.SpaceMemberRole) error {
	return s.db.WithContext(ctx).Model(&domain.SpaceMember{}).
		Where("space_id = ? AND user_id = ?", spaceID, userID).
		Update("role", role).Error
}

func (s *spaceRepository) RemoveMember(ctx context.Context, spaceID, userID int64) error {
	return s.db.WithContext(ctx).
		Where("space_id = ? AND user_id = ?", spaceID, userID).
		Delete(&domain.SpaceMember{}).Error
}

func (s *spaceRepository) AddDocument(ctx context.Context, spaceDocument *domain.SpaceDocument) error {
	return s.db.WithContext(ctx).Create(spaceDocument).Error
}

func (s *spaceRepository) RemoveDocument(ctx context.Context, spaceID, documentID int64) error {
	return s.db.WithContext(ctx).
		Where("space_id = ? AND document_id = ?", spaceID, documentID).
		Delete(&domain.SpaceDocument{}).Error
}

func (s *spaceRepository) GetSpaceDocuments(ctx context.Context, spaceID int64) ([]*domain.Document, error) {
	var documents []*domain.Document

	err := s.db.WithContext(ctx).
		Table("documents d").
		Select("d.*").
		Joins("JOIN space_documents sd ON d.id = sd.document_id").
		Where("sd.space_id = ? ", spaceID).
		Order("sd.added_at DESC").
		Find(&documents).Error

	if err != nil {
		return nil, err
	}

	return documents, nil
}

func (s *spaceRepository) IsDocumentInSpace(ctx context.Context, spaceID, documentID int64) (bool, error) {
	var count int64
	err := s.db.WithContext(ctx).
		Model(&domain.SpaceDocument{}).
		Where("space_id = ? AND document_id = ?", spaceID, documentID).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func NewSpaceRepository(db *gorm.DB) domain.SpaceRepository {
	return &spaceRepository{db: db}
}
