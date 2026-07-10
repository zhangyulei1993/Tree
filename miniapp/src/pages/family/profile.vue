<template>
  <view class="archive-page profile-page">
    <MiniBackHome />

    <MiniFamilyPageSkeleton v-if="loading && !family" variant="profile" />

    <view v-else-if="errorMessage && !family" class="profile-state archive-form-panel">
      <MiniNotice tone="warm" title="加载失败">{{ errorMessage }}</MiniNotice>
      <MiniButton variant="secondary" class="profile-action" @click="loadFamily">重新加载</MiniButton>
    </view>

    <template v-else-if="family">
      <view class="profile-head archive-page-head">
        <view>
          <text class="archive-kicker">Family Dossier</text>
          <text class="archive-title">家庭档案</text>
          <text class="archive-subtitle">基本资料与公开展示摘要</text>
        </view>
        <view class="archive-seal">{{ family.familySurname.slice(0, 1) }}</view>
      </view>

      <view class="profile-context">
        <text class="context-back" @click="openFamilyDetail">返回详情</text>
      </view>

      <view class="profile-cover">
        <view class="profile-cover-copy">
          <text class="profile-surname">{{ family.familySurname }}氏</text>
          <text class="profile-name">{{ family.familyName }}</text>
          <text class="profile-region">{{ regionSummary }}</text>
          <view class="profile-status-line">
            <text class="profile-status-badge">{{ publicStatusText(family.publicDisplayStatus) }}</text>
            <text v-if="family.nativePlace" class="archive-chip">{{ family.nativePlace }}</text>
            <text v-if="family.regionText" class="archive-chip">{{ family.regionText }}</text>
          </view>
        </view>
        <view class="profile-spine archive-book-spine">
          <text>档</text>
          <text>案</text>
        </view>
      </view>

      <view class="archive-form-panel">
        <view class="profile-panel-head">
          <view class="archive-section-head profile-section-head">
            <text class="archive-section-title">基本资料</text>
            <text class="archive-section-subtitle">家庭名称、籍贯与地区</text>
          </view>
          <text
            v-if="canManageFamily"
            class="archive-thin-button profile-edit-action"
            @tap="toggleProfileEdit"
          >
            {{ editingProfile ? '取消编辑' : '编辑资料' }}
          </text>
        </view>

        <view v-if="editingProfile" key="edit" class="profile-panel">
          <view class="readonly-field">
            <text class="readonly-label">家庭主姓氏</text>
            <text class="readonly-value">{{ family.familySurname }}</text>
          </view>
          <input v-model.trim="profileForm.familyName" class="tree-input" maxlength="120" placeholder="家庭名称" />
          <input v-model.trim="profileForm.nativePlace" class="tree-input" maxlength="120" placeholder="籍贯" />
          <input v-model.trim="profileForm.regionText" class="tree-input" maxlength="120" placeholder="地区" />
          <textarea v-model.trim="profileForm.description" class="tree-textarea" maxlength="1000" placeholder="家庭简介" />
          <MiniNotice tone="security" title="资料范围">
            主姓氏创建后不可在此修改；公开联系方式、搜索开关和公开审核在“公开展示与权限”中管理。
          </MiniNotice>
          <text v-if="profileError" class="tree-field-error">{{ profileError }}</text>
          <MiniButton class="profile-action" :loading="savingProfile" :disabled="savingProfile" @click="saveProfile">
            保存家庭资料
          </MiniButton>
        </view>
        <view v-else key="view" class="profile-panel">
          <view class="tree-info-row">
            <text class="tree-info-label">姓氏</text>
            <text class="tree-info-value">{{ family.familySurname }}</text>
          </view>
          <view class="tree-info-row">
            <text class="tree-info-label">家庭名称</text>
            <text class="tree-info-value">{{ family.familyName }}</text>
          </view>
          <view class="tree-info-row">
            <text class="tree-info-label">籍贯</text>
            <text class="tree-info-value">{{ family.nativePlace || '未设置' }}</text>
          </view>
          <view class="tree-info-row">
            <text class="tree-info-label">地区</text>
            <text class="tree-info-value">{{ family.regionText || '未设置' }}</text>
          </view>
        </view>
      </view>

      <view v-if="!editingProfile" class="archive-form-panel">
        <view class="archive-section-head">
          <text class="archive-section-title">家庭简介</text>
          <text class="archive-section-subtitle">对内展示的家庭说明</text>
        </view>
        <text class="profile-description">{{ family.description || '未设置' }}</text>
      </view>

      <view v-if="!editingProfile" class="archive-form-panel">
        <view class="archive-section-head">
          <text class="archive-section-title">公开展示摘要</text>
          <text class="archive-section-subtitle">公开状态与联系方式在“公开展示与权限”中维护</text>
        </view>
        <view class="tree-info-row">
          <text class="tree-info-label">公开状态</text>
          <text class="tree-info-value">{{ publicStatusText(family.publicDisplayStatus) }}</text>
        </view>
        <view class="tree-info-row">
          <text class="tree-info-label">允许被搜索</text>
          <text class="tree-info-value">{{ family.searchable ? '是' : '否' }}</text>
        </view>
        <view class="tree-info-row">
          <text class="tree-info-label">公开联系方式</text>
          <text class="tree-info-value">{{ publicContactText }}</text>
        </view>
      </view>

      <view v-if="family.role !== 'FOUNDER'" class="archive-danger-panel">
        <text class="archive-danger-title">退出家庭</text>
        <text class="archive-danger-desc">
          仅解除账号绑定，不删除家庭树资料。创建者需先转让身份后才能退出。
        </text>
        <MiniButton class="profile-action" variant="danger" :loading="leaving" :disabled="leaving" @click="confirmLeave">
          退出该家庭
        </MiniButton>
      </view>
    </template>
  </view>
