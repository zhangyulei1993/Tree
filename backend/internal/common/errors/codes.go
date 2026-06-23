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

const (
	CodeVerificationPhoneInvalid Code = 40001
	CodeVerificationSceneInvalid Code = 40002
	CodeVerificationTooFrequent  Code = 40003
	CodeVerificationPhoneExists  Code = 40004
	CodeVerificationPhoneMissing Code = 40005
	CodeVerificationStatusDenied Code = 40006
	CodeVerificationCodeInvalid  Code = 40007
	CodeVerificationCodeExpired  Code = 40008
	CodeVerificationCodeUsed     Code = 40009
)

const (
	CodeRegisterCodeInvalid       Code = 40101
	CodeRegisterCodeExpired       Code = 40102
	CodeRegisterPhoneExists       Code = 40103
	CodeRegisterPasswordWeak      Code = 40104
	CodeRegisterStatusDenied      Code = 40105
	CodeRegisterClaimFailed       Code = 40106
	CodeLoginPhonePasswordInvalid Code = 40201
	CodeLoginDisabled             Code = 40202
	CodeLoginCancelled            Code = 40203
	CodeLoginMerged               Code = 40204
	CodeLoginPendingClaim         Code = 40205
	CodeLoginTooManyFailures      Code = 40206
)

const (
	CodeWechatCodeInvalid          Code = 40301
	CodeWechatOpenIDFetchFailed    Code = 40302
	CodeWechatIdentityExpired      Code = 40303
	CodeWechatAccountInvalid       Code = 40304
	CodeWechatConfigError          Code = 40305
	CodeWechatPhoneCodeInvalid     Code = 40306
	CodeWechatPhoneFetchFailed     Code = 40307
	CodeWechatPhoneLoginInvalid    Code = 40308
	CodeBindPhoneLoginRequired     Code = 40401
	CodeBindPhoneCodeInvalid       Code = 40402
	CodeBindPhoneStatusDenied      Code = 40403
	CodeBindPhoneExists            Code = 40404
	CodeAccountMergeFailed         Code = 40405
	CodeAccountClaimFailed         Code = 40406
	CodeAccountMergeMemberConflict Code = 40504
	CodeProfileNicknameInvalid     Code = 40601
	CodeChangeOldPhoneCodeInvalid  Code = 40701
	CodeChangeNewPhoneCodeInvalid  Code = 40702
	CodeChangeNewPhoneExists       Code = 40703
	CodeCancelStatusDenied         Code = 41001
	CodeCancelActiveFamily         Code = 41002
	CodeCancelFounder              Code = 41003
	CodeCancelPendingTransfer      Code = 41004
	CodeCancelPendingDissolution   Code = 41005
	CodeCancelCodeInvalid          Code = 41006
)

const (
	CodeAdminUsernameOrPasswordInvalid Code = 47001
	CodeAdminDisabled                  Code = 47002
	CodeAdminLocked                    Code = 47003
	CodeAdminDeleted                   Code = 47004
	CodeOperationLogForbidden          Code = 48001
	CodeOperationLogNotFound           Code = 48002
	CodeOperationLogQueryInvalid       Code = 48003
	CodeOperationLogTimeRangeInvalid   Code = 48004
	CodeOperationLogDetailForbidden    Code = 48005
)

const (
	CodeContentCategoryNotFound Code = 49001
	CodeContentArticleNotFound  Code = 49002
	CodeContentInvalidStatus    Code = 49003
	CodeContentDuplicateKey     Code = 49004
	CodeContentForbidden        Code = 49005
	CodeContentInvalidInput     Code = 49006
)

