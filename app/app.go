package app

import (
	document "DOC/Document"
	"DOC/auth"
	"DOC/organization"
	"DOC/space"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	email2 "DOC/email"
	redis2 "DOC/internal/repository/redis"
	"DOC/internal/websocket"
	"DOC/internal/workers/email"

	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"DOC/config"

	"DOC/domain"
	"DOC/internal/database"

	"DOC/internal/repository/mysql"
	"DOC/internal/rest"

	"DOC/pkg/jwt"
	"DOC/user"
)

// App 应用结构
type App struct {
	config *config.Config
	db     *gorm.DB
	redis  *redis.Client
	router *gin.Engine
	server *http.Server

	// 用户仓储层
	userRepo  domain.UserRepository
	userCache domain.UserCacheRepository
	// 认证仓储层
	authRepo  domain.AuthRepository
	authCache domain.AuthCacheRepository
	// 组织仓储层
	organizationRepo domain.OrganizationRepository
	spaceRepo        domain.SpaceRepository
	//文档仓储层
	documentRepo           domain.DocumentRepository
	documentPermissionRepo domain.DocumentPermissionRepository
	documentFavoriteRepo   domain.DocumentFavoriteRepository
	documentShareRepo      domain.DocumentShareRepository

	emailRep domain.EmailRepository

	// 用户业务层
	userUsecase domain.UserUsecase

	// 认证业务层
	authUsecase domain.AuthUsecase
	// 组织业务层
	organizationUsecase domain.OrganizationUsecase
	// 空间业务层
	spaceUsecase domain.SpaceUsecase
	// 文档业务层
	documentUsecase           domain.DocumentUsecase
	documentFavoriteUsecase   domain.DocumentFavoriteUsecase
	documentShareUsecase      domain.DocumentShareUsecase
	documentPermissionUsecase domain.DocumentPermissionUsecase
	DocumentAggregateUsecase  domain.DocumentAggregateUsecase
	emailUseCase              domain.EmailUsecase

	// 邮件发送服务
	emailSender domain.EmailSender

	// 工作者
	emailWorker *email.EmailWorker

	// WebSocket 服务
	wsHub    *websocket.Hub
	wsServer *websocket.Server
}

// NewApp 创建新的应用实例
func NewApp(configPath string) (*App, error) {
	app := &App{}

	// 加载配置
	if err := app.loadConfig(configPath); err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}

	// 初始化数据库
	if err := app.initDatabase(); err != nil {
		return nil, fmt.Errorf("failed to init database: %v", err)
	}

	// 初始化 Redis
	if err := app.initRedis(); err != nil {
		return nil, fmt.Errorf("failed to init redis: %v", err)
	}

	// 初始化仓储层
	app.initRepositories()

	// 初始化业务层
	app.initUsecases()

	// 初始化 WebSocket 服务
	app.initWebSocket()

	// 初始化路由
	app.initRouter()

	// 初始化 HTTP 服务器
	app.initServer()

	log.Println("Application initialized successfully")
	return app, nil
}

// loadConfig 加载配置
func (a *App) loadConfig(configPath string) error {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return err
	}
	a.config = cfg
	log.Printf("Config loaded from: %s", configPath)
	return nil
}

// initDatabase 初始化数据库
func (a *App) initDatabase() error {
	db, err := database.NewMySQLConnection(&a.config.Database)
	if err != nil {
		return err
	}
	a.db = db

	// 自动迁移
	if err := database.AutoMigrate(db); err != nil {
		return fmt.Errorf("failed to migrate database: %v", err)
	}

	// 初始化种子数据
	if err := database.SeedData(db); err != nil {
		return fmt.Errorf("failed to seed data: %v", err)
	}

	return nil
}

// initRedis 初始化 Redis
func (a *App) initRedis() error {
	rdb, err := database.NewRedisConnection(&a.config.Redis)
	if err != nil {
		return err
	}
	a.redis = rdb
	return nil
}

// initRepositories 初始化仓储层
func (a *App) initRepositories() {
	// 初始化用户仓储
	a.userRepo = mysql.NewMysqlUserRepository(a.db)
	a.userCache = redis2.NewUserCacheRepository(a.redis)

	// 初始化认证仓储
	a.authRepo = mysql.NewMysqlAuthRepository(a.db)
	a.authCache = redis2.NewAuthCacheRepository(a.redis)

	// 初始化组织仓储
	a.organizationRepo = mysql.NewOrganizationRepository(a.db)

	// 初始化空间
	a.spaceRepo = mysql.NewSpaceRepository(a.db)

	// 初始化文档仓储
	a.documentRepo = mysql.NewDocumentRepository(a.db)
	a.documentShareRepo = mysql.NewDocumentShareRepository(a.db)
	a.documentFavoriteRepo = mysql.NewDocumentFavoriteRepository(a.db)
	a.documentPermissionRepo = mysql.NewDocumentPermissionRepository(a.db)

	// 初始化邮件仓储
	a.emailRep = mysql.NewEmailRepository(a.db)

	log.Println("Repositories initialized")
}

