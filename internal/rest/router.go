package rest

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"DOC/config"
	"DOC/domain"
	"DOC/internal/rest/middleware"
	"DOC/pkg/jwt"
)

// RouterConfig 路由配置
type RouterConfig struct {
	UserUsecase              domain.UserUsecase
	AuthUsecase              domain.AuthUsecase
	OrganizationUsecase      domain.OrganizationUsecase
	SpaceUsecase             domain.SpaceUsecase
	DocumentAggregateUsecase domain.DocumentAggregateUsecase // 文档聚合服务
	Config                   *config.Config
}

// NewRouter 创建新的路由器
func NewRouter(cfg RouterConfig) *gin.Engine {
	// 设置 Gin 模式
	gin.SetMode(cfg.Config.Server.Mode)

	// 创建 Gin 引擎
	router := gin.New()

	// 添加中间件
	setupMiddlewares(router, cfg.Config)

	// 设置路由
	setupRoutes(router, cfg)

	return router
}

// setupMiddlewares 设置中间件
func setupMiddlewares(router *gin.Engine, cfg *config.Config) {
	// 恢复中间件
	if cfg.Server.Mode == "debug" {
		router.Use(middleware.DebugRecoveryMiddleware())
	} else {
		router.Use(middleware.CustomRecoveryMiddleware())
	}

	// 日志中间件
	if cfg.Server.Mode == "debug" {
		router.Use(middleware.LoggerMiddleware())
	} else {
		router.Use(middleware.StructuredLoggerMiddleware())
	}

	// CORS 中间件
	if cfg.App.EnableCORS {
		router.Use(middleware.CORSMiddleware())
	}

	// 可以在这里添加更多中间件
	// - 限流中间件
	// - 认证中间件
	// - 监控中间件等
}

// setupRoutes 设置路由
func setupRoutes(router *gin.Engine, cfg RouterConfig) {
	// 健康检查端点
	router.GET("/health", healthCheck)
	router.GET("/ping", ping)

	// API 路由分组（添加版本号以匹配前端期望）
	api := router.Group("/api")
	{
		// v1 版本路由（匹配前端期望的路径）
		v1 := api.Group("/v1")
		{
			// 认证相关路由
			if cfg.AuthUsecase != nil {
				setupAuthRoutesV1(v1, cfg.AuthUsecase, cfg.UserUsecase, cfg.Config)
			}

			// 用户管理相关路由
			if cfg.UserUsecase != nil {
				setupUserRoutesV1(v1, cfg.UserUsecase, cfg.Config)
			}

			// 组织管理相关路由
			if cfg.OrganizationUsecase != nil {
				setupOrganizationRoutesV1(v1, cfg.OrganizationUsecase, cfg.Config)
			}

			// 空间管理相关路由
			if cfg.SpaceUsecase != nil {
				setupSpaceRoutesV1(v1, cfg.SpaceUsecase, cfg.Config)
			}

			// 文档管理相关路由
			if cfg.DocumentAggregateUsecase != nil {
				setupDocumentRoutesV1(v1, cfg.DocumentAggregateUsecase, cfg.Config)
			}
		}
	}

	// 404 处理
	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, ResponseData{
			Code:      http.StatusNotFound,
			Message:   "The requested resource was not found",
			Data:      nil,
			Timestamp: time.Now().Unix(),
		})
	})

	// 405 处理
	router.NoMethod(func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, ResponseData{
			Code:      http.StatusMethodNotAllowed,
			Message:   "The request method is not allowed for this resource",
			Data:      nil,
			Timestamp: time.Now().Unix(),
		})
	})
}

