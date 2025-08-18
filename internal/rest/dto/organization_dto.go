package dto

import "DOC/domain"

// === 请求结构体 ===

// CreateOrganizationRequest 创建组织请求结构（对应 CreateOrganizationDto）
type CreateOrganizationRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=100"` // 组织名称
	Description string `json:"description" binding:"max=1000"`        // 组织描述
	Logo        string `json:"logo" binding:"max=500"`                // 组织图标URL
	Website     string `json:"website" binding:"max=500"`             // 网站地址
	IsPublic    bool   `json:"is_public"`                             // 是否公开组织
}

// UpdateOrganizationRequest 更新组织请求结构（对应 UpdateOrganizationDto）
type UpdateOrganizationRequest struct {
	Name        string `json:"name" binding:"max=100"`         // 组织名称
	Description string `json:"description" binding:"max=1000"` // 组织描述
	Logo        string `json:"logo" binding:"max=500"`         // 组织图标URL
	Website     string `json:"website" binding:"max=500"`      // 网站地址
	IsPublic    *bool  `json:"is_public"`                      // 是否公开组织
}

// InviteMemberRequest 邀请成员请求结构（对应 InviteMemberDto）
type InviteMemberRequest struct {
	Email   string                        `json:"email" binding:"required,email"`                   // 被邀请人邮箱
	Message string                        `json:"message" binding:"max=500"`                        // 邀请消息
	Role    domain.OrganizationMemberRole `json:"role" binding:"required,oneof=OWNER ADMIN MEMBER"` // 邀请人角色
}

// JoinRequestRequest 申请加入组织请求结构（对应 JoinRequestDto）
type JoinRequestRequest struct {
	Message string `json:"message" binding:"max=500"` // 申请消息
}

// ProcessJoinRequestRequest 处理加入申请请求结构
type ProcessJoinRequestRequest struct {
	Approve bool   `json:"approve"`                // 是否批准
	Note    string `json:"note" binding:"max=500"` // 处理备注
}

// UpdateMemberRoleRequest 更新成员角色请求结构（对应 UpdateMemberRoleDto）
type UpdateMemberRoleRequest struct {
	Role domain.OrganizationMemberRole `json:"role" binding:"required,oneof=OWNER ADMIN MEMBER"` // 成员角色
}

// GetOrganizationsRequest 获取组织列表请求结构
type GetOrganizationsRequest struct {
	IsPublic *bool  `form:"isPublic"`         // 是否公开
	Keyword  string `form:"keyword"`          // 搜索关键词
	Limit    int    `form:"limit,default=20"` // 每页数量
	Offset   int    `form:"offset,default=0"` // 偏移量
}

// === 响应结构体 ===

// OrganizationResponse 组织响应结构
type OrganizationResponse struct {
	ID          int64  `json:"id"`           // 组织ID
	Name        string `json:"name"`         // 组织名称
	Description string `json:"description"`  // 组织描述
	Logo        string `json:"logo"`         // 组织图标URL
	Website     string `json:"website"`      // 网站地址
	IsPublic    bool   `json:"is_public"`    // 是否公开
	MemberCount int    `json:"member_count"` // 成员数量
	SpaceCount  int    `json:"space_count"`  // 空间数量
	CreatedAt   string `json:"created_at"`   // 创建时间
	UpdatedAt   string `json:"updated_at"`   // 更新时间

	// 扩展信息（根据上下文可选包含）
	Creator *UserResponse                 `json:"creator,omitempty"` // 创建者信息
	Members []*OrganizationMemberResponse `json:"members,omitempty"` // 成员列表
}

// OrganizationMemberResponse 组织成员响应结构
type OrganizationMemberResponse struct {
	ID             int64                         `json:"id"`              // 成员ID
	OrganizationID int64                         `json:"organization_id"` // 组织ID
	UserID         int64                         `json:"user_id"`         // 用户ID
	Role           domain.OrganizationMemberRole `json:"role"`            // 角色
	JoinedAt       string                        `json:"joined_at"`       // 加入时间

	// 关联用户信息
	User    *UserResponse `json:"user,omitempty"`    // 用户信息
	Inviter *UserResponse `json:"inviter,omitempty"` // 邀请者信息
}

// OrganizationInvitationResponse 组织邀请响应结构
type OrganizationInvitationResponse struct {
	ID             int64                         `json:"id"`              // 邀请ID
	OrganizationID int64                         `json:"organization_id"` // 组织ID
	Email          string                        `json:"email"`           // 邮箱
	Role           domain.OrganizationMemberRole `json:"role"`            // 角色
	Message        string                        `json:"message"`         // 邀请消息
	IsUsed         bool                          `json:"is_used"`         // 是否已使用
	InvitedAt      string                        `json:"invited_at"`      // 邀请时间
	ExpiresAt      string                        `json:"expires_at"`      // 过期时间
	AcceptedAt     *string                       `json:"accepted_at"`     // 接受时间

	// 关联信息
	Organization *OrganizationResponse `json:"organization,omitempty"` // 组织信息
	Inviter      *UserResponse         `json:"inviter,omitempty"`      // 邀请者信息
}

