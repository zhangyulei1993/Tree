# 22. 安全与隐私合规检查清单 v1

> 本项目涉及手机号、微信身份、家庭关系、姓名、性别、出生日期、籍贯、留言联系方式等信息，必须严格控制访问和日志。

## 1. 数据分类

普通信息：

```text
昵称
头像
公开留言昵称
```

敏感或较敏感信息：

```text
手机号
真实姓名
性别
出生日期
籍贯
家庭成员关系
家庭角色
微信 openid / unionid
留言联系方式
```

系统敏感信息：

```text
password_hash
验证码 hash
JWT secret
access token
refresh token
微信 AppSecret
数据库密码
Redis 密码
```

## 2. 账号安全

```text
密码只保存 hash
验证码只保存 hash
验证码 5 分钟有效
token 可失效
账号合并后旧 token 失效
用户注销后当前 token 失效
后台管理员登录失败记录
SUPER_ADMIN / PLATFORM_ADMIN 可被登录失败锁定
ROOT_ADMIN 不因登录失败自动锁定，但必须写失败日志
```

## 3. Token 安全

必须区分：

```text
普通用户 token
后台管理员 token
```

禁止：

```text
普通用户 token 访问 /api/admin/*
后台管理员 token 作为家庭成员身份操作家庭端接口
```

## 4. 微信安全

禁止：

```text
小程序前端保存 AppSecret
日志打印 session_key
日志打印完整 openid / unionid
前端暴露服务端微信接口密钥
```

后台展示：

```text
openidMasked
unionidMasked
```

## 5. 手机号安全

```text
后台当前版本允许 PLATFORM_ADMIN 查看完整手机号
前台公开页面不直接展示 users.phone
公开联系方式来自 families.public_contact_*
operation_logs 中优先保存 phone_masked / phone_hash
用户注销后释放 phone
```

## 6. 家庭数据权限

```text
游客只能看公开家庭资料
游客不能看私有家庭树
未绑定手机号不能加入家庭、留言、接受邀请
普通 MEMBER 只能查看家庭和编辑本人资料
FAMILY_ADMIN 可维护家庭成员
FOUNDER 可设置 FAMILY_ADMIN 和发起高风险申请
PLATFORM_ADMIN 可协助编辑家庭和成员，但不能审核创始人转让 / 家庭解散
```

## 7. 操作日志

必须写日志：

```text
后台登录成功 / 失败
管理员锁定 / 解锁
创建 / 编辑 / 删除管理员
编辑 / 禁用用户
预创建账号
账号合并
账号认领
创建 / 编辑 / 删除家庭
创建 / 编辑 / 删除成员
创建 / 删除关系
公开申请审核
留言审核
创始人转让审核
家庭解散审核
```

日志禁止记录：

```text
password_hash
验证码明文
access token
refresh token
session_key
完整 openid
完整 unionid
微信 AppSecret
数据库密码
Redis 密码
```

## 8. 删除与注销

注销前校验：

```text
不是任何家庭 FOUNDER
没有 ACTIVE family_member_user_links
没有 PENDING 创始人转让
没有 PENDING 家庭解散
```

注销后：

```text
users.status = CANCELLED
users.phone = NULL
phone_verified = 0
身份解绑
token 失效
手机号可重新注册
```

## 9. API 安全

```text
所有写接口校验 token
所有后台接口校验 admin token
所有家庭操作校验 family role
参数校验
分页 pageSize 上限
游客留言限流
验证码发送限流
CORS 白名单
文件上传类型和大小限制
```

## 10. 上线前检查

```text
生产 .env 不在 Git
JWT secret 已替换
admin 默认密码已修改
MySQL 不公网暴露
Redis 不公网暴露
HTTPS 有效
Nginx 不暴露内部目录
Swagger 在生产环境受限或关闭
日志无敏感字段
小程序 AppSecret 只在后端
```

## 11. 仓库自动安全检查

运行：

```bash
bash scripts/security-check.sh
```

脚本检查：

```text
真实 .env 是否被 Git 跟踪
私有小程序配置、证书容器和数据库备份是否入库
跟踪文件中是否包含私钥标记
源代码和部署脚本中是否存在疑似硬编码凭据
backend/.env、数据库目录、Redis 目录和构建产物是否被忽略
前端源码是否引用微信 AppSecret 配置名
```

自动扫描只是最低门禁，不能替代代码审查、依赖漏洞扫描、服务器权限检查和运行时日志抽查。

## 12. 日志抽查方法

在隔离测试环境中使用唯一的非敏感哨兵值执行登录失败、验证码、微信 mock、游客留言和审核流程，然后检查：

```text
应用日志
Nginx access/error log
operation_logs.detail_json
operation_logs.before_json
operation_logs.after_json
operation_logs.error_message
```

检查日志中不存在密码、password_hash、完整 Token、验证码、session_key、AppSecret、完整 openid/unionid、数据库密码或 Redis 密码。发现泄露后必须清理历史日志并完成密钥或 Token 轮换，不能只修复新日志。

## 13. 环境隔离要求

```text
开发、staging、production 使用不同数据库和凭据
自动化破坏性测试只连接名称以 _test 结尾的数据库
staging 部署脚本必须拒绝 production 标识
生产密钥通过服务器密钥管理或受限环境文件注入
备份文件不得进入 Git，也不得放在公开下载目录
本地 docker-compose.yml 中的开发密码禁止复用于 staging/production
```
