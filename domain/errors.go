package domain

import "errors"

// 定义领域层的业务错误
// 这些错误表示业务规则违反，不依赖于任何外部框架

var (
	// 通用错误
	ErrInternalServerError = errors.New("internal server error")
	ErrBadParamInput       = errors.New("given param is not valid")
	ErrTimeout             = errors.New("timeout")
	ErrConflict            = errors.New("your item already exist")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrForbidden           = errors.New("forbidden")

	// 用户相关错误
	ErrUserNotFound        = errors.New("user not found")
	ErrUserAlreadyExist    = errors.New("user already exist")
	ErrInvalidUserEmail    = errors.New("invalid user email")
	ErrInvalidGithubId     = errors.New("invalid githubID")
	ErrInvalidUserPassword = errors.New("invalid user password")
	ErrInvalidUserName     = errors.New("invalid username")
	ErrUserIDRequired      = errors.New("user id is required")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrUserNotActive       = errors.New("user is not active")
	ErrInvalidUser         = errors.New("invalid user")
	ErrUserInfo            = errors.New("user info error")

	// 认证相关错误
	ErrInvalidToken                    = errors.New("invalid token")
	ErrTokenExpired                    = errors.New("token expired")
	ErrSessionNotFound                 = errors.New("session not found")
	ErrSessionExpired                  = errors.New("session expired")
	ErrVerificationCodeNotFound        = errors.New("verification code not found or expired")
	ErrInvalidVerificationCode         = errors.New("invalid verification code")
	ErrVerificationCodeSendTooFrequent = errors.New("verification code send too frequent")
	ErrOAuthCodeInvalid                = errors.New("oauth code invalid")
	ErrOAuthStateNotFound              = errors.New("oauth state not found")
	ErrOAuthStateMismatch              = errors.New("oauth state mismatch")
	ErrOAuthUserInfoFailed             = errors.New("failed to get oauth user info")
	ErrGitHubAccountLinked             = errors.New("github account already linked to another user")

	// 文档相关错误
	ErrDocumentNotFound     = errors.New("document not found")
	ErrDocumentAlreadyExist = errors.New("document already exist")
	ErrInvalidDocumentTitle = errors.New("invalid document title")
	ErrInvalidDocumentType  = errors.New("invalid document type")
	ErrInvalidDocumentBody  = errors.New("invalid document body")
	ErrDocumentIDRequired   = errors.New("document id is required")
	ErrDocumentLocked       = errors.New("document is locked")
	ErrAuthorIDRequired     = errors.New("author id is required")
	ErrInvalidDocument      = errors.New("invalid document")

	// 分享相关错误
	ErrShareLinkNotFound    = errors.New("share link not found")
	ErrShareLinkExpired     = errors.New("share link expired")
	ErrShareLinkDisabled    = errors.New("share link disabled")
	ErrInvalidSharePassword = errors.New("invalid share password")
	ErrShareLinkAccess      = errors.New("share link access failed")

	// 权限相关错误
	ErrPermissionDenied       = errors.New("permission denied")
	ErrInvalidPermission      = errors.New("invalid permission")
	ErrPermissionNotFound     = errors.New("permission not found")
	ErrInvalidRole            = errors.New("invalid role")
	ErrInsufficientPermission = errors.New("insufficient permission")

	// 文件上传相关错误
	ErrFileNotFound      = errors.New("file not found")
	ErrFileUploadFailed  = errors.New("file upload failed")
	ErrInvalidFileType   = errors.New("invalid file type")
	ErrFileSizeExceeded  = errors.New("file size exceeded")
	ErrInvalidFileHash   = errors.New("invalid file hash")
	ErrChunkUploadFailed = errors.New("chunk upload failed")
	ErrFileStorageError  = errors.New("file storage error")

	// AI相关错误
	ErrAIServiceUnavailable = errors.New("ai service unavailable")
	ErrAITaskNotFound       = errors.New("ai task not found")
	ErrAITaskFailed         = errors.New("ai task failed")
	ErrInvalidAIRequest     = errors.New("invalid ai request")
	ErrAIQuotaExceeded      = errors.New("ai quota exceeded")

	// 通知相关错误
	ErrNotificationNotFound         = errors.New("notification not found")
	ErrInvalidNotification          = errors.New("invalid notification")
	ErrInvalidNotificationTitle     = errors.New("invalid notification title")
	ErrInvalidNotificationContent   = errors.New("invalid notification content")
	ErrInvalidNotificationType      = errors.New("invalid notification type")
	ErrInvalidNotificationPriority  = errors.New("invalid notification priority")
	ErrRecipientIDRequired          = errors.New("recipient id is required")
	ErrNotificationAlreadyRead      = errors.New("notification already read")
	ErrNotificationPermissionDenied = errors.New("notification permission denied")

	// 协作相关错误
	ErrCollaborationSessionNotFound  = errors.New("collaboration session not found")
	ErrCollaborationPermissionDenied = errors.New("collaboration permission denied")
	ErrInvalidCollaborationOperation = errors.New("invalid collaboration operation")

	// 邮件相关错误
	ErrEmailNotFound         = errors.New("email not found")
	ErrInvalidEmailAddress   = errors.New("invalid email address")
	ErrInvalidEmailSubject   = errors.New("invalid email subject")
	ErrInvalidEmailContent   = errors.New("invalid email content")
	ErrInvalidEmailType      = errors.New("invalid email type")
	ErrEmailSendFailed       = errors.New("email send failed")
	ErrEmailTemplateNotFound = errors.New("email template not found")
	ErrEmailConfigInvalid    = errors.New("email config invalid")

	// 缓存相关错误
	ErrMarsh        = errors.New("marshal error")
	ErrUnMarsh      = errors.New("unmarshal error")
	ErrGenerateCode = errors.New("generate code error")

	// 事件相关错误
	ErrInvalidEventType = errors.New("invalid event type")
	ErrInvalidActorID   = errors.New("invalid actor id")

	// 批量操作错误
	ErrInvalidBatchRequest = errors.New("invalid batch request")
	ErrBatchSizeExceeded   = errors.New("batch size exceeded")

	// 组织相关错误（为未来扩展预留）
	ErrOrganizationNotFound     = errors.New("organization not found")
	ErrOrganizationAlreadyExist = errors.New("organization already exist")
	ErrInvalidOrganizationName  = errors.New("invalid organization name")
	ErrOrganizationIDRequired   = errors.New("organization id is required")
	ErrNotOrganizationMember    = errors.New("not organization member")
	ErrInvalidOrganization      = errors.New("invalid organization")
	ErrNotFound                 = errors.New("not found invitation")

	// 知识库相关错误（为未来扩展预留）
	ErrKnowledgeBaseNotFound     = errors.New("knowledge base not found")
	ErrKnowledgeBaseAlreadyExist = errors.New("knowledge base already exist")
	ErrInvalidKnowledgeBaseName  = errors.New("invalid knowledge base name")
	ErrKnowledgeBaseIDRequired   = errors.New("knowledge base id is required")

	// 模板相关错误（为未来扩展预留）
	ErrTemplateNotFound     = errors.New("template not found")
	ErrTemplateAlreadyExist = errors.New("template already exist")
	ErrInvalidTemplateName  = errors.New("invalid template name")

	// 空间相关错误
	ErrSpaceNotFound         = errors.New("space not found")
	ErrSpaceAlreadyExist     = errors.New("space already exist")
	ErrInvalidSpaceName      = errors.New("invalid space name")
	ErrInvalidSpaceType      = errors.New("invalid space type")
	ErrSpaceIDRequired       = errors.New("space id is required")
	ErrNotSpaceMember        = errors.New("not space member")
	ErrInvalidSpace          = errors.New("invalid space")
	ErrSpacePermissionDenied = errors.New("space permission denied")

	// 系统内错误
	ErrInternalError = errors.New("internal error")
)
