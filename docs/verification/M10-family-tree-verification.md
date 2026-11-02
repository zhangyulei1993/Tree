# M10 Family Tree Verification

## Status

- Stage: M10 - Family Tree Assembly and Graph-Version Cache
- Automated verification: PASS
- Manual API verification: PENDING
- Access token: omitted; a real token must never be written to this file

## Environment

- Base URL: `http://127.0.0.1:8080`
- Runtime: `tree-dev`
- MySQL service: `tree-mysql`
- Redis service: `tree-redis`
- Database: `tree_platform`
- Private endpoint token type: USER

## Codex Self-Test Items

| Check | Status |
|---|---|
| nodes / edges / tree basic assembly | PASS |
| Root identification | PASS |
| Parent-child relationships | PASS |
| Spouse and multiple-spouse relationships | PASS |
| Multiple parents | PASS |
| Independent nodes | PASS |
| Cycle-safe traversal | PASS |
| Soft-deleted members and relationships filtered | PASS |
| SIBLING edges filtered | PASS |
| Private permission denial | PASS |
| Public NORMAL and APPROVED checks | PASS |
| Cache miss and cache write | PASS |
| Cache hit skips snapshot query | PASS |
| Redis failure falls back to database | PASS |
| graph_version produces a new cache key | PASS |
| Serialized response excludes sensitive field names | PASS |
| `go test ./...` | PASS |

## Pending Manual Verification

| Check | Status |
|---|---|
| Private family tree with a family-member USER token | PENDING |
| Private family tree denied for a non-member | PENDING |
| Public tree for NORMAL and APPROVED family | PENDING |
| Public tree denied for non-public family | PENDING |
| Redis cache key created after cache miss | PENDING |
| Second request returns the same graph version and payload | PENDING |
| New graph version uses a new cache key | PENDING |

## Curl Verification

Use shell variables only. Do not write the complete token to this file:

```bash
export BASE_URL=http://127.0.0.1:8080
export TOKEN='<USER_ACCESS_TOKEN_OMITTED>'
export FAMILY_ID='<FAMILY_ID>'

curl -i "$BASE_URL/api/families/$FAMILY_ID/tree" \
  -H "Authorization: Bearer $TOKEN"

curl -i "$BASE_URL/api/public/families/$FAMILY_ID/tree"
```

Verify that:

- `treeMode` is `LIST_TREE`.
- `nodes`, `edges`, and `tree` are arrays.
- `edges[*].relationshipType` is only `PARENT_CHILD` or `SPOUSE`.
- Node JSON contains no account identifiers or authentication secrets.

## Database And Redis Checks

```bash
docker exec tree-mysql mysql -utree_user -ptree_pass tree_platform -e \
  "SELECT id,status,public_display_status,tree_mode,graph_version FROM families WHERE id=$FAMILY_ID;"

docker exec tree-redis redis-cli --scan \
  --pattern "family:$FAMILY_ID:tree:v*"

docker exec tree-redis redis-cli GET \
  "family:$FAMILY_ID:tree:v<GRAPH_VERSION>"
```

To verify version switching, perform an already supported M8/M9 tree mutation in a disposable test family, confirm `families.graph_version` increases, then request the tree again and check that a new versioned key exists.

## Notes

- M10 reads but never increments `graph_version`.
- Manual verification must not be marked PASS until the HTTP, database, and Redis checks are actually executed.
- No complete access token, password, verification code, openid, unionid, or session key is recorded here.
