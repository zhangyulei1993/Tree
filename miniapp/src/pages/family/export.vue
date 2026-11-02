<template>
  <view class="archive-page export-page">
    <MiniBackHome :return-url="returnUrl" :return-label="returnLabel" />
    <view class="archive-page-head">
      <view>
        <text class="archive-kicker">Family Scroll</text>
        <text class="archive-title">家谱长卷</text>
        <text class="archive-subtitle">先确定成员范围，再选择留存、分享或公开用途。</text>
      </view>
    </view>

    <view class="selection-section">
      <text class="selection-label">成员范围</text>
      <view class="scope-list">
        <view
          v-for="item in memberModes"
          :key="item.key"
          class="scope-item"
          :class="{ active: memberMode === item.key }"
          @tap="memberMode = item.key"
        >
          <text class="scope-mark">{{ item.mark }}</text>
          <view>
            <text class="scope-title">{{ item.title }}</text>
            <text class="scope-desc">{{ item.desc }}</text>
          </view>
        </view>
      </view>
    </view>

    <view class="selection-section">
      <text class="selection-label">导出用途</text>
      <view class="scope-list">
        <view
          v-for="item in scopes"
          :key="item.key"
          class="scope-item"
          :class="{ active: scope === item.key }"
          @tap="scope = item.key"
        >
          <text class="scope-mark">{{ item.mark }}</text>
          <view>
            <text class="scope-title">{{ item.title }}</text>
            <text class="scope-desc">{{ item.desc }}</text>
          </view>
        </view>
      </view>
    </view>

    <MiniNotice v-if="blockingMessage" tone="warm" title="暂不能生成长卷">{{ blockingMessage }}</MiniNotice>
    <MiniButton :loading="loading" @click="generate">生成预览</MiniButton>

    <view v-if="archive?.quality?.incompleteMemberCount" class="quality-warning">
      当前有 {{ archive.quality.incompleteMemberCount }} 位成员缺少{{ archive.quality.missingFields.join('、') }}；分享长卷会标记为资料待完善，公开展示暂不可生成。
    </view>

    <view v-if="archive && tree && lineage" class="scroll-preview">
      <view class="scroll-top"><text>{{ archive.family.familySurname }}氏家谱</text><text>{{ memberModeTitle }}</text></view>
      <view class="scroll-title"><text>{{ archive.family.familyName }}</text><text>{{ archive.family.regionText || archive.family.nativePlace || '家脉留存' }}</text></view>
      <view class="scroll-stats">
        <view><b>{{ lineage.visibleMemberCount }}</b><text>导出成员</text></view>
        <view><b>{{ tree.edges.length }}</b><text>亲属关系</text></view>
        <view><b>{{ scopeTitle }}</b><text>导出用途</text></view>
      </view>
      <view class="lineage-explain"><text>{{ memberModeTitle }}</text><text>{{ memberModeExplanation }}</text></view>
      <view class="lineage-explain usage-explain"><text>{{ scopeTitle }}</text><text>{{ scopeExplanation }}</text></view>
      <FamilyExportLineageTree :tree="tree" :scope="scope" :member-mode="memberMode" />
      <view class="scroll-foot"><text>Tree · 家谱记录与家人协作</text><text>{{ date }}</text></view>
    </view>

    <view v-if="archive && tree && lineage" class="export-actions">
      <MiniButton :loading="exporting" @click="createImage">生成家谱长卷图片</MiniButton>
      <MiniButton variant="secondary" :disabled="!posterPaths.length" @click="saveImages">保存到相册</MiniButton>
      <MiniButton variant="secondary" :disabled="!posterPaths.length" @click="shareImage">分享图片</MiniButton>
    </view>

    <canvas
      v-if="posterCanvasHeight"
      canvas-id="archivePosterCanvas"
      id="archivePosterCanvas"
      class="export-canvas"
      :style="{ width: `${posterCanvasWidth}px`, height: `${posterCanvasHeight}px` }"
      :width="posterCanvasWidth"
      :height="posterCanvasHeight"
    />
  </view>
</template>

<script setup lang="ts">
import { computed, getCurrentInstance, nextTick, ref, watch } from 'vue'
import { onLoad } from '@dcloudio/uni-app'

