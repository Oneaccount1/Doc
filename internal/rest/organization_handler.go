package rest

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"DOC/domain"
	"DOC/internal/rest/dto"
)

// OrganizationHandler 组织 HTTP 处理器
// 负责处理组织相关的 HTTP 请求
type OrganizationHandler struct {
	organizationUsecase domain.OrganizationUsecase
}

// NewOrganizationHandler 创建新的组织处理器
func NewOrganizationHandler(organizationUsecase domain.OrganizationUsecase) *OrganizationHandler {
	return &OrganizationHandler{
		organizationUsecase: organizationUsecase,
	}
}

// CreateOrganization 创建新组织
// @Summary 创建新组织
// @Description 创建新的组织
// @Tags organizations
// @Accept json
// @Produce json
// @Param organization body dto.CreateOrganizationRequest true "组织信息"
// @Success 201 {object} dto.OrganizationResponse "创建成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 409 {object} ErrorResponse "组织已存在"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Security BearerAuth
// @Router /api/v1/organizations [post]
func (h *OrganizationHandler) CreateOrganization(c *gin.Context) {
	var req dto.CreateOrganizationRequest

	// 1. 绑定和验证请求数据
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseBadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 2. 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		ResponseUnauthorized(c, "未授权访问")
		return
	}

	uid, ok := userID.(int64)
	if !ok {
		ResponseInternalServerError(c, "用户ID格式错误")
		return
	}

	// 3. 调用业务逻辑创建组织
	organization, err := h.organizationUsecase.CreateOrganization(
		c.Request.Context(),
		uid,
		req.Name,
		req.Description,
		req.Logo,
		req.Website,
		req.IsPublic,
	)
	if err != nil {
		switch err {
		case domain.ErrOrganizationAlreadyExist:
			ResponseConflict(c, "组织名称已存在")
		case domain.ErrInvalidOrganizationName:
			ResponseBadRequest(c, "组织名称无效")
		default:
			ResponseInternalServerError(c, "创建组织失败"+err.Error())
		}
		return
	}

	// 4. 返回成功响应
	ResponseCreated(c, "创建成功", dto.ToOrganizationResponse(organization))
}

// GetOrganizations 获取组织列表
// @Summary 获取组织列表
// @Description 获取组织列表（公开组织）
// @Tags organizations
// @Accept json
// @Produce json
// @Param isPublic query bool false "是否公开"
// @Param keyword query string false "搜索关键词"
// @Param limit query int false "每页数量" default(20)
// @Param offset query int false "偏移量" default(0)
// @Success 200 {object} dto.OrganizationListResponse "获取成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Security BearerAuth
// @Router /api/v1/organizations [get]
func (h *OrganizationHandler) GetOrganizations(c *gin.Context) {
	var req dto.GetOrganizationsRequest

	// 1. 绑定查询参数
	if err := c.ShouldBindQuery(&req); err != nil {
		ResponseBadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 2. 获取组织列表
	organizations, err := h.organizationUsecase.SearchOrganizations(c, req.Keyword, req.IsPublic, req.Limit, req.Offset)
	//var err error
	//
	//if req.Keyword != "" {
	//	// 搜索组织
	//	organizations, err = h.organizationUsecase.SearchOrganizations(
	//		c.Request.Context(),
	//		req.Keyword,
	//		req.IsPublic,
	//		req.Limit,
	//		req.Offset,
	//	)
	//} else {
	//	// 获取公开组织
	//	organizations, err = h.organizationUsecase.GetPublicOrganizations(
	//		c.Request.Context(),
	//		req.Limit,
	//		req.Offset,
	//	)
	//}

	if err != nil {
		ResponseInternalServerError(c, "获取组织列表失败")
		return
	}

	// 3. 构建响应
	response := &dto.OrganizationListResponse{
		Organizations: make([]*dto.OrganizationResponse, len(organizations)),
		Total:         int64(len(organizations)),
	}

	for i, org := range organizations {
		response.Organizations[i] = dto.ToOrganizationResponse(org)
	}

	ResponseOK(c, "获取成功", response)
}

// GetMyOrganizations 获取我的组织
// @Summary 获取我的组织
// @Description 获取当前用户的组织列表
// @Tags organizations
// @Accept json
// @Produce json
// @Success 200 {array} dto.OrganizationResponse "获取成功"
// @Failure 401 {object} ErrorResponse "未授权"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Security BearerAuth
// @Router /api/v1/organizations/my [get]
func (h *OrganizationHandler) GetMyOrganizations(c *gin.Context) {
	// 1. 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		ResponseUnauthorized(c, "未授权访问")
		return
	}

	uid, ok := userID.(int64)
	if !ok {
		ResponseInternalServerError(c, "用户ID格式错误")
		return
	}

	// 2. 获取用户的组织列表
	organizations, err := h.organizationUsecase.GetMyOrganizations(c.Request.Context(), uid)
	if err != nil {
		ResponseInternalServerError(c, "获取组织列表失败")
		return
	}

	// 3. 转换为响应格式
	response := make([]*dto.OrganizationResponse, len(organizations))
	for i, org := range organizations {
		response[i] = dto.ToOrganizationResponse(org)
	}

	ResponseOK(c, "获取成功", response)
}

