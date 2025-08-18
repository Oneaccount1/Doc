package dto

import (
	"encoding/json"
	"time"

	"DOC/domain"
)

// === 文档相关的DTO定义 ===
// 这些DTO对应API规范中的数据结构，用于HTTP请求和响应

// CreateDocumentDto 创建文档请求DTO
// 对应API规范中的CreateDocumentDto
type CreateDocumentDto struct {
	Title     string          `json:"title" binding:"required" validate:"required,min=1,max=255"` // 文档标题
	Content   json.RawMessage `json:"content,omitempty"`                                          // 文档内容（JSON格式）
	Type      string          `json:"type,omitempty" validate:"omitempty,oneof=FILE FOLDER"`      // 文档类型
	ParentID  *int64          `json:"parent_id,omitempty"`                                        // 父文件夹ID
	SortOrder int             `json:"sort_order,omitempty"`                                       // 排序顺序
	IsStarred bool            `json:"is_starred,omitempty"`                                       // 是否星标
	SpaceID   *int64          `json:"space_id,omitempty"`                                         // 所属空间ID
}

// ToDocumentType 转换为领域模型的文档类型
func (dto *CreateDocumentDto) ToDocumentType() domain.DocumentType {
	if dto.Type == "" {
		return domain.DocumentTypeFile // 默认为文件
	}
	return domain.DocumentType(dto.Type)
}

// GetContentString 获取内容的字符串形式
func (dto *CreateDocumentDto) GetContentString() string {
	if dto.Content == nil {
		return ""
	}
	return string(dto.Content)
}

// UpdateDocumentDto 更新文档请求DTO
// 对应API规范中的UpdateDocumentDto
type UpdateDocumentDto struct {
	Title     *string `json:"title,omitempty" validate:"omitempty,min=1,max=255"`    // 文档标题
	Type      *string `json:"type,omitempty" validate:"omitempty,oneof=FILE FOLDER"` // 文档类型
	ParentID  *int64  `json:"parent_id,omitempty"`                                   // 父文件夹ID
	SortOrder *int    `json:"sort_order,omitempty"`                                  // 排序顺序
	IsStarred *bool   `json:"is_starred,omitempty"`                                  // 是否星标
}

// ToDocumentType 转换为领域模型的文档类型
func (dto *UpdateDocumentDto) ToDocumentType() *domain.DocumentType {
	if dto.Type == nil {
		return nil
	}
	docType := domain.DocumentType(*dto.Type)
	return &docType
}

// UpdateDocumentContentDto 更新文档内容请求DTO
// 对应API规范中的UpdateDocumentContentDto
type UpdateDocumentContentDto struct {
	Content json.RawMessage `json:"content" binding:"required"` // 文档内容（JSON格式）
}

// GetContentString 获取内容的字符串形式
func (dto *UpdateDocumentContentDto) GetContentString() string {
	if dto.Content == nil {
		return ""
	}
	return string(dto.Content)
}

// ShareDocumentDto 分享文档请求DTO
// 对应API规范中的ShareDocumentDto
type ShareDocumentDto struct {
	Permission       string  `json:"permission" binding:"required" validate:"required,oneof=VIEW COMMENT EDIT MANAGE FULL"` // 分享权限
	Password         *string `json:"password,omitempty"`                                                                    // 密码保护
	ExpiresAt        *string `json:"expires_at,omitempty"`                                                                  // 过期时间（ISO格式字符串）
	ShareWithUserIDs []int64 `json:"shareWithUserIds,omitempty"`                                                            // 指定分享给的用户ID列表
}

// ToPermission 转换为领域模型的权限类型
func (dto *ShareDocumentDto) ToPermission() domain.Permission {
	return domain.Permission(dto.Permission)
}

