# 30-PROJECT-REPO-SETUP-GUIDE.md

# 项目仓库初始化指南 v1

> 本文件用于把文档包转成实际开发仓库结构。

---

# 1. 创建仓库

建议仓库名：

```text
Tree
```

初始化：

```bash
mkdir Tree
cd Tree
git init
```

---

# 2. 推荐目录

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

# 3. 放置文档

将 v10 文档包中的项目文档放到：

```text
docs/
```

将 `AGENTS.md` 放到：

```text
项目根目录 AGENTS.md
```

不要只放在 docs 内。

---

# 4. 第一阶段前目录

M1 开始前可以只有：

```text
Tree
├── AGENTS.md
├── README.md
└── docs
```

然后让 Codex 根据 `CODEX_M1_START_PROMPT.md` 创建：

```text
backend
docker
scripts
.gitignore
```

---

# 5. Git 忽略规则

根目录 `.gitignore` 至少包含：

```gitignore
.env
.env.*
!.env.example
docker/mysql/data
docker/redis/data
uploads
logs
node_modules
dist
.DS_Store
```

---

# 6. 分支策略

建议：

```text
main            稳定分支
develop         开发集成分支
m1-*            M1 阶段分支
m2-*            M2 阶段分支
feature/*       功能分支
fix/*           修复分支
```

---

# 7. 第一次提交

建议第一提交只包含文档：

```bash
git add AGENTS.md docs README.md
git commit -m "docs: add project specification and codex guide"
```

第二提交再让 Codex 做 M1：

```bash
git checkout -b m1-project-skeleton
```

---

# 8. Codex 工作流

```text
1. 新建阶段分支。
2. 给 Codex 单阶段 prompt。
3. Codex 修改代码。
4. 运行测试。
5. 人工检查 diff。
6. 用 PR Review prompt 审查。
7. 通过后合并。
8. 进入下一阶段。
```

---

# 9. 不建议

不要：

```text
1. 直接让 Codex 在 main 上做全部功能。
2. 一个 PR 包含多个里程碑。
3. 没跑测试就合并。
4. 没看 diff 就合并。
5. 没有 AGENTS.md 就启动 Codex。
```
