package mysql

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"DOC/domain"
)

// organizationRepository MySQL组织仓储实现
type organizationRepository struct {
	db *gorm.DB
}

// NewOrganizationRepository 创建新的组织仓储实例
func NewOrganizationRepository(db *gorm.DB) domain.OrganizationRepository {
	return &organizationRepository{
		db: db,
	}
}

// Store 保存组织
func (r *organizationRepository) Store(ctx context.Context, organization *domain.Organization) error {
	if err := r.db.WithContext(ctx).Create(organization).Error; err != nil {
		return err
	}
	return nil
}

// GetByID 根据ID获取组织
func (r *organizationRepository) GetByID(ctx context.Context, id int64) (*domain.Organization, error) {
	var organization domain.Organization
	if err := r.db.WithContext(ctx).Where("id = ? AND status != ?", id, domain.OrganizationStatusDeleted).First(&organization).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrOrganizationNotFound
		}
		return nil, err
	}
	return &organization, nil
}

// GetByName 根据名称获取组织
func (r *organizationRepository) GetByName(ctx context.Context, name string) (*domain.Organization, error) {
	var organization domain.Organization
	if err := r.db.WithContext(ctx).Where("name = ? AND status != ?", name, domain.OrganizationStatusDeleted).First(&organization).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrOrganizationNotFound
		}
		return nil, err
	}
	return &organization, nil
}

// Update 更新组织
func (r *organizationRepository) Update(ctx context.Context, organization *domain.Organization) error {
	return r.db.WithContext(ctx).Save(organization).Error
}

// Delete 删除组织（软删除）
func (r *organizationRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Model(&domain.Organization{}).
		Where("id = ?", id).
		Update("status", domain.OrganizationStatusDeleted).Error
}

// GetUserOrganizations 获取用户的组织列表
func (r *organizationRepository) GetUserOrganizations(ctx context.Context, userID int64) ([]*domain.Organization, error) {
	var organizations []*domain.Organization

	err := r.db.WithContext(ctx).
		Table("organizations o").
		Select("o.*").
		Joins("JOIN organization_members om ON o.id = om.organization_id").
		Where("om.user_id = ? AND o.status != ?", userID, domain.OrganizationStatusDeleted).
		Order("o.created_at DESC").
		Find(&organizations).Error

	if err != nil {
		return nil, err
	}

	return organizations, nil
}

// GetPublicOrganizations 获取公开组织列表
func (r *organizationRepository) GetPublicOrganizations(ctx context.Context, limit, offset int) ([]*domain.Organization, error) {
	var organizations []*domain.Organization

	err := r.db.WithContext(ctx).
		Where("is_public = ? AND status = ?", true, domain.OrganizationStatusActive).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&organizations).Error

	if err != nil {
		return nil, err
	}

	return organizations, nil
}

// SearchOrganizations 搜索组织
func (r *organizationRepository) SearchOrganizations(ctx context.Context, keyword string, isPublic *bool, limit, offset int) ([]*domain.Organization, error) {
	var organizations []*domain.Organization

	query := r.db.WithContext(ctx).Where("status = ?", domain.OrganizationStatusActive)

	if keyword != "" {
		query = query.Where("name LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	if isPublic != nil {
		query = query.Where("is_public = ?", *isPublic)
	}

	err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&organizations).Error

	if err != nil {
		return nil, err
	}

	return organizations, nil
}

// AddMember 添加组织成员
func (r *organizationRepository) AddMember(ctx context.Context, member *domain.OrganizationMember) error {
	return r.db.WithContext(ctx).Create(member).Error
}

// GetMember 获取组织成员
func (r *organizationRepository) GetMember(ctx context.Context, orgID, userID int64) (*domain.OrganizationMember, error) {
	var member domain.OrganizationMember
	if err := r.db.WithContext(ctx).Where("organization_id = ? AND user_id = ?", orgID, userID).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return &member, nil
}

// GetMembers 获取组织成员列表
func (r *organizationRepository) GetMembers(ctx context.Context, orgID int64) ([]*domain.OrganizationMember, error) {
	var members []*domain.OrganizationMember

	err := r.db.WithContext(ctx).
		Preload("User").
		Where("organization_id = ?", orgID).
		Order("joined_at ASC").
		Find(&members).Error

	if err != nil {
		return nil, err
	}

	return members, nil
}

