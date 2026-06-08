# M12 Family Public Display And Visitor Message Verification

## Environment

- Project: Tree
- Backend: Go + Gin + GORM
- Database: MySQL 8 in Docker Compose
- Redis: Redis 7 in Docker Compose
- Base URL: `http://127.0.0.1:8080`
- Manual verification date: 2026-06-08
- Manual verification families: `familyId=4` for approve/take-down flow, `familyId=5` for reject flow
- Token policy: accessToken values are omitted from this document.

## Codex Self-Test Items

| Item | Status |
|---|---|
| publicdisplay/message code compiles | PASS |
| service tests cover public application creation, duplicate PENDING, cancel, approve, reject, take down | PASS |
| service tests cover visitor message submit, non-public reject, empty/too-long content, rate limits | PASS |
| service tests cover public visitor list only returning APPROVED messages | PASS |
| service tests cover public visitor list not returning visitorPhone / visitorWechat | PASS |
| service tests cover admin approve/reject/delete message and operation logs | PASS |
| service tests cover operation logs not containing visitorPhone / visitorWechat plaintext | PASS |

## Pending Manual Verification

| API | Status |
|---|---|
| POST `/api/families/{familyId}/public-applications` | PASS |
| GET `/api/families/{familyId}/public-applications` | PASS |
| POST `/api/families/{familyId}/public-applications/{applicationId}/cancel` | PASS |
| GET `/api/admin/family-public-applications` | PASS |
| POST `/api/admin/family-public-applications/{applicationId}/approve` | PASS |
| POST `/api/admin/family-public-applications/{applicationId}/reject` | PASS |
| POST `/api/admin/families/{familyId}/take-down-public` | PASS |
| POST `/api/public/families/{familyId}/visitor-messages` | PASS |
| GET `/api/public/families/{familyId}/visitor-messages` | PASS |
| GET `/api/admin/visitor-messages` | PASS |
| POST `/api/admin/visitor-messages/{messageId}/approve` | PASS |
| POST `/api/admin/visitor-messages/{messageId}/reject` | PASS |
| DELETE `/api/admin/visitor-messages/{messageId}` | PASS |

## Manual Verification Results

| Check | Result |
|---|---|
| Admin login succeeded and ADMIN token was kept in memory only | PASS |
| User login succeeded and USER token was kept in memory only | PASS |
| Created two temporary verification families | PASS |
| Submit -> list -> cancel public application | PASS |
| Resubmit -> admin approve public application | PASS |
| Admin reject public application on another family | PASS |
| Approved family public detail accessible | PASS |
| Approved family public tree accessible | PASS |
| Rejected family public detail blocked with `42103` | PASS |
| Visitor submit creates `PENDING` message | PASS |
| Admin visitor-message list returns pending message | PASS |
| Admin approve visitor message | PASS |
| Public visitor-message list returns approved message | PASS |
| Public visitor-message list omits `visitorPhone` and `visitorWechat` | PASS |
| Admin reject visitor message | PASS |
| Admin delete visitor message by status change to `DELETED` | PASS |
| Same IP and family one-minute rate limit returns `45105` | PASS |
| Admin take down public family | PASS |
| Taken-down public tree blocked with `42402` | PASS |
| Taken-down family visitor message submit blocked with `45106` | PASS |
| `graph_version` remained unchanged for M12 public application/review/take-down operations | PASS |
| Required operation log actions were written | PASS |

## Curl Verification Steps

Use environment variables instead of writing tokens into files:

```bash
export BASE_URL="http://127.0.0.1:8080"
export USER_TOKEN="<omitted>"
export ADMIN_TOKEN="<omitted>"
export FAMILY_ID="<family-id>"
```

Submit public application:

```bash
curl -i -X POST "$BASE_URL/api/families/$FAMILY_ID/public-applications" \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"applicationReason":"申请公开展示"}'
```

List family applications:

```bash
curl -i "$BASE_URL/api/families/$FAMILY_ID/public-applications?page=1&pageSize=20" \
  -H "Authorization: Bearer $USER_TOKEN"
```

Admin list and approve:

```bash
curl -i "$BASE_URL/api/admin/family-public-applications?status=PENDING&page=1&pageSize=20" \
  -H "Authorization: Bearer $ADMIN_TOKEN"

curl -i -X POST "$BASE_URL/api/admin/family-public-applications/$APPLICATION_ID/approve" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"reviewComment":"审核通过"}'
```

Submit visitor message after approval:

```bash
curl -i -X POST "$BASE_URL/api/public/families/$FAMILY_ID/visitor-messages" \
  -H "Content-Type: application/json" \
  -d '{"visitorName":"访客","messageContent":"祝家族兴旺"}'
```

List public visitor messages:

```bash
curl -i "$BASE_URL/api/public/families/$FAMILY_ID/visitor-messages?page=1&pageSize=20"
```

Admin handle visitor message:

```bash
curl -i "$BASE_URL/api/admin/visitor-messages?status=PENDING&familyId=$FAMILY_ID&page=1&pageSize=20" \
  -H "Authorization: Bearer $ADMIN_TOKEN"

curl -i -X POST "$BASE_URL/api/admin/visitor-messages/$MESSAGE_ID/approve" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"reviewComment":"内容合规"}'

curl -i -X POST "$BASE_URL/api/admin/visitor-messages/$MESSAGE_ID/reject" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"reviewComment":"不展示"}'

curl -i -X DELETE "$BASE_URL/api/admin/visitor-messages/$MESSAGE_ID" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"deleteReason":"测试删除"}'
```

Take down public family:

```bash
curl -i -X POST "$BASE_URL/api/admin/families/$FAMILY_ID/take-down-public" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"reason":"测试下架"}'
```

## Database Check Commands

Run from `tree-dev`:

```bash
mysql -h mysql -utree_user -ptree_pass tree_platform \
  -e "SELECT id, public_display_status, public_applied_at, public_approved_at, public_taken_down_at, graph_version FROM families WHERE id = $FAMILY_ID;"

mysql -h mysql -utree_user -ptree_pass tree_platform \
  -e "SELECT id, family_id, application_status, reviewed_by_admin_id, reviewed_at, cancelled_at FROM family_public_applications WHERE family_id = $FAMILY_ID ORDER BY id DESC LIMIT 5;"

mysql -h mysql -utree_user -ptree_pass tree_platform \
  -e "SELECT id, family_id, status, reviewed_by_admin_id, reviewed_at, deleted_at FROM visitor_messages WHERE family_id = $FAMILY_ID ORDER BY id DESC LIMIT 5;"
```

## Skipped / Notes

- No full accessToken, password, verification code, inviteToken, visitorPhone, or visitorWechat is stored in this document.
- Visitor message submit intentionally does not write operation_logs, so contact fields cannot enter log detail JSON through submit.
- Local verification reset one existing test user's password hash in the local database only; no code, migration, or secret file was changed.
