# 13. UI 页面规格说明 v1

> 本文档不是高保真设计稿，而是开发级页面规格。前端、设计师和 Codex 应以此确定页面范围、字段、按钮、权限和对应 API。

## 1. UI 总原则

一期优先：流程完整、字段清晰、权限明确、异常提示清楚。暂不做复杂家谱图谱、拖拽式树编辑、动态称谓展示、复杂通知中心。

---

# 2. 管理后台 admin-web

技术栈：Vue 3 + TypeScript + Vite + Element Plus。

## 2.1 后台登录页

路径：`/admin/login`

字段：`username`、`password`

API：

```http
POST /api/admin/auth/login
```

异常：用户名或密码错误、账号禁用、账号锁定。

## 2.2 仪表盘

路径：`/admin/dashboard`

展示：用户总数、家庭总数、公开家庭数、待审核公开申请、待审核留言、待审核创始人转让、待审核家庭解散、最近操作日志。

## 2.3 用户管理

路径：`/admin/users`

权限：ROOT_ADMIN / SUPER_ADMIN / PLATFORM_ADMIN。

表格字段：用户 ID、手机号、昵称、真实姓名、账号来源、注册端、手机号验证状态、账号状态、创建时间、最后登录时间。

筛选：手机号、昵称/姓名、状态、账号来源、注册端、手机号验证状态、创建时间。

按钮：查看详情、编辑资料、禁用用户、恢复用户、预创建账号。

API：

```http
GET  /api/admin/users
GET  /api/admin/users/{userId}
PUT  /api/admin/users/{userId}
POST /api/admin/users/{userId}/disable
POST /api/admin/users/{userId}/restore
POST /api/admin/users/pre-create
```

## 2.4 用户详情

路径：`/admin/users/:userId`

展示：基础资料、手机号状态、微信身份摘要、账号来源、账号状态、加入家庭列表、绑定成员列表、账号合并/认领摘要、操作日志摘要。

不可编辑：phone、openid、unionid、account_origin、merged_to_user_id、claimed_at。

## 2.5 家庭管理

路径：`/admin/families`

权限：ROOT_ADMIN / SUPER_ADMIN / PLATFORM_ADMIN。

表格字段：家庭 ID、家族名称、家族姓氏、籍贯、地区、状态、是否可搜索、公开状态、当前创始人、成员数量、graph_version、创建时间。

筛选：家族名称/姓氏/地区关键词、状态、公开状态、地区、创建时间。

按钮：查看详情、编辑家庭、成员管理、禁用家庭、删除家庭、恢复家庭、下架公开家庭。

权限边界：PLATFORM_ADMIN 可编辑、禁用、下架公开家庭；不可删除家庭、不可恢复家庭。

## 2.6 家庭详情

路径：`/admin/families/:familyId`

Tab：基础信息、成员列表、关系维护、公开申请、加入申请、邀请记录、留言记录、操作日志。

## 2.7 家庭成员管理

路径：`/admin/families/:familyId/members`

展示：列表树/层级树、成员列表、成员详情侧边栏。

成员字段：成员 ID、展示姓名、姓、字辈、名、性别、出生日期、成员类型、绑定状态、成员状态。

按钮：新增成员、编辑成员、删除成员、标记无需绑定用户、站内邀请、添加父亲、添加母亲、添加子女、添加兄弟姐妹、添加配偶、删除关系。

关键规则：存在下级成员时禁止删除；添加兄弟姐妹但无父母节点时提示先创建父母节点；relationship_type 不显示 SIBLING。

## 2.8 公开申请审核

路径：`/admin/public-applications`

权限：ROOT_ADMIN / SUPER_ADMIN / PLATFORM_ADMIN。

表格字段：申请 ID、家庭名称、家族姓氏、地区、申请人类型、申请人、申请理由、状态、创建时间。

按钮：查看、审核通过、审核拒绝。

规则：PLATFORM_ADMIN 不能审核自己发起的公开申请。

## 2.9 游客留言审核

路径：`/admin/visitor-messages`

筛选：状态、家庭名称、留言关键词、留言人、审核管理员、时间范围。

按钮：查看详情、审核通过、审核拒绝、删除。

公开留言列表不展示 visitorPhone / visitorWechat。

## 2.10 创始人转让审核

路径：`/admin/founder-transfer-requests`

权限：ROOT_ADMIN / SUPER_ADMIN。PLATFORM_ADMIN 不可访问。

按钮：审核通过、审核拒绝、查看详情。

## 2.11 家庭解散审核

路径：`/admin/dissolution-requests`

权限：ROOT_ADMIN / SUPER_ADMIN。PLATFORM_ADMIN 不可访问。

## 2.12 后台管理员管理

路径：`/admin/admin-users`

权限：ROOT_ADMIN / SUPER_ADMIN。PLATFORM_ADMIN 不可访问。

规则：ROOT_ADMIN 有且只有一个；任何管理员不能管理同级管理员；只有 ROOT_ADMIN 可解锁管理员。

## 2.13 操作日志

路径：`/admin/operation-logs`

筛选：操作者类型、角色、管理员、普通用户、模块、动作、目标类型、目标 ID、家庭 ID、成员 ID、关联用户 ID、结果、时间范围、关键词。

详情展示：before_json、after_json、detail_json、error_message、user_agent。

---

# 3. PC/H5 前台 web

## 3.1 首页

路径：`/`

内容：平台介绍、家庭搜索入口、公开家庭推荐、登录/注册入口。

## 3.2 家庭搜索

路径：`/families`

搜索：家族名称关键词、家族姓氏、籍贯、地区。

结果：家族名称、姓氏、地址籍贯、创始人姓名、公开联系方式、公开状态。

API：`GET /api/families`

## 3.3 公开家庭主页

路径：`/families/:familyId/public`

展示：家庭名称、姓氏、籍贯、地区、简介、公开联系方式、公开树入口、公开留言、留言按钮。

API：

```http
GET /api/public/families/{familyId}/profile
GET /api/public/families/{familyId}/visitor-messages
POST /api/public/families/{familyId}/visitor-messages
```

## 3.4 公开家庭树

路径：`/families/:familyId/tree/public`

展示：顶点、第二排、第三排、第四排等列表树。不展示动态称谓和复杂图谱。

## 3.5 登录 / 注册

路径：`/login`、`/register`

登录：手机号、密码。注册：手机号、验证码、密码、昵称。

## 3.6 个人中心

路径：`/me`

展示：基础资料、手机号绑定状态、我的家庭、我的邀请、我的加入申请、注销账号入口。

## 3.7 邀请详情

路径：`/invite/:inviteToken`

流程：未登录引导登录；未绑定手机号引导绑定；已绑定手机号则允许接受/拒绝。

---

# 4. 微信小程序 miniapp

技术栈：uni-app + Vue 3 + TypeScript。

页面：

```text
首页
微信快捷登录
绑定手机号
家庭搜索
公开家庭主页
公开家庭树
我的家庭
家庭详情
家庭成员列表树
邀请确认
加入家庭申请
个人中心
```

未绑定手机号允许：浏览部分非互动信息、查看第一个公开家庭树。

必须绑定手机号：创建家庭、加入家庭、留言、查看完整家庭树、接受邀请。

---

# 5. 通用页面状态

所有端都要处理：loading、empty、error、no-permission、not-found、login-required、phone-bind-required、operation-success、operation-failed。

# 6. 必须确认弹窗

删除成员、删除配偶、删除关系、标记无需绑定用户、禁用用户、禁用家庭、下架公开家庭、审核拒绝、创始人转让审核通过、家庭解散审核通过、注销账号。
