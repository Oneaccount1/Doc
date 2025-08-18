package rest

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"DOC/domain"
	"DOC/internal/rest/dto"
)

// SpaceHandler 空间处理器
type SpaceHandler struct {
	spaceUsecase domain.SpaceUsecase
}

// NewSpaceHandler 创建新的空间处理器
func NewSpaceHandler(spaceUsecase domain.SpaceUsecase) *SpaceHandler {
	return &SpaceHandler{
		spaceUsecase: spaceUsecase,
	}
}

// handleSpaceError 处理空间相关错误
func (h *SpaceHandler) handleSpaceError(c *gin.Context, err error) {
	switch err {
	case domain.ErrSpaceNotFound:
		ResponseNotFound(c, "空间不存在")
	case domain.ErrSpaceAlreadyExist:
		ResponseConflict(c, "空间已存在")
	case domain.ErrSpacePermissionDenied:
		ResponseForbidden(c, "权限不足")
	case domain.ErrNotSpaceMember:
		ResponseForbidden(c, "不是空间成员")
	case domain.ErrUserNotFound:
		ResponseNotFound(c, "用户不存在")
	case domain.ErrUserNotActive:
		ResponseBadRequest(c, "用户未激活")
	case domain.ErrInvalidSpaceName:
		ResponseBadRequest(c, "无效的空间名称")
	case domain.ErrInvalidSpaceType:
		ResponseBadRequest(c, "无效的空间类型")
	case domain.ErrInvalidRole:
		ResponseBadRequest(c, "无效的角色")
	case domain.ErrConflict:
		ResponseConflict(c, "资源冲突")
	case domain.ErrPermissionDenied:
		ResponseForbidden(c, "权限不足")
	case domain.ErrDocumentNotFound:
		ResponseNotFound(c, "文档不存在")
	case domain.ErrBadParamInput:
		ResponseBadRequest(c, "参数错误")
	default:
		ResponseInternalServerError(c, "服务器内部错误")
	}
}

// Create 创建空间
// @Summary 创建新空间
// @Description 创建一个新的空间
// @Tags spaces
// @Accept json
// @Produce json
// @Param request body dto.CreateSpaceDto true "创建空间请求"
// @Success 201 {object} dto.SpaceResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security Bearer
// @Router /api/v1/spaces [post]
func (h *SpaceHandler) Create(c *gin.Context) {
	var req dto.CreateSpaceDto
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseBadRequest(c, "请求参数错误")
		return
	}

	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		ResponseUnauthorized(c, "用户未认证")
		return
	}

	uid, ok := userID.(int64)
	if !ok {
		ResponseUnauthorized(c, "无效的用户ID")
		return
	}

	// 设置默认值
	spaceType := req.Type
	if spaceType == "" {
		spaceType = domain.SpaceTypeWorkspace
	}
	// 组装参数
	createSpacePara := domain.CreateSpacePara{
		UserID:    uid,
		Name:      req.Name,
		Desc:      req.Description,
		Icon:      req.Icon,
		Color:     req.Color,
		SpaceType: spaceType,
		IsPublic:  req.IsPublic,
		OrgID:     req.OrganizationID,
	}

	// 调用业务逻辑
	space, err := h.spaceUsecase.CreateSpace(
		c.Request.Context(),
		createSpacePara,
	)
	if err != nil {
		h.handleSpaceError(c, err)
		return
	}

	// 返回响应
	ResponseCreated(c, "空间创建成功", dto.ToSpaceResponse(space))
}

// FindAll 获取用户的空间列表
// @Summary 获取用户的空间列表
// @Description 获取当前用户的所有空间
// @Tags spaces
// @Accept json
// @Produce json
// @Success 200 {object} dto.SpaceListResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security Bearer
// @Router /api/v1/spaces [get]
func (h *SpaceHandler) FindAll(c *gin.Context) {
	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		ResponseUnauthorized(c, "用户未认证")
		return
	}

	uid, ok := userID.(int64)
	if !ok {
		ResponseUnauthorized(c, "无效的用户ID")
		return
	}

	// 调用业务逻辑
	spaces, err := h.spaceUsecase.GetMySpaces(c.Request.Context(), uid)
	if err != nil {
		h.handleSpaceError(c, err)
		return
	}

	// 返回响应
	response := dto.ToSpaceListResponse(spaces, int64(len(spaces)), 1, len(spaces))
	ResponseOK(c, "获取成功", response)
}

