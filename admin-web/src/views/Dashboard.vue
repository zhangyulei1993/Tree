<template>
  <div class="page-stack">
    <PageHeader title="仪表盘" description="展示后台运营与待处理事项的 mock 概览。" />
    <div class="stats-grid">
      <section v-for="item in dashboardStats" :key="item.label" class="surface stat-card">
        <span>{{ item.label }}</span>
        <strong>{{ item.value }}</strong>
        <em>{{ item.trend }}</em>
      </section>
    </div>
    <div class="dashboard-grid">
      <section class="surface panel">
        <h3>待处理审核</h3>
        <el-timeline>
          <el-timeline-item timestamp="公开申请" type="warning">李氏宗亲提交公开展示申请</el-timeline-item>
          <el-timeline-item timestamp="游客留言" type="primary">张氏家族有 1 条新留言待审核</el-timeline-item>
          <el-timeline-item timestamp="高风险" type="danger">王氏家谱提交解散申请</el-timeline-item>
        </el-timeline>
      </section>
      <section class="surface panel">
        <h3>最近操作日志</h3>
        <el-table :data="operationLogs.slice(0, 3)" size="small">
          <el-table-column prop="action" label="动作" min-width="210" />
          <el-table-column prop="actor" label="操作者" width="150" />
          <el-table-column prop="result" label="结果" width="110" />
        </el-table>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import PageHeader from '@/components/PageHeader.vue'
import { dashboardStats, operationLogs } from '@/mock/data'
</script>

<style scoped>
.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 14px;
}

.stat-card {
  display: flex;
  min-height: 118px;
  flex-direction: column;
  justify-content: space-between;
  padding: 18px;
}

.stat-card span,
.stat-card em {
  color: var(--color-text-secondary);
  font-size: 13px;
  font-style: normal;
}

.stat-card strong {
  font-size: 30px;
}

.dashboard-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.panel {
  padding: 18px;
}

h3 {
  margin: 0 0 16px;
}
</style>