// GetOrganization 获取组织详情
// @Summary 获取组织详情
// @Description 获取指定组织的详细信息
// @Tags organizations
// @Accept json
// @Produce json
// @Param id path int true "组织ID"
// @Success 200 {object} dto.OrganizationResponse "获取成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 404 {object} ErrorResponse "组织不存在"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Security BearerAuth
// @Router /api/v1/organizations/{id} [get]
func (h *OrganizationHandler) GetOrganization(c *gin.Context) {
	// 1. 获取组织ID
	orgID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		ResponseBadRequest(c, "组织ID格式错误")
		return
	}

	// 2. 获取组织信息
	organization, err := h.organizationUsecase.GetOrganization(c.Request.Context(), orgID)
	if err != nil {
		switch err {
		case domain.ErrOrganizationNotFound:
			ResponseNotFound(c, "组织不存在")
		default:
			ResponseInternalServerError(c, "获取组织信息失败")
		}
		return
	}

	ResponseOK(c, "获取成功", dto.ToOrganizationResponse(organization))
}

// UpdateOrganization 更新组织信息
// @Summary 更新组织信息
// @Description 更新指定组织的信息
// @Tags organizations
// @Accept json
// @Produce json
// @Param id path int true "组织ID"
// @Param organization body dto.UpdateOrganizationRequest true "组织信息"
// @Success 200 {object} dto.OrganizationResponse "更新成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 403 {object} ErrorResponse "权限不足"
// @Failure 404 {object} ErrorResponse "组织不存在"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Security BearerAuth
// @Router /api/v1/organizations/{id} [put]
func (h *OrganizationHandler) UpdateOrganization(c *gin.Context) {
	var req dto.UpdateOrganizationRequest

	// 1. 获取组织ID
	orgID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		ResponseBadRequest(c, "组织ID格式错误")
		return
	}

	// 2. 绑定和验证请求数据
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseBadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 3. 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		ResponseUnauthorized(c, "未授权访问")
		return
	}

	uid, ok := userID.(int64)
	if !ok {
		ResponseInternalServerError(c, "用户ID格式错误")
		return
	}

	// 4. 更新组织信息
	organization, err := h.organizationUsecase.UpdateOrganization(
		c.Request.Context(),
		uid,
		orgID,
		req.Name,
		req.Description,
		req.Logo,
		req.Website,
		req.IsPublic,
	)
	if err != nil {
		switch err {
		case domain.ErrOrganizationNotFound:
			ResponseNotFound(c, "组织不存在")
		case domain.ErrPermissionDenied:
			ResponseForbidden(c, "权限不足")
		case domain.ErrOrganizationAlreadyExist:
			ResponseConflict(c, "组织名称已存在")
		default:
			ResponseInternalServerError(c, "更新组织失败")
		}
		return
	}

	ResponseOK(c, "更新成功", dto.ToOrganizationResponse(organization))
}

// DeleteOrganization 删除组织
// @Summary 删除组织
// @Description 删除指定组织
// @Tags organizations
// @Accept json
// @Produce json
// @Param id path int true "组织ID"
// @Success 200 {object} SuccessResponse "删除成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 403 {object} ErrorResponse "权限不足"
// @Failure 404 {object} ErrorResponse "组织不存在"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Security BearerAuth
// @Router /api/v1/organizations/{id} [delete]
func (h *OrganizationHandler) DeleteOrganization(c *gin.Context) {
	// 1. 获取组织ID
	orgID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		ResponseBadRequest(c, "组织ID格式错误")
		return
	}

	// 2. 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		ResponseUnauthorized(c, "未授权访问")
		return
	}

	uid, ok := userID.(int64)
	if !ok {
		ResponseInternalServerError(c, "用户ID格式错误")
		return
	}

	// 3. 删除组织
	if err := h.organizationUsecase.DeleteOrganization(c.Request.Context(), uid, orgID); err != nil {
		switch err {
		case domain.ErrOrganizationNotFound:
			ResponseNotFound(c, "组织不存在")
		case domain.ErrPermissionDenied:
			ResponseForbidden(c, "权限不足")
		default:
			ResponseInternalServerError(c, "删除组织失败")
		}
		return
	}

	ResponseOK(c, "删除成功", map[string]interface{}{
		"message": "组织删除成功",
		"success": true,
	})
}

