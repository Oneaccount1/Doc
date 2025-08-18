package database

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"DOC/config"
	"DOC/domain"
)

// NewMySQLConnection 创建 MySQL 数据库连接
func NewMySQLConnection(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	dsn := cfg.GetDSN()

	// 配置 GORM
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now().Local()
		},
		SkipDefaultTransaction:                   true,
		DisableForeignKeyConstraintWhenMigrating: true,
	}

	// 连接数据库
	db, err := gorm.Open(mysql.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	// 获取底层的 sql.DB 对象进行连接池配置
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %v", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(10)           // 最大空闲连接数
	sqlDB.SetMaxOpenConns(100)          // 最大打开连接数
	sqlDB.SetConnMaxLifetime(time.Hour) // 连接最大生存时间

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	log.Println("Successfully connected to MySQL database")
	return db, nil
}

// AutoMigrate 自动迁移数据库表结构
func AutoMigrate(db *gorm.DB) error {
	log.Println("Starting database migration...")

	// 自动迁移表结构
	// 按照依赖关系顺序迁移表，避免外键约束错误
	err := db.AutoMigrate(
		&domain.User{},                    // 用户表 - 基础表
		&domain.VerificationCode{},        // 验证码表
		&domain.AuthSession{},             // 认证会话表
		&domain.OAuthState{},              // 验证状态表
		&domain.Organization{},            // 组织表
		&domain.OrganizationMember{},      // 组织成员表
		&domain.OrganizationInvitation{},  // 组织邀请表
		&domain.OrganizationJoinRequest{}, // 组织加入申请表
		&domain.Space{},                   // 空间表
		&domain.SpaceMember{},             //空间成员表
		&domain.SpaceDocument{},           // 空间文档表 多对多
		&domain.Document{},                // 文档表
		&domain.DocumentFavorite{},        // 文档收藏表
		&domain.DocumentPermission{},      // 文档权限表
		&domain.DocumentShare{},           // 文档分享表
		&domain.Email{},                   // 邮件表
	)
	if err != nil {
		return fmt.Errorf("failed to migrate database: %v", err)
	}

	log.Println("Database migration completed successfully")
	return nil
}

func SeedData(db *gorm.DB) error {
	return nil
}
