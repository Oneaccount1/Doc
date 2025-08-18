package rest

import (
	"DOC/internal/rest/dto"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"DOC/domain"
)

// AuthHandler 认证 HTTP 处理器
// 负责处理认证相关的 HTTP 请求，包括登录、注册、OAuth、验证码等
// 这一层负责 HTTP 协议的处理，数据格式转换，以及调用业务逻辑层
type AuthHandler struct {
	authUsecase domain.AuthUsecase // 认证业务逻辑接口
	userUsecase domain.UserUsecase // 用户业务逻辑接口
}

// NewAuthHandler 创建新的认证处理器
// 参数说明：
// - authUsecase: 认证业务逻辑接口
// - userUsecase: 用户业务逻辑接口
// 返回值：
// - *AuthHandler: 认证处理器实例
func NewAuthHandler(authUsecase domain.AuthUsecase, userUsecase domain.UserUsecase) *AuthHandler {
	return &AuthHandler{
		authUsecase: authUsecase,
		userUsecase: userUsecase,
	}
}

// Register 用户注册处理器
// @Summary 用户注册
// @Description 创建新用户账户
// @Tags 认证
// @Accept json
// @Produce json
// @Param user body RegisterRequest true "注册信息"
// @Success 201 {object} UserResponse "注册成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 409 {object} ErrorResponse "用户已存在"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest

	// 1. 绑定和验证请求数据
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseBadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 2. 创建用户实体
	user := &domain.User{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
		Name:     req.DisplayName,
	}

	// 3. 调用业务逻辑层进行注册
	authResponse, err := h.authUsecase.Register(c.Request.Context(), user)
	if err != nil {
		// 根据不同的错误类型返回相应的 HTTP 状态码
		switch err {
		case domain.ErrUserAlreadyExist:
			ResponseConflict(c, "用户已存在")
		case domain.ErrInvalidUserEmail:
			ResponseBadRequest(c, "邮箱格式无效")
		case domain.ErrInvalidUserPassword:
			ResponseBadRequest(c, "密码格式无效")
		default:
			ResponseInternalServerError(c, "注册失败")
		}
		return
	}

	// 4. 返回成功响应 - 注册成功后返回完整的认证响应
	authData := dto.AuthResponse{
		User:             dto.ToUserResponse(authResponse.User),
		Token:            authResponse.AccessToken,
		RefreshToken:     authResponse.RefreshToken,
		ExpiresIn:        authResponse.ExpiresIn,  // domain层返回的已经是毫秒
		RefreshExpiresIn: 7 * 24 * 60 * 60 * 1000, // 7天，毫秒
	}
	ResponseCreated(c, "注册成功", authData)
}

// Login 用户登录处理器
// @Summary 用户登录
// @Description 用户身份验证并获取访问令牌
// @Tags 认证
// @Accept json
// @Produce json
// @Param credentials body LoginRequest true "登录凭据"
// @Success 200 {object} AuthResponse "登录成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "认证失败"
// @Failure 403 {object} ErrorResponse "账户被禁用"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest

	// 1. 绑定和验证请求数据
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseBadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 2. 调用业务逻辑层进行登录验证
	authResponse, err := h.authUsecase.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		// 根据错误类型返回相应状态码
		switch err {
		case domain.ErrInvalidCredentials:
			ResponseUnauthorized(c, "邮箱或密码错误")
		case domain.ErrUserNotActive:
			ResponseForbidden(c, "账户已被禁用")
		default:
			ResponseInternalServerError(c, "登录失败")
		}
		return
	}

	// 3. 设置认证cookie（24小时过期）
	h.setAuthCookie(c, authResponse.AccessToken, 24*60*60)

	// 4. 返回登录成功响应
	authData := dto.AuthResponse{
		User:             dto.ToUserResponse(authResponse.User),
		Token:            authResponse.AccessToken,
		RefreshToken:     authResponse.RefreshToken,
		ExpiresIn:        authResponse.ExpiresIn,  // domain层返回的是毫秒，保持原样
		RefreshExpiresIn: 7 * 24 * 60 * 60 * 1000, // 7天，毫秒
	}

	ResponseOK(c, "登录成功", authData)
}