// setupAuthRoutesV1 设置认证相关路由（v1版本，匹配前端期望）
// 配置用户认证相关的所有路由
// 参数说明：
// - rg: 路由组
// - authUsecase: 认证业务逻辑接口
// - userUsecase: 用户业务逻辑接口
// - cfg: 应用配置
func setupAuthRoutesV1(rg *gin.RouterGroup, authUsecase domain.AuthUsecase, userUsecase domain.UserUsecase, cfg *config.Config) {
	// 创建 JWT 管理器
	jwtManager := jwt.NewJWTManager(
		cfg.App.JWTSecret,
		time.Duration(cfg.App.JWTExpireHours)*time.Hour,
	)

	// 创建认证中间件
	authMiddleware := middleware.NewAuthMiddleware(jwtManager)

	// 创建认证处理器
	authHandler := NewAuthHandler(authUsecase, userUsecase)

	// 认证相关路由（匹配前端期望的路径）
	auth := rg.Group("/auth")
	{
		// 无需认证的路由
		auth.POST("/register", authHandler.Register)             // 用户注册
		auth.POST("/login", authHandler.Login)                   // 用户登录
		auth.POST("/email/send-code", authHandler.SendEmailCode) // 发送邮箱验证码
		auth.POST("/email/login", authHandler.EmailLogin)        // 邮箱验证码登录
		auth.POST("/verify", authHandler.VerifyToken)            // 验证令牌
		auth.POST("/refresh", authHandler.RefreshToken)          // 刷新令牌

		// GitHub OAuth 路由
		auth.GET("/github", authHandler.GitHubLogin)             // GitHub 登录
		auth.GET("/github/callback", authHandler.GitHubCallback) // GitHub 回调
		auth.GET("/bind/github", authHandler.GitHubBind)

		// 需要认证的路由
		authProtected := auth.Group("")
		authProtected.Use(authMiddleware.RequireAuth())
		{
			authProtected.GET("/profile", authHandler.GetProfile)    // 获取当前用户资料
			authProtected.POST("/logout", authHandler.Logout)        // 退出登录
			authProtected.POST("/logout-all", authHandler.LogoutAll) // 退出所有设备
		}
	}
}

// setupUserRoutesV1 设置用户管理相关路由（v1版本，匹配前端期望）
// 配置用户信息管理相关的所有路由
// 参数说明：
// - rg: 路由组
// - userUsecase: 用户业务逻辑接口
// - cfg: 应用配置
func setupUserRoutesV1(rg *gin.RouterGroup, userUsecase domain.UserUsecase, cfg *config.Config) {
	// 创建 JWT 管理器
	jwtManager := jwt.NewJWTManager(
		cfg.App.JWTSecret,
		time.Duration(cfg.App.JWTExpireHours)*time.Hour,
	)

	// 创建认证中间件
	authMiddleware := middleware.NewAuthMiddleware(jwtManager)

	// 创建用户处理器
	userHandler := NewUserHandler(userUsecase)

	// 用户管理路由（需要认证）
	users := rg.Group("/users")
	users.Use(authMiddleware.RequireAuth()) // 应用认证中间件
	{
		// 用户信息相关路由（匹配 API 文档）
		users.GET("", userHandler.GetProfile)                  // GET /api/v1/users - 获取当前用户信息
		users.GET("/profile", userHandler.GetProfile)          // GET /api/v1/users/profile - 获取个人资料
		users.PUT("/profile", userHandler.UpdateProfile)       // PUT /api/v1/users/profile - 更新个人资料
		users.PUT("/password", userHandler.ChangePassword)     // PUT /api/v1/users/password - 修改密码
		users.GET("/search", userHandler.SearchUsers)          // GET /api/v1/users/search - 搜索用户
		users.GET("/:id", userHandler.GetUserByID)             // GET /api/v1/users/:id - 根据ID获取用户
		users.GET("/email/:email", userHandler.GetUserByEmail) // GET /api/v1/users/email/:email - 根据邮箱获取用户
		users.GET("/github/", userHandler.GetUserByGitHubId)   // GET /api/v1/users/github/ - 根据githubID获取用户
	}

	// 管理员路由（需要管理员权限）
	admin := rg.Group("/admin/users")
	admin.Use(authMiddleware.AdminAuth()) // 应用管理员认证中间件
	{
		//admin.GET("", userHandler.ListUsers) // 获取用户列表
	}
}

// healthCheck 健康检查处理器
func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "Service is healthy",
		"service": "InkwaveDocNet API",
		"version": "1.0.0",
	})
}

// ping 简单的 ping 处理器
func ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}

