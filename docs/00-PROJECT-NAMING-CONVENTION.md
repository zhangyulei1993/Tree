# 00-PROJECT-NAMING-CONVENTION.md

# Tree 项目命名规范 v1

> 本文档用于统一项目名称、仓库名、目录名、服务名、数据库名和部署名。  
> 后续 Codex、开发团队和部署文档必须优先遵守本命名规范。

---

# 1. 项目名称

```text
Tree
```

说明：

```text
Tree 是项目正式名称。
业务领域仍然是家族 / 家庭关系 / 亲缘树平台。
不要把业务术语 family、family_member、families 等数据库概念替换掉。
```

---

# 2. 仓库与本地目录

推荐仓库名：

```text
Tree
```

推荐本地目录：

```text
~/Projects/Tree
```

项目根目录：

```text
Tree
├── AGENTS.md
├── README.md
├── docs
├── backend
├── admin-web
├── web
├── miniapp
├── docker
└── scripts
```

---

# 3. 子项目名称

| 模块 | 名称 |
|---|---|
| 后端 API | `tree-api` |
| 管理后台 | `tree-admin-web` |
| PC/H5 前台 | `tree-web` |
| 微信小程序 | `tree-miniapp` |
| MySQL 容器 | `tree-mysql` |
| Redis 容器 | `tree-redis` |

---

# 4. 数据库与账号

推荐数据库名：

```text
tree_platform
```

本地开发数据库用户：

```text
tree_user
```

本地开发数据库密码示例：

```text
tree_pass
```

注意：

```text
生产环境必须使用强密码，不得使用示例密码。
```

---

# 5. 后端服务

后端二进制推荐名称：

```text
tree-api
```

systemd 服务名：

```text
tree-api.service
```

Docker service 名：

```text
tree-api
```

---

# 6. API 与业务命名

API 路由仍然使用业务资源名：

```text
/api/families
/api/families/{familyId}/members
/api/families/{familyId}/relationships
/api/admin/users
```

不要改成：

```text
/api/trees
```

原因：

```text
当前业务核心对象仍是 family / family_member / relationship。
Tree 是产品名，不是所有业务表和 API 资源名。
```

---

# 7. 数据库表名

数据库表名保持不变：

```text
users
families
family_members
family_relationships
family_member_user_links
family_invitations
family_join_requests
family_public_applications
visitor_messages
admin_users
operation_logs
```

不要因为项目名 Tree 而把表名改成：

```text
tree_members
tree_relationships
```

---

# 8. 文档命名

后续文档包建议命名为：

```text
tree-docs-codex-ready-v11-project-name.zip
```

后续版本：

```text
tree-docs-codex-ready-v12-xxx.zip
tree-docs-codex-ready-v13-xxx.zip
```

---

# 9. Codex 注意事项

Codex 必须遵守：

```text
1. 项目名使用 Tree。
2. 本地目录使用 ~/Projects/Tree。
3. 后端服务使用 tree-api。
4. 数据库使用 tree_platform。
5. Docker 容器使用 tree-mysql / tree-redis。
6. 不要把业务概念 family / family_member 错误替换成 tree。
7. 不要修改已确认的数据库业务表名。
```
