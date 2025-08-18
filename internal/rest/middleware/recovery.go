package middleware

import (
	"fmt"
	"log"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

// CustomRecoveryMiddleware 自定义恢复中间件
func CustomRecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		// 记录错误日志
		log.Printf("Panic recovered: %v\n", recovered)
		log.Printf("Stack trace:\n%s", debug.Stack())

		// 返回统一的错误响应
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": "An unexpected error occurred",
		})

		c.Abort()
	})
}

// RecoveryConfig 恢复中间件配置
type RecoveryConfig struct {
	// EnableStackTrace 是否启用堆栈跟踪
	EnableStackTrace bool

	// LogFunc 自定义日志函数
	LogFunc func(c *gin.Context, err interface{}, stack []byte)

	// RecoveryFunc 自定义恢复处理函数
	RecoveryFunc func(c *gin.Context, err interface{})
}

// DefaultRecoveryConfig 默认恢复配置
func DefaultRecoveryConfig() RecoveryConfig {
	return RecoveryConfig{
		EnableStackTrace: true,
		LogFunc: func(c *gin.Context, err interface{}, stack []byte) {
			log.Printf("[Recovery] panic recovered:\n%v\n%s", err, stack)
		},
		RecoveryFunc: func(c *gin.Context, err interface{}) {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Internal Server Error",
				"message": "An unexpected error occurred",
			})
			c.Abort()
		},
	}
}

// RecoveryWithConfig 带配置的恢复中间件
func RecoveryWithConfig(config RecoveryConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 获取堆栈跟踪
				var stack []byte
				if config.EnableStackTrace {
					stack = debug.Stack()
				}

				// 调用自定义日志函数
				if config.LogFunc != nil {
					config.LogFunc(c, err, stack)
				}

				// 调用自定义恢复处理函数
				if config.RecoveryFunc != nil {
					config.RecoveryFunc(c, err)
				} else {
					// 默认处理
					c.JSON(http.StatusInternalServerError, gin.H{
						"error":   "Internal Server Error",
						"message": "An unexpected error occurred",
					})
					c.Abort()
				}
			}
		}()

		c.Next()
	}
}

// DebugRecoveryMiddleware 调试模式的恢复中间件（包含详细错误信息）
func DebugRecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		// 记录详细错误信息
		stack := debug.Stack()
		log.Printf("Panic recovered: %v\n", recovered)
		log.Printf("Stack trace:\n%s", stack)

		// 在调试模式下返回详细错误信息
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":       "Internal Server Error",
			"message":     "An unexpected error occurred",
			"debug_info":  fmt.Sprintf("%v", recovered),
			"stack_trace": string(stack),
		})

		c.Abort()
	})
}
