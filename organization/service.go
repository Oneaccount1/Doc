package organization

import (
	"DOC/pkg/common"
	"context"
	"errors"
	"fmt"
	"time"

	"DOC/domain"
	"DOC/pkg/utils"
)

// organizationService 组织业务服务实现
type organizationService struct {
	organizationRepo domain.OrganizationRepository
	userRepo         domain.UserRepository
	emailService     domain.EmailUsecase
	contextTimeout   time.Duration
}

// NewOrganizationService 创建新的组织服务实例
func NewOrganizationService(
	organizationRepo domain.OrganizationRepository,
	userRepo domain.UserRepository,
	emailService domain.EmailUsecase,
	timeout time.Duration,
) domain.OrganizationUsecase {
	return &organizationService{
		organizationRepo: organizationRepo,
		userRepo:         userRepo,
		emailService:     emailService,
		contextTimeout:   timeout,
	}
}

// CreateOrganization 创建组织
func (s *organizationService) CreateOrganization(ctx context.Context, userID int64, name, description, logo, website string, isPublic bool) (*domain.Organization, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	// 1. 验证用户是否存在
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}

	// 2. 检查组织名称是否已存在
	if existOrg, err := s.organizationRepo.GetByName(ctx, name); err != nil {
		if !errors.Is(err, domain.ErrOrganizationNotFound) {
			return nil, fmt.Errorf("检查组织名称失败: %w", err)
		}
	} else if existOrg != nil {
		return nil, domain.ErrOrganizationAlreadyExist
	}

	// 3. 创建组织实体
	organization := &domain.Organization{
		Name:        name,
		Description: description,
		Logo:        logo,
		Website:     website,
		IsPublic:    isPublic,
		Status:      domain.OrganizationStatusActive,
		CreatedBy:   userID,
		MemberCount: 1, // 创建者自动成为成员
	}

	// 4. 验证组织实体
	if err := organization.Validate(); err != nil {
		return nil, err
	}

	// 5. 保存组织
	if err := s.organizationRepo.Store(ctx, organization); err != nil {
		return nil, fmt.Errorf("保存组织失败: %w", err)
	}

	// 6. 添加创建者为所有者
	member := &domain.OrganizationMember{
		OrganizationID: organization.ID,
		UserID:         userID,
		Role:           domain.OrgRoleOwner,
		InvitedBy:      userID,
	}

	if err := s.organizationRepo.AddMember(ctx, member); err != nil {
		return nil, fmt.Errorf("添加组织成员失败: %w", err)
	}

	organization.Creator = user
	return organization, nil
}

// GetOrganization 获取组织详情
func (s *organizationService) GetOrganization(ctx context.Context, id int64) (*domain.Organization, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	organization, err := s.organizationRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return organization, nil
}

// UpdateOrganization 更新组织信息
func (s *organizationService) UpdateOrganization(ctx context.Context, userID, orgID int64, name, description, logo, website string, isPublic *bool) (*domain.Organization, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	// 1. 检查权限
	if canManage, err := s.checkManagePermission(ctx, userID, orgID); err != nil {
		return nil, err
	} else if !canManage {
		return nil, domain.ErrPermissionDenied
	}

	// 2. 获取现有组织
	organization, err := s.organizationRepo.GetByID(ctx, orgID)
	if err != nil {
		return nil, err
	}

	// 3. 检查名称是否已被其他组织使用
	if name != "" && name != organization.Name {
		if existOrg, err := s.organizationRepo.GetByName(ctx, name); err != nil {
			if !errors.Is(err, domain.ErrOrganizationNotFound) {
				return nil, fmt.Errorf("检查组织名称失败: %w", err)
			}
		} else if existOrg != nil && existOrg.ID != orgID {
			return nil, domain.ErrOrganizationAlreadyExist
		}
	}

	// 4. 更新字段
	if name != "" {
		organization.Name = name
	}
	if description != "" {
		organization.Description = description
	}
	if logo != "" {
		organization.Logo = logo
	}
	if website != "" {
		organization.Website = website
	}
	if isPublic != nil {
		organization.IsPublic = *isPublic
	}

	// 5. 保存更新
	if err := s.organizationRepo.Update(ctx, organization); err != nil {
		return nil, fmt.Errorf("更新组织失败: %w", err)
	}

	return organization, nil
}

