<template>
  <view class="archive-page historical-showcase-page">
    <view class="archive-page-head">
      <view>
        <text class="archive-kicker">Imperial Lineage</text>
        <text class="archive-title">帝王世家</text>
        <text class="archive-subtitle">历史示例，按后台发布内容展示，仅供浏览</text>
      </view>
      <view class="archive-seal">谱</view>
    </view>

    <view v-if="loading && items.length === 0" class="archive-form-panel page-state"><MiniEmptyState symbol="谱" title="正在加载" description="正在读取展示家谱，请稍候。" /></view>
    <view v-else-if="errorMessage && items.length === 0" class="archive-form-panel page-state"><MiniNotice tone="warm" title="加载失败">{{ errorMessage }}</MiniNotice><MiniButton variant="secondary" @click="load">重新加载</MiniButton></view>

    <template v-else>
      <view class="historical-list archive-panel">
        <text class="historical-list-label">已发布世系</text>
        <view v-for="item in items" :key="item.id" class="historical-row" :class="{ active: item.id === selectedId }" @click="selectFamily(item.id)">
          <view><text class="historical-name">{{ item.displayName }}</text><text class="historical-meta">{{ item.regionText || '历史示例' }} · {{ item.memberCount }} 位人物</text></view>
          <text class="archive-arrow">›</text>
        </view>
      </view>

      <view v-if="selected" class="historical-intro archive-panel"><text class="historical-label">历史示例</text><text class="historical-summary">{{ selected.summary }}</text><text class="historical-count">{{ selected.memberCount }} 位人物 · {{ selected.relationshipCount }} 条关系</text></view>
      <view v-if="treeLoading" class="archive-form-panel page-state"><MiniEmptyState symbol="…" title="正在展开家谱" description="正在整理人物关系。" /></view>
      <view v-else-if="tree" class="historical-tree"><view class="tree-note"><text>可左右、上下滑动查看完整世系；人物仅展示姓名与身份。</text></view><FamilyTreeStructureView :tree="tree" /></view>
    </template>
  </view>
</template>

<script setup lang="ts">
import { onLoad } from '@dcloudio/uni-app'
import { computed, ref } from 'vue'
import { apiErrorMessage } from '@/api/client'
import { getPublishedShowcaseFamilyTree, listPublishedShowcaseFamilies } from '@/api/showcaseFamilies'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
import FamilyTreeStructureView from '@/components/family/FamilyTreeStructureView.vue'
import type { FamilyTreeResult, ShowcaseFamilyItem } from '@/types/api'

const items = ref<ShowcaseFamilyItem[]>([])
const selectedId = ref<number | null>(null)
const tree = ref<FamilyTreeResult | null>(null)
const loading = ref(false)
const treeLoading = ref(false)
const errorMessage = ref('')
const selected = computed(() => items.value.find((item) => item.id === selectedId.value) || null)

async function selectFamily(id: number) {
  selectedId.value = id
  tree.value = null
  treeLoading.value = true
  try { tree.value = await getPublishedShowcaseFamilyTree(id) } catch (error) { errorMessage.value = apiErrorMessage(error, '展示家谱暂不可访问。') } finally { treeLoading.value = false }
}

async function load(preferredId?: number) {
  loading.value = true; errorMessage.value = ''
  try { items.value = await listPublishedShowcaseFamilies(); const next = items.value.find((item) => item.id === preferredId)?.id || items.value.find((item) => item.rotationPinned)?.id || items.value[0]?.id; if (next) await selectFamily(next) } catch (error) { errorMessage.value = apiErrorMessage(error, '展示家谱列表加载失败。') } finally { loading.value = false }
}

onLoad((options) => { const id = Number(options?.familyId); void load(Number.isFinite(id) && id > 0 ? id : undefined) })
</script>

<style scoped>
.historical-showcase-page { padding-top: 28rpx; padding-bottom: calc(80rpx + env(safe-area-inset-bottom)); }
.page-state { margin-top: 24rpx; }
.historical-list, .historical-intro, .historical-tree { margin-top: 20rpx; }
.historical-list { padding: 18rpx 0; }
.historical-list-label, .historical-label { display: block; color: var(--archive-cinnabar); font-size: 20rpx; font-weight: 700; letter-spacing: 3rpx; }
.historical-row { display: flex; align-items: center; justify-content: space-between; gap: 14rpx; border-top: 1rpx solid var(--archive-line); padding: 18rpx 0; }
.historical-row:first-of-type { margin-top: 12rpx; }
.historical-row.active .historical-name { color: var(--archive-cinnabar); }
.historical-name, .historical-meta, .historical-summary, .historical-count { display: block; }
.historical-name { color: var(--archive-ink); font-size: 27rpx; font-weight: 700; }
.historical-meta, .historical-count { margin-top: 5rpx; color: var(--archive-ink-soft); font-size: 21rpx; }
.historical-intro { padding: 20rpx 0; }
.historical-summary { margin-top: 12rpx; color: var(--archive-ink); font-size: 25rpx; line-height: 1.7; }
.tree-note { margin-bottom: 12rpx; color: var(--archive-ink-soft); font-size: 21rpx; line-height: 1.5; }
</style>
