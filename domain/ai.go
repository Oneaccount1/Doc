package domain

import (
	"context"
	"time"
)

// AITaskType AI任务类型枚举
type AITaskType string

const (
	AITaskTypeTextCorrection    AITaskType = "text_correction"    // 文本纠错
	AITaskTypeDiagramGeneration AITaskType = "diagram_generation" // 图表生成
	AITaskTypeContentSummary    AITaskType = "content_summary"    // 内容摘要
	AITaskTypeTranslation       AITaskType = "translation"        // 翻译
)

// AITaskStatus AI任务状态枚举
type AITaskStatus int

const (
	AITaskStatusPending    AITaskStatus = iota // 待处理
	AITaskStatusProcessing                     // 处理中
	AITaskStatusCompleted                      // 已完成
	AITaskStatusFailed                         // 失败
	AITaskStatusCancelled                      // 已取消
)

// DiagramType 图表类型枚举
type DiagramType string

const (
	DiagramTypeFlowchart DiagramType = "flowchart" // 流程图
	DiagramTypeSequence  DiagramType = "sequence"  // 序列图
	DiagramTypeClass     DiagramType = "class"     // 类图
	DiagramTypeGantt     DiagramType = "gantt"     // 甘特图
	DiagramTypePie       DiagramType = "pie"       // 饼图
	DiagramTypeMindmap   DiagramType = "mindmap"   // 思维导图
)

// AITask AI任务实体
// 管理AI相关的异步任务处理
type AITask struct {
	ID          int64        `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID      int64        `json:"user_id" gorm:"not null;index"`
	DocumentID  *int64       `json:"document_id" gorm:"index"` // 可选，关联的文档ID
	Type        AITaskType   `json:"type" gorm:"type:varchar(50);not null;index"`
	Status      AITaskStatus `json:"status" gorm:"type:tinyint;default:0;index"`
	InputText   string       `json:"input_text" gorm:"type:longtext;not null"`
	OutputText  string       `json:"output_text" gorm:"type:longtext"`
	Parameters  string       `json:"parameters" gorm:"type:json"` // JSON格式的参数
	ErrorMsg    string       `json:"error_message" gorm:"type:text"`
	ProcessTime int64        `json:"process_time" gorm:"default:0"` // 处理时间（毫秒）
	CreatedAt   time.Time    `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time    `json:"updated_at" gorm:"autoUpdateTime"`
	CompletedAt *time.Time   `json:"completed_at"`

	// 关联数据
	User     *User     `json:"user,omitempty" gorm:"-"`
	Document *Document `json:"document,omitempty" gorm:"-"`
}

// Validate 验证AI任务实体
func (t *AITask) Validate() error {
	if t.UserID <= 0 {
		return ErrInvalidUser
	}
	if t.InputText == "" {
		return ErrInvalidDocument
	}
	if !t.isValidType() {
		return ErrInvalidEventType
	}
	return nil
}

// isValidType 验证任务类型
func (t *AITask) isValidType() bool {
	validTypes := []AITaskType{
		AITaskTypeTextCorrection, AITaskTypeDiagramGeneration,
		AITaskTypeContentSummary, AITaskTypeTranslation,
	}
	for _, validType := range validTypes {
		if t.Type == validType {
			return true
		}
	}
	return false
}

// MarkAsProcessing 标记为处理中
func (t *AITask) MarkAsProcessing() {
	t.Status = AITaskStatusProcessing
	t.UpdatedAt = time.Now()
}

// MarkAsCompleted 标记为已完成
func (t *AITask) MarkAsCompleted(output string, processTime int64) {
	t.Status = AITaskStatusCompleted
	t.OutputText = output
	t.ProcessTime = processTime
	now := time.Now()
	t.CompletedAt = &now
	t.UpdatedAt = now
}

// MarkAsFailed 标记为失败
func (t *AITask) MarkAsFailed(errorMsg string) {
	t.Status = AITaskStatusFailed
	t.ErrorMsg = errorMsg
	t.UpdatedAt = time.Now()
}

// MarkAsCancelled 标记为已取消
func (t *AITask) MarkAsCancelled() {
	t.Status = AITaskStatusCancelled
	t.UpdatedAt = time.Now()
}

// IsCompleted 检查是否已完成
func (t *AITask) IsCompleted() bool {
	return t.Status == AITaskStatusCompleted
}

// IsFailed 检查是否失败
func (t *AITask) IsFailed() bool {
	return t.Status == AITaskStatusFailed
}

