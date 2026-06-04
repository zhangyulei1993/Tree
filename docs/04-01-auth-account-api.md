# 04-01. 认证与账号接口完整规格 v1

## 1. 模块范围

覆盖：

```text
验证码发送
PC / H5 手机号注册
PC / H5 手机号密码登录
微信小程序快捷登录
微信小程序绑定手机号
后台预创建账号认领
账号合并
手机号换绑
用户注销
退出登录
token 失效
```

涉及表：

```text
users
user_auth_identities
verification_codes
user_phone_history
user_account_merge_logs
user_account_claim_logs
family_member_user_links
operation_logs
```

## 2. 通用规则

验证码：

```text
有效期 5 分钟
只保存 code_hash
只能使用一次
按 scene 区分
```

scene：

```text
REGISTER
LOGIN
BIND_PHONE
CHANGE_PHONE_OLD
CHANGE_PHONE_NEW
CANCEL_ACCOUNT
ACCOUNT_MERGE
ACCOUNT_CLAIM
```

用户状态：

```text
PENDING_PHONE_BIND
PENDING_CLAIM
ACTIVE
MERGED
DISABLED
CANCELLED
DELETED
```

## 3. POST /api/auth/send-code

### 请求

```json
{
  "phone": "13800000000",
  "scene": "REGISTER",
  "clientType": "H5_WEB"
}
```

### 规则

```text
REGISTER：手机号未被 ACTIVE user 占用；PENDING_CLAIM 可发送用于认领
LOGIN：手机号必须存在 ACTIVE user
BIND_PHONE：用于小程序临时账号绑定手机号
CHANGE_PHONE_OLD：发送给当前旧手机号
CHANGE_PHONE_NEW：发送给新手机号
CANCEL_ACCOUNT：发送给当前绑定手机号
```

### 返回

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "expireSeconds": 300,
    "cooldownSeconds": 60
  }
}
```

## 4. POST /api/auth/register-phone

PC / H5 手机号注册。必须设置密码。

### 请求

```json
{
  "phone": "13800000000",
  "code": "123456",
  "password": "Password123!",
  "nickname": "张三",
  "clientType": "H5_WEB"
}
```

### 流程

```text
校验 REGISTER / ACCOUNT_CLAIM 验证码
查询 users.phone
不存在 -> 创建 ACTIVE user
存在 PENDING_CLAIM -> 认领预创建账号
存在 ACTIVE -> 返回已注册
返回普通用户 token
```

### 数据库影响

```text
users
verification_codes
user_account_claim_logs
user_phone_history
operation_logs
```

## 5. POST /api/auth/login-phone

手机号 + 密码登录。

### 请求

```json
{
  "phone": "13800000000",
  "password": "Password123!",
  "clientType": "PC_WEB"
}
```

### 状态处理

```text
ACTIVE -> 登录成功
PENDING_CLAIM -> 提示先认领
MERGED -> 提示账号已合并，需要重新登录目标账号
DISABLED / CANCELLED / DELETED -> 禁止登录
```

## 6. POST /api/auth/wechat-mini/login

微信小程序快捷登录。

### 请求

```json
{
  "code": "wx_login_code",
  "clientType": "WECHAT_MINI_PROGRAM"
}
```

### 流程

```text
后端用 code 换 openid / unionid
查 ACTIVE user_auth_identities
存在 -> 找到 user 并登录
不存在 -> 创建 PENDING_PHONE_BIND user + ACTIVE identity
```

### 注销后重新进入

```text
同一个微信注销后重新进入，按新用户处理，不自动恢复旧账号。
```

## 7. POST /api/auth/wechat-mini/bind-phone

小程序绑定手机号。

### 请求

```json
{
  "phone": "13800000000",
  "code": "123456"
}
```

### 分支

```text
手机号不存在 -> 当前临时 user 变 ACTIVE
手机号对应 ACTIVE user -> 触发账号合并
手机号对应 PENDING_CLAIM user -> 触发账号认领
DISABLED / DELETED -> 拒绝绑定
```

账号合并 / 认领后：

```text
旧 token 立即失效
用户必须重新登录
```

## 8. POST /api/auth/change-phone

手机号换绑。

### 请求

```json
{
  "oldPhoneCode": "111111",
  "newPhone": "13900000000",
  "newPhoneCode": "222222"
}
```

### 规则

```text
旧手机号验证码通过
新手机号验证码通过
新手机号未被占用 -> 更新 users.phone
新手机号已被 ACTIVE user 占用 -> 进入账号合并 / 冲突处理
```

## 9. POST /api/auth/cancel-account

用户注销。

### 请求

```json
{
  "phoneCode": "123456",
  "cancelReason": "用户主动注销"
}
```

### 正常注销前校验

```text
不是任何家庭 FOUNDER
没有 ACTIVE family_member_user_links
没有 PENDING 创始人转让
没有 PENDING 家庭解散
没有未完成账号合并 / 认领
```

### 注销处理

```text
users.status = CANCELLED
users.phone = NULL
users.phone_verified = 0
nickname / avatar_url / real_name / password_hash 清空
user_auth_identities 置 CANCELLED / UNLINKED
写 user_phone_history
写 operation_logs
当前 token 失效
```

## 10. POST /api/auth/logout

当前 token 失效。

## 11. 内部流程：账号合并

```text
source_user = 被合并账号
target_user = 保留账号
source_user.status = MERGED
source_user.merged_to_user_id = target_user.id
source_user.phone = NULL
迁移 user_auth_identities
迁移或处理 family_member_user_links
写 user_account_merge_logs
写 user_phone_history
写 operation_logs
旧 token 全部失效
```

同 family 不同 member 冲突：

```text
不能自动合并绑定，必须让用户选择保留哪个 member。
```

## 12. 内部流程：账号认领

```text
target_user = 后台预创建账号
target_user.status = ACTIVE
target_user.claimed_at = 当前时间
target_user.claimed_via = PHONE_REGISTER_CLAIM / WECHAT_MINI_PROGRAM_BIND_PHONE
如果存在 temp_user，则 temp_user.status = MERGED
迁移微信 identity
写 user_account_claim_logs
写 user_phone_history
写 operation_logs
```

## 13. 错误码范围

```text
40000-40099 验证码
40100-40199 注册
40200-40299 登录
40300-40399 微信小程序登录
40400-40499 手机号绑定
40500-40599 账号合并
40600-40699 账号认领
40700-40799 手机号换绑
41000-41099 用户注销
```
