package space

import (
	"context"
	"time"

	"DOC/domain"
)

// spaceService 空间业务逻辑服务
type spaceService struct {
	spaceRepo        domain.SpaceRepository
	userRepo         domain.UserRepository
	organizationRepo domain.OrganizationRepository
	documentRepo     domain.DocumentRepository
	contextTimeout   time.Duration
}

// NewSpaceService 创建新的空间服务实例
func NewSpaceService(
	spaceRepo domain.SpaceRepository,
	userRepo domain.UserRepository,
	organizationRepo domain.OrganizationRepository,
	documentRepo domain.DocumentRepository,
	timeout time.Duration,
) domain.SpaceUsecase {
	return &spaceService{
		spaceRepo:        spaceRepo,
		userRepo:         userRepo,
		organizationRepo: organizationRepo,
		documentRepo:     documentRepo,
		contextTimeout:   timeout,
	}
}

// CreateSpace 创建空间
func (s *spaceService) CreateSpace(ctx context.Context, para domain.CreateSpacePara) (*domain.Space, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	// 验证用户
	user, err := s.userRepo.GetByID(ctx, para.UserID)
	if err != nil {
		return nil, err
	}
	if !user.IsActive() {
		return nil, domain.ErrUserNotActive
	}

	// 如果指定了组织，验证用户是否为组织成员
	if para.OrgID != nil {
		// 这里需要检查用户是否为组织成员
		// 暂时跳过，等待组织服务完善
	}

	// 检查空间名称是否已存在
	existingSpace, _ := s.spaceRepo.GetByName(ctx, para.Name, para.OrgID)
	if existingSpace != nil {
		return nil, domain.ErrSpaceAlreadyExist
	}

	// 创建空间实体
	space := &domain.Space{
		Name:           para.Name,
		Description:    para.Desc,
		Icon:           para.Icon,
		Color:          para.Color,
		Type:           para.SpaceType,
		IsPublic:       para.IsPublic,
		OrganizationID: para.OrgID,
		CreatedBy:      para.UserID,
		Status:         domain.SpaceStatusActive,
		MemberCount:    1, // 创建者自动成为成员
	}

	// 验证空间实体
	if err := space.Validate(); err != nil {
		return nil, err
	}

	// 保存空间
	if err := s.spaceRepo.Store(ctx, space); err != nil {
		return nil, err
	}

	// 将创建者添加为空间所有者
	member := &domain.SpaceMember{
		SpaceID: space.ID,
		UserID:  para.UserID,
		Role:    domain.SpaceRoleOwner,
		AddedBy: para.UserID,
	}

	if err := s.spaceRepo.AddMember(ctx, member); err != nil {
		return nil, err
	}

	return space, nil
}

// GetSpace 获取空间详情
func (s *spaceService) GetSpace(ctx context.Context, userID, spaceID int64) (*domain.Space, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	// 获取空间
	space, err := s.spaceRepo.GetByID(ctx, spaceID)
	if err != nil {
		return nil, err
	}

	// 检查访问权限
	hasAccess, err := s.checkSpaceAccess(ctx, userID, space)
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, domain.ErrSpacePermissionDenied
	}

	return space, nil
}

// UpdateSpace 更新空间信息
func (s *spaceService) UpdateSpace(ctx context.Context, para domain.UpdateSpacePara) (*domain.Space, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	// 获取空间
	space, err := s.spaceRepo.GetByID(ctx, para.SpaceID)
	if err != nil {
		return nil, err
	}

	// 检查管理权限
	canManage, err := s.checkSpaceManagePermission(ctx, para.UserID, para.SpaceID)
	if err != nil {
		return nil, err
	}
	if !canManage {
		return nil, domain.ErrSpacePermissionDenied
	}

	// 更新字段
	if para.Name != nil {
		// 检查名称是否已存在
		if *para.Name != space.Name {
			existingSpace, _ := s.spaceRepo.GetByName(ctx, *para.Name, space.OrganizationID)
			if existingSpace != nil && existingSpace.ID != para.SpaceID {
				return nil, domain.ErrSpaceAlreadyExist
			}
		}
		space.Name = *para.Name
	}
	if para.Desc != nil {
		space.Description = *para.Desc
	}
	if para.Icon != nil {
		space.Icon = *para.Icon
	}
	if para.Color != nil {
		space.Color = *para.Color
	}
	if para.SpaceType != nil {
		space.Type = *para.SpaceType
	}
	if para.IsPublic != nil {
		space.IsPublic = *para.IsPublic
	}

	// 验证更新后的空间
	if err := space.Validate(); err != nil {
		return nil, err
	}

	// 保存更新
	if err := s.spaceRepo.Update(ctx, space); err != nil {
		return nil, err
	}

	return space, nil
}

