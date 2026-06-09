<template>
  <header class="header">
    <RouterLink class="brand" to="/">
      <span>Tree</span>
      <strong>家脉亲缘</strong>
    </RouterLink>
    <nav>
      <RouterLink to="/">首页</RouterLink>
      <RouterLink to="/families">搜索家庭</RouterLink>
      <RouterLink to="/me/families">我的家庭</RouterLink>
      <RouterLink to="/me">我的</RouterLink>
    </nav>
    <RouterLink class="button secondary" to="/login">登录 / 注册</RouterLink>
  </header>
</template>

<style scoped>
.header {
  position: sticky;
  top: 0;
  z-index: 10;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
  min-height: 68px;
  padding: 0 max(16px, calc((100vw - 1120px) / 2));
  border-bottom: 1px solid var(--color-border);
  background: rgba(247, 243, 234, 0.94);
  backdrop-filter: blur(12px);
}

.brand {
  display: flex;
  align-items: center;
  gap: 10px;
  font-weight: 700;
}

.brand span {
  display: grid;
  width: 34px;
  height: 34px;
  place-items: center;
  border-radius: 8px;
  background: var(--color-primary);
  color: #fff;
}

nav {
  display: flex;
  gap: 18px;
  color: var(--color-text-secondary);
}

@media (max-width: 760px) {
  .header {
    align-items: flex-start;
    flex-direction: column;
    padding-block: 12px;
  }

  nav {
    flex-wrap: wrap;
  }
}
</style>
