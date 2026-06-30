<template>
  <PageShell>
    <section class="hero">
      <div class="container hero-inner">
        <div class="hero-copy">
          <span class="eyebrow">Tree 家脉亲缘</span>
          <h1>
            <span>把家人关系整理清楚</span>
            <span>把家族记忆长期保存</span>
          </h1>
          <p class="hero-lead">
            Tree 用家庭空间管理成员、关系、邀请与公开展示。私密资料留在家庭内部，公开主页经过审核后再对外展示。
          </p>
          <div class="hero-tags">
            <span>私密协作</span>
            <span>审核公开</span>
            <span>家谱关系</span>
          </div>
          <div class="actions">
            <RouterLink class="button" to="/me/families">进入我的家庭</RouterLink>
            <RouterLink class="button secondary" to="/families">寻找公开家族</RouterLink>
          </div>
        </div>

        <div class="hero-panel">
          <div class="relation-visual" aria-hidden="true">
            <span class="orbit"></span>
            <span class="line line-a"></span>
            <span class="line line-b"></span>
            <span class="line line-c"></span>
            <span class="node node-core"></span>
            <span class="node node-top"></span>
            <span class="node node-left"></span>
            <span class="node node-right"></span>
            <span class="node node-bottom"></span>
          </div>
          <div class="panel-copy">
            <strong>家庭空间</strong>
            <p>先整理成员和关系，再邀请亲人一起补充、核对和维护。</p>
          </div>
          <div class="panel-steps">
            <span>成员资料</span>
            <span>亲缘关系</span>
            <span>公开展示</span>
          </div>
        </div>
      </div>
    </section>

    <section class="container main-actions" aria-label="常用入口">
      <RouterLink class="action-card action-primary" to="/me/families">
        <span class="action-icon home-icon"></span>
        <strong>我的家庭</strong>
        <small>查看成员、关系和私有家谱</small>
      </RouterLink>
      <RouterLink class="action-card" to="/families">
        <span class="action-icon search-icon"></span>
        <strong>寻找家族</strong>
        <small>浏览公开主页与公开树</small>
      </RouterLink>
      <RouterLink class="action-card" to="/me/invitations">
        <span class="action-icon invite-icon"></span>
        <strong>我的邀请</strong>
        <small>处理收到的家庭邀请</small>
      </RouterLink>
      <RouterLink class="action-card" to="/me/join-requests">
        <span class="action-icon request-icon"></span>
        <strong>加入申请</strong>
        <small>查看申请处理进度</small>
      </RouterLink>
    </section>

    <section id="reading" class="container content-section">
      <div class="section-head">
        <span class="eyebrow">阅读精选</span>
        <div>
          <h2>故事、典故与使用指南</h2>
          <p>用教程快速上手，用故事和典故理解家族记录的价值。</p>
        </div>
      </div>
      <div class="reading-layout">
        <RouterLink v-if="featuredArticle" class="featured-article" :to="`/content/${featuredArticle.id}`">
          <span class="article-label">{{ featuredArticle.categoryName }}</span>
          <h3>{{ featuredArticle.title }}</h3>
          <p>{{ featuredArticle.summary || '阅读全文，了解家庭记录与协作方法。' }}</p>
        </RouterLink>
        <div v-else class="featured-article empty-content">{{ contentError || '暂无精选内容' }}</div>
        <div class="article-grid">
          <RouterLink v-for="item in readingCards" :key="item.id" class="article-card" :class="contentTone(item.categoryKey)" :to="`/content/${item.id}`">
            <span>{{ item.categoryName }}</span>
            <h3>{{ item.title }}</h3>
            <p>{{ item.summary || '阅读全文' }}</p>
          </RouterLink>
        </div>
      </div>
    </section>

    <section class="container section">
      <div class="section-head">
        <span class="eyebrow">公开家族</span>
        <div>
          <h2>看看别人如何展示家族主页</h2>
          <p>经审核后对外展示家族简介、公开树和已审核留言。</p>
        </div>
      </div>
      <div class="grid families">
        <FamilyCard v-for="family in homeFamilyCards" :key="family.id" :family="family" />
        <p v-if="familyError" class="section-state">{{ familyError }}</p>
        <p v-else-if="!familiesLoading && publicFamilies.length === 0" class="section-state">暂无已公开家庭</p>
      </div>
    </section>

    <section class="container workflow-section">
      <div class="workflow-copy">
        <span class="eyebrow">使用流程</span>
        <h2>从核心成员开始，逐步整理完整家谱。</h2>
        <p>不需要一次性录完全部资料。先建立家庭，再邀请亲人共同补充，最后按需申请公开展示。</p>
      </div>
      <div class="workflow-list">
        <div v-for="step in workflowSteps" :key="step.title" class="workflow-item">
          <span>{{ step.no }}</span>
          <div>
            <strong>{{ step.title }}</strong>
            <small>{{ step.desc }}</small>
          </div>
        </div>
      </div>
    </section>
  </PageShell>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'

