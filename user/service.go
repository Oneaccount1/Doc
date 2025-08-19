package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"DOC/domain"
	"DOC/pkg/utils"
)

// userService 用户业务服务实现
// 专注于用户资料管理、用户查询和权限检查
// 不包含认证相关逻辑（已分离到auth服务）
type userService struct {
	userRepo       domain.UserRepository      // 用户仓储接口
	userCache      domain.UserCacheRepository // 用户缓存仓储接口
	contextTimeout time.Duration              // 上下文超时时间
}

// NewUserService 创建新的用户服务实例
func NewUserService(
	userRepo domain.UserRepository,
	userCache domain.UserCacheRepository,
	timeout time.Duration,
) domain.UserUsecase {
	return &userService{
		userRepo:       userRepo,
		userCache:      userCache,
		contextTimeout: timeout,
	}
}

// GetProfile 获取用户资料
func (s *userService) GetProfile(ctx context.Context, userID int64) (*domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	if userID <= 0 {
		return nil, domain.ErrInvalidUser
	}

	// 先尝试从缓存获取
	cacheKey := fmt.Sprintf("user:%d", userID)
	if cachedUser, err := s.userCache.GetCachedUser(ctx, cacheKey); err == nil && cachedUser != nil {
		return cachedUser, nil
	}

	// 从数据库获取
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 缓存用户信息
	if err := s.userCache.CacheUser(ctx, user); err != nil {
		// 记录日志但不影响主流程
		fmt.Printf("缓存用户信息失败: %v\n", err)
	}

	return user, nil
}

// UpdateProfile 更新用户资料
func (s *userService) UpdateProfile(ctx context.Context, userID int64, updates *domain.UserProfileUpdate) error {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	if userID <= 0 {
		return domain.ErrInvalidUser
	}

	if updates == nil {
		return domain.ErrBadParamInput
	}

	// 获取当前用户信息
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	// 应用更新
	if updates.Name != nil {
		user.Name = *updates.Name
	}
	if updates.Bio != nil {
		user.Bio = *updates.Bio
	}
	if updates.Company != nil {
		user.Company = *updates.Company
	}
	if updates.Location != nil {
		user.Location = *updates.Location
	}
	if updates.WebsiteURL != nil {
		user.WebsiteURL = *updates.WebsiteURL
	}
	if updates.AvatarURL != nil {
		user.AvatarURL = *updates.AvatarURL
	}

	// 保存到数据库
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("更新用户资料失败: %w", err)
	}

	// 清除缓存
	cacheKey := fmt.Sprintf("user:%d", userID)
	if err := s.userCache.DeleteCachedUser(ctx, cacheKey); err != nil {
		fmt.Printf("清除用户缓存失败: %v\n", err)
	}

	return nil
}

// ChangePassword 修改密码
func (s *userService) ChangePassword(ctx context.Context, userID int64, oldPassword, newPassword string) error {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	if userID <= 0 {
		return domain.ErrInvalidUser
	}

	if oldPassword == "" || newPassword == "" {
		return domain.ErrInvalidUserPassword
	}

	// 获取用户信息
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	// 验证旧密码
	if !utils.CheckPassword(oldPassword, user.Password) {
		return domain.ErrInvalidUserPassword
	}

	// 生成新密码哈希
	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}

	// 更新密码
	user.Password = hashedPassword

	// 保存到数据库
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("更新密码失败: %w", err)
	}

	// 清除缓存
	cacheKey := fmt.Sprintf("user:%d", userID)
	if err := s.userCache.DeleteCachedUser(ctx, cacheKey); err != nil {
		fmt.Printf("清除用户缓存失败: %v\n", err)
	}

	return nil
}

// GetUserByID 根据ID获取用户
func (s *userService) GetUserByID(ctx context.Context, userID int64) (*domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	if userID <= 0 {
		return nil, domain.ErrInvalidUser
	}
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, domain.ErrInternalServerError
	}
	return user, nil
}

// GetUserByEmail 根据邮箱获取用户
func (s *userService) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	if email == "" || !utils.IsValidEmail(email) {
		return nil, domain.ErrInvalidUserEmail
	}

	return s.userRepo.GetByEmail(ctx, email)
}

// GetUserByUsername 根据用户名获取用户
func (s *userService) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	if username == "" {
		return nil, domain.ErrInvalidUserName
	}

	return s.userRepo.GetByUsername(ctx, username)
}

func (s *userService) GetUserByGitHubId(ctx context.Context, githubId string) (*domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	if githubId == "" {
		return nil, domain.ErrInvalidGithubId
	}

	return s.userRepo.GetByGithubId(ctx, githubId)
}

// SearchUsers 搜索用户
func (s *userService) SearchUsers(ctx context.Context, query string, limit, offset int) ([]*domain.User, int64, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	if query == "" {
		return nil, 0, domain.ErrBadParamInput
	}

	if limit <= 0 || limit > 100 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	// 搜索用户
	users, err := s.userRepo.Search(ctx, query, offset, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("搜索用户失败: %w", err)
	}

	// 获取总数
	total := int64(len(users))

	return users, total, nil
}
