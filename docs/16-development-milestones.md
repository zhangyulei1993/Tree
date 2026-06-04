# 16. 开发里程碑计划 v1

> 目标：把一期项目拆成可以逐步交付和验收的阶段。Codex 或开发团队应按阶段推进，不要一次性实现全部功能。

## M1. 项目骨架与本地环境

范围：Go 后端骨架、admin-web、web、miniapp、docker-compose、`.env.example`、统一返回、错误码、日志框架、健康检查。

验收：MySQL/Redis 可启动；`go run ./cmd/server` 可启动；`GET /api/health` 成功；前端项目能启动。

## M2. 数据库 migration 与基础模型

范围：创建所有核心表、GORM model、seed ROOT_ADMIN。

验收：migration 可执行；表结构符合 `14-database-migration-ddl.md`；GORM model 可编译。

## M3. 认证与账号

接口：send-code、register-phone、login-phone、logout、wechat-mini/login、wechat-mini/bind-phone、change-phone、cancel-account。

验收：验证码只保存 hash；小程序临时账号可绑定手机号；绑定已有手机号触发合并；绑定预创建账号触发认领；注销释放 phone；旧 token 失效。

## M4. 后台管理员认证

接口：admin login/logout/me/password。

验收：ROOT_ADMIN 可登录；后台 token 与用户 token 分离；SUPER_ADMIN/PLATFORM_ADMIN 登录失败可锁定；ROOT_ADMIN 登录失败不自动锁定。

## M5. 家庭基础

接口：创建家庭、搜索家庭、家庭详情、编辑家庭、公开主页、家庭树。

验收：创建家庭时创建 founder member 和 FOUNDER link；游客只能搜索 NORMAL + searchable 家庭；未公开家庭不能访问公开主页；家庭树返回 nodes/edges/tree。

## M6. 家庭成员与关系

范围：成员增删改、无需绑定、父母/子女/兄弟姐妹/配偶关系、graph_version。

验收：存在下级成员不能删除；兄弟姐妹无父母节点必须失败；SIBLING 不落库；PRIMARY 父/母唯一；多个配偶允许。

## M7. 邀请与加入家庭

范围：站内邀请、分享链接邀请、接受/拒绝/撤销邀请、加入申请、处理申请。

验收：PLATFORM_ADMIN 不能生成分享邀请链接；邀请 7 天有效；接受邀请创建 ACTIVE link；接受邀请不触发 graph_version。

## M8. 公开申请与游客留言

验收：同一 family 只能有一个 PENDING 公开申请；PLATFORM_ADMIN 不能审核自己发起的公开申请；留言审核通过后展示；公开留言不展示联系方式。

## M9. 家庭角色、创始人转让、家庭解散

验收：只有 FOUNDER 可设置 FAMILY_ADMIN；NOT_REQUIRED 成员不能成为管理员；创始人转让和家庭解散必须 ROOT_ADMIN/SUPER_ADMIN 审核；PLATFORM_ADMIN 无权审核。

## M10. 后台管理页面与接口

范围：用户管理、家庭管理、成员管理、公开申请审核、留言审核、转让审核、解散审核、管理员管理。

验收：权限边界正确；高风险操作有确认弹窗；所有敏感操作写日志。

## M11. 操作日志

验收：后台敏感操作、普通用户重要操作、失败操作均写日志；日志可筛选和查看详情；敏感字段过滤。

## M12. PC/H5 前台

验收：首页、搜索、公开主页、公开树、登录注册、个人中心、我的家庭、邀请详情、加入申请可用。

## M13. 微信小程序

验收：微信快捷登录、绑定手机号、家庭搜索、公开主页、公开树、我的家庭、邀请确认、加入申请、个人中心可用。

## M14. 测试与修复

验收：`11-test-cases-and-acceptance.md` 中 P0 全部通过；P1 主要流程通过；权限越权、账号合并、graph_version、operation_logs 测试通过。

## M15. 部署上线准备

范围：Dockerfile、docker-compose.prod.yml、Nginx、HTTPS、MySQL 备份、Redis 配置、上传目录、日志目录、默认密码修改。

## 推荐顺序

```text
M1 -> M2 -> M3 -> M4 -> M5 -> M6 -> M7 -> M8 -> M9 -> M10 -> M11 -> M12 -> M13 -> M14 -> M15
```

关键风险阶段：M3 账号合并、M6 成员关系、M9 转让/解散、M11 日志。
