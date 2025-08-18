package mysql

import (
	"context"
	"errors"
	"time"

	"DOC/domain"

	"gorm.io/gorm"
)

// emailRepository MySQL邮件仓储实现
// 实现 domain.EmailRepository 接口
type emailRepository struct {
	db *gorm.DB
}

// NewEmailRepository 创建新的邮件仓储实例
func NewEmailRepository(db *gorm.DB) domain.EmailRepository {
	return &emailRepository{
		db: db,
	}
}

// Store 保存邮件
func (r *emailRepository) Store(ctx context.Context, email *domain.Email) error {
	if err := r.db.WithContext(ctx).Create(email).Error; err != nil {
		// 数据库层面的错误直接返回
		return err
	}
	return nil
}

// GetByID 根据ID获取邮件
func (r *emailRepository) GetByID(ctx context.Context, id int64) (*domain.Email, error) {
	var email domain.Email
	if err := r.db.WithContext(ctx).First(&email, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrEmailNotFound
		}
		return nil, err
	}
	return &email, nil
}

// Update 更新邮件
func (r *emailRepository) Update(ctx context.Context, email *domain.Email) error {
	email.UpdatedAt = time.Now()
	if err := r.db.WithContext(ctx).Save(email).Error; err != nil {
		return err
	}
	return nil
}

// Delete 删除邮件
func (r *emailRepository) Delete(ctx context.Context, id int64) error {
	if err := r.db.WithContext(ctx).Delete(&domain.Email{}, id).Error; err != nil {
		return err
	}
	return nil
}

// ListByRecipientID 根据接收者ID获取邮件列表
func (r *emailRepository) ListByRecipientID(ctx context.Context, recipientID int64, offset, limit int) ([]*domain.Email, error) {
	var emails []*domain.Email
	if err := r.db.WithContext(ctx).
		Where("recipient_id = ?", recipientID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&emails).Error; err != nil {
		return nil, err
	}
	return emails, nil
}

// ListByStatus 根据状态获取邮件列表
func (r *emailRepository) ListByStatus(ctx context.Context, status domain.EmailStatus, offset, limit int) ([]*domain.Email, error) {
	var emails []*domain.Email
	if err := r.db.WithContext(ctx).
		Where("status = ?", status).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&emails).Error; err != nil {
		return nil, err
	}
	return emails, nil
}

// ListByType 根据类型获取邮件列表
func (r *emailRepository) ListByType(ctx context.Context, emailType domain.EmailType, offset, limit int) ([]*domain.Email, error) {
	var emails []*domain.Email
	if err := r.db.WithContext(ctx).
		Where("type = ?", emailType).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&emails).Error; err != nil {
		return nil, err
	}
	return emails, nil
}

// ListPendingEmails 获取待发送邮件列表
func (r *emailRepository) ListPendingEmails(ctx context.Context, limit int) ([]*domain.Email, error) {
	var emails []*domain.Email
	if err := r.db.WithContext(ctx).
		Where("status = ?", domain.EmailStatusPending).
		Order("priority ASC, created_at ASC"). // 按优先级和创建时间排序
		Limit(limit).
		Find(&emails).Error; err != nil {
		return nil, err
	}
	return emails, nil
}

// ListFailedEmails 获取失败邮件列表
func (r *emailRepository) ListFailedEmails(ctx context.Context, limit int) ([]*domain.Email, error) {
	var emails []*domain.Email
	if err := r.db.WithContext(ctx).
		Where("status = ? AND retry_count < max_retries", domain.EmailStatusFailed).
		Order("priority ASC, updated_at ASC"). // 按优先级和更新时间排序
		Limit(limit).
		Find(&emails).Error; err != nil {
		return nil, err
	}
	return emails, nil
}

// CountByStatus 根据状态统计邮件数量
func (r *emailRepository) CountByStatus(ctx context.Context, status domain.EmailStatus) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&domain.Email{}).
		Where("status = ?", status).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// CountByRecipientID 根据接收者ID统计邮件数量
func (r *emailRepository) CountByRecipientID(ctx context.Context, recipientID int64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&domain.Email{}).
		Where("recipient_id = ?", recipientID).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// BatchUpdateStatus 批量更新邮件状态
func (r *emailRepository) BatchUpdateStatus(ctx context.Context, ids []int64, status domain.EmailStatus) error {
	if len(ids) == 0 {
		return nil
	}

	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}

	if err := r.db.WithContext(ctx).
		Model(&domain.Email{}).
		Where("id IN ?", ids).
		Updates(updates).Error; err != nil {
		return err
	}
	return nil
}

// DeleteOldEmails 删除旧邮件
func (r *emailRepository) DeleteOldEmails(ctx context.Context, olderThan time.Time) error {
	if err := r.db.WithContext(ctx).
		Where("created_at < ?", olderThan).
		Delete(&domain.Email{}).Error; err != nil {
		return err
	}
	return nil
}

// GetEmailStatsByDateRange 获取指定日期范围内的邮件统计
func (r *emailRepository) GetEmailStatsByDateRange(ctx context.Context, startDate, endDate time.Time) (map[domain.EmailStatus]int64, error) {
	var results []struct {
		Status domain.EmailStatus `json:"status"`
		Count  int64              `json:"count"`
	}

	if err := r.db.WithContext(ctx).
		Model(&domain.Email{}).
		Select("status, COUNT(*) as count").
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Group("status").
		Find(&results).Error; err != nil {
		return nil, err
	}

	stats := make(map[domain.EmailStatus]int64)
	for _, result := range results {
		stats[result.Status] = result.Count
	}

	return stats, nil
}

// GetEmailStatsByType 获取按类型分组的邮件统计
func (r *emailRepository) GetEmailStatsByType(ctx context.Context) (map[domain.EmailType]int64, error) {
	var results []struct {
		Type  domain.EmailType `json:"type"`
		Count int64            `json:"count"`
	}

	if err := r.db.WithContext(ctx).
		Model(&domain.Email{}).
		Select("type, COUNT(*) as count").
		Group("type").
		Find(&results).Error; err != nil {
		return nil, err
	}

	stats := make(map[domain.EmailType]int64)
	for _, result := range results {
		stats[result.Type] = result.Count
	}

	return stats, nil
}

// GetRecentFailedEmails 获取最近失败的邮件
func (r *emailRepository) GetRecentFailedEmails(ctx context.Context, hours int, limit int) ([]*domain.Email, error) {
	cutoffTime := time.Now().Add(-time.Duration(hours) * time.Hour)

	var emails []*domain.Email
	if err := r.db.WithContext(ctx).
		Where("status = ? AND updated_at >= ?", domain.EmailStatusFailed, cutoffTime).
		Order("updated_at DESC").
		Limit(limit).
		Find(&emails).Error; err != nil {
		return nil, err
	}
	return emails, nil
}
