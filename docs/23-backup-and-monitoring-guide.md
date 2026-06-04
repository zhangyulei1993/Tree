# 23. 备份、日志与监控指南 v1

> 一期规模不大也必须具备基本备份、日志、恢复和服务监控能力。

## 1. 需要备份的数据

```text
MySQL 数据库
uploads 上传文件
.env 生产配置备份，安全保存
Nginx 配置
后端发布包
前端 dist 版本
```

Redis 不作为唯一持久数据源：

```text
验证码
token 黑名单
缓存数据
限流数据
```

## 2. MySQL 备份策略

建议：

```text
每日全量备份
重要发布前手动备份
保留最近 7 天每日备份
每周保留 1 份，保留 4 周
每月保留 1 份，保留 6 个月
```

备份命令：

```bash
mysqldump -h 127.0.0.1 -u tree_user -p tree_platform \
  --single-transaction \
  --routines \
  --triggers \
  > /opt/Tree/backups/mysql/tree_platform_$(date +%F_%H%M%S).sql
```

压缩：

```bash
gzip tree_platform_2026-06-04_120000.sql
```

## 3. 上传文件备份

```bash
tar -czf uploads_$(date +%F_%H%M%S).tar.gz /opt/Tree/uploads
```

如果后续使用 OSS / COS：

```text
启用对象存储版本控制
启用生命周期策略
定期校验文件可访问性
```

## 4. 备份位置

不要只存在同一台服务器。

建议：

```text
本地服务器一份
对象存储一份
异地备份一份，后续可选
```

## 5. 恢复演练

至少每月一次：

```text
新建临时数据库
导入最近备份
启动测试后端连接该库
验证用户、家庭、成员、关系、日志可查询
```

## 6. 日志分类

系统运行日志：

```text
Go 后端运行日志
Nginx access.log
Nginx error.log
MySQL error log
Redis log
```

业务审计日志：

```text
operation_logs
```

## 7. Go 后端日志

目录：

```text
/opt/Tree/backend/logs
```

建议：

```text
app-2026-06-04.log
error-2026-06-04.log
```

日志级别：

```text
dev: debug
staging: info
production: info / warn
```

禁止记录：

```text
密码
验证码
token
session_key
AppSecret
数据库密码
完整 openid / unionid
```

## 8. 监控指标

一期至少检查：

```text
服务器 CPU
内存
磁盘空间
MySQL 连接数
Redis 内存
后端进程状态
Nginx 5xx 数量
接口响应时间
数据库备份是否成功
HTTPS 证书是否临近过期
```

## 9. 告警建议

```text
磁盘使用率 > 80%
后端服务不可用
MySQL 不可用
Redis 不可用
HTTPS 证书即将过期
数据库备份失败
Nginx 5xx 异常增加
```

## 10. 服务自动重启

systemd：

```text
Restart=always
RestartSec=5
```

Docker：

```yaml
restart: unless-stopped
```

## 11. 发布记录

每次发布记录：

```text
发布时间
发布人
后端版本
前端版本
小程序版本
数据库 migration 版本
变更内容
是否已备份
是否回滚过
```

建议文件：

```text
RELEASE_LOG.md
```