// SendEmailCode 发送邮箱验证码处理器
// @Summary 发送邮箱验证码
// @Description 向指定邮箱发送验证码，用于注册或登录验证
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body SendEmailCodeRequest true "发送验证码请求"
// @Success 200 {object} SendCodeResponse "发送成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 429 {object} ErrorResponse "发送过于频繁"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/auth/email/send-code [post]
func (h *AuthHandler) SendEmailCode(c *gin.Context) {
	var req dto.SendEmailCodeRequest

	// 1. 绑定和验证请求数据
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseBadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 2. 调用业务逻辑层发送验证码
	if err := h.authUsecase.SendVerificationCode(c.Request.Context(), req.Email); err != nil {
		switch err {
		case domain.ErrInvalidUserEmail:
			ResponseBadRequest(c, "邮箱格式无效")
		case domain.ErrVerificationCodeSendTooFrequent:
			ResponseTooManyRequests(c, "验证码发送过于频繁，请稍后再试")
		default:
			ResponseInternalServerError(c, "发送验证码失败")
		}
		return
	}

	// 3. 返回成功响应
	codeData := dto.SendCodeResponse{
		Success: true,
		Message: "验证码发送成功",
	}
	ResponseCreated(c, "验证码发送成功", codeData)
}

// EmailLogin 邮箱验证码登录处理器
// @Summary 邮箱验证码登录
// @Description 使用邮箱验证码登录用户账户
// @Tags 认证
// @Accept json
// @Produce json
// @Param credentials body EmailLoginRequest true "登录凭据"
// @Success 200 {object} AuthResponse "登录成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "认证失败"
// @Failure 403 {object} ErrorResponse "账户被禁用"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/auth/email/login [post]
func (h *AuthHandler) EmailLogin(c *gin.Context) {
	var req dto.EmailLoginRequest

	// 1. 绑定和验证请求数据
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseBadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 2. 调用业务逻辑层进行邮箱登录
	authResponse, err := h.authUsecase.VerifyCode(c.Request.Context(), req.Email, req.Code)
	if err != nil {
		// 根据错误类型返回相应状态码
		switch err {
		case domain.ErrInvalidCredentials:
			ResponseUnauthorized(c, "邮箱或验证码错误")
		case domain.ErrInvalidUserEmail:
			ResponseBadRequest(c, "邮箱格式无效")
		case domain.ErrVerificationCodeNotFound:
			ResponseBadRequest(c, "验证码不存在或已过期")
		case domain.ErrInvalidVerificationCode:
			ResponseBadRequest(c, "验证码错误")
		case domain.ErrUserNotActive:
			ResponseForbidden(c, "账户已被禁用")
		default:
			ResponseInternalServerError(c, "登录失败")
		}
		return
	}

	// 3. 设置认证cookie（24小时过期）
	h.setAuthCookie(c, authResponse.AccessToken, 24*60*60)

	// 4. 构建响应数据
	authData := dto.AuthResponse{
		User:             dto.ToUserResponse(authResponse.User),
		Token:            authResponse.AccessToken,
		RefreshToken:     authResponse.RefreshToken,
		ExpiresIn:        authResponse.ExpiresIn,  // domain层返回的是毫秒，保持原样
		RefreshExpiresIn: 7 * 24 * 60 * 60 * 1000, // 7天，毫秒
	}

	ResponseCreated(c, "登录成功", authData)
}

// VerifyToken 验证令牌处理器
// @Summary 验证JWT令牌
// @Description 验证JWT令牌的有效性并返回用户信息
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body VerifyTokenRequest true "验证令牌请求"
// @Success 200 {object} VerifyTokenResponse "验证成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "令牌无效"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/auth/verify [post]
func (h *AuthHandler) VerifyToken(c *gin.Context) {
	var req dto.VerifyTokenRequest

	// 1. 绑定和验证请求数据
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseBadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 2. 调用业务逻辑层验证令牌
	user, err := h.authUsecase.ValidateToken(c.Request.Context(), req.Token)
	if err != nil {
		switch err {
		case domain.ErrInvalidToken:
			ResponseUnauthorized(c, "令牌无效")
		case domain.ErrTokenExpired:
			ResponseUnauthorized(c, "令牌已过期")
		default:
			ResponseInternalServerError(c, "验证失败")
		}
		return
	}

	// 3. 返回验证结果
	response := dto.VerifyTokenResponse{
		Success: true,
		User:    dto.ToUserResponse(user),
	}

	ResponseOK(c, "验证成功", response)
}

