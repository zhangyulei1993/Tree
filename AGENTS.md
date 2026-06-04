# AGENTS.md

## Project

This project is a family genealogy platform.

The project includes:

- Go backend API
- Admin web dashboard
- PC/H5 public web
- WeChat Mini Program
- MySQL database
- Redis cache
- Deployment and operations documentation

## Tech Stack

Backend:

- Go
- Gin
- GORM
- MySQL 8
- Redis 7
- golang-migrate
- go-playground/validator
- zap
- JWT

Admin Web:

- Vue 3
- TypeScript
- Vite
- Element Plus
- Pinia
- Axios
- Vue Router

PC/H5 Web:

- Vue 3
- TypeScript
- Vite
- Pinia
- Axios
- Vue Router

Mini Program:

- uni-app
- Vue 3
- TypeScript

Deployment:

- Docker Compose
- Nginx
- HTTPS
- Linux server

## Important Docs

Before starting any task, read the relevant docs in `/docs`.

Always read:

- docs/00-TECH-STACK-FINAL.md
- docs/01-system-architecture-and-tech-stack.md
- docs/02-GO-BACKEND-ARCHITECTURE.md
- docs/CODEX_TASKS.md
- docs/16-development-milestones.md

For database tasks, also read:

- docs/14-database-migration-ddl.md
- docs/15-error-code-summary.md

For UI tasks, also read:

- docs/13-ui-page-spec.md
- docs/18-design-language.md
- docs/19-core-wireframes.md

For deployment and release tasks, also read:

- docs/20-deployment-guide.md
- docs/21-mini-program-launch-checklist.md
- docs/22-security-and-privacy-checklist.md
- docs/23-backup-and-monitoring-guide.md
- docs/24-production-readiness-checklist.md

## Non-negotiable Business Rules

These rules must never be changed unless the user explicitly updates the product requirement.

1. Do not merge `users` and `family_members`.
2. `users` are platform accounts.
3. `family_members` are genealogy/family-tree nodes.
4. A `family_member` can exist without a user account.
5. A `family_member` can be marked as `NOT_REQUIRED` when binding a real user is impossible or unnecessary.
6. One active `user` can only bind to one active `family_member` within the same family.
7. One active `family_member` can only be bound by one active `user`.
8. The same `user` may appear in different families through different family-member links.
9. A family currently supports only one main family surname.
10. `relationship_type` only supports `PARENT_CHILD` and `SPOUSE`.
11. Do not store `SIBLING` in the database.
12. `ADD_SIBLING` must be implemented through shared parent-child relationships.
13. If a member has no father or mother node, adding a sibling must fail and prompt the user to create a parent node first.
14. A member with descendants cannot be deleted directly.
15. A family founder cannot directly exit or be deleted before transferring founder status.
16. `PLATFORM_ADMIN` cannot approve founder transfer.
17. `PLATFORM_ADMIN` cannot approve family dissolution.
18. `PLATFORM_ADMIN` cannot generate share invitation links.
19. Important operations must write `operation_logs`.
20. Tree structure changes must update `graph_version`.
21. User token and admin token must be separated.
22. Verification codes must not be stored in plain text.
23. Passwords must not be stored in plain text.
24. Tokens, verification codes, `session_key`, `AppSecret`, `openid`, and `unionid` must not be logged in full.

## Go Backend Rules

1. Handler only handles HTTP input/output.
2. Business rules must be implemented in service layer.
3. Database access must be implemented in repository layer.
4. Complex operations must use transactions.
5. Do not hard-code test users in production code.
6. Do not log sensitive data.
7. Do not put complex business logic in Gin handlers.
8. Use GORM transactions for multi-table writes.
9. Use `OperationLogService` for audit logs.
10. Use `GraphVersionService` for tree updates.
11. Do not create a Java, Spring Boot, Node.js, NestJS, Python, FastAPI, or React backend.
12. Do not change the backend stack without explicit user approval.

## Required Backend Layers

Each backend module should follow this structure when applicable:

- handler
- service
- repository
- model
- dto
- vo
- enum

Common services that must be preserved:

