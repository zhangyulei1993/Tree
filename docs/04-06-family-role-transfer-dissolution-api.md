# 04-06. 家庭角色、创始人转让与家庭解散接口完整规格 v1

## 1. 模块范围

覆盖：

```text
设置家族管理员
取消家族管理员
发起创始人转让申请
后台审核创始人转让
撤销创始人转让
发起家庭解散申请
后台审核家庭解散
撤销家庭解散
后台恢复家庭
```

涉及表：

```text
families
family_members
family_member_user_links
family_founder_transfer_requests
family_dissolution_requests
admin_users
operation_logs
```

## 2. 核心规则

```text
家庭角色存放在 family_member_user_links.family_role
当前家庭角色：FOUNDER / FAMILY_ADMIN / MEMBER
families.current_founder_member_id 必须与 FOUNDER link 一致
无 ACTIVE user 绑定或 user_binding_policy = NOT_REQUIRED 的节点不能成为 FOUNDER / FAMILY_ADMIN
创始人不能直接退出家庭，必须先完成创始人转让
创始人转让和家庭解散必须由 ROOT_ADMIN / SUPER_ADMIN 审核
PLATFORM_ADMIN 不能审核创始人转让和家庭解散
家庭解散后不物理删除数据
家庭恢复一期不单独建恢复申请表
```

## 3. POST /api/families/{familyId}/members/{memberId}/set-admin

设置家族管理员。

权限：FOUNDER。

校验：

```text
family.status = NORMAL
目标 member ACTIVE
目标 member 有 ACTIVE link
目标 member.user_binding_policy != NOT_REQUIRED
目标 member 当前不是 FOUNDER
```

影响：

```text
family_member_user_links.family_role = FAMILY_ADMIN
写 operation_logs
不触发 graph_version
```

## 4. POST /api/families/{familyId}/members/{memberId}/unset-admin

取消家族管理员。

权限：FOUNDER。

影响：

```text
family_member_user_links.family_role = MEMBER
写 operation_logs
不触发 graph_version
```

## 5. POST /api/families/{familyId}/founder-transfer-requests

发起创始人转让申请。

权限：FOUNDER。

请求：

```json
{
  "toMemberId": 20002,
  "requestReason": "本人不再维护该家庭，转让给其他成员管理"
}
```

校验：

```text
当前 user 是 FOUNDER
toMember 属于同 family
toMember ACTIVE
toMember 有 ACTIVE link
toMember.user_binding_policy != NOT_REQUIRED
同 family 不存在 PENDING 转让申请
```

## 6. GET /api/admin/founder-transfer-requests

后台创始人转让申请列表。

权限：

```text
ROOT_ADMIN
SUPER_ADMIN
```

## 7. POST /api/admin/founder-transfer-requests/{requestId}/approve

审核通过创始人转让。

权限：

```text
ROOT_ADMIN
SUPER_ADMIN
```

事务内：

```text
request_status = COMPLETED
review_result = APPROVED
families.current_founder_member_id = to_member_id
原创始人 link.family_role = MEMBER
新创始人 link.family_role = FOUNDER
写 operation_logs
```

不触发 graph_version。

## 8. POST /api/admin/founder-transfer-requests/{requestId}/reject

拒绝创始人转让，不修改家庭创始人。

## 9. POST /api/families/{familyId}/founder-transfer-requests/{requestId}/cancel

申请人撤销未审核转让申请。

## 10. POST /api/families/{familyId}/dissolution-requests

发起家庭解散申请。

权限：FOUNDER。

校验：

```text
family.status = NORMAL
当前 user 是 FOUNDER
同 family 不存在 PENDING 解散申请
```

## 11. GET /api/admin/dissolution-requests

后台解散申请列表。

权限：

```text
ROOT_ADMIN
SUPER_ADMIN
```

## 12. POST /api/admin/dissolution-requests/{requestId}/approve

审核通过家庭解散。

事务内：

```text
request_status = COMPLETED
review_result = APPROVED
families.status = DISSOLVED
families.dissolved_at = 当前时间
families.searchable = 0
families.public_display_status = PRIVATE
写 operation_logs
```

不触发 graph_version。

## 13. POST /api/admin/dissolution-requests/{requestId}/reject

拒绝家庭解散，不修改家庭状态。

## 14. POST /api/families/{familyId}/dissolution-requests/{requestId}/cancel

申请人撤销未审核解散申请。

## 15. POST /api/admin/families/{familyId}/restore

恢复家庭。

权限：

```text
ROOT_ADMIN
SUPER_ADMIN
```

请求：

```json
{
  "restoreStatus": "NORMAL",
  "publicDisplayStatus": "PRIVATE",
  "searchable": true,
  "restoreReason": "误操作解散，确认恢复"
}
```

规则：

```text
恢复不是简单回滚，而是进入恢复编辑流程
恢复后 publicDisplayStatus 建议 PRIVATE
历史邀请不重新激活
```

## 16. 错误码范围

```text
46000-46999
```
