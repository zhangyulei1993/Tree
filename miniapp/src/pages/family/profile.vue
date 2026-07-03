<template>
  <view class="tree-page profile-page">
    <MiniBackHome />
    <FamilyContextHeader
      v-if="family"
      :family-name="family.familyName"
      section="家庭档案"
      subtitle="基本资料与公开展示摘要"
      :role-label="roleText(family.role)"
      back-label="返回详情"
      @back="openFamilyDetail"
    />

    <MiniFamilyPageSkeleton v-if="loading && !family" variant="profile" />

    <MiniCard v-else-if="errorMessage && !family">
      <MiniNotice tone="warm" title="加载失败">{{ errorMessage }}</MiniNotice>
      <MiniButton variant="secondary" @click="loadFamily">重新加载</MiniButton>
    </MiniCard>

    <template v-else-if="family">
      <view class="tree-space">
        <view class="tree-space-head">
          <view>
            <text class="tree-space-title">档案信息</text>
            <text class="tree-space-subtitle">家庭基本资料与公开展示状态</text>
          </view>
          <text v-if="canManageFamily" class="section-action" @tap="toggleProfileEdit">
            {{ editingProfile ? '取消编辑' : '编辑资料' }}
          </text>
        </view>
        <view v-if="editingProfile" key="edit" class="tree-space-body section-pad profile-panel">
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
          <MiniButton :loading="savingProfile" :disabled="savingProfile" @click="saveProfile">
            保存家庭资料
          </MiniButton>
        </view>
        <view v-else key="view" class="tree-space-body section-pad profile-panel">
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
          <view class="tree-info-row">
            <text class="tree-info-label">简介</text>
            <text class="tree-info-value">{{ family.description || '未设置' }}</text>
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
      </view>

      <MiniCard v-if="family.role !== 'FOUNDER'" class="danger-zone">
        <MiniSectionHeader title="退出家庭" subtitle="仅解除账号绑定，不删除家谱资料" />
        <MiniButton variant="danger" :loading="leaving" :disabled="leaving" @click="confirmLeave">
          退出该家庭
        </MiniButton>
      </MiniCard>
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
import MiniCard from '@/components/base/MiniCard.vue'
import MiniFamilyPageSkeleton from '@/components/base/MiniFamilyPageSkeleton.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
import MiniSectionHeader from '@/components/base/MiniSectionHeader.vue'
import FamilyContextHeader from '@/components/family/FamilyContextHeader.vue'
import { useSessionStore } from '@/stores/session'
import type { FamilyDetail } from '@/types/api'

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
  if (!profileForm.familyName) {
    profileError.value = '请填写家庭名称。'
    return
  }
  savingProfile.value = true
  profileError.value = ''
  try {
    family.value = await updateFamily(familyId.value, { ...profileForm })
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
    content: '退出后你的账号将与当前成员节点解除绑定，家谱资料不会删除。再次加入需要重新申请或接受邀请。确定继续吗？',
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
      uni.redirectTo({ url: '/pages/family/my' })
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

function roleText(role: string) {
  return ({ FOUNDER: '创建者', FAMILY_ADMIN: '家庭管理员', MEMBER: '普通成员' } as Record<string, string>)[role] || '成员'
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
  background:
    radial-gradient(circle at 92% 0%, rgba(216, 175, 104, 0.14), transparent 260rpx),
    radial-gradient(circle at 0% 18%, rgba(24, 54, 83, 0.08), transparent 300rpx);
}

.tree-space-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16rpx;
  padding: 0 8rpx 12rpx;
}

.section-action {
  flex-shrink: 0;
  border: 1rpx solid rgba(47, 107, 87, 0.18);
  border-radius: 999rpx;
  background: rgba(47, 107, 87, 0.08);
  color: var(--tree-green);
  padding: 8rpx 16rpx;
  font-size: 22rpx;
  font-weight: 800;
  transition: transform 180ms ease-out, background-color 180ms ease-out;
}

.section-action:active {
  background-color: rgba(47, 107, 87, 0.14);
  transform: translateY(1rpx) scale(0.99);
}

.profile-panel {
  animation: profile-panel-in 200ms ease-out;
}

@keyframes profile-panel-in {
  from {
    opacity: 0;
    transform: translateY(8rpx);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.readonly-field {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16rpx;
  border-radius: 20rpx;
  background: rgba(248, 250, 252, 0.9);
  padding: 20rpx 24rpx;
}

.readonly-label {
  color: var(--tree-text-secondary);
  font-size: 24rpx;
}

.readonly-value {
  color: var(--tree-text-primary);
  font-size: 26rpx;
  font-weight: 800;
}

.tree-input,
.tree-textarea {
  margin-bottom: 16rpx;
}

.section-pad {
  padding: 8rpx 24rpx 16rpx;
}

.section-pad :deep(.mini-notice) {
  margin-bottom: 18rpx;
}

.danger-zone {
  margin-top: 20rpx;
  border-color: rgba(181, 71, 60, 0.2);
  background: rgba(255, 248, 246, 0.92);
}
</style>