import { apiErrorMessage } from '@/api/client'
import { exportFamilyArchive, getFamilyDetail, type FamilyArchiveExport } from '@/api/families'
import { listFamilyMembers } from '@/api/members'
import { getPrivateTree } from '@/api/tree'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import FamilyExportLineageTree from '@/components/family/FamilyExportLineageTree.vue'
import { buildExportLineagePosterPages, drawExportLineagePoster } from '@/features/family-tree/exportLineagePoster'
import { buildExportLineageTree, type ExportLineageMode } from '@/features/family-tree/exportLineageTree'
import { calculateFamilyHealthScore, summarizeFamilyHealth } from '@/features/family-tree/familyHealthSummary'
import type { FamilyTreeResult } from '@/types/api'

type Scope = FamilyArchiveExport['scope']

const familyId = ref('')
const returnTo = ref('')
const scope = ref<Scope>('SHARE')
const memberMode = ref<ExportLineageMode>('SURNAME')
const archive = ref<FamilyArchiveExport | null>(null)
const tree = ref<FamilyTreeResult | null>(null)
const loading = ref(false)
const exporting = ref(false)
const posterPaths = ref<string[]>([])
const posterCanvasWidth = ref(750)
const posterCanvasHeight = ref(0)
const blockingMessage = ref('')
const instance = getCurrentInstance()
const exportScoreThreshold = 90
const returnUrl = computed(() => returnTo.value === 'tree'
  ? `/pages/family/private-tree?familyId=${encodeURIComponent(familyId.value)}`
  : `/pages/family/health?familyId=${encodeURIComponent(familyId.value)}${returnTo.value === 'tree' ? '&returnTo=tree' : ''}`)
const returnLabel = computed(() => returnTo.value === 'tree' ? '返回家庭树' : '返回家庭健康')

const memberModes: Array<{ key: ExportLineageMode; mark: string; title: string; desc: string }> = [
  { key: 'SURNAME', mark: '宗', title: '姓氏成员', desc: '按本姓宗支展开；女性成员仅延伸下一代，不显示配偶。' },
  { key: 'FAMILY', mark: '家', title: '家庭成员', desc: '不区分男女成员，所有本家成员均按完整后代分支导出。' }
]
const scopes: Array<{ key: Scope; mark: string; title: string; desc: string }> = [
  { key: 'INTERNAL', mark: '藏', title: '家庭留存', desc: '保留成员年代与前任、现任婚姻注记，适合内部归档。' },
  { key: 'SHARE', mark: '赠', title: '分享亲友', desc: '隐藏前任婚姻，仅保留现任配偶，适合亲友留念。' },
  { key: 'PUBLIC', mark: '展', title: '公开展示', desc: '隐藏前任婚姻并脱敏现任配偶，适合展示家庭脉络。' }
]

const scopeTitle = computed(() => scopes.find((item) => item.key === scope.value)?.title || '家庭留存')
const memberModeTitle = computed(() => memberModes.find((item) => item.key === memberMode.value)?.title || '姓氏成员')
const date = computed(() => archive.value ? new Date(archive.value.generatedAt).toLocaleDateString('zh-CN') : '')
const lineage = computed(() => tree.value ? buildExportLineageTree(tree.value, memberMode.value) : null)
const memberModeExplanation = computed(() => memberMode.value === 'SURNAME'
  ? '以男性姓氏成员为主；女性姓氏成员保留本人及其下一代，后续分支不继续展开。仅显示本姓父母归属，不显示妻、夫或族外配偶。'
  : '所有本家成员均按完整后代分支展开，不因成员性别截断分支；族外配偶仅以妻／夫注记。')
const scopeExplanation = computed(() => memberMode.value === 'SURNAME'
  ? scope.value === 'PUBLIC'
    ? '姓名与本姓父母归属仅保留姓氏加单个星号；不展示年份、配偶与前任婚姻。'
    : '仅为已离世成员显示生卒年份，并保留本姓父母归属；不展示妻、夫、前任或现任配偶。'
  : scope.value === 'INTERNAL'
  ? '仅为已离世成员显示生卒年份，并保留实际父母来源、前任与现任族外配偶注记，仅供家庭内部留存。'
  : scope.value === 'SHARE'
    ? '仅为已离世成员显示生卒年份，并保留世代、实际父母来源与现任配偶注记；前任婚姻不展示。'
    : '姓名与父母、现任配偶仅保留姓氏加单个星号；不展示年份与前任婚姻。')

watch([scope, memberMode], () => {
  posterPaths.value = []
  posterCanvasWidth.value = 750
  posterCanvasHeight.value = 0
})

