<template>
  <view class="archive-page stories-page">
    <MiniBackHome />

    <view v-if="loading" class="stories-state archive-form-panel">
      <MiniEmptyState symbol="…" title="正在加载" description="正在整理家庭变动记录..." />
    </view>
    <view v-else-if="errorMessage" class="stories-state archive-form-panel">
      <MiniNotice tone="warm" title="加载失败">{{ errorMessage }}</MiniNotice>
      <MiniButton variant="secondary" @click="loadPage">重新加载</MiniButton>
    </view>

    <template v-else>
      <view class="stories-head archive-page-head">
        <view>
          <text class="archive-kicker">Family Stories</text>
          <text class="archive-title">家庭事迹</text>
          <text class="archive-subtitle">从家庭成员与关系变动中，留下家庭走过的痕迹</text>
        </view>
        <view class="archive-seal">记</view>
      </view>

      <view v-if="families.length > 1" class="family-switcher archive-panel">
        <text class="family-switcher-label">选择家庭</text>
        <scroll-view scroll-x class="family-switcher-scroll">
          <view class="family-switcher-list">
            <text
              v-for="item in families"
              :key="item.id"
              class="family-switcher-item"
              :class="{ active: String(item.id) === selectedFamilyId }"
              @click="selectFamily(item.id)"
            >
              {{ item.familyName }}
            </text>
          </view>
        </scroll-view>
      </view>

      <view v-if="!family" class="archive-form-panel">
        <MiniEmptyState symbol="记" title="暂无家庭事迹" description="创建或加入家庭后，这里会展示家庭成员与关系的变动记录。" />
      </view>

      <template v-else>
        <view class="stories-context archive-panel">
          <view>
            <text class="stories-family-name">{{ family.familyName }}</text>
            <text class="stories-family-desc">{{ family.regionText || family.nativePlace || '家庭记录' }}</text>
          </view>
          <text class="stories-count">{{ storyItems.length }} 条记录</text>
        </view>

        <MiniNotice tone="info" class="stories-notice">
          当前版本先根据成员、关系和家庭资料的变动记录整理展示，后续可继续补充完整的家庭故事。
        </MiniNotice>

        <view v-if="storyItems.length > 0" class="story-list">
          <view v-for="item in storyItems" :key="item.id" class="story-item">
            <view class="story-marker" />
            <view class="story-content">
              <view class="story-meta">
                <text>{{ item.date }}</text>
                <text>{{ item.source }}</text>
              </view>
              <text class="story-title">{{ item.title }}</text>
              <text class="story-description">{{ item.description }}</text>
            </view>
          </view>
        </view>
        <MiniEmptyState
          v-else
          symbol="记"
          title="暂无变动记录"
          description="添加或调整家庭成员后，这里会自动形成家庭事迹时间线。"
        />
      </template>
    </template>
  </view>
</template>

<script setup lang="ts">
import { onLoad } from '@dcloudio/uni-app'
import { computed, ref } from 'vue'

import { apiErrorMessage, isRealApiMode } from '@/api/client'
import { listMyFamilies, getFamilyDetail } from '@/api/families'
import { listFamilyMembers } from '@/api/members'
import { listFamilyOperationLogs } from '@/api/publicApplications'
import { ensureToolEnabled } from '@/api/toolConfigs'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
import { useSessionStore } from '@/stores/session'
import type { FamilyDetail, FamilyMember, FamilyOperationLog, FamilySummary } from '@/types/api'

interface StoryItem {
  id: string
  date: string
  title: string
  description: string
  source: string
  timestamp: number
}

const session = useSessionStore()
const families = ref<FamilySummary[]>([])
const selectedFamilyId = ref('')
const family = ref<FamilyDetail | null>(null)
const members = ref<FamilyMember[]>([])
const operationLogs = ref<FamilyOperationLog[]>([])
const loading = ref(false)
const errorMessage = ref('')

const canReadOperationLogs = computed(() =>
  family.value?.role === 'FOUNDER' || family.value?.role === 'FAMILY_ADMIN'
)

