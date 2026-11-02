# Tree Database Schema Overview

数据库：MySQL 8，默认字符集 `utf8mb4`，本地库名 `tree_platform`。

M2 migrations 创建 18 张表。当前 SQL 没有声明外键，表关系通过 ID 字段、索引和 service 层约束维护。

## Account Domain

| 表 | 用途 |
|---|---|
| `users` | 平台普通用户账号。保存手机号、密码 hash、账号状态、合并/认领/注销状态。 |
| `user_auth_identities` | 用户第三方身份，目前用于微信小程序 openid/unionid 与用户的绑定。 |
| `verification_codes` | 手机验证码记录，只保存 `code_hash`，包含 scene、client type、有效期和使用状态。 |
| `user_phone_history` | 绑定、换绑、注销、认领等手机号历史。 |
| `user_account_merge_logs` | 账号合并审计，记录 source user 与 target user。 |
| `user_account_claim_logs` | PENDING_CLAIM 账号认领审计。 |

主要关系：

- `user_auth_identities.user_id -> users.id`
- 三类账号日志通过 user ID 指向 `users`
- 合并后 source user 的 `merged_to_user_id` 指向保留的 target user

## Family Domain

| 表 | 用途 |
|---|---|
| `families` | 家庭主体、唯一主姓氏、公开状态、创始 member、树版本。 |
| `family_members` | 家谱节点。与 `users` 是不同实体，可在没有平台账号时独立存在。 |
| `family_relationships` | member 之间的 `PARENT_CHILD` 或 `SPOUSE` 关系；不存储 `SIBLING`。 |
| `family_member_user_links` | 将平台 user 绑定到某家庭中的 member，并保存 `FOUNDER`、`FAMILY_ADMIN`、`MEMBER` 角色。 |

主要关系：

- `family_members.family_id -> families.id`
- `family_relationships.family_id -> families.id`
- `from_member_id`、`to_member_id -> family_members.id`
- `family_member_user_links` 同时连接 `families`、`family_members`、`users`
- `families.current_founder_member_id -> family_members.id`

M7 创建家庭时，会在一个事务中创建 `families`、创建者占位 `family_members`、`FOUNDER` link，并更新 `current_founder_member_id`。

## Family Workflow Domain

| 表 | 用途 |
|---|---|
| `family_invitations` | 邀请用户绑定目标 member，M1-M7 尚无 API。 |
| `family_join_requests` | 用户申请加入家庭，M1-M7 尚无 API。 |
| `family_public_applications` | 家庭公开展示申请及审核，M1-M7 尚无 API。 |
| `visitor_messages` | 公开家庭访客留言，M1-M7 尚无 API。 |
| `family_founder_transfer_requests` | 创始人转让申请，M1-M7 尚无 API。 |
| `family_dissolution_requests` | 家庭解散申请。M7 已实现创建、查询当前 PENDING、取消。 |

解散申请保存 requester user/member。M7 创建申请时将家庭状态改为 `DISSOLUTION_PENDING`，取消后恢复 `NORMAL`；真正审核和解散尚未实现。

## Administration And Audit

| 表 | 用途 |
|---|---|
| `admin_users` | ROOT_ADMIN、SUPER_ADMIN、PLATFORM_ADMIN 后台账号和锁定状态。 |
| `operation_logs` | 用户、管理员和系统关键操作审计。 |

`000014_seed_root_admin` 插入 username 为 `admin` 的 ROOT_ADMIN，密码字段为 bcrypt hash。生产部署必须替换该 hash。

## Status And Integrity Notes

- 状态字段使用 `VARCHAR`，不使用 MySQL ENUM。
- 多数核心表有 `created_at`、`updated_at`；部分日志表只有 `created_at`。
- 支持软删除的实体使用 `deleted_at`。
- migrations 中的多项唯一性和复杂约束标为 service TODO，例如每家庭唯一 FOUNDER、每 member 唯一 ACTIVE user link。
- `families.graph_version` 用于家谱结构版本；M7 的基础资料与解散申请不会递增它。
