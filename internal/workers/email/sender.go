package email

import (
	"DOC/config"
	"context"
	"crypto/tls"
	"fmt"
	"html/template"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
	"time"

	"DOC/domain"
)

// emailSender 邮件发送器实现
// 实现 domain.EmailSender 接口，负责实际的邮件发送操作
type emailSender struct {
	config    config.EmailConfig
	templates map[string]*template.Template
}

// NewEmailSender 创建新的邮件发送器
func NewEmailSender(config config.EmailConfig) (domain.EmailSender, error) {
	sender := &emailSender{
		config:    config,
		templates: make(map[string]*template.Template),
	}

	// 加载邮件模板
	if err := sender.loadTemplates(); err != nil {
		return nil, fmt.Errorf("加载邮件模板失败: %w", err)
	}

	return sender, nil
}

// loadTemplates 加载邮件模板 - 直接读取模板目录下的所有HTML文件
func (s *emailSender) loadTemplates() error {
	if s.config.TemplateDir == "" {
		return nil // 如果没有配置模板目录，跳过模板加载
	}

	// 检查模板目录是否存在
	if _, err := os.Stat(s.config.TemplateDir); os.IsNotExist(err) {
		return fmt.Errorf("模板目录不存在: %s", s.config.TemplateDir)
	}

	// 读取模板目录下的所有HTML文件
	pattern := filepath.Join(s.config.TemplateDir, "*.html")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("读取模板文件失败: %w", err)
	}

	// 加载每个模板文件
	for _, file := range files {
		// 获取文件名（不包含扩展名）作为模板名
		templateName := strings.TrimSuffix(filepath.Base(file), ".html")

		tmpl, err := template.ParseFiles(file)
		if err != nil {
			return fmt.Errorf("解析模板文件 %s 失败: %w", file, err)
		}

		s.templates[templateName] = tmpl
	}

	return nil
}

// Send 发送邮件
func (s *emailSender) Send(ctx context.Context, email *domain.Email) error {
	// 构建邮件内容
	content, err := s.buildEmailContent(email)
	if err != nil {
		return fmt.Errorf("构建邮件内容失败: %w", err)
	}

	// 发送邮件
	return s.sendSMTP(ctx, email.To, email.Subject, content)
}

// SendBatch 批量发送邮件
func (s *emailSender) SendBatch(ctx context.Context, emails []*domain.Email) error {
	for _, email := range emails {
		if err := s.Send(ctx, email); err != nil {
			return fmt.Errorf("批量发送邮件失败，邮件ID %d: %w", email.ID, err)
		}
	}
	return nil
}

// HealthCheck 健康检查
func (s *emailSender) HealthCheck(ctx context.Context) error {
	// 创建带超时的上下文
	checkCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	addr := fmt.Sprintf("%s:%d", s.config.SMTPHost, s.config.SMTPPort)
	auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.SMTPHost)

	// 在goroutine中进行健康检查
	errCh := make(chan error, 1)
	go func() {
		client, err := smtp.Dial(addr)
		if err != nil {
			errCh <- fmt.Errorf("连接SMTP服务器失败: %w", err)
			return
		}
		defer client.Close()

		// 启用TLS
		if err := client.StartTLS(&tls.Config{ServerName: s.config.SMTPHost}); err != nil {
			errCh <- fmt.Errorf("启用TLS失败: %w", err)
			return
		}

		// 认证
		if err := client.Auth(auth); err != nil {
			errCh <- fmt.Errorf("SMTP认证失败: %w", err)
			return
		}

		errCh <- nil
	}()

	// 等待健康检查完成或超时
	select {
	case <-checkCtx.Done():
		return fmt.Errorf("健康检查超时: %w", checkCtx.Err())
	case err := <-errCh:
		return err
	}
}

// buildEmailContent 构建邮件内容
func (s *emailSender) buildEmailContent(email *domain.Email) (string, error) {

	// 如果有模板，渲染模板
	if email.Template != "" {
		return s.renderTemplate(email.Template, email.Data)
	}

	return "", fmt.Errorf("邮件内容和模板都为空")
}

