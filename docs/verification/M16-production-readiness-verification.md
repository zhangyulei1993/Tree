# M16 Production Readiness Verification

## Scope

This first M16 implementation round covers:

- repository security checks;
- repeatable local verification commands;
- guarded staging deployment and application rollback templates;
- production and mini program release checklists.

It does not add P0 backend tests, deploy to a real server, run production
migrations, or mark the product production-ready.

## Codex Self-Test

| Item | Status | Evidence |
|---|---|---|
| `scripts/security-check.sh` | PASS | Completed with `Security check result: PASS` |
| `scripts/m16-verify.sh` | PASS | Completed with `M16 verification result: PASS` |
| staging deploy refuses non-staging environment | PASS | Missing staging context and production-like targets were rejected |
| rollback contains no destructive/down migration | PASS | Shell syntax and static command scan passed |
| backend tests | PASS | `go test ./...` passed in `tree-dev` |
| admin web build | PASS | `pnpm build` passed in `tree-dev` |
| PC/H5 web build | PASS | `pnpm build` passed in `tree-dev` |
| mini program build | PASS | `pnpm build:mp-weixin` passed in `tree-dev` |

Any failed command keeps the M16 result at FAIL.

## Pending P0 Work

- Add automated coverage for authentication, JWT/token separation, admin
  lockout, family creation, operation logs, and API-level authorization.
- Run isolated integration tests against a database whose name ends in
  `_test`; never run destructive tests against development, staging, or
  production databases.
- Replace all mock frontend flows with reviewed real API integration before
  production release.
- Complete a real staging deployment, backup, restore drill, HTTPS check,
  observability check, and mini program experience-version test.

## Staging Verification

Before using the templates:

1. Build reviewed backend and frontend artifacts.
2. Provide an executable backup hook and a forward-only migration hook.
3. Export only staging values through the deployment environment.
4. Keep credentials outside the repository and command history.
5. Verify the health endpoint after deployment.
6. Test `rollback-staging.sh` using application releases with compatible
   schemas. It never runs migration down.

Required explicit confirmations:

```text
STAGING_DEPLOY_CONFIRM=DEPLOY_STAGING
STAGING_ROLLBACK_CONFIRM=ROLLBACK_STAGING
```

## Production Gate

Production remains blocked until all P0 items in
`docs/24-production-readiness-checklist.md` have evidence and are marked PASS.
Build success alone is not production approval.

## Sensitive Data Rules

Do not place the following in this verification document:

- access or refresh tokens;
- passwords or password hashes;
- verification codes;
- full phone numbers;
- openid or unionid;
- WeChat AppSecret;
- database or Redis credentials.

## Final Result

```text
M16 first-round verification: PASS
Production readiness: NOT APPROVED
```
