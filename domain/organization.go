package domain

import (
	"context"
	"time"
)

// OrganizationStatus 组织状态枚举
type OrganizationStatus int

const (
	OrganizationStatusActive    OrganizationStatus = iota // 激活
	OrganizationStatusSuspended                           // 暂停
	OrganizationStatusDeleted                             // 已删除
)

// OrganizationMemberRole 组织成员角色枚举（对应API中的OWNER/ADMIN/MEMBER）
type OrganizationMemberRole string

const (
	OrgRoleOwner  OrganizationMemberRole = "OWNER"  // 所有者
	OrgRoleAdmin  OrganizationMemberRole = "ADMIN"  // 管理员
	OrgRoleMember OrganizationMemberRole = "MEMBER" // 普通成员
)

// JoinRequestStatus 加入申请状态枚举
type JoinRequestStatus int

const (
	JoinRequestStatusPending  JoinRequestStatus = iota // 待处理
	JoinRequestStatusApproved                          // 已批准
	JoinRequestStatusRejected                          // 已拒绝
	JoinRequestStatusExpired                           // 已过期
)

// Organization 组织实体（对应 CreateOrganizationDto/UpdateOrganizationDto）
type Organization struct {
	ID          int64              `json:"id" gorm:"primaryKey;autoIncrement"`
	Name        string             `json:"name" gorm:"type:varchar(100);not null;index"`
	Description string             `json:"description" gorm:"type:text"`
	Logo        string             `json:"logo" gorm:"type:varchar(500)"`        // 对应API中的logo
	Website     string             `json:"website" gorm:"type:varchar(500)"`     // 对应API中的website
	IsPublic    bool               `json:"is_public" gorm:"default:false;index"` // 对应API中的is_public
	Status      OrganizationStatus `json:"status" gorm:"type:tinyint;default:0"`

	// 创建者信息
	CreatedBy int64     `json:"created_by" gorm:"not null;index"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// 统计信息
	MemberCount int `json:"member_count" gorm:"default:0"`
	SpaceCount  int `json:"space_count" gorm:"default:0"`

	// 关联数据（不存储在数据库中）
	Creator *User                 `json:"creator,omitempty" gorm:"-"`
	Members []*OrganizationMember `json:"members,omitempty" gorm:"-"`
}

// OrganizationMember 组织成员实体
type OrganizationMember struct {
	ID             int64                  `json:"id" gorm:"primaryKey;autoIncrement"`
	OrganizationID int64                  `json:"organization_id" gorm:"not null;index"`
	UserID         int64                  `json:"user_id" gorm:"not null;index"`
	Role           OrganizationMemberRole `json:"role" gorm:"type:varchar(20);not null"`
	InvitedBy      int64                  `json:"invited_by" gorm:"not null"`
	JoinedAt       time.Time              `json:"joined_at" gorm:"autoCreateTime"`

	// 关联数据 - 正确的GORM关系定义
	Organization *Organization `json:"organization,omitempty" gorm:"foreignKey:OrganizationID"`
	User         *User         `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Inviter      *User         `json:"inviter,omitempty" gorm:"foreignKey:InvitedBy"`
}

// OrganizationInvitation 组织邀请实体（对应 InviteMemberDto）
type OrganizationInvitation struct {
	ID             int64                  `json:"id" gorm:"primaryKey;autoIncrement"`
	OrganizationID int64                  `json:"organization_id" gorm:"not null;index"`
	Email          string                 `json:"email" gorm:"type:varchar(255);not null;index"`
	Role           OrganizationMemberRole `json:"role" gorm:"type:varchar(20);not null"`
	Message        string                 `json:"message" gorm:"type:text"`
	Token          string                 `json:"token" gorm:"type:varchar(255);uniqueIndex;not null"`
	InvitedBy      int64                  `json:"invited_by" gorm:"not null"`
	InvitedAt      time.Time              `json:"invited_at" gorm:"autoCreateTime"`
	ExpiresAt      time.Time              `json:"expires_at" gorm:"not null;index"`
	AcceptedAt     *time.Time             `json:"accepted_at"`
	IsUsed         bool                   `json:"is_used" gorm:"default:false"`

	// 关联数据
	Organization *Organization `json:"organization,omitempty" gorm:"foreignKey:OrganizationID"`
	Inviter      *User         `json:"inviter,omitempty" gorm:"foreignKey:InvitedBy"`
}

