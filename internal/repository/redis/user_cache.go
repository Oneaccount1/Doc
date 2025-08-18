package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"DOC/domain"
	"github.com/redis/go-redis/v9"
)

// 用户相关键前缀
const (
	UserCachePrefix  = "user_cache:"  // 用户信息缓存
	TokenCachePrefix = "token_cache:" // Token缓存
)

// 时间常量
const (
	UserCacheExpire  = 24 * time.Hour // 用户缓存24小时
	TokenCacheExpire = 24 * time.Hour // Token缓存24小时
)

// UserCacheRepository  Redis缓存仓储实现
// 实现domain.CacheService接口，提供验证码存储、频率限制等功能
type UserCacheRepository struct {
	client *redis.Client // 保留原始客户端用于SetNX等高级操作
}

// NewUserCacheRepository 创建新的缓存仓储实例
func NewUserCacheRepository(client *redis.Client) domain.UserCacheRepository {
	return &UserCacheRepository{
		client: client,
	}
}
func (r *UserCacheRepository) CacheUser(ctx context.Context, user *domain.User) error {
	key := r.generateUserCacheKey(user.Email)
	// 序列化对象
	userData, err := json.Marshal(user)
	if err != nil {
		return domain.ErrMarsh
	}
	// 存储用户信息, 同时设置用户信息过期时间
	return r.client.Set(ctx, key, userData, UserCacheExpire).Err()
}

func (r *UserCacheRepository) GetCachedUser(ctx context.Context, key string) (*domain.User, error) {
	result, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	// 反序列化用户对象
	var user domain.User

	if err := json.Unmarshal([]byte(result), &user); err != nil {
		return nil, domain.ErrUnMarsh
	}

	return &user, nil
}

func (r *UserCacheRepository) DeleteCachedUser(ctx context.Context, key string) error {
	deleted, err := r.client.Del(ctx, key).Result()
	if err != nil {
		return err
	}
	if deleted == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}
func (r *UserCacheRepository) generateUserCacheKey(email string) string {
	return fmt.Sprintf("%s%s", UserCachePrefix, email)
}
