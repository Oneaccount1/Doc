package email

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"refatorSiwu/domain"
	"refatorSiwu/email/mocks"
)

// TestSendEmail 测试发送邮件功能
func TestSendEmail(t *testing.T) {
	// 创建模拟对象
	mockEmailRepo := new(mocks.EmailRepository)
	mockEmailSender := new(mocks.EmailService)

	// 创建邮件服务
	emailService := NewEmailService(mockEmailRepo, mockEmailSender, 30*time.Second)

	tests := []struct {
		name          string
		email         *domain.Email
		setupMocks    func()
		expectedError error
	}{
		{
			name: "成功发送邮件",
			email: &domain.Email{
				To:      "test@example.com",
				Subject: "测试邮件",
				Content: "这是一封测试邮件",
				Type:    domain.EmailTypeSystem,
			},
			setupMocks: func() {
				// 模拟保存邮件成功
				mockEmailRepo.On("Store", mock.Anything, mock.AnythingOfType("*domain.Email")).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "邮件验证失败",
			email: &domain.Email{
				To:      "", // 空邮箱地址
				Subject: "测试邮件",
				Content: "这是一封测试邮件",
				Type:    domain.EmailTypeSystem,
			},
			setupMocks:    func() {},
			expectedError: domain.ErrInvalidEmailAddress,
		},
		{
			name: "保存邮件失败",
			email: &domain.Email{
				To:      "test@example.com",
				Subject: "测试邮件",
				Content: "这是一封测试邮件",
				Type:    domain.EmailTypeSystem,
			},
			setupMocks: func() {
				// 模拟保存邮件失败
				mockEmailRepo.On("Store", mock.Anything, mock.AnythingOfType("*domain.Email")).Return(domain.ErrInternalServerError)
			},
			expectedError: domain.ErrInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 重置模拟对象
			mockEmailRepo.ExpectedCalls = nil
			mockEmailSender.ExpectedCalls = nil

			// 设置模拟
			tt.setupMocks()

			// 执行测试
			err := emailService.SendEmail(context.Background(), tt.email)

			// 验证结果
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			// 验证模拟调用
			mockEmailRepo.AssertExpectations(t)
		})
	}
}

// TestSendVerificationEmail 测试发送验证码邮件
func TestSendVerificationEmail(t *testing.T) {
	// 创建模拟对象
	mockEmailRepo := new(mocks.EmailRepository)
	mockEmailSender := new(mocks.EmailService)

	// 创建邮件服务
	emailService := NewEmailService(mockEmailRepo, mockEmailSender, 30*time.Second)

	tests := []struct {
		name          string
		to            string
		code          string
		codeType      domain.VerificationCodeType
		setupMocks    func()
		expectedError error
	}{
		{
			name:     "成功发送注册验证码",
			to:       "test@example.com",
			code:     "123456",
			codeType: domain.VerificationCodeTypeRegister,
			setupMocks: func() {
				mockEmailRepo.On("Store", mock.Anything, mock.AnythingOfType("*domain.Email")).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:     "成功发送登录验证码",
			to:       "test@example.com",
			code:     "654321",
			codeType: domain.VerificationCodeTypeLogin,
			setupMocks: func() {
				mockEmailRepo.On("Store", mock.Anything, mock.AnythingOfType("*domain.Email")).Return(nil)
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 重置模拟对象
			mockEmailRepo.ExpectedCalls = nil
			mockEmailSender.ExpectedCalls = nil

			// 设置模拟
			tt.setupMocks()

			// 执行测试
			err := emailService.SendVerificationEmail(context.Background(), tt.to, tt.code, tt.codeType)

			// 验证结果
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			// 验证模拟调用
			mockEmailRepo.AssertExpectations(t)
		})
	}
}

// TestSendWelcomeEmail 测试发送欢迎邮件
func TestSendWelcomeEmail(t *testing.T) {
	// 创建模拟对象
	mockEmailRepo := new(mocks.EmailRepository)
	mockEmailSender := new(mocks.EmailService)

	// 创建邮件服务
	emailService := NewEmailService(mockEmailRepo, mockEmailSender, 30*time.Second)

	// 设置模拟
	mockEmailRepo.On("Store", mock.Anything, mock.AnythingOfType("*domain.Email")).Return(nil)

	// 执行测试
	err := emailService.SendWelcomeEmail(context.Background(), "test@example.com", "testuser")

	// 验证结果
	assert.NoError(t, err)

	// 验证模拟调用
	mockEmailRepo.AssertExpectations(t)
}

// TestGetEmailStats 测试获取邮件统计
func TestGetEmailStats(t *testing.T) {
	// 创建模拟对象
	mockEmailRepo := new(mocks.EmailRepository)
	mockEmailSender := new(mocks.EmailService)

	// 创建邮件服务
	emailService := NewEmailService(mockEmailRepo, mockEmailSender, 30*time.Second)

	// 设置模拟
	mockEmailRepo.On("CountByStatus", mock.Anything, domain.EmailStatusSent).Return(int64(100), nil)
	mockEmailRepo.On("CountByStatus", mock.Anything, domain.EmailStatusFailed).Return(int64(10), nil)
	mockEmailRepo.On("CountByStatus", mock.Anything, domain.EmailStatusPending).Return(int64(5), nil)
	mockEmailRepo.On("CountByStatus", mock.Anything, domain.EmailStatusRetrying).Return(int64(2), nil)

	// 执行测试
	stats, err := emailService.GetEmailStats(context.Background())

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, int64(100), stats.TotalSent)
	assert.Equal(t, int64(10), stats.TotalFailed)
	assert.Equal(t, int64(5), stats.TotalPending)
	assert.Equal(t, int64(2), stats.TotalRetrying)
	assert.InDelta(t, 90.91, stats.SuccessRate, 0.01) // 100/(100+10)*100 ≈ 90.91

	// 验证模拟调用
	mockEmailRepo.AssertExpectations(t)
}

// TestRetryFailedEmails 测试重试失败邮件
func TestRetryFailedEmails(t *testing.T) {
	// 创建模拟对象
	mockEmailRepo := new(mocks.EmailRepository)
	mockEmailSender := new(mocks.EmailService)

	// 创建邮件服务
	emailService := NewEmailService(mockEmailRepo, mockEmailSender, 30*time.Second)

	// 创建失败邮件
	failedEmail := &domain.Email{
		ID:         1,
		To:         "test@example.com",
		Subject:    "测试邮件",
		Content:    "测试内容",
		Status:     domain.EmailStatusFailed,
		RetryCount: 1,
		MaxRetries: 3,
	}

	// 设置模拟
	mockEmailRepo.On("ListFailedEmails", mock.Anything, 10).Return([]*domain.Email{failedEmail}, nil)
	mockEmailRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Email")).Return(nil)

	// 执行测试
	err := emailService.RetryFailedEmails(context.Background(), 10)

	// 验证结果
	assert.NoError(t, err)

	// 验证模拟调用
	mockEmailRepo.AssertExpectations(t)
}
