package rest

import (
	"DOC/internal/rest/dto"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"DOC/domain"
)

// UserHandler 用户 HTTP 处理器
// 负责处理用户相关的 HTTP 请求，专注于用户资料管理、搜索等功能
// 认证相关功能已移至 AuthHandler
// 这一层负责 HTTP 协议的处理，数据格式转换，以及调用业务逻辑层
type UserHandler struct {
	userUsecase domain.UserUsecase // 用户业务逻辑接口
}

// NewUserHandler 创建新的用户处理器
// 参数说明：
// - userUsecase: 用户业务逻辑接口
// 返回值：
// - *UserHandler: 用户处理器实例
func NewUserHandler(userUsecase domain.UserUsecase) *UserHandler {
	return &UserHandler{
		userUsecase: userUsecase,
	}
}

// === HTTP 处理器方法 ===
// 注意：请求和响应结构体已移动到 dto.go 文件中，避免重复定义

// GetProfile 获取用户资料处理器
// @Summary 获取用户资料
// @Description 获取当前登录用户的详细信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} UserResponse "获取成功"
// @Failure 401 {object} ErrorResponse "未授权"
// @Failure 404 {object} ErrorResponse "用户不存在"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/users/profile [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	// 1. 从上下文中获取当前用户ID
	// 这个ID是在认证中间件中设置的
	userID, exists := c.Get("user_id")
	if !exists {
		ResponseUnauthorized(c, "未授权访问")
		return
	}

	// 2. 类型断言，确保 userID 是 int64 类型
	uid, ok := userID.(int64)
	if !ok {
		ResponseInternalServerError(c, "用户ID格式错误")
		return
	}

	// 3. 调用业务逻辑层获取用户信息
	user, err := h.userUsecase.GetProfile(c.Request.Context(), uid)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			ResponseNotFound(c, "用户不存在")
		} else {
			ResponseInternalServerError(c, "获取用户信息失败")
		}
		return
	}

	// 4. 返回用户信息
	ResponseOK(c, "获取成功", dto.ToUserResponse(user))
}

// UpdateProfile 更新用户资料处理器
// @Summary 更新用户资料
// @Description 更新当前登录用户的个人信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param profile body UpdateProfileRequest true "用户资料"
// @Success 200 {object} UserResponse "更新成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "未授权"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/users/profile [put]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
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

	// 2. 绑定请求数据
	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseBadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 3. 创建用户更新结构
	updates := &domain.UserProfileUpdate{
		Name:      &req.Name,
		AvatarURL: &req.AvatarURL,
		Bio:       &req.Bio,
	}

	// 4. 调用业务逻辑层更新用户信息
	if err := h.userUsecase.UpdateProfile(c.Request.Context(), uid, updates); err != nil {
		ResponseInternalServerError(c, "更新用户信息失败")
		return
	}

	// 5. 获取更新后的用户信息并返回
	updatedUser, err := h.userUsecase.GetProfile(c.Request.Context(), uid)
	if err != nil {
		ResponseInternalServerError(c, "获取更新后的用户信息失败")
		return
	}

	ResponseOK(c, "更新成功", dto.ToUserResponse(updatedUser))
}

// ChangePassword 修改密码处理器
// @Summary 修改密码
// @Description 修改当前登录用户的密码
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param password body ChangePasswordRequest true "密码信息"
// @Success 200 {object} SuccessResponse "修改成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "未授权"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/users/password [put]
func (h *UserHandler) ChangePassword(c *gin.Context) {
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

	// 2. 绑定请求数据
	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseBadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 3. 调用业务逻辑层修改密码
	if err := h.userUsecase.ChangePassword(c.Request.Context(), uid, req.OldPassword, req.NewPassword); err != nil {
		if err.Error() == "原密码错误" {
			ResponseBadRequest(c, "原密码错误")
		} else if err.Error() == "密码强度不足" {
			ResponseBadRequest(c, err.Error())
		} else {
			ResponseInternalServerError(c, "修改密码失败")
		}
		return
	}

	ResponseOK(c, "密码修改成功", nil)
}

