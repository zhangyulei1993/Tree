# Codex Execution Document: Mac Docker Local Development Environment for Tree

## 0. Role

You are Codex working inside the `Tree` project repository.

Your task is to make this project runnable on a clean Mac where the host machine only has:

- Git
- Docker Desktop
- VS Code / Cursor
- WeChat Developer Tools

Do **not** require the Mac host to install Go, Node.js, pnpm, npm, MySQL, Redis, Java, Flutter, or other development runtimes.

All backend, frontend, and database runtime dependencies must be controlled through Docker.

---

## 1. Project Background

The project is named `Tree` / `家脉亲缘`.

Expected project structure:

```text
Tree/
├─ backend/        Go + Gin backend
├─ admin/          Vue3 + TypeScript + Vite admin panel
├─ web/            Vue3 + TypeScript + Vite public website / H5
├─ miniprogram/    Native WeChat Mini Program
├─ docs/           Project documentation
└─ README.md
```

Expected technology stack:

```text
backend: Go, Gin, GORM, MySQL, JWT
admin: Vue3, TypeScript, Vite, Element Plus
web: Vue3, TypeScript, Vite
miniprogram: Native WeChat Mini Program
database: MySQL
```

The Docker environment should cover:

```text
1. MySQL
2. backend
3. admin
4. web
```

The WeChat Mini Program should remain opened by WeChat Developer Tools on the Mac host. Do not attempt to Dockerize the GUI tool.

---

## 2. Core Requirements

### 2.1 Do Not Pollute the Mac Host

Do not add instructions requiring host-level installation of:

```text
node
npm
pnpm
yarn
go
mysql
redis
java
flutter
dart
```

The host should only run Docker and Git.

### 2.2 Do Not Rewrite Business Logic

This task is only about local development environment setup.

Do **not** rewrite:

```text
backend business code
admin pages
web pages
miniprogram code
kinship algorithm
database schema
```

Only add or minimally adjust environment-related files.

### 2.3 Avoid Destructive Operations

Do **not** run commands that delete user files or global Mac environment, such as:

```bash
rm -rf ~
rm -rf ~/.ssh
rm -rf /opt/homebrew
rm -rf /usr/local
docker system prune -a --volumes
```

Inside the project repository, only modify files necessary for Docker development setup.

---

## 3. Files to Create or Update

Create or update these files:

```text
Tree/
├─ docker-compose.yml
├─ .env.example
├─ .gitignore
├─ README.md                    # add Docker development section only
├─ backend/
│  └─ Dockerfile.dev
├─ admin/
│  └─ Dockerfile.dev
└─ web/
   └─ Dockerfile.dev
```

Optional if needed:

```text
Tree/
└─ database/
   └─ init.sql
```

Only create `database/init.sql` if the project already has an SQL initialization script or if the existing backend clearly expects one.

---

## 4. Before Editing: Inspect the Repository

First inspect the project:

```bash
pwd
ls -la
find . -maxdepth 2 -type f | sort
```

Check backend:

```bash
ls -la backend
cat backend/go.mod
find backend -maxdepth 3 -type f | sort
```

Find the backend entrypoint. Prefer one of:

```text
backend/cmd/server/main.go
backend/main.go
backend/cmd/main.go
```

Check admin package manager:

```bash
ls -la admin
cat admin/package.json
ls admin | grep -E "pnpm-lock.yaml|package-lock.json|yarn.lock"
```

Check web package manager:

```bash
ls -la web
cat web/package.json
ls web | grep -E "pnpm-lock.yaml|package-lock.json|yarn.lock"
```

Then adapt Docker commands to the actual project files.

---

## 5. Root `.env.example`

Create `Tree/.env.example`:

```env
# MySQL
MYSQL_ROOT_PASSWORD=root123456
MYSQL_DATABASE=tree
MYSQL_USER=tree_user
MYSQL_PASSWORD=tree_pass

# Ports
BACKEND_PORT=8080
WEB_PORT=5173
ADMIN_PORT=5174

# Backend
JWT_SECRET=tree-dev-secret

# Timezone
TZ=Asia/Shanghai
```

Do not commit a real `.env` file.

If `.env` does not exist, tell the user to copy it manually:

```bash
cp .env.example .env
```

---

## 6. Root `.gitignore`

Ensure `.gitignore` contains at least:

```gitignore
# Environment
.env
.env.local
.env.*.local

# Dependencies
node_modules/
.pnpm-store/

# Build outputs
dist/
build/
.out/

# Logs
*.log
npm-debug.log*
pnpm-debug.log*
yarn-debug.log*

# Temporary files
*.tmp
*.bak
*.bak_*
.trash_*/
.DS_Store

# IDE
.vscode/settings.json
.idea/

# Docker local data
docker-data/
```

Do not remove existing important `.gitignore` rules.

---

## 7. Backend Dockerfile

Create `backend/Dockerfile.dev`.

First read `backend/go.mod` and match the Go version.

If `go.mod` says `go 1.22`, use:

