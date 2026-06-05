package errors

type Code int

const (
	CodeSuccess              Code = 0
	CodeSystemError          Code = 10000
	CodeInvalidParams        Code = 10001
	CodeMethodNotAllowed     Code = 10002
	CodeNotFound             Code = 10003
	CodeInvalidDataStatus    Code = 10004
	CodeTooManyRequests      Code = 10005
	CodeForbidden            Code = 10006
	CodeUnauthorized         Code = 10007
	CodeLoginExpired         Code = 10008
	CodePhoneBindRequired    Code = 10009
	CodeAccountStatusInvalid Code = 10010
	CodeDuplicateOperation   Code = 10011
	CodeResourceNotFound     Code = 10012
	CodeUploadFailed         Code = 10013
	CodeDataConflict         Code = 10014
)

var messages = map[Code]string{
	CodeSuccess:              "success",
	CodeSystemError:          "系统错误",
	CodeInvalidParams:        "参数错误",
	CodeMethodNotAllowed:     "请求方法不支持",
	CodeNotFound:             "数据不存在",
	CodeInvalidDataStatus:    "数据状态不允许操作",
	CodeTooManyRequests:      "操作过于频繁",
	CodeForbidden:            "无权访问",
	CodeUnauthorized:         "未登录",
	CodeLoginExpired:         "登录已过期",
	CodePhoneBindRequired:    "请先绑定手机号",
	CodeAccountStatusInvalid: "当前账号状态异常",
	CodeDuplicateOperation:   "重复操作",
	CodeResourceNotFound:     "请求资源不存在",
	CodeUploadFailed:         "文件上传失败",
	CodeDataConflict:         "数据冲突",
}

func Message(code Code) string {
	if message, ok := messages[code]; ok {
		return message
	}
	return messages[CodeSystemError]
}

type BusinessError struct {
	Code    Code
	Message string
}

func New(code Code) *BusinessError {
	return &BusinessError{
		Code:    code,
		Message: Message(code),
	}
}

func (e *BusinessError) Error() string {
	return e.Message
}
