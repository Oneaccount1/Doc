package rest

import (
	"DOC/internal/rest/middleware"
	"encoding/json"
	"errors"

	"github.com/gin-gonic/gin"

	"DOC/domain"
	"DOC/internal/rest/dto"
)

// DocumentHandler 文档HTTP处理器
// 负责处理文档相关的HTTP请求，调用聚合服务完成业务操作
type DocumentHandler struct {
	aggregateService domain.DocumentAggregateUsecase // 文档聚合服务
}

// NewDocumentHandler 创建新的文档处理器实例
func NewDocumentHandler(aggregateService domain.DocumentAggregateUsecase) *DocumentHandler {
	return &DocumentHandler{
		aggregateService: aggregateService,
	}
}

// === 文档核心操作处理器 ===

// CreateDocument 创建文档
// POST /api/v1/documents
func (h *DocumentHandler) CreateDocument(c *gin.Context) {
	// 1. 获取当前用户ID（从中间件中获取）
	userID, exists := c.Get("user_id")
	if !exists {
		ResponseUnauthorized(c, "未授权访问")
		//c.JSON(http.StatusUnauthorized, dto.ErrorResponse("用户未登录", "UNAUTHORIZED"))
		return
	}

	// 2. 绑定请求参数
	var req dto.CreateDocumentDto
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseBadRequest(c, "请求参数无效")
		//c.JSON(http.StatusBadRequest, dto.ErrorResponse("请求参数无效: "+err.Error(), "INVALID_PARAMS"))
		return
	}

	// 3. 调用业务服务创建文档
	document, err := h.aggregateService.CreateDocument(
		c.Request.Context(),
		userID.(int64),
		req.Title,
		req.GetContentString(),
		req.ToDocumentType(),
		req.ParentID,
		req.SpaceID,
		req.SortOrder,
		req.IsStarred,
	)

	// 4. 处理业务错误
	if err != nil {
		h.handleBusinessError(c, err)
		return
	}

	// 5. 返回成功响应
	ResponseCreated(c, "Created", document)
}

// GetDocument 获取文档详情
// GET /api/v1/documents/:id
func (h *DocumentHandler) GetDocument(c *gin.Context) {
	// 1. 获取用户ID

	userID, exist := middleware.GetCurrentUserID(c)
	if userID == 0 || !exist {
		return
	}

	// 2. 获取文档ID参数
	var param dto.IDParamDto
	if err := c.ShouldBindUri(&param); err != nil {
		ResponseBadRequest(c, "无效的文档ID")
		//c.JSON(http.StatusBadRequest, dto.ErrorResponse("无效的文档ID", "INVALID_DOCUMENT_ID"))
		return
	}

	// 3. 调用业务服务获取文档
	document, err := h.aggregateService.GetDocument(c.Request.Context(), userID, param.ID)
	if err != nil {
		h.handleBusinessError(c, err)
		return
	}

	// 4. 返回文档信息
	ResponseOK(c, "Success", document)
}

// UpdateDocument 更新文档信息
// PUT /api/v1/documents/:id
func (h *DocumentHandler) UpdateDocument(c *gin.Context) {
	// 1. 获取用户ID和文档ID
	userID, exist := middleware.GetCurrentUserID(c)
	if userID == 0 || !exist {
		return
	}

	var param dto.IDParamDto
	if err := c.ShouldBindUri(&param); err != nil {
		ResponseBadRequest(c, "无效的文档ID")
		//c.JSON(http.StatusBadRequest, dto.ErrorResponse("无效的文档ID", "INVALID_DOCUMENT_ID"))
		return
	}

	// 2. 绑定更新参数
	var req dto.UpdateDocumentDto
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseBadRequest(c, "请求参数无效"+err.Error())
		//c.JSON(http.StatusBadRequest, dto.ErrorResponse("请求参数无效: "+err.Error(), "INVALID_PARAMS"))
		return
	}

	// 3. 调用业务服务更新文档
	document, err := h.aggregateService.UpdateDocument(
		c.Request.Context(),
		userID,
		param.ID,
		*req.Title, // 假设前端总是传递标题
		req.ToDocumentType(),
		req.ParentID,
		req.SortOrder,
		req.IsStarred,
	)

	if err != nil {
		h.handleBusinessError(c, err)
		return
	}

	// 4. 返回更新后的文档
	ResponseOK(c, "Success", document)
}

