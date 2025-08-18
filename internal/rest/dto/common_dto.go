package dto

import (
	"DOC/domain"
	"time"
)

// === 通用DTO定义 ===

// UserInfoDto 用户信息DTO
// 用于在其他DTO中引用用户信息
type UserInfoDto struct {
	ID       int64   `json:"id"`                 // 用户ID
	Username string  `json:"username"`           // 用户名
	Email    string  `json:"email"`              // 邮箱
	Nickname *string `json:"nickname,omitempty"` // 昵称
	Avatar   *string `json:"avatar,omitempty"`   // 头像URL
}

// FromUser 从用户领域模型转换为DTO
func FromUser(user *domain.User) *UserInfoDto {
	if user == nil {
		return nil
	}

	return &UserInfoDto{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		//Nickname: user.Nickname,
		//Avatar:   user.Avatar,
	}
}

// === 文档分享相关DTO ===

// DocumentShareResponseDto 文档分享响应DTO
type DocumentShareResponseDto struct {
	ID          int64        `json:"id"`                   // 分享ID
	DocumentID  int64        `json:"document_id"`          // 文档ID
	LinkID      string       `json:"link_id"`              // 分享链接ID
	Permission  string       `json:"permission"`           // 分享权限
	HasPassword bool         `json:"has_password"`         // 是否设置密码
	ExpiresAt   *time.Time   `json:"expires_at,omitempty"` // 过期时间
	CreatedAt   time.Time    `json:"created_at"`           // 创建时间
	AccessCount int          `json:"access_count"`         // 访问次数
	IsActive    bool         `json:"is_active"`            // 是否激活
	ShareURL    string       `json:"share_url"`            // 完整分享URL
	CreatedBy   *UserInfoDto `json:"created_by,omitempty"` // 创建者信息
}

// FromDocumentShare 从分享领域模型转换为DTO
func FromDocumentShare(share *domain.DocumentShare) *DocumentShareResponseDto {
	if share == nil {
		return nil
	}

	dto := &DocumentShareResponseDto{
		ID:          share.ID,
		DocumentID:  share.DocumentID,
		LinkID:      share.LinkID,
		Permission:  string(share.Permission),
		HasPassword: share.Password != "",
		ExpiresAt:   share.ExpiresAt,
		CreatedAt:   share.CreatedAt,
		//AccessCount: share.AccessCount,
		//IsActive:    share.IsActive(),
		//ShareURL:    share.GetShareURL(), // 假设领域模型有此方法
	}

	//if share.CreatedBy != nil {
	//	dto.CreatedBy = FromUser(share.CreatedBy)
	//}

	return dto
}

//// === 文档权限相关DTO ===
//
//// DocumentPermissionResponseDto 文档权限响应DTO
//type DocumentPermissionResponseDto struct {
//	ID            int64        `json:"id"`                        // 权限ID
//	DocumentID    int64        `json:"document_id"`               // 文档ID
//	UserID        int64        `json:"user_id"`                   // 用户ID
//	Permission    string       `json:"permission"`                // 权限级别
//	GrantedAt     time.Time    `json:"granted_at"`                // 授权时间
//	GrantedBy     int64        `json:"granted_by"`                // 授权者ID
//	User          *UserInfoDto `json:"user,omitempty"`            // 用户信息
//	GrantedByUser *UserInfoDto `json:"granted_by_user,omitempty"` // 授权者信息
//}

// === 通用响应结构 ===

// ApiResponse 通用API响应结构
type ApiResponse struct {
	Success bool        `json:"success"`           // 操作是否成功
	Message string      `json:"message,omitempty"` // 响应消息
	Data    interface{} `json:"data,omitempty"`    // 响应数据
	Code    string      `json:"code,omitempty"`    // 错误代码
}

// ErrorResponse 错误响应
func ErrorResponse(message string, code ...string) *ApiResponse {
	resp := &ApiResponse{
		Success: false,
		Message: message,
	}

	if len(code) > 0 {
		resp.Code = code[0]
	}

	return resp
}

// === 分页相关DTO ===

// PaginationDto 分页参数DTO
type PaginationDto struct {
	Page     int `form:"page,omitempty" validate:"omitempty,min=1" default:"1"`               // 页码
	PageSize int `form:"page_size,omitempty" validate:"omitempty,min=1,max=100" default:"20"` // 每页数量
}

// === 参数验证相关 ===

// IDParamDto ID路径参数DTO
type IDParamDto struct {
	ID int64 `uri:"id" binding:"required,min=1"` // ID参数
}

// DocumentIDParamDto 文档ID路径参数DTO
type DocumentIDParamDto struct {
	DocumentID int64 `uri:"documentId" binding:"required,min=1"` // 文档ID参数
}

// ShareLinkParamDto 分享链接路径参数DTO
type ShareLinkParamDto struct {
	LinkID string `uri:"linkId" binding:"required"` // 分享链接ID参数
}

// === 查询参数DTO ===

// DocumentQueryDto 文档查询参数DTO
type DocumentQueryDto struct {
	ParentID       *int64  `form:"parent_id,omitempty"`                                                                 // 父文件夹ID
	SpaceID        *int64  `form:"space_id,omitempty"`                                                                  // 空间ID
	IncludeDeleted bool    `form:"include_deleted,omitempty"`                                                           // 是否包含已删除的文档
	Type           *string `form:"type,omitempty" validate:"omitempty,oneof=FILE FOLDER"`                               // 文档类型过滤
	SortBy         string  `form:"sort_by,omitempty" validate:"omitempty,oneof=title created_at updated_at sort_order"` // 排序字段
	SortOrder      string  `form:"sort_order,omitempty" validate:"omitempty,oneof=asc desc"`                            // 排序方向
	PaginationDto          // 嵌入分页参数
}
