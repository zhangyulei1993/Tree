# 29-CODEX-PR-REVIEW-PROMPTS.md

# Codex PR Review 提示词 v1

> 每个阶段完成后，可让 Codex 再做一次代码审查。  
> 以下提示词可直接复制到 PR 或 Codex 对话中。

---

# 通用 PR Review 提示词

请审查当前 PR，重点检查：

```text
1. 是否违反 AGENTS.md。
2. 是否违反 docs 中的业务规则。
3. 是否修改了不应修改的目录。
4. 是否混用了 user token 和 admin token。
5. 是否把验证码、密码、token、session_key、AppSecret、openid、unionid 写入日志。
6. 多表写入是否使用事务。
7. 是否遗漏 operation_logs。
8. 是否遗漏 graph_version。
9. 是否把业务逻辑写进 handler。
10. 是否提前实现了下一阶段内容。
11. 是否存在 mock data 进入生产代码。
12. 是否有测试覆盖关键路径。
```

请输出：

```text
1. 阻塞问题
2. 高风险问题
3. 中风险问题
4. 建议优化
5. 是否建议合并
```

---

# 数据库 PR Review

请重点检查：

```text
1. users 和 family_members 是否分离。
2. family_member_user_links 是否能表达 user-member 绑定。
3. relationship_type 是否只允许 PARENT_CHILD / SPOUSE。
4. 是否错误加入 SIBLING。
5. operation_logs 是否包含必要字段。
6. 软删除字段是否完整。
7. created_at / updated_at 是否完整。
8. 索引是否覆盖主要查询。
9. 复杂唯一约束是否在 Service TODO 中标出。
10. ROOT_ADMIN seed 是否安全。
```

---

# 认证模块 PR Review

请重点检查：

```text
1. 密码是否 hash。
2. 验证码是否 hash。
3. 验证码是否 5 分钟有效。
4. token 是否可失效。
5. user token 和 admin token 是否分离。
6. DISABLED / CANCELLED / MERGED 用户是否不能登录。
7. 账号合并后旧 token 是否失效。
8. 微信 AppSecret 是否只在后端。
9. session_key / openid / unionid 是否未完整写日志。
```

---

# 家庭成员与关系 PR Review

请重点检查：

```text
1. 删除存在下级成员是否被阻止。
2. FOUNDER 是否不能直接删除。
3. NOT_REQUIRED 是否不能绑定用户。
4. ADD_SIBLING 是否不落库。
5. ADD_SIBLING 是否通过共同父母实现。
6. 无父母时 ADD_SIBLING 是否失败。
7. PRIMARY 父亲 / 母亲唯一性是否校验。
8. 多配偶是否允许。
9. 关系变化是否更新 graph_version。
10. 操作是否写 operation_logs。
```

---

# 权限 PR Review

请重点检查：

```text
1. PLATFORM_ADMIN 是否不能审核创始人转让。
2. PLATFORM_ADMIN 是否不能审核家庭解散。
3. PLATFORM_ADMIN 是否不能生成分享邀请链接。
4. 普通 MEMBER 是否不能管理其他成员。
5. 游客是否不能查看私有家庭树。
6. 未绑定手机号是否不能加入家庭、留言、接受邀请。
7. 后台管理员是否不能管理同级管理员。
8. ROOT_ADMIN 是否不能被删除、禁用、降级。
```

---

# UI PR Review

请重点检查：

```text
1. 是否符合 Genealogy Calm Design System。
2. 是否使用设计令牌。
3. 是否引入了未经确认的新 UI 框架。
4. 后台是否使用 Element Plus。
5. PLATFORM_ADMIN 是否隐藏高风险审核入口。
6. 危险操作是否有确认弹窗。
7. 是否有 loading / empty / error / no-permission 状态。
8. 原型任务是否没有接真实 API。
9. mock data 是否没有进入生产接口逻辑。
```

---

# 部署 PR Review

请重点检查：

```text
1. 是否泄露 .env 或生产密钥。
2. Docker Compose 是否持久化 MySQL 数据。
3. Redis 是否未暴露公网。
4. Nginx 是否正确配置 HTTPS。
5. Swagger 生产环境是否受限或关闭。
6. 日志目录是否持久化。
7. 上传目录是否持久化。
8. 是否有数据库备份说明。
9. 小程序合法域名是否说明清楚。
```
