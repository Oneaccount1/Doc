package domain

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"
)

// FileType 文件类型枚举
type FileType string

const (
	FileTypeImage    FileType = "image"    // 图片
	FileTypeVideo    FileType = "video"    // 视频
	FileTypeAudio    FileType = "audio"    // 音频
	FileTypeDocument FileType = "document" // 文档
	FileTypeOther    FileType = "other"    // 其他
)

// UploadStatus 上传状态枚举
type UploadStatus int

const (
	UploadStatusPending   UploadStatus = iota // 待上传
	UploadStatusUploading                     // 上传中
	UploadStatusCompleted                     // 已完成
	UploadStatusFailed                        // 失败
	UploadStatusCancelled                     // 已取消
)

// File 文件实体
// 管理用户上传的文件信息
type File struct {
	ID                int64        `json:"id" gorm:"primaryKey;autoIncrement"`
	FileID            string       `json:"file_id" gorm:"type:varchar(64);uniqueIndex;not null"` // 唯一文件标识
	OriginalName      string       `json:"original_name" gorm:"type:varchar(500);not null"`
	ProcessedName     string       `json:"processed_name" gorm:"type:varchar(500);not null"`
	FileHash          string       `json:"file_hash" gorm:"type:varchar(64);index;not null"`
	FileSize          int64        `json:"file_size" gorm:"not null"`
	FileType          FileType     `json:"file_type" gorm:"type:varchar(20);not null;index"`
	MimeType          string       `json:"mime_type" gorm:"type:varchar(100);not null"`
	ProcessedMimeType string       `json:"processed_mime_type" gorm:"type:varchar(100)"`
	FileURL           string       `json:"file_url" gorm:"type:varchar(1000);not null"`
	ThumbnailURL      string       `json:"thumbnail_url" gorm:"type:varchar(1000)"`
	UploaderID        int64        `json:"uploader_id" gorm:"not null;index"`
	Status            UploadStatus `json:"status" gorm:"type:tinyint;default:0;index"`
	ErrorMessage      string       `json:"error_message" gorm:"type:text"`

	// 分块上传相关
	TotalChunks    int   `json:"total_chunks" gorm:"default:1"`
	UploadedChunks int   `json:"uploaded_chunks" gorm:"default:0"`
	ChunkSize      int64 `json:"chunk_size" gorm:"default:0"`

	// 外部服务相关（如ImageKit）
	ExternalFileID string `json:"external_file_id" gorm:"type:varchar(200)"`
	ExternalURL    string `json:"external_url" gorm:"type:varchar(1000)"`

	// 时间戳
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	CompletedAt *time.Time `json:"completed_at"`

	// 关联数据
	Uploader *User `json:"uploader,omitempty" gorm:"-"`
}

