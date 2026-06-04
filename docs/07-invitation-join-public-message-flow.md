# 07. 邀请、加入、公开申请与留言流程

## 1. 邀请流程

### 平台管理员

```text
填写节点信息
手机号搜索用户池
找到用户 -> 创建节点 + 站内邀请
找不到用户 -> 不能发邀请，只能选择无需绑定用户
```

### 家族管理员 / 家庭创始人

```text
创建或定位节点
生成分享链接
用户点击链接
查看邀请详情
确认绑定前必须登录 / 注册 / 绑定手机号
接受后创建绑定
```

## 2. 邀请规则

```text
有效期 7 天
访问时实时判断过期
拒绝后可以重新邀请
接受后默认 family_role = MEMBER
不同 member 可以同时有邀请
同一个 member 同一时间只能有一个 PENDING 邀请
```

## 3. 用户申请加入家庭

申请条件：

```text
user.status = ACTIVE
phone_verified = 1
family.status = NORMAL
user 未加入该 family
不存在 PENDING 申请
```

处理人为 FOUNDER / FAMILY_ADMIN 为主，平台管理员只协助。

## 4. 家庭公开申请

```text
FOUNDER / FAMILY_ADMIN 可以申请公开
PLATFORM_ADMIN 可以代申请
PLATFORM_ADMIN 可以审核公开申请
但不能审核自己发起的申请
同一 family 只能有一个 PENDING 公开申请
```

## 5. 游客留言

```text
允许匿名
联系方式非必填
留言内容必填
1000 字以内
1 分钟同 IP 最多 1 条
1 天同 IP 最多 20 条
审核通过后直接展示
created_at DESC
```