import { listContentArticles } from '@/api/content'
import { listPublicFamilies } from '@/api/families'
import FamilyCard from '@/components/FamilyCard.vue'
import PageShell from '@/components/PageShell.vue'
import type { ContentArticleSummary, PublicFamilyListItem } from '@/types/api'

const publicFamilies = ref<PublicFamilyListItem[]>([])
const featuredArticle = ref<ContentArticleSummary | null>(null)
const readingCards = ref<ContentArticleSummary[]>([])
const familyError = ref('')
const contentError = ref('')
const familiesLoading = ref(true)
const homeFamilyCards = computed(() => publicFamilies.value.map((family) => ({
  id: String(family.id),
  name: family.familyName,
  surname: family.familySurname,
  nativePlace: family.nativePlace || '未填写',
  regionText: family.regionText || '未填写',
  description: family.description || '该家庭暂未填写简介。',
  publicContact: family.publicContactVisible
    ? [family.publicContactName, family.publicContactNote].filter(Boolean).join(' / ')
    : '',
  publicStatus: 'APPROVED' as const
})))

function contentTone(categoryKey: string) {
  if (categoryKey === 'story') return 'story'
  if (categoryKey === 'surname') return 'surname'
  return 'article'
}

async function loadHomeData() {
  familiesLoading.value = true
  const [familiesResult, contentResult] = await Promise.allSettled([
    listPublicFamilies({ page: 1, pageSize: 2 }),
    listContentArticles({ featured: true, page: 1, pageSize: 4 })
  ])
  if (familiesResult.status === 'fulfilled') {
    publicFamilies.value = familiesResult.value.items
  } else {
    familyError.value = '公开家庭暂时无法加载，请稍后重试。'
  }
  if (contentResult.status === 'fulfilled') {
    featuredArticle.value = contentResult.value.items[0] || null
    readingCards.value = contentResult.value.items.slice(1, 4)
  } else {
    contentError.value = '精选内容暂时无法加载，请稍后重试。'
  }
  familiesLoading.value = false
}

onMounted(loadHomeData)

const workflowSteps = [
  { no: '01', title: '创建家庭', desc: '填写姓氏、家庭名称和基础介绍' },
  { no: '02', title: '补充成员', desc: '录入成员资料，维护父母子女与配偶关系' },
  { no: '03', title: '邀请亲人', desc: '通过邀请链接让家人一起核对补充' },
  { no: '04', title: '申请公开', desc: '审核通过后展示公开主页与公开树' }
]
</script>

<style scoped>
.hero {
  position: relative;
  padding: 88px 0 44px;
  overflow: hidden;
}

.hero::before {
  content: '';
  position: absolute;
  inset: 28px max(20px, calc((100vw - 1180px) / 2)) 0;
  border: 1px solid rgba(255, 255, 255, 0.18);
  border-radius: 42px;
  background:
    radial-gradient(circle at 88% 12%, rgba(200, 164, 93, 0.24), transparent 280px),
    radial-gradient(circle at 12% 88%, rgba(47, 107, 87, 0.30), transparent 320px),
    linear-gradient(135deg, #172f4a 0%, #244f54 58%, #2f6b57 100%);
  box-shadow: 0 34px 90px rgba(31, 58, 95, 0.22);
}

.hero-inner {
  display: grid;
  grid-template-columns: minmax(0, 1.35fr) minmax(320px, 0.9fr);
  gap: 32px;
  align-items: center;
}

.hero-copy {
  position: relative;
  z-index: 1;
}

.eyebrow {
  display: inline-flex;
  border: 1px solid rgba(255, 255, 255, 0.18);
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.11);
  color: rgba(248, 231, 194, 0.92);
  padding: 7px 14px;
  font-size: 13px;
  font-weight: 800;
  letter-spacing: 0.08em;
}

