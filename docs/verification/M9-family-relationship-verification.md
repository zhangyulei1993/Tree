# M9 Family Relationship Verification

## Status

- Stage: M9 - Family Relationship
- Verification status: PASS
- Verification date: 2026-06-07
- Branch: `m9-family-relation`
- Access token: omitted; a real token must never be written to this file

Manual API, database, and automated checks were executed successfully.

## Environment

- Base URL: `http://127.0.0.1:8080`
- Runtime: `tree-dev`
- MySQL service: `tree-mysql`
- Database: `tree_platform`
- Token type: USER
- Required role: `FOUNDER` or `FAMILY_ADMIN`
- Test user ID: `8`
- Test family ID: `2`
- Test role: `FOUNDER`
- Base member ID: `6`

The access token and login password were used only in the verification shell
session. They were not written to this document or any project file.

## Automated Checks

| Check | Status |
|---|---|
| `gofmt -w .` | PASS |
| `go test ./...` | PASS |
| `go vet ./...` | PASS |
| OpenAPI syntax and references | PASS |

Commands:

```bash
docker exec tree-dev bash -lc "cd /workspace/backend && gofmt -w ."
docker exec tree-dev bash -lc "cd /workspace/backend && go test ./..."
docker exec tree-dev bash -lc "cd /workspace/backend && go vet ./..."
```

Automated checks passed on 2026-06-07.

## Manual API Results

The API verification used shell variables without writing the token to a file:

```bash
export BASE_URL=http://127.0.0.1:8080
export TOKEN='<USER_ACCESS_TOKEN_OMITTED>'
export FAMILY_ID=2
export BASE_MEMBER_ID=6
```

| Verification | Result | Evidence |
|---|---|---|
| Backend startup and M9 route registration | PASS | POST, PUT, DELETE routes registered on port 8080 |
| `GET /api/health` | PASS | Returned code `0` and status `ok` |
| USER login | PASS | Token obtained; token and password omitted |
| `ADD_CHILD` | PASS | Member `7`; relationship `1`, direction `6 -> 7`, type `PARENT_CHILD` |
| `ADD_FATHER` | PASS | Member `8`; relationship `2`, direction `8 -> 6`, gender `MALE` |
| Duplicate PRIMARY father | PASS | Returned `43304` |
| `ADD_MOTHER` | PASS | Member `9`; relationship `3`, direction `9 -> 6`, gender `FEMALE` |
| Duplicate PRIMARY mother | PASS | Returned `43305` |
| Two `ADD_SPOUSE` calls | PASS | Members `10`, `11`; relationships `4`, `5` |
| `ADD_SIBLING` without parents | PASS | Returned `43301`; member count remained unchanged |
| `ADD_SIBLING` with parents | PASS | Member `12`; relationships `6`, `7` share parents `8`, `9` |
| PUT relationship note | PASS | Relationship `1` note updated without changing member IDs/type |
| PUT structural field | PASS | Attempt to change `relationshipType` returned `43306` |
| DELETE relationship | PASS | Relationship `4` soft deleted with status `DELETED` |
| Permission test | PASS | ACTIVE non-member user `14` received `43307` |
| Failed-operation rollback | PASS | No duplicate/invalid/unauthorized member rows remained |
| Graph version | PASS | Family `2` changed from `6` to `14` for eight successful mutations |
| Operation logs | PASS | Six CREATE, one UPDATE, and one DELETE success record |
| Stored relationship types | PASS | Only `PARENT_CHILD` and `SPOUSE`; zero `SIBLING` rows |
| M8 member API regression | PASS | Member list and member `6` detail returned code `0` |

## Database Checks

```sql
SELECT id, family_id, from_member_id, to_member_id,
       relationship_type, parent_link_type, status, deleted_at
FROM family_relationships
WHERE family_id = 2
ORDER BY id DESC;

SELECT DISTINCT relationship_type
FROM family_relationships;

SELECT id, graph_version
FROM families
WHERE id = 2;

SELECT action, result, target_id, family_id, created_at
FROM operation_logs
WHERE module = 'FAMILY_RELATIONSHIP'
ORDER BY id DESC;
```

Database results:

- Active relationship rows: `6`
- Soft-deleted relationship: `4`
- Sibling member `12` has two active shared-parent relationships.
- `relationship_type = SIBLING` row count: `0`
- FAMILY_RELATIONSHIP success operation-log count: `8`

## Final Conclusion

M9 automated tests, HTTP integration tests, transaction rollback checks,
database assertions, permission checks, graph-version checks, operation logs,
and the M8 member-query regression all passed.