var messages = map[Code]string{
	CodeSuccess:                        "success",
	CodeSystemError:                    "系统错误",
	CodeInvalidParams:                  "参数错误",
	CodeMethodNotAllowed:               "请求方法不支持",
	CodeNotFound:                       "数据不存在",
	CodeInvalidDataStatus:              "数据状态不允许操作",
	CodeTooManyRequests:                "操作过于频繁",
	CodeForbidden:                      "无权访问",
	CodeUnauthorized:                   "未登录",
	CodeLoginExpired:                   "登录已过期",
	CodePhoneBindRequired:              "请先绑定手机号",
	CodeAccountStatusInvalid:           "当前账号状态异常",
	CodeDuplicateOperation:             "重复操作",
	CodeResourceNotFound:               "请求资源不存在",
	CodeUploadFailed:                   "文件上传失败",
	CodeDataConflict:                   "数据冲突",
	CodeVerificationPhoneInvalid:       "手机号格式错误",
	CodeVerificationSceneInvalid:       "验证码场景错误",
	CodeVerificationTooFrequent:        "发送过于频繁",
	CodeVerificationPhoneExists:        "手机号已注册",
	CodeVerificationPhoneMissing:       "手机号未注册",
	CodeVerificationStatusDenied:       "当前账号状态不允许发送验证码",
	CodeVerificationCodeInvalid:        "验证码错误",
	CodeVerificationCodeExpired:        "验证码已过期",
	CodeVerificationCodeUsed:           "验证码已使用",
	CodeRegisterCodeInvalid:            "验证码错误",
	CodeRegisterCodeExpired:            "验证码已过期",
	CodeRegisterPhoneExists:            "手机号已注册",
	CodeRegisterPasswordWeak:           "密码不符合规则",
	CodeRegisterStatusDenied:           "账号状态不允许注册",
	CodeRegisterClaimFailed:            "预创建账号认领失败",
	CodeLoginPhonePasswordInvalid:      "手机号或密码错误",
	CodeLoginDisabled:                  "账号已禁用",
	CodeLoginCancelled:                 "账号已注销",
	CodeLoginMerged:                    "账号已合并",
	CodeLoginPendingClaim:              "账号待认领",
	CodeLoginTooManyFailures:           "登录失败次数过多",
	CodeWechatCodeInvalid:              "微信登录 code 无效",
	CodeWechatOpenIDFetchFailed:        "微信 openid 获取失败",
	CodeWechatIdentityExpired:          "微信身份已失效",
	CodeWechatAccountInvalid:           "当前微信账号状态异常",
	CodeWechatConfigError:              "小程序配置错误",
	CodeWechatPhoneCodeInvalid:         "微信手机号 code 无效",
	CodeWechatPhoneFetchFailed:         "微信手机号获取失败",
	CodeWechatPhoneLoginInvalid:        "微信手机号登录账号状态异常",
	CodeBindPhoneLoginRequired:         "请先登录",
	CodeBindPhoneCodeInvalid:           "验证码错误或已过期",
	CodeBindPhoneStatusDenied:          "当前账号不能绑定手机号",
	CodeBindPhoneExists:                "手机号已绑定其他账号",
	CodeAccountMergeFailed:             "账号合并失败",
	CodeAccountClaimFailed:             "账号认领失败",
	CodeAccountMergeMemberConflict:     "同一家庭下存在多个成员绑定冲突",
	CodeProfileNicknameInvalid:         "昵称格式不正确",
	CodeChangeOldPhoneCodeInvalid:      "旧手机号验证码错误",
	CodeChangeNewPhoneCodeInvalid:      "新手机号验证码错误",
	CodeChangeNewPhoneExists:           "新手机号已被其他账号绑定",
	CodeCancelStatusDenied:             "当前账号状态不允许注销",
	CodeCancelActiveFamily:             "请先退出所有家庭",
	CodeCancelFounder:                  "家庭创始人必须先转让创始人身份",
	CodeCancelPendingTransfer:          "存在未处理的创始人转让申请",
	CodeCancelPendingDissolution:       "存在未处理的家庭解散申请",
	CodeCancelCodeInvalid:              "验证码错误或已过期",
	CodeAdminUsernameOrPasswordInvalid: "用户名或密码错误",
	CodeAdminDisabled:                  "管理员账号已禁用",
	CodeAdminLocked:                    "管理员账号已锁定",
	CodeAdminDeleted:                   "管理员账号已删除",
	CodeOperationLogForbidden:          "无权查看操作日志",
	CodeOperationLogNotFound:           "操作日志不存在",
	CodeOperationLogQueryInvalid:       "日志查询参数错误",
	CodeOperationLogTimeRangeInvalid:   "时间范围错误",
	CodeOperationLogDetailForbidden:    "日志详情不可访问",
	CodeContentCategoryNotFound:        "内容分类不存在",
	CodeContentArticleNotFound:         "内容不存在",
	CodeContentInvalidStatus:           "内容状态不可操作",
	CodeContentDuplicateKey:            "内容标识已存在",
	CodeContentForbidden:               "无权操作内容",
	CodeContentInvalidInput:            "内容参数错误",
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
