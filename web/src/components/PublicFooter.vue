<template>
  <footer class="footer">
    <div class="container">
      <div>
        <strong>Tree 家脉亲缘</strong>
        <span>整理家庭关系，保存家族记忆。</span>
      </div>
      <div class="footer-meta">
        <span>家庭资料仅在授权范围内可见，公开展示需经过审核。</span>
        <span class="build-meta">版本 {{ buildInfo.version }} · {{ buildInfo.commit }} · {{ buildInfo.environment }}</span>
      </div>
    </div>
  </footer>
</template>

<script setup lang="ts">
import { buildInfo } from '@/buildInfo'
</script>

<style scoped>
.footer {
  margin-top: 64px;
  border-top: 1px solid rgba(148, 163, 184, 0.14);
  background: rgba(255, 255, 255, 0.68);
  padding: 36px 0 42px;
  color: var(--color-text-secondary);
  backdrop-filter: blur(16px);
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

.footer-meta {
  justify-items: end;
  text-align: right;
}

.build-meta {
  color: rgba(100, 116, 139, 0.72);
  font-size: 12px;
  font-variant-numeric: tabular-nums;
}

strong {
  display: block;
  color: var(--color-primary);
  font-size: 15px;
}

@media (max-width: 720px) {
  .container {
    flex-direction: column;
  }

  .footer-meta {
    justify-items: center;
    text-align: center;
  }
}
</style>
