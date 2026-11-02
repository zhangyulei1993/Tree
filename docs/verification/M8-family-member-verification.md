# M8 Family Member Verification

## Verification Result

- Stage: M8 - Family Member
- Result: PASS
- Verification type: Manual API verification and automated Go tests
- Verification date: 2026-06-07

## Verification Environment

- Branch: `m8-family-member`
- Base URL: `http://127.0.0.1:8080`
- Runtime: `tree-dev` Docker development container
- Database: MySQL service `tree-mysql`, database `tree_platform`
- Redis: Redis service `tree-redis`
- Test user ID: `8`
- Test phone: `13715460762`
- Family ID: `2`
- Family role: `FOUNDER`
- Created test member ID: `6`

The USER access token was used only as an HTTP Authorization credential. The
complete access token is intentionally omitted from this document and was not
written to any project file.

## Command Summary

The manual verification used commands equivalent to the following:

```bash
curl http://127.0.0.1:8080/api/health

curl -X POST http://127.0.0.1:8080/api/auth/login-phone \
  -H 'Content-Type: application/json' \
  -d '{"phone":"13715460762","password":"<OMITTED>","clientType":"H5_WEB"}'

curl http://127.0.0.1:8080/api/families \
  -H 'Authorization: Bearer <USER_ACCESS_TOKEN_OMITTED>'

curl -X POST http://127.0.0.1:8080/api/families/2/members \
  -H 'Authorization: Bearer <USER_ACCESS_TOKEN_OMITTED>' \
  -H 'Content-Type: application/json' \
  -d '{"name":"M8测试成员","gender":"MALE","birthYear":1992,"isAlive":true}'

curl http://127.0.0.1:8080/api/families/2/members \
  -H 'Authorization: Bearer <USER_ACCESS_TOKEN_OMITTED>'

curl http://127.0.0.1:8080/api/families/2/members/6 \
  -H 'Authorization: Bearer <USER_ACCESS_TOKEN_OMITTED>'

curl -X PUT http://127.0.0.1:8080/api/families/2/members/6 \
  -H 'Authorization: Bearer <USER_ACCESS_TOKEN_OMITTED>' \
  -H 'Content-Type: application/json' \
  -d '{"name":"M8测试成员-已更新"}'

curl -X DELETE http://127.0.0.1:8080/api/families/2/members/6 \
  -H 'Authorization: Bearer <USER_ACCESS_TOKEN_OMITTED>' \
  -H 'Content-Type: application/json' \
  -d '{"reason":"人工验证软删除"}'
```

Automated tests:

```bash
cd backend
go test ./...
```

The closeout test was executed in the `tree-dev` development container on
2026-06-07. All packages compiled and `go test ./...` completed successfully.
The `family/member/service` tests passed; packages without test files were also
loaded successfully.

## Verification Checklist

| Item | Result | Notes |
|---|---|---|
| Backend service startup | PASS | API service started successfully |
| `go test ./...` | PASS | Executed in `tree-dev`; all packages passed |
| `GET /api/health` | PASS | Returned success |
| `POST /api/auth/login-phone` | PASS | USER token obtained; token omitted |
| `GET /api/families` | PASS | Family `2` returned with role `FOUNDER` |
| `POST /api/families/2/members` | PASS | Created member `6` |
| `GET /api/families/2/members` | PASS | Created member appeared in list |
| `GET /api/families/2/members/6` | PASS | Member detail returned successfully |
| `PUT /api/families/2/members/6` | PASS | Member update succeeded |
| `DELETE /api/families/2/members/6` | PASS | Member soft deletion succeeded |
| Query member `6` after deletion | PASS | Returned error code `43101` |
| `POST /api/families/2/members/6/bind-user` | SKIPPED | Not included in this manual verification |
| `POST /api/families/2/members/6/unbind-user` | SKIPPED | Not included in this manual verification |

## Final Conclusion

The manually verified M8 family-member creation, listing, detail, update, and
soft-delete workflow passed. Querying the deleted member returned business
error code `43101` as expected.

The `bind-user` and `unbind-user` endpoints remain unverified in this manual
verification and are explicitly marked as `SKIPPED`.