// RefreshToken 刷新令牌处理器
// @Summary 刷新访问令牌
// @Description 使用刷新令牌获取新的访问令牌
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "刷新令牌请求"
// @Success 200 {object} AuthResponse "刷新成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "刷新令牌无效"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest

	// 1. 绑定和验证请求数据
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseBadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 2. 调用业务逻辑层刷新令牌
	tokenResponse, err := h.authUsecase.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		switch err {
		case domain.ErrInvalidToken:
			ResponseUnauthorized(c, "刷新令牌无效")
		case domain.ErrTokenExpired:
			ResponseUnauthorized(c, "刷新令牌已过期")
		default:
			ResponseInternalServerError(c, "刷新失败")
		}
		return
	}

	// 3. 获取用户信息（从 token 中解析）
	_, err = h.authUsecase.ValidateToken(c.Request.Context(), tokenResponse.AccessToken)
	if err != nil {
		ResponseInternalServerError(c, "获取用户信息失败")
		return
	}

	// 4. 设置新的认证cookie
	h.setAuthCookie(c, tokenResponse.AccessToken, 24*60*60)

	// 5. 返回新令牌 - 使用专门的刷新令牌响应结构
	refreshData := dto.TokenRefreshResponse{
		Success:          true,
		Token:            tokenResponse.AccessToken,
		RefreshToken:     req.RefreshToken,        // 保持原有的刷新令牌
		ExpiresIn:        tokenResponse.ExpiresIn, // domain层返回的是毫秒，保持原样
		RefreshExpiresIn: 7 * 24 * 60 * 60 * 1000, // 7天，毫秒
	}

	ResponseOK(c, "刷新成功", refreshData)
}

// GetProfile 获取当前用户资料处理器
// @Summary 获取当前用户资料
// @Description 获取当前登录用户的详细信息
// @Tags 认证
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} UserResponse "获取成功"
// @Failure 401 {object} ErrorResponse "未授权"
// @Failure 404 {object} ErrorResponse "用户不存在"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/auth/profile [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	// 1. 从上下文中获取当前用户ID
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

	// 3. 调用用户业务逻辑获取用户信息
	user, err := h.userUsecase.GetUserByID(c.Request.Context(), uid)
	if err != nil {
		switch err {
		case domain.ErrUserNotFound:
			ResponseNotFound(c, "用户不存在")
		default:
			ResponseInternalServerError(c, "获取用户信息失败")
		}
		return
	}

	// 4. 返回用户信息
	userResponse := dto.ToUserResponse(user)
	ResponseOK(c, "获取成功", userResponse)
}

// Logout 退出登录处理器
// @Summary 退出登录
// @Description 退出当前用户登录状态
// @Tags 认证
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} SuccessResponse "退出成功"
// @Failure 401 {object} ErrorResponse "未授权"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// 1. 获取当前令牌
	token := c.GetHeader("Authorization")
	if token != "" && strings.HasPrefix(token, "Bearer ") {
		token = token[7:] // 去掉 "Bearer " 前缀
		// 2. 调用业务逻辑层退出登录（将令牌加入黑名单）
		if err := h.authUsecase.Logout(c.Request.Context(), token); err != nil {
			ResponseInternalServerError(c, "退出失败")
			return
		}
	}

	// 3. 清除认证相关的cookie
	h.clearAuthCookie(c)

	// 4. 返回成功响应
	ResponseOK(c, "退出成功", map[string]interface{}{
		"message": "已成功退出登录",
		"success": true,
	})
}