const storyItems = computed(() => {
  const logStories = operationLogs.value
    .filter((log) => ['FAMILY', 'FAMILY_MEMBER', 'FAMILY_RELATIONSHIP'].includes(log.module))
    .map((log) => {
      const timestamp = Date.parse(log.createdAt)
      const subject = log.memberId ? memberName(log.memberId) : ''
      return {
        id: `log-${log.id}`,
        date: formatDate(log.createdAt),
        title: operationActionText(log.action),
        description: subject ? `${subject} · ${operationModuleText(log.module)}` : operationModuleText(log.module),
        source: '家庭变动',
        timestamp: Number.isFinite(timestamp) ? timestamp : 0
      }
    })

  if (logStories.length > 0) return logStories.sort((left, right) => right.timestamp - left.timestamp).slice(0, 30)

  return members.value
    .flatMap((member) => memberStoryItems(member))
    .sort((left, right) => right.timestamp - left.timestamp)
    .slice(0, 30)
})

async function loadPage(options?: Record<string, unknown>) {
  session.restoreSession()
  if (!(await ensureToolEnabled('FAMILY_STORIES'))) return
  const route = '/pages/family/stories'
  if (!session.isLoggedIn) {
    session.requireLogin(route)
    return
  }

  loading.value = true
  errorMessage.value = ''
  try {
    families.value = await listMyFamilies()
    if (families.value.length === 0) {
      family.value = null
      members.value = []
      operationLogs.value = []
      return
    }

    const requestedFamilyId = String(options?.familyId || '')
    selectedFamilyId.value = families.value.some((item) => String(item.id) === requestedFamilyId)
      ? requestedFamilyId
      : String(families.value[0].id)
    await loadSelectedFamily()
  } catch (error) {
    errorMessage.value = apiErrorMessage(error, '家庭事迹加载失败。')
  } finally {
    loading.value = false
  }
}

async function loadSelectedFamily() {
  if (!selectedFamilyId.value) return
  const familyId = selectedFamilyId.value
  const [familyResult, memberResult] = await Promise.all([
    getFamilyDetail(familyId),
    listFamilyMembers(familyId)
  ])
  family.value = familyResult
  members.value = memberResult
  operationLogs.value = []

  if (isRealApiMode && canReadOperationLogs.value) {
    try {
      const result = await listFamilyOperationLogs(familyId)
      operationLogs.value = result.items
    } catch {
      operationLogs.value = []
    }
  }
}

async function selectFamily(familyId: number | string) {
  if (String(familyId) === selectedFamilyId.value) return
  selectedFamilyId.value = String(familyId)
  loading.value = true
  try {
    await loadSelectedFamily()
  } catch (error) {
    errorMessage.value = apiErrorMessage(error, '家庭事迹加载失败。')
  } finally {
    loading.value = false
  }
}

function memberStoryItems(member: FamilyMember): StoryItem[] {
  const items: StoryItem[] = []
  const createdTimestamp = parseTimestamp(member.createdAt)
  const updatedTimestamp = parseTimestamp(member.updatedAt)
  if (createdTimestamp > 0) {
    items.push({
      id: `member-created-${member.memberId}`,
      date: formatDate(member.createdAt),
      title: `记录成员：${member.name}`,
      description: '家庭成员加入家庭记录。',
      source: '成员变动',
      timestamp: createdTimestamp
    })
  }
  if (updatedTimestamp > createdTimestamp + 60000) {
    items.push({
      id: `member-updated-${member.memberId}`,
      date: formatDate(member.updatedAt),
      title: `更新成员：${member.name}`,
      description: '成员资料或关系信息有过更新。',
      source: '成员变动',
      timestamp: updatedTimestamp
    })
  }
  return items
}

function memberName(memberId: number | string) {
  return members.value.find((member) => String(member.memberId) === String(memberId))?.name || '成员'
}

function parseTimestamp(value?: string) {
  if (!value) return 0
  const timestamp = Date.parse(value)
  return Number.isFinite(timestamp) && timestamp > 0 ? timestamp : 0
}

function formatDate(value?: string) {
  if (!value) return '待确定'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '待确定'
  return date.toLocaleDateString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit' })
}

