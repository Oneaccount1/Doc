package dto

import "DOC/domain"

// CreateSpaceDto 创建空间请求DTO
type CreateSpaceDto struct {
	Name           string           `json:"name" validate:"required,min=1,max=100"`
	Description    string           `json:"description,omitempty" validate:"max=1000"`
	Icon           string           `json:"icon,omitempty" validate:"max=200"`
	Color          string           `json:"color,omitempty" validate:"max=50"`
	Type           domain.SpaceType `json:"type,omitempty" validate:"omitempty,oneof=WORKSPACE PROJECT PERSONAL"`
	IsPublic       bool             `json:"is_public"`
	OrganizationID *int64           `json:"organization_id,omitempty"`
}

// UpdateSpaceDto 更新空间请求DTO
type UpdateSpaceDto struct {
	Name        *string           `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Description *string           `json:"description,omitempty" validate:"omitempty,max=1000"`
	Icon        *string           `json:"icon,omitempty" validate:"omitempty,max=200"`
	Color       *string           `json:"color,omitempty" validate:"omitempty,max=50"`
	Type        *domain.SpaceType `json:"type,omitempty" validate:"omitempty,oneof=WORKSPACE PROJECT PERSONAL"`
	IsPublic    *bool             `json:"is_public,omitempty"`
	IsDeleted   *bool             `json:"is_deleted,omitempty"`
}

// AddMemberDto 添加空间成员请求DTO
type AddMemberDto struct {
	UserID int64                  `json:"userId" validate:"required,min=1"`
	Role   domain.SpaceMemberRole `json:"role" validate:"required,oneof=OWNER ADMIN EDITOR VIEWER GUEST"`
}

// UpdateMemberRoleDto 更新成员角色请求DTO
type UpdateMemberRoleDto struct {
	Role domain.SpaceMemberRole `json:"role" validate:"required,oneof=OWNER ADMIN EDITOR VIEWER GUEST"`
}

// AddDocumentDto 添加文档到空间请求DTO
type AddDocumentDto struct {
	DocumentID int64 `json:"documentId" validate:"required,min=1"`
}

// SpaceResponse 空间响应DTO
type SpaceResponse struct {
	ID             int64              `json:"id"`
	Name           string             `json:"name"`
	Description    string             `json:"description"`
	Icon           string             `json:"icon"`
	Color          string             `json:"color"`
	Type           domain.SpaceType   `json:"type"`
	IsPublic       bool               `json:"is_public"`
	IsDeleted      bool               `json:"is_deleted"`
	Status         domain.SpaceStatus `json:"status"`
	OrganizationID *int64             `json:"organization_id"`
	CreatedBy      int64              `json:"created_by"`
	CreatedAt      string             `json:"created_at"`
	UpdatedAt      string             `json:"updated_at"`
	DocumentCount  int                `json:"document_count"`
	MemberCount    int                `json:"member_count"`

	// 关联信息
	Organization *OrganizationResponse  `json:"organization,omitempty"`
	Creator      *UserResponse          `json:"creator,omitempty"`
	Members      []*SpaceMemberResponse `json:"members,omitempty"`
	Documents    []*DocumentResponseDto `json:"documents,omitempty"`
}

// SpaceMemberResponse 空间成员响应DTO
type SpaceMemberResponse struct {
	ID      int64                  `json:"id"`
	SpaceID int64                  `json:"space_id"`
	UserID  int64                  `json:"user_id"`
	Role    domain.SpaceMemberRole `json:"role"`
	AddedBy int64                  `json:"added_by"`
	AddedAt string                 `json:"added_at"`

	// 关联用户信息
	User        *UserResponse `json:"user,omitempty"`
	AddedByUser *UserResponse `json:"added_by_user,omitempty"`
}

// SpaceDocumentResponse 空间文档响应DTO
type SpaceDocumentResponse struct {
	ID         int64  `json:"id"`
	SpaceID    int64  `json:"space_id"`
	DocumentID int64  `json:"document_id"`
	AddedBy    int64  `json:"added_by"`
	AddedAt    string `json:"added_at"`

	// 关联信息
	Document *DocumentResponseDto `json:"document,omitempty"`
}

// SpaceListResponse 空间列表响应DTO
type SpaceListResponse struct {
	Spaces []*SpaceResponse `json:"spaces"`
	Total  int64            `json:"total"`
	Page   int              `json:"page"`
	Size   int              `json:"size"`
}

// SpaceMemberListResponse 空间成员列表响应DTO
type SpaceMemberListResponse struct {
	Members []*SpaceMemberResponse `json:"members"`
	Total   int64                  `json:"total"`
}

// SpaceDocumentListResponse 空间文档列表响应DTO
type SpaceDocumentListResponse struct {
	Documents []*DocumentResponseDto `json:"documents"`
	Total     int64                  `json:"total"`
}

// === 转换方法 ===

// ToSpaceResponse 将Space实体转换为响应DTO
func ToSpaceResponse(space *domain.Space) *SpaceResponse {
	if space == nil {
		return nil
	}

	response := &SpaceResponse{
		ID:             space.ID,
		Name:           space.Name,
		Description:    space.Description,
		Icon:           space.Icon,
		Color:          space.Color,
		Type:           space.Type,
		IsPublic:       space.IsPublic,
		Status:         space.Status,
		OrganizationID: space.OrganizationID,
		CreatedBy:      space.CreatedBy,
		CreatedAt:      space.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:      space.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		DocumentCount:  space.DocumentCount,
		MemberCount:    space.MemberCount,
	}

	// 转换关联数据
	if space.Organization != nil {
		response.Organization = ToOrganizationResponse(space.Organization)
	}
	if space.Creator != nil {
		response.Creator = ToUserResponse(space.Creator)
	}
	if len(space.Members) > 0 {
		response.Members = make([]*SpaceMemberResponse, len(space.Members))
		for i, member := range space.Members {
			response.Members[i] = ToSpaceMemberResponse(member)
		}
	}
	if len(space.Documents) > 0 {
		response.Documents = make([]*DocumentResponseDto, len(space.Documents))
		for i, doc := range space.Documents {
			response.Documents[i] = FromDocument(doc)
		}
	}

	return response
}

// ToSpaceMemberResponse 将SpaceMember实体转换为响应DTO
func ToSpaceMemberResponse(member *domain.SpaceMember) *SpaceMemberResponse {
	if member == nil {
		return nil
	}

	response := &SpaceMemberResponse{
		ID:      member.ID,
		SpaceID: member.SpaceID,
		UserID:  member.UserID,
		Role:    member.Role,
		AddedBy: member.AddedBy,
		AddedAt: member.AddedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	// 转换关联数据
	if member.User != nil {
		response.User = ToUserResponse(member.User)
	}
	if member.AddedByUser != nil {
		response.AddedByUser = ToUserResponse(member.AddedByUser)
	}

	return response
}

// ToSpaceListResponse 将Space列表转换为响应DTO
func ToSpaceListResponse(spaces []*domain.Space, total int64, page, size int) *SpaceListResponse {
	response := &SpaceListResponse{
		Total: total,
		Page:  page,
		Size:  size,
	}

	if len(spaces) > 0 {
		response.Spaces = make([]*SpaceResponse, len(spaces))
		for i, space := range spaces {
			response.Spaces[i] = ToSpaceResponse(space)
		}
	} else {
		response.Spaces = []*SpaceResponse{}
	}

	return response
}

// ToSpaceMemberListResponse 将SpaceMember列表转换为响应DTO
func ToSpaceMemberListResponse(members []*domain.SpaceMember) *SpaceMemberListResponse {
	response := &SpaceMemberListResponse{
		Total: int64(len(members)),
	}

	if len(members) > 0 {
		response.Members = make([]*SpaceMemberResponse, len(members))
		for i, member := range members {
			response.Members[i] = ToSpaceMemberResponse(member)
		}
	} else {
		response.Members = []*SpaceMemberResponse{}
	}

	return response
}
