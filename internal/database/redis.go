package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"

	"DOC/config"
)

// NewRedisConnection 创建 Redis 连接
func NewRedisConnection(cfg *config.RedisConfig) (*redis.Client, error) {
	// 创建 Redis 客户端
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.GetRedisAddr(),
		Password: cfg.Password,
		DB:       cfg.DB,

		// 连接池配置
		PoolSize:     10,               // 连接池大小
		MinIdleConns: 5,                // 最小空闲连接数
		MaxIdleConns: 10,               // 最大空闲连接数
		PoolTimeout:  30 * time.Second, // 连接池超时时间

		// 重试配置
		MaxRetries:      3,
		MinRetryBackoff: 8 * time.Millisecond,
		MaxRetryBackoff: 512 * time.Millisecond,

		// 读写超时
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %v", err)
	}

	log.Println("Successfully connected to Redis")
	return rdb, nil
}
