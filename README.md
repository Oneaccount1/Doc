# DOC

一个基于 Clean Architecture 设计的现代化 Go 语言文档协作平台，使用 Gin + GORM + Redis + MySQL 技术栈。项目采用领域驱动设计（DDD）和分层架构，提供高性能的多人实时协作文档编辑功能。

## 🏗️ 项目架构

本项目采用 Clean Architecture（整洁架构）设计原则，确保代码的可维护性、可测试性和可扩展性。

```
refatorSiwu/
├── domain/                           # 领域层 - 实体和业务规则
│   ├── user.go                      # 用户实体和接口
│   ├── organization.go              # 组织实体和接口  
│   ├── space.go                     # 空间实体和接口
│   ├── document.go                  # 文档实体和接口
│   ├── document_core.go             # 文档核心业务
│   ├── document_favorite.go         # 文档收藏功能
│   ├── document_permission.go       # 文档权限管理
│   ├── document_share.go            # 文档分享功能
│   ├── auth.go                      # 认证相关接口
│   ├── email.go                     # 邮件服务接口
│   ├── collaboration.go             # 协作功能接口
│   ├── ai.go                        # AI 功能接口
│   ├── upload.go                    # 文件上传接口
│   └── errors.go                    # 领域错误定义
├── user/                            # 用户业务服务层
│   ├── service.go                   # 用户业务逻辑
│   └── service_test.go              # 用户服务测试
├── auth/                            # 认证业务服务层
│   └── service.go                   # 认证业务逻辑
├── organization/                    # 组织业务服务层
│   └── service.go                   # 组织业务逻辑
├── space/                           # 空间业务服务层
│   └── service.go                   # 空间业务逻辑
├── Document/                        # 文档业务服务层
│   ├── core.go                      # 文档核心服务
│   ├── core_test.go                 # 文档核心测试
│   ├── aggregate_service.go         # 文档聚合服务
│   ├── favorite.go                  # 文档收藏服务
│   ├── permission.go                # 文档权限服务
│   ├── share.go                     # 文档分享服务
│   └── example_integration.go       # 集成示例
├── email/                           # 邮件业务服务层
│   ├── service.go                   # 邮件服务实现
│   ├── service_test.go              # 邮件服务测试
│   └── mocks/                       # 邮件服务 Mock
├── internal/                        # 内部实现层
│   ├── database/                    # 数据库连接
│   │   ├── mysql.go                 # MySQL 连接和迁移
│   │   └── redis.go                 # Redis 连接
│   ├── repository/                  # 仓储实现层
│   │   ├── mysql/                   # MySQL 仓储实现
│   │   │   ├── user_repository.go   # 用户仓储
│   │   │   ├── auth_repository.go   # 认证仓储
│   │   │   ├── organization_repository.go # 组织仓储
│   │   │   ├── space_repository.go  # 空间仓储
│   │   │   ├── document_repository.go # 文档仓储
│   │   │   ├── document_favorite.go # 文档收藏仓储
│   │   │   ├── document_permission_repository.go # 文档权限仓储
│   │   │   ├── document_share_repository.go # 文档分享仓储
│   │   │   └── email_repository.go  # 邮件仓储
│   │   └── redis/                   # Redis 仓储实现
│   │       ├── base.go              # Redis 基础配置
│   │       ├── user_cache.go        # 用户缓存
│   │       └── auth_cache.go        # 认证缓存
│   ├── rest/                        # REST API 层
│   │   ├── router.go                # 路由配置
│   │   ├── response.go              # 响应结构
│   │   ├── auth_handler.go          # 认证处理器
│   │   ├── user_handler.go          # 用户处理器
│   │   ├── organization_handler.go  # 组织处理器
│   │   ├── space_handler.go         # 空间处理器
│   │   ├── document_handler.go      # 文档处理器
│   │   ├── dto/                     # 数据传输对象
│   │   │   ├── auth_dto.go          # 认证 DTO
│   │   │   ├── user_dto.go          # 用户 DTO
│   │   │   ├── organization_dto.go  # 组织 DTO
│   │   │   ├── space_dto.go         # 空间 DTO
│   │   │   ├── document_dto.go      # 文档 DTO
│   │   │   └── common_dto.go        # 通用 DTO
│   │   └── middleware/              # 中间件
│   │       ├── auth.go              # 认证中间件
│   │       ├── cors.go              # CORS 中间件
│   │       ├── logger.go            # 日志中间件
│   │       ├── recovery.go          # 恢复中间件
│   │       └── validation.go        # 验证中间件
│   ├── websocket/                   # WebSocket 实时通信
│   │   ├── hub.go                   # WebSocket 中心
│   │   ├── client.go                # WebSocket 客户端
│   │   ├── server.go                # WebSocket 服务器
│   │   └── example_test.go          # WebSocket 测试示例
│   └── workers/                     # 后台工作者
│       └── email/                   # 邮件工作者
│           ├── worker.go            # 邮件工作者
│           └── sender.go            # 邮件发送器
├── app/                             # 应用启动层
│   └── app.go                       # 应用初始化和启动
├── config/                          # 配置管理
│   └── config.go                    # 配置结构和加载
├── pkg/                             # 公共包
│   ├── jwt/                         # JWT 工具
│   │   └── jwt.go                   # JWT 管理器
│   └── utils/                       # 工具函数
│       ├── hash.go                  # 哈希工具
│       ├── check_valid.go           # 验证工具
│       └── jsonmap.go               # JSON 映射工具
├── templates/                       # 模板文件
│   └── email/                       # 邮件模板
│       ├── verification_register.html    # 注册验证邮件
│       ├── verification_login.html       # 登录验证邮件
│       ├── verification_code.html        # 验证码邮件
│       ├── password_reset.html           # 密码重置邮件
│       ├── organization_invitation.html  # 组织邀请邮件
│       └── welcome.html                  # 欢迎邮件
├── docs/                            # 项目文档
│   ├── api/                         # API 文档
│   │   ├── modules/                 # 模块API文档
│   │   ├── openapi/                 # OpenAPI 规范
│   │   ├── postman/                 # Postman 集合
│   │   ├── quick-start.md           # 快速开始
│   │   └── README.md                # API 文档说明
│   ├── architecture/                # 架构文档
│   │   └── notification_design_principles.md
│   ├── document_module_guide.md     # 文档模块指南
│   └── email_service.md             # 邮件服务文档
├── config.yaml                      # 配置文件
├── go.mod                           # Go 模块文件
├── go.sum                           # Go 依赖校验
├── main.go                          # 应用入口
├── Makefile                         # 构建脚本
└── README.md                        # 项目说明
```

