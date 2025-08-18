package domain

import (
	"context"
	"time"
)

// AuthProvider 认证提供商枚举
type AuthProvider string

const (
	AuthProviderEmail  AuthProvider = "email"  // 邮箱密码认证
	AuthProviderCode   AuthProvider = "code"   // 邮箱验证码认证
	AuthProviderGitHub AuthProvider = "github" // GitHub OAuth认证
)

// TokenType 令牌类型枚举
type TokenType string

const (
	TokenTypeAccess  TokenType = "access"  // 访问令牌
	TokenTypeRefresh TokenType = "refresh" // 刷新令牌
)

// SessionStatus 会话状态枚举
type SessionStatus int

const (
	SessionStatusActive  SessionStatus = iota // 活跃
	SessionStatusExpired                      // 已过期
	SessionStatusRevoked                      // 已撤销
)

// AuthSession 认证会话实体
// 管理用户的登录会话，支持多设备登录
type AuthSession struct {
	ID           int64         `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID       int64         `json:"user_id" gorm:"not null;index"`
	SessionToken string        `json:"session_token" gorm:"type:varchar(512);uniqueIndex;not null"`
	RefreshToken string        `json:"refresh_token" gorm:"type:varchar(512);uniqueIndex;not null"`
	Provider     AuthProvider  `json:"provider" gorm:"type:varchar(20);not null"`
	Status       SessionStatus `json:"status" gorm:"type:tinyint;default:0;index"`

	// 时间信息
	ExpiresAt  time.Time  `json:"expires_at" gorm:"not null;index"`
	LastUsedAt time.Time  `json:"last_used_at" gorm:"autoUpdateTime"`
	CreatedAt  time.Time  `json:"created_at" gorm:"autoCreateTime"`
	RevokedAt  *time.Time `json:"revoked_at"`

	// 关联数据
	User *User `json:"user,omitempty" gorm:"-"`
}

// Validate 验证认证会话
func (s *AuthSession) Validate() error {
	if s.UserID <= 0 {
		return ErrInvalidUser
	}
	if s.SessionToken == "" {
		return ErrInvalidCredentials
	}
	if s.RefreshToken == "" {
		return ErrInvalidCredentials
	}
	if s.ExpiresAt.Before(time.Now()) {
		return ErrUnauthorized
	}
	return nil
}

// IsExpired 检查会话是否过期
func (s *AuthSession) IsExpired() bool {
	return s.ExpiresAt.Before(time.Now()) || s.Status == SessionStatusExpired
}

// IsActive 检查会话是否活跃
func (s *AuthSession) IsActive() bool {
	return s.Status == SessionStatusActive && !s.IsExpired()
}

// Revoke 撤销会话
func (s *AuthSession) Revoke() {
	s.Status = SessionStatusRevoked
	now := time.Now()
	s.RevokedAt = &now
}

// UpdateLastUsed 更新最后使用时间
func (s *AuthSession) UpdateLastUsed() {
	s.LastUsedAt = time.Now()
}

// VerificationCode 验证码实体
type VerificationCode struct {
	ID    int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	Email string `json:"email" gorm:"type:varchar(255);not null;index"`
	Code  string `json:"code" gorm:"type:varchar(10);not null"`
	Used  bool   `json:"used" gorm:"default:false"`
	// 时间字段
	ExpiresAt time.Time  `json:"expires_at" gorm:"not null;index"`
	CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UsedAt    *time.Time `json:"used_at"`
}

// Validate 验证验证码
func (v *VerificationCode) Validate() error {
	if v.Email == "" {
		return ErrInvalidEmailAddress
	}
	if v.Code == "" {
		return ErrInvalidVerificationCode
	}
	if v.ExpiresAt.Before(time.Now()) {
		return ErrVerificationCodeNotFound
	}
	return nil
}

// IsExpired 检查验证码是否过期
func (v *VerificationCode) IsExpired() bool {
	return v.ExpiresAt.Before(time.Now())
}

// IsUsed 检查验证码是否已使用
func (v *VerificationCode) IsUsed() bool {
	return v.Used
}

// MarkAsUsed 标记为已使用
func (v *VerificationCode) MarkAsUsed() {
	v.Used = true
	now := time.Now()
	v.UsedAt = &now
}

// OAuthState OAuth状态实体
// 管理OAuth认证过程中的状态信息
type OAuthState struct {
	ID        int64        `json:"id" gorm:"primaryKey;autoIncrement"`
	State     string       `json:"state" gorm:"type:varchar(255);uniqueIndex;not null"`
	Provider  AuthProvider `json:"provider" gorm:"type:varchar(20);not null"`
	UserID    *int64       `json:"user_id" gorm:"index"` // 可选，用于账户关联
	ExpiresAt time.Time    `json:"expires_at" gorm:"not null;index"`
	CreatedAt time.Time    `json:"created_at" gorm:"autoCreateTime"`
	UsedAt    *time.Time   `json:"used_at"`
}

// Validate 验证OAuth状态
func (o *OAuthState) Validate() error {
	if o.State == "" {
		return ErrOAuthStateMismatch
	}
	if o.ExpiresAt.Before(time.Now()) {
		return ErrOAuthCodeInvalid
	}
	return nil
}

// IsExpired 检查OAuth状态是否过期
func (o *OAuthState) IsExpired() bool {
	return o.ExpiresAt.Before(time.Now())
}

// MarkAsUsed 标记为已使用
func (o *OAuthState) MarkAsUsed() {
	now := time.Now()
	o.UsedAt = &now
}

// AuthRepository 认证仓储接口
type AuthRepository interface {
	// 会话管理
	StoreSession(ctx context.Context, session *AuthSession) error
	GetSessionByToken(ctx context.Context, token string) (*AuthSession, error)
	GetSessionsByUserID(ctx context.Context, userID int64) ([]*AuthSession, error)
	UpdateSession(ctx context.Context, session *AuthSession) error
	RevokeSession(ctx context.Context, sessionID int64) error
	RevokeAllUserSessions(ctx context.Context, userID int64) error

	// 验证码管理
	StoreVerificationCode(ctx context.Context, code *VerificationCode) error
	GetVerificationCode(ctx context.Context, email string) (*VerificationCode, error)
	DeleteVerificationCode(ctx context.Context, id int64) error
	CleanupExpiredCodes(ctx context.Context) error

	// OAuth状态管理
	StoreOAuthState(ctx context.Context, state *OAuthState) error
	GetOAuthState(ctx context.Context, state string) (*OAuthState, error)
	DeleteOAuthState(ctx context.Context, id int64) error
	CleanupExpiredStates(ctx context.Context) error
}

// AuthUsecase 认证业务逻辑接口
type AuthUsecase interface {
	// 基础认证
	Login(ctx context.Context, email, password string) (*AuthResponse, error)
	Register(ctx context.Context, user *User) (*AuthResponse, error)
	Logout(ctx context.Context, sessionToken string) error
	LogoutAll(ctx context.Context, userID int64) error

	// 验证码认证
	SendVerificationCode(ctx context.Context, email string) error
	VerifyCode(ctx context.Context, email, code string) (*AuthResponse, error)

	// OAuth认证
	GenerateOAuthURL(ctx context.Context, provider AuthProvider) (string, string, error)
	HandleOAuthCallback(ctx context.Context, provider AuthProvider, code, state string) (*AuthResponse, error)
	GetFrontedURL(response *AuthResponse) string

	// 令牌管理
	RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error)
	ValidateToken(ctx context.Context, accessToken string) (*User, error)
	RevokeToken(ctx context.Context, sessionToken string) error

	// 会话管理
	GetUserSessions(ctx context.Context, userID int64) ([]*AuthSession, error)
	RevokeSession(ctx context.Context, userID int64, sessionID int64) error
}

// AuthResponse 认证响应
type AuthResponse struct {
	User         *User  `json:"user"`
	AccessToken  string `json:"token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int64  `json:"expires_in"`
}

