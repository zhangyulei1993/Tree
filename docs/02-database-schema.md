# 02. 数据库核心表设计草案

> 本文档整理已经确认的核心表。字段仍可在开发前进一步微调，但业务方向已经基本确定。

## 1. 表清单

```text
users
user_auth_identities
user_phone_history
verification_codes
user_account_merge_logs
user_account_claim_logs

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

# 2. users

普通用户账号表，不等于家庭成员节点。

关键字段：

```text
id
phone
phone_verified
email
email_verified
password_hash
nickname
avatar_url
real_name
remark
account_origin
register_client
status
created_by_admin_id
claimed_at
claimed_via
merged_to_user_id
merged_at
cancelled_at
cancelled_reason
disabled_at
disabled_by_admin_id
disabled_reason
deleted_at
deleted_by_admin_id
last_login_at
last_login_client
last_login_ip
created_at
updated_at
```

status：

```text
PENDING_PHONE_BIND
PENDING_CLAIM
ACTIVE
MERGED
DISABLED
CANCELLED
DELETED
```

account_origin：

```text
SELF_REGISTER_PHONE_PC
SELF_REGISTER_PHONE_H5
WECHAT_MINI_PROGRAM_AUTO_CREATED
ADMIN_PRE_CREATED
SYSTEM_CREATED
```

注销时释放手机号：

```text
phone = NULL
phone_verified = 0
status = CANCELLED
```

---

# 3. families

关键字段：

```text
id
family_name
family_surname
native_place
region_code
region_text
description
avatar_url
creator_user_id
current_founder_member_id
status
searchable
public_display_status
public_applied_at
public_approved_at
public_taken_down_at
public_contact_name
public_contact_phone
public_contact_wechat
public_contact_note
public_contact_visible
tree_mode
graph_version
dissolved_at
deleted_at
restored_at
created_at
updated_at
```

status：

```text
NORMAL
DISABLED
DISSOLVED
DELETED
```

public_display_status：

```text
PRIVATE
PENDING
APPROVED
REJECTED
TAKEN_DOWN
```

当前版本一个家庭只允许一个核心姓氏。

---

# 4. family_members

关键字段：

```text
id
family_id
member_type
surname
generation_character
given_name
display_name
gender
birth_date
death_date
birth_order
manual_order
native_place
region_text
is_living
is_lineage_main
is_terminal_node
terminal_reason
manual_lineage_override
lineage_note_type
lineage_note
user_binding_policy
unbound_reason
unbound_note
status
created_by_user_id
created_by_admin_id
deleted_at
deleted_by_user_id
deleted_by_admin_id
created_at
updated_at
```

member_type：

```text
LINEAGE_MEMBER
SPOUSE
EXTERNAL_MEMBER
```

user_binding_policy：

```text
OPTIONAL
REQUIRED
NOT_REQUIRED
DISABLED
```

status：

```text
ACTIVE
DELETED
HIDDEN
```

---

# 5. family_relationships

关键字段：

```text
id
family_id
from_member_id
to_member_id
relationship_type
parent_link_type
relation_note_type
relation_note
status
created_by_user_id
created_by_admin_id
deleted_at
deleted_by_user_id
deleted_by_admin_id
created_at
updated_at
```

relationship_type：

```text
PARENT_CHILD
SPOUSE
```

parent_link_type：

```text
PRIMARY
STEP
ADOPTIVE
SUCCESSION
NOTE_ONLY
OTHER
```

---

# 6. family_member_user_links

关键字段：

```text
id
family_id
member_id
user_id
link_status
link_source
family_role
role_granted_at
role_granted_by_user_id
role_granted_by_admin_id
invitation_id
claimed_at
linked_at
unlinked_at
unlinked_by_user_id
unlinked_by_admin_id
unlinked_reason
merged_from_user_id
merged_to_link_id
merged_at
created_by_user_id
created_by_admin_id
created_at
updated_at
```

family_role：

```text
FOUNDER
FAMILY_ADMIN
MEMBER
```

link_status：

```text
ACTIVE
UNLINKED
CANCELLED_BY_USER
MERGED
REVOKED
ERROR_CORRECTED
```

---

# 7. family_invitations

关键字段：

```text
id
family_id
target_member_id
inviter_user_id
inviter_admin_id
target_user_id
target_phone
invite_type
invite_channel
invite_actor_type
family_role_after_accept
invite_token
status
invite_message
accepted_by_user_id
accepted_at
rejected_by_user_id
rejected_at
reject_reason
cancelled_by_user_id
cancelled_by_admin_id
cancelled_at
cancel_reason
expired_at
created_at
updated_at
```

invite_channel：

```text
IN_APP
SHARE_LINK
```

invite_actor_type：

```text
PLATFORM_ADMIN
FAMILY_FOUNDER
FAMILY_ADMIN
```

---

# 8. 其他表

## verification_codes

用于手机号验证码。有效期 5 分钟，只保存 hash。

## user_auth_identities

用于微信 openid / unionid，不保存明文 session_key。

## user_phone_history

用于记录手机号绑定、换绑、注销释放、账号合并释放历史。

## user_account_merge_logs

账号合并日志。旧 token 立即失效。

## user_account_claim_logs

后台预创建账号认领日志。PC/H5 用预注册手机号注册时，直接认领 PENDING_CLAIM user。

## family_public_applications

家庭公开申请历史。

## visitor_messages

游客留言。

## family_join_requests

用户主动申请加入家庭。

## admin_users

后台管理员账号。

## operation_logs

操作日志。一期支持筛选搜索，不支持导出。

## family_founder_transfer_requests

创始人转让申请。

## family_dissolution_requests

解散家庭申请。
