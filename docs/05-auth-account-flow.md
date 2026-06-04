# 05. 认证、账号合并与认领流程

## 1. PC / H5 注册

```text
用户输入手机号
发送验证码
验证码通过
设置密码
如果手机号对应 PENDING_CLAIM user -> 认领预创建账号
否则创建 ACTIVE user
```

PC / H5 注册必须设置密码。

邮箱当前非必填，不做邮箱注册 / 登录。

## 2. 微信小程序登录

```text
wx.login
后端 code 换 openid / unionid
查询 user_auth_identities ACTIVE 记录
```

如果没找到：

```text
创建 PENDING_PHONE_BIND user
创建 ACTIVE user_auth_identities
```

小程序用户当前版本不需要设置密码。

## 3. 小程序绑定手机号

```text
用户绑定手机号
验证码通过
检查该手机号是否已有 user
```

情况：

```text
不存在 user -> 当前小程序临时 user 变 ACTIVE
存在 ACTIVE user -> 触发账号合并
存在 PENDING_CLAIM user -> 触发账号认领
```

## 4. 账号合并

合并依据：手机号验证码通过。

```text
source_user = 被合并账号
target_user = 保留账号
source_user.status = MERGED
source_user.merged_to_user_id = target_user.id
旧 token 立即失效
用户必须重新登录
```

如果两个账号在同一 family 绑定不同 member：

```text
用户必须选择保留一个 member 绑定。
```

## 5. 账号认领

后台预创建账号：

```text
users.status = PENDING_CLAIM
users.account_origin = ADMIN_PRE_CREATED
users.phone = 预留手机号
```

用户用该手机号注册 / 登录 / 小程序绑定时：

```text
验证码通过
不新建 user
认领该 PENDING_CLAIM user
target_user.status = ACTIVE
claimed_at = 当前时间
```

## 6. 手机号换绑

```text
旧手机号验证码通过
新手机号验证码通过
更新 users.phone
写 user_phone_history
写 operation_logs
```

## 7. 用户注销

正常注销前应先退出所有家庭并解除所有 ACTIVE 绑定。

注销时：

```text
users.status = CANCELLED
users.phone = NULL
users.phone_verified = 0
nickname = NULL
avatar_url = NULL
real_name = NULL
password_hash = NULL
user_auth_identities.status = CANCELLED / UNLINKED
family_member_user_links ACTIVE 绑定解除
```

## 8. 验证码规则

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
