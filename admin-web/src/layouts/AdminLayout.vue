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
}

.admin-main {
  min-width: 0;
}

.admin-content {
  padding: 20px 24px 32px;
}
</style>
