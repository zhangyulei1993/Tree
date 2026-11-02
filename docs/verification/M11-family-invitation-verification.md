# M11 Family Invitation Verification

## Status

- Automated service tests: PASS
- Full backend tests: PASS
- `go vet ./...`: PASS
- Manual API verification: PENDING
- Access token, password, verification code and original invite token: omitted

## Codex Self-Test Items

| Check | Status |
|---|---|
| Share and in-app invitation creation | PASS |
| Seven-day expiration | PASS |
| Token stored as SHA-256 digest | PASS |
| Duplicate, NOT_REQUIRED and bound-member checks | PASS |
| In-app target-user check | PASS |
| Accept creates ACTIVE MEMBER link | PASS |
| Accept does not change graph_version | PASS |
| Expired/non-PENDING accept rejected | PASS |
| Reject/cancel and cancellation permission | PASS |
| Join request creation and duplicate check | PASS |
| Already-linked applicant rejected | PASS |
| Family-admin processing permission | PASS |
| BIND_EXISTING_MEMBER approval | PASS |
| CREATE_NEW_MEMBER approval and graph_version +1 | PASS |
| Reject and applicant-only cancel | PASS |
| Sensitive fields absent from audit input | PASS |

## Pending Manual Verification

All 12 M11 HTTP routes, transaction rollback, database state and operation logs remain `PENDING` until executed against MySQL.

## Curl Verification

```bash
export BASE_URL=http://127.0.0.1:8080
export ADMIN_USER_TOKEN='<OMITTED>'
export APPLICANT_TOKEN='<OMITTED>'
export FAMILY_ID='<FAMILY_ID>'
export MEMBER_ID='<UNBOUND_MEMBER_ID>'

curl -X POST "$BASE_URL/api/families/$FAMILY_ID/members/$MEMBER_ID/invite" \
  -H "Authorization: Bearer $ADMIN_USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"inviteChannel":"SHARE_LINK"}'

# Keep the returned token only in the shell session.
export INVITE_TOKEN='<OMITTED>'
curl "$BASE_URL/api/invitations/$INVITE_TOKEN"

curl -X POST "$BASE_URL/api/families/$FAMILY_ID/join-requests" \
  -H "Authorization: Bearer $APPLICANT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"applicantMessage":"M11 verification"}'
```

Continue with accept/reject/cancel and both approval modes using disposable users and members.

## Database Checks

```sql
SELECT id, family_id, target_member_id, invite_channel, status, expired_at
FROM family_invitations ORDER BY id DESC;

SELECT id, family_id, applicant_user_id, request_status, approve_mode,
       bound_member_id, created_member_id
FROM family_join_requests ORDER BY id DESC;

SELECT family_id, member_id, user_id, link_status, link_source,
       invitation_id, join_request_id
FROM family_member_user_links ORDER BY id DESC;

SELECT id, graph_version FROM families WHERE id = <FAMILY_ID>;

SELECT module, action, result, family_id, member_id, target_id
FROM operation_logs
WHERE module IN ('FAMILY_INVITATION','FAMILY_JOIN_REQUEST')
ORDER BY id DESC;
```

## Skipped

- PLATFORM_ADMIN invitation/approval assistance: intentionally outside M11.
- Admin routes: intentionally not implemented.
- Frontend and mini-program flows: outside this backend stage.
