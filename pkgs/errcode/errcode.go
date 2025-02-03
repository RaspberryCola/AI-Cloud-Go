package errcode

const (
	// 系统级错误码
	ServerError        = 10000
	ParamBindError     = 10001
	AuthorizationError = 10002
	TokenError         = 10003
	TokenExpired       = 10004

	// 用户模块错误码 (20000-29999)
	UserNotFound      = 20001
	UserAlreadyExists = 20002
	PasswordError     = 20003
	UserRegisterError = 20004

	// 文件模块错误码 (30000-39999)
	FileNotFound    = 30001
	FileUploadError = 30002
	FileDeleteError = 30003
)

// 错误码消息映射
var codeMessages = map[int]string{
	ServerError:        "服务器内部错误",
	ParamBindError:     "参数绑定错误",
	AuthorizationError: "认证失败",
	TokenError:         "无效的Token",
	TokenExpired:       "Token已过期",

	UserNotFound:      "用户不存在",
	UserAlreadyExists: "用户已存在",
	PasswordError:     "密码错误",
	UserRegisterError: "用户注册错误",

	FileNotFound:    "文件不存在",
	FileUploadError: "文件上传失败",
	FileDeleteError: "文件删除失败",
}

// GetMessage 获取错误码对应的消息
func GetMessage(code int) string {
	if msg, ok := codeMessages[code]; ok {
		return msg
	}
	return "未知错误"
}
