package dto

// RegisterRequest 用户注册请求结构
type RegisterRequest struct {
	Username    string `json:"username" binding:"required,min=3,max=50"`  // 用户名，3-50字符
	Email       string `json:"email" binding:"required,email"`            // 邮箱，必须是有效邮箱格式
	Password    string `json:"password" binding:"required,min=8,max=128"` // 密码，8-128字符
	DisplayName string `json:"display_name" binding:"max=100"`            // 显示名称，可选，最大100字符
}

// LoginRequest 用户登录请求结构
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`    // 邮箱
	Password string `json:"password" binding:"required,min=1"` // 密码
}

// SendEmailCodeRequest 发送邮箱验证码请求结构
type SendEmailCodeRequest struct {
	Email string `json:"email" binding:"required,email"` // 邮箱地址
}

// EmailLoginRequest 邮箱验证码登录请求结构
type EmailLoginRequest struct {
	Email string `json:"email" binding:"required,email"` // 邮箱地址
	Code  string `json:"code" binding:"required,len=6"`  // 验证码，6位数字
}

// ChangePasswordRequest 修改密码请求结构
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`       // 原密码
	NewPassword string `json:"new_password" binding:"required,min=8"` // 新密码
}

// VerifyTokenRequest 验证令牌请求结构
type VerifyTokenRequest struct {
	Token string `json:"token" binding:"required"` // JWT 令牌
}

// RefreshTokenRequest 刷新令牌请求结构
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"` // 刷新令牌
}

// 响应部分

// TokenRefreshResponse 令牌刷新响应结构（匹配API规范）
type TokenRefreshResponse struct {
	Success          bool   `json:"success"`          // 操作是否成功
	Token            string `json:"token"`            // JWT访问令牌
	RefreshToken     string `json:"refreshToken"`     // JWT刷新令牌
	ExpiresIn        int64  `json:"expiresIn"`        // 访问令牌过期时间（毫秒）
	RefreshExpiresIn int64  `json:"refreshExpiresIn"` // 刷新令牌过期时间（毫秒）
}

// SendCodeResponse 发送验证码响应结构
type SendCodeResponse struct {
	Success bool   `json:"success"` // 发送是否成功
	Message string `json:"message"` // 响应消息
}

// VerifyTokenResponse 验证令牌响应结构
type VerifyTokenResponse struct {
	Success bool          `json:"success"` // 验证是否成功
	User    *UserResponse `json:"user"`    // 用户信息
}

// AuthResponse 认证响应结构（匹配前端期望的AuthResponse接口）
// 根据用户提供的结构调整字段名和类型
type AuthResponse struct {
	User             *UserResponse `json:"user"`                       // 用户信息
	Token            string        `json:"token"`                      // JWT 访问令牌
	RefreshToken     string        `json:"refreshToken,omitempty"`     // 刷新令牌（可选）
	ExpiresIn        int64         `json:"expiresIn"`                  // 令牌过期时间（毫秒）
	RefreshExpiresIn int64         `json:"refreshExpiresIn,omitempty"` // 刷新令牌过期时间（毫秒）
}
