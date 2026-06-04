# 06. 家庭成员与关系规则

## 1. 成员节点类型

```text
LINEAGE_MEMBER
SPOUSE
EXTERNAL_MEMBER
```

## 2. 姓名结构

```text
surname               姓
generation_character  字辈
given_name            名
display_name          展示名
```

`generation_character` 可填可空，不强制自动提取。

## 3. 用户绑定策略

```text
OPTIONAL
REQUIRED
NOT_REQUIRED
DISABLED
```

## 4. 家庭树绑定状态

接口派生字段：

```text
BOUND
INVITING
NOT_REQUIRED
UNBOUND
```

## 5. relationship_type

一期只保留：

```text
PARENT_CHILD
SPOUSE
```

兄弟姐妹不单独存 `SIBLING`，由共同父母推导。

## 6. parent_link_type

```text
PRIMARY
STEP
ADOPTIVE
SUCCESSION
NOTE_ONLY
OTHER
```

## 7. 添加兄弟姐妹

有父母节点：

```text
直接从本人添加兄弟姐妹。
系统自动挂到共同父母下。
最终写入 PARENT_CHILD。
```

无父母节点：

```text
提示先创建父母节点。
```

## 8. 多父母规则

本姓男性直系成员的子女：

```text
PRIMARY 父亲只能一个
母亲可以多个
```

本姓女性直系成员的子女：

```text
PRIMARY 母亲只能一个
父亲可以多个
```

## 9. 多配偶规则

男性节点可以存在多个女性配偶。

女性节点也可以存在多个男性配偶。

## 10. 删除成员

```text
存在下级成员 -> 禁止删除
主成员有配偶 -> 提示会连同配偶一起删除
删除配偶节点 -> 只删除配偶节点
```
