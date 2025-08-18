package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

// Config 应用配置结构
type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Redis     RedisConfig     `mapstructure:"redis"`
	App       AppConfig       `mapstructure:"app"`
	Email     EmailConfig     `mapstructure:"email"`
	WebSocket WebSocketConfig `mapstructure:"websocket"`
	OAuth     OAuthConfig     `mapstructure:"oauth"`
	Auth      AuthConfig      `mapstructure:"auth"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port string `mapstructure:"port"`
	Mode string `mapstructure:"mode"` // debug, release, test
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	Charset  string `mapstructure:"charset"`
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// AppConfig 应用配置
type AppConfig struct {
	Name            string `mapstructure:"name"`
	Version         string `mapstructure:"version"`
	ContextTimeout  int    `mapstructure:"context_timeout"`
	JWTSecret       string `mapstructure:"jwt_secret"`
	JWTExpireHours  int    `mapstructure:"jwt_expire_hours"`
	LogLevel        string `mapstructure:"log_level"`
	EnableCORS      bool   `mapstructure:"enable_cors"`
	EnableRateLimit bool   `mapstructure:"enable_rate_limit"`
	EnableWebSocket bool   `mapstructure:"enable_websocket"`
	MaxFileSize     int64  `mapstructure:"max_file_size"` // 文件上传大小限制（字节）
}

// EmailConfig 邮件配置
type EmailConfig struct {
	SMTPHost          string `mapstructure:"smtp_host"`
	SMTPPort          int    `mapstructure:"smtp_port"` // 改为 int 类型
	Username          string `mapstructure:"username"`
	Password          string `mapstructure:"password"`
	FromEmail         string `mapstructure:"from_email"`
	FromName          string `mapstructure:"from_name"`
	TemplateDir       string `mapstructure:"template_dir"`
	EnableTLS         bool   `mapstructure:"enable_tls"`
	ConnectionTimeout int    `mapstructure:"connection_timeout"`
	SendTimeout       int    `mapstructure:"send_timeout"`
	MaxRetries        int    `mapstructure:"max_retries"`
	RetryInterval     int    `mapstructure:"retry_interval"`
}

// WebSocketConfig WebSocket 配置
type WebSocketConfig struct {
	Enable          bool   `mapstructure:"enable"`
	Path            string `mapstructure:"path"`
	ReadBufferSize  int    `mapstructure:"read_buffer_size"`
	WriteBufferSize int    `mapstructure:"write_buffer_size"`
	MaxMessageSize  int64  `mapstructure:"max_message_size"`
	PingPeriod      int    `mapstructure:"ping_period"` // 秒
	PongWait        int    `mapstructure:"pong_wait"`   // 秒
	WriteWait       int    `mapstructure:"write_wait"`  // 秒
	MaxConnections  int    `mapstructure:"max_connections"`
}

// OAuthConfig OAuth 认证配置
type OAuthConfig struct {
	GitHub GitHubOAuthConfig `mapstructure:"github"`
	Google GoogleOAuthConfig `mapstructure:"google"`
	WeChat WeChatOAuthConfig `mapstructure:"wechat"`
}

// GitHubOAuthConfig GitHub OAuth 配置
type GitHubOAuthConfig struct {
	ClientID       string   `mapstructure:"client_id"`
	ClientSecret   string   `mapstructure:"client_secret"`
	RedirectURL    string   `mapstructure:"redirect_url"`
	APICallbackURL string   `mapstructure:"api_callback_url"`
	Scopes         []string `mapstructure:"scopes"`
	APIBaseURL     string   `mapstructure:"api_base_url"`
	AuthURL        string   `mapstructure:"auth_url"`
	TokenURL       string   `mapstructure:"token_url"`
	Timeout        int      `mapstructure:"timeout"`
}

// GoogleOAuthConfig Google OAuth 配置
type GoogleOAuthConfig struct {
	Enabled      bool   `mapstructure:"enabled"`
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	RedirectURL  string `mapstructure:"redirect_url"`
}

// WeChatOAuthConfig 微信 OAuth 配置
type WeChatOAuthConfig struct {
	Enabled     bool   `mapstructure:"enabled"`
	AppID       string `mapstructure:"app_id"`
	AppSecret   string `mapstructure:"app_secret"`
	RedirectURL string `mapstructure:"redirect_url"`
}

// AuthConfig 认证相关配置
type AuthConfig struct {
	VerificationCode VerificationCodeConfig `mapstructure:"verification_code"`
	Session          SessionConfig          `mapstructure:"session"`
	Security         SecurityConfig         `mapstructure:"security"`
}

// VerificationCodeConfig 验证码配置
type VerificationCodeConfig struct {
	Length              int `mapstructure:"length"`
	ExpireMinutes       int `mapstructure:"expire_minutes"`
	SendIntervalSeconds int `mapstructure:"send_interval_seconds"`
	MaxAttempts         int `mapstructure:"max_attempts"`
}

// SessionConfig 会话配置
type SessionConfig struct {
	MaxSessionsPerUser   int  `mapstructure:"max_sessions_per_user"`
	CleanupIntervalHours int  `mapstructure:"cleanup_interval_hours"`
	ExtendOnUse          bool `mapstructure:"extend_on_use"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	PasswordMinLength      int  `mapstructure:"password_min_length"`
	PasswordRequireSpecial bool `mapstructure:"password_require_special"`
	MaxLoginAttempts       int  `mapstructure:"max_login_attempts"`
	LockoutDurationMinutes int  `mapstructure:"lockout_duration_minutes"`
	Enable2FA              bool `mapstructure:"enable_2fa"`
}