// DeleteDocument 删除文档
// DELETE /api/v1/documents/:id
func (h *DocumentHandler) DeleteDocument(c *gin.Context) {
	// 1. 获取用户ID和文档ID
	userID, exist := middleware.GetCurrentUserID(c)
	if userID == 0 || !exist {
		return
	}

	var param dto.IDParamDto
	if err := c.ShouldBindUri(&param); err != nil {
		ResponseBadRequest(c, "无效的文档ID")
		//c.JSON(http.StatusBadRequest, dto.ErrorResponse("无效的文档ID", "INVALID_DOCUMENT_ID"))
		return
	}

	// 2. 获取查询参数（是否永久删除）
	permanent := c.Query("permanent") == "true"

	// 3. 调用业务服务删除文档
	var err error
	if permanent {
		// 永久删除（暂未实现在领域服务中，这里作为扩展）
		err = h.aggregateService.DeleteDocument(c.Request.Context(), userID, param.ID)
	} else {
		// 软删除
		err = h.aggregateService.DeleteDocument(c.Request.Context(), userID, param.ID)
	}

	if err != nil {
		h.handleBusinessError(c, err)
		return
	}

	// 4. 返回成功响应
	ResponseOK(c, "Success", nil)
}

// === 文档内容操作处理器 ===

// GetDocumentContent 获取文档内容
// GET /api/v1/documents/:id/content
func (h *DocumentHandler) GetDocumentContent(c *gin.Context) {
	// 1. 获取用户ID和文档ID
	userID, exist := middleware.GetCurrentUserID(c)
	if userID == 0 || !exist {
		return
	}

	var param dto.IDParamDto
	if err := c.ShouldBindUri(&param); err != nil {
		ResponseBadRequest(c, "无效的文档ID")
		//c.JSON(http.StatusBadRequest, dto.ErrorResponse("无效的文档ID", "INVALID_DOCUMENT_ID"))
		return
	}

	// 2. 调用业务服务获取内容
	content, err := h.aggregateService.GetDocumentContent(c.Request.Context(), userID, param.ID)
	if err != nil {
		h.handleBusinessError(c, err)
		return
	}

	// 3. 解析JSON内容
	var jsonContent json.RawMessage
	if content != "" {
		jsonContent = json.RawMessage(content)
	} else {
		jsonContent = json.RawMessage("{}")
	}

	// 4. 返回内容
	ResponseOK(c, "Success", jsonContent)
}

// UpdateDocumentContent 更新文档内容
// PUT /api/v1/documents/:id/content
func (h *DocumentHandler) UpdateDocumentContent(c *gin.Context) {
	// 1. 获取用户ID和文档ID
	userID, exist := middleware.GetCurrentUserID(c)
	if userID == 0 || !exist {
		return
	}

	var param dto.IDParamDto
	if err := c.ShouldBindUri(&param); err != nil {
		ResponseBadRequest(c, "无效的文档ID")
		//c.JSON(http.StatusBadRequest, dto.ErrorResponse("无效的文档ID", "INVALID_DOCUMENT_ID"))
		return
	}

	// 2. 绑定内容参数
	var req dto.UpdateDocumentContentDto
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseBadRequest(c, "请求参数无效"+err.Error())
		//c.JSON(http.StatusBadRequest, dto.ErrorResponse("请求参数无效: "+err.Error(), "INVALID_PARAMS"))
		return
	}

	// 3. 调用业务服务更新内容
	err := h.aggregateService.UpdateDocumentContent(
		c.Request.Context(),
		userID,
		param.ID,
		req.GetContentString(),
	)

	if err != nil {
		h.handleBusinessError(c, err)
		return
	}

	// 4. 返回成功响应
	ResponseOK(c, "Success", nil)
}

// === 文档查询和搜索处理器 ===

// GetMyDocuments 获取我的文档列表
// GET /api/v1/documents
func (h *DocumentHandler) GetMyDocuments(c *gin.Context) {
	// 1. 获取用户ID
	userID, exist := middleware.GetCurrentUserID(c)
	if userID == 0 || !exist {
		return
	}

	// 2. 绑定查询参数
	var query dto.DocumentQueryDto
	if err := c.ShouldBindQuery(&query); err != nil {
		ResponseBadRequest(c, "查询参数无效"+err.Error())
		//c.JSON(http.StatusBadRequest, dto.ErrorResponse("查询参数无效: "+err.Error(), "INVALID_QUERY"))
		return
	}

	// 3. 调用业务服务获取文档列表
	documents, err := h.aggregateService.GetMyDocuments(
		c.Request.Context(),
		userID,
		query.ParentID,
		query.IncludeDeleted,
	)

	if err != nil {
		h.handleBusinessError(c, err)
		return
	}

	// 4. 转换为响应DTO
	responseDTOs := make([]*dto.DocumentResponseDto, len(documents))
	for i, doc := range documents {
		responseDTOs[i] = dto.FromDocument(doc)
	}

	// 5. 返回文档列表
	//  返回的数据应当有owned shared todo
	ResponseOK(c, "Success", responseDTOs)
}

