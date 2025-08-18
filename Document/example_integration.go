package document

//
//import (
//	"gorm.io/gorm"
//
//	"refatorSiwu/domain"
//	"refatorSiwu/internal/repository/mysql"
//)
//
//// ExampleIntegration 示例：如何在应用中整合文档模块的所有组件
//// 这个文件展示了如何将文档核心服务、聚合服务、仓储等组件组装在一起
//// 在实际项目中，这些代码应该放在 app/app.go 或类似的应用初始化文件中
//
//// DocumentModuleConfig 文档模块配置
//type DocumentModuleConfig struct {
//	DB                      *gorm.DB                             // 数据库连接
//	UserRepository          domain.UserRepository                // 用户仓储
//	DocumentShareUsecase    domain.DocumentShareUsecase         // 分享子域服务
//	DocumentPermissionUsecase domain.DocumentPermissionUsecase  // 权限子域服务
//	DocumentFavoriteUsecase domain.DocumentFavoriteUsecase      // 收藏子域服务
//}
//
//// SetupDocumentModule 设置文档模块
//// 创建并返回文档聚合服务，该服务整合了所有文档相关的功能
//func SetupDocumentModule(config DocumentModuleConfig) domain.DocumentAggregateService {
//	// 1. 创建文档仓储
//	documentRepo := mysql.NewDocumentRepository(config.DB)
//
//	// 2. 创建文档核心业务服务
//	documentService := NewDocumentService(
//		documentRepo,
//		config.DocumentShareUsecase,
//		config.DocumentPermissionUsecase,
//		config.DocumentFavoriteUsecase,
//		config.UserRepository,
//	)
//
//	// 3. 创建文档聚合服务
//	aggregateService := NewDocumentAggregateService(
//		documentService,
//		config.DocumentShareUsecase,
//		config.DocumentPermissionUsecase,
//		config.DocumentFavoriteUsecase,
//		config.UserRepository,
//	)
//
//	return aggregateService
//}
//
//// ExampleUsage 示例用法
//// 展示如何在 app/app.go 中使用文档模块
///*
//func SetupApplication() {
//	// 初始化数据库连接
//	db := setupDatabase()
//
//	// 初始化其他仓储
//	userRepo := mysql.NewUserRepository(db)
//
//	// 初始化子域服务（分享、权限、收藏）
//	shareRepo := mysql.NewDocumentShareRepository(db)
//	shareUsecase := share.NewDocumentShareUsecase(shareRepo, userRepo)
//
//	permRepo := mysql.NewDocumentPermissionRepository(db)
//	permUsecase := permission.NewDocumentPermissionUsecase(permRepo, userRepo)
//
//	favoriteRepo := mysql.NewDocumentFavoriteRepository(db)
//	favoriteUsecase := favorite.NewDocumentFavoriteUsecase(favoriteRepo, userRepo)
//
//	// 设置文档模块
//	documentService := document.SetupDocumentModule(document.DocumentModuleConfig{
//		DB:                        db,
//		UserRepository:            userRepo,
//		DocumentShareUsecase:      shareUsecase,
//		DocumentPermissionUsecase: permUsecase,
//		DocumentFavoriteUsecase:   favoriteUsecase,
//	})
//
//	// 初始化路由配置
//	routerConfig := rest.RouterConfig{
//		UserUsecase:              userUsecase,
//		AuthUsecase:              authUsecase,
//		OrganizationUsecase:      orgUsecase,
//		SpaceUsecase:             spaceUsecase,
//		DocumentAggregateService: documentService, // 添加文档聚合服务
//		Config:                   config,
//	}
//
//	// 创建并启动HTTP服务器
//	router := rest.NewRouter(routerConfig)
//	router.Run(":8080")
//}
//*/
//
//// ExampleAPIUsage 示例API使用
//// 展示如何通过HTTP API使用文档功能
///*
//# 创建文档
//POST /api/v1/documents
//Content-Type: application/json
//Authorization: Bearer <token>
//
//{
//  "title": "我的新文档",
//  "content": {
//    "blocks": [],
//    "entityMap": {}
//  },
//  "type": "FILE",
//  "parent_id": null,
//  "space_id": 1,
//  "sort_order": 0,
//  "is_starred": false
//}
//
//# 获取文档列表
//GET /api/v1/documents?parent_id=1&page=1&page_size=20
//Authorization: Bearer <token>
//
//# 搜索文档
//GET /api/v1/documents/search?keyword=测试&type=FILE&limit=20&offset=0
//Authorization: Bearer <token>
//
//# 更新文档内容
//PUT /api/v1/documents/123/content
//Content-Type: application/json
//Authorization: Bearer <token>
//
//{
//  "content": {
//    "blocks": [
//      {
//        "type": "paragraph",
//        "data": {
//          "text": "这是更新后的内容"
//        }
//      }
//    ],
//    "entityMap": {}
//  }
//}
//
//# 创建分享链接
//POST /api/v1/documents/123/share
//Content-Type: application/json
//Authorization: Bearer <token>
//
//{
//  "permission": "VIEW",
//  "password": "secret123",
//  "expires_at": "2024-12-31T23:59:59Z",
//  "shareWithUserIds": [456, 789]
//}
//
//# 通过分享链接访问文档（无需登录）
//GET /api/v1/documents/shared/abc123def?password=secret123
//
//# 切换文档收藏状态
//POST /api/v1/documents/123/shared/favorite
//Authorization: Bearer <token>
//
//# 设置文档权限
//PUT /api/v1/documents/123/acl
//Content-Type: application/json
//Authorization: Bearer <token>
//
//{
//  "target_user_id": 456,
//  "permission": "EDIT"
//}
//
//# 批量删除文档
//DELETE /api/v1/documents/batch
//Content-Type: application/json
//Authorization: Bearer <token>
//
//{
//  "document_ids": [123, 124, 125]
//}
//*/