// InviteMember 邀请成员加入组织
// @Summary 邀请成员加入组织
// @Description 向指定邮箱发送组织邀请
// @Tags organizations
// @Accept json
// @Produce json
// @Param id path int true "组织ID"
// @Param invitation body dto.InviteMemberRequest true "邀请信息"
// @Success 201 {object} SuccessResponse "邀请发送成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 403 {object} ErrorResponse "权限不足"
// @Failure 404 {object} ErrorResponse "组织不存在"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Security BearerAuth
// @Router /api/v1/organizations/{id}/invite [post]
func (h *OrganizationHandler) InviteMember(c *gin.Context) {
	var req dto.InviteMemberRequest

	// 1. 获取组织ID
	orgID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		ResponseBadRequest(c, "组织ID格式错误")
		return
	}

	// 2. 绑定和验证请求数据
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseBadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 3. 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		ResponseUnauthorized(c, "未授权访问")
		return
	}

	uid, ok := userID.(int64)
	if !ok {
		ResponseInternalServerError(c, "用户ID格式错误")
		return
	}

	// 4. 发送邀请
	if err := h.organizationUsecase.InviteMember(
		c.Request.Context(),
		uid,
		orgID,
		req.Email,
		req.Message,
		req.Role,
	); err != nil {
		switch err {
		case domain.ErrOrganizationNotFound:
			ResponseNotFound(c, "组织不存在")
		case domain.ErrPermissionDenied:
			ResponseForbidden(c, "权限不足")
		case domain.ErrInvalidEmailAddress:
			ResponseBadRequest(c, "邮箱格式无效")
		case domain.ErrUserAlreadyExist:
			ResponseConflict(c, "用户已是组织成员")
		default:
			ResponseInternalServerError(c, "发送邀请失败"+err.Error())
		}
		return
	}

	ResponseCreated(c, "邀请发送成功", map[string]interface{}{
		"message": "邀请已发送到指定邮箱",
		"success": true,
	})
}

// RequestJoinOrganization 申请加入组织
// @Summary 申请加入组织
// @Description 向组织提交加入申请
// @Tags organizations
// @Accept json
// @Produce json
// @Param id path int true "组织ID"
// @Param request body dto.JoinRequestRequest true "申请信息"
// @Success 201 {object} SuccessResponse "申请提交成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 403 {object} ErrorResponse "权限不足"
// @Failure 404 {object} ErrorResponse "组织不存在"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Security BearerAuth
// @Router /api/v1/organizations/{id}/join-request [post]
func (h *OrganizationHandler) RequestJoinOrganization(c *gin.Context) {
	var req dto.JoinRequestRequest

	// 1. 获取组织ID
	orgID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		ResponseBadRequest(c, "组织ID格式错误")
		return
	}

	// 2. 绑定和验证请求数据
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseBadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 3. 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		ResponseUnauthorized(c, "未授权访问")
		return
	}

	uid, ok := userID.(int64)
	if !ok {
		ResponseInternalServerError(c, "用户ID格式错误")
		return
	}

	// 4. 提交申请
	if err := h.organizationUsecase.RequestJoinOrganization(
		c.Request.Context(),
		uid,
		orgID,
		req.Message,
	); err != nil {
		switch err {
		case domain.ErrOrganizationNotFound:
			ResponseNotFound(c, "组织不存在")
		case domain.ErrPermissionDenied:
			ResponseForbidden(c, "该组织不允许申请加入")
		case domain.ErrUserAlreadyExist:
			ResponseConflict(c, "您已是组织成员")
		case domain.ErrBadParamInput:
			ResponseBadRequest(c, "您已有待处理的申请")
		default:
			ResponseInternalServerError(c, "提交申请失败")
		}
		return
	}

	ResponseCreated(c, "申请提交成功", map[string]interface{}{
		"message": "加入申请已提交，请等待管理员审核",
		"success": true,
	})
}