// SearchDocuments 搜索文档
// GET /api/v1/documents/search
func (h *DocumentHandler) SearchDocuments(c *gin.Context) {
	// 1. 获取用户ID
	userID, exist := middleware.GetCurrentUserID(c)
	if userID == 0 || !exist {
		return
	}

	// 2. 绑定搜索参数
	var query dto.DocumentSearchQueryDto
	if err := c.ShouldBindQuery(&query); err != nil {
		ResponseBadRequest(c, "搜索参数无效"+err.Error())
		//c.JSON(http.StatusBadRequest, dto.ErrorResponse("搜索参数无效: "+err.Error(), "INVALID_SEARCH_QUERY"))
		return
	}

	// 3. 设置默认值
	if query.Limit <= 0 {
		query.Limit = 20
	}

	// 4. 调用业务服务搜索文档
	results, err := h.aggregateService.SearchDocuments(
		c.Request.Context(),
		userID,
		query.Keyword,
		query.ToDocumentType(),
		query.Limit,
		query.Offset,
	)

	if err != nil {
		h.handleBusinessError(c, err)
		return
	}

	// 5. 转换为响应DTO
	searchDTOs := make([]*dto.DocumentSearchItemDto, len(results))
	for i, result := range results {
		searchDTOs[i] = dto.FromDocumentSearchResult(result)
	}

	// 6. 返回搜索结果
	// todo 确认返回结构
	ResponseOK(c, "Success", searchDTOs)
	//c.JSON(http.StatusOK, dto.SuccessResponse(&dto.DocumentSearchResponseDto{
	//	Documents: searchDTOs,
	//	Total:     len(searchDTOs),
	//	Keyword:   query.Keyword,
	//}))
}

// === 文档分享操作处理器 ===

// CreateShareLink 创建分享链接
// POST /api/v1/documents/:id/share
func (h *DocumentHandler) CreateShareLink(c *gin.Context) {
	// 1. 获取用户ID和文档ID
	userID, exist := middleware.GetCurrentUserID(c)
	if userID == 0 || !exist {
		return
	}

	var param dto.IDParamDto
	if err := c.ShouldBindUri(&param); err != nil {
		ResponseBadRequest(c, "无效的文档ID")
		//c.JSON(http.StatusBadRequest, dto.ErrorResponse("无效的文档ID", "INVALID_DOCUMENT_ID"))
		return
	}

	// 2. 绑定分享参数
	var req dto.ShareDocumentDto
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseBadRequest(c, "分享参数无效"+err.Error())
		//c.JSON(http.StatusBadRequest, dto.ErrorResponse("分享参数无效: "+err.Error(), "INVALID_SHARE_PARAMS"))
		return
	}

	// 3. 解析过期时间
	expiresAt, err := req.GetExpiresAt()
	if err != nil {
		ResponseBadRequest(c, "过期时间格式无效")
		//c.JSON(http.StatusBadRequest, dto.ErrorResponse("过期时间格式无效", "INVALID_EXPIRES_AT"))
		return
	}

	// 4. 调用业务服务创建分享
	shareInfo, err := h.aggregateService.ShareDocument(
		c.Request.Context(),
		userID,
		param.ID,
		req.ToPermission(),
		*req.Password,
		expiresAt,
		req.ShareWithUserIDs,
	)

	if err != nil {
		h.handleBusinessError(c, err)
		return
	}

	// 5. 返回分享信息
	// todo c
	ResponseCreated(c, "Created", shareInfo)
	//c.JSON(http.StatusCreated, dto.SuccessResponseWithMessage(
	//	dto.FromDocumentShare(shareInfo),
	//	"分享链接创建成功",
	//))
}

