package domain

import (
	"context"
	"time"
)

type SpaceType string

const (
	SpaceTypeWorkspace SpaceType = "WORKSPACE" // 工作空间（默认）
	SpaceTypeProject   SpaceType = "PROJECT"   // 项目空间
	SpaceTypePersonal  SpaceType = "PERSONAL"  // 个人空间
)

type SpaceStatus int

const (
	SpaceStatusActive    SpaceStatus = iota // 激活
	SpaceStatusArchived                     // 归档（软删除，可恢复）
	SpaceStatusDeleted                      // 已删除（硬删除标记）
	SpaceStatusSuspended                    // 暂停（管理员操作）
)

type SpaceMemberRole string

const (
	SpaceRoleOwner  SpaceMemberRole = "OWNER"  // 所有者
	SpaceRoleAdmin  SpaceMemberRole = "ADMIN"  // 管理员
	SpaceRoleEditor SpaceMemberRole = "EDITOR" // 编辑者
	SpaceRoleViewer SpaceMemberRole = "VIEWER" // 查看者
	SpaceRoleGuest  SpaceMemberRole = "GUEST"  // 访客
)

// Space 空间实体（对应 CreateSpaceDto/UpdateSpaceDto）
type Space struct {
	ID          int64       `json:"id" gorm:"primaryKey;autoIncrement"`
	Name        string      `json:"name" gorm:"type:varchar(100);not null;index"`
	Description string      `json:"description" gorm:"type:text"`
	Icon        string      `json:"icon" gorm:"type:varchar(200)"`                             // 空间图标
	Color       string      `json:"color" gorm:"type:varchar(50)"`                             // 空间颜色
	Type        SpaceType   `json:"type" gorm:"type:varchar(20);not null;default:'WORKSPACE'"` // 空间类型
	IsPublic    bool        `json:"is_public" gorm:"default:false;index"`                      // 是否公开
	Status      SpaceStatus `json:"status" gorm:"type:tinyint;default:0"`

	// 关联关系
	OrganizationID *int64 `json:"organization_id" gorm:"index"` // 所属组织ID（可选，个人空间为nil）
	CreatedBy      int64  `json:"created_by" gorm:"not null;index"`

	// 时间字段
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// 统计信息
	DocumentCount int `json:"document_count" gorm:"default:0"`
	MemberCount   int `json:"member_count" gorm:"default:0"`

	// 关联数据 - 正确的GORM关系定义
	Organization *Organization  `json:"organization,omitempty" gorm:"foreignKey:OrganizationID"`
	Creator      *User          `json:"creator,omitempty" gorm:"foreignKey:CreatedBy"`
	Members      []*SpaceMember `json:"members,omitempty" gorm:"foreignKey:SpaceID"`
	Documents    []*Document    `json:"documents,omitempty" gorm:"many2many:space_documents"`
}

// SpaceMember 空间成员实体
type SpaceMember struct {
	ID      int64           `json:"id" gorm:"primaryKey;autoIncrement"`
	SpaceID int64           `json:"space_id" gorm:"not null;index"`
	UserID  int64           `json:"user_id" gorm:"not null;index"`
	Role    SpaceMemberRole `json:"role" gorm:"type:varchar(20);not null"`
	AddedBy int64           `json:"added_by" gorm:"not null"`
	AddedAt time.Time       `json:"added_at" gorm:"autoCreateTime"`

	// 关联数据 - 正确的GORM关系定义
	Space       *Space `json:"space,omitempty" gorm:"foreignKey:SpaceID"`
	User        *User  `json:"user,omitempty" gorm:"foreignKey:UserID"`
	AddedByUser *User  `json:"added_by_user,omitempty" gorm:"foreignKey:AddedBy"`
}

// SpaceDocument 空间文档关联实体（多对多关系）
type SpaceDocument struct {
	ID         int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	SpaceID    int64     `json:"space_id" gorm:"not null;index"`
	DocumentID int64     `json:"document_id" gorm:"not null;index"`
	AddedBy    int64     `json:"added_by" gorm:"not null"`
	AddedAt    time.Time `json:"added_at" gorm:"autoCreateTime"`

	// 关联数据 - 正确的GORM关系定义
	Space    *Space    `json:"space,omitempty" gorm:"foreignKey:SpaceID"`
	Document *Document `json:"document,omitempty" gorm:"foreignKey:DocumentID"`
}

// === 实体方法 ===

// Validate 验证空间实体
func (s *Space) Validate() error {
	if s.Name == "" {
		return ErrInvalidSpaceName
	}
	if s.CreatedBy <= 0 {
		return ErrInvalidUser
	}
	if s.Type != SpaceTypeWorkspace && s.Type != SpaceTypeProject && s.Type != SpaceTypePersonal {
		return ErrInvalidSpaceType
	}
	return nil
}

// IsActive 检查空间是否激活
func (s *Space) IsActive() bool {
	return s.Status == SpaceStatusActive
}

// CanAccess 检查是否可以访问空间
func (s *Space) CanAccess(userID int64) bool {
	// 创建者总是可以访问
	if s.CreatedBy == userID {
		return true
	}
	// 公开空间可以访问
	if s.IsPublic {
		return true
	}
	// 其他情况需要检查成员权限
	return false
}

// IsPersonalSpace 检查是否为个人空间
func (s *Space) IsPersonalSpace() bool {
	return s.Type == SpaceTypePersonal
}