// GetUserByID 根据ID获取用户处理器
// @Summary 根据ID获取用户
// @Description 根据用户ID获取用户基本信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Success 200 {object} UserResponse "获取成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "未授权"
// @Failure 404 {object} ErrorResponse "用户不存在"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/users/{id} [get]
func (h *UserHandler) GetUserByID(c *gin.Context) {
	var req dto.GetUserByIdRequest

	// 1. 绑定URI参数
	if err := c.ShouldBindUri(&req); err != nil {
		ResponseBadRequest(c, "用户ID格式错误: "+err.Error())
		return
	}

	// 2. 调用业务逻辑层获取用户信息
	user, err := h.userUsecase.GetUserByID(c.Request.Context(), req.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			ResponseNotFound(c, "用户不存在")
		} else {
			ResponseInternalServerError(c, "获取用户信息失败")
		}
		return
	}

	// 3. 返回用户信息
	ResponseOK(c, "获取成功", dto.ToUserResponse(user))
}

// GetUserByEmail 根据邮箱获取用户处理器
// @Summary 根据邮箱获取用户
// @Description 根据邮箱地址获取用户基本信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param email path string true "用户邮箱"
// @Success 200 {object} UserResponse "获取成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "未授权"
// @Failure 404 {object} ErrorResponse "用户不存在"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/users/email/{email} [get]
func (h *UserHandler) GetUserByEmail(c *gin.Context) {
	var req dto.GetUserByEmailRequest

	// 1. 绑定URI参数
	if err := c.ShouldBindUri(&req); err != nil {
		ResponseBadRequest(c, "邮箱格式错误: "+err.Error())
		return
	}

	// 2. 调用业务逻辑层获取用户信息
	user, err := h.userUsecase.GetUserByEmail(c.Request.Context(), req.Email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			ResponseNotFound(c, "用户不存在")
		} else {
			ResponseInternalServerError(c, "获取用户信息失败")
		}
		return
	}

	// 3. 返回用户信息
	ResponseOK(c, "获取成功", dto.ToUserResponse(user))
}

func (h *UserHandler) GetUserByGitHubId(c *gin.Context) {
	query := c.Query("githubId")
	if query == "" {
		ResponseBadRequest(c, "githubId不能为空")
		return
	}
	user, err := h.userUsecase.GetUserByGitHubId(c, query)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			ResponseNotFound(c, "用户不存在")
		} else {
			ResponseInternalServerError(c, "获取用户信息失败")
		}
	}
	// 返回用户信息
	ResponseOK(c, "获取成功", dto.ToUserResponse(user))

}

// SearchUsers 搜索用户处理器（新增，匹配前端期望）
// @Summary 搜索用户
// @Description 根据关键词搜索用户（姓名或邮箱）
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param q query string true "搜索关键词"
// @Param limit query int false "返回结果数量限制" default(10)
// @Param offset query int false "偏移量" default(0)
// @Success 200 {object} SearchUsersResponse "搜索成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "未授权"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/users/search [get]
func (h *UserHandler) SearchUsers(c *gin.Context) {
	// 1. 解析查询参数
	query := c.Query("q")
	if query == "" {
		ResponseBadRequest(c, "搜索关键词不能为空")
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	// 参数验证
	if limit < 1 || limit > 100 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	// 2. 调用业务逻辑层搜索用户
	users, total, err := h.userUsecase.SearchUsers(c.Request.Context(), query, limit, offset)
	if err != nil {
		ResponseInternalServerError(c, "搜索用户失败")
		return
	}

	// 3. 转换响应格式
	userResponses := make([]*dto.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = dto.ToUserResponse(user)
	}

	SearchResponse(c, http.StatusOK, userResponses, total)

}