// OrganizationJoinRequest 组织加入申请实体（对应 JoinRequestDto）
type OrganizationJoinRequest struct {
	ID             int64             `json:"id" gorm:"primaryKey;autoIncrement"`
	OrganizationID int64             `json:"organization_id" gorm:"not null;index"`
	UserID         int64             `json:"user_id" gorm:"not null;index"`
	Message        string            `json:"message" gorm:"type:text"`
	Status         JoinRequestStatus `json:"status" gorm:"type:tinyint;default:0;index"`
	ProcessedBy    *int64            `json:"processed_by" gorm:"index"`
	ProcessedAt    *time.Time        `json:"processed_at"`
	ProcessNote    string            `json:"process_note" gorm:"type:text"`
	CreatedAt      time.Time         `json:"created_at" gorm:"autoCreateTime"`

	// 关联数据
	Organization *Organization `json:"organization,omitempty" gorm:"foreignKey:OrganizationID"`
	User         *User         `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Processor    *User         `json:"processor,omitempty" gorm:"foreignKey:ProcessedBy"`
}

// === 实体方法 ===

// Validate 验证组织实体
func (o *Organization) Validate() error {
	if o.Name == "" {
		return ErrInvalidOrganizationName
	}
	if o.CreatedBy <= 0 {
		return ErrInvalidUser
	}
	return nil
}

// IsActive 检查组织是否激活
func (o *Organization) IsActive() bool {
	return o.Status == OrganizationStatusActive
}

// CanJoin 检查是否可以加入组织
func (o *Organization) CanJoin() bool {
	return o.IsActive() && o.IsPublic
}

// Validate 验证组织成员
func (om *OrganizationMember) Validate() error {
	if om.OrganizationID <= 0 {
		return ErrOrganizationIDRequired
	}
	if om.UserID <= 0 {
		return ErrInvalidUser
	}
	if !om.isValidRole() {
		return ErrInvalidRole
	}
	return nil
}

// isValidRole 验证成员角色
func (om *OrganizationMember) isValidRole() bool {
	validRoles := []OrganizationMemberRole{OrgRoleOwner, OrgRoleAdmin, OrgRoleMember}
	for _, role := range validRoles {
		if om.Role == role {
			return true
		}
	}
	return false
}

// IsOwner 是否为所有者
func (om *OrganizationMember) IsOwner() bool {
	return om.Role == OrgRoleOwner
}

// IsAdmin 是否为管理员
func (om *OrganizationMember) IsAdmin() bool {
	return om.Role == OrgRoleAdmin
}

// CanManageMembers 是否可以管理成员
func (om *OrganizationMember) CanManageMembers() bool {
	return om.Role == OrgRoleOwner || om.Role == OrgRoleAdmin
}

// Validate 验证邀请
func (oi *OrganizationInvitation) Validate() error {
	if oi.OrganizationID <= 0 {
		return ErrOrganizationIDRequired
	}
	if oi.Email == "" {
		return ErrInvalidEmailAddress
	}
	if oi.InvitedBy <= 0 {
		return ErrInvalidUser
	}
	return nil
}

// IsExpired 检查邀请是否过期
func (oi *OrganizationInvitation) IsExpired() bool {
	return oi.ExpiresAt.Before(time.Now())
}

// IsValid 检查邀请是否有效
func (oi *OrganizationInvitation) IsValid() bool {
	return !oi.IsUsed && !oi.IsExpired()
}

// Accept 接受邀请
func (oi *OrganizationInvitation) Accept() {
	oi.IsUsed = true
	now := time.Now()
	oi.AcceptedAt = &now
}

// Validate 验证加入申请
func (ojr *OrganizationJoinRequest) Validate() error {
	if ojr.OrganizationID <= 0 {
		return ErrOrganizationIDRequired
	}
	if ojr.UserID <= 0 {
		return ErrInvalidUser
	}
	return nil
}

// IsPending 是否待处理
func (ojr *OrganizationJoinRequest) IsPending() bool {
	return ojr.Status == JoinRequestStatusPending
}

// Approve 批准申请
func (ojr *OrganizationJoinRequest) Approve(processedBy int64, note string) {
	ojr.Status = JoinRequestStatusApproved
	ojr.ProcessedBy = &processedBy
	ojr.ProcessNote = note
	now := time.Now()
	ojr.ProcessedAt = &now
}

// Reject 拒绝申请
func (ojr *OrganizationJoinRequest) Reject(processedBy int64, note string) {
	ojr.Status = JoinRequestStatusRejected
	ojr.ProcessedBy = &processedBy
	ojr.ProcessNote = note
	now := time.Now()
	ojr.ProcessedAt = &now
}

// === 仓储接口 ===

// OrganizationRepository 组织仓储接口
type OrganizationRepository interface {
	// 组织基本操作
	Store(ctx context.Context, organization *Organization) error
	GetByID(ctx context.Context, id int64) (*Organization, error)
	GetByName(ctx context.Context, name string) (*Organization, error)
	Update(ctx context.Context, organization *Organization) error
	Delete(ctx context.Context, id int64) error

	// 组织列表查询
	GetUserOrganizations(ctx context.Context, userID int64) ([]*Organization, error)
	GetPublicOrganizations(ctx context.Context, limit, offset int) ([]*Organization, error)
	SearchOrganizations(ctx context.Context, keyword string, isPublic *bool, limit, offset int) ([]*Organization, error)

	// 成员管理
	AddMember(ctx context.Context, member *OrganizationMember) error
	GetMember(ctx context.Context, orgID, userID int64) (*OrganizationMember, error)
	GetMembers(ctx context.Context, orgID int64) ([]*OrganizationMember, error)
	UpdateMemberRole(ctx context.Context, orgID, userID int64, role OrganizationMemberRole) error
	RemoveMember(ctx context.Context, orgID, userID int64) error

	// 邀请管理
	StoreInvitation(ctx context.Context, invitation *OrganizationInvitation) error
	GetInvitationByToken(ctx context.Context, token string) (*OrganizationInvitation, error)
	GetInvitationsByOrg(ctx context.Context, orgID int64) ([]*OrganizationInvitation, error)
	UpdateInvitation(ctx context.Context, invitation *OrganizationInvitation) error
	DeleteInvitation(ctx context.Context, id int64) error

	// 加入申请管理
	StoreJoinRequest(ctx context.Context, request *OrganizationJoinRequest) error
	GetJoinRequest(ctx context.Context, id int64) (*OrganizationJoinRequest, error)
	GetJoinRequestsByOrg(ctx context.Context, orgID int64) ([]*OrganizationJoinRequest, error)
	GetJoinRequestsByUser(ctx context.Context, userID int64) ([]*OrganizationJoinRequest, error)
	UpdateJoinRequest(ctx context.Context, request *OrganizationJoinRequest) error

	// 清理操作
	CleanupExpiredInvitations(ctx context.Context) error
	CleanupExpiredJoinRequests(ctx context.Context) error
}

// OrganizationUsecase 组织业务逻辑接口
type OrganizationUsecase interface {
	// 组织管理
	CreateOrganization(ctx context.Context, userID int64, name, description, logo, website string, isPublic bool) (*Organization, error)
	GetOrganization(ctx context.Context, id int64) (*Organization, error)
	UpdateOrganization(ctx context.Context, userID, orgID int64, name, description, logo, website string, isPublic *bool) (*Organization, error)
	DeleteOrganization(ctx context.Context, userID, orgID int64) error

	// 组织列表
	GetMyOrganizations(ctx context.Context, userID int64) ([]*Organization, error)
	GetPublicOrganizations(ctx context.Context, limit, offset int) ([]*Organization, error)
	SearchOrganizations(ctx context.Context, keyword string, isPublic *bool, limit, offset int) ([]*Organization, error)

	// 成员管理
	InviteMember(ctx context.Context, userID, orgID int64, email, message string, role OrganizationMemberRole) error
	ProcessJoinRequest(ctx context.Context, userID, requestID int64, approve bool, note string) error
	RequestJoinOrganization(ctx context.Context, userID, orgID int64, message string) error
	UpdateMemberRole(ctx context.Context, userID, orgID, memberID int64, role OrganizationMemberRole) error
	RemoveMember(ctx context.Context, userID, orgID, memberID int64) error
	LeaveOrganization(ctx context.Context, userID, orgID int64) error

	// 成员查询
	GetOrganizationMembers(ctx context.Context, userID, orgID int64) ([]*OrganizationMember, error)
	GetMemberRole(ctx context.Context, userID, orgID int64) (OrganizationMemberRole, error)

	// 权限检查
	CheckPermission(ctx context.Context, userID, orgID int64, action string) (bool, error)
	IsOrganizationMember(ctx context.Context, userID, orgID int64) (bool, error)
	IsOrganizationAdmin(ctx context.Context, userID, orgID int64) (bool, error)
	IsOrganizationOwner(ctx context.Context, userID, orgID int64) (bool, error)
}