// DeleteOrganization 删除组织
func (s *organizationService) DeleteOrganization(ctx context.Context, userID, orgID int64) error {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	// 1. 检查是否为组织所有者
	if isOwner, err := s.IsOrganizationOwner(ctx, userID, orgID); err != nil {
		return err
	} else if !isOwner {
		return domain.ErrPermissionDenied
	}

	// 2. 删除组织
	if err := s.organizationRepo.Delete(ctx, orgID); err != nil {
		return fmt.Errorf("删除组织失败: %w", err)
	}

	return nil
}

// GetMyOrganizations 获取我的组织列表
func (s *organizationService) GetMyOrganizations(ctx context.Context, userID int64) ([]*domain.Organization, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	organizations, err := s.organizationRepo.GetUserOrganizations(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("获取用户组织失败: %w", err)
	}

	return organizations, nil
}

// GetPublicOrganizations 获取公开组织列表
func (s *organizationService) GetPublicOrganizations(ctx context.Context, limit, offset int) ([]*domain.Organization, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	organizations, err := s.organizationRepo.GetPublicOrganizations(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("获取公开组织失败: %w", err)
	}

	return organizations, nil
}

// SearchOrganizations 搜索组织
func (s *organizationService) SearchOrganizations(ctx context.Context, keyword string, isPublic *bool, limit, offset int) ([]*domain.Organization, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	organizations, err := s.organizationRepo.SearchOrganizations(ctx, keyword, isPublic, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("搜索组织失败: %w", err)
	}

	return organizations, nil
}

// InviteMember 邀请成员加入组织
func (s *organizationService) InviteMember(ctx context.Context, userID, orgID int64, email, message string, role domain.OrganizationMemberRole) error {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	// 1. 检查邀请者权限
	if canManage, err := s.checkManagePermission(ctx, userID, orgID); err != nil {
		return err
	} else if !canManage {
		return domain.ErrPermissionDenied
	}

	// 2. 验证邮箱格式
	if !utils.IsValidEmail(email) {
		return domain.ErrInvalidEmailAddress
	}

	// 3. 检查用户是否已是成员
	if user, err := s.userRepo.GetByEmail(ctx, email); err == nil {
		if member, err := s.organizationRepo.GetMember(ctx, orgID, user.ID); err == nil && member != nil {
			return domain.ErrUserAlreadyExist
		}
	}

	// 4. 生成邀请令牌
	token, err := utils.GenerateRandomString(32)
	if err != nil {
		return fmt.Errorf("生成邀请令牌失败: %w", err)
	}

	// 5. 创建邀请记录
	invitation := &domain.OrganizationInvitation{
		OrganizationID: orgID,
		Email:          email,
		Role:           role,
		Message:        message,
		Token:          token,
		InvitedBy:      userID,
		ExpiresAt:      time.Now().Add(7 * 24 * time.Hour), // 7天过期
	}

	if err := invitation.Validate(); err != nil {
		return err
	}

	// 6. 保存邀请
	if err := s.organizationRepo.StoreInvitation(ctx, invitation); err != nil {
		return fmt.Errorf("保存邀请失败: %w", err)
	}

	// 构建模板数据
	templateData := s.buildInvitationTemplateData(invitation, token)

	// 使用模板发送邮件
	subject := fmt.Sprintf("邀请您加入组织 - %s", "墨协") // 可以后续从组织信息中获取组织名称
	if err := s.emailService.SendOrganizationInvitationEmail(ctx, email, subject, common.NewJSONMap(templateData)); err != nil {
		return fmt.Errorf("发送邀请邮件失败: %w", err)
	}

	return nil
}

