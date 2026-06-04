# 01-system-architecture-and-tech-stack.md

# 系统架构与技术栈说明 v2 - Go 后端版

> 本文档是项目主架构文档。  
> 它把业务文档中的功能模块落到具体开发端、技术语言、部署组件和模块职责上。  
> 当前版本后端正式采用 **Go + Gin + GORM**。

---

# 1. 系统总体架构

项目采用前后端分离架构：

```text
管理后台 admin-web  ─┐
PC/H5 前台 web       ├── 后端 API backend ── MySQL
微信小程序 miniapp   ┘                    └── Redis
```

后台、前台、小程序都通过 RESTful API 调用后端。

---

# 2. 各端职责

## 2.1 后端 API

负责：

```text
认证与账号
微信小程序登录
账号合并与认领
家庭管理
成员管理
关系推导
邀请与加入申请
公开申请
留言审核
后台权限
操作日志
graph_version
缓存控制
```

技术：

```text
Go + Gin + GORM + MySQL 8 + Redis 7
```

---

## 2.2 管理后台

负责：

```text
后台登录
用户管理
家庭管理
成员管理
公开申请审核
游客留言审核
创始人转让审核
家庭解散审核
管理员管理
操作日志查看
```

技术：

```text
Vue 3 + TypeScript + Vite + Element Plus
```

---

## 2.3 PC / H5 前台

负责：

```text
家庭搜索
公开家庭主页
公开家庭树展示
用户注册登录
我的家庭
个人中心
邀请确认
加入申请
```

技术：

```text
Vue 3 + TypeScript + Vite
```

---

## 2.4 微信小程序

负责：

```text
微信快捷登录
绑定手机号
家庭搜索
公开主页
公开树
我的家庭
邀请确认
加入申请
个人中心
```

技术：

```text
uni-app + Vue 3 + TypeScript
```

---

# 3. Go 后端模块划分

```text
auth                  认证、验证码、token
user                  普通用户
account               账号合并、账号认领
family/core           家庭基础
family/member         家庭成员
family/relationship   家庭关系
family/tree           家庭树组装
family/invitation     邀请
family/joinrequest    加入申请
family/publicdisplay  公开申请
family/message        游客留言
family/role           家庭角色
family/transfer       创始人转让
family/dissolution    家庭解散
admin                 后台管理员与后台管理
operationlog          操作日志
common/security       权限认证
common/errors         错误码
common/cache          Redis 缓存
common/transaction    事务封装
```

---

# 4. Go 后端分层

每个模块采用：

```text
handler
service
repository
model
dto
vo
enum
```

职责：

```text
handler：HTTP 入参、出参
service：业务规则、事务、权限、日志、graph_version
repository：数据库查询
model：GORM model
dto：请求对象
vo：返回对象
enum：状态和常量
```

---

# 5. 数据库组件

数据库：

```text
MySQL 8
```

数据库访问：

```text
GORM
```

数据库迁移：

```text
golang-migrate
```

核心表：

```text
users
user_auth_identities
user_phone_history
verification_codes
families
family_members
family_relationships
family_member_user_links
family_invitations
family_join_requests
family_public_applications
visitor_messages
admin_users
operation_logs
family_founder_transfer_requests
family_dissolution_requests
```

---

# 6. Redis 组件

Redis 用途：

```text
验证码缓存
token 黑名单
家庭树缓存
游客留言限流
接口限流
```

Go 客户端：

```text
go-redis
```

---

# 7. 权限系统

权限分两套：

```text
后台管理员权限
家庭业务权限
```

后台管理员：

```text
ROOT_ADMIN
SUPER_ADMIN
PLATFORM_ADMIN
```

家庭业务角色：

```text
FOUNDER
FAMILY_ADMIN
MEMBER
VISITOR
```

关键规则：

```text
后台 token 与普通用户 token 必须分离。
PLATFORM_ADMIN 不能审核创始人转让。
PLATFORM_ADMIN 不能审核家庭解散。
FOUNDER 才能设置 FAMILY_ADMIN。
```

建议实现：

```text
AdminAuthMiddleware
UserAuthMiddleware
RequireAdminRole
RequirePhoneVerified
FamilyPermissionService
```

---

# 8. 家庭树架构

一期只做：

```text
列表树 / 层级树
```

不做：

```text
复杂关系图谱
动态称谓显示
中心人物关系网
```

家庭树接口返回：

```text
nodes
edges
tree
```

缓存 key：

```text
family:{familyId}:tree:v{graphVersion}
```

---

# 9. graph_version 设计

触发更新：

```text
新增成员
删除成员
恢复成员
新增关系
删除关系
修改成员关键字段
修改关系性质
```

不触发：

```text
邀请
接受邀请
用户绑定
角色变更
公开申请
留言
创始人转让
家庭解散
```

实现建议：

```text
GraphVersionService.Increment(familyID)
GraphVersionService.Get(familyID)
GraphVersionService.BuildTreeCacheKey(familyID, graphVersion)
```

---

# 10. 操作日志架构

所有敏感操作写入：

```text
operation_logs
```

操作者类型：

```text
ADMIN
USER
SYSTEM
```

必须记录：

```text
operator
module
action
target
before_json
after_json
result
ip
user_agent
created_at
```

实现建议：

```text
OperationLogService.WriteSuccess(...)
OperationLogService.WriteFailed(...)
```

---

# 11. 部署架构建议

一期部署：

```text
Nginx
Go 后端二进制服务
MySQL 8
Redis 7
管理后台静态文件
PC/H5 静态文件
微信小程序发布到微信平台
```

可使用：

```text
Docker Compose
```

本地开发建议：

```text
docker-compose 启动 MySQL + Redis
backend 使用 go run ./cmd/server 启动
admin-web / web 使用 npm run dev
miniapp 使用 uni-app 工具链启动
```

---

# 12. 开发顺序

推荐顺序：

```text
1. Go 后端项目骨架
2. 数据库 schema 和 migration
3. 配置、日志、统一响应、错误码
4. Redis、JWT、权限中间件
5. 认证与账号
6. 后台 admin 登录
7. operation_logs
8. 家庭基础
9. family_members
10. family_relationships
11. family tree
12. invitation / join request
13. public application / visitor message
14. founder transfer / dissolution
15. 管理后台页面
16. 前台网站页面
17. 小程序页面
18. 测试与验收
```
