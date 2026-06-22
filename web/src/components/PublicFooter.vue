<template>
  <footer class="footer">
    <div class="container">
      <div>
        <strong>Tree 家脉亲缘</strong>
        <span>整理家庭关系，保存家族记忆。</span>
      </div>
      <span>家庭资料仅在授权范围内可见，公开展示需经过审核。</span>
    </div>
  </footer>
</template>

<style scoped>
.footer {
  margin-top: 64px;
  border-top: 1px solid rgba(148, 163, 184, 0.14);
  background: rgba(255, 255, 255, 0.62);
  padding: 32px 0;
  color: var(--color-text-secondary);
}

.container {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
}

.container div {
  display: grid;
  gap: 4px;
}

strong {
  display: block;
  color: var(--color-primary);
}

@media (max-width: 720px) {
  .container {
    flex-direction: column;
  }
}
</style>