h1 {
  max-width: 780px;
  margin: 18px 0 16px;
  color: #fff;
  font-size: clamp(36px, 4.7vw, 62px);
  line-height: 1.04;
  letter-spacing: -0.03em;
}

h1 span {
  display: block;
}

.hero-lead {
  max-width: 680px;
  margin: 0;
  color: rgba(255, 255, 255, 0.76);
  font-size: 18px;
  line-height: 1.85;
}

.hero-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-top: 22px;
}

.hero-tags span {
  border-radius: 999px;
  border: 1px solid rgba(255, 255, 255, 0.14);
  background: rgba(255, 255, 255, 0.10);
  color: rgba(255, 255, 255, 0.88);
  padding: 7px 12px;
  font-size: 13px;
  font-weight: 800;
}

.hero-tags span:nth-child(2) {
  background: rgba(255, 255, 255, 0.10);
  color: rgba(255, 255, 255, 0.88);
}

.hero-tags span:nth-child(3) {
  background: rgba(255, 255, 255, 0.10);
  color: rgba(255, 255, 255, 0.88);
}

.actions {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  margin-top: 28px;
}

.hero-panel {
  position: relative;
  min-height: 430px;
  border: 1px solid rgba(255, 255, 255, 0.18);
  border-radius: var(--radius-card-xl);
  background:
    radial-gradient(circle at 78% 18%, rgba(200, 164, 93, 0.20), transparent 170px),
    rgba(255, 255, 255, 0.10);
  padding: 32px;
  overflow: hidden;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.18),
    0 24px 58px rgba(0, 0, 0, 0.12);
}

.relation-visual {
  position: relative;
  width: 280px;
  height: 230px;
  margin: 8px auto 26px;
}

.relation-visual span {
  position: absolute;
  display: block;
}

.orbit {
  inset: 8px 20px 0 20px;
  border: 1px dashed rgba(255, 255, 255, 0.24);
  border-radius: 50%;
}

.line {
  height: 3px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.28);
  transform-origin: center;
}

.line-a {
  top: 76px;
  left: 138px;
  width: 56px;
  transform: rotate(90deg);
}

.line-b {
  top: 136px;
  left: 78px;
  width: 92px;
  transform: rotate(-28deg);
}

.line-c {
  top: 140px;
  right: 62px;
  width: 86px;
  transform: rotate(28deg);
}

.node {
  border-radius: 50%;
  background: #fff;
  box-sizing: border-box;
}

.node-core {
  top: 92px;
  left: 116px;
  width: 54px;
  height: 54px;
  border: 5px solid rgba(255, 255, 255, 0.96);
  box-shadow: 0 0 0 13px rgba(255, 255, 255, 0.12);
}

.node-top {
  top: 28px;
  left: 134px;
  width: 22px;
  height: 22px;
  border: 4px solid rgba(248, 231, 194, 0.95);
}

.node-left,
.node-right,
.node-bottom {
  width: 25px;
  height: 25px;
  border: 4px solid var(--color-heritage-green);
}

.node-left {
  left: 52px;
  top: 158px;
  border-color: rgba(160, 197, 231, 0.95);
}

.node-right {
  right: 48px;
  top: 148px;
  border-color: rgba(167, 216, 198, 0.95);
}

.node-bottom {
  left: 134px;
  bottom: 12px;
  border-color: rgba(216, 175, 104, 0.95);
}

.panel-copy {
  position: relative;
  z-index: 1;
  border: 1px solid rgba(255, 255, 255, 0.18);
  border-radius: 24px;
  background: rgba(255, 255, 255, 0.12);
  padding: 20px;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.8);
}

.panel-copy strong {
  color: #fff;
  font-size: 24px;
}

