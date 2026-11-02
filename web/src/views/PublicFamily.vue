<template>
  <PageShell>
    <section class="container family-page">
      <section v-if="loadingFamily" class="card state-panel">正在加载优质家庭案例...</section>
      <section v-else-if="familyError" class="card state-panel error" role="alert">
        <strong>优质家庭案例不可访问</strong>
        <span>{{ familyError }}</span>
        <button class="button secondary" @click="loadFamily">重新加载</button>
      </section>

      <template v-else-if="family">
        <div class="card profile showcase-banner">
          <div class="profile-head">
            <div class="family-seal">{{ family.familySurname.slice(0, 1) }}</div>
            <div class="profile-copy">
              <span class="eyebrow">优质家庭案例</span>
              <h1>{{ family.familyName }}</h1>
              <p class="surname-line">{{ family.familySurname }}氏</p>
            </div>
          </div>
          <p class="intro">{{ family.description || '该家庭暂未填写公开简介。' }}</p>
          <div class="profile-tags">
            <span v-if="family.nativePlace" class="tag-chip">{{ family.nativePlace }}</span>
            <span v-if="family.regionText" class="tag-chip green">{{ family.regionText }}</span>
          </div>
          <dl class="facts">
            <div><dt>姓氏</dt><dd>{{ family.familySurname }}</dd></div>
            <div><dt>籍贯</dt><dd>{{ family.nativePlace || '未设置' }}</dd></div>
            <div><dt>地区</dt><dd>{{ family.regionText || '未设置' }}</dd></div>
          </dl>
          <div class="contact surface-soft">
            <template v-if="family.publicContactVisible && hasPublicContact">
              <strong>{{ family.publicContactName || '公开联系方式' }}</strong>
              <span v-if="family.publicContactPhone">电话：{{ family.publicContactPhone }}</span>
              <span v-if="family.publicContactWechat">微信：{{ family.publicContactWechat }}</span>
              <span v-if="family.publicContactNote">{{ family.publicContactNote }}</span>
            </template>
            <span v-else>该家庭未设置公开联系方式，如需联系请通过平台协助。</span>
          </div>
          <div class="actions">
            <RouterLink class="button" :to="`/families/${family.id}/tree/public`">学习家庭树</RouterLink>
            <span class="learning-note">仅供学习参考，家庭成员请通过家庭内部邀请加入。</span>
          </div>
        </div>

      </template>
    </section>
  </PageShell>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'

import { apiErrorMessage } from '@/api/client'
import { getPublicFamilyDetail } from '@/api/families'
import PageShell from '@/components/PageShell.vue'
import type { PublicFamily } from '@/types/api'

const route = useRoute()
const familyId = computed(() => String(route.params.familyId))
const family = ref<PublicFamily | null>(null)
const loadingFamily = ref(true)
const familyError = ref('')

const hasPublicContact = computed(() => Boolean(
  family.value?.publicContactName ||
  family.value?.publicContactPhone ||
  family.value?.publicContactWechat ||
  family.value?.publicContactNote
))

async function loadFamily() {
  loadingFamily.value = true
  familyError.value = ''
  try {
    family.value = await getPublicFamilyDetail(familyId.value)
  } catch (requestError) {
    family.value = null
    familyError.value = apiErrorMessage(requestError, '无法加载优质家庭案例。')
  } finally {
    loadingFamily.value = false
  }
}

onMounted(loadFamily)
</script>

<style scoped>
.family-page {
  display: grid;
  gap: 18px;
  padding: 48px 0 32px;
}

.profile,
.panel {
  padding: 28px;
}

.showcase-banner {
  position: relative;
  overflow: hidden;
  background: var(--gradient-hero);
  border-color: var(--color-border);
  box-shadow: var(--shadow-soft);
}

.showcase-banner::before {
  content: '';
  position: absolute;
  top: -24px;
  right: -24px;
  width: 140px;
  height: 140px;
  border: 1px solid var(--color-border);
  border-radius: 50%;
  opacity: 0.22;
  pointer-events: none;
}

.profile-head {
  display: flex;
  gap: 18px;
  align-items: flex-start;
}

.family-seal {
  display: grid;
  flex-shrink: 0;
  place-items: center;
  width: 82px;
  height: 82px;
  border: 2px solid var(--color-warm-gold);
  border-radius: 28px;
  background: linear-gradient(145deg, #fffdf9 0%, #edf8f2 100%);
  color: var(--color-primary);
  font-size: 32px;
  font-weight: 800;
}

.profile-copy h1 {
  margin: 8px 0 0;
}

.surname-line {
  margin: 6px 0 0;
  color: var(--color-warm-gold-text);
  font-weight: 700;
}

.intro {
  margin-top: 16px;
}

.profile-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 14px;
}

.tag-chip {
  display: inline-flex;
  border-radius: 999px;
  background: rgba(31, 58, 95, 0.08);
  color: var(--color-primary);
  padding: 6px 12px;
  font-size: 12px;
  font-weight: 700;
}

.tag-chip.green {
  background: var(--color-heritage-green-light);
  color: var(--color-heritage-green);
}

.facts {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
  margin-top: 18px;
  border-radius: 20px;
  background: rgba(255, 255, 255, 0.68);
  padding: 16px;
}

.eyebrow {
  color: var(--color-warm-gold-text);
  font-weight: 700;
  letter-spacing: 0.08em;
}

h1 {
  font-size: clamp(30px, 4vw, 40px);
  color: var(--color-primary);
}

h2,
p {
  margin-top: 0;
}

p {
  line-height: 1.8;
}

.contact {
  display: grid;
  gap: 6px;
  margin-top: 18px;
  padding: 14px;
}

.actions,
.panel {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.actions {
  align-items: flex-start;
  flex-direction: row;
  margin-top: 18px;
}

.panel-heading,
.form-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.panel-heading h2 {
  margin: 0;
}

.form-meta {
  color: var(--color-text-secondary);
  font-size: 13px;
}

.button:disabled {
  cursor: not-allowed;
  opacity: 0.6;
}

.feedback,
.inline-state {
  margin: 0;
  font-weight: 700;
}

.feedback.error,
.inline-state.error,
.state-panel.error {
  color: var(--color-danger);
}

.feedback.success {
  color: var(--color-success);
}

.inline-state,
.state-panel {
  padding: 22px;
  text-align: center;
}

.state-panel {
  display: grid;
  min-height: 220px;
  place-items: center;
  gap: 12px;
}

dt {
  color: var(--color-text-secondary);
}

dd {
  margin: 4px 0 0;
  font-weight: 700;
}

@media (max-width: 820px) {
  .facts {
    grid-template-columns: 1fr;
  }

  .actions {
    align-items: stretch;
    flex-direction: column;
  }
}
</style>