function operationModuleText(module: string) {
  return ({
    FAMILY: '家庭资料',
    FAMILY_MEMBER: '成员资料',
    FAMILY_RELATIONSHIP: '家庭关系'
  } as Record<string, string>)[module] || '家庭记录'
}

function operationActionText(action: string) {
  return ({
    CREATE_FAMILY: '创建家庭',
    UPDATE_FAMILY: '更新家庭资料',
    CREATE_MEMBER: '记录新成员',
    UPDATE_MEMBER: '更新成员资料',
    DELETE_MEMBER: '移除成员记录',
    CREATE_RELATIONSHIP: '建立家庭关系',
    UPDATE_RELATIONSHIP: '调整家庭关系',
    DELETE_RELATIONSHIP: '解除家庭关系',
    PLACE_EXISTING_MEMBER: '接入暂存成员'
  } as Record<string, string>)[action] || '更新家庭记录'
}

onLoad((options) => {
  void loadPage(options as Record<string, unknown> | undefined)
})
</script>

<style scoped>
.stories-page {
  padding-top: 28rpx;
}

.stories-state {
  margin-top: 24rpx;
}

.family-switcher {
  margin-top: 18rpx;
  padding: 18rpx 20rpx;
}

.family-switcher-label,
.stories-family-name,
.stories-family-desc,
.stories-count,
.story-meta,
.story-title,
.story-description {
  display: block;
}

.family-switcher-label {
  color: var(--archive-ink-soft);
  font-size: 20rpx;
}

.family-switcher-scroll {
  margin-top: 10rpx;
  white-space: nowrap;
}

.family-switcher-list {
  display: inline-flex;
  gap: 10rpx;
}

.family-switcher-item {
  border: 1rpx solid var(--archive-line-strong);
  padding: 8rpx 16rpx;
  color: var(--archive-ink-soft);
  font-size: 21rpx;
}

.family-switcher-item.active {
  border-color: var(--archive-blue);
  background: rgba(35, 73, 98, 0.08);
  color: var(--archive-blue);
  font-weight: 700;
}

.stories-context {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18rpx;
  margin-top: 18rpx;
  border-top: 4rpx solid var(--archive-cinnabar);
  padding: 22rpx 24rpx;
  background: rgba(255, 252, 244, 0.78);
}

.stories-family-name {
  color: var(--archive-ink);
  font-family: 'Songti SC', 'STSong', serif;
  font-size: 30rpx;
  font-weight: 800;
}

.stories-family-desc {
  margin-top: 5rpx;
  color: var(--archive-ink-soft);
  font-size: 20rpx;
}

.stories-count {
  flex-shrink: 0;
  color: var(--archive-blue);
  font-size: 20rpx;
}

.stories-notice {
  margin-top: 18rpx;
}

.story-list {
  position: relative;
  margin-top: 22rpx;
  padding-left: 20rpx;
}

.story-list::before {
  position: absolute;
  top: 6rpx;
  bottom: 6rpx;
  left: 6rpx;
  width: 1rpx;
  background: var(--archive-line-strong);
  content: '';
}

.story-item {
  position: relative;
  display: flex;
  gap: 16rpx;
  padding-bottom: 24rpx;
}

.story-marker {
  position: relative;
  z-index: 1;
  width: 12rpx;
  height: 12rpx;
  flex-shrink: 0;
  margin-top: 8rpx;
  margin-left: -20rpx;
  border: 3rpx solid var(--archive-cinnabar);
  border-radius: 50%;
  background: var(--archive-paper-light);
}

.story-content {
  min-width: 0;
  flex: 1;
  border-bottom: 1rpx solid var(--archive-line);
  padding-bottom: 18rpx;
}

.story-meta {
  color: var(--archive-ink-soft);
  font-size: 19rpx;
}

.story-meta text + text {
  margin-left: 14rpx;
  color: var(--archive-cinnabar);
}

.story-title {
  margin-top: 7rpx;
  color: var(--archive-ink);
  font-size: 27rpx;
  font-weight: 750;
}

.story-description {
  margin-top: 5rpx;
  color: var(--archive-ink-soft);
  font-size: 21rpx;
  line-height: 1.5;
}
</style>
