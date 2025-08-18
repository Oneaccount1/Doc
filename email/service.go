package email

import (
	"DOC/pkg/utils"
	"context"
	"fmt"
	"time"

	"DOC/domain"
)

// emailService 邮件业务服务实现
// 实现 domain.EmailUsecase 接口，包含邮件相关的所有业务逻辑
// 这一层负责协调邮件实体、仓储接口和邮件发送服务
type emailService struct {
	emailRepo      domain.EmailRepository // 邮件仓储接口
	emailSender    domain.EmailSender     // 邮件发送服务接口
	contextTimeout time.Duration          // 上下文超时时间
}

// NewEmailService 创建新的邮件服务实例
// 这是依赖注入的入口点，外部通过这个函数创建服务实例
// 参数说明：
// - emailRepo: 邮件仓储接口实现
// - emailSender: 邮件发送服务实现
// - timeout: 操作超时时间
// 返回值：
// - domain.EmailUsecase: 邮件业务逻辑接口
func NewEmailService(emailRepo domain.EmailRepository, emailSender domain.EmailSender, timeout time.Duration) domain.EmailUsecase {
	return &emailService{
		emailRepo:      emailRepo,
		emailSender:    emailSender,
		contextTimeout: timeout,
	}
}

// SendEmail 发送邮件
// 实现邮件发送的完整业务流程
// 业务流程：
// 1. 验证邮件实体
// 2. 保存邮件到数据库
// 3. 标记为发送中
// 4. 调用邮件发送服务
// 5. 更新发送状态
func (s *emailService) SendEmail(ctx context.Context, email *domain.Email) error {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	// 1. 验证邮件实体
	if err := email.Validate(); err != nil {
		return fmt.Errorf("邮件验证失败: %w", err)
	}

	// 2. 设置默认值
	if email.Status == 0 {
		email.Status = domain.EmailStatusPending
	}
	if email.Priority == 0 {
		email.Priority = domain.EmailPriorityNormal
	}
	if email.MaxRetries == 0 {
		email.MaxRetries = 3
	}

	// 3. 保存邮件到数据库
	if err := s.emailRepo.Store(ctx, email); err != nil {
		return fmt.Errorf("保存邮件失败: %w", err)
	}

	// 4. 异步发送邮件
	go s.sendEmailAsync(email)

	return nil
}

