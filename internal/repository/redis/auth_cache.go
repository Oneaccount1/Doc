package redis

import (
	"DOC/domain"
	"DOC/pkg/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// 数据库底层错误直接返回

// 验证码相关键前缀
const (
	VerificationCodePrefix = "email_code:"           // 验证码存储
	SendTimePrefix         = "email_code_send_time:" // 发送时间限制
	SessionCachePrefix     = "session:"              // 会话缓存
	OAuthStateCachePrefix  = "oauth_state:"          // OAuth状态缓存
)

// 时间常量
const (
	VerificationCodeExpire = 10 * time.Minute // 验证码10分钟过期
	SendIntervalLimit      = 1 * time.Minute  // 发送间隔1分钟
	SessionCacheExpire     = 24 * time.Hour   // 会话缓存24小时
	OAuthStateExpire       = 10 * time.Minute // OAuth状态10分钟过期
)

type AuthCacheRepository struct {
	client *redis.Client
}

func NewAuthCacheRepository(client *redis.Client) domain.AuthCacheRepository {
	return &AuthCacheRepository{
		client: client,
	}
}

func (a *AuthCacheRepository) StoreEmailCode(ctx context.Context, email, code string, expiration time.Duration) error {
	// 1.首先生成 email 加密后的键
	key := generateEmailCodeKey(email)
	// 使用传入的过期时间，如果为0则使用默认值
	if expiration == 0 {
		expiration = VerificationCodeExpire
	}
	return a.client.Set(ctx, key, code, expiration).Err()
}

func (a *AuthCacheRepository) GetEmailCode(ctx context.Context, email string) (string, error) {
	// 1.首先生成 email 加密后的键
	key := generateEmailCodeKey(email)
	result, err := a.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", domain.ErrVerificationCodeNotFound
		}
		return "", err

	}
	return result, nil
}

func (a *AuthCacheRepository) DeleteEmailCode(ctx context.Context, email string) error {
	// 1.首先生成 email 加密后的键
	key := generateEmailCodeKey(email)
	deleted, err := a.client.Del(ctx, key).Result()
	if err != nil {
		return err
	}
	if deleted == 0 {
		return domain.ErrVerificationCodeNotFound
	}
	return nil
}

func (a *AuthCacheRepository) CheckSendInterval(ctx context.Context, email string) (bool, error) {
	// 1.首先生成发送时间限制的键
	key := generateSendTimeKey(email)
	canSet, err := a.client.SetNX(ctx, key, 1, SendIntervalLimit).Result()
	if err != nil {
		return false, err
	}
	return canSet, nil
}

func (a *AuthCacheRepository) SetSendInterval(ctx context.Context, email string, interval time.Duration) error {
	// 1.首先生成发送时间限制的键
	key := generateSendTimeKey(email)
	return a.client.Set(ctx, key, 1, interval).Err()
}

func (a *AuthCacheRepository) CacheSession(ctx context.Context, session *domain.AuthSession) error {
	// 生成会话缓存键
	key := generateSessionCacheKey(session.SessionToken)

	// 序列化会话对象
	sessionData, err := json.Marshal(session)
	if err != nil {
		return err
	}

	// 存储会话信息，设置过期时间
	return a.client.Set(ctx, key, sessionData, SessionCacheExpire).Err()
}

func (a *AuthCacheRepository) GetCachedSession(ctx context.Context, token string) (*domain.AuthSession, error) {
	// 生成会话缓存键
	key := generateSessionCacheKey(token)

	// 从Redis获取会话数据
	result, err := a.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, domain.ErrSessionNotFound
		}
		return nil, err
	}

	// 反序列化会话对象
	var session domain.AuthSession
	if err := json.Unmarshal([]byte(result), &session); err != nil {
		return nil, domain.ErrUnMarsh
	}

	return &session, nil
}

func (a *AuthCacheRepository) DeleteCachedSession(ctx context.Context, token string) error {
	// 生成会话缓存键
	key := generateSessionCacheKey(token)

	// 删除会话缓存
	deleted, err := a.client.Del(ctx, key).Result()
	if err != nil {
		return err
	}
	if deleted == 0 {
		return domain.ErrSessionNotFound
	}

	return nil
}

func (a *AuthCacheRepository) CacheOAuthState(ctx context.Context, state string, data interface{}, expiration time.Duration) error {
	// 生成OAuth状态缓存键
	key := generateOAuthStateCacheKey(state)

	// 序列化数据
	stateData, err := json.Marshal(data)
	if err != nil {
		return domain.ErrMarsh
	}

	// 存储OAuth状态，使用传入的过期时间，如果为0则使用默认值
	if expiration == 0 {
		expiration = OAuthStateExpire
	}

	return a.client.Set(ctx, key, stateData, expiration).Err()
}

func (a *AuthCacheRepository) GetCachedOAuthState(ctx context.Context, state string) (interface{}, error) {
	// 生成OAuth状态缓存键
	key := generateOAuthStateCacheKey(state)

	// 从Redis获取状态数据
	result, err := a.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, domain.ErrOAuthStateMismatch
		}
		return nil, err
	}

	// 反序列化数据到通用接口
	var data interface{}
	if err := json.Unmarshal([]byte(result), &data); err != nil {
		return nil, domain.ErrUnMarsh
	}

	return data, nil
}

func (a *AuthCacheRepository) DeleteCachedOAuthState(ctx context.Context, state string) error {
	// 生成OAuth状态缓存键
	key := generateOAuthStateCacheKey(state)

	// 删除OAuth状态缓存
	deleted, err := a.client.Del(ctx, key).Result()
	if err != nil {
		return err
	}
	if deleted == 0 {
		return domain.ErrOAuthStateNotFound
	}

	return nil
}

// 新的简化键生成方法
func generateEmailCodeKey(email string) string {
	hashedEmail := utils.MD5Hash(email)
	return fmt.Sprintf("%s%s", VerificationCodePrefix, hashedEmail)
}

func generateSendTimeKey(email string) string {
	hashedEmail := utils.MD5Hash(email)
	return fmt.Sprintf("%s%s", SendTimePrefix, hashedEmail)
}

// generateSessionCacheKey 生成会话缓存键
func generateSessionCacheKey(token string) string {
	hashedToken := utils.MD5Hash(token)
	return fmt.Sprintf("%s%s", SessionCachePrefix, hashedToken)
}

// generateOAuthStateCacheKey 生成OAuth状态缓存键
func generateOAuthStateCacheKey(state string) string {
	hashedState := utils.MD5Hash(state)
	return fmt.Sprintf("%s%s", OAuthStateCachePrefix, hashedState)
}
