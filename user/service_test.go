package user_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"refatorSiwu/domain"
	"refatorSiwu/pkg/jwt"
	"refatorSiwu/user"
	"refatorSiwu/user/mocks"
)

// TestUserService_Register 测试用户注册功能
// 这个测试验证用户注册的完整流程，包括数据验证、唯一性检查、密码加密等
func TestUserService_Register(t *testing.T) {
	// 创建 mock 仓储
	mockUserRepo := new(mocks.UserRepository)
	mockUserCache := new(mocks.UserCacheRepository)
	mockEmailService := new(mocks.EmailUsecase)

	// 创建 JWT 管理器
	jwtManager := jwt.NewJWTManager("test-secret", time.Hour)

	// 创建用户服务
	userService := user.NewUserService(mockUserRepo, mockUserCache, mockEmailService, jwtManager, time.Second*30)

	t.Run("成功注册", func(t *testing.T) {
		// 准备测试数据
		testUser := &domain.User{
			Username:    "testuser",
			Email:       "test@example.com",
			Password:    "password123",
			DisplayName: "Test User",
		}

		// 设置 mock 期望
		// 1. 检查邮箱唯一性 - 返回用户不存在
		mockUserRepo.On("GetByEmail", mock.Anything, testUser.Email).
			Return(nil, domain.ErrUserNotFound).Once()

		// 2. 检查用户名唯一性 - 返回用户不存在
		mockUserRepo.On("GetByUsername", mock.Anything, testUser.Username).
			Return(nil, domain.ErrUserNotFound).Once()

		// 3. 保存用户 - 返回成功
		mockUserRepo.On("Store", mock.Anything, mock.AnythingOfType("*domain.User")).
			Return(nil).Once()

		// 执行测试
		err := userService.Register(context.Background(), testUser)

		// 验证结果
		assert.NoError(t, err)
		assert.Equal(t, domain.UserStatusActive, testUser.Status)
		assert.NotEmpty(t, testUser.Password)                // 密码应该被加密
		assert.NotEqual(t, "password123", testUser.Password) // 密码不应该是明文

		// 验证 mock 调用
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("邮箱已存在", func(t *testing.T) {
		testUser := &domain.User{
			Username: "testuser2",
			Email:    "existing@example.com",
			Password: "password123",
		}

		// 设置 mock 期望 - 邮箱已存在
		existingUser := &domain.User{ID: 1, Email: "existing@example.com"}
		mockUserRepo.On("GetByEmail", mock.Anything, testUser.Email).
			Return(existingUser, nil).Once()

		// 执行测试
		err := userService.Register(context.Background(), testUser)

		// 验证结果
		assert.Error(t, err)
		assert.Equal(t, domain.ErrUserAlreadyExist, err)

		mockUserRepo.AssertExpectations(t)
	})

	t.Run("用户名已存在", func(t *testing.T) {
		testUser := &domain.User{
			Username: "existinguser",
			Email:    "new@example.com",
			Password: "password123",
		}

		// 设置 mock 期望
		// 1. 邮箱检查 - 不存在
		mockUserRepo.On("GetByEmail", mock.Anything, testUser.Email).
			Return(nil, domain.ErrUserNotFound).Once()

		// 2. 用户名检查 - 已存在
		existingUser := &domain.User{ID: 1, Username: "existinguser"}
		mockUserRepo.On("GetByUsername", mock.Anything, testUser.Username).
			Return(existingUser, nil).Once()

		// 执行测试
		err := userService.Register(context.Background(), testUser)

		// 验证结果
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "用户名已存在")

		mockUserRepo.AssertExpectations(t)
	})

	t.Run("密码强度不足", func(t *testing.T) {
		testUser := &domain.User{
			Username: "testuser3",
			Email:    "test3@example.com",
			Password: "123", // 密码太短
		}

		// 执行测试
		err := userService.Register(context.Background(), testUser)

		// 验证结果
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "密码强度不足")
	})
}

