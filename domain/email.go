package domain

import (
	"DOC/pkg/utils"
	"context"
	"time"
)

// Email 邮件实体 - 核心业务实体
type Email struct {
	ID      int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	To      string `json:"to" gorm:"type:varchar(255);not null;index"`
	Subject string `json:"subject" gorm:"type:varchar(500);not null"`

	Template string `json:"template" gorm:"type:varchar(100)"`
	// 模板数据
	Data utils.JSONMap `json:"data,omitempty" gorm:"type:json;"`

	Type     EmailType     `json:"type" gorm:"type:varchar(50);not null;index"`
	Status   EmailStatus   `json:"status" gorm:"type:tinyint;not null;default:0;index"`
	Priority EmailPriority `json:"priority" gorm:"type:tinyint;not null;default:2"`

	// 发送相关字段
	SentAt     *time.Time `json:"sent_at"`
	FailReason string     `json:"fail_reason" gorm:"type:text"`
	RetryCount int        `json:"retry_count" gorm:"default:0"`
	MaxRetries int        `json:"max_retries" gorm:"default:3"`

	// 时间戳
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// EmailType 邮件类型枚举
type EmailType string

const (
	EmailTypeVerification EmailType = "verification" // 验证码邮件
	EmailTypeWelcome      EmailType = "welcome"      // 欢迎邮件
	EmailTypeNotification EmailType = "notification" // 通知邮件
	EmailTypeMarketing    EmailType = "marketing"    // 营销邮件
	EmailTypeSystem       EmailType = "system"       // 系统邮件
	EmailTypeInvitation   EmailType = "invitation"   // 系统邮件
)

// EmailStatus 邮件状态枚举
type EmailStatus int

const (
	EmailStatusPending   EmailStatus = iota // 待发送
	EmailStatusSending                      // 发送中
	EmailStatusSent                         // 已发送
	EmailStatusFailed                       // 发送失败
	EmailStatusRetrying                     // 重试中
	EmailStatusCancelled                    // 已取消
)

// EmailPriority 邮件优先级
type EmailPriority int

const (
	EmailPriorityHigh   EmailPriority = 1 // 高优先级
	EmailPriorityNormal EmailPriority = 2 // 普通优先级
	EmailPriorityLow    EmailPriority = 3 // 低优先级
)

// Validate 验证邮件实体的业务规则
func (e *Email) Validate() error {
	if e.To == "" {
		return ErrInvalidEmailAddress
	}
	if e.Subject == "" {
		return ErrInvalidEmailSubject
	}
	if e.Template == "" {
		return ErrInvalidEmailContent
	}
	if !e.isValidType() {
		return ErrInvalidEmailType
	}
	return nil
}

// isValidType 验证邮件类型
func (e *Email) isValidType() bool {
	validTypes := []EmailType{
		EmailTypeVerification, EmailTypeWelcome, EmailTypeNotification,
		EmailTypeMarketing, EmailTypeSystem, EmailTypeInvitation,
	}
	for _, validType := range validTypes {
		if e.Type == validType {
			return true
		}
	}
	return false
}

// MarkAsSending 标记为发送中
func (e *Email) MarkAsSending() {
	e.Status = EmailStatusSending
	e.UpdatedAt = time.Now()
}

// MarkAsSent 标记为已发送
func (e *Email) MarkAsSent() {
	e.Status = EmailStatusSent
	now := time.Now()
	e.SentAt = &now
	e.UpdatedAt = now
}

// MarkAsFailed 标记为发送失败
func (e *Email) MarkAsFailed(reason string) {
	e.Status = EmailStatusFailed
	e.FailReason = reason
	e.UpdatedAt = time.Now()
}

// CanRetry 检查是否可以重试
func (e *Email) CanRetry() bool {
	return e.Status == EmailStatusFailed && e.RetryCount < e.MaxRetries
}

// IncrementRetry 增加重试次数
func (e *Email) IncrementRetry() {
	e.RetryCount++
	e.Status = EmailStatusRetrying
	e.UpdatedAt = time.Now()
}

// IsHighPriority 检查是否为高优先级
func (e *Email) IsHighPriority() bool {
	return e.Priority == EmailPriorityHigh
}

// IsPending 检查是否待发送
func (e *Email) IsPending() bool {
	return e.Status == EmailStatusPending
}

// IsSent 检查是否已发送
func (e *Email) IsSent() bool {
	return e.Status == EmailStatusSent
}

// EmailRepository 邮件仓储接口
type EmailRepository interface {
	// 基础CRUD操作
	Store(ctx context.Context, email *Email) error
	GetByID(ctx context.Context, id int64) (*Email, error)
	Update(ctx context.Context, email *Email) error
	Delete(ctx context.Context, id int64) error

	// 查询操作

	ListByStatus(ctx context.Context, status EmailStatus, offset, limit int) ([]*Email, error)
	ListByType(ctx context.Context, emailType EmailType, offset, limit int) ([]*Email, error)
	ListPendingEmails(ctx context.Context, limit int) ([]*Email, error)
	ListFailedEmails(ctx context.Context, limit int) ([]*Email, error)

	// 统计操作
	CountByStatus(ctx context.Context, status EmailStatus) (int64, error)

	// 批量操作
	BatchUpdateStatus(ctx context.Context, ids []int64, status EmailStatus) error
	DeleteOldEmails(ctx context.Context, olderThan time.Time) error
}

// EmailUsecase 邮件业务逻辑接口
// 定义邮件相关的所有业务用例，遵循Clean Architecture原则
type EmailUsecase interface {
	// 基础邮件操作
	SendEmail(ctx context.Context, email *Email) error

	// 特定类型邮件发送
	// todo检查对应模板 所需要的数据
	SendVerificationEmail(ctx context.Context, to, code string) error
	SendWelcomeEmail(ctx context.Context, to, username string) error
	SendNotificationEmail(ctx context.Context, to, subject, content string) error
	SendSystemEmail(ctx context.Context, to, subject, content string) error
	SendOrganizationInvitationEmail(ctx context.Context, to, subject string, data utils.JSONMap) error

	// 邮件管理
	GetEmailByID(ctx context.Context, id int64) (*Email, error)
	GetEmailsByStatus(ctx context.Context, status EmailStatus, offset, limit int) ([]*Email, error)

	// 重试和状态管理
	RetryFailedEmails(ctx context.Context, limit int) error
	UpdateEmailStatus(ctx context.Context, id int64, status EmailStatus, reason string) error

	// 统计和监控
	GetEmailStats(ctx context.Context) (*EmailStats, error)
	CleanupOldEmails(ctx context.Context, olderThanDays int) error
}

// EmailStats 邮件统计信息
type EmailStats struct {
	TotalSent     int64   `json:"total_sent"`
	TotalFailed   int64   `json:"total_failed"`
	TotalPending  int64   `json:"total_pending"`
	TotalRetrying int64   `json:"total_retrying"`
	SuccessRate   float64 `json:"success_rate"`
}

// EmailSender 邮件发送服务接口
// 负责实际的邮件发送操作，由基础设施层实现
type EmailSender interface {
	// 发送邮件
	Send(ctx context.Context, email *Email) error

	// 批量发送
	SendBatch(ctx context.Context, emails []*Email) error

	// 健康检查
	HealthCheck(ctx context.Context) error
}