// LoadConfig 加载配置文件
func LoadConfig(path string) (*Config, error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// 设置环境变量前缀
	viper.SetEnvPrefix("REFATOR_SIWU")
	viper.AutomaticEnv()

	// 设置默认值
	setDefaults()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: Could not read config file: %v", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unable to decode config: %v", err)
	}

	return &config, nil
}

// setDefaults 设置默认配置值
func setDefaults() {
	// Server defaults
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.mode", "debug")

	// Database defaults
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", "3306")
	viper.SetDefault("database.username", "root")
	viper.SetDefault("database.password", "")
	viper.SetDefault("database.dbname", "refator_siwu")
	viper.SetDefault("database.charset", "utf8mb4")

	// Redis defaults
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", "6379")
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)

	// App defaults
	viper.SetDefault("app.name", "InkwaveDocNet")
	viper.SetDefault("app.version", "1.0.0")
	viper.SetDefault("app.context_timeout", 30)
	viper.SetDefault("app.jwt_secret", "your-secret-key-change-in-production")
	viper.SetDefault("app.jwt_expire_hours", 24)
	viper.SetDefault("app.log_level", "info")
	viper.SetDefault("app.enable_cors", true)
	viper.SetDefault("app.enable_rate_limit", true)
	viper.SetDefault("app.enable_websocket", true)
	viper.SetDefault("app.max_file_size", 10485760) // 10MB

	// Email defaults
	viper.SetDefault("email.smtp_host", "smtp.gmail.com")
	viper.SetDefault("email.smtp_port", "587")
	viper.SetDefault("email.username", "")
	viper.SetDefault("email.password", "")
	viper.SetDefault("email.from_email", "noreply@example.com")
	viper.SetDefault("email.from_name", "墨协")
	viper.SetDefault("email.template_dir", "./templates/email")

	// WebSocket defaults
	viper.SetDefault("websocket.enable", true)
	viper.SetDefault("websocket.path", "/ws")
	viper.SetDefault("websocket.read_buffer_size", 1024)
	viper.SetDefault("websocket.write_buffer_size", 1024)
	viper.SetDefault("websocket.max_message_size", 512)
	viper.SetDefault("websocket.ping_period", 54) // 54秒
	viper.SetDefault("websocket.pong_wait", 60)   // 60秒
	viper.SetDefault("websocket.write_wait", 10)  // 10秒
	viper.SetDefault("websocket.max_connections", 1000)

	// OAuth defaults
	// GitHub OAuth
	viper.SetDefault("oauth.github.client_id", "")
	viper.SetDefault("oauth.github.client_secret", "")
	viper.SetDefault("oauth.github.redirect_url", "http://localhost:3000/auth/callback")
	viper.SetDefault("oauth.github.api_callback_url", "http://localhost:8080/api/v1/auth/github/callback")
	viper.SetDefault("oauth.github.scopes", []string{"user:email", "read:user"})
	viper.SetDefault("oauth.github.api_base_url", "https://api.github.com")
	viper.SetDefault("oauth.github.auth_url", "https://github.com/login/oauth/authorize")
	viper.SetDefault("oauth.github.token_url", "https://github.com/login/oauth/access_token")
	viper.SetDefault("oauth.github.timeout", 30)

	// Google OAuth
	viper.SetDefault("oauth.google.enabled", false)
	viper.SetDefault("oauth.google.client_id", "")
	viper.SetDefault("oauth.google.client_secret", "")
	viper.SetDefault("oauth.google.redirect_url", "http://localhost:3000/auth/google/callback")

	// WeChat OAuth
	viper.SetDefault("oauth.wechat.enabled", false)
	viper.SetDefault("oauth.wechat.app_id", "")
	viper.SetDefault("oauth.wechat.app_secret", "")
	viper.SetDefault("oauth.wechat.redirect_url", "http://localhost:3000/auth/wechat/callback")

	// Auth defaults
	// 验证码配置
	viper.SetDefault("auth.verification_code.length", 6)
	viper.SetDefault("auth.verification_code.expire_minutes", 10)
	viper.SetDefault("auth.verification_code.send_interval_seconds", 60)
	viper.SetDefault("auth.verification_code.max_attempts", 5)

	// 会话配置
	viper.SetDefault("auth.session.max_sessions_per_user", 5)
	viper.SetDefault("auth.session.cleanup_interval_hours", 24)
	viper.SetDefault("auth.session.extend_on_use", true)

	// 安全配置
	viper.SetDefault("auth.security.password_min_length", 8)
	viper.SetDefault("auth.security.password_require_special", true)
	viper.SetDefault("auth.security.max_login_attempts", 5)
	viper.SetDefault("auth.security.lockout_duration_minutes", 30)
	viper.SetDefault("auth.security.enable_2fa", false)
}

// GetDSN 获取数据库连接字符串
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local",
		c.Username,
		c.Password,
		c.Host,
		c.Port,
		c.DBName,
		c.Charset,
	)
}

// GetRedisAddr 获取 Redis 连接地址
func (c *RedisConfig) GetRedisAddr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}
