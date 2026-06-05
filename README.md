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