async function generate() {
  loading.value = true
  blockingMessage.value = ''
  archive.value = null
  tree.value = null
  posterPaths.value = []
  posterCanvasWidth.value = 750
  posterCanvasHeight.value = 0
  try {
    const [currentTree, currentFamily, currentMembers] = await Promise.all([
      getPrivateTree(familyId.value),
      getFamilyDetail(familyId.value),
      listFamilyMembers(familyId.value)
    ])
    const healthScore = calculateFamilyHealthScore(summarizeFamilyHealth(currentFamily, currentMembers, currentTree))
    if (healthScore < exportScoreThreshold) {
      blockingMessage.value = `当前家庭健康度为 ${healthScore} 分，达到 ${exportScoreThreshold} 分后才能生成长卷。请先返回家庭健康处理待补充事项。`
      return
    }
    archive.value = {
      scope: scope.value,
      generatedAt: new Date().toISOString(),
      family: currentFamily,
      members: currentMembers.map((member) => ({ memberId: member.memberId, name: member.name, gender: member.gender, birthYear: member.birthYear, deathYear: member.deathYear })),
      relationships: currentTree.edges.map((edge) => ({ fromMemberId: edge.fromMemberId, toMemberId: edge.toMemberId, relationshipType: edge.relationshipType })),
      quality: { incompleteMemberCount: healthSummaryIncompleteCount(currentFamily, currentMembers, currentTree), missingFields: healthSummaryMissingFields(currentFamily, currentMembers, currentTree) }
    }
    tree.value = currentTree
  } catch (error) {
    uni.showModal({ title: '生成失败', content: apiErrorMessage(error, '无法生成家谱长卷。'), showCancel: false })
  } finally {
    loading.value = false
  }
}

function canvasToTempFilePath(width: number, height: number): Promise<string> {
  return new Promise((resolve, reject) => {
    uni.canvasToTempFilePath({
      canvasId: 'archivePosterCanvas', x: 0, y: 0, width, height,
      destWidth: width, destHeight: height, fileType: 'png',
      success: (result) => resolve(result.tempFilePath), fail: reject
    }, instance?.proxy)
  })
}

async function createImage() {
  if (!archive.value || !lineage.value) return
  exporting.value = true
  try {
    const paths: string[] = []
    const pages = buildExportLineagePosterPages(lineage.value.generations, `${archive.value.family.familyName}家谱`, `${memberModeTitle.value} · ${scopeTitle.value}`, scope.value)
    for (const page of pages) {
      posterCanvasWidth.value = page.width
      posterCanvasHeight.value = page.height
      await nextTick()
      const context = uni.createCanvasContext('archivePosterCanvas', instance?.proxy)
      drawExportLineagePoster(context, page)
      await new Promise<void>((resolve) => context.draw(false, resolve))
      paths.push(await canvasToTempFilePath(page.width, page.height))
    }
    posterPaths.value = paths
    let auditSynced = true
    try {
      await exportFamilyArchive(familyId.value, scope.value)
    } catch {
      auditSynced = false
    }
    uni.showModal({
      title: '图片已生成',
      content: `${paths.length > 1 ? `已生成 ${paths.length} 张家谱长卷图片，可全部保存到相册。` : '已生成家谱长卷，可保存或分享。'}${auditSynced ? '' : ' 导出记录暂未同步，不影响图片使用。'}`,
      showCancel: false
    })
  } catch (error) {
    uni.showModal({ title: '生成失败', content: apiErrorMessage(error, '图片生成失败，请稍后重试。'), showCancel: false })
  } finally {
    exporting.value = false
  }
}

function healthSummaryIncompleteCount(family: Awaited<ReturnType<typeof getFamilyDetail>>, members: Awaited<ReturnType<typeof listFamilyMembers>>, treeResult: FamilyTreeResult) {
  return summarizeFamilyHealth(family, members, treeResult).total
}

function healthSummaryMissingFields(family: Awaited<ReturnType<typeof getFamilyDetail>>, members: Awaited<ReturnType<typeof listFamilyMembers>>, treeResult: FamilyTreeResult) {
  const summary = summarizeFamilyHealth(family, members, treeResult)
  const fields: string[] = []
  if (summary.unknownGender) fields.push('性别')
  if (summary.missingBirth) fields.push('出生年份')
  if (summary.isolated || summary.missingSpouse || summary.birthOrderConflict) fields.push('亲属关系')
  if (summary.requiredUnbound) fields.push('账号绑定')
  if (summary.missingPlace) fields.push('地区')
  return fields
}

