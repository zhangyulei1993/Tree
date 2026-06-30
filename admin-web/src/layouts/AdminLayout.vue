<template>
  <div class="admin-shell">
    <AdminSidebar />
    <main class="admin-main">
      <AdminTopbar />
      <section class="admin-content">
        <router-view />
      </section>
    </main>
  </div>
</template>

<script setup lang="ts">
import AdminSidebar from '@/components/AdminSidebar.vue'
import AdminTopbar from '@/components/AdminTopbar.vue'
</script>

<style scoped>
.admin-shell {
  display: grid;
  min-height: 100vh;
  grid-template-columns: 248px minmax(0, 1fr);
  background:
    radial-gradient(circle at 16% 0%, rgba(59, 110, 168, 0.06), transparent 320px),
    radial-gradient(circle at 90% 10%, rgba(47, 107, 87, 0.06), transparent 360px),
    var(--color-bg-admin);
}

.admin-main {
  min-width: 0;
}

.admin-content {
  padding: 24px 28px 40px;
}

@media (max-width: 960px) {
  .admin-shell {
    grid-template-columns: 1fr;
  }

  .admin-content {
    padding: 18px 16px 28px;
  }
}
</style>