// FindOne 获取空间详情
// @Summary 获取空间详情
// @Description 根据ID获取空间的详细信息
// @Tags spaces
// @Accept json
// @Produce json
// @Param id path int true "空间ID"
// @Success 200 {object} dto.SpaceResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security Bearer
// @Router /api/v1/spaces/{id} [get]
func (h *SpaceHandler) FindOne(c *gin.Context) {
	// 获取空间ID
	spaceID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		ResponseBadRequest(c, "无效的空间ID")
		return
	}

	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		ResponseUnauthorized(c, "用户未认证")
		return
	}

	uid, ok := userID.(int64)
	if !ok {
		ResponseUnauthorized(c, "无效的用户ID")
		return
	}

	// 调用业务逻辑
	space, err := h.spaceUsecase.GetSpace(c.Request.Context(), uid, spaceID)
	if err != nil {
		h.handleSpaceError(c, err)
		return
	}

	// 返回响应
	ResponseOK(c, "获取成功", dto.ToSpaceResponse(space))
}

// Update 更新空间信息
// @Summary 更新空间信息
// @Description 更新空间的基本信息
// @Tags spaces
// @Accept json
// @Produce json
// @Param id path int true "空间ID"
// @Param request body dto.UpdateSpaceDto true "更新空间请求"
// @Success 200 {object} dto.SpaceResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security Bearer
// @Router /api/v1/spaces/{id} [put]
func (h *SpaceHandler) Update(c *gin.Context) {
	// 获取空间ID
	spaceID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		ResponseBadRequest(c, "无效的空间ID")
		return
	}

	var req dto.UpdateSpaceDto
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseBadRequest(c, "请求参数错误")
		return
	}

	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		ResponseUnauthorized(c, "用户未认证")
		return
	}

	uid, ok := userID.(int64)
	if !ok {
		ResponseUnauthorized(c, "无效的用户ID")
		return
	}
	// 组装参数
	updateSpacePara := domain.UpdateSpacePara{
		UserID:    uid,
		SpaceID:   spaceID,
		Name:      req.Name,
		Desc:      req.Description,
		Icon:      req.Icon,
		Color:     req.Color,
		SpaceType: req.Type,
		IsPublic:  req.IsPublic,
	}

	// 调用业务逻辑
	space, err := h.spaceUsecase.UpdateSpace(
		c.Request.Context(),
		updateSpacePara,
	)
	if err != nil {
		h.handleSpaceError(c, err)
		return
	}

	// 返回响应
	ResponseOK(c, "更新成功", dto.ToSpaceResponse(space))
}

// Remove 删除空间
// @Summary 删除空间
// @Description 删除指定的空间
// @Tags spaces
// @Accept json
// @Produce json
// @Param id path int true "空间ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security Bearer
// @Router /api/v1/spaces/{id} [delete]
func (h *SpaceHandler) Remove(c *gin.Context) {
	// 获取空间ID
	spaceID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		ResponseBadRequest(c, "无效的空间ID")
		return
	}

	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		ResponseUnauthorized(c, "用户未认证")
		return
	}

	uid, ok := userID.(int64)
	if !ok {
		ResponseUnauthorized(c, "无效的用户ID")
		return
	}

	// 调用业务逻辑
	err = h.spaceUsecase.DeleteSpace(c.Request.Context(), uid, spaceID)
	if err != nil {
		h.handleSpaceError(c, err)
		return
	}

	// 返回响应
	ResponseOK(c, "删除成功", nil)
}

