# M13 Family Role Transfer Dissolution Verification

## Environment

- Project: Tree
- Backend: Go + Gin + GORM
- Database: MySQL 8 in Docker Compose
- Base URL: `http://127.0.0.1:8080`
- Token policy: accessToken values must be kept in memory or shell variables only and omitted from this document.

## Codex Self-Test Items

| Item | Status |
|---|---|
| role service tests cover set/unset admin rules | PASS |
| founder transfer service tests cover create/cancel/review rules | PASS |
| dissolution service tests cover admin review and restore rules | PASS |
| `go test ./...` | PASS |
| `go vet ./...` | PASS |

## Pending Manual Verification

| API | Status |
|---|---|
| POST `/api/families/{familyId}/members/{memberId}/set-admin` | PENDING |
| POST `/api/families/{familyId}/members/{memberId}/unset-admin` | PENDING |
| POST `/api/families/{familyId}/founder-transfer-requests` | PENDING |
| POST `/api/families/{familyId}/founder-transfer-requests/{requestId}/cancel` | PENDING |
| GET `/api/admin/founder-transfer-requests` | PENDING |
| POST `/api/admin/founder-transfer-requests/{requestId}/approve` | PENDING |
| POST `/api/admin/founder-transfer-requests/{requestId}/reject` | PENDING |
| POST `/api/families/{familyId}/dissolution-requests` | PENDING |
| GET `/api/families/{familyId}/dissolution-requests/current` | PENDING |
| POST `/api/families/{familyId}/dissolution-requests/{requestId}/cancel` | PENDING |
| GET `/api/admin/dissolution-requests` | PENDING |
| POST `/api/admin/dissolution-requests/{requestId}/approve` | PENDING |
| POST `/api/admin/dissolution-requests/{requestId}/reject` | PENDING |
| POST `/api/admin/families/{familyId}/restore` | PENDING |

## Curl Verification Steps

Use shell variables and do not write tokens into files:

```bash
export BASE_URL="http://127.0.0.1:8080"
export USER_TOKEN="<omitted>"
export ADMIN_TOKEN="<omitted>"
export FAMILY_ID="<family-id>"
export MEMBER_ID="<bound-member-id>"
```

Role:

```bash
curl -i -X POST "$BASE_URL/api/families/$FAMILY_ID/members/$MEMBER_ID/set-admin" \
  -H "Authorization: Bearer $USER_TOKEN" -H "Content-Type: application/json" -d '{}'

curl -i -X POST "$BASE_URL/api/families/$FAMILY_ID/members/$MEMBER_ID/unset-admin" \
  -H "Authorization: Bearer $USER_TOKEN" -H "Content-Type: application/json" -d '{}'
```

Founder transfer:

```bash
curl -i -X POST "$BASE_URL/api/families/$FAMILY_ID/founder-transfer-requests" \
  -H "Authorization: Bearer $USER_TOKEN" -H "Content-Type: application/json" \
  -d '{"toMemberId":'$MEMBER_ID',"requestReason":"verify transfer"}'

curl -i "$BASE_URL/api/admin/founder-transfer-requests?status=PENDING&familyId=$FAMILY_ID" \
  -H "Authorization: Bearer $ADMIN_TOKEN"

curl -i -X POST "$BASE_URL/api/admin/founder-transfer-requests/$REQUEST_ID/approve" \
  -H "Authorization: Bearer $ADMIN_TOKEN" -H "Content-Type: application/json" -d '{}'
```

Dissolution and restore:

```bash
curl -i -X POST "$BASE_URL/api/families/$FAMILY_ID/dissolution-requests" \
  -H "Authorization: Bearer $USER_TOKEN" -H "Content-Type: application/json" \
  -d '{"requestReason":"verify dissolution"}'

curl -i "$BASE_URL/api/admin/dissolution-requests?status=PENDING&familyId=$FAMILY_ID" \
  -H "Authorization: Bearer $ADMIN_TOKEN"

curl -i -X POST "$BASE_URL/api/admin/dissolution-requests/$REQUEST_ID/approve" \
  -H "Authorization: Bearer $ADMIN_TOKEN" -H "Content-Type: application/json" -d '{}'

curl -i -X POST "$BASE_URL/api/admin/families/$FAMILY_ID/restore" \
  -H "Authorization: Bearer $ADMIN_TOKEN" -H "Content-Type: application/json" -d '{}'
```

## Database Check Commands

```bash
mysql -h mysql -utree_user -ptree_pass tree_platform \
  -e "SELECT id,status,public_display_status,searchable,current_founder_member_id,graph_version FROM families WHERE id=$FAMILY_ID;"

mysql -h mysql -utree_user -ptree_pass tree_platform \
  -e "SELECT id,member_id,user_id,family_role,link_status FROM family_member_user_links WHERE family_id=$FAMILY_ID ORDER BY id;"

mysql -h mysql -utree_user -ptree_pass tree_platform \
  -e "SELECT id,request_status,review_result,reviewed_by_admin_id FROM family_founder_transfer_requests WHERE family_id=$FAMILY_ID ORDER BY id DESC LIMIT 5;"

mysql -h mysql -utree_user -ptree_pass tree_platform \
  -e "SELECT id,request_status,review_result,reviewed_by_admin_id FROM family_dissolution_requests WHERE family_id=$FAMILY_ID ORDER BY id DESC LIMIT 5;"
```

## Skipped / Notes

- Manual curl verification has not been executed yet.
- No full accessToken, password, verification code, inviteToken, phone, openid, or unionid is stored in this document.