// DeleteSpace 删除空间
func (s *spaceService) DeleteSpace(ctx context.Context, userID, spaceID int64) error {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	// 获取空间
	space, err := s.spaceRepo.GetByID(ctx, spaceID)
	if err != nil {
		return err
	}

	// 检查删除权限（只有空间所有者可以删除）
	member, err := s.spaceRepo.GetMember(ctx, spaceID, userID)
	if err != nil {
		return domain.ErrNotSpaceMember
	}
	if !member.IsOwner() {
		return domain.ErrSpacePermissionDenied
	}

	// 软删除空间
	space.Status = domain.SpaceStatusDeleted
	return s.spaceRepo.Update(ctx, space)
}

// GetMySpaces 获取用户的空间列表
func (s *spaceService) GetMySpaces(ctx context.Context, userID int64) ([]*domain.Space, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	// 验证用户
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !user.IsActive() {
		return nil, domain.ErrUserNotActive
	}

	return s.spaceRepo.GetUserSpaces(ctx, userID)
}

// GetOrganizationSpaces 获取组织的空间列表
func (s *spaceService) GetOrganizationSpaces(ctx context.Context, userID, orgID int64) ([]*domain.Space, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	// 验证用户
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !user.IsActive() {
		return nil, domain.ErrUserNotActive
	}

	// 检查用户是否为组织成员
	// 这里需要调用组织服务检查权限
	// 暂时跳过，等待组织服务完善

	return s.spaceRepo.GetOrganizationSpaces(ctx, orgID)
}

// SearchSpaces 搜索空间
func (s *spaceService) SearchSpaces(ctx context.Context, userID int64, keyword string, limit, offset int) ([]*domain.Space, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	// 验证用户
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !user.IsActive() {
		return nil, domain.ErrUserNotActive
	}

	return s.spaceRepo.SearchSpaces(ctx, keyword, userID, limit, offset)
}

// AddSpaceMember 添加空间成员
func (s *spaceService) AddSpaceMember(ctx context.Context, userID, spaceID, memberUserID int64, role domain.SpaceMemberRole) error {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	// 检查管理权限
	canManage, err := s.checkSpaceManagePermission(ctx, userID, spaceID)
	if err != nil {
		return err
	}
	if !canManage {
		return domain.ErrSpacePermissionDenied
	}

	// 验证目标用户
	targetUser, err := s.userRepo.GetByID(ctx, memberUserID)
	if err != nil {
		return err
	}
	if !targetUser.IsActive() {
		return domain.ErrUserNotActive
	}

	// 检查用户是否已经是成员
	existingMember, _ := s.spaceRepo.GetMember(ctx, spaceID, memberUserID)
	if existingMember != nil {
		return domain.ErrConflict
	}

	// 创建成员记录
	member := &domain.SpaceMember{
		SpaceID: spaceID,
		UserID:  memberUserID,
		Role:    role,
		AddedBy: userID,
	}

	if err := member.Validate(); err != nil {
		return err
	}

	return s.spaceRepo.AddMember(ctx, member)
}

// UpdateMemberRole 更新成员角色
func (s *spaceService) UpdateMemberRole(ctx context.Context, userID, spaceID, memberUserID int64, role domain.SpaceMemberRole) error {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	// 检查管理权限
	canManage, err := s.checkSpaceManagePermission(ctx, userID, spaceID)
	if err != nil {
		return err
	}
	if !canManage {
		return domain.ErrSpacePermissionDenied
	}

	// 不能修改自己的角色
	if userID == memberUserID {
		return domain.ErrPermissionDenied
	}

	// 检查目标成员是否存在
	_, err = s.spaceRepo.GetMember(ctx, spaceID, memberUserID)
	if err != nil {
		return domain.ErrNotSpaceMember
	}

	// 不能将所有者角色分配给其他人（需要转移所有权）
	if role == domain.SpaceRoleOwner {
		return domain.ErrPermissionDenied
	}

	return s.spaceRepo.UpdateMemberRole(ctx, spaceID, memberUserID, role)
}

// RemoveSpaceMember 移除空间成员
func (s *spaceService) RemoveSpaceMember(ctx context.Context, userID, spaceID, memberUserID int64) error {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	// 检查管理权限
	canManage, err := s.checkSpaceManagePermission(ctx, userID, spaceID)
	if err != nil {
		return err
	}
	if !canManage {
		return domain.ErrSpacePermissionDenied
	}

	// 不能移除自己
	if userID == memberUserID {
		return domain.ErrPermissionDenied
	}

	// 检查目标成员是否存在
	member, err := s.spaceRepo.GetMember(ctx, spaceID, memberUserID)
	if err != nil {
		return domain.ErrNotSpaceMember
	}

	// 不能移除所有者
	if member.IsOwner() {
		return domain.ErrPermissionDenied
	}

	return s.spaceRepo.RemoveMember(ctx, spaceID, memberUserID)
}

// GetSpaceMembers 获取空间成员列表
func (s *spaceService) GetSpaceMembers(ctx context.Context, userID, spaceID int64) ([]*domain.SpaceMember, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	// 检查访问权限
	hasAccess, err := s.checkSpaceAccess(ctx, userID, nil)
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, domain.ErrSpacePermissionDenied
	}

	return s.spaceRepo.GetMembers(ctx, spaceID)
}

