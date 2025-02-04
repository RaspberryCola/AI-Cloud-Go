package errcode

const (
	/******************** 基础错误码 (10000-19999) ********************/
	// 通用系统错误
	SuccessCode         = 0     // 特殊保留成功码
	InternalServerError = 10001 // 服务器内部错误
	DatabaseError       = 10002 // 数据库操作错误
	CacheError          = 10003 // 缓存服务错误
	RateLimitExceeded   = 10004 // 请求频率限制

	// 请求参数相关
	ParamBindError     = 10101 // 参数绑定错误
	ParamValidateError = 10102 // 参数验证失败

	// 认证授权相关
	UnauthorizedError = 10201 // 身份未认证
	ForbiddenError    = 10202 // 无访问权限
	TokenExpired      = 10203 // Token过期
	TokenInvalid      = 10204 // Token无效
	TokenMissing      = 10205 // Token缺失

	/******************** 业务错误码 ********************/
	// 用户模块 (20000-20999)
	UserNotFound      = 20001 // 用户不存在
	UserAlreadyExists = 20002 // 用户已存在
	PasswordMismatch  = 20003 // 密码错误
	UserDisabled      = 20004 // 用户被禁用
	EmailInvalid      = 20005 // 邮箱格式错误

	// 文件模块 (21000-21999)
	FileNotFound     = 21001 // 文件不存在
	FileUploadFailed = 21002 // 文件上传失败
	FileDeleteFailed = 21003 // 文件删除失败
	FileSizeExceeded = 21004 // 文件大小超限
	FileTypeInvalid  = 21005 // 文件类型无效
	FileParseFailed  = 21006 // 文件解析失败
	FileListFailed   = 21007 // 文杰列表获取失败
	FileSearchFailed = 21008 // 文件搜索失败
	// 订单模块 (22000-22999)
	// 可后续扩展...
)