// OrganizationJoinRequestResponse 组织加入申请响应结构
type OrganizationJoinRequestResponse struct {
	ID             int64                    `json:"id"`              // 申请ID
	OrganizationID int64                    `json:"organization_id"` // 组织ID
	UserID         int64                    `json:"user_id"`         // 用户ID
	Message        string                   `json:"message"`         // 申请消息
	Status         domain.JoinRequestStatus `json:"status"`          // 申请状态
	ProcessNote    string                   `json:"process_note"`    // 处理备注
	CreatedAt      string                   `json:"created_at"`      // 申请时间
	ProcessedAt    *string                  `json:"processed_at"`    // 处理时间

	// 关联信息
	Organization *OrganizationResponse `json:"organization,omitempty"` // 组织信息
	User         *UserResponse         `json:"user,omitempty"`         // 申请用户信息
	Processor    *UserResponse         `json:"processor,omitempty"`    // 处理者信息
}

// OrganizationListResponse 组织列表响应结构
type OrganizationListResponse struct {
	Organizations []*OrganizationResponse `json:"organizations"` // 组织列表
	Total         int64                   `json:"total"`         // 总数
	//Limit         int                     `json:"limit"`         // 每页数量
	//Offset        int                     `json:"offset"`        // 偏移量
}

// === 转换函数 ===

// ToOrganizationResponse 将组织实体转换为响应结构
func ToOrganizationResponse(org *domain.Organization) *OrganizationResponse {
	if org == nil {
		return nil
	}

	response := &OrganizationResponse{
		ID:          org.ID,
		Name:        org.Name,
		Description: org.Description,
		Logo:        org.Logo,
		Website:     org.Website,
		IsPublic:    org.IsPublic,
		MemberCount: org.MemberCount,
		SpaceCount:  org.SpaceCount,
		CreatedAt:   org.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   org.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	// 可选的关联数据
	if org.Creator != nil {
		response.Creator = ToUserResponse(org.Creator)
	}

	if len(org.Members) > 0 {
		response.Members = make([]*OrganizationMemberResponse, len(org.Members))
		for i, member := range org.Members {
			response.Members[i] = ToOrganizationMemberResponse(member)
		}
	}

	return response
}

// ToOrganizationMemberResponse 将组织成员实体转换为响应结构
func ToOrganizationMemberResponse(member *domain.OrganizationMember) *OrganizationMemberResponse {
	if member == nil {
		return nil
	}

	response := &OrganizationMemberResponse{
		ID:             member.ID,
		OrganizationID: member.OrganizationID,
		UserID:         member.UserID,
		Role:           member.Role,
		JoinedAt:       member.JoinedAt.Format("2006-01-02T15:04:05Z"),
	}

	// 可选的关联数据
	if member.User != nil {
		response.User = ToUserResponse(member.User)
	}

	if member.Inviter != nil {
		response.Inviter = ToUserResponse(member.Inviter)
	}

	return response
}

// ToOrganizationInvitationResponse 将组织邀请实体转换为响应结构
func ToOrganizationInvitationResponse(invitation *domain.OrganizationInvitation) *OrganizationInvitationResponse {
	if invitation == nil {
		return nil
	}

	response := &OrganizationInvitationResponse{
		ID:             invitation.ID,
		OrganizationID: invitation.OrganizationID,
		Email:          invitation.Email,
		Role:           invitation.Role,
		Message:        invitation.Message,
		IsUsed:         invitation.IsUsed,
		InvitedAt:      invitation.InvitedAt.Format("2006-01-02T15:04:05Z"),
		ExpiresAt:      invitation.ExpiresAt.Format("2006-01-02T15:04:05Z"),
	}

	// 处理可选的时间字段
	if invitation.AcceptedAt != nil {
		acceptedAt := invitation.AcceptedAt.Format("2006-01-02T15:04:05Z")
		response.AcceptedAt = &acceptedAt
	}

	// 可选的关联数据
	if invitation.Organization != nil {
		response.Organization = ToOrganizationResponse(invitation.Organization)
	}

	if invitation.Inviter != nil {
		response.Inviter = ToUserResponse(invitation.Inviter)
	}

	return response
}

// ToOrganizationJoinRequestResponse 将组织加入申请实体转换为响应结构
func ToOrganizationJoinRequestResponse(request *domain.OrganizationJoinRequest) *OrganizationJoinRequestResponse {
	if request == nil {
		return nil
	}

	response := &OrganizationJoinRequestResponse{
		ID:             request.ID,
		OrganizationID: request.OrganizationID,
		UserID:         request.UserID,
		Message:        request.Message,
		Status:         request.Status,
		ProcessNote:    request.ProcessNote,
		CreatedAt:      request.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	// 处理可选的时间字段
	if request.ProcessedAt != nil {
		processedAt := request.ProcessedAt.Format("2006-01-02T15:04:05Z")
		response.ProcessedAt = &processedAt
	}

	// 可选的关联数据
	if request.Organization != nil {
		response.Organization = ToOrganizationResponse(request.Organization)
	}

	if request.User != nil {
		response.User = ToUserResponse(request.User)
	}

	if request.Processor != nil {
		response.Processor = ToUserResponse(request.Processor)
	}

	return response
}