// GetDocuments 获取空间中的文档
// @Summary 获取空间中的文档
// @Description 获取指定空间中的所有文档
// @Tags spaces
// @Accept json
// @Produce json
// @Param id path int true "空间ID"
// @Success 200 {object} dto.SpaceDocumentListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security Bearer
// @Router /api/v1/spaces/{id}/documents [get]
func (h *SpaceHandler) GetDocuments(c *gin.Context) {
	// 获取空间ID
	spaceID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		ResponseBadRequest(c, "无效的空间ID")
		return
	}

	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		ResponseUnauthorized(c, "用户未认证")
		return
	}

	uid, ok := userID.(int64)
	if !ok {
		ResponseUnauthorized(c, "无效的用户ID")
		return
	}

	// 调用业务逻辑
	documents, err := h.spaceUsecase.GetSpaceDocuments(c.Request.Context(), uid, spaceID)
	if err != nil {
		h.handleSpaceError(c, err)
		return
	}

	// 构建响应
	documentResponses := make([]*dto.DocumentResponseDto, len(documents))
	for i, doc := range documents {
		documentResponses[i] = dto.FromDocument(doc)
	}

	response := &dto.SpaceDocumentListResponse{
		Documents: documentResponses,
		Total:     int64(len(documents)),
	}

	ResponseOK(c, "获取成功", response)
}

// AddDocument 添加文档到空间
// @Summary 添加文档到空间
// @Description 将指定文档添加到空间中
// @Tags spaces
// @Accept json
// @Produce json
// @Param id path int true "空间ID"
// @Param request body dto.AddDocumentDto true "添加文档请求"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security Bearer
// @Router /api/v1/spaces/{id}/documents [post]
func (h *SpaceHandler) AddDocument(c *gin.Context) {
	// 获取空间ID
	spaceID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		ResponseBadRequest(c, "无效的空间ID")
		return
	}

	var req dto.AddDocumentDto
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseBadRequest(c, "请求参数错误")
		return
	}

	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		ResponseUnauthorized(c, "用户未认证")
		return
	}

	uid, ok := userID.(int64)
	if !ok {
		ResponseUnauthorized(c, "无效的用户ID")
		return
	}

	// 调用业务逻辑
	err = h.spaceUsecase.AddDocumentToSpace(c.Request.Context(), uid, spaceID, req.DocumentID)
	if err != nil {
		h.handleSpaceError(c, err)
		return
	}

	// 返回响应
	ResponseOK(c, "添加成功", nil)
}

// RemoveDocument 从空间移除文档
// @Summary 从空间移除文档
// @Description 将指定文档从空间中移除
// @Tags spaces
// @Accept json
// @Produce json
// @Param id path int true "空间ID"
// @Param documentId path int true "文档ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security Bearer
// @Router /api/v1/spaces/{id}/documents/{documentId} [delete]
func (h *SpaceHandler) RemoveDocument(c *gin.Context) {
	// 获取空间ID
	spaceID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		ResponseBadRequest(c, "无效的空间ID")
		return
	}

	// 获取文档ID
	documentID, err := strconv.ParseInt(c.Param("documentId"), 10, 64)
	if err != nil {
		ResponseBadRequest(c, "无效的文档ID")
		return
	}

	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		ResponseUnauthorized(c, "用户未认证")
		return
	}

	uid, ok := userID.(int64)
	if !ok {
		ResponseUnauthorized(c, "无效的用户ID")
		return
	}

	// 调用业务逻辑
	err = h.spaceUsecase.RemoveDocumentFromSpace(c.Request.Context(), uid, spaceID, documentID)
	if err != nil {
		h.handleSpaceError(c, err)
		return
	}

	// 返回响应
	ResponseOK(c, "移除成功", nil)
}

// GetMembers 获取空间成员列表
// @Summary 获取空间成员列表
// @Description 获取指定空间的所有成员
// @Tags spaces
// @Accept json
// @Produce json
// @Param id path int true "空间ID"
// @Success 200 {object} dto.SpaceMemberListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security Bearer
// @Router /api/v1/spaces/{id}/members [get]
func (h *SpaceHandler) GetMembers(c *gin.Context) {
	// 获取空间ID
	spaceID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		ResponseBadRequest(c, "无效的空间ID")
		return
	}

	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		ResponseUnauthorized(c, "用户未认证")
		return
	}

	uid, ok := userID.(int64)
	if !ok {
		ResponseUnauthorized(c, "无效的用户ID")
		return
	}

	// 调用业务逻辑
	members, err := h.spaceUsecase.GetSpaceMembers(c.Request.Context(), uid, spaceID)
	if err != nil {
		h.handleSpaceError(c, err)
		return
	}

	// 返回响应
	response := dto.ToSpaceMemberListResponse(members)
	ResponseOK(c, "获取成功", response)
}

