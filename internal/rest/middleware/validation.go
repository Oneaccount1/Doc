package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"DOC/internal/rest/dto"
)

// ValidationMiddleware 验证中间件
// 提供通用的请求参数验证功能
type ValidationMiddleware struct {
	validator *validator.Validate
}

// NewValidationMiddleware 创建新的验证中间件
func NewValidationMiddleware() *ValidationMiddleware {
	return &ValidationMiddleware{
		validator: validator.New(),
	}
}

// ValidateRequest 验证请求中间件
// 自动验证绑定的DTO结构体
func (v *ValidationMiddleware) ValidateRequest() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// 继续处理请求
		c.Next()

		// 如果有验证错误，在这里处理
		if len(c.Errors) > 0 {
			// 提取验证错误
			var validationErrors []string
			for _, err := range c.Errors {
				var ve validator.ValidationErrors
				if errors.As(err.Err, &ve) {
					for _, fieldErr := range ve {
						validationErrors = append(validationErrors, fieldErr.Error())
					}
				}
			}
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Validation",
				"message": "请求参数验证失败" + strings.Join(validationErrors, ", "),
			})
			c.Abort()
		}
	})
}

// ContentTypeValidation 内容类型验证中间件
// 确保请求的Content-Type符合预期
func ContentTypeValidation(expectedTypes ...string) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		contentType := c.GetHeader("Content-Type")

		// 如果没有期望的类型，跳过验证
		if len(expectedTypes) == 0 {
			c.Next()
			return
		}

		// 检查内容类型是否匹配
		for _, expectedType := range expectedTypes {
			if strings.HasPrefix(contentType, expectedType) {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusUnsupportedMediaType, gin.H{
			"error":   "UNSUPPORTED_MEDIA_TYPE",
			"message": "不支持的Content-Type" + contentType,
		})
		c.Abort()
	})
}

// DocumentIDValidation 文档ID验证中间件
// 验证路径参数中的文档ID是否有效
func DocumentIDValidation() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		docID := c.Param("id")
		if docID == "" {
			docID = c.Param("documentId")
		}

		if docID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "MISSING_DOCUMENT_ID",
				"message": "缺少文档ID参数",
			})
			c.Abort()
			return
		}

		// 这里可以添加更多的ID格式验证
		// 例如检查是否为有效的数字ID
		if len(docID) > 20 {
			//c.JSON(http.StatusBadRequest, dto.ErrorResponse(
			//	"文档ID格式无效",
			//	"INVALID_DOCUMENT_ID_FORMAT",
			//))
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "INVALID_DOCUMENT_ID_FORMAT",
				"message": "文档ID格式无效",
			})
			c.Abort()
			return
		}

		c.Next()
	})
}

// RequestSizeLimit 请求大小限制中间件
// 限制请求体的大小，防止过大的请求占用过多资源
func RequestSizeLimit(maxSize int64) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		if c.Request.ContentLength > maxSize {
			//c.JSON(http.StatusRequestEntityTooLarge, dto.ErrorResponse(
			//	"请求体过大",
			//	"REQUEST_TOO_LARGE",
			//))
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error":   "REQUEST_TOO_LARGE",
				"message": "请求体过大",
			})
			c.Abort()
			return
		}

		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)
		c.Next()
	})
}

// SearchParameterValidation 搜索参数验证中间件
// 验证搜索相关的查询参数
func SearchParameterValidation() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		keyword := c.Query("keyword")

		// 验证关键词长度
		if len(keyword) == 0 {
			//c.JSON(http.StatusBadRequest, dto.ErrorResponse(
			//	"搜索关键词不能为空",
			//	"EMPTY_SEARCH_KEYWORD",
			//))
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "EMPTY_SEARCH_KEYWORD",
				"message": "搜索关键词不能为空",
			})
			c.Abort()
			return
		}

		if len(keyword) > 100 {
			//c.JSON(http.StatusBadRequest, dto.ErrorResponse(
			//	"搜索关键词太长，最多100个字符",
			//	"SEARCH_KEYWORD_TOO_LONG",
			//))
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "SEARCH_KEYWORD_TOO_LONG",
				"message": "搜索关键词太长，最多100个字符",
			})
			c.Abort()
			return
		}

		// 过滤危险字符
		if strings.ContainsAny(keyword, "<>\"'&") {
			//c.JSON(http.StatusBadRequest, dto.ErrorResponse(
			//	"搜索关键词包含非法字符",
			//	"INVALID_SEARCH_KEYWORD",
			//))
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "INVALID_SEARCH_KEYWORD",
				"message": "搜索关键词包含非法字符",
			})
			c.Abort()
			return
		}

		c.Next()
	})
}

