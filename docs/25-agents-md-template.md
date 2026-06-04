# 25. AGENTS.md 模板

> 将本文件内容复制到项目根目录 `AGENTS.md`。Codex 会优先读取该文件。

```md
# AGENTS.md

## Project

This project is a family genealogy platform.

## Tech Stack

Backend:
- Go
- Gin
- GORM
- MySQL 8
- Redis 7
- golang-migrate
- zap
- JWT

Frontend:
- Vue 3
- TypeScript
- Vite
- Element Plus

Mini Program:
- uni-app
- Vue 3
- TypeScript

## Important Docs

Always read:
- docs/00-TECH-STACK-FINAL.md
- docs/02-GO-BACKEND-ARCHITECTURE.md
- docs/CODEX_TASKS.md
- docs/16-development-milestones.md

For UI tasks, also read:
- docs/18-design-language.md
- docs/19-core-wireframes.md

## Non-negotiable Business Rules

1. Do not merge `users` and `family_members`.
2. `family_member` can exist without a user account.
3. `relationship_type` only supports `PARENT_CHILD` and `SPOUSE`.
4. Do not store `SIBLING` in the database.
5. `ADD_SIBLING` must be implemented through shared parent-child relationships.
6. `PLATFORM_ADMIN` cannot approve founder transfer.
7. `PLATFORM_ADMIN` cannot approve family dissolution.
8. `PLATFORM_ADMIN` cannot generate share invitation links.
9. Important operations must write `operation_logs`.
10. Tree structure changes must update `graph_version`.
11. User token and admin token must be separated.
12. Verification codes must not be stored in plain text.
13. Passwords must not be stored in plain text.
14. Tokens, verification codes, session_key, AppSecret, openid and unionid must not be logged in full.

## Go Backend Rules

1. Handler only handles HTTP input/output.
2. Business rules must be in service.
3. Database access must be in repository.
4. Complex operations must use transactions.
5. Do not hard-code test users in production code.
6. Do not log sensitive data.
7. Do not put complex business logic in Gin handlers.
8. Use GORM transactions for multi-table writes.
9. Use OperationLogService for audit logs.
10. Use GraphVersionService for tree updates.

## UI Rules

1. Use the Genealogy Calm Design System.
2. Admin pages use Element Plus.
3. Public web uses warm, calm and trustworthy style.
4. Mini program pages should be short and task-focused.
5. Do not connect real APIs in prototype tasks unless explicitly requested.
6. Use mock data for UI prototypes.
7. Do not introduce a new UI framework without approval.

## Before Finishing Each Task

Run if applicable:
- gofmt on changed Go files
- go test ./...
- pnpm build for changed frontend apps

Then report:
1. Changed files
2. New APIs
3. New database tables / fields
4. Tests run
5. Test results
6. Known limitations
7. Next recommended task

## Forbidden

Do not:
1. Change the tech stack.
2. Generate Java / Spring Boot backend.
3. Generate Node.js / NestJS backend.
4. Merge users and family_members.
5. Store SIBLING in relationships.
6. Skip permission checks.
7. Skip operation_logs.
8. Skip graph_version.
9. Use mock data in production code.
10. Commit .env secrets.
```
