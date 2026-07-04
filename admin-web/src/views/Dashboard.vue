<template>
  <div class="page-stack">
    <section class="dashboard-hero">
      <div>
        <span class="dashboard-kicker">Tree 管理后台</span>
        <h1>平台运营总览</h1>
        <p>聚合用户、家庭、审核和高风险流程，优先处理需要人工介入的事项。</p>
      </div>
      <div class="hero-meter">
        <span>待处理</span>
        <strong>{{ pendingTotal }}</strong>
      </div>
    </section>
    <PageHeader title="仪表盘" description="平台实时业务概览。"/>
    <el-alert v-if="error" :title="error" type="error" show-icon/>
    <div v-loading="loading" class="stats-grid">
      <section v-for="item in cards" :key="item.label" class="surface stat-card" :class="{ urgent: item.urgent }">
        <span>{{ item.label }}</span>
        <strong>{{ item.value }}</strong>
        <em>{{ item.note }}</em>
      </section>
    </div>
  </div>
</template>
<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { getDashboard } from '@/api/management'; import { getApiErrorMessage } from '@/api/client'; import PageHeader from '@/components/PageHeader.vue'; import type { DashboardStats } from '@/types/api'
const stats=ref<DashboardStats>({users:0,families:0,pendingPublicApplications:0,pendingFounderTransfers:0,pendingDissolutions:0});const loading=ref(true);const error=ref('')
const pendingTotal = computed(() => stats.value.pendingPublicApplications + stats.value.pendingFounderTransfers + stats.value.pendingDissolutions)
const cards=computed(()=>[{label:'用户数',value:stats.value.users,note:'平台账号'},{label:'家庭数',value:stats.value.families,note:'全部家庭'},{label:'公开申请',value:stats.value.pendingPublicApplications,note:'待审核',urgent:stats.value.pendingPublicApplications>0},{label:'创建者转让',value:stats.value.pendingFounderTransfers,note:'待审核',urgent:stats.value.pendingFounderTransfers>0},{label:'家庭解散',value:stats.value.pendingDissolutions,note:'待审核',urgent:stats.value.pendingDissolutions>0}])
onMounted(async()=>{try{stats.value=await getDashboard()}catch(e){error.value=getApiErrorMessage(e)}finally{loading.value=false}})
</script>
<style scoped>
.stats-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 16px;
}

.dashboard-hero {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 24px;
  border: 1px solid rgba(255, 255, 255, 0.16);
  border-radius: 28px;
  background:
    radial-gradient(circle at 88% 10%, rgba(200, 164, 93, 0.24), transparent 280px),
    radial-gradient(circle at 10% 100%, rgba(47, 107, 87, 0.26), transparent 260px),
    linear-gradient(135deg, #172f4a 0%, #244f54 62%, #2f6b57 100%);
  padding: 30px;
  color: #fff;
  box-shadow: 0 28px 70px rgba(18, 32, 51, 0.22);
}

.dashboard-kicker {
  display: inline-flex;
  border: 1px solid rgba(255, 255, 255, 0.18);
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.10);
  color: rgba(248, 231, 194, 0.92);
  padding: 7px 12px;
  font-size: 12px;
  font-weight: 800;
  letter-spacing: 0.08em;
}

.dashboard-hero h1 {
  margin: 16px 0 10px;
  font-size: 40px;
  line-height: 1.05;
  letter-spacing: -0.03em;
}

.dashboard-hero p {
  max-width: 620px;
  margin: 0;
  color: rgba(255, 255, 255, 0.74);
  line-height: 1.7;
}

.hero-meter {
  min-width: 150px;
  border: 1px solid rgba(255, 255, 255, 0.18);
  border-radius: 22px;
  background: rgba(255, 255, 255, 0.12);
  padding: 18px;
  text-align: right;
}

.hero-meter span {
  display: block;
  color: rgba(255, 255, 255, 0.72);
  font-size: 13px;
  font-weight: 700;
}

.hero-meter strong {
  display: block;
  margin-top: 6px;
  color: #fff;
  font-size: 42px;
  line-height: 1;
}

.stat-card {
  position: relative;
  display: flex;
  min-height: 132px;
  flex-direction: column;
  justify-content: space-between;
  padding: 20px;
  background:
    radial-gradient(circle at 100% 0%, rgba(200, 164, 93, 0.10), transparent 120px),
    linear-gradient(135deg, rgba(255, 255, 255, 0.96) 0%, rgba(244, 250, 247, 0.92) 100%);
  overflow: hidden;
}

.stat-card::before {
  content: '';
  position: absolute;
  left: 0;
  top: 22px;
  bottom: 22px;
  width: 5px;
  border-radius: 0 999px 999px 0;
  background: linear-gradient(180deg, var(--color-primary) 0%, var(--color-success) 100%);
}

.stat-card.urgent::before {
  background: linear-gradient(180deg, var(--color-warning) 0%, var(--color-danger) 100%);
}

.stat-card span,
.stat-card em {
  color: var(--color-text-secondary);
  font-size: 13px;
  font-style: normal;
}

.stat-card strong {
  color: var(--color-primary);
  font-size: 34px;
  letter-spacing: -0.03em;
}

@media (max-width: 900px) {
  .stats-grid {
    grid-template-columns: 1fr 1fr;
  }
}

@media (max-width: 640px) {
  .stats-grid {
    grid-template-columns: 1fr;
  }
}
</style>