// AddMember 添加空间成员
// @Summary 添加空间成员
// @Description 向空间添加新成员
// @Tags spaces
// @Accept json
// @Produce json
// @Param id path int true "空间ID"
// @Param request body dto.AddMemberDto true "添加成员请求"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security Bearer
// @Router /api/v1/spaces/{id}/members [post]
func (h *SpaceHandler) AddMember(c *gin.Context) {
	// 获取空间ID
	spaceID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		ResponseBadRequest(c, "无效的空间ID")
		return
	}

	var req dto.AddMemberDto
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseBadRequest(c, "请求参数错误")
		return
	}

	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		ResponseUnauthorized(c, "用户未认证")
		return
	}

	uid, ok := userID.(int64)
	if !ok {
		ResponseUnauthorized(c, "无效的用户ID")
		return
	}

	// 调用业务逻辑
	err = h.spaceUsecase.AddSpaceMember(c.Request.Context(), uid, spaceID, req.UserID, req.Role)
	if err != nil {
		h.handleSpaceError(c, err)
		return
	}

	// 返回响应
	ResponseOK(c, "添加成功", nil)
}

// UpdateMemberRole 更新成员角色
// @Summary 更新成员角色
// @Description 更新空间成员的角色
// @Tags spaces
// @Accept json
// @Produce json
// @Param id path int true "空间ID"
// @Param userId path int true "用户ID"
// @Param request body dto.UpdateMemberRoleDto true "更新角色请求"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security Bearer
// @Router /api/v1/spaces/{id}/members/{userId} [put]
func (h *SpaceHandler) UpdateMemberRole(c *gin.Context) {
	// 获取空间ID
	spaceID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		ResponseBadRequest(c, "无效的空间ID")
		return
	}

	// 获取目标用户ID
	targetUserID, err := strconv.ParseInt(c.Param("userId"), 10, 64)
	if err != nil {
		ResponseBadRequest(c, "无效的用户ID")
		return
	}

	var req dto.UpdateMemberRoleDto
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseBadRequest(c, "请求参数错误")
		return
	}

	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		ResponseUnauthorized(c, "用户未认证")
		return
	}

	uid, ok := userID.(int64)
	if !ok {
		ResponseUnauthorized(c, "无效的用户ID")
		return
	}

	// 调用业务逻辑
	err = h.spaceUsecase.UpdateMemberRole(c.Request.Context(), uid, spaceID, targetUserID, req.Role)
	if err != nil {
		h.handleSpaceError(c, err)
		return
	}

	// 返回响应
	ResponseOK(c, "更新成功", nil)
}

// RemoveMember 移除空间成员
// @Summary 移除空间成员
// @Description 从空间中移除指定成员
// @Tags spaces
// @Accept json
// @Produce json
// @Param id path int true "空间ID"
// @Param userId path int true "用户ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security Bearer
// @Router /api/v1/spaces/{id}/members/{userId} [delete]
func (h *SpaceHandler) RemoveMember(c *gin.Context) {
	// 获取空间ID
	spaceID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		ResponseBadRequest(c, "无效的空间ID")
		return
	}

	// 获取目标用户ID
	targetUserID, err := strconv.ParseInt(c.Param("userId"), 10, 64)
	if err != nil {
		ResponseBadRequest(c, "无效的用户ID")
		return
	}

	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		ResponseUnauthorized(c, "用户未认证")
		return
	}

	uid, ok := userID.(int64)
	if !ok {
		ResponseUnauthorized(c, "无效的用户ID")
		return
	}

	// 调用业务逻辑
	err = h.spaceUsecase.RemoveSpaceMember(c.Request.Context(), uid, spaceID, targetUserID)
	if err != nil {
		h.handleSpaceError(c, err)
		return
	}

	// 返回响应
	ResponseOK(c, "移除成功", nil)
}