// GenerateFileHash 生成文件哈希
func GenerateFileHash(reader io.Reader) (string, error) {
	hash := md5.New()
	if _, err := io.Copy(hash, reader); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// DetermineFileType 根据MIME类型确定文件类型
func DetermineFileType(mimeType string) FileType {
	switch {
	case strings.HasPrefix(mimeType, "image/"):
		return FileTypeImage
	case strings.HasPrefix(mimeType, "video/"):
		return FileTypeVideo
	case strings.HasPrefix(mimeType, "audio/"):
		return FileTypeAudio
	case strings.Contains(mimeType, "pdf") ||
		strings.Contains(mimeType, "document") ||
		strings.Contains(mimeType, "text"):
		return FileTypeDocument
	default:
		return FileTypeOther
	}
}

// Validate 验证文件实体
func (f *File) Validate() error {
	if f.OriginalName == "" {
		return ErrInvalidDocument
	}
	if f.FileSize <= 0 {
		return ErrInvalidDocument
	}
	if f.UploaderID <= 0 {
		return ErrInvalidUser
	}
	if f.FileHash == "" {
		return ErrInvalidDocument
	}
	return nil
}

// IsImage 检查是否为图片文件
func (f *File) IsImage() bool {
	return f.FileType == FileTypeImage
}

// IsVideo 检查是否为视频文件
func (f *File) IsVideo() bool {
	return f.FileType == FileTypeVideo
}

// IsCompleted 检查上传是否完成
func (f *File) IsCompleted() bool {
	return f.Status == UploadStatusCompleted
}

// IsFailed 检查上传是否失败
func (f *File) IsFailed() bool {
	return f.Status == UploadStatusFailed
}

// CanCancel 检查是否可以取消
func (f *File) CanCancel() bool {
	return f.Status == UploadStatusPending || f.Status == UploadStatusUploading
}

// MarkAsUploading 标记为上传中
func (f *File) MarkAsUploading() {
	f.Status = UploadStatusUploading
	f.UpdatedAt = time.Now()
}

// MarkAsCompleted 标记为上传完成
func (f *File) MarkAsCompleted(fileURL string) {
	f.Status = UploadStatusCompleted
	f.FileURL = fileURL
	now := time.Now()
	f.CompletedAt = &now
	f.UpdatedAt = now
}

// MarkAsFailed 标记为上传失败
func (f *File) MarkAsFailed(errorMsg string) {
	f.Status = UploadStatusFailed
	f.ErrorMessage = errorMsg
	f.UpdatedAt = time.Now()
}

// MarkAsCancelled 标记为已取消
func (f *File) MarkAsCancelled() {
	f.Status = UploadStatusCancelled
	f.UpdatedAt = time.Now()
}

// UpdateProgress 更新上传进度
func (f *File) UpdateProgress(uploadedChunks int) {
	f.UploadedChunks = uploadedChunks
	f.UpdatedAt = time.Now()
}

// GetProgress 获取上传进度百分比
func (f *File) GetProgress() float64 {
	if f.TotalChunks == 0 {
		return 0
	}
	return float64(f.UploadedChunks) / float64(f.TotalChunks) * 100
}

// GetFileExtension 获取文件扩展名
func (f *File) GetFileExtension() string {
	return strings.ToLower(filepath.Ext(f.OriginalName))
}

// ChunkUpload 分块上传信息
type ChunkUpload struct {
	FileID      string `json:"file_id"`
	ChunkNumber int    `json:"chunk_number"`
	TotalChunks int    `json:"total_chunks"`
	ChunkSize   int64  `json:"chunk_size"`
	TotalSize   int64  `json:"total_size"`
	FileName    string `json:"file_name"`
	FileHash    string `json:"file_hash"`
	MimeType    string `json:"mime_type"`
	Data        []byte `json:"-"` // 分块数据
}

// Validate 验证分块上传信息
func (c *ChunkUpload) Validate() error {
	if c.FileID == "" {
		return ErrInvalidDocument
	}
	if c.ChunkNumber < 0 || c.ChunkNumber >= c.TotalChunks {
		return ErrInvalidDocument
	}
	if c.TotalChunks <= 0 {
		return ErrInvalidDocument
	}
	if c.ChunkSize <= 0 {
		return ErrInvalidDocument
	}
	if len(c.Data) == 0 {
		return ErrInvalidDocument
	}
	return nil
}

// FileRepository 文件仓储接口
type FileRepository interface {
	// 基础CRUD操作
	Store(ctx context.Context, file *File) error
	GetByID(ctx context.Context, id int64) (*File, error)
	GetByFileID(ctx context.Context, fileID string) (*File, error)
	GetByHash(ctx context.Context, hash string) (*File, error)
	Update(ctx context.Context, file *File) error
	Delete(ctx context.Context, id int64) error

	// 查询操作
	GetByUploaderID(ctx context.Context, uploaderID int64, offset, limit int) ([]*File, error)
	GetByType(ctx context.Context, fileType FileType, offset, limit int) ([]*File, error)
	GetByStatus(ctx context.Context, status UploadStatus, limit int) ([]*File, error)

	// 统计操作
	CountByUploaderID(ctx context.Context, uploaderID int64) (int64, error)
	GetTotalSizeByUploaderID(ctx context.Context, uploaderID int64) (int64, error)

	// 清理操作
	CleanupFailedUploads(ctx context.Context, olderThan time.Time) error
	CleanupOrphanedFiles(ctx context.Context) error
}

// UploadUsecase 文件上传业务逻辑接口
type UploadUsecase interface {
	// 文件上传
	UploadFile(ctx context.Context, userID int64, fileName string, fileSize int64, mimeType string, data io.Reader) (*File, error)
	UploadImage(ctx context.Context, userID int64, fileName string, data io.Reader) (*File, error)
	UploadAvatar(ctx context.Context, userID int64, fileName string, data io.Reader) (*File, error)

	// 分块上传
	InitiateChunkUpload(ctx context.Context, userID int64, fileName string, fileSize int64, mimeType string, chunkSize int64) (*File, error)
	UploadChunk(ctx context.Context, userID int64, chunk *ChunkUpload) (*File, error)
	CompleteChunkUpload(ctx context.Context, userID int64, fileID string) (*File, error)
	CancelUpload(ctx context.Context, userID int64, fileID string) error

	// 文件管理
	GetFile(ctx context.Context, userID int64, fileID string) (*File, error)
	GetUserFiles(ctx context.Context, userID int64, offset, limit int) ([]*File, error)
	DeleteFile(ctx context.Context, userID int64, fileID string) error

	// 文件检查
	CheckFileExists(ctx context.Context, hash string) (*File, error)
	GetUploadStatus(ctx context.Context, userID int64, fileID string) (*File, error)
	GetUploadedChunks(ctx context.Context, userID int64, fileID string) ([]int, error)

	// 健康检查
	HealthCheck(ctx context.Context) error
}

// FileStorage 文件存储服务接口
// 定义与外部存储服务的交互接口，由基础设施层实现
type FileStorage interface {
	// 文件操作
	Store(ctx context.Context, fileName string, data io.Reader) (string, error)
	Get(ctx context.Context, fileName string) (io.ReadCloser, error)
	Delete(ctx context.Context, fileName string) error
	Exists(ctx context.Context, fileName string) (bool, error)

	// URL生成
	GetPublicURL(ctx context.Context, fileName string) (string, error)
	GetSignedURL(ctx context.Context, fileName string, expiry time.Duration) (string, error)

	// 分块上传
	InitiateMultipartUpload(ctx context.Context, fileName string) (string, error)
	UploadPart(ctx context.Context, uploadID, fileName string, partNumber int, data io.Reader) (string, error)
	CompleteMultipartUpload(ctx context.Context, uploadID, fileName string, parts []string) (string, error)
	AbortMultipartUpload(ctx context.Context, uploadID, fileName string) error

	// 健康检查
	HealthCheck(ctx context.Context) error
}
