# 09. 缓存与性能规则

## 1. 核心判断

家庭树和关系计算属于读多写少、关系型查询，适合缓存。

一期不强依赖 Redis，但必须预留缓存设计。

## 2. graph_version

`families.graph_version` 用于家庭树缓存失效。

缓存 key 建议：

```text
family:{familyId}:tree:v{graphVersion}
```

## 3. 触发 graph_version + 1

```text
新增成员
删除成员
恢复成员
修改成员关键字段
新增关系
删除关系
修改 parent_link_type
修改关系备注
修改字辈
修改人工承继
```

## 4. 不一定触发 graph_version

```text
修改家庭简介
修改公开联系方式
修改公开展示状态
用户绑定 / 解绑 member
邀请接受 / 拒绝
角色变化
创始人转让
家庭解散
游客留言
```

## 5. 查询建议

家庭树接口不应层层查数据库。

推荐：

```text
一次性加载 family 下 ACTIVE members
一次性加载 family 下 ACTIVE relationships
后端内存组装 tree
返回 nodes + edges + tree
```