// buildInvitationTemplateData 构建邀请邮件模板数据
func (s *organizationService) buildInvitationTemplateData(invitation *domain.OrganizationInvitation, token string) map[string]interface{} {
	// 获取组织名称首字母作为图标
	organizationName := "墨协" // 可以后续从组织信息中获取
	organizationInitial := "墨"

	templateData := map[string]interface{}{
		"organizationInitial": organizationInitial,
		"organizationName":    organizationName,
		"roleDisplayName":     invitation.Role,

		"invitedAt": invitation.InvitedAt.Format("2006-01-02 15:04:05"),
		"expiresAt": invitation.ExpiresAt.Format("2006-01-02 15:04:05"),

		"message": invitation.Message,
		"token":   token,
	}

	// 可以添加邀请人信息（需要从数据库获取）
	if invitation.InvitedBy > 0 {
		inviter, err := s.userRepo.GetByID(context.Background(), invitation.InvitedBy)
		if err == nil {
			templateData["inviterName"] = inviter.Username
		}
	}

	// 可以添加接受邀请的URL
	templateData["acceptUrl"] = fmt.Sprintf("https://your-domain.com/organizations/invitations/accept?token=%s", token)

	return templateData
}

// ProcessJoinRequest 处理加入申请
func (s *organizationService) ProcessJoinRequest(ctx context.Context, userID, requestID int64, approve bool, note string) error {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	// 1. 获取申请记录
	request, err := s.organizationRepo.GetJoinRequest(ctx, requestID)
	if err != nil {
		return err
	}

	// 2. 检查处理者权限
	if canManage, err := s.checkManagePermission(ctx, userID, request.OrganizationID); err != nil {
		return err
	} else if !canManage {
		return domain.ErrPermissionDenied
	}

	// 3. 检查申请状态
	if !request.IsPending() {
		return domain.ErrBadParamInput
	}

	// 4. 处理申请
	if approve {
		request.Approve(userID, note)

		// 添加成员
		member := &domain.OrganizationMember{
			OrganizationID: request.OrganizationID,
			UserID:         request.UserID,
			Role:           domain.OrgRoleMember,
			InvitedBy:      userID,
		}

		if err := s.organizationRepo.AddMember(ctx, member); err != nil {
			return fmt.Errorf("添加组织成员失败: %w", err)
		}
	} else {
		request.Reject(userID, note)
	}

	// 5. 更新申请状态
	if err := s.organizationRepo.UpdateJoinRequest(ctx, request); err != nil {
		return fmt.Errorf("更新申请状态失败: %w", err)
	}

	return nil
}

// RequestJoinOrganization 申请加入组织
func (s *organizationService) RequestJoinOrganization(ctx context.Context, userID, orgID int64, message string) error {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	// 1. 检查组织是否存在且可加入
	organization, err := s.organizationRepo.GetByID(ctx, orgID)
	if err != nil {
		return err
	}

	if !organization.CanJoin() {
		return domain.ErrPermissionDenied
	}

	// 2. 检查是否已是成员
	if member, err := s.organizationRepo.GetMember(ctx, orgID, userID); err == nil && member != nil {
		return domain.ErrUserAlreadyExist
	}

	// 3. 检查是否已有待处理的申请
	userRequests, err := s.organizationRepo.GetJoinRequestsByUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("获取用户申请失败: %w", err)
	}

	for _, req := range userRequests {
		if req.OrganizationID == orgID && req.IsPending() {
			return domain.ErrBadParamInput
		}
	}

	// 4. 创建申请记录
	request := &domain.OrganizationJoinRequest{
		OrganizationID: orgID,
		UserID:         userID,
		Message:        message,
		Status:         domain.JoinRequestStatusPending,
	}

	if err := request.Validate(); err != nil {
		return err
	}

	// 5. 保存申请
	if err := s.organizationRepo.StoreJoinRequest(ctx, request); err != nil {
		return fmt.Errorf("保存加入申请失败: %w", err)
	}

	return nil
}

// UpdateMemberRole 更新成员角色
func (s *organizationService) UpdateMemberRole(ctx context.Context, userID, orgID, memberID int64, role domain.OrganizationMemberRole) error {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	// 1. 检查权限
	if canManage, err := s.checkManagePermission(ctx, userID, orgID); err != nil {
		return err
	} else if !canManage {
		return domain.ErrPermissionDenied
	}

	// 2. 不能修改自己的角色
	if userID == memberID {
		return domain.ErrPermissionDenied
	}

	// 3. 获取目标成员
	member, err := s.organizationRepo.GetMember(ctx, orgID, memberID)
	if err != nil {
		return err
	}

	// 4. 不能修改其他所有者的角色（除非自己也是所有者）
	if member.IsOwner() {
		if isOwner, err := s.IsOrganizationOwner(ctx, userID, orgID); err != nil {
			return err
		} else if !isOwner {
			return domain.ErrPermissionDenied
		}
	}

	// 5. 更新角色
	if err := s.organizationRepo.UpdateMemberRole(ctx, orgID, memberID, role); err != nil {
		return fmt.Errorf("更新成员角色失败: %w", err)
	}

	return nil
}