// CanCancel 检查是否可以取消
func (t *AITask) CanCancel() bool {
	return t.Status == AITaskStatusPending || t.Status == AITaskStatusProcessing
}

// TextCorrectionRequest 文本纠错请求
// 对应前端的CorrectTextParams接口
type TextCorrectionRequest struct {
	Text        string      `json:"text" validate:"required"`
	DiagramType DiagramType `json:"diagramType,omitempty"`
}

// TextCorrectionResponse 文本纠错响应
// 对应前端的CorrectTextResponse接口
type TextCorrectionResponse struct {
	OriginalText  string           `json:"originalText"`
	CorrectedText string           `json:"correctedText"`
	Corrections   []TextCorrection `json:"correction"`
	HasErrors     bool             `json:"hasErrors"`
	MermaidCode   string           `json:"mermaidCode,omitempty"`
	ErrorMessage  string           `json:"errorMessage,omitempty"`
}

// TextCorrection 文本纠错详情
type TextCorrection struct {
	Original  string `json:"original"`
	Corrected string `json:"corrected"`
}

// DiagramGenerationRequest 图表生成请求
type DiagramGenerationRequest struct {
	Text        string      `json:"text" validate:"required"`
	DiagramType DiagramType `json:"diagramType" validate:"required"`
	Style       string      `json:"style,omitempty"`
}

// DiagramGenerationResponse 图表生成响应
type DiagramGenerationResponse struct {
	MermaidCode  string `json:"mermaidCode"`
	DiagramType  string `json:"diagramType"`
	ErrorMessage string `json:"errorMessage,omitempty"`
}

// AITaskRepository AI任务仓储接口
type AITaskRepository interface {
	// 基础CRUD操作
	Store(ctx context.Context, task *AITask) error
	GetByID(ctx context.Context, id int64) (*AITask, error)
	Update(ctx context.Context, task *AITask) error
	Delete(ctx context.Context, id int64) error

	// 查询操作
	GetByUserID(ctx context.Context, userID int64, offset, limit int) ([]*AITask, error)
	GetByStatus(ctx context.Context, status AITaskStatus, limit int) ([]*AITask, error)
	GetByType(ctx context.Context, taskType AITaskType, offset, limit int) ([]*AITask, error)
	GetPendingTasks(ctx context.Context, limit int) ([]*AITask, error)

	// 统计操作
	CountByUserID(ctx context.Context, userID int64) (int64, error)
	CountByStatus(ctx context.Context, status AITaskStatus) (int64, error)

	// 清理操作
	CleanupOldTasks(ctx context.Context, olderThan time.Time) error
}

// AIUsecase AI业务逻辑接口
type AIUsecase interface {
	// 文本纠错
	CorrectText(ctx context.Context, userID int64, request *TextCorrectionRequest) (*TextCorrectionResponse, error)
	CorrectTextAsync(ctx context.Context, userID int64, request *TextCorrectionRequest) (*AITask, error)

	// 图表生成
	GenerateDiagram(ctx context.Context, userID int64, request *DiagramGenerationRequest) (*DiagramGenerationResponse, error)
	GenerateDiagramAsync(ctx context.Context, userID int64, request *DiagramGenerationRequest) (*AITask, error)

	// 任务管理
	GetTask(ctx context.Context, userID int64, taskID int64) (*AITask, error)
	GetUserTasks(ctx context.Context, userID int64, offset, limit int) ([]*AITask, error)
	CancelTask(ctx context.Context, userID int64, taskID int64) error

	// 任务处理（后台服务使用）
	ProcessPendingTasks(ctx context.Context, limit int) error
	ProcessTask(ctx context.Context, taskID int64) error
}

// AIService AI服务接口
// 定义与外部AI服务的交互接口，由基础设施层实现
type AIService interface {
	// 文本处理
	CorrectText(ctx context.Context, text string) (*TextCorrectionResponse, error)
	GenerateDiagram(ctx context.Context, text string, diagramType DiagramType) (*DiagramGenerationResponse, error)
	SummarizeContent(ctx context.Context, content string) (string, error)
	TranslateText(ctx context.Context, text, targetLang string) (string, error)

	// 健康检查
	HealthCheck(ctx context.Context) error
}

// AITaskProcessor AI任务处理器接口
// 用于后台异步处理AI任务
type AITaskProcessor interface {
	ProcessTextCorrection(ctx context.Context, task *AITask) error
	ProcessDiagramGeneration(ctx context.Context, task *AITask) error
	ProcessContentSummary(ctx context.Context, task *AITask) error
	ProcessTranslation(ctx context.Context, task *AITask) error
}
