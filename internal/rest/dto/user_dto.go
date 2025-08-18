package dto

import "DOC/domain"

// GetUserByIdRequest 根据ID获取用户请求结构
type GetUserByIdRequest struct {
	UserID int64 `uri:"id" binding:"required,min=1"` // 用户ID
}

// GetUserByEmailRequest 根据邮箱获取用户请求结构
type GetUserByEmailRequest struct {
	Email string `uri:"email" binding:"required,email"` // 邮箱地址
}

// UpdateProfileRequest 更新用户资料请求结构
type UpdateProfileRequest struct {
	Name      string `json:"name" binding:"required,min=1,max=100"` // 显示名称，对应前端的name字段
	AvatarURL string `json:"avatar_url" binding:"max=500"`          // 头像URL，对应前端的avatar_url字段
	Bio       string `json:"bio" binding:"max=500"`                 // 个人简介
}

// 响应部分

// UserResponse 用户响应结构
// 定义返回给前端的用户信息格式，与前端期望的User接口完全匹配
type UserResponse struct {
	ID          int64   `json:"id"`            // 用户ID
	Name        string  `json:"name"`          // 显示名称
	AvatarURL   string  `json:"avatar_url"`    // 头像URL
	Bio         *string `json:"bio"`           // 个人简介（可为null）
	Company     *string `json:"company"`       // 公司（可为null）
	CreatedAt   string  `json:"created_at"`    // 创建时间
	Email       *string `json:"email"`         // 邮箱（可为null）
	GitHubID    *string `json:"github_id"`     // GitHub ID（可为null）
	LastLoginAt string  `json:"last_login_at"` // 最后登录时间
	Location    *string `json:"location"`      // 位置（可为null）
	UpdatedAt   string  `json:"updated_at"`    // 更新时间
	WebsiteURL  *string `json:"website_url"`   // 个人网站（可为null）
}

// ToUserResponse 将领域实体转换为响应结构
func ToUserResponse(user *domain.User) *UserResponse {
	if user == nil {
		return nil
	}

	var lastLoginAt string
	if user.LastLoginAt != nil {
		lastLoginAt = user.LastLoginAt.Format("2006-01-02T15:04:05Z")
	}

	// 处理可为null的字段
	var bio *string
	if user.Bio != "" {
		bio = &user.Bio
	}

	var company *string
	if user.Company != "" {
		company = &user.Company
	}

	var email *string
	if user.Email != "" {
		email = &user.Email
	}

	var githubID *string
	if user.GitHubID != "" {
		githubID = &user.GitHubID
	}

	var location *string
	if user.Location != "" {
		location = &user.Location
	}

	var websiteURL *string
	if user.WebsiteURL != "" {
		websiteURL = &user.WebsiteURL
	}

	return &UserResponse{
		ID:          user.ID,
		Email:       email,
		Name:        user.Name,
		AvatarURL:   user.AvatarURL,
		GitHubID:    githubID,
		Bio:         bio,
		Location:    location,
		Company:     company,
		WebsiteURL:  websiteURL,
		LastLoginAt: lastLoginAt,
		CreatedAt:   user.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   user.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