// BatchOperationValidation 批量操作验证中间件
// 验证批量操作的参数
func BatchOperationValidation() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		var req dto.BatchOperationDto

		// 绑定请求参数
		if err := c.ShouldBindJSON(&req); err != nil {
			//c.JSON(http.StatusBadRequest, dto.ErrorResponse(
			//	"批量操作参数格式无效: "+err.Error(),
			//	"INVALID_BATCH_PARAMS",
			//))
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "INVALID_BATCH_PARAMS",
				"message": "批量操作参数格式无效" + err.Error(),
			})
			c.Abort()
			return
		}

		// 验证文档ID数量
		if len(req.DocumentIDs) == 0 {
			//c.JSON(http.StatusBadRequest, dto.ErrorResponse(
			//	"至少需要选择一个文档",
			//	"NO_DOCUMENTS_SELECTED",
			//))
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "NO_DOCUMENTS_SELECTED",
				"message": "至少需要选择一个文档",
			})
			c.Abort()
			return
		}

		if len(req.DocumentIDs) > 100 {
			//c.JSON(http.StatusBadRequest, dto.ErrorResponse(
			//	"批量操作最多支持100个文档",
			//	"TOO_MANY_DOCUMENTS",
			//))
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "TOO_MANY_DOCUMENTS",
				"message": "批量操作最多支持100个文档",
			})
			c.Abort()
			return
		}

		// 验证文档ID的有效性
		for _, docID := range req.DocumentIDs {
			if docID <= 0 {
				//c.JSON(http.StatusBadRequest, dto.ErrorResponse(
				//	"文档ID必须大于0",
				//	"INVALID_DOCUMENT_ID",
				//))
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "INVALID_DOCUMENT_ID",
					"message": "文档ID必须大于0",
				})
				c.Abort()
				return
			}
		}

		// 检查是否有重复的文档ID
		seen := make(map[int64]bool)
		for _, docID := range req.DocumentIDs {
			if seen[docID] {
				//c.JSON(http.StatusBadRequest, dto.ErrorResponse(
				//	"文档列表包含重复的ID",
				//	"DUPLICATE_DOCUMENT_IDS",
				//))
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "DUPLICATE_DOCUMENT_IDS",
					"message": "文档列表包含重复的ID",
				})
				c.Abort()
				return
			}
			seen[docID] = true
		}

		// 将验证过的参数存储到上下文中
		c.Set("batch_request", req)
		c.Next()
	})
}

// ShareParameterValidation 分享参数验证中间件
// 验证分享相关的参数
func ShareParameterValidation() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		var req dto.ShareDocumentDto

		// 绑定请求参数
		if err := c.ShouldBindJSON(&req); err != nil {
			//c.JSON(http.StatusBadRequest, dto.ErrorResponse(
			//	"分享参数格式无效: "+err.Error(),
			//	"INVALID_SHARE_PARAMS",
			//))
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "INVALID_SHARE_PARAMS",
				"message": "分享参数格式无效",
			})
			c.Abort()
			return
		}

		// 验证权限类型
		validPermissions := map[string]bool{
			"VIEW": true, "COMMENT": true, "EDIT": true, "MANAGE": true, "FULL": true,
		}
		if !validPermissions[req.Permission] {
			//c.JSON(http.StatusBadRequest, dto.ErrorResponse(
			//	"无效的权限类型",
			//	"INVALID_PERMISSION",
			//))
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "INVALID_PERMISSION",
				"message": "无效的权限类型",
			})
			c.Abort()
			return
		}

		// 验证密码长度
		if req.Password != nil && (len(*req.Password) < 4 || len(*req.Password) > 50) {
			//c.JSON(http.StatusBadRequest, dto.ErrorResponse(
			//	"分享密码长度必须在4-50个字符之间",
			//	"INVALID_PASSWORD_LENGTH",
			//))
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "INVALID_PASSWORD_LENGTH",
				"message": "分享密码长度必须在4",
			})
			c.Abort()
			return
		}

		// 验证过期时间
		if req.ExpiresAt != nil {
			if _, err := req.GetExpiresAt(); err != nil {
				//c.JSON(http.StatusBadRequest, dto.ErrorResponse(
				//	"过期时间格式无效，请使用ISO 8601格式",
				//	"INVALID_EXPIRES_AT",
				//))
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "INVALID_EXPIRES_AT",
					"message": "过期时间格式无效，请使用ISO 8601格式",
				})
				c.Abort()
				return
			}
		}

		// 验证分享用户ID列表
		if len(req.ShareWithUserIDs) > 50 {
			//c.JSON(http.StatusBadRequest, dto.ErrorResponse(
			//	"分享用户列表最多支持50个用户",
			//	"TOO_MANY_SHARE_USERS",
			//))
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "TOO_MANY_SHARE_USERS",
				"message": "分享用户列表最多支持50个用户",
			})
			c.Abort()
			return
		}

		for _, userID := range req.ShareWithUserIDs {
			if userID <= 0 {
				//c.JSON(http.StatusBadRequest, dto.ErrorResponse(
				//	"用户ID必须大于0",
				//	"INVALID_USER_ID",
				//))
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "INVALID_USER_ID",
					"message": "用户ID必须大于0",
				})
				c.Abort()
				return
			}
		}

		// 将验证过的参数存储到上下文中
		c.Set("share_request", req)
		c.Next()
	})
}
