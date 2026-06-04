# 18. Tree 家脉亲缘平台设计语言 v1

> 用于约束管理后台、PC/H5 前台、微信小程序的视觉风格、组件规则和 Codex 生成 UI 原型时的边界。

## 1. 设计理念

关键词：

```text
清晰、稳重、可信、家族感、低干扰、管理友好、移动端简洁
```

不同端的重点：

| 端 | 设计重点 |
|---|---|
| 管理后台 | 效率、信息密度、权限清晰、危险操作可控 |
| PC/H5 前台 | 信任感、家族文化感、搜索与公开展示清晰 |
| 微信小程序 | 流程短、按钮明确、邀请确认顺畅 |

避免：

```text
大红大金堆满、国潮过重、复杂动效、暗黑科技风、娱乐化短视频风
```

## 2. 色彩令牌

建议建立：

```text
admin-web/src/styles/tokens.css
web/src/styles/tokens.css
miniapp/src/styles/tokens.scss
```

基础 token：

```css
:root {
  --color-primary: #1F3A5F;
  --color-primary-hover: #2B4A73;
  --color-heritage-green: #2F6B57;
  --color-warm-gold: #C8A45D;
  --color-bg-public: #F7F3EA;
  --color-bg-admin: #F5F7FA;
  --color-bg-card: #FFFFFF;
  --color-text-primary: #1F2937;
  --color-text-secondary: #6B7280;
  --color-text-disabled: #9CA3AF;
  --color-border: #E5E7EB;
  --color-danger: #C0392B;
  --color-warning: #D97706;
  --color-success: #2F6B57;
}
```

## 3. 字体

```css
body {
  font-family:
    Inter,
    system-ui,
    -apple-system,
    BlinkMacSystemFont,
    "Segoe UI",
    "PingFang SC",
    "Microsoft YaHei",
    sans-serif;
}
```

## 4. 圆角与间距

```css
:root {
  --radius-card: 12px;
  --radius-button: 8px;
  --radius-input: 8px;
  --radius-dialog: 14px;
  --radius-tag: 999px;

  --shadow-light: 0 2px 8px rgba(15, 23, 42, 0.06);
  --shadow-medium: 0 8px 24px rgba(15, 23, 42, 0.10);
}
```

间距使用 4px 网格：

```text
4 / 8 / 12 / 16 / 20 / 24 / 32 / 40 / 48
```

## 5. 后台 UI 规则

后台基于：

```text
Vue 3 + TypeScript + Vite + Element Plus
```

组件：

```text
AdminLayout
AdminSidebar
AdminTopbar
PageHeader
SearchPanel
DataTable
StatusTag
RoleTag
ActionButtonGroup
ConfirmDialog
AuditDialog
DetailDrawer
JsonViewer
FamilyTreePanel
MemberDetailPanel
```

规则：

```text
1. 所有列表页顶部都有 PageHeader。
2. 筛选条件放在 SearchPanel。
3. 表格统一使用 DataTable。
4. 状态统一用 StatusTag。
5. 危险操作统一用 ConfirmDialog。
6. 审核操作统一用 AuditDialog。
7. 详情信息优先用 DetailDrawer 或详情页。
8. 后台不使用重装饰传统纹样。
```

## 6. 前台 UI 规则

组件：

```text
PublicHeader
PublicFooter
HeroSearch
FamilyCard
PublicFamilyProfileCard
PublicTreeList
VisitorMessageList
VisitorMessageForm
LoginCard
RegisterCard
EmptyState
```

规则：

```text
1. 留白比后台更多。
2. 家族卡片使用温和背景和清晰边框。
3. 公开主页可以有轻微传统文化感。
4. 不直接展示 users.phone，只展示 families.public_contact_*。
```

## 7. 小程序 UI 规则

组件：

```text
MiniPage
MiniCard
MiniPrimaryButton
MiniSecondaryButton
PhoneBindCard
InviteInfoCard
FamilyMiniCard
MiniTreeList
MiniEmptyState
```

规则：

```text
1. 一屏只解决一个任务。
2. 主按钮必须明显。
3. 表单尽量短。
4. 邀请和家庭信息用卡片展示。
5. 关键操作必须二次确认。
6. 未绑定手机号时明确提示下一步。
```

## 8. 业务 UI 规则

```text
1. PLATFORM_ADMIN 不应看到创始人转让和家庭解散审核按钮。
2. NOT_REQUIRED 成员必须有清晰标签。
3. INVITING 成员必须展示邀请状态。
4. 删除存在下级成员时必须显示阻断提示。
5. ADD_SIBLING 无父母节点时提示：请先创建父亲或母亲节点。
6. 公开家庭无联系方式时显示：该家庭未设置公开联系方式，如需联系请通过平台协助。
7. 高风险操作必须二次确认。
```

## 9. Codex UI 生成要求

```text
1. 不接真实 API，先使用 mock data。
2. 不修改后端。
3. 不引入新的 UI 框架。
4. 不写死散乱颜色，优先使用 CSS variables。
5. 不实现复杂业务逻辑。
6. 输出页面路径、组件结构和运行方式。
```