// ProcessJoinRequest 处理加入申请
// @Summary 处理加入申请
// @Description 审批组织加入申请
// @Tags organizations
// @Accept json
// @Produce json
// @Param requestId path int true "申请ID"
// @Param process body dto.ProcessJoinRequestRequest true "处理信息"
// @Success 200 {object} SuccessResponse "处理成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 403 {object} ErrorResponse "权限不足"
// @Failure 404 {object} ErrorResponse "申请不存在"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Security BearerAuth
// @Router /api/v1/organizations/join-requests/{requestId}/process [post]
func (h *OrganizationHandler) ProcessJoinRequest(c *gin.Context) {
	var req dto.ProcessJoinRequestRequest

	// 1. 获取申请ID
	requestID, err := strconv.ParseInt(c.Param("requestId"), 10, 64)
	if err != nil {
		ResponseBadRequest(c, "申请ID格式错误")
		return
	}

	// 2. 绑定和验证请求数据
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseBadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 3. 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		ResponseUnauthorized(c, "未授权访问")
		return
	}

	uid, ok := userID.(int64)
	if !ok {
		ResponseInternalServerError(c, "用户ID格式错误")
		return
	}

	// 4. 处理申请
	if err := h.organizationUsecase.ProcessJoinRequest(
		c.Request.Context(),
		uid,
		requestID,
		req.Approve,
		req.Note,
	); err != nil {
		switch err {
		case domain.ErrNotFound:
			ResponseNotFound(c, "申请不存在")
		case domain.ErrPermissionDenied:
			ResponseForbidden(c, "权限不足")
		case domain.ErrBadParamInput:
			ResponseBadRequest(c, "申请状态无效")
		default:
			ResponseInternalServerError(c, "处理申请失败")
		}
		return
	}

	action := "拒绝"
	if req.Approve {
		action = "批准"
	}

	ResponseOK(c, "处理成功", map[string]interface{}{
		"message": "申请已" + action,
		"success": true,
	})
}

// GetOrganizationMembers 获取组织成员列表
// @Summary 获取组织成员列表
// @Description 获取指定组织的成员列表
// @Tags organizations
// @Accept json
// @Produce json
// @Param id path int true "组织ID"
// @Success 200 {array} dto.OrganizationMemberResponse "获取成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 403 {object} ErrorResponse "权限不足"
// @Failure 404 {object} ErrorResponse "组织不存在"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Security BearerAuth
// @Router /api/v1/organizations/{id}/members [get]
func (h *OrganizationHandler) GetOrganizationMembers(c *gin.Context) {
	// 1. 获取组织ID
	orgID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		ResponseBadRequest(c, "组织ID格式错误")
		return
	}

	// 2. 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		ResponseUnauthorized(c, "未授权访问")
		return
	}

	uid, ok := userID.(int64)
	if !ok {
		ResponseInternalServerError(c, "用户ID格式错误")
		return
	}

	// 3. 获取成员列表
	members, err := h.organizationUsecase.GetOrganizationMembers(c.Request.Context(), uid, orgID)
	if err != nil {
		switch err {
		case domain.ErrOrganizationNotFound:
			ResponseNotFound(c, "组织不存在")
		case domain.ErrPermissionDenied:
			ResponseForbidden(c, "权限不足")
		default:
			ResponseInternalServerError(c, "获取成员列表失败")
		}
		return
	}

	// 4. 转换为响应格式
	response := make([]*dto.OrganizationMemberResponse, len(members))
	for i, member := range members {
		response[i] = dto.ToOrganizationMemberResponse(member)
	}

	ResponseOK(c, "获取成功", response)
}

