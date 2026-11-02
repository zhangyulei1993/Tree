# 20. 服务器部署指南 v1

> 用于测试环境和生产环境部署。当前后端：Go + Gin + GORM + MySQL 8 + Redis 7。

## 1. 环境划分

至少区分：

```text
staging 测试环境
production 正式环境
```

不要将 production 当测试环境使用。

## 2. 推荐架构

```text
Nginx
  ├── admin-web 静态文件
  ├── web 静态文件
  └── /api 反向代理到 Go 后端

Go 后端服务
MySQL 8
Redis 7
uploads 文件目录或对象存储
logs 日志目录
```

## 3. 推荐目录

```text
/opt/Tree
├── backend
│   ├── tree-api
│   ├── .env
│   └── logs
├── admin-web/dist
├── web/dist
├── uploads
├── docker/docker-compose.yml
├── nginx/Tree.conf
└── backups
```

## 4. 域名规划

| 用途 | 域名示例 |
|---|---|
| PC/H5 前台 | `https://www.example.com` |
| 管理后台 | `https://admin.example.com` |
| 后端 API | `https://api.example.com` |
| 文件访问 | `https://static.example.com` |

小程序正式请求建议使用：

```text
https://api.example.com
```

## 5. HTTPS

生产环境必须启用 HTTPS。

建议：

```text
Nginx 终止 TLS
Go 后端只监听内网端口
```

证书来源：

```text
Let's Encrypt
云服务商免费证书
商业证书
```

## 6. Nginx 示例

```nginx
server {
    listen 80;
    server_name www.example.com admin.example.com api.example.com;
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl http2;
    server_name api.example.com;

    ssl_certificate     /etc/nginx/certs/api.example.com.pem;
    ssl_certificate_key /etc/nginx/certs/api.example.com.key;

    client_max_body_size 20m;

    location /api/ {
        proxy_pass http://127.0.0.1:8080/api/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## 7. 生产 .env 要求

```env
APP_ENV=production
APP_PORT=8080

MYSQL_HOST=127.0.0.1
MYSQL_PORT=3306
MYSQL_USER=tree_user
MYSQL_PASSWORD=strong_password
MYSQL_DATABASE=tree_platform

REDIS_HOST=127.0.0.1
REDIS_PORT=6379
REDIS_PASSWORD=strong_redis_password
REDIS_DB=0

JWT_USER_SECRET=replace_with_strong_random_secret
JWT_ADMIN_SECRET=replace_with_another_strong_random_secret

WECHAT_MINI_APP_ID=
WECHAT_MINI_APP_SECRET=

FILE_STORAGE_TYPE=local
FILE_STORAGE_LOCAL_PATH=/opt/Tree/uploads
LOG_LEVEL=info
```

禁止：

```text
使用示例 JWT_SECRET
使用弱 MySQL 密码
Redis 暴露公网
WECHAT_MINI_APP_SECRET 写入前端
生产日志打印验证码、token、openid、unionid
```

## 8. systemd 示例

```ini
[Unit]
Description=Tree API
After=network.target

[Service]
Type=simple
WorkingDirectory=/opt/Tree/backend
ExecStart=/opt/Tree/backend/tree-api
Restart=always
RestartSec=5
EnvironmentFile=/opt/Tree/backend/.env

[Install]
WantedBy=multi-user.target
```

## 9. 发布顺序

```text
1. 数据库备份
2. 执行 migration
3. 部署 Go 后端
4. 重启 Go 服务
5. 验证 /api/health
6. 部署 admin-web
7. 部署 web
8. 配置 Nginx
9. 验证 HTTPS
10. 验证小程序 request 域名
11. 上传小程序体验版
12. 体验版测试
13. 提交小程序审核
14. 审核通过后发布
```

## 10. 回滚策略

准备：

```text
上一版本后端二进制
上一版本 admin-web dist
上一版本 web dist
数据库备份
Nginx 旧配置
小程序上一稳定版本
```

## 11. 备案提醒

如果服务器部署在中国大陆，通常需要处理：

```text
ICP备案
APP / 小程序备案
域名实名认证
HTTPS
隐私协议
用户协议
```

最终以服务器所在地、主体、分发平台和最新官方规则为准。

## 12. Staging 安全部署模板

仓库提供：

```text
scripts/deploy-staging.sh
scripts/rollback-staging.sh
```

这些脚本是安全模板，不包含服务器地址、数据库地址、账号或密钥，也不能直接用于 production。

部署脚本要求：

```text
APP_ENV 必须严格等于 staging
必须显式设置 STAGING_DEPLOY_CONFIRM=DEPLOY_STAGING
STAGING_ROOT、服务名和健康检查地址不得包含 production/prod 标识
必须提供可执行的备份 hook
必须提供可执行的 forward-only migration hook
发布物必须已在部署前构建并审核
健康检查失败时部署命令返回失败
```

必要变量：

```text
APP_ENV
STAGING_DEPLOY_CONFIRM
STAGING_ROOT
RELEASE_ID
BACKEND_BINARY
ADMIN_DIST
WEB_DIST
STAGING_BACKUP_SCRIPT
STAGING_MIGRATE_SCRIPT
STAGING_SERVICE_NAME
STAGING_HEALTH_URL
```

变量只通过安全的服务器环境注入，不写入仓库。`STAGING_MIGRATE_SCRIPT` 只接受向前执行的 `up` 操作。

## 13. Staging 回滚边界

`rollback-staging.sh` 仅切换到上一应用发布目录并重启服务：

```text
不执行 migration down
不自动恢复数据库
不删除当前或历史数据库数据
不允许 APP_ENV=production
健康检查失败时恢复回滚前的应用版本
```

数据库 schema 发生不兼容变化时，必须使用向后兼容的 migration、分阶段发布或经审核的独立恢复方案，不能依赖 destructive/down migration。

## 14. 发布前本地门禁

在部署 staging 前运行：

```bash
bash scripts/security-check.sh
bash scripts/m16-verify.sh
```

任一命令失败时停止发布，不得跳过或手工改写为 PASS。当前 `admin-web`、`web` 和 `miniapp` 仍包含 mock 原型流程，在完成真实 API 对接和体验环境验收前不得作为生产发布物。
