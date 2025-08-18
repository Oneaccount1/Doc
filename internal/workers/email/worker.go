package email

import (
	"context"
	"log"
	"sync"
	"time"

	"DOC/domain"
)

// EmailWorker 邮件工作者 - 简化版本
type EmailWorker struct {
	emailRepo   domain.EmailRepository
	emailSender domain.EmailSender

	// 基本配置
	workerCount  int           // 工作协程数量
	pollInterval time.Duration // 轮询间隔

	// 控制
	stopCh  chan struct{}
	running bool
	mu      sync.RWMutex
	wg      sync.WaitGroup
}

// WorkerConfig 工作者配置
type WorkerConfig struct {
	WorkerCount  int           `json:"worker_count"`  // 工作协程数量，默认2
	PollInterval time.Duration `json:"poll_interval"` // 轮询间隔，默认30秒
}

// NewEmailWorker 创建新的邮件工作者
func NewEmailWorker(emailRepo domain.EmailRepository, emailSender domain.EmailSender, config WorkerConfig) *EmailWorker {
	// 设置默认值
	if config.WorkerCount <= 0 {
		config.WorkerCount = 2
	}
	if config.PollInterval <= 0 {
		config.PollInterval = 30 * time.Second
	}

	return &EmailWorker{
		emailRepo:    emailRepo,
		emailSender:  emailSender,
		workerCount:  config.WorkerCount,
		pollInterval: config.PollInterval,
		stopCh:       make(chan struct{}),
	}
}

// Start 启动邮件工作者
func (w *EmailWorker) Start() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.running {
		return nil // 已经在运行，直接返回
	}

	w.running = true
	log.Printf("启动邮件工作者，工作协程数: %d", w.workerCount)

	// 启动工作协程
	for i := 0; i < w.workerCount; i++ {
		w.wg.Add(1)
		go w.worker(i)
	}

	return nil
}

// Stop 停止邮件工作者
func (w *EmailWorker) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.running {
		return // 没有在运行，直接返回
	}

	log.Println("停止邮件工作者...")
	close(w.stopCh)
	w.wg.Wait()
	w.running = false
	log.Println("邮件工作者已停止")
}

// worker 工作协程 - 简化版本
func (w *EmailWorker) worker(id int) {
	defer w.wg.Done()

	log.Printf("邮件工作协程 %d 启动", id)
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopCh:
			log.Printf("邮件工作协程 %d 停止", id)
			return
		case <-ticker.C:
			w.processEmails(id)
		}
	}
}

// processEmails 处理邮件 - 简化版本
func (w *EmailWorker) processEmails(workerID int) {
	ctx := context.Background()

	// 获取待发送邮件（每次只取5封）
	emails, err := w.emailRepo.ListPendingEmails(ctx, 5)
	if err != nil {
		log.Printf("工作协程 %d 获取待发送邮件失败: %v", workerID, err)
		return
	}

	if len(emails) == 0 {
		return // 没有待发送邮件
	}

	log.Printf("工作协程 %d 发现 %d 封待发送邮件", workerID, len(emails))

	// 逐个处理邮件
	for _, email := range emails {
		w.sendEmail(email, workerID)
	}
}

// sendEmail 发送单个邮件 - 简化版本
func (w *EmailWorker) sendEmail(email *domain.Email, workerID int) {
	ctx := context.Background()

	log.Printf("工作协程 %d 发送邮件 ID: %d, 收件人: %s", workerID, email.ID, email.To)

	// 标记为发送中
	email.MarkAsSending()
	if err := w.emailRepo.Update(ctx, email); err != nil {
		log.Printf("更新邮件状态失败: %v", err)
		return
	}

	// 发送邮件
	if err := w.emailSender.Send(ctx, email); err != nil {
		log.Printf("发送邮件失败 ID: %d, 错误: %v", email.ID, err)

		// 标记为失败
		email.MarkAsFailed(err.Error())
		w.emailRepo.Update(ctx, email)
		return
	}

	// 标记为成功
	email.MarkAsSent()
	if err := w.emailRepo.Update(ctx, email); err != nil {
		log.Printf("更新邮件状态失败: %v", err)
	} else {
		log.Printf("邮件发送成功 ID: %d", email.ID)
	}
}