// Validate 验证空间成员
func (sm *SpaceMember) Validate() error {
	if sm.SpaceID <= 0 {
		return ErrSpaceIDRequired
	}
	if sm.UserID <= 0 {
		return ErrInvalidUser
	}
	if !sm.isValidRole() {
		return ErrInvalidRole
	}
	return nil
}

// isValidRole 验证成员角色
func (sm *SpaceMember) isValidRole() bool {
	validRoles := []SpaceMemberRole{SpaceRoleOwner, SpaceRoleAdmin, SpaceRoleEditor, SpaceRoleViewer, SpaceRoleGuest}
	for _, role := range validRoles {
		if sm.Role == role {
			return true
		}
	}
	return false
}

// IsOwner 是否为所有者
func (sm *SpaceMember) IsOwner() bool {
	return sm.Role == SpaceRoleOwner
}

// IsAdmin 是否为管理员
func (sm *SpaceMember) IsAdmin() bool {
	return sm.Role == SpaceRoleAdmin
}

// CanManageMembers 是否可以管理成员
func (sm *SpaceMember) CanManageMembers() bool {
	return sm.Role == SpaceRoleOwner || sm.Role == SpaceRoleAdmin
}

// CanEditDocuments 是否可以编辑文档
func (sm *SpaceMember) CanEditDocuments() bool {
	return sm.Role == SpaceRoleOwner || sm.Role == SpaceRoleAdmin || sm.Role == SpaceRoleEditor
}

// CanViewDocuments 是否可以查看文档
func (sm *SpaceMember) CanViewDocuments() bool {
	return sm.Role != SpaceRoleGuest || sm.Role == SpaceRoleViewer || sm.CanEditDocuments()
}

// === 仓储接口 ===

// SpaceRepository 空间仓储接口
type SpaceRepository interface {
	// 空间基本操作
	Store(ctx context.Context, space *Space) error
	GetByID(ctx context.Context, id int64) (*Space, error)
	GetByName(ctx context.Context, name string, orgID *int64) (*Space, error)
	Update(ctx context.Context, space *Space) error
	Delete(ctx context.Context, id int64) error

	// 空间列表查询
	GetUserSpaces(ctx context.Context, userID int64) ([]*Space, error)
	GetOrganizationSpaces(ctx context.Context, orgID int64) ([]*Space, error)
	GetPublicSpaces(ctx context.Context, limit, offset int) ([]*Space, error)
	SearchSpaces(ctx context.Context, keyword string, userID int64, limit, offset int) ([]*Space, error)

	// 成员管理
	AddMember(ctx context.Context, member *SpaceMember) error
	GetMember(ctx context.Context, spaceID, userID int64) (*SpaceMember, error)
	GetMembers(ctx context.Context, spaceID int64) ([]*SpaceMember, error)
	UpdateMemberRole(ctx context.Context, spaceID, userID int64, role SpaceMemberRole) error
	RemoveMember(ctx context.Context, spaceID, userID int64) error

	// 文档管理
	AddDocument(ctx context.Context, spaceDocument *SpaceDocument) error
	RemoveDocument(ctx context.Context, spaceID, documentID int64) error
	GetSpaceDocuments(ctx context.Context, spaceID int64) ([]*Document, error)
	IsDocumentInSpace(ctx context.Context, spaceID, documentID int64) (bool, error)
}

// SpaceUsecase 空间业务逻辑接口
type SpaceUsecase interface {
	// 空间管理
	CreateSpace(ctx context.Context, para CreateSpacePara) (*Space, error)
	GetSpace(ctx context.Context, userID, spaceID int64) (*Space, error)
	UpdateSpace(ctx context.Context, para UpdateSpacePara) (*Space, error)
	DeleteSpace(ctx context.Context, userID, spaceID int64) error

	// 空间列表
	GetMySpaces(ctx context.Context, userID int64) ([]*Space, error)
	GetOrganizationSpaces(ctx context.Context, userID, orgID int64) ([]*Space, error)
	SearchSpaces(ctx context.Context, userID int64, keyword string, limit, offset int) ([]*Space, error)

	// 成员管理
	AddSpaceMember(ctx context.Context, userID, spaceID, memberUserID int64, role SpaceMemberRole) error
	UpdateMemberRole(ctx context.Context, userID, spaceID, memberUserID int64, role SpaceMemberRole) error
	RemoveSpaceMember(ctx context.Context, userID, spaceID, memberUserID int64) error
	GetSpaceMembers(ctx context.Context, userID, spaceID int64) ([]*SpaceMember, error)

	// 文档管理
	AddDocumentToSpace(ctx context.Context, userID, spaceID, documentID int64) error
	RemoveDocumentFromSpace(ctx context.Context, userID, spaceID, documentID int64) error
	GetSpaceDocuments(ctx context.Context, userID, spaceID int64) ([]*Document, error)

	// 权限检查
	CheckSpacePermission(ctx context.Context, userID, spaceID int64, action string) (bool, error)
	IsSpaceMember(ctx context.Context, userID, spaceID int64) (bool, error)
	GetUserSpaceRole(ctx context.Context, userID, spaceID int64) (SpaceMemberRole, error)
}

type CreateSpacePara struct {
	UserID    int64
	Name      string
	Desc      string
	Icon      string
	Color     string
	SpaceType SpaceType
	IsPublic  bool
	OrgID     *int64
}
type UpdateSpacePara struct {
	UserID    int64
	SpaceID   int64
	Name      *string
	Desc      *string
	Icon      *string
	Color     *string
	SpaceType *SpaceType
	IsPublic  *bool
}