// setupOrganizationRoutesV1 设置组织相关路由
func setupOrganizationRoutesV1(v1 *gin.RouterGroup, organizationUsecase domain.OrganizationUsecase, config *config.Config) {
	// 创建组织处理器
	organizationHandler := NewOrganizationHandler(organizationUsecase)

	// 创建 JWT 认证中间件
	jwtManager := jwt.NewJWTManager(
		config.App.JWTSecret,
		time.Duration(config.App.JWTExpireHours)*time.Hour,
	)
	authMiddleware := middleware.NewAuthMiddleware(jwtManager)

	// 组织路由组
	organizations := v1.Group("/organizations")
	organizations.Use(authMiddleware.RequireAuth()) // 应用JWT认证中间件
	{
		// 组织基本操作
		organizations.POST("", organizationHandler.CreateOrganization)       // 创建组织
		organizations.GET("", organizationHandler.GetOrganizations)          // 获取组织列表
		organizations.GET("/my", organizationHandler.GetMyOrganizations)     // 获取我的组织
		organizations.GET("/:id", organizationHandler.GetOrganization)       // 获取组织详情
		organizations.PUT("/:id", organizationHandler.UpdateOrganization)    // 更新组织信息
		organizations.DELETE("/:id", organizationHandler.DeleteOrganization) // 删除组织

		// 成员管理
		organizations.POST("/:id/invite", organizationHandler.InviteMember)                    // 邀请成员
		organizations.POST("/:id/join-request", organizationHandler.RequestJoinOrganization)   // 申请加入
		organizations.GET("/:id/members", organizationHandler.GetOrganizationMembers)          // 获取成员列表
		organizations.PUT("/:id/members/:memberId/role", organizationHandler.UpdateMemberRole) // 更新成员角色
		organizations.DELETE("/:id/members/:memberId", organizationHandler.RemoveMember)       // 移除成员
		organizations.DELETE("/:id/leave", organizationHandler.LeaveOrganization)              // 退出组织

		// 申请处理
		organizations.POST("/join-requests/:requestId/process", organizationHandler.ProcessJoinRequest) // 处理加入申请
	}
}

// setupSpaceRoutesV1 设置空间相关路由
func setupSpaceRoutesV1(v1 *gin.RouterGroup, spaceUsecase domain.SpaceUsecase, config *config.Config) {
	// 创建空间处理器
	spaceHandler := NewSpaceHandler(spaceUsecase)

	// 创建 JWT 认证中间件
	jwtManager := jwt.NewJWTManager(
		config.App.JWTSecret,
		time.Duration(config.App.JWTExpireHours)*time.Hour,
	)
	authMiddleware := middleware.NewAuthMiddleware(jwtManager)

	// 空间路由组
	spaces := v1.Group("/spaces")
	spaces.Use(authMiddleware.RequireAuth()) // 应用JWT认证中间件
	{
		// 空间基本操作
		spaces.POST("", spaceHandler.Create)       // 创建空间
		spaces.GET("", spaceHandler.FindAll)       // 获取用户的空间列表
		spaces.GET("/:id", spaceHandler.FindOne)   // 获取空间详情
		spaces.PUT("/:id", spaceHandler.Update)    // 更新空间信息
		spaces.DELETE("/:id", spaceHandler.Remove) // 删除空间

		// 空间文档管理
		spaces.GET("/:id/documents", spaceHandler.GetDocuments)                  // 获取空间中的文档
		spaces.POST("/:id/documents", spaceHandler.AddDocument)                  // 添加文档到空间
		spaces.DELETE("/:id/documents/:documentId", spaceHandler.RemoveDocument) // 从空间移除文档

		// 空间成员管理
		spaces.GET("/:id/members", spaceHandler.GetMembers)               // 获取空间成员列表
		spaces.POST("/:id/members", spaceHandler.AddMember)               // 添加空间成员
		spaces.PUT("/:id/members/:userId", spaceHandler.UpdateMemberRole) // 更新成员角色
		spaces.DELETE("/:id/members/:userId", spaceHandler.RemoveMember)  // 移除空间成员
	}
}

