# 24. 生产上线准备检查清单 v1

> 所有 P0 项必须完成后才能上线。

## 1. 代码与分支

| 检查项 | 级别 |
|---|---|
| main / release 分支代码已冻结 | P0 |
| 所有 P0 测试通过 | P0 |
| go test ./... 通过 | P0 |
| 前端 build 成功 | P0 |
| 小程序 build 成功 | P0 |
| 无调试按钮 / mock 数据 | P0 |
| 无测试账号密码硬编码 | P0 |

## 2. 服务器与域名

| 检查项 | 级别 |
|---|---|
| 服务器已准备 | P0 |
| 域名已解析 | P0 |
| HTTPS 证书有效 | P0 |
| Nginx 配置正确 | P0 |
| API 域名可访问 | P0 |
| admin 域名可访问 | P0 |
| web 域名可访问 | P0 |
| 证书自动续期或到期提醒 | P1 |

## 3. 备案与小程序

| 检查项 | 级别 |
|---|---|
| 域名实名信息正确 | P0 |
| ICP / APP / 小程序备案按实际要求处理 | P0 |
| 小程序 request 合法域名已配置 | P0 |
| 小程序隐私保护指引已配置 | P0 |
| 用户协议可访问 | P0 |
| 隐私政策可访问 | P0 |
| 注销账号入口可访问 | P0 |
| 小程序体验版测试通过 | P0 |
| 小程序审核说明已准备 | P1 |

## 4. 数据库

| 检查项 | 级别 |
|---|---|
| MySQL 8 正常运行 | P0 |
| migration 已执行 | P0 |
| ROOT_ADMIN 已创建 | P0 |
| 默认 admin 密码已修改 | P0 |
| 数据库生产密码强度足够 | P0 |
| 数据库不公网暴露 | P0 |
| 发布前已备份数据库 | P0 |
| 备份恢复已演练 | P1 |

## 5. Redis

| 检查项 | 级别 |
|---|---|
| Redis 正常运行 | P0 |
| Redis 不公网暴露 | P0 |
| Redis 密码已设置，若使用公网或共享环境 | P0 |
| 验证码缓存正常 | P0 |
| token 黑名单正常 | P0 |
| 游客留言限流正常 | P1 |

## 6. 后端配置

| 检查项 | 级别 |
|---|---|
| APP_ENV=production | P0 |
| JWT_USER_SECRET 已替换 | P0 |
| JWT_ADMIN_SECRET 已替换 | P0 |
| WECHAT_MINI_APP_SECRET 只在后端 | P0 |
| 日志级别不是 debug | P0 |
| Swagger 生产环境受限或关闭 | P1 |
| CORS 白名单正确 | P0 |
| 文件上传路径正确 | P1 |

## 7. 核心业务流程

必须测试：

```text
后台登录
普通用户注册 / 登录
微信小程序登录
小程序绑定手机号
后台预创建账号
账号认领
账号合并
创建家庭
新增成员
添加父亲 / 母亲 / 子女 / 兄弟姐妹 / 配偶
删除成员
邀请绑定成员
接受邀请
申请加入家庭
审核加入申请
申请公开家庭
审核公开申请
游客留言
审核留言
创始人转让
家庭解散
操作日志查询
```

## 8. 权限测试

P0：

```text
普通用户不能访问后台接口
后台 token 不能当普通用户 token 使用
PLATFORM_ADMIN 不能审核创始人转让
PLATFORM_ADMIN 不能审核家庭解散
PLATFORM_ADMIN 不能生成分享邀请链接
普通成员不能删除其他成员
游客不能查看私有家庭树
未绑定手机号不能加入家庭
未绑定手机号不能留言
删除存在下级成员必须失败
```

## 9. 安全检查

P0：

```text
密码不明文存储
验证码不明文存储
日志不记录 token
日志不记录验证码
日志不记录 password_hash
日志不记录 session_key
日志不记录 AppSecret
openid / unionid 不完整展示
生产 .env 不提交 Git
```

## 10. 上线当天操作顺序

```text
1. 冻结代码
2. 备份数据库
3. 执行 migration
4. 部署后端
5. 验证 /api/health
6. 部署 admin-web
7. 部署 web
8. 验证 HTTPS 和 Nginx
9. 上传小程序体验版
10. 真机测试
11. 提交审核 / 发布
12. 观察日志
13. 记录发布结果
```

## 11. 不建议上线的情况

```text
HTTPS 未完成
小程序合法域名未配置
隐私政策未准备
数据库未备份
默认 admin 密码未修改
权限测试未通过
账号合并测试未通过
家庭成员关系测试未通过
operation_logs 未写入
```

## 12. M16 自动门禁

发布候选版本必须执行：

```bash
bash scripts/security-check.sh
bash scripts/m16-verify.sh
```

`m16-verify.sh` 包含：

```text
Docker Compose 配置校验
后端 go test ./...
admin-web 构建
PC/H5 web 构建
微信小程序构建
仓库安全检查
```

任一子命令失败时总体结论必须为 FAIL。禁止跳过失败、忽略退出码或仅凭人工页面可打开标记 PASS。

## 13. 上线证据要求

每个 P0 项必须有可追溯证据：

```text
命令和退出码
测试报告或脱敏截图
staging 发布版本
migration 版本
数据库备份记录
恢复演练记录
权限越权测试结果
operation_logs 抽查结果
graph_version 操作矩阵结果
小程序体验版设备矩阵结果
```

证据中不得包含真实密码、Token、验证码、openid、unionid、AppSecret、数据库密码或 Redis 密码。

## 14. 当前原型限制

如果 `admin-web`、`web` 或 `miniapp` 仍使用 mock data 或 mock 登录：

```text
构建成功只能证明工程可编译
不能证明真实认证、权限或 API 流程可用
不能作为生产上线通过依据
体验版不得保留 mock、测试按钮或示例数据
```

必须完成真实 API 对接、安全评审和 staging 验收后，才能将对应 P0 项标记为 PASS。

## 15. Staging 部署门禁

使用 `scripts/deploy-staging.sh` 前确认：

```text
APP_ENV=staging
已执行发布前备份
只执行 forward-only migration
发布物与 Git 版本可追溯
健康检查地址属于 staging
上一应用发布目录可用于回滚
rollback-staging.sh 不执行 migration down
```

staging 验证失败时不得继续 production 发布。
