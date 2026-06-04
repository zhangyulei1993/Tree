# 11. 测试用例与验收标准 v1

> 本文档用于一期开发验收。  
> 范围覆盖账号认证、微信小程序、账号合并与认领、家庭创建、成员关系、邀请、加入申请、公开申请、留言、后台权限、操作日志、缓存与 graph_version。

---

# 1. 测试分级

## 1.1 P0 阻塞级

P0 问题必须修复后才能上线。

```text
登录注册失败
手机号绑定错误
账号合并错误
家庭创建失败
成员关系错乱
权限越权
平台管理员执行高风险审核
删除存在下级成员
一个 user 在同 family 绑定多个 member
一个 member 被多个 user 绑定
操作日志缺失
```

## 1.2 P1 高优先级

```text
公开申请状态异常
留言审核异常
邀请过期判断异常
graph_version 不更新
后台筛选错误
用户注销残留登录状态
```

## 1.3 P2 普通问题

```text
提示文案不清晰
列表排序错误
非关键字段展示异常
部分筛选条件缺失
```

---

# 2. 认证与账号测试用例

## TC-AUTH-001 PC/H5 手机号注册成功

前置条件：

```text
手机号不存在 ACTIVE user
验证码有效
密码符合规则
```

步骤：

```text
发送 REGISTER 验证码
调用 POST /api/auth/register-phone
```

预期：

```text
创建 ACTIVE user
phone_verified = 1
password_hash 已写入
返回 accessToken
写 user_phone_history
写 operation_logs
```

验收标准：

```text
用户可用手机号和密码登录。
```

## TC-AUTH-002 手机号注册命中 PENDING_CLAIM

前置条件：

```text
后台已预创建 users.status = PENDING_CLAIM
phone = 测试手机号
```

步骤：

```text
用户用该手机号注册并验证成功
```

预期：

```text
不新建 user
原 PENDING_CLAIM user -> ACTIVE
claimed_at 写入
写 user_account_claim_logs
写 user_phone_history
写 operation_logs
```

## TC-AUTH-003 手机号密码登录成功

预期：

```text
ACTIVE user 登录成功
last_login_at 更新
返回 token
写 USER_LOGIN 日志
```

## TC-AUTH-004 DISABLED 用户禁止登录

预期：

```text
返回账号已禁用
不返回 token
写失败日志
```

## TC-AUTH-005 微信小程序首次登录

前置条件：

```text
openid 未存在 ACTIVE identity
```

预期：

```text
创建 PENDING_PHONE_BIND user
创建 ACTIVE user_auth_identities
返回 token
needBindPhone = true
```

## TC-AUTH-006 小程序绑定新手机号

预期：

```text
当前 user -> ACTIVE
phone_verified = 1
写 user_phone_history
不触发账号合并
```

## TC-AUTH-007 小程序绑定已存在 ACTIVE 手机号

预期：

```text
触发账号合并
临时 user -> MERGED
微信 identity 迁移到目标 user
旧 token 失效
needRelogin = true
写 merge_logs
```

## TC-AUTH-008 小程序绑定后台预创建手机号

预期：

```text
触发账号认领
预创建 user -> ACTIVE
临时 user -> MERGED
微信 identity 迁移
旧 token 失效
写 claim_logs
```

## TC-AUTH-009 用户注销正常失败：仍在家庭中

前置条件：

```text
用户存在 ACTIVE family_member_user_links
```

预期：

```text
注销失败
提示先退出所有家庭
不清空 phone
```

## TC-AUTH-010 用户注销成功

前置条件：

```text
无 ACTIVE 家庭绑定
非 FOUNDER
无未完成高风险流程
```

预期：

```text
users.status = CANCELLED
users.phone = NULL
phone_verified = 0
nickname/avatar/real_name/password_hash 清空
identity 失效
token 失效
手机号可重新注册
```

---

# 3. 家庭基础测试用例

## TC-FAM-001 创建家庭成功

预期：

```text
创建 families
创建 founder family_member
创建 family_member_user_links
family_role = FOUNDER
current_founder_member_id 正确
graph_version = 1
写 CREATE_FAMILY 日志
```

## TC-FAM-002 未绑定手机号不能创建家庭