.panel-copy p {
  margin: 8px 0 0;
  color: rgba(255, 255, 255, 0.72);
  line-height: 1.75;
}

.panel-steps {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 10px;
  margin-top: 14px;
}

.panel-steps span {
  border-radius: 14px;
  border: 1px solid rgba(255, 255, 255, 0.14);
  background: rgba(255, 255, 255, 0.10);
  color: rgba(255, 255, 255, 0.84);
  padding: 12px;
  text-align: center;
  font-size: 13px;
  font-weight: 800;
}

.main-actions {
  display: grid;
  grid-template-columns: 1.18fr repeat(3, 1fr);
  gap: 16px;
  padding: 14px 0 48px;
}

.action-card {
  position: relative;
  min-height: 164px;
  border: 1px solid rgba(148, 163, 184, 0.14);
  border-radius: 28px;
  background:
    radial-gradient(circle at 100% 0%, rgba(47, 107, 87, 0.06), transparent 130px),
    rgba(255, 255, 255, 0.90);
  padding: 24px;
  overflow: hidden;
  box-shadow: 0 14px 40px rgba(31, 58, 95, 0.06);
  transition: transform 0.18s ease, box-shadow 0.18s ease;
}

.action-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 18px 46px rgba(31, 58, 95, 0.09);
}

.action-primary {
  background: linear-gradient(135deg, #1f3a5f 0%, #28605b 100%);
}

.action-card strong {
  display: block;
  margin-top: 18px;
  color: var(--color-primary);
  font-size: 22px;
  letter-spacing: -0.02em;
}

.action-primary strong {
  color: #fff;
}

.action-card small {
  display: block;
  margin-top: 8px;
  color: var(--color-text-secondary);
  line-height: 1.55;
}

.action-primary small {
  color: rgba(255, 255, 255, 0.76);
}

.action-icon {
  display: block;
  position: relative;
  width: 46px;
  height: 46px;
  border-radius: 16px;
  background: rgba(47, 107, 87, 0.1);
}

.action-primary .action-icon {
  background: rgba(255, 255, 255, 0.16);
}

.action-icon::before,
.action-icon::after {
  content: '';
  position: absolute;
  box-sizing: border-box;
}

.home-icon::before {
  left: 13px;
  top: 14px;
  width: 20px;
  height: 17px;
  border: 2px solid rgba(255, 255, 255, 0.76);
  border-radius: 4px;
}

.search-icon::before {
  left: 12px;
  top: 12px;
  width: 18px;
  height: 18px;
  border: 3px solid var(--color-heritage-green);
  border-radius: 50%;
}

.search-icon::after {
  left: 28px;
  top: 30px;
  width: 12px;
  height: 3px;
  border-radius: 999px;
  background: var(--color-heritage-green);
  transform: rotate(45deg);
}

.invite-icon::before,
.request-icon::before {
  left: 10px;
  top: 12px;
  width: 26px;
  height: 20px;
  border: 2px solid var(--color-primary);
  border-radius: 5px;
}

.request-icon::after {
  left: 16px;
  top: 18px;
  width: 14px;
  height: 2px;
  background: var(--color-primary);
  box-shadow: 0 6px 0 var(--color-primary);
}

.section,
.content-section,
.workflow-section {
  padding: 28px 0 58px;
}

.section-head {
  display: flex;
  align-items: end;
  justify-content: space-between;
  gap: 28px;
  margin-bottom: 22px;
}

.section-head h2,
.workflow-copy h2 {
  margin: 8px 0 6px;
  color: var(--color-primary);
  font-size: clamp(26px, 3vw, 38px);
  line-height: 1.16;
  letter-spacing: -0.02em;
}

.section-head p,
.workflow-copy p {
  margin: 0;
  color: var(--color-text-secondary);
  line-height: 1.75;
}

.reading-layout {
  display: grid;
  grid-template-columns: minmax(300px, 0.82fr) 1.18fr;
  gap: 16px;
}

.featured-article {
  position: relative;
  display: block;
  min-height: 322px;
  border-radius: 32px;
  background: linear-gradient(142deg, rgba(31, 58, 95, 0.97) 0%, rgba(47, 107, 87, 0.92) 100%);
  padding: 28px;
  overflow: hidden;
  color: inherit;
  text-decoration: none;
  box-shadow: 0 20px 54px rgba(31, 58, 95, 0.13);
}

.featured-article::after {
  content: '';
  position: absolute;
  right: -56px;
  bottom: -72px;
  width: 220px;
  height: 220px;
  border: 1px solid rgba(255, 255, 255, 0.16);
  border-radius: 50%;
}

.article-label {
  display: inline-flex;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.13);
  color: rgba(255, 255, 255, 0.82);
  padding: 7px 13px;
  font-size: 13px;
  font-weight: 800;
}