- AuthService
- TokenService
- VerificationCodeService
- AccountMergeService
- AccountClaimService
- FamilyPermissionService
- FamilyMemberService
- FamilyRelationshipService
- FamilyTreeService
- FamilyInvitationService
- FamilyJoinRequestService
- PublicDisplayService
- VisitorMessageService
- FamilyRoleService
- FounderTransferService
- FamilyDissolutionService
- OperationLogService
- GraphVersionService
- TransactionManager

## Transaction Rules

The following operations must use database transactions:

- Create family
- Account merge
- Account claim
- Accept invitation
- Approve join request
- Delete member
- Add relationship
- Approve founder transfer
- Approve family dissolution
- Restore family

If any step fails, the whole operation must roll back.

## Graph Version Rules

The following operations must increment `families.graph_version`:

- Create member
- Delete member
- Restore member
- Create relationship
- Delete relationship
- Update member key fields
- Update relationship nature

The following operations must not increment `graph_version`:

- Accept invitation
- Bind user to member
- Change family role
- Public application
- Visitor message
- Founder transfer
- Family dissolution

## Operation Log Rules

Sensitive and important operations must write `operation_logs`.

Must log:

- Admin login success/failure
- Admin lock/unlock
- Create/edit/delete admin user
- Edit/disable/restore user
- Pre-create account
- Account merge
- Account claim
- Create/edit/delete family
- Create/edit/delete member
- Create/delete relationship
- Public application review
- Visitor message review
- Founder transfer review
- Family dissolution review

Must not log:

- Plain password
- Password hash
- Plain verification code
- Access token
- Refresh token
- WeChat session_key
- WeChat AppSecret
- Full openid
- Full unionid
- Database password
- Redis password

## UI Rules

1. Use the Genealogy Calm Design System.
2. Admin pages use Element Plus.
3. Public web pages use a warm, calm, trustworthy style.
4. Mini program pages should be short and task-focused.
5. Do not connect real APIs in prototype tasks unless explicitly requested.
6. Use mock data for UI prototypes.
7. Do not introduce a new UI framework without approval.
8. Do not implement high-fidelity visual effects before the user confirms the low-fidelity prototype.

## Local Development Rules

The project should support macOS local development.

Default local development approach:

- Go backend runs locally.
- Admin web runs locally.
- PC/H5 web runs locally.
- Mini program runs locally through uni-app and WeChat DevTools.
- MySQL 8 and Redis 7 run through Docker Compose.

Required local artifacts:

- docker/docker-compose.yml
- backend/.env.example
- backend/configs/config.example.yaml
- backend/migrations/
- backend/cmd/server/main.go
- GET /api/health

## Before Finishing Each Task

Run if applicable:

- gofmt on changed Go files
- go test ./...
- pnpm build for changed frontend apps
- pnpm typecheck if configured
- pnpm lint if configured

Then report:

1. Completed work
2. Changed files
3. New APIs
4. New database tables / fields
5. Commands run
6. Test results
7. Known limitations
8. Risks
9. Next recommended task

## Forbidden

Do not:

1. Change the tech stack.
2. Generate Java / Spring Boot backend.
3. Generate Node.js / NestJS backend.
4. Generate Python / FastAPI backend.
5. Merge `users` and `family_members`.
6. Store `SIBLING` in relationships.
7. Skip permission checks.
8. Skip `operation_logs`.
9. Skip `graph_version`.
10. Use mock data in production code.
11. Commit `.env` secrets.
12. Log sensitive values.
13. Modify docs unless the user explicitly asks for documentation updates.
14. Implement multiple milestones in one task unless explicitly requested.


## Project Naming

The official project name is:

```text
Tree
```

Naming conventions:

```text
Local directory: ~/Projects/Tree
Backend service: tree-api
Admin web: tree-admin-web
Public web: tree-web
Mini program: tree-miniapp
Database: tree_platform
MySQL container: tree-mysql
Redis container: tree-redis
```

Do not rename domain concepts:

```text
family
families
family_member
family_relationships
family_member_user_links
```

These are business-domain names and must remain unchanged.