预期：

```text
返回请先绑定手机号
不创建 family
```

## TC-FAM-003 游客搜索家庭

预期：

```text
只返回 status = NORMAL 且 searchable = 1 的家庭
返回公开联系方式来自 families.public_contact_*，不是 users.phone
```

## TC-FAM-004 未公开家庭不可查看公开主页

预期：

```text
public_display_status != APPROVED 时返回家庭未公开
```

## TC-FAM-005 家庭公开联系方式为空

预期：

```text
公开主页仍可展示
联系方式字段为 null
返回 contactTip
```

---

# 4. 家庭成员与关系测试用例

## TC-MEM-001 创建普通成员

预期：

```text
family_members 新增
status = ACTIVE
user_binding_policy = OPTIONAL
graph_version + 1
写 CREATE_MEMBER 日志
```

## TC-MEM-002 创建无需绑定用户成员

请求：

```text
userBindingPolicy = NOT_REQUIRED
unboundReason = DECEASED
```

预期：

```text
成员创建成功
userBindingState = NOT_REQUIRED
无 invitation
无 family_member_user_links
```

## TC-MEM-003 普通 MEMBER 只能编辑本人资料

预期：

```text
可以改本人 displayName / birthDate / nativePlace 等
不能改关系、角色、绑定状态
```

## TC-MEM-004 删除存在下级成员失败

前置条件：

```text
member 存在 ACTIVE 子女关系
```

预期：

```text
删除失败
返回 descendantCount
写 FAILED operation_logs
graph_version 不变
```

## TC-MEM-005 删除主成员且有配偶需要确认

预期：

```text
未传 confirmDeleteSpouse -> 返回需确认
传 true -> 主成员、配偶、配偶关系软删除
graph_version + 1
```

## TC-MEM-006 添加兄弟姐妹但无父母节点

预期：

```text
返回“请先创建父亲或母亲节点”
不创建成员
不创建关系
```

## TC-MEM-007 添加兄弟姐妹且有父母节点

预期：

```text
创建新 sibling member
为已有父母分别创建 PARENT_CHILD 到新 member
不创建 SIBLING 关系
graph_version + 1
```

## TC-MEM-008 重复 PRIMARY 父亲冲突

预期：

```text
同一 child 已有 PRIMARY 父亲时，再添加 PRIMARY 父亲应失败
STEP / NOTE_ONLY 可允许
```

## TC-MEM-009 添加多个配偶

预期：

```text
允许多个 SPOUSE 节点
每个配偶通过 SPOUSE relationship 关联
```

---

# 5. 邀请与加入家庭测试用例

## TC-INV-001 平台管理员站内邀请已注册用户

预期：

```text
手机号搜索到 user
创建 IN_APP invitation
target_user_id 有值
状态 PENDING
```

## TC-INV-002 平台管理员搜索不到手机号

预期：

```text
不能创建邀请
返回可标记无需绑定用户
不能生成 SHARE_LINK
```

## TC-INV-003 家族管理员生成分享邀请链接

预期：

```text
创建 SHARE_LINK invitation
target_user_id 可为空
生成 inviteToken
有效期 7 天
```

## TC-INV-004 接受邀请成功

预期：

```text
invitation -> ACCEPTED
创建 ACTIVE family_member_user_links
family_role = MEMBER
不触发 graph_version
```

## TC-INV-005 接受已过期邀请失败

预期：

```text
访问时更新 invitation.status = EXPIRED
接受失败
写 EXPIRE_INVITATION 系统日志
```

## TC-JOIN-001 用户主动申请加入家庭

预期：

```text
用户已登录且绑定手机号
创建 PENDING family_join_requests
不创建 member link
```

## TC-JOIN-002 通过申请并绑定已有成员

预期：

```text
创建 ACTIVE link
request_status = APPROVED
handle_result = APPROVED_BIND_EXISTING_MEMBER
不触发 graph_version
```

## TC-JOIN-003 通过申请并创建新成员

预期：

```text
创建 member
创建 ACTIVE link
默认 family_role = MEMBER
graph_version + 1
```

---

# 6. 家庭公开与留言测试用例

## TC-PUBLIC-001 家庭创始人申请公开

预期：

