# Tree

Tree is a family genealogy platform. This repository currently contains the M1 project skeleton and local Docker development environment.

## Local Development With Docker Dev Container

This project does not require Go, Node.js, or pnpm to be installed directly on the Mac. The development tools run inside the `tree-dev` container.

Start the local environment:

```bash
docker compose -f docker/docker-compose.yml up -d
```

Verify tool versions inside the dev container:

```bash
docker exec tree-dev go version
docker exec tree-dev node -v
docker exec tree-dev pnpm -v
docker exec tree-dev git --version
```

Run backend tests:

```bash
docker exec tree-dev bash -lc "cd /workspace/backend && go test ./..."
```

## Database Migration

Migration files live in `backend/migrations`.

The project is prepared for `golang-migrate`, but the CLI is not installed in the dev container yet. After installing it, run:

```bash
docker exec tree-dev bash -lc 'cd /workspace/backend && migrate -path ./migrations -database "mysql://tree_user:tree_pass@tcp(mysql:3306)/tree_platform?multiStatements=true" up'
```

Until the migration CLI is added, you can validate or apply the SQL files manually against the local MySQL container:

```bash
for file in backend/migrations/*.up.sql; do
  docker exec -i tree-mysql mysql -utree_user -ptree_pass tree_platform < "$file"
done
```

The `000014_seed_root_admin.up.sql` migration inserts `username = admin` with a bcrypt hash placeholder. Replace the hash before production use; do not use or document a plain default password.

## Backend Infrastructure

M3 adds the backend infrastructure foundation:

- Unified API response helpers and staged error code constants.
- Separate JWT manager paths for user tokens and admin tokens.
- Gin middleware skeletons for user auth, admin auth, and phone verification.
- zap logger initialization helpers with safe field masking/hash utilities.
- Redis client initialization and token blacklist interface.
- GORM transaction manager via `TransactionManager.WithTransaction`.
- OperationLogService interface plus a GORM-backed base implementation.

Security notes:

- `JWT_USER_SECRET` and `JWT_ADMIN_SECRET` must be different.
- Do not log access tokens, refresh tokens, verification codes, passwords, `session_key`, AppSecret, full openid, or full unionid.
- Current auth middleware only parses tokens and injects context. Login, logout, token revocation rules, and user/admin status checks are implemented in later stages.

## Admin Authentication

M4 adds backend admin authentication APIs:

```http
POST /api/admin/auth/login
POST /api/admin/auth/logout
GET  /api/admin/me
PUT  /api/admin/me/password
```

Admin tokens use `JWT_ADMIN_SECRET` and are separated from user tokens. Logout writes the current admin token ID into the Redis token blacklist. Password checks use bcrypt hash verification only.

Admin lock settings:

```env
ADMIN_SECURITY_LOCK_MAX_FAILURES=5
ADMIN_SECURITY_LOCK_MINUTES=30
```

`ROOT_ADMIN` login failures are written to `operation_logs`, but do not trigger automatic lockout. `SUPER_ADMIN` and `PLATFORM_ADMIN` can be locked after the configured failure limit. Do not log plaintext passwords, password hashes, access tokens, refresh tokens, verification codes, or other sensitive secrets.

## User Phone Authentication

M5 adds user phone authentication APIs:

```http
POST /api/auth/send-code
POST /api/auth/register-phone
POST /api/auth/login-phone
POST /api/auth/logout
```

Verification codes are stored in `verification_codes` as bcrypt hashes and expire after 5 minutes by default. In non-production environments the response includes `devCode` for local testing; production must not expose or log plaintext verification codes.

Default verification-code settings:

```env
VERIFY_CODE_EXPIRE_SECONDS=300
VERIFY_CODE_COOLDOWN_SECONDS=60
```

Phone registration creates an `ACTIVE` user with `phone_verified = true` and a bcrypt `password_hash`. Phone login uses `JWT_USER_SECRET`, while admin login continues to use `JWT_ADMIN_SECRET`. Logout revokes the current user token through the Redis token blacklist.

## WeChat Mini Program Account Flow

M6 adds account extension APIs:

```http
POST /api/auth/wechat-mini/login
POST /api/auth/wechat-mini/bind-phone
POST /api/auth/change-phone
POST /api/auth/cancel-account
```

Local development can use safe WeChat mock mode:

```env
WECHAT_MOCK_ENABLED=true
WECHAT_MINI_APP_ID=
WECHAT_MINI_APP_SECRET=
```

In production, disable mock mode and configure the real WeChat Mini Program AppID/AppSecret. AppSecret is read only by the backend and must never be logged or exposed to frontend code.

M6 account operations use transactions for phone binding with account merge/claim, phone change, and account cancellation. Account merge checks whether the two accounts are bound to different members in the same family; when a conflict exists, the API returns a conflict error instead of merging automatically. Account merge and cancellation revoke old user tokens through Redis user-token revocation.

## Family Core

M7 adds the authenticated family core APIs and an anonymous public profile endpoint:

```http
POST /api/families
GET  /api/families
GET  /api/families/:familyId
GET  /api/families/:familyId/public
PUT  /api/families/:familyId
POST /api/families/:familyId/dissolution-requests
GET  /api/families/:familyId/dissolution-requests/current
POST /api/families/:familyId/dissolution-requests/:requestId/cancel
```

Creating a family requires an `ACTIVE`, phone-verified user. The family, founder placeholder member, and `FOUNDER` user link are created in one database transaction. Family detail and dissolution status require membership; family updates and dissolution creation require `FOUNDER` or `FAMILY_ADMIN`. Public family details are anonymous but only available for `NORMAL` families whose public display status is `APPROVED`.

Start the backend API from inside the dev container:

```bash
docker exec tree-dev bash -lc "cd /workspace/backend && go run ./cmd/server"
```

Health check:

```bash
curl http://127.0.0.1:8080/api/health
```

Expected response:

```json
{"code":0,"message":"success","data":{"status":"ok"}}
```

## Local Service Addresses

When the backend runs inside the `tree-dev` container, use Docker service names:

```env
MYSQL_HOST=mysql
REDIS_HOST=redis
```

MySQL:

```text
host: mysql
port: 3306
database: tree_platform
user: tree_user
password: tree_pass
root password: root_pass
```

Redis:

```text
host: redis
port: 6379
appendonly: yes
```

## Project Structure

```text
Tree
├── .devcontainer
├── backend
├── docker
├── docs
└── README.md
```