// RemoveMember 移除成员
func (s *organizationService) RemoveMember(ctx context.Context, userID, orgID, memberID int64) error {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	// 1. 检查权限
	if canManage, err := s.checkManagePermission(ctx, userID, orgID); err != nil {
		return err
	} else if !canManage {
		return domain.ErrPermissionDenied
	}

	// 2. 不能移除自己
	if userID == memberID {
		return domain.ErrPermissionDenied
	}

	// 3. 移除成员
	if err := s.organizationRepo.RemoveMember(ctx, orgID, memberID); err != nil {
		return fmt.Errorf("移除成员失败: %w", err)
	}

	return nil
}

// LeaveOrganization 退出组织
func (s *organizationService) LeaveOrganization(ctx context.Context, userID, orgID int64) error {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	// 1. 检查是否为成员
	member, err := s.organizationRepo.GetMember(ctx, orgID, userID)
	if err != nil {
		return err
	}

	// 2. 所有者不能直接退出，需要先转移所有权
	if member.IsOwner() {
		return domain.ErrPermissionDenied
	}

	// 3. 移除成员
	if err := s.organizationRepo.RemoveMember(ctx, orgID, userID); err != nil {
		return fmt.Errorf("退出组织失败: %w", err)
	}

	return nil
}

// GetOrganizationMembers 获取组织成员列表
func (s *organizationService) GetOrganizationMembers(ctx context.Context, userID, orgID int64) ([]*domain.OrganizationMember, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	// 1. 检查是否为组织成员
	if isMember, err := s.IsOrganizationMember(ctx, userID, orgID); err != nil {
		return nil, err
	} else if !isMember {
		return nil, domain.ErrPermissionDenied
	}

	// 2. 获取成员列表
	members, err := s.organizationRepo.GetMembers(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("获取组织成员失败: %w", err)
	}

	return members, nil
}

// GetMemberRole 获取成员角色
func (s *organizationService) GetMemberRole(ctx context.Context, userID, orgID int64) (domain.OrganizationMemberRole, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	member, err := s.organizationRepo.GetMember(ctx, orgID, userID)
	if err != nil {
		return "", err
	}

	return member.Role, nil
}

// CheckPermission 检查权限
func (s *organizationService) CheckPermission(ctx context.Context, userID, orgID int64, action string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	member, err := s.organizationRepo.GetMember(ctx, orgID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return false, nil
		}
		return false, err
	}

	switch action {
	case "manage":
		return member.CanManageMembers(), nil
	case "view":
		return true, nil
	default:
		return true, nil
	}
}

// IsOrganizationMember 是否为组织成员
func (s *organizationService) IsOrganizationMember(ctx context.Context, userID, orgID int64) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	_, err := s.organizationRepo.GetMember(ctx, orgID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// IsOrganizationAdmin 是否为组织管理员
func (s *organizationService) IsOrganizationAdmin(ctx context.Context, userID, orgID int64) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	member, err := s.organizationRepo.GetMember(ctx, orgID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return false, nil
		}
		return false, err
	}

	return member.IsAdmin() || member.IsOwner(), nil
}

// IsOrganizationOwner 是否为组织所有者
func (s *organizationService) IsOrganizationOwner(ctx context.Context, userID, orgID int64) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	member, err := s.organizationRepo.GetMember(ctx, orgID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return false, nil
		}
		return false, err
	}

	return member.IsOwner(), nil
}

// === 辅助方法 ===

// checkManagePermission 检查管理权限
func (s *organizationService) checkManagePermission(ctx context.Context, userID, orgID int64) (bool, error) {
	return s.IsOrganizationAdmin(ctx, userID, orgID)
}