```dockerfile
FROM golang:1.22-alpine
```

If `go.mod` says `go 1.23`, use:

```dockerfile
FROM golang:1.23-alpine
```

If unsure, use the latest compatible stable Go Alpine image.

Template:

```dockerfile
FROM golang:1.23-alpine

WORKDIR /app

RUN apk add --no-cache git bash tzdata

ENV CGO_ENABLED=0
ENV GOPROXY=https://goproxy.cn,direct

EXPOSE 8080
```

Do not copy source into the image for development; source will be mounted by Docker Compose.

---

## 8. Admin Dockerfile

Create `admin/Dockerfile.dev`:

```dockerfile
FROM node:22-alpine

WORKDIR /app

RUN apk add --no-cache bash git

EXPOSE 5173
```

---

## 9. Web Dockerfile

Create `web/Dockerfile.dev`:

```dockerfile
FROM node:22-alpine

WORKDIR /app

RUN apk add --no-cache bash git

EXPOSE 5173
```

---

## 10. Docker Compose

Create `Tree/docker-compose.yml`.

Use this base version, but adapt backend command if the entrypoint differs.

```yaml
services:
  mysql:
    image: mysql:8.0
    container_name: tree_mysql
    restart: unless-stopped
    environment:
      MYSQL_ROOT_PASSWORD: ${MYSQL_ROOT_PASSWORD:-root123456}
      MYSQL_DATABASE: ${MYSQL_DATABASE:-tree}
      MYSQL_USER: ${MYSQL_USER:-tree_user}
      MYSQL_PASSWORD: ${MYSQL_PASSWORD:-tree_pass}
      TZ: ${TZ:-Asia/Shanghai}
    ports:
      - "3306:3306"
    volumes:
      - tree_mysql_data:/var/lib/mysql
    command:
      [
        "--character-set-server=utf8mb4",
        "--collation-server=utf8mb4_unicode_ci"
      ]
    healthcheck:
      test: ["CMD-SHELL", "mysqladmin ping -h localhost -p$${MYSQL_ROOT_PASSWORD} || exit 1"]
      interval: 5s
      timeout: 3s
      retries: 20

  backend:
    build:
      context: ./backend
      dockerfile: Dockerfile.dev
    container_name: tree_backend
    restart: unless-stopped
    working_dir: /app
    volumes:
      - ./backend:/app
      - tree_go_mod_cache:/go/pkg/mod
      - tree_go_build_cache:/root/.cache/go-build
    ports:
      - "${BACKEND_PORT:-8080}:8080"
    environment:
      TZ: ${TZ:-Asia/Shanghai}
      DB_HOST: mysql
      DB_PORT: 3306
      DB_NAME: ${MYSQL_DATABASE:-tree}
      DB_USER: ${MYSQL_USER:-tree_user}
      DB_PASSWORD: ${MYSQL_PASSWORD:-tree_pass}
      JWT_SECRET: ${JWT_SECRET:-tree-dev-secret}
    depends_on:
      mysql:
        condition: service_healthy
    command: sh -c "go mod download && go run ./cmd/server"

  admin:
    build:
      context: ./admin
      dockerfile: Dockerfile.dev
    container_name: tree_admin
    restart: unless-stopped
    working_dir: /app
    volumes:
      - ./admin:/app
      - tree_admin_node_modules:/app/node_modules
    ports:
      - "${ADMIN_PORT:-5174}:5173"
    environment:
      TZ: ${TZ:-Asia/Shanghai}
      VITE_API_BASE_URL: http://localhost:${BACKEND_PORT:-8080}
    command: >
      sh -c "
      if [ -f pnpm-lock.yaml ]; then
        corepack enable && pnpm install && pnpm dev --host 0.0.0.0;
      elif [ -f package-lock.json ]; then
        npm install && npm run dev -- --host 0.0.0.0;
      elif [ -f yarn.lock ]; then
        corepack enable && yarn install && yarn dev --host 0.0.0.0;
      else
        corepack enable && pnpm install && pnpm dev --host 0.0.0.0;
      fi
      "

  web:
    build:
      context: ./web
      dockerfile: Dockerfile.dev
    container_name: tree_web
    restart: unless-stopped
    working_dir: /app
    volumes:
      - ./web:/app
      - tree_web_node_modules:/app/node_modules
    ports:
      - "${WEB_PORT:-5173}:5173"
    environment:
      TZ: ${TZ:-Asia/Shanghai}
      VITE_API_BASE_URL: http://localhost:${BACKEND_PORT:-8080}
    command: >
      sh -c "
      if [ -f pnpm-lock.yaml ]; then
        corepack enable && pnpm install && pnpm dev --host 0.0.0.0;
      elif [ -f package-lock.json ]; then
        npm install && npm run dev -- --host 0.0.0.0;
      elif [ -f yarn.lock ]; then
        corepack enable && yarn install && yarn dev --host 0.0.0.0;
      else
        corepack enable && pnpm install && pnpm dev --host 0.0.0.0;
      fi
      "

volumes:
  tree_mysql_data:
  tree_go_mod_cache:
  tree_go_build_cache:
  tree_admin_node_modules:
  tree_web_node_modules:
```