async function saveImages() {
  if (!posterPaths.value.length) return
  try {
    for (const filePath of posterPaths.value) {
      await new Promise<void>((resolve, reject) => uni.saveImageToPhotosAlbum({ filePath, success: () => resolve(), fail: reject }))
    }
    uni.showToast({ title: '已保存到相册', icon: 'success' })
  } catch {
    uni.showModal({ title: '无法保存图片', content: '请允许小程序保存到相册后重试。', showCancel: false })
  }
}

function shareImage() {
  if (!posterPaths.value.length) return
  const wxApi = (globalThis as typeof globalThis & { wx?: { showShareImageMenu?: (options: { path: string }) => void } }).wx
  if (wxApi?.showShareImageMenu) {
    wxApi.showShareImageMenu({ path: posterPaths.value[0]! })
    return
  }
  uni.previewImage({ urls: posterPaths.value, current: posterPaths.value[0] })
}

onLoad((options) => { familyId.value = String(options?.familyId || ''); returnTo.value = String(options?.returnTo || '') })
</script>

<style scoped>
.export-page { padding-top: 28rpx; }
.tree-poster-preview { margin-top: 24rpx; overflow: hidden; border-top: 1rpx solid #c9a56b; border-bottom: 1rpx solid #c9a56b; }
.tree-poster-preview-title, .tree-poster-preview-desc { display: block; }
.tree-poster-preview-title { padding: 22rpx 22rpx 0; color: #193856; font-size: 27rpx; font-weight: 700; }
.tree-poster-preview-desc { padding: 8rpx 22rpx 16rpx; color: #718096; font-size: 22rpx; line-height: 1.5; }
.tree-poster-preview-tree { max-height: 740rpx; }
.selection-section { margin-top: 26rpx; }
.selection-label { display: block; color: var(--archive-ink); font-size: 26rpx; font-weight: 700; }
.scope-list { margin-top: 14rpx; }
.scope-item { display: flex; gap: 18rpx; margin-top: 14rpx; padding: 24rpx; border: 1rpx solid var(--archive-line); background: rgba(255, 255, 255, .38); }
.scope-item.active { border-color: #9e4c3b; background: #fff9ee; }
.scope-mark { width: 52rpx; height: 52rpx; border: 1rpx solid #b88a52; color: #8d5138; font-size: 25rpx; line-height: 52rpx; text-align: center; }
.scope-title, .scope-desc, .scroll-title text, .tree-member b, .tree-member text { display: block; }
.scope-title { color: var(--archive-ink); font-size: 29rpx; font-weight: 700; }
.scope-desc { margin-top: 5rpx; color: var(--archive-ink-soft); font-size: 24rpx; line-height: 1.5; }
.quality-warning { margin-top: 22rpx; padding: 18rpx 20rpx; border-left: 4rpx solid #b34a3a; background: #fff4e4; color: #8d5138; font-size: 24rpx; line-height: 1.65; }
.scroll-preview { margin-top: 28rpx; padding: 38rpx 30rpx; border: 2rpx solid #b88a52; background: #f8f0df; color: #193856; }
.scroll-top, .scroll-foot { display: flex; justify-content: space-between; color: #8d5138; font-size: 22rpx; }
.scroll-title { display: flex; flex-direction: column; align-items: center; gap: 12rpx; margin: 34rpx 0 26rpx; font-size: 42rpx; font-weight: 700; }
.scroll-title text:last-child { color: #687488; font-size: 24rpx; font-weight: 400; }
.scroll-stats { display: flex; justify-content: space-around; padding: 22rpx 0; border-top: 1rpx solid #c9a56b; border-bottom: 1rpx solid #c9a56b; }
.scroll-stats view { display: flex; flex-direction: column; align-items: center; color: #687488; font-size: 21rpx; }
.scroll-stats b { margin-bottom: 6rpx; color: #193856; font-size: 28rpx; }
.lineage-explain { display: flex; gap: 12rpx; padding: 20rpx 0 2rpx; color: #687488; font-size: 21rpx; line-height: 1.55; }
.usage-explain { padding-top: 12rpx; }
.lineage-explain text:first-child { flex: 0 0 auto; color: #8d5138; font-weight: 700; }
.scroll-foot { margin-top: 28rpx; padding-top: 18rpx; border-top: 1rpx solid #c9a56b; }
.export-actions { display: flex; flex-wrap: wrap; gap: 16rpx; margin-top: 22rpx; }
.export-actions :deep(.mini-button) { min-width: calc(50% - 8rpx); flex: 1; }
.export-canvas { position: fixed; top: -10000px; left: -10000px; opacity: 0; pointer-events: none; }
</style>