// UpdateMemberRole 更新成员角色
func (r *organizationRepository) UpdateMemberRole(ctx context.Context, orgID, userID int64, role domain.OrganizationMemberRole) error {
	return r.db.WithContext(ctx).Model(&domain.OrganizationMember{}).
		Where("organization_id = ? AND user_id = ?", orgID, userID).
		Update("role", role).Error
}

// RemoveMember 移除组织成员
func (r *organizationRepository) RemoveMember(ctx context.Context, orgID, userID int64) error {
	return r.db.WithContext(ctx).
		Where("organization_id = ? AND user_id = ?", orgID, userID).
		Delete(&domain.OrganizationMember{}).Error
}

// StoreInvitation 保存邀请
func (r *organizationRepository) StoreInvitation(ctx context.Context, invitation *domain.OrganizationInvitation) error {
	return r.db.WithContext(ctx).Create(invitation).Error
}

// GetInvitationByToken 根据令牌获取邀请
func (r *organizationRepository) GetInvitationByToken(ctx context.Context, token string) (*domain.OrganizationInvitation, error) {
	var invitation domain.OrganizationInvitation
	if err := r.db.WithContext(ctx).Where("token = ?", token).First(&invitation).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &invitation, nil
}

// GetInvitationsByOrg 获取组织的邀请列表
func (r *organizationRepository) GetInvitationsByOrg(ctx context.Context, orgID int64) ([]*domain.OrganizationInvitation, error) {
	var invitations []*domain.OrganizationInvitation

	err := r.db.WithContext(ctx).
		Where("organization_id = ?", orgID).
		Order("invited_at DESC").
		Find(&invitations).Error

	if err != nil {
		return nil, err
	}

	return invitations, nil
}

// UpdateInvitation 更新邀请
func (r *organizationRepository) UpdateInvitation(ctx context.Context, invitation *domain.OrganizationInvitation) error {
	return r.db.WithContext(ctx).Save(invitation).Error
}

// DeleteInvitation 删除邀请
func (r *organizationRepository) DeleteInvitation(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&domain.OrganizationInvitation{}, id).Error
}

// StoreJoinRequest 保存加入申请
func (r *organizationRepository) StoreJoinRequest(ctx context.Context, request *domain.OrganizationJoinRequest) error {
	return r.db.WithContext(ctx).Create(request).Error
}

// GetJoinRequest 获取加入申请
func (r *organizationRepository) GetJoinRequest(ctx context.Context, id int64) (*domain.OrganizationJoinRequest, error) {
	var request domain.OrganizationJoinRequest
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&request).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &request, nil
}

// GetJoinRequestsByOrg 获取组织的加入申请列表
func (r *organizationRepository) GetJoinRequestsByOrg(ctx context.Context, orgID int64) ([]*domain.OrganizationJoinRequest, error) {
	var requests []*domain.OrganizationJoinRequest

	err := r.db.WithContext(ctx).
		Preload("User").
		Where("organization_id = ?", orgID).
		Order("created_at DESC").
		Find(&requests).Error

	if err != nil {
		return nil, err
	}

	return requests, nil
}

// GetJoinRequestsByUser 获取用户的加入申请列表
func (r *organizationRepository) GetJoinRequestsByUser(ctx context.Context, userID int64) ([]*domain.OrganizationJoinRequest, error) {
	var requests []*domain.OrganizationJoinRequest

	err := r.db.WithContext(ctx).
		Preload("Organization").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&requests).Error

	if err != nil {
		return nil, err
	}

	return requests, nil
}

// UpdateJoinRequest 更新加入申请
func (r *organizationRepository) UpdateJoinRequest(ctx context.Context, request *domain.OrganizationJoinRequest) error {
	return r.db.WithContext(ctx).Save(request).Error
}

// CleanupExpiredInvitations 清理过期邀请
func (r *organizationRepository) CleanupExpiredInvitations(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expires_at < NOW() AND is_used = false").
		Delete(&domain.OrganizationInvitation{}).Error
}

// CleanupExpiredJoinRequests 清理过期加入申请
func (r *organizationRepository) CleanupExpiredJoinRequests(ctx context.Context) error {
	// 将30天前的待处理申请标记为过期
	return r.db.WithContext(ctx).Model(&domain.OrganizationJoinRequest{}).
		Where("created_at < DATE_SUB(NOW(), INTERVAL 30 DAY) AND status = ?", domain.JoinRequestStatusPending).
		Update("status", domain.JoinRequestStatusExpired).Error
}
