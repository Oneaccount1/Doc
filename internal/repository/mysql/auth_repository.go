package mysql

import (
	"DOC/domain"
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

// 数据库底层错误直接返回

type mysqlAuthRepository struct {
	db *gorm.DB
}

func NewMysqlAuthRepository(db *gorm.DB) domain.AuthRepository {
	return &mysqlAuthRepository{
		db: db,
	}
}

func (m *mysqlAuthRepository) GetSessionByRefreshToken(ctx context.Context, token string) (*domain.AuthSession, error) {
	var session domain.AuthSession

	if err := m.db.WithContext(ctx).
		Where("refresh_token = ? AND status = ?", token, domain.SessionStatusActive).
		First(&session).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrSessionNotFound
		}
		return nil, err
	}

	return &session, nil
}
func (m *mysqlAuthRepository) StoreSession(ctx context.Context, session *domain.AuthSession) error {
	// 设置创建和更新时间
	now := time.Now()
	session.CreatedAt = now
	session.LastUsedAt = now

	if err := m.db.WithContext(ctx).Create(session).Error; err != nil {
		return err
	}

	return nil
}

func (m *mysqlAuthRepository) GetSessionByToken(ctx context.Context, token string) (*domain.AuthSession, error) {
	var session domain.AuthSession

	if err := m.db.WithContext(ctx).
		Where("session_token = ? AND status = ?", token, domain.SessionStatusActive).
		First(&session).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrSessionNotFound
		}
		return nil, err
	}

	return &session, nil
}

func (m *mysqlAuthRepository) GetSessionsByUserID(ctx context.Context, userID int64) ([]*domain.AuthSession, error) {
	var sessions []*domain.AuthSession

	if err := m.db.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, domain.SessionStatusActive).
		Order("last_used_at DESC").
		Find(&sessions).Error; err != nil {
		return nil, err
	}

	return sessions, nil
}

func (m *mysqlAuthRepository) UpdateSession(ctx context.Context, session *domain.AuthSession) error {
	// 更新最后使用时间
	session.LastUsedAt = time.Now()

	if err := m.db.WithContext(ctx).Save(session).Error; err != nil {
		return err
	}

	return nil
}

func (m *mysqlAuthRepository) RevokeSession(ctx context.Context, sessionID int64) error {
	now := time.Now()

	result := m.db.WithContext(ctx).Model(&domain.AuthSession{}).
		Where("id = ? AND status = ?", sessionID, domain.SessionStatusActive).
		Updates(map[string]interface{}{
			"status":     domain.SessionStatusRevoked,
			"revoked_at": &now,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return domain.ErrSessionNotFound
	}

	return nil
}

func (m *mysqlAuthRepository) RevokeAllUserSessions(ctx context.Context, userID int64) error {
	now := time.Now()

	if err := m.db.WithContext(ctx).Model(&domain.AuthSession{}).
		Where("user_id = ? AND status = ?", userID, domain.SessionStatusActive).
		Updates(map[string]interface{}{
			"status":     domain.SessionStatusRevoked,
			"revoked_at": &now,
		}).Error; err != nil {
		return err
	}

	return nil
}

func (m *mysqlAuthRepository) StoreVerificationCode(ctx context.Context, code *domain.VerificationCode) error {
	// 设置创建时间
	code.CreatedAt = time.Now()

	if err := m.db.WithContext(ctx).Create(code).Error; err != nil {
		return err
	}

	return nil
}

func (m *mysqlAuthRepository) GetVerificationCode(ctx context.Context, email string) (*domain.VerificationCode, error) {
	var code domain.VerificationCode

	if err := m.db.WithContext(ctx).
		Where("email = ? AND used = ? AND expires_at > ?", email, false, time.Now()).
		Order("created_at DESC").
		First(&code).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrVerificationCodeNotFound
		}
		return nil, err
	}

	return &code, nil
}

func (m *mysqlAuthRepository) DeleteVerificationCode(ctx context.Context, id int64) error {
	// 标记为已使用而不是物理删除
	now := time.Now()

	result := m.db.WithContext(ctx).Model(&domain.VerificationCode{}).
		Where("id = ? AND used = ?", id, false).
		Updates(map[string]interface{}{
			"used":    true,
			"used_at": &now,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return domain.ErrVerificationCodeNotFound
	}

	return nil
}

func (m *mysqlAuthRepository) CleanupExpiredCodes(ctx context.Context) error {
	// 物理删除过期的验证码记录
	if err := m.db.WithContext(ctx).Where("expires_at < ?", time.Now()).Delete(&domain.VerificationCode{}).Error; err != nil {
		return err
	}

	return nil
}

func (m *mysqlAuthRepository) StoreOAuthState(ctx context.Context, state *domain.OAuthState) error {
	// 设置创建时间
	state.CreatedAt = time.Now()

	if err := m.db.WithContext(ctx).Create(state).Error; err != nil {
		return err
	}

	return nil
}

func (m *mysqlAuthRepository) GetOAuthState(ctx context.Context, state string) (*domain.OAuthState, error) {
	var oauthState domain.OAuthState

	if err := m.db.WithContext(ctx).Where("state = ? AND expires_at > ?", state, time.Now()).
		First(&oauthState).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrOAuthStateMismatch
		}
		return nil, err
	}

	return &oauthState, nil
}

func (m *mysqlAuthRepository) DeleteOAuthState(ctx context.Context, id int64) error {
	// 标记为已使用
	now := time.Now()

	result := m.db.WithContext(ctx).Model(&domain.OAuthState{}).
		Where("id = ? AND used_at IS NULL", id).
		Update("used_at", &now)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return domain.ErrOAuthStateMismatch
	}

	return nil
}

func (m *mysqlAuthRepository) CleanupExpiredStates(ctx context.Context) error {
	// 物理删除过期的OAuth状态记录
	if err := m.db.WithContext(ctx).Where("expires_at < ?", time.Now()).Delete(&domain.OAuthState{}).Error; err != nil {
		return err
	}

	return nil
}