// initUsecases 初始化业务层
func (a *App) initUsecases() {
	timeout := time.Duration(a.config.App.ContextTimeout) * time.Second

	// 创建 JWT 管理器
	jwtManager := jwt.NewJWTManager(
		a.config.App.JWTSecret,
		time.Duration(a.config.App.JWTExpireHours)*time.Hour,
	)
	// 初始化邮件发送服务
	sender, err := email.NewEmailSender(a.config.Email)
	if err != nil {
		log.Fatalf("初始化邮件发送失败: %v", err)
	}
	a.emailSender = sender
	// 初始化邮件工作者
	workerConfig := email.WorkerConfig{
		WorkerCount:  2,
		PollInterval: 20 * time.Second,
	}
	a.emailWorker = email.NewEmailWorker(a.emailRep, a.emailSender, workerConfig)

	// 初始化邮件业务服务
	a.emailUseCase = email2.NewEmailService(
		a.emailRep,
		a.emailSender,
		timeout,
	)

	// 初始化用户业务服务
	a.userUsecase = user.NewUserService(
		a.userRepo,
		a.userCache,
		timeout,
	)

	// 初始化 验证服务
	a.authUsecase = auth.NewAuthService(
		a.userRepo,
		a.authRepo,
		a.authCache,
		a.emailUseCase,
		jwtManager,
		a.config,
		timeout,
	)

	// 初始化组织业务服务
	a.organizationUsecase = organization.NewOrganizationService(
		a.organizationRepo,
		a.userRepo,
		a.emailUseCase,
		timeout,
	)
	// 初始化空间业务服务
	a.spaceUsecase = space.NewSpaceService(
		a.spaceRepo,
		a.userRepo,
		a.organizationRepo,
		nil,
		timeout,
	)
	// 初始化各个文档服务
	a.documentShareUsecase = document.NewDocumentShareService(
		a.documentShareRepo,
		a.documentRepo,
		a.documentPermissionRepo,
	)
	// 权限
	a.documentPermissionUsecase = document.NewDocumentPermissionService(
		a.documentPermissionRepo,
		a.documentRepo,
	)
	// 分享
	a.documentFavoriteUsecase = document.NewDocumentFavoriteService(
		a.documentFavoriteRepo,
		a.documentRepo,
	)
	// 聚合
	a.documentUsecase = document.NewDocumentService(
		a.documentRepo,
		a.documentShareUsecase,
		a.documentPermissionUsecase,
		a.documentFavoriteUsecase,
		a.userRepo,
	)

	// 初始化文档聚合服务
	a.DocumentAggregateUsecase = document.NewDocumentAggregateService(
		a.documentUsecase,
		a.documentShareUsecase,
		a.documentPermissionUsecase,
		a.documentFavoriteUsecase,
		a.userRepo,
	)

	log.Println("Usecases initialized")
}

// initWebSocket 初始化 WebSocket 服务
func (a *App) initWebSocket() {
	// 创建 JWT 管理器
	jwtManager := jwt.NewJWTManager(
		a.config.App.JWTSecret,
		time.Duration(a.config.App.JWTExpireHours)*time.Hour,
	)

	// 创建 WebSocket Hub
	a.wsHub = websocket.NewHub(nil) // 暂时传入 nil，后续可以添加协作仓储

	// 创建 WebSocket 服务器
	a.wsServer = websocket.NewServer(a.wsHub, jwtManager, nil) // 暂时传入 nil，后续可以添加协作用例

	// 启动 WebSocket 服务
	a.wsServer.Start()

	log.Println("WebSocket service initialized")
}

// initRouter 初始化路由
func (a *App) initRouter() {
	routerConfig := rest.RouterConfig{
		UserUsecase:              a.userUsecase,
		AuthUsecase:              a.authUsecase,
		OrganizationUsecase:      a.organizationUsecase,
		SpaceUsecase:             a.spaceUsecase,
		DocumentAggregateUsecase: a.DocumentAggregateUsecase,
		Config:                   a.config,
	}
	a.router = rest.NewRouter(routerConfig)

	// 注册 WebSocket 路由
	a.wsServer.RegisterRoutes(a.router)

	log.Println("Router initialized")
}

// initServer 初始化 HTTP 服务器
func (a *App) initServer() {
	addr := fmt.Sprintf("%s:%s", a.config.Server.Host, a.config.Server.Port)
	a.server = &http.Server{
		Addr:           addr,
		Handler:        a.router,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}
	log.Printf("HTTP server configured on %s", addr)
}

// Run 运行应用
func (a *App) Run() error {
	// 启动服务器
	go func() {
		log.Printf("Starting server on %s", a.server.Addr)
		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Printf("🚀 %s server started on %s", a.config.App.Name, a.server.Addr)
	log.Printf("📖 API documentation available at: http://%s/health", a.server.Addr)

	// 等待中断信号
	return a.gracefulShutdown()
}

// gracefulShutdown 优雅关闭
func (a *App) gracefulShutdown() error {
	// 创建一个接收系统信号的通道
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 阻塞等待信号
	<-quit
	log.Println("Shutting down server...")

	// 创建一个超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 关闭 WebSocket 服务
	if a.wsServer != nil {
		a.wsServer.Stop()
		log.Println("WebSocket server stopped")
	}

	// 关闭邮件工作者
	if a.emailWorker != nil {
		a.emailWorker.Stop()
		log.Println("Email worker stopped")
	}

	// 关闭 HTTP 服务器
	if err := a.server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
		return err
	}

	// 关闭数据库连接
	if a.db != nil {
		sqlDB, err := a.db.DB()
		if err == nil {
			sqlDB.Close()
		}
		log.Println("Database connection closed")
	}

	// 关闭 Redis 连接
	if a.redis != nil {
		a.redis.Close()
		log.Println("Redis connection closed")
	}

	log.Println("Server exited")
	return nil
}

// GetConfig 获取配置
func (a *App) GetConfig() *config.Config {
	return a.config
}

// GetDB 获取数据库连接
func (a *App) GetDB() *gorm.DB {
	return a.db
}

// GetRedis 获取 Redis 连接
func (a *App) GetRedis() *redis.Client {
	return a.redis
}
