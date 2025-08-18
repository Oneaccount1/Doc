package middleware

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

// LoggerConfig 日志中间件配置
type LoggerConfig struct {
	// Output 日志输出目标，默认为 os.Stdout
	Output io.Writer

	// SkipPaths 跳过记录日志的路径
	SkipPaths []string

	// TimeFormat 时间格式，默认为 RFC3339
	TimeFormat string

	// UTC 是否使用 UTC 时间
	UTC bool
}

// DefaultLoggerConfig 默认日志配置
func DefaultLoggerConfig() LoggerConfig {
	return LoggerConfig{
		Output:     os.Stdout,
		SkipPaths:  []string{"/health", "/metrics"},
		TimeFormat: time.RFC3339,
		UTC:        false,
	}
}

// LoggerMiddleware 日志中间件
func LoggerMiddleware() gin.HandlerFunc {
	return LoggerWithConfig(DefaultLoggerConfig())
}

// LoggerWithConfig 带配置的日志中间件
func LoggerWithConfig(config LoggerConfig) gin.HandlerFunc {
	// 设置默认值
	if config.Output == nil {
		config.Output = os.Stdout
	}
	if config.TimeFormat == "" {
		config.TimeFormat = time.RFC3339
	}

	// 创建跳过路径的映射
	skipPaths := make(map[string]bool)
	for _, path := range config.SkipPaths {
		skipPaths[path] = true
	}

	return gin.LoggerWithConfig(gin.LoggerConfig{
		Output:    config.Output,
		SkipPaths: config.SkipPaths,
		Formatter: func(param gin.LogFormatterParams) string {
			var timeStamp time.Time
			if config.UTC {
				timeStamp = param.TimeStamp.UTC()
			} else {
				timeStamp = param.TimeStamp
			}

			// 自定义日志格式
			return fmt.Sprintf("[%s] %s %s %s %d %s \"%s\" %s \"%s\" %s\n",
				timeStamp.Format(config.TimeFormat),
				param.ClientIP,
				param.Method,
				param.Path,
				param.StatusCode,
				param.Latency,
				param.Request.UserAgent(),
				param.ErrorMessage,
				param.Request.Referer(),
				param.Keys,
			)
		},
	})
}

// CustomLoggerMiddleware 自定义日志中间件
func CustomLoggerMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// 自定义日志格式
		return fmt.Sprintf("%s - [%s] \"%s %s %s\" %d %s \"%s\" \"%s\"\n",
			param.ClientIP,
			param.TimeStamp.Format("02/Jan/2006:15:04:05 -0700"),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}

// StructuredLoggerMiddleware 结构化日志中间件
func StructuredLoggerMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// JSON 格式的结构化日志
		return fmt.Sprintf(`{"time":"%s","client_ip":"%s","method":"%s","path":"%s","status":%d,"latency":"%s","user_agent":"%s","error":"%s"}%s`,
			param.TimeStamp.Format(time.RFC3339),
			param.ClientIP,
			param.Method,
			param.Path,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
			"\n",
		)
	})
}