.featured-article h3 {
  position: relative;
  z-index: 1;
  margin: 74px 0 12px;
  color: #fff;
  font-size: 32px;
  line-height: 1.25;
}

.featured-article p {
  position: relative;
  z-index: 1;
  margin: 0;
  color: rgba(255, 255, 255, 0.76);
  line-height: 1.8;
}

.article-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 14px;
}

.article-card {
  display: block;
  min-height: 190px;
  border: 1px solid rgba(148, 163, 184, 0.12);
  border-radius: 28px;
  padding: 22px;
  color: inherit;
  text-decoration: none;
  box-shadow: 0 10px 26px rgba(31, 58, 95, 0.04);
  transition:
    transform 0.18s ease,
    box-shadow 0.18s ease;
}

.article-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 18px 40px rgba(31, 58, 95, 0.08);
}

.article-card.story {
  background: #fff2f7;
}

.article-card.surname {
  background: #fff8e8;
}

.article-card.article {
  background: #edf8f2;
}

.article-card span {
  color: var(--color-text-secondary);
  font-size: 13px;
  font-weight: 800;
}

.article-card h3 {
  margin: 18px 0 10px;
  color: var(--color-primary);
  font-size: 21px;
  line-height: 1.35;
}

.article-card p {
  margin: 0;
  color: var(--color-text-secondary);
  line-height: 1.65;
}

.families {
  grid-template-columns: repeat(2, 1fr);
}

.workflow-section {
  display: grid;
  grid-template-columns: minmax(0, 0.82fr) minmax(320px, 1.18fr);
  gap: 30px;
  align-items: start;
}

.workflow-list {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 14px;
}

.workflow-item {
  display: flex;
  gap: 14px;
  border: 1px solid rgba(148, 163, 184, 0.14);
  border-radius: 24px;
  background: rgba(255, 255, 255, 0.88);
  padding: 20px;
  box-shadow: 0 12px 34px rgba(31, 58, 95, 0.05);
}

.workflow-item span {
  display: grid;
  flex: 0 0 42px;
  width: 42px;
  height: 42px;
  place-items: center;
  border-radius: 14px;
  background: rgba(31, 58, 95, 0.08);
  color: var(--color-primary);
  font-size: 13px;
  font-weight: 900;
}

.workflow-item strong,
.workflow-item small {
  display: block;
}

.workflow-item strong {
  color: var(--color-primary);
  font-size: 18px;
}

.workflow-item small {
  margin-top: 6px;
  color: var(--color-text-secondary);
  line-height: 1.55;
}

@media (max-width: 980px) {
  .hero-inner,
  .reading-layout,
  .workflow-section {
    grid-template-columns: 1fr;
  }

  .main-actions,
  .article-grid,
  .families,
  .workflow-list {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 680px) {
  .hero {
    padding: 30px 12px 36px;
  }

  .hero::before {
    inset: 0 12px 0;
    border-radius: 32px;
  }

  .hero-inner {
    gap: 24px;
  }

  .hero-copy {
    padding: 0 2px;
  }

  h1 {
    font-size: clamp(30px, 9vw, 36px);
    line-height: 1.12;
    letter-spacing: -0.035em;
  }

  .hero-panel {
    min-height: auto;
    padding: 24px 18px;
  }

  .relation-visual {
    width: 240px;
    transform: scale(0.9);
    transform-origin: center top;
    margin-bottom: 0;
  }

  .main-actions,
  .article-grid,
  .families,
  .workflow-list,
  .panel-steps {
    grid-template-columns: 1fr;
  }

  .section-head {
    align-items: start;
    flex-direction: column;
  }
}
</style>