// GetExpiresAt 获取过期时间
func (dto *ShareDocumentDto) GetExpiresAt() (*time.Time, error) {
	if dto.ExpiresAt == nil || *dto.ExpiresAt == "" {
		return nil, nil
	}

	t, err := time.Parse(time.RFC3339, *dto.ExpiresAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// DocumentSearchQueryDto 文档搜索查询DTO
type DocumentSearchQueryDto struct {
	Keyword string  `form:"keyword" binding:"required" validate:"required,min=1"`            // 搜索关键词
	Type    *string `form:"type,omitempty" validate:"omitempty,oneof=FILE FOLDER"`           // 文档类型过滤
	Limit   int     `form:"limit,omitempty" validate:"omitempty,min=1,max=100" default:"20"` // 每页数量
	Offset  int     `form:"offset,omitempty" validate:"omitempty,min=0" default:"0"`         // 偏移量
}

// ToDocumentType 转换为领域模型的文档类型
func (dto *DocumentSearchQueryDto) ToDocumentType() *domain.DocumentType {
	if dto.Type == nil {
		return nil
	}
	docType := domain.DocumentType(*dto.Type)
	return &docType
}

// BatchOperationDto 批量操作请求DTO
type BatchOperationDto struct {
	DocumentIDs []int64 `json:"document_ids" binding:"required,min=1,max=100"` // 文档ID列表
	NewParentID *int64  `json:"new_parent_id,omitempty"`                       // 新父文件夹ID（用于批量移动）
}

// === 响应DTO ===

// DocumentResponseDto 文档响应DTO
// 对应领域模型的Document，但包含额外的UI相关信息
type DocumentResponseDto struct {
	ID        int64     `json:"id"`                  // 文档ID
	Title     string    `json:"title"`               // 文档标题
	Type      string    `json:"type"`                // 文档类型
	ParentID  *int64    `json:"parent_id,omitempty"` // 父文件夹ID
	SpaceID   *int64    `json:"space_id,omitempty"`  // 所属空间ID
	OwnerID   int64     `json:"owner_id"`            // 所有者ID
	SortOrder int       `json:"sort_order"`          // 排序顺序
	IsStarred bool      `json:"is_starred"`          // 是否星标
	CreatedAt time.Time `json:"created_at"`          // 创建时间
	UpdatedAt time.Time `json:"updated_at"`          // 更新时间

	// 关联信息（可选）
	Owner         *UserInfoDto        `json:"owner,omitempty"`         // 所有者信息
	Parent        *DocumentBriefDto   `json:"parent,omitempty"`        // 父文档信息
	Children      []*DocumentBriefDto `json:"children,omitempty"`      // 子文档列表（仅文件夹）
	ChildrenCount int                 `json:"childrenCount,omitempty"` // 子项数量
}

// DocumentBriefDto 文档简要信息DTO
// 用于在列表或关联中显示文档的基本信息
type DocumentBriefDto struct {
	ID        int64     `json:"id"`         // 文档ID
	Title     string    `json:"title"`      // 文档标题
	Type      string    `json:"type"`       // 文档类型
	IsStarred bool      `json:"is_starred"` // 是否星标
	UpdatedAt time.Time `json:"updated_at"` // 更新时间
}

// DocumentSearchItemDto 文档搜索结果项DTO
// 对应API规范中的DocumentSearchItemDto
type DocumentSearchItemDto struct {
	ID            int64             `json:"id"`                      // 文档ID
	Title         string            `json:"title"`                   // 文档标题
	Type          string            `json:"type"`                    // 文档类型
	IsStarred     bool              `json:"is_starred"`              // 是否星标
	IsFavorite    bool              `json:"is_favorite"`             // 是否收藏
	CreatedAt     time.Time         `json:"created_at"`              // 创建时间
	UpdatedAt     time.Time         `json:"updated_at"`              // 更新时间
	LastViewed    *time.Time        `json:"last_viewed,omitempty"`   // 最后查看时间
	Parent        *DocumentBriefDto `json:"parent,omitempty"`        // 父文档信息
	ChildrenCount int               `json:"childrenCount,omitempty"` // 子项数量（仅文件夹）
	IsOwner       bool              `json:"isOwner"`                 // 是否为所有者
	Permission    string            `json:"permission"`              // 用户权限
	Owner         *UserInfoDto      `json:"owner,omitempty"`         // 所有者信息
	ParentPath    string            `json:"parent_path,omitempty"`   // 父路径
	MatchScore    float64           `json:"match_score,omitempty"`   // 匹配分数
}

// === DTO转换函数 ===

// FromDocument 从领域模型转换为响应DTO
func FromDocument(doc *domain.Document) *DocumentResponseDto {
	if doc == nil {
		return nil
	}

	dto := &DocumentResponseDto{
		ID:        doc.ID,
		Title:     doc.Title,
		Type:      string(doc.Type),
		ParentID:  doc.ParentID,
		SpaceID:   doc.SpaceID,
		OwnerID:   doc.OwnerID,
		SortOrder: doc.SortOrder,
		IsStarred: doc.IsStarred,
		CreatedAt: doc.CreatedAt,
		UpdatedAt: doc.UpdatedAt,
	}

	// 转换关联信息
	if doc.Owner != nil {
		dto.Owner = FromUser(doc.Owner)
	}

	if doc.Parent != nil {
		dto.Parent = FromDocumentBrief(doc.Parent)
	}

	if len(doc.Children) > 0 {
		dto.Children = make([]*DocumentBriefDto, len(doc.Children))
		for i, child := range doc.Children {
			dto.Children[i] = FromDocumentBrief(child)
		}
		dto.ChildrenCount = len(doc.Children)
	}

	return dto
}

// FromDocumentBrief 从领域模型转换为简要信息DTO
func FromDocumentBrief(doc *domain.Document) *DocumentBriefDto {
	if doc == nil {
		return nil
	}

	return &DocumentBriefDto{
		ID:        doc.ID,
		Title:     doc.Title,
		Type:      string(doc.Type),
		IsStarred: doc.IsStarred,
		UpdatedAt: doc.UpdatedAt,
	}
}

// FromDocumentSearchResult 从搜索结果转换为DTO
func FromDocumentSearchResult(result *domain.DocumentSearchResult) *DocumentSearchItemDto {
	if result == nil {
		return nil
	}

	dto := &DocumentSearchItemDto{
		ID:         result.ID,
		Title:      result.Title,
		Type:       string(result.Type),
		IsStarred:  result.IsStarred,
		IsFavorite: result.IsFavorite,
		UpdatedAt:  result.UpdatedAt,
		IsOwner:    result.Owner != nil,
		Permission: string(result.Permission),
		ParentPath: result.ParentPath,
		MatchScore: result.MatchScore,
	}

	if result.Owner != nil {
		dto.Owner = FromUser(result.Owner)
		dto.IsOwner = true
	}

	return dto
}