// LogoutAll 退出所有登录处理器
// @Summary 退出所有设备登录
// @Description 退出当前用户在所有设备上的登录状态
// @Tags 认证
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} SuccessResponse "退出成功"
// @Failure 401 {object} ErrorResponse "未授权"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/auth/logout-all [post]
func (h *AuthHandler) LogoutAll(c *gin.Context) {
	// 1. 从上下文中获取当前用户ID
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

	// 2. 调用业务逻辑层退出所有登录
	if err := h.authUsecase.LogoutAll(c.Request.Context(), uid); err != nil {
		ResponseInternalServerError(c, "退出失败")
		return
	}

	// 3. 清除认证相关的cookie
	h.clearAuthCookie(c)

	// 4. 返回成功响应
	ResponseOK(c, "退出成功", map[string]interface{}{
		"message": "已成功退出所有设备登录",
		"success": true,
	})
}

// === GitHub OAuth 相关处理器 ===

// GitHubLogin GitHub OAuth 登录处理器
// @Summary GitHub OAuth登录
// @Description 重定向到GitHub授权页面，开始OAuth登录流程
// @Tags 认证
// @Success 302 "重定向到GitHub授权页面"
// @Failure 500 "启动GitHub登录失败"
// @Router /api/v1/auth/github [get]
func (h *AuthHandler) GitHubLogin(c *gin.Context) {
	// 1. 调用业务逻辑层生成 OAuth URL
	authURL, _, err := h.authUsecase.GenerateOAuthURL(c.Request.Context(), domain.AuthProviderGitHub)
	if err != nil {
		ResponseInternalServerError(c, "启动GitHub登录失败")
		return
	}

	// 2. 重定向到 GitHub 授权页面
	c.Redirect(http.StatusFound, authURL)
}

// GitHubCallback GitHub OAuth 回调处理器
// @Summary GitHub OAuth回调
// @Description 处理GitHub OAuth授权回调，验证授权码并生成JWT令牌
// @Tags 认证
// @Param code query string true "GitHub OAuth授权码"
// @Param state query string false "状态参数，用于区分登录和绑定操作"
// @Success 302 "成功处理授权，重定向到前端页面并携带令牌"
// @Failure 401 "授权码缺失或无效"
// @Router /api/v1/auth/github/callback [get]
func (h *AuthHandler) GitHubCallback(c *gin.Context) {
	// 1. 获取授权码和状态参数
	code := c.Query("code")
	state := c.Query("state")

	if code == "" {
		ResponseUnauthorized(c, "授权码缺失")
		return
	}

	// 2. 调用业务逻辑层处理 OAuth 回调
	authResponse, err := h.authUsecase.HandleOAuthCallback(c.Request.Context(), domain.AuthProviderGitHub, code, state)
	if err != nil {
		ResponseInternalServerError(c, "GitHub登录失败"+err.Error())
		return
	}

	// 3. 设置认证cookie
	h.setAuthCookie(c, authResponse.AccessToken, 24*60*60)

	// 4. 重定向到前端页面，携带用户信息
	frontendURL := h.authUsecase.GetFrontedURL(authResponse)
	c.Redirect(302, frontendURL)
}

func (h *AuthHandler) GitHubBind(c *gin.Context) {
	// todo
}

// === Cookie 处理相关的辅助方法 ===

// setAuthCookie 设置认证相关的cookie
func (h *AuthHandler) setAuthCookie(c *gin.Context, token string, maxAge int) {
	// 设置token cookie
	c.SetCookie(
		"auth_token", // cookie名称
		token,        // cookie值
		maxAge,       // 过期时间（秒）
		"/",          // 路径
		"",           // 域名
		false,        // secure（HTTPS环境下应设为true）
		false,        // httpOnly（前端需要读取，所以设为false）
	)

	// 设置一个标识cookie，用于前端判断登录状态
	c.SetCookie(
		"isLoggedIn",
		"true",
		maxAge,
		"/",
		"",
		false,
		false,
	)
}

// clearAuthCookie 清除认证相关的cookie
func (h *AuthHandler) clearAuthCookie(c *gin.Context) {
	// 清除token cookie
	c.SetCookie(
		"auth_token",
		"",
		-1, // 设置为过期
		"/",
		"",
		false,
		false,
	)

	// 清除登录状态cookie
	c.SetCookie(
		"isLoggedIn",
		"",
		-1,
		"/",
		"",
		false,
		false,
	)
}