// todo测试文档相关接口
// setupDocumentRoutesV1 设置文档相关路由（v1版本，匹配API规范）
// 配置文档管理相关的所有路由，包括文档的CRUD、分享、权限、收藏等功能
// 参数说明：
// - rg: 路由组
// - documentService: 文档聚合服务接口
// - cfg: 应用配置
func setupDocumentRoutesV1(rg *gin.RouterGroup, documentService domain.DocumentAggregateUsecase, cfg *config.Config) {
	// 创建 JWT 管理器
	jwtManager := jwt.NewJWTManager(
		cfg.App.JWTSecret,
		time.Duration(cfg.App.JWTExpireHours)*time.Hour,
	)

	// 创建认证中间件
	authMiddleware := middleware.NewAuthMiddleware(jwtManager)

	// 创建文档处理器
	documentHandler := NewDocumentHandler(documentService)

	// 文档相关路由（需要认证）
	documents := rg.Group("/documents")
	documents.Use(authMiddleware.RequireAuth()) // 应用认证中间件
	{
		// === 文档基本操作 ===
		documents.POST("", documentHandler.CreateDocument)       // POST /api/v1/documents - 创建文档
		documents.GET("", documentHandler.GetMyDocuments)        // GET /api/v1/documents - 获取我的文档列表
		documents.GET("/:id", documentHandler.GetDocument)       // GET /api/v1/documents/:id - 获取文档详情
		documents.PUT("/:id", documentHandler.UpdateDocument)    // PUT /api/v1/documents/:id - 更新文档信息
		documents.DELETE("/:id", documentHandler.DeleteDocument) // DELETE /api/v1/documents/:id - 删除文档

		// === 文档内容操作 ===
		documents.GET("/:id/content", documentHandler.GetDocumentContent)    // GET /api/v1/documents/:id/content - 获取文档内容
		documents.PUT("/:id/content", documentHandler.UpdateDocumentContent) // PUT /api/v1/documents/:id/content - 更新文档内容

		// === 文档搜索 ===
		documents.GET("/search", documentHandler.SearchDocuments) // GET /api/v1/documents/search - 搜索文档

		// === 文档分享操作 ===
		documents.POST("/:id/share", documentHandler.CreateShareLink)           // POST /api/v1/documents/:id/share - 创建分享链接
		documents.GET("/shared-via-link", documentHandler.GetMySharedDocuments) // GET /api/v1/documents/shared-via-link - 获取我分享的文档

		// === 文档权限操作 ===
		documents.PUT("/:id/acl", documentHandler.SetUserPermission) // PUT /api/v1/documents/:id/acl - 设置用户权限

		// === 文档收藏操作 ===
		documents.POST("/:id/shared/favorite", documentHandler.ToggleFavoriteDocument) // POST /api/v1/documents/:id/shared/favorite - 切换收藏状态
		documents.PUT("/:id/shared/title", documentHandler.SetFavoriteCustomTitle)     // PUT /api/v1/documents/:id/shared/title - 设置收藏自定义标题
		documents.DELETE("/:id/shared", documentHandler.RemoveFavoriteDocument)        // DELETE /api/v1/documents/:id/shared - 移除收藏

		// === 批量操作 ===
		documents.DELETE("/batch", documentHandler.BatchDeleteDocuments) // DELETE /api/v1/documents/batch - 批量删除文档
		documents.PUT("/batch/move", documentHandler.BatchMoveDocuments) // PUT /api/v1/documents/batch/move - 批量移动文档
	}

	// === 公开的分享访问路由（无需认证） ===
	// 通过分享链接访问文档的路由，不需要用户登录
	documents.GET("/shared/:linkId", documentHandler.GetSharedDocument) // GET /api/v1/documents/shared/:linkId - 通过分享链接访问文档

	// === 用户权限检查路由 ===
	// 检查用户对特定文档的访问权限
	userRoutes := rg.Group("/users")
	userRoutes.Use(authMiddleware.RequireAuth())
	{
		userRoutes.GET("/check-document-access/:documentId", documentHandler.CheckDocumentAccess) // GET /api/v1/users/check-document-access/:documentId - 检查文档访问权限
	}
}