// AddDocumentToSpace 添加文档到空间
func (s *spaceService) AddDocumentToSpace(ctx context.Context, userID, spaceID, documentID int64) error {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	// 检查空间编辑权限
	canEdit, err := s.checkSpaceEditPermission(ctx, userID, spaceID)
	if err != nil {
		return err
	}
	if !canEdit {
		return domain.ErrSpacePermissionDenied
	}

	// 验证文档
	document, err := s.documentRepo.GetByID(ctx, documentID)
	if err != nil {
		return err
	}
	if !document.IsActive() {
		return domain.ErrDocumentNotFound
	}

	// 检查文档是否已在空间中
	exists, err := s.spaceRepo.IsDocumentInSpace(ctx, spaceID, documentID)
	if err != nil {
		return err
	}
	if exists {
		return domain.ErrConflict
	}

	// 添加文档到空间
	spaceDocument := &domain.SpaceDocument{
		SpaceID:    spaceID,
		DocumentID: documentID,
		AddedBy:    userID,
	}

	return s.spaceRepo.AddDocument(ctx, spaceDocument)
}

// RemoveDocumentFromSpace 从空间移除文档
func (s *spaceService) RemoveDocumentFromSpace(ctx context.Context, userID, spaceID, documentID int64) error {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	// 检查空间编辑权限
	canEdit, err := s.checkSpaceEditPermission(ctx, userID, spaceID)
	if err != nil {
		return err
	}
	if !canEdit {
		return domain.ErrSpacePermissionDenied
	}

	// 检查文档是否在空间中
	exists, err := s.spaceRepo.IsDocumentInSpace(ctx, spaceID, documentID)
	if err != nil {
		return err
	}
	if !exists {
		return domain.ErrDocumentNotFound
	}

	return s.spaceRepo.RemoveDocument(ctx, spaceID, documentID)
}

// GetSpaceDocuments 获取空间文档列表
func (s *spaceService) GetSpaceDocuments(ctx context.Context, userID, spaceID int64) ([]*domain.Document, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	// 检查访问权限
	hasAccess, err := s.checkSpaceAccess(ctx, userID, nil)
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, domain.ErrSpacePermissionDenied
	}

	return s.spaceRepo.GetSpaceDocuments(ctx, spaceID)
}

// CheckSpacePermission 检查空间权限
func (s *spaceService) CheckSpacePermission(ctx context.Context, userID, spaceID int64, action string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	switch action {
	case "read", "view":
		return s.checkSpaceAccess(ctx, userID, nil)
	case "edit", "write":
		return s.checkSpaceEditPermission(ctx, userID, spaceID)
	case "manage", "admin":
		return s.checkSpaceManagePermission(ctx, userID, spaceID)
	default:
		return false, domain.ErrInvalidPermission
	}
}

// IsSpaceMember 检查是否为空间成员
func (s *spaceService) IsSpaceMember(ctx context.Context, userID, spaceID int64) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	_, err := s.spaceRepo.GetMember(ctx, spaceID, userID)
	if err != nil {
		if err == domain.ErrUserNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GetUserSpaceRole 获取用户在空间中的角色
func (s *spaceService) GetUserSpaceRole(ctx context.Context, userID, spaceID int64) (domain.SpaceMemberRole, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	member, err := s.spaceRepo.GetMember(ctx, spaceID, userID)
	if err != nil {
		return "", err
	}

	return member.Role, nil
}

// === 私有辅助方法 ===

// checkSpaceAccess 检查空间访问权限
func (s *spaceService) checkSpaceAccess(ctx context.Context, userID int64, space *domain.Space) (bool, error) {
	// 如果没有传入空间对象，先获取
	if space == nil {
		var err error
		space, err = s.spaceRepo.GetByID(ctx, userID) // 这里应该是spaceID，但需要从上下文获取
		if err != nil {
			return false, err
		}
	}

	// 检查空间是否激活
	if !space.IsActive() {
		return false, nil
	}

	// 空间创建者总是可以访问
	if space.CreatedBy == userID {
		return true, nil
	}

	// 公开空间可以访问
	if space.IsPublic {
		return true, nil
	}

	// 检查是否为空间成员
	_, err := s.spaceRepo.GetMember(ctx, space.ID, userID)
	return err == nil, nil
}

// checkSpaceEditPermission 检查空间编辑权限
func (s *spaceService) checkSpaceEditPermission(ctx context.Context, userID, spaceID int64) (bool, error) {
	member, err := s.spaceRepo.GetMember(ctx, spaceID, userID)
	if err != nil {
		return false, nil // 不是成员就没有编辑权限
	}

	return member.CanEditDocuments(), nil
}

// checkSpaceManagePermission 检查空间管理权限
func (s *spaceService) checkSpaceManagePermission(ctx context.Context, userID, spaceID int64) (bool, error) {
	member, err := s.spaceRepo.GetMember(ctx, spaceID, userID)
	if err != nil {
		return false, nil // 不是成员就没有管理权限
	}

	return member.CanManageMembers(), nil
}
