package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"DOC/pkg/jwt"
)

// AuthMiddleware JWT 认证中间件
// 负责验证请求中的 JWT token，并将用户信息设置到上下文中
// 这个中间件会拦截需要认证的请求，验证 Authorization header 中的 Bearer token
type AuthMiddleware struct {
	jwtManager *jwt.JWTManager // JWT 管理器，用于验证 token
}

// NewAuthMiddleware 创建新的认证中间件
// 参数说明：
// - jwtManager: JWT 管理器实例
// 返回值：
// - *AuthMiddleware: 认证中间件实例
func NewAuthMiddleware(jwtManager *jwt.JWTManager) *AuthMiddleware {
	return &AuthMiddleware{
		jwtManager: jwtManager,
	}
}

// RequireAuth 需要认证的中间件
// 这是一个 Gin 中间件函数，用于保护需要登录才能访问的路由
// 工作流程：
// 1. 从 Authorization header 中提取 Bearer token
// 2. 验证 token 的有效性
// 3. 解析用户信息并设置到上下文中
// 4. 如果验证失败，返回 401 错误
// 返回值：
// - gin.HandlerFunc: Gin 中间件函数
func (a *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 尝试从多个来源获取 token
		tokenString := a.extractToken(c)
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "缺少认证令牌",
			})
			c.Abort() // 终止请求处理
			return
		}

		// 4. 验证 token 并解析用户信息
		claims, err := a.jwtManager.ValidateToken(tokenString)
		if err != nil {
			// token 验证失败，可能是过期、签名错误或格式错误
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "认证令牌无效或已过期",
			})
			c.Abort()
			return
		}

		// 5. 将用户信息设置到上下文中
		// 后续的处理器可以通过 c.Get() 方法获取这些信息
		c.Set("user_id", claims.UserID)    // 用户ID
		c.Set("username", claims.Username) // 用户名
		c.Set("email", claims.Email)       // 邮箱
		c.Set("token", tokenString)        // 原始 token（用于刷新等操作）

		// 6. 继续处理请求
		c.Next()
	}
}

// OptionalAuth 可选认证中间件
// 这个中间件不会强制要求认证，但如果提供了有效的 token，会解析用户信息
// 适用于一些既可以匿名访问，也可以登录访问的接口
// 例如：公开的文档列表，登录用户可以看到更多信息
func (a *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 尝试从多个来源获取 token
		tokenString := a.extractToken(c)
		if tokenString == "" {
			// 没有提供 token，继续处理请求（匿名访问）
			c.Next()
			return
		}

		// 4. 尝试验证 token
		claims, err := a.jwtManager.ValidateToken(tokenString)
		if err != nil {
			// token 无效，但不阻止请求（匿名访问）
			c.Next()
			return
		}

		// 5. token 有效，设置用户信息到上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("token", tokenString)
		c.Set("authenticated", true) // 标记为已认证

		c.Next()
	}
}

// AdminAuth 管理员认证中间件
// 这个中间件不仅验证用户身份，还检查用户是否具有管理员权限
// 注意：这里简化实现，实际项目中应该从数据库查询用户角色
func (a *AuthMiddleware) AdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 首先进行基本的身份认证
		tokenString := a.extractToken(c)
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "缺少认证令牌",
			})
			c.Abort()
			return
		}

		claims, err := a.jwtManager.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "认证令牌无效或已过期",
			})
			c.Abort()
			return
		}

		// 2. 检查管理员权限
		// 简化实现：假设用户名为 "admin" 的用户是管理员
		// 实际项目中应该查询数据库中的用户角色
		if claims.Username != "admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Forbidden",
				"message": "需要管理员权限",
			})
			c.Abort()
			return
		}

		// 3. 设置用户信息到上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("token", tokenString)
		c.Set("is_admin", true) // 标记为管理员

		c.Next()
	}
}

// RefreshTokenMiddleware 刷新 token 中间件
// 用于处理 token 刷新请求
func (a *AuthMiddleware) RefreshTokenMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 获取当前 token
		tokenString := a.extractToken(c)
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "缺少认证令牌",
			})
			c.Abort()
			return
		}

		// 2. 尝试刷新 token
		newToken, err := a.jwtManager.RefreshToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "令牌刷新失败: " + err.Error(),
			})
			c.Abort()
			return
		}

		// 3. 设置新的token到cookie（24小时过期，使用前端期望的名称：auth_token）
		c.SetCookie(
			"auth_token", // cookie名称（与前端期望一致）
			newToken,     // 新token
			24*60*60,     // 过期时间（24小时）
			"/",          // 路径
			"",           // 域名
			false,        // secure
			false,        // httpOnly（前端需要读取）
		)

		// 4. 返回新的 token，匹配前端期望的格式
		c.JSON(http.StatusOK, gin.H{
			"token":      newToken,
			"expires_in": 24 * 60 * 60, // 24小时，以秒为单位
		})
	}
}

// GetCurrentUserID 从上下文中获取当前用户ID的辅助函数
// 这是一个便捷函数，用于在处理器中快速获取当前用户ID
// 参数说明：
// - c: Gin 上下文
// 返回值：
// - int64: 用户ID，如果未认证返回 0
// - bool: 是否成功获取到用户ID
func GetCurrentUserID(c *gin.Context) (int64, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}

	uid, ok := userID.(int64)
	if !ok {
		return 0, false
	}

	return uid, true
}

// GetCurrentUsername 从上下文中获取当前用户名的辅助函数
func GetCurrentUsername(c *gin.Context) (string, bool) {
	username, exists := c.Get("username")
	if !exists {
		return "", false
	}

	name, ok := username.(string)
	if !ok {
		return "", false
	}

	return name, true
}

// IsAuthenticated 检查当前请求是否已认证的辅助函数
func IsAuthenticated(c *gin.Context) bool {
	_, exists := c.Get("user_id")
	return exists
}

// IsAdmin 检查当前用户是否为管理员的辅助函数
func IsAdmin(c *gin.Context) bool {
	isAdmin, exists := c.Get("is_admin")
	if !exists {
		return false
	}

	admin, ok := isAdmin.(bool)
	return ok && admin
}

// extractToken 从多个来源提取token的辅助方法
// 优先级：Authorization header > Cookie
// 这样既支持传统的Bearer token，也支持cookie认证
func (a *AuthMiddleware) extractToken(c *gin.Context) string {
	// 1. 首先尝试从 Authorization header 获取
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		const bearerPrefix = "Bearer "
		if strings.HasPrefix(authHeader, bearerPrefix) {
			tokenString := authHeader[len(bearerPrefix):]
			if tokenString != "" {
				return tokenString
			}
		}
	}

	// 2. 如果header中没有，尝试从cookie获取（使用前端期望的名称：auth_token）
	if tokenCookie, err := c.Cookie("auth_token"); err == nil && tokenCookie != "" {
		return tokenCookie
	}

	// 3. 都没有找到，返回空字符串
	return ""
}