## 🚀 技术栈

- **Web 框架**: [Gin](https://github.com/gin-gonic/gin) - 高性能的 Go Web 框架
- **ORM**: [GORM](https://gorm.io/) - Go 语言 ORM 库
- **数据库**: MySQL - 关系型数据库
- **缓存**: Redis - 内存数据库

## 📋 功能特性

### 🔥 核心功能
- ✅ **多人实时协作**: WebSocket 实时文档编辑，支持多用户同时编辑
- ✅ **智能文档管理**: 文档创建、编辑、删除、锁定、版本控制
- ✅ **组织空间管理**: 多级组织架构，空间权限隔离
- ✅ **权限控制系统**: 细粒度权限管理，支持读写、分享等权限
- ✅ **文档分享功能**: 支持内部分享和外部链接分享
- ✅ **收藏系统**: 个人文档收藏和快速访问

### 🔐 认证与安全
- ✅ **JWT 认证**: 安全的 Token 认证机制
- ✅ **邮箱验证**: 注册、登录、密码重置邮箱验证
- ✅ **OAuth 集成**: 支持 GitHub OAuth 第三方登录
- ✅ **安全中间件**: 认证、CORS、限流等安全防护

### 📧 通知系统
- ✅ **邮件服务**: 异步邮件队列，支持 HTML 模板
- ✅ **系统通知**: 组织邀请、文档分享等通知功能
- ✅ **邮件模板**: 丰富的邮件模板，支持多种场景

### 🏗️ 架构特性  
- ✅ **Clean Architecture**: 分层架构，依赖倒置原则
- ✅ **领域驱动设计**: DDD 设计模式，业务逻辑清晰
- ✅ **高性能缓存**: Redis 缓存支持，提升响应速度
- ✅ **数据库优化**: GORM ORM，自动迁移和种子数据
- ✅ **WebSocket 通信**: 实时双向通信，支持协作功能
- ✅ **后台任务**: 邮件发送、数据清理等异步任务
- ✅ **配置管理**: YAML 配置文件，支持环境变量
- ✅ **优雅关闭**: 信号监听，资源清理，数据安全
- ✅ **单元测试**: 完整的测试覆盖，Mock 支持
- ✅ **API 文档**: OpenAPI 规范，Postman 集合

## 🛠️ 快速开始

### 环境要求

- Go 1.24+
- MySQL 8.0+
- Redis 6.0+

### 安装依赖

```bash
go mod tidy
```

### 配置数据库

1. 创建 MySQL 数据库：
```sql
CREATE DATABASE docflow CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

2. 修改 `config.yaml` 中的数据库配置：
```yaml
database:
  host: "localhost"
  port: "3306" 
  username: "root"
  password: "your_password"
  dbname: "docflow"
  charset: "utf8mb4"

redis:
  host: "localhost"
  port: "6379"
  password: ""
  db: 0

app:
  name: "RefatorSiwu"
  version: "1.0.0"
  jwt_secret: "your-super-secret-jwt-key-change-in-production"
  jwt_expire_hours: 240

# 邮件配置（可选）
email:
  smtp_host: "smtp.qq.com"
  smtp_port: 587
  username: "your_email@qq.com"
  password: "your_smtp_password"
  from_email: "your_email@qq.com"
  from_name: "RefatorSiwu"
```

### 启动应用

```bash
# 使用默认配置
go run main.go

# 指定配置文件路径
go run main.go -config ./config
```

应用将在 `http://localhost:8080` 启动。

### 健康检查

```bash
curl http://localhost:8080/health
```

## 🔧 配置说明

项目使用 YAML 格式的配置文件，支持环境变量覆盖：

```yaml
# 服务器配置
server:
  host: "localhost"
  port: "8080"
  mode: "debug"  # debug, release, test

# 数据库配置
database:
  host: "localhost"
  port: "3306"
  username: "root"
  password: "123456"
  dbname: "docflow"
  charset: "utf8mb4"

# Redis 配置
redis:
  host: "localhost"
  port: "6379"
  password: ""
  db: 0

# 应用配置
app:
  name: "RefatorSiwu"
  version: "1.0.0"
  context_timeout: 30
  jwt_secret: "your-super-secret-jwt-key-change-in-production"
  jwt_expire_hours: 240
  log_level: "debug"
  enable_cors: true
  enable_rate_limit: true
  enable_websocket: true
  max_file_size: 10485760  # 10MB

# 邮件配置
email:
  smtp_host: "smtp.qq.com"
  smtp_port: 587
  username: "your_email@qq.com"
  password: "your_smtp_password"
  from_email: "your_email@qq.com"
  from_name: "RefatorSiwu"
  template_dir: "./templates/email"
  enable_tls: true
  connection_timeout: 30
  send_timeout: 60
  max_retries: 3
  retry_interval: 5

# OAuth 配置
oauth:
  github:
    client_id: "your_github_client_id"
    client_secret: "your_github_client_secret"
    redirect_url: "http://localhost:3000/auth/callback"
    api_callback_url: "http://localhost:8080/api/v1/auth/github/callback"
    scopes: ["user:email", "read:user"]

# 认证安全配置
auth:
  verification_code:
    length: 6
    expire_minutes: 10
    send_interval_seconds: 60
    max_attempts: 5
  security:
    password_min_length: 8
    password_require_special: true
    max_login_attempts: 5
    lockout_duration_minutes: 30
```

## 🏛️ 架构设计原则

### 🔄 Clean Architecture（整洁架构）
- **依赖倒置**: 高层模块不依赖低层模块，都依赖于抽象接口
- **关注点分离**: 每层专注于自己的职责，减少耦合
- **依赖注入**: 使用接口抽象，便于测试和扩展

### 📦 分层架构
- **领域层 (Domain)**: 业务实体、值对象、领域服务和接口定义
- **应用层 (UseCase)**: 业务用例、服务协调和业务流程
- **基础设施层 (Infrastructure)**: 数据持久化、外部服务接入
- **接口层 (Interface)**: HTTP API、WebSocket、中间件

### 🧪 测试驱动设计
- **接口抽象**: 所有外部依赖都通过接口抽象
- **Mock 支持**: 使用 testify/mock 进行单元测试
- **集成测试**: 提供完整的集成测试示例
- **测试覆盖**: 核心业务逻辑 100% 测试覆盖

### ⚡ 性能优化
- **缓存策略**: Redis 多级缓存，提升响应速度
- **连接池**: 数据库连接池，避免频繁连接创建
- **异步处理**: 邮件发送、通知推送等异步任务处理
- **优雅关闭**: 资源清理，确保数据一致性

### 🔒 安全设计
- **JWT 认证**: 无状态认证，支持分布式部署
- **权限控制**: 基于角色的访问控制 (RBAC)
- **输入验证**: 请求参数验证，防止注入攻击
- **安全中间件**: CORS、限流、日志审计等



<div align="center">
  <strong>⭐ 如果这个项目对你有帮助，请给我们一个 Star！ ⭐</strong>
</div>
