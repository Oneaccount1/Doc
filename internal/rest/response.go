package rest

import (
	"DOC/internal/rest/dto"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ResponseData 基本响应结构
type ResponseData struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data"`
	Timestamp int64       `json:"timestamp"`
}

// SearchUsersResponse 搜索用户响应结构（匹配前端期望）
type SearchUsersResponse struct {
	Users []*dto.UserResponse `json:"users"` // 用户列表
	Total int64               `json:"total"` // 总数
}

func SearchResponse(c *gin.Context, statusCode int, users []*dto.UserResponse, total int64) {
	c.JSON(statusCode, SearchUsersResponse{
		Users: users,
		Total: total,
	})
}

// Response 返回响应
func Response(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, ResponseData{
		Code:      statusCode,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().Unix(),
	})
}

// ResponseBadRequest 返回400错误
func ResponseBadRequest(c *gin.Context, message string) {
	Response(c, http.StatusBadRequest, message, nil)
}

// ResponseUnauthorized 返回401错误
func ResponseUnauthorized(c *gin.Context, message string) {
	Response(c, http.StatusUnauthorized, message, nil)
}

// ResponseForbidden 返回403错误
func ResponseForbidden(c *gin.Context, message string) {
	Response(c, http.StatusForbidden, message, nil)
}

// ResponseNotFound 返回404错误
func ResponseNotFound(c *gin.Context, message string) {
	Response(c, http.StatusNotFound, message, nil)
}

// ResponseConflict 返回409错误
func ResponseConflict(c *gin.Context, message string) {
	Response(c, http.StatusConflict, message, nil)
}

// ResponseTooManyRequests 返回429错误
func ResponseTooManyRequests(c *gin.Context, message string) {
	Response(c, http.StatusTooManyRequests, message, nil)
}

// ResponseInternalServerError 返回500错误
func ResponseInternalServerError(c *gin.Context, message string) {
	Response(c, http.StatusInternalServerError, message, nil)
}

// ResponseCreated 返回201成功响应
func ResponseCreated(c *gin.Context, message string, data interface{}) {
	Response(c, http.StatusCreated, message, data)
}

// ResponseOK 返回200成功响应
func ResponseOK(c *gin.Context, message string, data interface{}) {

	Response(c, http.StatusOK, message, data)
}

//// DocumentResponseOK 返回我的文档200成功响应
//func DocumentResponseOK(c *gin.Context, users []*dto.UserResponse, total int64) {
//	SearchResponse(c, http.StatusOK, users, total)
//}
