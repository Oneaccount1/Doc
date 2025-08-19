package domain

import (
	"context"
	"encoding/json"
	"regexp"
	"time"

	"gorm.io/gorm"
)

// UserStatus 用户状态枚举
type UserStatus int

const (
	UserStatusInactive  UserStatus = iota // 未激活
	UserStatusActive                      // 激活
	UserStatusSuspended                   // 暂停
	UserStatusDeleted                     // 已删除
)

// User 用户实体
// 领域层核心实体，包含用户基本信息和业务规则
type User struct {
	ID         int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	Username   string `json:"username" gorm:"type:varchar(50);uniqueIndex;not null" validate:"required,min=3,max=50"`
	Name       string `json:"name" gorm:"type:varchar(100)"` // 显示名称
	Email      string `json:"email" gorm:"type:varchar(100);uniqueIndex;not null" validate:"required,email"`
	Password   string `json:"-" gorm:"type:varchar(255)"`           // 密码，OAuth用户可为空
	AvatarURL  string `json:"avatar_url" gorm:"type:varchar(500)"`  // 头像URL
	Bio        string `json:"bio" gorm:"type:text"`                 // 个人简介
	Company    string `json:"company" gorm:"type:varchar(200)"`     // 公司
	Location   string `json:"location" gorm:"type:varchar(200)"`    // 位置
	WebsiteURL string `json:"website_url" gorm:"type:varchar(500)"` // 个人网站
	// todo 目前未使用Role字段
	Role   string     `json:"role" gorm:"type:varchar(50);default:'user'"`   // 角色
	Status UserStatus `json:"status" gorm:"type:tinyint;not null;default:0"` // 用户状态

	// OAuth相关
	GitHubID    string `json:"github_id" gorm:"type:varchar(50);index"` // GitHub ID
	Preferences string `json:"preferences" gorm:"type:json"`            // 用户偏好（JSON）

	// 时间戳
	LastLoginAt *time.Time `json:"last_login_at"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

// BeforeCreate 钩子 - 在创建记录前自动调用
func (u *User) BeforeCreate(tx *gorm.DB) error {
	// 确保 Preferences 是有效的 JSON
	if !json.Valid([]byte(u.Preferences)) || u.Preferences == "" {
		u.Preferences = "{}" // 设置为空 JSON 对象
	}
	return nil
}

// BeforeUpdate 钩子 - 在更新记录前自动调用
func (u *User) BeforeUpdate(tx *gorm.DB) error {
	// 确保 Preferences 是有效的 JSON
	if !json.Valid([]byte(u.Preferences)) || u.Preferences == "" {
		u.Preferences = "{}"
	}
	return nil
}

// Validate 验证用户实体的业务规则
// 参考原项目实现，确保数据完整性和业务规则
func (u *User) Validate() error {
	// 验证用户名
	if u.Username == "" {
		return ErrInvalidUser
	}
	if len(u.Username) < 3 || len(u.Username) > 50 {
		return ErrInvalidUserName
	}

	// 验证邮箱
	if !u.IsValidEmail() {
		return ErrInvalidUserEmail
	}

	// 密码验证（OAuth用户可以没有密码）
	if u.Password == "" && u.GitHubID == "" {
		return ErrInvalidUserPassword
	}

	// 设置默认值
	if u.Name == "" {
		u.Name = u.Username
	}
	// 预留字段，未使用
	if u.Role == "" {
		u.Role = "user"
	}

	return nil
}

// IsValidEmail 验证邮箱格式
func (u *User) IsValidEmail() bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(u.Email)
}

// IsActive 检查用户是否激活（使用新的IsActive字段）
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}

// CanLogin 检查用户是否可以登录
func (u *User) CanLogin() bool {
	return u.Status == UserStatusActive
}

// Activate 激活用户
func (u *User) Activate() {
	u.Status = UserStatusActive
	u.UpdatedAt = time.Now()
}

// Suspend 暂停用户
func (u *User) Suspend() {
	u.Status = UserStatusSuspended
	u.UpdatedAt = time.Now()
}

// UpdateLastLogin 更新最后登录时间
func (u *User) UpdateLastLogin() {
	now := time.Now()
	u.LastLoginAt = &now
	u.UpdatedAt = now
}

// UserRepository 定义用户数据访问接口
// 这是领域层定义的接口，具体实现在基础设施层
type UserRepository interface {
	// 基础CRUD操作
	GetByID(ctx context.Context, id int64) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	GetByGithubId(ctx context.Context, id string) (*User, error)

	Store(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id int64) error

	// 查询操作
	List(ctx context.Context, offset, limit int) ([]*User, error)
	Search(ctx context.Context, query string, offset, limit int) ([]*User, error)
	Count(ctx context.Context) (int64, error)
	CountByStatus(ctx context.Context, status UserStatus) (int64, error)

	// 批量操作
	BatchUpdateStatus(ctx context.Context, userIDs []int64, status UserStatus) error
}

// UserUsecase 用户业务逻辑接口
// 专注于用户资料和信息管理，不包含认证逻辑
type UserUsecase interface {
	// 用户资料管理
	GetProfile(ctx context.Context, userID int64) (*User, error)
	UpdateProfile(ctx context.Context, userID int64, updates *UserProfileUpdate) error
	ChangePassword(ctx context.Context, userID int64, oldPassword, newPassword string) error

	// 用户查询
	GetUserByID(ctx context.Context, userID int64) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	GetUserByGitHubId(ctx context.Context, query string) (*User, error)
	SearchUsers(ctx context.Context, query string, limit, offset int) ([]*User, int64, error)
}

// UserProfileUpdate 用户资料更新结构
type UserProfileUpdate struct {
	Name       *string `json:"name,omitempty"`
	Bio        *string `json:"bio,omitempty"`
	Company    *string `json:"company,omitempty"`
	Location   *string `json:"location,omitempty"`
	WebsiteURL *string `json:"website_url,omitempty"`
	AvatarURL  *string `json:"avatar_url,omitempty"`
}

// UserCacheRepository 定义用户缓存数据访问接口
// 专注于用户信息缓存，不包含认证相关缓存
type UserCacheRepository interface {
	// 用户信息缓存
	CacheUser(ctx context.Context, user *User) error
	GetCachedUser(ctx context.Context, key string) (*User, error)
	DeleteCachedUser(ctx context.Context, key string) error

	//// 用户搜索结果缓存（可选优化）
	//CacheSearchResult(ctx context.Context, query string, users []*User, expiration time.Duration) error
	//GetCachedSearchResult(ctx context.Context, query string) ([]*User, error)
	//DeleteCachedSearchResult(ctx context.Context, query string) error
}