// GetSharedDocument 通过分享链接获取文档
// GET /api/v1/documents/shared/:linkId
func (h *DocumentHandler) GetSharedDocument(c *gin.Context) {
	// 1. 获取分享链接参数
	var param dto.ShareLinkParamDto
	if err := c.ShouldBindUri(&param); err != nil {
		ResponseBadRequest(c, "无效的分享链接")
		//c.JSON(http.StatusBadRequest, dto.ErrorResponse("无效的分享链接", "INVALID_SHARE_LINK"))
		return
	}

	// 2. 获取密码参数
	password := c.Query("password")

	// 3. 获取访问IP
	accessIP := c.ClientIP()

	// 4. 调用业务服务获取分享文档
	accessInfo, err := h.aggregateService.GetSharedDocument(
		c.Request.Context(),
		param.LinkID,
		password,
		accessIP,
	)

	if err != nil {
		h.handleBusinessError(c, err)
		return
	}

	// 5. 返回文档访问信息
	ResponseOK(c, "Success", accessInfo)
	//c.JSON(http.StatusOK, dto.SuccessResponse(dto.FromDocumentAccessInfo(accessInfo)))
}

// === 文档收藏操作处理器 ===

// ToggleFavoriteDocument 切换文档收藏状态
// POST /api/v1/documents/:id/shared/favorite
func (h *DocumentHandler) ToggleFavoriteDocument(c *gin.Context) {
	// 1. 获取用户ID和文档ID
	userID, exist := middleware.GetCurrentUserID(c)
	if userID == 0 || !exist {
		return
	}

	var param dto.IDParamDto
	if err := c.ShouldBindUri(&param); err != nil {
		ResponseBadRequest(c, "无效的文档ID")
		//c.JSON(http.StatusBadRequest, dto.ErrorResponse("无效的文档ID", "INVALID_DOCUMENT_ID"))
		return
	}

	// 2. 调用业务服务切换收藏状态
	isFavorite, err := h.aggregateService.ToggleDocumentFavorite(c.Request.Context(), userID, param.ID)
	if err != nil {
		h.handleBusinessError(c, err)
		return
	}

	// 3. 返回收藏状态
	message := "已取消收藏"
	if isFavorite {
		message = "已添加到收藏"
	}

	ResponseOK(c, "Success", message)

	//c.JSON(http.StatusOK, dto.SuccessResponseWithMessage(
	//	map[string]bool{"is_favorite": isFavorite},
	//	message,
	//))
}

// === 文档权限操作处理器 ===