// TokenResponse 令牌响应
type TokenResponse struct {
	AccessToken string `json:"token"`
	ExpiresIn   int64  `json:"expires_in"`
}

// AuthCacheRepository 认证缓存数据访问接口
// 专注于认证相关的缓存功能
type AuthCacheRepository interface {
	// 邮箱验证码缓存
	StoreEmailCode(ctx context.Context, email, code string, expiration time.Duration) error
	GetEmailCode(ctx context.Context, email string) (string, error)
	DeleteEmailCode(ctx context.Context, email string) error

	// 发送间隔控制
	CheckSendInterval(ctx context.Context, email string) (bool, error)
	SetSendInterval(ctx context.Context, email string, interval time.Duration) error

	// 会话缓存（可选优化）
	CacheSession(ctx context.Context, session *AuthSession) error
	GetCachedSession(ctx context.Context, token string) (*AuthSession, error)
	DeleteCachedSession(ctx context.Context, token string) error

	// OAuth状态缓存
	CacheOAuthState(ctx context.Context, state string, data interface{}, expiration time.Duration) error
	GetCachedOAuthState(ctx context.Context, state string) (interface{}, error)
	DeleteCachedOAuthState(ctx context.Context, state string) error
}

// OAuthUserInfo OAuth用户信息
type OAuthUserInfo struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Avatar   string `json:"avatar"`
	Username string `json:"username"`
}