```text
创建 family_public_applications
application_status = PENDING
families.public_display_status = PENDING
```

## TC-PUBLIC-002 同一家庭重复提交 PENDING 公开申请

预期：

```text
第二次提交失败
提示已存在待审核公开申请
```

## TC-PUBLIC-003 平台管理员不能审核自己代发起的公开申请

预期：

```text
返回无权审核自己发起的申请
```

## TC-PUBLIC-004 审核通过公开申请

预期：

```text
application_status = APPROVED
families.public_display_status = APPROVED
公开主页可访问
```

## TC-PUBLIC-005 下架公开家庭

预期：

```text
families.public_display_status = TAKEN_DOWN
公开主页不可访问
写 TAKE_DOWN_PUBLIC_DISPLAY 日志
```

## TC-MSG-001 匿名游客留言

预期：

```text
不登录也可提交
联系方式非必填
status = PENDING
审核前不展示
```

## TC-MSG-002 审核通过留言

预期：

```text
status = APPROVED
公开留言列表按 created_at DESC 展示
不公开 visitorPhone / visitorWechat
```

## TC-MSG-003 IP 频率限制

预期：

```text
同一 IP 1 分钟第二次提交失败
同一 IP 一天超过 20 条失败
```

---

# 7. 家庭角色、转让与解散测试用例

## TC-ROLE-001 FOUNDER 设置 FAMILY_ADMIN

预期：

```text
目标 member 有 ACTIVE link
family_role -> FAMILY_ADMIN
不触发 graph_version
```

## TC-ROLE-002 NOT_REQUIRED 成员不能设为管理员

预期：

```text
操作失败
```

## TC-TRANSFER-001 发起创始人转让

预期：

```text
只有 FOUNDER 可发起
目标成员必须有 ACTIVE link
创建 PENDING request
```

## TC-TRANSFER-002 PLATFORM_ADMIN 不能审核创始人转让

预期：

```text
审核失败
写 FAILED operation_logs
```

## TC-TRANSFER-003 SUPER_ADMIN 审核通过创始人转让

预期：

```text
request_status = COMPLETED
review_result = APPROVED
families.current_founder_member_id = 新 member
原创始人 role = MEMBER
新创始人 role = FOUNDER
```

## TC-DISSOLVE-001 发起家庭解散申请

预期：

```text
只有 FOUNDER 可发起
同 family 不允许多个 PENDING
```

## TC-DISSOLVE-002 审核通过家庭解散

预期：

```text
families.status = DISSOLVED
public_display_status = PRIVATE
searchable = 0
历史成员、关系、留言、邀请保留
```

## TC-RESTORE-001 恢复家庭

预期：

```text
ROOT_ADMIN / SUPER_ADMIN 可恢复
恢复进入编辑流程
publicDisplayStatus 默认 PRIVATE
历史邀请不重新激活
```

---

# 8. 后台权限测试用例

## TC-ADMIN-001 admin 有且只有一个

预期：

```text
不能创建第二个 ROOT_ADMIN
不能删除 / 禁用 / 降级 admin
```

## TC-ADMIN-002 SUPER_ADMIN 不能管理同级

预期：

```text
不能禁用 / 删除其他 SUPER_ADMIN
```

## TC-ADMIN-003 PLATFORM_ADMIN 不能管理后台管理员

预期：

```text
访问 /api/admin/admin-users 返回无权限
```

## TC-ADMIN-004 ROOT_ADMIN 登录失败不自动锁定

预期：

```text
多次失败不进入 LOCKED
但写登录失败日志
```

## TC-ADMIN-005 SUPER_ADMIN / PLATFORM_ADMIN 登录失败锁定

预期：

```text
达到配置阈值进入 LOCKED
只有 ROOT_ADMIN 可解锁
```

## TC-ADMIN-006 PLATFORM_ADMIN 禁用普通用户

预期：

```text
可禁用普通用户
若目标是 FOUNDER，应提示高风险升级处理
```

## TC-ADMIN-007 PLATFORM_ADMIN 不能恢复家庭

预期：

```text
访问恢复接口失败
```

---

# 9. 操作日志测试用例

## TC-LOG-001 后台登录成功写日志

预期：

