package mysql

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"DOC/domain"
)

// mysqlUserRepository MySQL 用户仓储实现
// 实现 domain.UserRepository 接口，负责用户数据的持久化操作
// 使用 GORM 作为 ORM 框架，简化数据库操作
type mysqlUserRepository struct {
	db *gorm.DB // GORM 数据库连接实例
}

// NewMysqlUserRepository 创建新的 MySQL 用户仓储实例
// 参数说明：
// - db: GORM 数据库连接实例
// 返回值：
// - domain.UserRepository: 用户仓储接口实现
func NewMysqlUserRepository(db *gorm.DB) domain.UserRepository {
	return &mysqlUserRepository{
		db: db,
	}
}

// GetByID 根据用户ID获取用户信息
// 这是最常用的查询方法，用于根据用户ID获取完整的用户信息
// 参数说明：
// - ctx: 上下文，用于超时控制和取消操作
// - id: 用户唯一标识ID
// 返回值：
// - *domain.User: 用户实体，如果未找到则为 nil
// - error: 查询过程中的错误
func (r *mysqlUserRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	var user domain.User

	// 使用 GORM 的 WithContext 方法传递上下文
	// First 方法查找第一条匹配的记录
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&user).Error

	if err != nil {
		// 如果是记录未找到的错误，返回领域层定义的错误
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUserNotFound
		}
		// 其他数据库错误直接返回
		return nil, err
	}

	return &user, nil
}

// GetByEmail 根据邮箱获取用户信息
// 用于登录验证和邮箱唯一性检查
// 邮箱是用户的唯一标识之一，常用于身份验证
// 参数说明：
// - ctx: 上下文
// - email: 用户邮箱地址
// 返回值：
// - *domain.User: 用户实体
// - error: 查询错误
func (r *mysqlUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User

	err := r.db.WithContext(ctx).
		Where("email = ?", email).
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

// GetByUsername 根据用户名获取用户信息
// 用于用户名唯一性检查和某些登录场景
// 参数说明：
// - ctx: 上下文
// - username: 用户名
// 返回值：
// - *domain.User: 用户实体
// - error: 查询错误
func (r *mysqlUserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	var user domain.User

	err := r.db.WithContext(ctx).
		Where("username = ?", username).
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *mysqlUserRepository) GetByGithubId(ctx context.Context, githubId string) (*domain.User, error) {
	var user domain.User

	err := r.db.WithContext(ctx).
		Where("github_id = ?", githubId).
		First(user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

// Store 创建新用户
// 用于用户注册，将新用户信息保存到数据库
// 注意：GORM 会自动设置创建时间和更新时间（如果字段存在）
// 参数说明：
// - ctx: 上下文
// - user: 待保存的用户实体
// 返回值：
// - error: 保存过程中的错误
func (r *mysqlUserRepository) Store(ctx context.Context, user *domain.User) error {
	// 设置创建和更新时间
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	// 使用 GORM 的 Create 方法插入新记录
	// Create 方法会自动填充主键ID到 user 结构体中
	err := r.db.WithContext(ctx).Create(user).Error
	if err != nil {
		return err
	}

	return nil
}

// Update 更新用户信息
// 用于用户资料修改，只更新非零值字段
// GORM 的 Updates 方法会忽略零值字段，避免意外覆盖
// 参数说明：
// - ctx: 上下文
// - user: 包含更新信息的用户实体
// 返回值：
// - error: 更新过程中的错误
func (r *mysqlUserRepository) Update(ctx context.Context, user *domain.User) error {
	// 设置更新时间
	user.UpdatedAt = time.Now()

	// 使用 Updates 方法更新记录
	// Where 条件确保只更新指定ID的用户
	result := r.db.WithContext(ctx).
		Where("id = ?", user.ID).
		Updates(user)

	if result.Error != nil {
		return result.Error
	}

	// 检查是否有记录被更新
	// 如果 RowsAffected 为 0，说明用户不存在
	if result.RowsAffected == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

// Delete 删除用户（软删除）
// 在实际业务中，通常不会物理删除用户数据
// 而是标记为删除状态，保留数据用于审计和恢复
// 参数说明：
// - ctx: 上下文
// - id: 待删除的用户ID
// 返回值：
// - error: 删除过程中的错误
func (r *mysqlUserRepository) Delete(ctx context.Context, id int64) error {
	// 使用 GORM 的软删除功能
	// 如果模型包含 DeletedAt 字段，Delete 方法会执行软删除
	result := r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&domain.User{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

// List 获取用户列表（分页）
// 用于管理后台的用户列表展示
// 支持分页查询，避免一次性加载大量数据
// 参数说明：
// - ctx: 上下文
// - offset: 偏移量，跳过的记录数
// - limit: 限制数量，返回的最大记录数
// 返回值：
// - []*domain.User: 用户列表
// - error: 查询错误
func (r *mysqlUserRepository) List(ctx context.Context, offset, limit int) ([]*domain.User, error) {
	var users []*domain.User

	// 使用 Offset 和 Limit 实现分页
	// Order 确保结果的一致性
	err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&users).Error

	if err != nil {
		return nil, err
	}

	return users, nil
}

// Count 获取用户总数
// 用于分页计算和统计展示
// 返回值：
// - int64: 用户总数
// - error: 查询错误
func (r *mysqlUserRepository) Count(ctx context.Context) (int64, error) {
	var count int64

	// 使用 Count 方法统计记录数
	err := r.db.WithContext(ctx).
		Model(&domain.User{}).
		Count(&count).Error

	if err != nil {
		return 0, err
	}

	return count, nil
}

// Search 搜索用户
// 根据查询条件搜索用户，支持用户名、邮箱、姓名模糊搜索
func (r *mysqlUserRepository) Search(ctx context.Context, query string, offset, limit int) ([]*domain.User, error) {
	var users []*domain.User

	db := r.db.WithContext(ctx)

	// 如果查询条件不为空，添加搜索条件
	if query != "" {
		searchPattern := "%" + query + "%"
		db = db.Where("username LIKE ? OR email LIKE ? OR name LIKE ?",
			searchPattern, searchPattern, searchPattern)
	}

	err := db.Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&users).Error

	if err != nil {
		return nil, err
	}

	return users, nil
}

// CountByStatus 根据状态统计用户数量
func (r *mysqlUserRepository) CountByStatus(ctx context.Context, status domain.UserStatus) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.User{}).
		Where("status = ?", status).
		Count(&count).Error
	return count, err
}

// BatchUpdateStatus 批量更新用户状态
func (r *mysqlUserRepository) BatchUpdateStatus(ctx context.Context, userIDs []int64, status domain.UserStatus) error {
	return r.db.WithContext(ctx).
		Model(&domain.User{}).
		Where("id IN ?", userIDs).
		Update("status", status).Error
}