// SetUserPermission 设置用户权限
// PUT /api/v1/documents/:id/acl
func (h *DocumentHandler) SetUserPermission(c *gin.Context) {
	// 1. 获取用户ID和文档ID
	userID, exist := middleware.GetCurrentUserID(c)
	if userID == 0 || !exist {
		return
	}

	var param dto.IDParamDto
	if err := c.ShouldBindUri(&param); err != nil {
		ResponseBadRequest(c, "无效的文档ID")
		//c.JSON(http.StatusBadRequest, dto.ErrorResponse("无效的文档ID", "INVALID_DOCUMENT_ID"))
		return
	}

	// 2. 绑定权限参数
	var req struct {
		TargetUserID int64  `json:"target_user_id" binding:"required"`
		Permission   string `json:"permission" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseBadRequest(c, "权限参数无效"+err.Error())
		//c.JSON(http.StatusBadRequest, dto.ErrorResponse("权限参数无效: "+err.Error(), "INVALID_PERMISSION_PARAMS"))
		return
	}

	// 3. 调用业务服务设置权限
	err := h.aggregateService.GrantDocumentPermission(
		c.Request.Context(),
		userID,
		param.ID,
		req.TargetUserID,
		domain.Permission(req.Permission),
	)

	if err != nil {
		h.handleBusinessError(c, err)
		return
	}

	// 4. 返回成功响应
	ResponseOK(c, "Success", nil)
	//c.JSON(http.StatusOK, dto.SuccessResponseWithMessage(nil, "权限设置成功"))
}

// === 批量操作处理器 ===

// BatchDeleteDocuments 批量删除文档
// DELETE /api/v1/documents/batch
func (h *DocumentHandler) BatchDeleteDocuments(c *gin.Context) {
	// 1. 获取用户ID
	userID, exist := middleware.GetCurrentUserID(c)
	if userID == 0 || !exist {
		return
	}

	// 2. 绑定批量操作参数
	var req dto.BatchOperationDto
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseBadRequest(c, "批量操作参数无效"+err.Error())
		//c.JSON(http.StatusBadRequest, dto.ErrorResponse("批量操作参数无效: "+err.Error(), "INVALID_BATCH_PARAMS"))
		return
	}

	// 3. 调用业务服务批量删除
	err := h.aggregateService.BatchDeleteDocuments(c.Request.Context(), userID, req.DocumentIDs)
	if err != nil {
		h.handleBusinessError(c, err)
		return
	}

	// 4. 返回成功响应
	ResponseOK(c, "Success", nil)
	//c.JSON(http.StatusOK, dto.SuccessResponseWithMessage(
	//	&dto.BatchOperationResponseDto{
	//		Success:      true,
	//		ProcessedIDs: req.DocumentIDs,
	//		Message:      "批量删除成功",
	//	},
	//	"批量删除成功",
	//))
}

// BatchMoveDocuments 批量移动文档
// PUT /api/v1/documents/batch/move
func (h *DocumentHandler) BatchMoveDocuments(c *gin.Context) {
	// 1. 获取用户ID
	userID, exist := middleware.GetCurrentUserID(c)
	if userID == 0 || !exist {
		return
	}

	// 2. 绑定批量操作参数
	var req dto.BatchOperationDto
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseBadRequest(c, "批量操作参数无效"+err.Error())
		//c.JSON(http.StatusBadRequest, dto.ErrorResponse("批量操作参数无效: "+err.Error(), "INVALID_BATCH_PARAMS"))
		return
	}

	// 3. 调用业务服务批量移动
	err := h.aggregateService.BatchMoveDocuments(c.Request.Context(), userID, req.DocumentIDs, req.NewParentID)
	if err != nil {
		h.handleBusinessError(c, err)
		return
	}

	// 4. 返回成功响应
	ResponseOK(c, "Success", nil)
	//c.JSON(http.StatusOK, dto.SuccessResponseWithMessage(
	//	&dto.BatchOperationResponseDto{
	//		Success:      true,
	//		ProcessedIDs: req.DocumentIDs,
	//		Message:      "批量移动成功",
	//	},
	//	"批量移动成功",
	//))
}

// GetMySharedDocuments 获取我分享的文档列表
// GET /api/v1/documents/shared-via-link
func (h *DocumentHandler) GetMySharedDocuments(c *gin.Context) {
	// 1. 获取用户ID
	userID, exist := middleware.GetCurrentUserID(c)
	if userID == 0 || !exist {
		return
	}

	// 2. 调用业务服务获取分享列表
	shares, err := h.aggregateService.GetMySharedDocuments(c.Request.Context(), userID)
	if err != nil {
		h.handleBusinessError(c, err)
		return
	}

	// 3. 转换为响应DTO
	shareDTOs := make([]*dto.DocumentShareResponseDto, len(shares))
	for i, share := range shares {
		shareDTOs[i] = dto.FromDocumentShare(share)
	}

	// 4. 返回分享列表
	ResponseOK(c, "Success", shareDTOs)
	//c.JSON(http.StatusOK, dto.SuccessResponse(shareDTOs))
}

// SetFavoriteCustomTitle 设置收藏的自定义标题
// PUT /api/v1/documents/:id/shared/title
func (h *DocumentHandler) SetFavoriteCustomTitle(c *gin.Context) {
	// 1. 获取用户ID和文档ID
	userID, exist := middleware.GetCurrentUserID(c)
	if userID == 0 || !exist {
		return
	}

	var param dto.IDParamDto
	if err := c.ShouldBindUri(&param); err != nil {
		ResponseBadRequest(c, "无效的文档ID")
		//c.JSON(http.StatusBadRequest, dto.ErrorResponse("无效的文档ID", "INVALID_DOCUMENT_ID"))
		return
	}

	// 2. 绑定请求参数
	var req struct {
		CustomTitle string `json:"custom_title" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseBadRequest(c, "请求参数无效"+err.Error())
		//c.JSON(http.StatusBadRequest, dto.ErrorResponse("请求参数无效: "+err.Error(), "INVALID_PARAMS"))
		return
	}

	// 3. 调用业务服务设置自定义标题
	err := h.aggregateService.SetFavoriteCustomTitle(c.Request.Context(), userID, param.ID, req.CustomTitle)
	if err != nil {
		h.handleBusinessError(c, err)
		return
	}

	// 4. 返回成功响应
	ResponseOK(c, "Success", nil)
	//c.JSON(http.StatusOK, dto.SuccessResponseWithMessage(nil, "自定义标题设置成功"))
}