```text
module = AUTH
action = ADMIN_LOGIN_SUCCESS
operator_type = ADMIN
```

## TC-LOG-002 后台登录失败写日志

预期：

```text
module = AUTH
action = ADMIN_LOGIN_FAILED
result = FAILED
```

## TC-LOG-003 删除存在下级成员失败写日志

预期：

```text
module = FAMILY_MEMBER
action = DELETE_MEMBER
result = FAILED
error_message 有明确原因
```

## TC-LOG-004 平台管理员越权审核解散写日志

预期：

```text
module = FAMILY_DISSOLUTION
action = APPROVE_DISSOLUTION
result = FAILED
```

## TC-LOG-005 日志列表筛选

筛选条件：

```text
operatorType
operatorRole
module
action
familyId
memberId
userId
result
time_range
```

预期：

```text
返回结果符合条件
按 created_at DESC 排序
```

## TC-LOG-006 日志详情敏感字段过滤

预期不返回：

```text
password_hash
验证码
accessToken
refreshToken
session_key
完整 openid
完整 unionid
```

---

# 10. 缓存与 graph_version 验收

## TC-GRAPH-001 创建成员触发 graph_version + 1

## TC-GRAPH-002 删除成员触发 graph_version + 1

## TC-GRAPH-003 添加关系触发 graph_version + 1

## TC-GRAPH-004 删除关系触发 graph_version + 1

## TC-GRAPH-005 接受邀请不触发 graph_version

## TC-GRAPH-006 设置家族管理员不触发 graph_version

## TC-GRAPH-007 家庭公开申请不触发 graph_version

## TC-GRAPH-008 家庭树缓存 key 包含 graph_version

建议缓存 key：

```text
family:{familyId}:tree:v{graphVersion}
```

---

# 11. 一期总体验收标准

## 11.1 功能验收

```text
用户可通过 PC/H5 注册登录。
微信小程序可快捷登录并绑定手机号。
账号认领和账号合并流程正确。
家庭可创建、搜索、查看、编辑。
成员可创建、编辑、删除、标记无需绑定。
关系可通过父母、子女、兄弟姐妹、配偶入口创建。
邀请和加入申请流程可闭环。
家庭公开申请、审核、下架可闭环。
游客留言、审核、展示可闭环。
家庭角色、创始人转让、解散、恢复可闭环。
后台用户、家庭、管理员管理可用。
操作日志可查询、筛选、查看详情。
```

## 11.2 权限验收

```text
普通成员不能管理其他成员。
家族管理员不能设置管理员或转让创始人。
平台管理员不能审核创始人转让或家庭解散。
平台管理员不能生成分享邀请链接。
平台管理员不能恢复家庭。
超级管理员不能管理同级超级管理员。
ROOT_ADMIN 有且只有一个。
```

## 11.3 数据一致性验收

```text
一个 user 在同一 family 只能绑定一个 ACTIVE member。
一个 member 只能被一个 ACTIVE user 绑定。
同一 family 只能有一个 FOUNDER。
同一 member 同时只能有一个 PENDING invitation。
同一 family 同时只能有一个 PENDING 公开申请。
同一 family 同时只能有一个 PENDING 创始人转让申请。
同一 family 同时只能有一个 PENDING 解散申请。
删除成员不能留下非法树结构。
```

## 11.4 安全验收

```text
验证码只保存 hash。
密码只保存 password_hash。
token 失效机制有效。
账号合并后旧 token 失效。
注销后手机号释放。
日志不泄露密码、验证码、token、session_key。
后台敏感操作必须写日志。
```

## 11.5 非功能验收

```text
家庭树接口一次性加载 members + relationships 后组装。
家庭树可通过 graph_version 缓存。
后台列表接口支持分页。
日志列表支持筛选。
游客留言有基础频率限制。
常见接口错误有明确错误码和提示。
```

---

# 12. 上线前检查清单

```text
数据库表结构已创建
核心唯一性约束已实现
接口权限中间件已实现
用户 token 与 admin token 分离
验证码服务可用
微信小程序登录配置可用
operation_logs 写入工具可用
graph_version 更新工具可用
错误码常量已整理
后台页面至少覆盖 P0 流程
测试用例 P0 全部通过
```