</template>

<script setup lang="ts">
import { onHide, onLoad, onShow, onUnload } from '@dcloudio/uni-app'
import { computed, reactive, ref } from 'vue'

import { apiErrorMessage } from '@/api/client'
import { getFamilyDetail, leaveFamily, updateFamily } from '@/api/families'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniFamilyPageSkeleton from '@/components/base/MiniFamilyPageSkeleton.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
import { useSessionStore } from '@/stores/session'
import type { FamilyDetail } from '@/types/api'
import { normalizeText, optionalText, validateTextFields } from '@/utils/inputValidation'

const session = useSessionStore()
const familyId = ref('')
const family = ref<FamilyDetail | null>(null)
const loading = ref(false)
const errorMessage = ref('')
const leaving = ref(false)
const editingProfile = ref(false)
const savingProfile = ref(false)
const profileError = ref('')
const profileForm = reactive({
  familyName: '',
  nativePlace: '',
  regionText: '',
  description: ''
})

const publicContactText = computed(() => {
  if (!family.value?.publicContactVisible) return '未公开'
  const contacts = [
    family.value.publicContactName,
    family.value.publicContactPhone,
    family.value.publicContactWechat,
    family.value.publicContactNote
  ].filter(Boolean)
  return contacts.length > 0 ? contacts.join(' / ') : '未设置'
})

const regionSummary = computed(() => {
  if (!family.value) return '地区未设置'
  const parts = [family.value.nativePlace, family.value.regionText].filter(Boolean)
  return parts.length > 0 ? parts.join(' · ') : '地区未设置'
})

const canManageFamily = computed(() =>
  family.value?.role === 'FOUNDER' || family.value?.role === 'FAMILY_ADMIN'
)

async function loadFamily() {
  session.restoreSession()
  if (!familyId.value) {
    errorMessage.value = '缺少家庭信息。'
    return
  }
  const route = `/pages/family/profile?familyId=${encodeURIComponent(familyId.value)}`
  if (!session.isLoggedIn) {
    resetPageData()
    session.requireLogin(route)
    return
  }
  const isInitialLoad = !family.value
  loading.value = isInitialLoad
  errorMessage.value = ''
  try {
    family.value = await getFamilyDetail(familyId.value)
    fillProfileForm()
  } catch (error) {
    errorMessage.value = apiErrorMessage(error, '家庭档案加载失败。')
  } finally {
    loading.value = false
  }
}

function resetTransientUI() {
  editingProfile.value = false
  profileError.value = ''
}

function resetPageData() {
  loading.value = false
  errorMessage.value = ''
  family.value = null
  resetTransientUI()
}

function fillProfileForm() {
  if (!family.value) return
  profileForm.familyName = family.value.familyName || ''
  profileForm.nativePlace = family.value.nativePlace || ''
  profileForm.regionText = family.value.regionText || ''
  profileForm.description = family.value.description || ''
}

function toggleProfileEdit() {
  if (!canManageFamily.value) return
  editingProfile.value = !editingProfile.value
  profileError.value = ''
  if (editingProfile.value) fillProfileForm()
}

async function saveProfile() {
  if (!family.value || savingProfile.value) return
  const validationMessage = validateTextFields([
    { value: profileForm.familyName, label: '家庭名称', kind: 'name', required: true, maxLength: 120 },
    { value: profileForm.nativePlace, label: '籍贯', maxLength: 120 },
    { value: profileForm.regionText, label: '地区', maxLength: 120 },
    { value: profileForm.description, label: '家庭简介', kind: 'multiLine', maxLength: 1000 }
  ])
  if (validationMessage) {
    profileError.value = validationMessage
    return
  }
  savingProfile.value = true
  profileError.value = ''
  try {
    family.value = await updateFamily(familyId.value, {
      ...profileForm,
      familyName: normalizeText(profileForm.familyName),
      nativePlace: optionalText(profileForm.nativePlace),
      regionText: optionalText(profileForm.regionText),
      description: optionalText(profileForm.description)
    })
    editingProfile.value = false
    uni.showToast({ title: '家庭资料已保存', icon: 'success' })
  } catch (error) {
    profileError.value = apiErrorMessage(error, '保存家庭资料失败。')
  } finally {
    savingProfile.value = false
  }
}

