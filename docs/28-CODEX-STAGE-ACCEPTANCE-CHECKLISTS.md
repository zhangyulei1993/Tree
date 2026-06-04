# 28-CODEX-STAGE-ACCEPTANCE-CHECKLISTS.md

# Codex 全阶段验收清单 v1

> 每个阶段完成后，用对应清单验收。没有通过则不要进入下一阶段。

---

# M1 验收：项目骨架与本地环境

必须满足：

```text
Go 后端骨架存在
Docker Compose 存在
MySQL / Redis 可启动
.env.example 存在
/api/health 正常返回
统一响应结构存在
错误码基础结构存在
README 有启动说明
没有提前实现业务模块
```

运行：

```bash
cd backend
go test ./...
go run ./cmd/server
docker compose -f docker/docker-compose.yml config
```

---

# M2 验收：数据库 migration 与 GORM model

必须满足：

```text
migration 文件存在
核心表覆盖完整
GORM model 可编译
ROOT_ADMIN seed 存在
Service 层强约束已标 TODO
没有实现业务接口
```

重点检查：

```text
users 和 family_members 分离
family_member_user_links 存在
relationship_type 没有 SIBLING
operation_logs 存在
```

---

# M3 验收：基础设施

必须满足：

```text
JWT 工具存在
UserAuthMiddleware 存在
AdminAuthMiddleware 存在
RequirePhoneVerified 存在
OperationLogService 存在
TransactionManager 存在
Redis client 初始化存在
日志不记录敏感字段
```

---

# M4 验收：后台认证

必须满足：

```text
admin 登录成功
admin 登录失败写日志
ROOT_ADMIN 不自动锁定
SUPER_ADMIN / PLATFORM_ADMIN 可锁定
admin token 与 user token 分离
修改密码可用
```

---

# M5 验收：普通用户认证

必须满足：

```text
验证码发送可用
验证码只保存 hash
手机号注册可用
手机号密码登录可用
logout 后 token 失效
DISABLED / CANCELLED / MERGED 用户不能登录
```

---

# M6 验收：微信登录、绑定、认领、合并

必须满足：

```text
微信登录生成 PENDING_PHONE_BIND 用户
绑定新手机号成功
绑定已有手机号触发合并
绑定 PENDING_CLAIM 手机号触发认领
合并后旧 token 失效
注销释放手机号
敏感微信字段不完整入日志
```

---

# M7 验收：家庭基础

必须满足：

```text
创建家庭成功
创建家庭时创建 founder member
创建家庭时创建 FOUNDER link
current_founder_member_id 正确
家庭搜索可用
公开主页只允许 APPROVED 家庭访问
```

---

# M8 验收：家庭成员

必须满足：

```text
创建成员成功并 graph_version +1
编辑关键字段 graph_version +1
删除成员软删除
存在下级成员不能删除
FOUNDER 不能直接删除
NOT_REQUIRED 成员不能绑定用户
操作写日志
```

---

# M9 验收：家庭关系

必须满足：

```text
ADD_FATHER 可用
ADD_MOTHER 可用
ADD_CHILD 可用
ADD_SPOUSE 可用
ADD_SIBLING 可用
SIBLING 不落库
baseMember 无父母时 ADD_SIBLING 失败
PRIMARY 父亲唯一
PRIMARY 母亲唯一
多个配偶允许
关系变化更新 graph_version
```

---

# M10 验收：家庭树

必须满足：

```text
私有家庭树返回 nodes / edges / tree
公开家庭树只允许 APPROVED 家庭
treeMode = LIST_TREE
不显示动态称谓
Redis 缓存 key 使用 graphVersion
```

---

# M11 验收：邀请与加入

必须满足：

```text
PLATFORM_ADMIN 只能站内邀请已注册用户
PLATFORM_ADMIN 不能生成分享链接
家庭管理员可生成分享链接
邀请 7 天有效
接受邀请前必须绑定手机号
接受邀请创建 ACTIVE link
接受邀请不更新 graph_version
加入申请可审核
```

---

# M12 验收：公开申请与留言

必须满足：

```text
公开申请可提交
同一 family 只能有一个 PENDING 公开申请
PLATFORM_ADMIN 不能审核自己发起的公开申请
审核通过后公开主页可访问
游客留言可提交
留言审核通过后展示
公开留言不展示联系方式
留言限流存在
```

---

# M13 验收：角色、转让、解散

必须满足：

```text
FOUNDER 可设置 FAMILY_ADMIN
NOT_REQUIRED 成员不能成为管理员
创始人转让必须 ROOT_ADMIN / SUPER_ADMIN 审核
PLATFORM_ADMIN 不能审核创始人转让
家庭解散必须 ROOT_ADMIN / SUPER_ADMIN 审核
PLATFORM_ADMIN 不能审核家庭解散
解散后数据保留
恢复后默认 PRIVATE
```

---

# M14 验收：管理后台 UI

必须满足：

```text
admin-web 可 build
页面范围完整
使用 Element Plus
使用设计令牌
PLATFORM_ADMIN 不显示高风险审核入口
危险操作有确认弹窗
静态原型不接真实 API，除非明确要求
```

---

# M15 验收：前台与小程序 UI

必须满足：

```text
web 可 build
miniapp 可 build:mp-weixin
家庭搜索页存在
公开主页存在
公开树存在
小程序邀请确认页存在
绑定手机号页存在
用户协议 / 隐私政策 / 注销入口存在
```

---

# M16 验收：测试与上线准备

必须满足：

```text
P0 测试通过
权限越权测试通过
operation_logs 检查通过
graph_version 检查通过
敏感日志检查通过
部署脚本存在
生产检查清单完成
小程序体验版测试清单完成
```

---

# 最终上线前必须满足

```text
HTTPS 完成
合法域名配置完成
隐私政策完成
数据库备份完成
默认 admin 密码已修改
权限测试通过
账号合并测试通过
家庭成员关系测试通过
operation_logs 正常写入
graph_version 正常更新
```