// renderTemplate 渲染模板
func (s *emailSender) renderTemplate(templateName string, data map[string]interface{}) (string, error) {
	tmpl, exists := s.templates[templateName]
	if !exists {
		return "", fmt.Errorf("模板 %s 不存在", templateName)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("执行模板失败: %w", err)
	}

	return buf.String(), nil
}

// sendSMTP 通过SMTP发送邮件
func (s *emailSender) sendSMTP(ctx context.Context, to, subject, content string) error {
	// 创建带超时的上下文
	sendCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// 构建邮件头
	from := fmt.Sprintf("%s <%s>", s.config.FromName, s.config.FromEmail)
	headers := map[string]string{
		"From":         from,
		"To":           to,
		"Subject":      subject,
		"MIME-Version": "1.0",
		"Content-Type": "text/html; charset=UTF-8",
		"Date":         time.Now().Format(time.RFC1123Z),
		"Message-ID":   fmt.Sprintf("<%d@%s>", time.Now().UnixNano(), s.config.SMTPHost),
	}

	// 构建邮件消息
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + content

	// SMTP配置
	addr := fmt.Sprintf("%s:%d", s.config.SMTPHost, s.config.SMTPPort)

	// 在goroutine中发送邮件，支持超时控制
	errCh := make(chan error, 1)
	go func() {
		// 使用更详细的SMTP发送逻辑
		err := s.sendWithDetailedSMTP(addr, s.config.Username, s.config.Password, s.config.FromEmail, []string{to}, []byte(message))
		errCh <- err
	}()

	// 等待发送完成或超时
	select {
	case <-sendCtx.Done():
		return fmt.Errorf("邮件发送超时: %w", sendCtx.Err())
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("发送邮件失败: %w", err)
		}
		return nil
	}
}

// sendWithDetailedSMTP 使用更详细的SMTP发送逻辑，提供更好的错误处理
func (s *emailSender) sendWithDetailedSMTP(addr, username, password, from string, to []string, msg []byte) error {
	var client *smtp.Client
	var err error

	// 根据端口判断是使用SSL还是TLS
	if s.config.SMTPPort == 465 {
		// 465端口使用SSL直连
		tlsConfig := &tls.Config{
			ServerName:         s.config.SMTPHost,
			InsecureSkipVerify: false,
		}

		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("SSL连接SMTP服务器失败: %w", err)
		}

		client, err = smtp.NewClient(conn, s.config.SMTPHost)
		if err != nil {
			conn.Close()
			return fmt.Errorf("创建SMTP客户端失败: %w", err)
		}
	} else {
		// 587端口或其他端口使用普通连接
		client, err = smtp.Dial(addr)
		if err != nil {
			return fmt.Errorf("连接SMTP服务器失败: %w", err)
		}

		// 启用STARTTLS（如果配置启用且不是465端口）
		if s.config.EnableTLS {
			tlsConfig := &tls.Config{
				ServerName:         s.config.SMTPHost,
				InsecureSkipVerify: false,
			}
			if err := client.StartTLS(tlsConfig); err != nil {
				client.Close()
				return fmt.Errorf("启用TLS失败: %w", err)
			}
		}
	}

	defer client.Close()

	// 认证
	auth := smtp.PlainAuth("", username, password, s.config.SMTPHost)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP认证失败: %w", err)
	}

	// 设置发件人
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("设置发件人失败: %w", err)
	}

	// 设置收件人
	for _, recipient := range to {
		if err := client.Rcpt(recipient); err != nil {
			return fmt.Errorf("设置收件人 %s 失败: %w", recipient, err)
		}
	}

	// 发送邮件内容
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("获取邮件写入器失败: %w", err)
	}
	defer writer.Close()

	if _, err := writer.Write(msg); err != nil {
		return fmt.Errorf("写入邮件内容失败: %w", err)
	}

	return nil
}
