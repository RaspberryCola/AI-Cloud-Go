package response

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`    // 业务码
	Message string      `json:"message"` // 提示信息
	Data    interface{} `json:"data"`    // 数据
}

// 分页数据结构
type PageData struct {
	Total int64       `json:"total"` // 总数
	List  interface{} `json:"list"`  // 数据列表
}

const (
	SUCCESS      = 0
	ERROR        = 1
	ERROR_PARAM  = 2
	ERROR_AUTH   = 401
	ERROR_SERVER = 500
)

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    SUCCESS,
		Message: "success",
		Data:    data,
	})
}

// SuccessWithMessage 成功响应带自定义消息
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    SUCCESS,
		Message: message,
		Data:    data,
	})
}

// PageSuccess 分页数据响应
func PageSuccess(c *gin.Context, list interface{}, total int64) {
	c.JSON(http.StatusOK, Response{
		Code:    SUCCESS,
		Message: "success",
		Data: PageData{
			List:  list,
			Total: total,
		},
	})
}

// Error 错误响应
func Error(c *gin.Context, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    ERROR,
		Message: message,
		Data:    nil,
	})
}

//// ErrorWithCode 错误响应带自定义错误码
//func ErrorWithCode(c *gin.Context, code int, message string) {
//	c.JSON(http.StatusOK, Response{
//		Code:    code,
//		Message: message,
//		Data:    nil,
//	})
//}

// 错误响应
func ErrorCustom(c *gin.Context, httpCode, code int, msg string, data interface{}) {
	c.JSON(httpCode, Response{
		Code:    code,
		Message: msg,
		Data:    data,
	})
}
func ParamError(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusBadRequest, Response{
		Code:    code,
		Message: msg,
	})
}

func UnauthorizedError(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusUnauthorized, Response{
		Code:    code,
		Message: msg,
	})
}

func InternalError(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusInternalServerError, Response{
		Code:    code,
		Message: msg,
	})
}