// TestUserService_Login 测试用户登录功能
func TestUserService_Login(t *testing.T) {
	mockUserRepo := new(mocks.UserRepository)
	mockUserCache := new(mocks.UserCacheRepository)
	mockEmailService := new(mocks.EmailUsecase)
	jwtManager := jwt.NewJWTManager("test-secret", time.Hour)
	userService := user.NewUserService(mockUserRepo, mockUserCache, mockEmailService, jwtManager, time.Second*30)

	t.Run("成功登录", func(t *testing.T) {
		// 准备测试数据
		email := "test@example.com"

		// 创建已加密密码的用户
		hashedPassword := "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi" // "password"
		existingUser := &domain.User{
			ID:       1,
			Username: "testuser",
			Email:    email,
			Password: hashedPassword,
			Status:   domain.UserStatusActive,
		}

		// 设置 mock 期望
		mockUserRepo.On("GetByEmail", mock.Anything, email).
			Return(existingUser, nil).Once()

		// 更新最后登录时间
		mockUserRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.User")).
			Return(nil).Once()

		// 执行测试
		user, token, err := userService.Login(context.Background(), email, "password")

		// 验证结果
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.NotEmpty(t, token)
		assert.Equal(t, existingUser.ID, user.ID)
		assert.Equal(t, existingUser.Email, user.Email)
		assert.Empty(t, user.Password) // 密码应该被清除

		// 验证 token 有效性
		claims, err := jwtManager.ValidateToken(token)
		assert.NoError(t, err)
		assert.Equal(t, existingUser.ID, claims.UserID)
		assert.Equal(t, existingUser.Username, claims.Username)
		assert.Equal(t, existingUser.Email, claims.Email)

		mockUserRepo.AssertExpectations(t)
	})

	t.Run("用户不存在", func(t *testing.T) {
		email := "nonexistent@example.com"
		password := "password123"

		// 设置 mock 期望
		mockUserRepo.On("GetByEmail", mock.Anything, email).
			Return(nil, domain.ErrUserNotFound).Once()

		// 执行测试
		user, token, err := userService.Login(context.Background(), email, password)

		// 验证结果
		assert.Error(t, err)
		assert.Equal(t, domain.ErrInvalidCredentials, err)
		assert.Nil(t, user)
		assert.Empty(t, token)

		mockUserRepo.AssertExpectations(t)
	})

	t.Run("密码错误", func(t *testing.T) {
		email := "test@example.com"
		wrongPassword := "wrongpassword"

		existingUser := &domain.User{
			ID:       1,
			Email:    email,
			Password: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // "password"
			Status:   domain.UserStatusActive,
		}

		// 设置 mock 期望
		mockUserRepo.On("GetByEmail", mock.Anything, email).
			Return(existingUser, nil).Once()

		// 执行测试
		user, token, err := userService.Login(context.Background(), email, wrongPassword)

		// 验证结果
		assert.Error(t, err)
		assert.Equal(t, domain.ErrInvalidCredentials, err)
		assert.Nil(t, user)
		assert.Empty(t, token)

		mockUserRepo.AssertExpectations(t)
	})

	t.Run("用户被暂停", func(t *testing.T) {
		email := "suspended@example.com"
		password := "password"

		suspendedUser := &domain.User{
			ID:       1,
			Email:    email,
			Password: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi",
			Status:   domain.UserStatusSuspended, // 用户被暂停
		}

		// 设置 mock 期望
		mockUserRepo.On("GetByEmail", mock.Anything, email).
			Return(suspendedUser, nil).Once()

		// 执行测试
		user, token, err := userService.Login(context.Background(), email, password)

		// 验证结果
		assert.Error(t, err)
		assert.Equal(t, domain.ErrUserNotActive, err)
		assert.Nil(t, user)
		assert.Empty(t, token)

		mockUserRepo.AssertExpectations(t)
	})
}

// TestUserService_GetProfile 测试获取用户资料功能
func TestUserService_GetProfile(t *testing.T) {
	mockUserRepo := new(mocks.UserRepository)
	mockUserCache := new(mocks.UserCacheRepository)
	mockEmailService := new(mocks.EmailUsecase)
	jwtManager := jwt.NewJWTManager("test-secret", time.Hour)
	userService := user.NewUserService(mockUserRepo, mockUserCache, mockEmailService, jwtManager, time.Second*30)

	t.Run("成功获取用户资料", func(t *testing.T) {
		userID := int64(1)
		expectedUser := &domain.User{
			ID:          userID,
			Username:    "testuser",
			Email:       "test@example.com",
			DisplayName: "Test User",
			Password:    "hashedpassword", // 这个应该被清除
		}

		// 设置 mock 期望
		mockUserRepo.On("GetByID", mock.Anything, userID).
			Return(expectedUser, nil).Once()

		// 执行测试
		user, err := userService.GetProfile(context.Background(), userID)

		// 验证结果
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, expectedUser.ID, user.ID)
		assert.Equal(t, expectedUser.Username, user.Username)
		assert.Equal(t, expectedUser.Email, user.Email)
		assert.Empty(t, user.Password) // 密码应该被清除

		mockUserRepo.AssertExpectations(t)
	})

	t.Run("用户不存在", func(t *testing.T) {
		userID := int64(999)

		// 设置 mock 期望
		mockUserRepo.On("GetByID", mock.Anything, userID).
			Return(nil, domain.ErrUserNotFound).Once()

		// 执行测试
		user, err := userService.GetProfile(context.Background(), userID)

		// 验证结果
		assert.Error(t, err)
		assert.Equal(t, domain.ErrUserNotFound, err)
		assert.Nil(t, user)

		mockUserRepo.AssertExpectations(t)
	})
}