function openFamilyDetail() {
  if (getCurrentPages().length > 1) {
    uni.navigateBack({ delta: 1 })
    return
  }
  uni.redirectTo({ url: `/pages/family/detail?familyId=${encodeURIComponent(familyId.value)}` })
}

function confirmLeave() {
  if (!family.value || leaving.value) return
  uni.showModal({
    title: '确认退出家庭',
    content: '退出后你的账号将与当前成员节点解除绑定，家庭树资料不会删除。再次加入需要重新申请或接受邀请。确定继续吗？',
    confirmColor: '#B5473C',
    success: (result) => {
      if (result.confirm) submitLeave()
    }
  })
}

async function submitLeave() {
  leaving.value = true
  try {
    await leaveFamily(familyId.value, '用户主动退出家庭')
    uni.showToast({ title: '已退出家庭', icon: 'success' })
    setTimeout(() => {
      uni.switchTab({ url: '/pages/family/my' })
    }, 500)
  } catch (error) {
    uni.showModal({
      title: '退出失败',
      content: apiErrorMessage(error, '暂时无法退出该家庭。'),
      showCancel: false
    })
  } finally {
    leaving.value = false
  }
}

function publicStatusText(status: string) {
  return ({
    APPROVED: '已公开',
    PENDING: '待审核',
    REJECTED: '已拒绝',
    PRIVATE: '未公开',
    TAKEN_DOWN: '已下架'
  } as Record<string, string>)[status] || '未知状态'
}

onLoad((options) => {
  familyId.value = String(options?.familyId || '')
})
onShow(loadFamily)
onHide(resetTransientUI)
onUnload(resetPageData)
</script>

<style scoped>
.profile-page {
  padding-top: 28rpx;
}

.profile-page :deep(.mini-back-home) {
  margin-bottom: 22rpx;
}

.profile-state :deep(.mini-notice) {
  margin-bottom: 18rpx;
}

.profile-context {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16rpx;
  margin-bottom: 24rpx;
}

.context-back {
  color: var(--archive-ink-soft);
  font-size: 23rpx;
  line-height: 1.4;
}

.context-back:active {
  color: var(--archive-cinnabar);
}

.profile-cover {
  display: flex;
  min-height: 210rpx;
  margin-bottom: 24rpx;
  border-top: 1rpx solid var(--archive-line-strong);
  border-bottom: 1rpx solid var(--archive-line);
}

.profile-cover-copy {
  flex: 1;
  min-width: 0;
  padding: 28rpx 28rpx 28rpx 0;
}

.profile-surname,
.profile-name,
.profile-region {
  display: block;
}

.profile-surname {
  color: var(--archive-cinnabar);
  font-size: 24rpx;
  font-weight: 700;
  letter-spacing: 2rpx;
}

.profile-name {
  margin-top: 10rpx;
  color: var(--archive-blue);
  font-family: 'Songti SC', 'STSong', serif;
  font-size: 42rpx;
  font-weight: 800;
  letter-spacing: 2rpx;
  line-height: 1.28;
}

.profile-region {
  margin-top: 12rpx;
  color: var(--archive-ink-soft);
  font-size: 23rpx;
  line-height: 1.55;
}

.profile-status-line {
  display: flex;
  flex-wrap: wrap;
  gap: 10rpx;
  margin-top: 16rpx;
}

.profile-status-badge {
  display: inline-flex;
  align-items: center;
  min-height: 42rpx;
  border: 1rpx solid rgba(22, 51, 83, 0.22);
  background: rgba(22, 51, 83, 0.08);
  color: var(--archive-blue);
  padding: 0 14rpx;
  font-size: 21rpx;
  font-weight: 650;
  line-height: 1.2;
}

.profile-spine {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12rpx;
  width: 118rpx;
  flex-shrink: 0;
}

.profile-spine text {
  color: rgba(255, 255, 255, 0.94);
  font-family: 'Songti SC', 'STSong', serif;
  font-size: 30rpx;
  font-weight: 800;
}

.profile-page .archive-form-panel + .archive-form-panel {
  margin-top: -8rpx;
}

.profile-panel-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16rpx;
  margin-bottom: 18rpx;
}

.profile-section-head {
  margin-bottom: 0;
  padding-bottom: 0;
  border-bottom: 0;
}

.profile-edit-action {
  flex-shrink: 0;
}

.profile-panel :deep(.mini-notice) {
  margin-bottom: 18rpx;
}

.profile-description {
  display: block;
  color: var(--archive-ink);
  font-size: 26rpx;
  line-height: 1.68;
  white-space: pre-wrap;
}

.profile-action {
  margin-top: 20rpx;
  align-self: flex-start;
}

.readonly-field {
  margin-bottom: 16rpx;
  border-bottom: 1rpx solid var(--archive-line);
  padding-bottom: 14rpx;
}

.readonly-label,
.readonly-value {
  display: block;
}

.readonly-label {
  color: var(--archive-cinnabar);
  font-size: 21rpx;
}

.readonly-value {
  margin-top: 8rpx;
  color: var(--archive-ink);
  font-size: 28rpx;
  font-weight: 700;
}
</style>