// RemoveFavoriteDocument 移除文档收藏
// DELETE /api/v1/documents/:id/shared
func (h *DocumentHandler) RemoveFavoriteDocument(c *gin.Context) {
	// 1. 获取用户ID和文档ID
	userID, exist := middleware.GetCurrentUserID(c)
	if userID == 0 || !exist {
		return
	}

	var param dto.IDParamDto
	if err := c.ShouldBindUri(&param); err != nil {
		ResponseBadRequest(c, "无效的文档ID")
		//c.JSON(http.StatusBadRequest, dto.ErrorResponse("无效的文档ID", "INVALID_DOCUMENT_ID"))
		return
	}

	// 2. 调用业务服务移除收藏
	err := h.aggregateService.RemoveDocumentFavorite(c.Request.Context(), userID, param.ID)
	if err != nil {
		h.handleBusinessError(c, err)
		return
	}

	// 3. 返回成功响应
	ResponseOK(c, "Success", nil)
	//c.JSON(http.StatusOK, dto.SuccessResponseWithMessage(nil, "已移除收藏"))
}

// CheckDocumentAccess 检查文档访问权限
// GET /api/v1/users/check-document-access/:documentId
func (h *DocumentHandler) CheckDocumentAccess(c *gin.Context) {
	// 1. 获取用户ID
	userID, exist := middleware.GetCurrentUserID(c)
	if userID == 0 || !exist {
		return
	}

	// 2. 获取文档ID参数
	var param dto.DocumentIDParamDto
	if err := c.ShouldBindUri(&param); err != nil {
		ResponseBadRequest(c, "无效的文档ID")
		//c.JSON(http.StatusBadRequest, dto.ErrorResponse("无效的文档ID", "INVALID_DOCUMENT_ID"))
		return
	}

	// 3. 调用业务服务检查访问权限
	accessInfo, err := h.aggregateService.GetDocumentWithAccessInfo(c.Request.Context(), userID, param.DocumentID)
	if err != nil {
		h.handleBusinessError(c, err)
		return
	}

	// 4. 返回访问权限信息
	ResponseOK(c, "Success", accessInfo)
	//c.JSON(http.StatusOK, dto.SuccessResponse(dto.FromDocumentAccessInfo(accessInfo)))
}

// handleBusinessError 处理业务错误，将领域错误转换为HTTP响应
func (h *DocumentHandler) handleBusinessError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrDocumentNotFound):
		ResponseNotFound(c, "权限不足")
	case errors.Is(err, domain.ErrPermissionDenied):
		ResponseUnauthorized(c, "权限不足")
		//c.JSON(http.StatusForbidden, dto.ErrorResponse("权限不足", "PERMISSION_DENIED"))
	case errors.Is(err, domain.ErrUserNotFound):
		ResponseNotFound(c, "用户不存在")
		//c.JSON(http.StatusNotFound, dto.ErrorResponse("用户不存在", "USER_NOT_FOUND"))
	case errors.Is(err, domain.ErrInvalidDocumentTitle):
		ResponseBadRequest(c, "文档标题无效")
		//c.JSON(http.StatusBadRequest, dto.ErrorResponse("文档标题无效", "INVALID_TITLE"))
	case errors.Is(err, domain.ErrInvalidDocumentType):
		ResponseBadRequest(c, "文档类型无效")
		//c.JSON(http.StatusBadRequest, dto.ErrorResponse("文档类型无效", "INVALID_TYPE"))
	case errors.Is(err, domain.ErrConflict):
		ResponseConflict(c, "操作冲突")
		//c.JSON(http.StatusConflict, dto.ErrorResponse("操作冲突", "CONFLICT"))
	case errors.Is(err, domain.ErrInvalidBatchRequest):
		ResponseBadRequest(c, "批量操作参数无效")
		//c.JSON(http.StatusBadRequest, dto.ErrorResponse("批量操作参数无效", "INVALID_BATCH"))
	case errors.Is(err, domain.ErrBatchSizeExceeded):
		ResponseBadRequest(c, "批量操作数量超限")
		//c.JSON(http.StatusBadRequest, dto.ErrorResponse("批量操作数量超限", "BATCH_SIZE_EXCEEDED"))
	default:
		// 记录未知错误（在实际项目中应该使用日志库）
		ResponseInternalServerError(c, "服务器内部错误")
		//c.JSON(http.StatusInternalServerError, dto.ErrorResponse("服务器内部错误", "INTERNAL_ERROR"))
	}
}