// SendVerificationEmail 发送验证码邮件
// 发送用户注册或登录验证码邮件
func (s *emailService) SendVerificationEmail(ctx context.Context, to, code string) error {
	var subject, template string
	subject = "验证码"
	template = "verification_default"
	//
	//switch codeType {
	//case domain.VerificationCodeTypeRegister:
	//	subject = "欢迎注册 - 验证码"
	//	template = "verification_register"
	//case domain.VerificationCodeTypeLogin:
	//	subject = "登录验证码"
	//	template = "verification_login"
	//default:
	//	subject = "验证码"
	//	template = "verification_default"
	//}
	data := utils.JSONMap{
		"code": code,
	}

	email := &domain.Email{
		To:         to,
		Subject:    subject,
		Data:       data,
		Template:   template,
		Type:       domain.EmailTypeVerification,
		Status:     domain.EmailStatusPending,
		Priority:   domain.EmailPriorityHigh, // 验证码邮件高优先级
		MaxRetries: 3,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	return s.SendEmail(ctx, email)
}

// SendWelcomeEmail 发送欢迎邮件
// 用户注册成功后发送欢迎邮件
func (s *emailService) SendWelcomeEmail(ctx context.Context, to, username string) error {
	data := utils.JSONMap{
		"username": username,
	}
	email := &domain.Email{
		To:         to,
		Subject:    "欢迎加入墨协！",
		Template:   "welcome",
		Data:       data,
		Type:       domain.EmailTypeWelcome,
		Status:     domain.EmailStatusPending,
		Priority:   domain.EmailPriorityNormal,
		MaxRetries: 3,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	return s.SendEmail(ctx, email)
}

// SendNotificationEmail 发送通知邮件
// 发送系统通知邮件
func (s *emailService) SendNotificationEmail(ctx context.Context, to, subject, content string) error {
	data := utils.JSONMap{
		"content": content,
	}
	email := &domain.Email{
		To:         to,
		Subject:    subject,
		Data:       data,
		Type:       domain.EmailTypeNotification,
		Status:     domain.EmailStatusPending,
		Priority:   domain.EmailPriorityNormal,
		MaxRetries: 3,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	return s.SendEmail(ctx, email)
}

// SendSystemEmail 发送系统邮件
// 发送系统级别的重要邮件
func (s *emailService) SendSystemEmail(ctx context.Context, to, subject, content string) error {
	data := utils.JSONMap{
		"content": content,
	}
	email := &domain.Email{
		To:         to,
		Subject:    subject,
		Data:       data,
		Type:       domain.EmailTypeSystem,
		Status:     domain.EmailStatusPending,
		Priority:   domain.EmailPriorityHigh, // 系统邮件高优先级
		MaxRetries: 5,                        // 系统邮件更多重试次数
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	return s.SendEmail(ctx, email)
}

// SendOrganizationInvitationEmail 发送组织邀请邮件
func (s *emailService) SendOrganizationInvitationEmail(ctx context.Context, to, subject string, data utils.JSONMap) error {

	emailData := &domain.Email{
		To:         to,
		Subject:    subject,
		Template:   "organization_invitation",
		Data:       data,
		Type:       domain.EmailTypeInvitation,
		Status:     domain.EmailStatusPending,
		Priority:   domain.EmailPriorityHigh,
		MaxRetries: 2,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	return s.SendEmail(ctx, emailData)
}

// GetEmailByID 根据ID获取邮件
func (s *emailService) GetEmailByID(ctx context.Context, id int64) (*domain.Email, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	return s.emailRepo.GetByID(ctx, id)
}

// GetEmailsByStatus 获取指定状态的邮件列表
func (s *emailService) GetEmailsByStatus(ctx context.Context, status domain.EmailStatus, offset, limit int) ([]*domain.Email, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	return s.emailRepo.ListByStatus(ctx, status, offset, limit)
}

// RetryFailedEmails 重试失败的邮件
// 查找可重试的失败邮件并重新发送
func (s *emailService) RetryFailedEmails(ctx context.Context, limit int) error {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	// 获取失败的邮件
	failedEmails, err := s.emailRepo.ListFailedEmails(ctx, limit)
	if err != nil {
		return fmt.Errorf("获取失败邮件列表失败: %w", err)
	}

	for _, email := range failedEmails {
		if email.CanRetry() {
			// 增加重试次数
			email.IncrementRetry()
			if err := s.emailRepo.Update(ctx, email); err != nil {
				fmt.Printf("更新邮件重试状态失败: %v\n", err)
				continue
			}

			// 异步重新发送
			go s.sendEmailAsync(email)
		}
	}

	return nil
}

// UpdateEmailStatus 更新邮件状态
func (s *emailService) UpdateEmailStatus(ctx context.Context, id int64, status domain.EmailStatus, reason string) error {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	email, err := s.emailRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("获取邮件失败: %w", err)
	}

	switch status {
	case domain.EmailStatusSent:
		email.MarkAsSent()
	case domain.EmailStatusFailed:
		email.MarkAsFailed(reason)
	case domain.EmailStatusSending:
		email.MarkAsSending()
	default:
		email.Status = status
		email.UpdatedAt = time.Now()
	}

	return s.emailRepo.Update(ctx, email)
}

// GetEmailStats 获取邮件统计信息
func (s *emailService) GetEmailStats(ctx context.Context) (*domain.EmailStats, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	stats := &domain.EmailStats{}

	// 获取各状态邮件数量
	// todo 协程优化
	sent, err := s.emailRepo.CountByStatus(ctx, domain.EmailStatusSent)
	if err != nil {
		return nil, fmt.Errorf("获取已发送邮件数量失败: %w", err)
	}
	stats.TotalSent = sent

	failed, err := s.emailRepo.CountByStatus(ctx, domain.EmailStatusFailed)
	if err != nil {
		return nil, fmt.Errorf("获取失败邮件数量失败: %w", err)
	}
	stats.TotalFailed = failed

	pending, err := s.emailRepo.CountByStatus(ctx, domain.EmailStatusPending)
	if err != nil {
		return nil, fmt.Errorf("获取待发送邮件数量失败: %w", err)
	}
	stats.TotalPending = pending

	retrying, err := s.emailRepo.CountByStatus(ctx, domain.EmailStatusRetrying)
	if err != nil {
		return nil, fmt.Errorf("获取重试中邮件数量失败: %w", err)
	}
	stats.TotalRetrying = retrying

	// 计算成功率
	total := stats.TotalSent + stats.TotalFailed
	if total > 0 {
		stats.SuccessRate = float64(stats.TotalSent) / float64(total) * 100
	}

	return stats, nil
}

// CleanupOldEmails 清理旧邮件
// 删除指定天数之前的邮件记录
func (s *emailService) CleanupOldEmails(ctx context.Context, olderThanDays int) error {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	cutoffTime := time.Now().AddDate(0, 0, -olderThanDays)
	return s.emailRepo.DeleteOldEmails(ctx, cutoffTime)
}

// sendEmailAsync 异步发送邮件
// 在后台协程中执行实际的邮件发送操作
func (s *emailService) sendEmailAsync(email *domain.Email) {
	ctx := context.Background()

	// 标记为发送中
	email.MarkAsSending()
	if err := s.emailRepo.Update(ctx, email); err != nil {
		fmt.Printf("更新邮件状态失败: %v\n", err)
		return
	}

	// 发送邮件
	if err := s.emailSender.Send(ctx, email); err != nil {
		// 发送失败，标记失败状态
		email.MarkAsFailed(err.Error())
		if updateErr := s.emailRepo.Update(ctx, email); updateErr != nil {
			fmt.Printf("更新邮件失败状态失败: %v\n", updateErr)
		}
		return
	}

	// 发送成功，标记成功状态
	email.MarkAsSent()
	if err := s.emailRepo.Update(ctx, email); err != nil {
		fmt.Printf("更新邮件成功状态失败: %v\n", err)
	}
}
