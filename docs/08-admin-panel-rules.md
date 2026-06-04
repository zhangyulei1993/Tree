# 08. 后台管理规则

## 1. 后台角色

```text
ROOT_ADMIN
SUPER_ADMIN
PLATFORM_ADMIN
```

## 2. ROOT_ADMIN

```text
username = admin
有且只有一个
不能删除
不能禁用
不能降级
登录失败不自动锁定
可以管理 SUPER_ADMIN
可以解除管理员锁定
可以配置管理员锁定策略
```

## 3. SUPER_ADMIN

```text
拥有全部业务权限
可以管理 PLATFORM_ADMIN
可以审核创始人转让
可以审核家庭解散
可以恢复已删除 / 已解散家庭
可以下架公开家庭
不能管理同级 SUPER_ADMIN
不能管理 ROOT_ADMIN
```

## 4. PLATFORM_ADMIN

可以：

```text
查看用户
查看完整手机号
编辑普通用户资料
禁用普通用户
查看家庭
编辑家庭
编辑成员
删除成员
协助用户完成家庭关系
审核家庭公开申请
审核游客留言
下架公开家庭
预创建账号
```

不可以：

```text
生成分享邀请链接
审核创始人转让
审核解散家庭
恢复已删除 / 已解散家庭
删除普通用户
重置手机号
修改 openid / unionid
导出数据
管理同级管理员
管理 SUPER_ADMIN
```

## 5. 操作日志

所有后台敏感操作必须写 `operation_logs`。

一期日志支持筛选、搜索、分页、详情查看，不支持导出、删除、清空。