// UpdateMemberRole 更新成员角色
// @Summary 更新成员角色
// @Description 更新组织成员的角色
// @Tags organizations
// @Accept json
// @Produce json
// @Param id path int true "组织ID"
// @Param memberId path int true "成员ID"
// @Param role body dto.UpdateMemberRoleRequest true "角色信息"
// @Success 200 {object} SuccessResponse "更新成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 403 {object} ErrorResponse "权限不足"
// @Failure 404 {object} ErrorResponse "组织或成员不存在"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Security BearerAuth
// @Router /api/v1/organizations/{id}/members/{memberId}/role [put]
func (h *OrganizationHandler) UpdateMemberRole(c *gin.Context) {
	var req dto.UpdateMemberRoleRequest

	// 1. 获取路径参数
	orgID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		ResponseBadRequest(c, "组织ID格式错误")
		return
	}

	memberID, err := strconv.ParseInt(c.Param("memberId"), 10, 64)
	if err != nil {
		ResponseBadRequest(c, "成员ID格式错误")
		return
	}

	// 2. 绑定和验证请求数据
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseBadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 3. 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		ResponseUnauthorized(c, "未授权访问")
		return
	}

	uid, ok := userID.(int64)
	if !ok {
		ResponseInternalServerError(c, "用户ID格式错误")
		return
	}

	// 4. 更新成员角色
	if err := h.organizationUsecase.UpdateMemberRole(
		c.Request.Context(),
		uid,
		orgID,
		memberID,
		req.Role,
	); err != nil {
		switch err {
		case domain.ErrOrganizationNotFound:
			ResponseNotFound(c, "组织不存在")
		case domain.ErrUserNotFound:
			ResponseNotFound(c, "成员不存在")
		case domain.ErrPermissionDenied:
			ResponseForbidden(c, "权限不足")
		default:
			ResponseInternalServerError(c, "更新角色失败")
		}
		return
	}

	ResponseOK(c, "更新成功", map[string]interface{}{
		"message": "成员角色更新成功",
		"success": true,
	})
}

// RemoveMember 移除成员
// @Summary 移除成员
// @Description 从组织中移除成员
// @Tags organizations
// @Accept json
// @Produce json
// @Param id path int true "组织ID"
// @Param memberId path int true "成员ID"
// @Success 200 {object} SuccessResponse "移除成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 403 {object} ErrorResponse "权限不足"
// @Failure 404 {object} ErrorResponse "组织或成员不存在"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Security BearerAuth
// @Router /api/v1/organizations/{id}/members/{memberId} [delete]
func (h *OrganizationHandler) RemoveMember(c *gin.Context) {
	// 1. 获取路径参数
	orgID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		ResponseBadRequest(c, "组织ID格式错误")
		return
	}

	memberID, err := strconv.ParseInt(c.Param("memberId"), 10, 64)
	if err != nil {
		ResponseBadRequest(c, "成员ID格式错误")
		return
	}

	// 2. 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		ResponseUnauthorized(c, "未授权访问")
		return
	}

	uid, ok := userID.(int64)
	if !ok {
		ResponseInternalServerError(c, "用户ID格式错误")
		return
	}

	// 3. 移除成员
	if err := h.organizationUsecase.RemoveMember(c.Request.Context(), uid, orgID, memberID); err != nil {
		switch err {
		case domain.ErrOrganizationNotFound:
			ResponseNotFound(c, "组织不存在")
		case domain.ErrUserNotFound:
			ResponseNotFound(c, "成员不存在")
		case domain.ErrPermissionDenied:
			ResponseForbidden(c, "权限不足")
		default:
			ResponseInternalServerError(c, "移除成员失败")
		}
		return
	}

	ResponseOK(c, "移除成功", map[string]interface{}{
		"message": "成员移除成功",
		"success": true,
	})
}

// LeaveOrganization 退出组织
// @Summary 退出组织
// @Description 用户主动退出组织
// @Tags organizations
// @Accept json
// @Produce json
// @Param id path int true "组织ID"
// @Success 200 {object} SuccessResponse "退出成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 403 {object} ErrorResponse "权限不足"
// @Failure 404 {object} ErrorResponse "组织不存在"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Security BearerAuth
// @Router /api/v1/organizations/{id}/leave [delete]
func (h *OrganizationHandler) LeaveOrganization(c *gin.Context) {
	// 1. 获取组织ID
	orgID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		ResponseBadRequest(c, "组织ID格式错误")
		return
	}

	// 2. 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		ResponseUnauthorized(c, "未授权访问")
		return
	}

	uid, ok := userID.(int64)
	if !ok {
		ResponseInternalServerError(c, "用户ID格式错误")
		return
	}

	// 3. 退出组织
	if err := h.organizationUsecase.LeaveOrganization(c.Request.Context(), uid, orgID); err != nil {
		switch err {
		case domain.ErrOrganizationNotFound:
			ResponseNotFound(c, "组织不存在")
		case domain.ErrUserNotFound:
			ResponseNotFound(c, "您不是该组织成员")
		case domain.ErrPermissionDenied:
			ResponseForbidden(c, "组织所有者不能直接退出，请先转移所有权")
		default:
			ResponseInternalServerError(c, "退出组织失败")
		}
		return
	}

	ResponseOK(c, "退出成功", map[string]interface{}{
		"message": "已成功退出组织",
		"success": true,
	})
}
