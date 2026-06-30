<template>
  <div class="page">
    <PublicHeader />
    <main>
      <slot />
    </main>
    <PublicFooter />
  </div>
</template>

<script setup lang="ts">
import PublicFooter from '@/components/PublicFooter.vue'
import PublicHeader from '@/components/PublicHeader.vue'
</script>

<style scoped>
.page {
  position: relative;
  overflow: clip;
}

.page::before,
.page::after {
  content: '';
  position: fixed;
  z-index: 0;
  width: 360px;
  height: 360px;
  border-radius: 50%;
  pointer-events: none;
  filter: blur(8px);
}

.page::before {
  top: -160px;
  left: -100px;
  background: radial-gradient(circle, rgba(59, 110, 168, 0.14) 0%, rgba(59, 110, 168, 0) 72%);
}

.page::after {
  right: -120px;
  bottom: 8vh;
  background: radial-gradient(circle, rgba(47, 107, 87, 0.14) 0%, rgba(47, 107, 87, 0) 70%);
}

main {
  position: relative;
  z-index: 1;
}
</style>