Important:

- If backend entrypoint is not `./cmd/server`, change only this part:

```yaml
command: sh -c "go mod download && go run ./cmd/server"
```

Examples:

```yaml
command: sh -c "go mod download && go run ."
```

or:

```yaml
command: sh -c "go mod download && go run ./main.go"
```

---

## 11. Backend Database Configuration

Inspect backend config loading.

Search:

```bash
grep -R "DB_HOST\|DB_PORT\|DB_NAME\|DB_USER\|DB_PASSWORD\|MYSQL" -n backend || true
```

If backend already reads environment variables, do not change code.

If backend only has hardcoded local MySQL config like:

```text
127.0.0.1
localhost
root
password
```

make the smallest necessary change so it can read these environment variables:

```text
DB_HOST=mysql
DB_PORT=3306
DB_NAME=tree
DB_USER=tree_user
DB_PASSWORD=tree_pass
```

Do not introduce a large configuration framework.

---

## 12. Vite Proxy / API Address

Inspect:

```bash
cat admin/vite.config.*
cat web/vite.config.*
grep -R "VITE_API_BASE_URL\|localhost:8080\|proxy" -n admin web || true
```

If Vite proxy already exists and points to backend, keep it.

If frontend directly reads `VITE_API_BASE_URL`, ensure Docker Compose passes:

```yaml
VITE_API_BASE_URL: http://localhost:${BACKEND_PORT:-8080}
```

For browser requests, `localhost:8080` is correct because the browser runs on the Mac host.

For container-to-container requests, use:

```text
http://backend:8080
```

Only use `backend:8080` inside containers, not in browser-side code.

---

## 13. README Update

Append a section to `README.md`:

```markdown
## Local Development with Docker on macOS

This project is designed to run on a clean Mac host with Docker Desktop.

The Mac host does not need local Go, Node.js, pnpm, npm, MySQL, or Redis.

### Required host tools

- Git
- Docker Desktop
- VS Code / Cursor
- WeChat Developer Tools for `miniprogram`

### Start

```bash
cp .env.example .env
docker compose up --build
```

### Visit

```text
web: http://localhost:5173
admin: http://localhost:5174
backend: http://localhost:8080
mysql: localhost:3306
```

### Common commands

```bash
docker compose up -d
docker compose logs -f
docker compose logs -f backend
docker compose logs -f web
docker compose logs -f admin
docker compose exec backend sh
docker compose exec mysql mysql -utree_user -ptree_pass tree
docker compose down
```

### Reset database

Warning: this deletes local database data.

```bash
docker compose down -v
docker compose up --build
```

### WeChat Mini Program

Open the `miniprogram/` directory directly with WeChat Developer Tools.

For simulator requests, use:

```text
http://localhost:8080
```

For real-device preview, use the Mac LAN IP, for example:

```bash
ipconfig getifaddr en0
```

Then configure the API base URL as:

```text
http://<Mac LAN IP>:8080
```
```

---

## 14. Validation Commands

After changes, run:

```bash
cp .env.example .env 2>/dev/null || true
docker compose config
docker compose up --build
```

In another terminal:

```bash
docker compose ps
docker compose logs --tail=100 backend
docker compose logs --tail=100 web
docker compose logs --tail=100 admin
```

Check inside containers:

```bash
docker compose exec backend go version
docker compose exec web node -v
docker compose exec admin node -v
docker compose exec mysql mysql --version
```

Expected browser URLs:

```text
http://localhost:5173
http://localhost:5174
http://localhost:8080
```

If backend has no root route, `http://localhost:8080` may return 404. That is acceptable as long as the backend container is running and API routes work.

---

## 15. Acceptance Criteria

The task is complete when:

```text
1. docker compose config passes.
2. docker compose up --build starts mysql, backend, admin, and web.
3. The backend container connects to MySQL using Docker service name `mysql`.
4. The web dev server is accessible at http://localhost:5173.
5. The admin dev server is accessible at http://localhost:5174.
6. No local host Go/Node/MySQL installation is required.
7. `.env` is ignored by Git.
8. `.env.example` is committed.
9. Dockerfiles are committed.
10. README contains clear Docker local development instructions.
```

---

## 16. Final Response Format

When done, summarize:

```text
Modified files:
- docker-compose.yml
- .env.example
- .gitignore
- backend/Dockerfile.dev
- admin/Dockerfile.dev
- web/Dockerfile.dev
- README.md

How to run:
1. cp .env.example .env
2. docker compose up --build

URLs:
- web: http://localhost:5173
- admin: http://localhost:5174
- backend: http://localhost:8080

Notes:
- Host machine does not need local Node.js, Go, or MySQL.
- WeChat Mini Program still uses WeChat Developer Tools on Mac.
```

If something cannot be completed, clearly state:

```text
Could not complete:
- ...
Reason:
- ...
Manual action needed:
- ...
```
