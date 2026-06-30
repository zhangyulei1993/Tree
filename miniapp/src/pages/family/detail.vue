<template>
  <view class="tree-page">
    <MiniBackHome />
    <MiniCard v-if="!authChecked">
      <MiniEmptyState symbol="…" title="正在确认登录状态" description="请稍候..." />
    </MiniCard>

    <MiniCard v-else-if="loading">
      <MiniEmptyState symbol="…" title="正在加载" description="正在加载家庭详情..." />
    </MiniCard>

    <MiniCard v-else-if="errorMessage">
      <MiniNotice tone="warm" title="加载失败">{{ errorMessage }}</MiniNotice>
      <MiniButton variant="secondary" @click="loadFamily">重新加载</MiniButton>
    </MiniCard>

    <template v-else-if="family">
      <MiniCard variant="hero" class="dossier-hero tree-pedigree-watermark">
        <view class="tree-dossier-head">
          <view class="tree-dossier-seal">{{ family.familySurname.slice(0, 1) }}</view>
          <view class="tree-dossier-copy">
            <text class="dossier-eyebrow">家庭档案</text>
            <text class="tree-page-title">{{ family.familyName }}</text>
            <view class="tag-row">
              <MiniStatusTag :label="familyStatusText(family.status)" :tone="statusTagTone(family.status)" />
              <MiniStatusTag
                :label="publicStatusText(family.publicDisplayStatus)"
                :tone="statusTagTone(family.publicDisplayStatus)"
              />
            </view>
          </view>
        </view>
        <text class="tree-muted dossier-desc">{{ family.description || '暂未填写家庭简介。' }}</text>
        <view class="tree-archive-ribbon">
          <text v-if="family.familySurname" class="tree-archive-chip gold">{{ family.familySurname }}氏</text>
          <text v-if="family.regionText || family.nativePlace" class="tree-archive-chip">
            {{ family.regionText || family.nativePlace }}
          </text>
        </view>
      </MiniCard>

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
        <view v-if="editingProfile" class="tree-space-body section-pad">
          <view class="readonly-field">
            <text class="readonly-label">家庭主姓氏</text>
            <text class="readonly-value">{{ family.familySurname }}</text>
          </view>
          <input v-model.trim="profileForm.familyName" class="tree-input" maxlength="120" placeholder="家庭名称" />
          <input v-model.trim="profileForm.nativePlace" class="tree-input" maxlength="120" placeholder="籍贯" />
          <input v-model.trim="profileForm.regionText" class="tree-input" maxlength="120" placeholder="地区" />
          <textarea v-model.trim="profileForm.description" class="tree-textarea" maxlength="1000" placeholder="家庭简介" />
          <MiniNotice tone="security" title="资料范围">
            主姓氏创建后不可在此修改；公开联系方式、搜索开关和公开审核在“公开展示”中管理。
          </MiniNotice>
          <text v-if="profileError" class="tree-field-error">{{ profileError }}</text>
          <MiniButton :loading="savingProfile" :disabled="savingProfile" @click="saveProfile">
            保存家庭资料
          </MiniButton>
        </view>
        <view v-else class="tree-space-body section-pad">
          <view class="tree-info-row">
            <text class="tree-info-label">姓氏</text>
            <text class="tree-info-value">{{ family.familySurname }}</text>
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

      <view class="tree-space">
        <view class="tree-space-head">
          <text class="tree-space-title">我的身份</text>
          <text class="tree-space-subtitle">你在该家庭中的角色</text>
        </view>
        <view class="tree-space-body section-pad">
          <view class="tree-info-row">
            <text class="tree-info-label">家庭角色</text>
            <text class="tree-info-value">{{ roleText(family.role) }}</text>
          </view>
          <view class="tree-info-row">
            <text class="tree-info-label">家庭状态</text>
            <text class="tree-info-value">{{ familyStatusText(family.status) }}</text>
          </view>
        </view>
      </view>

      <view class="tree-space">
        <view class="tree-space-head">
          <text class="tree-space-title">查看家庭</text>
          <text class="tree-space-subtitle">成员名册与家谱结构</text>
        </view>
        <view class="tree-space-body">
          <view class="tree-action-tile clickable-tile" @click="openMembers">
            <view class="tree-action-tile-icon">
              <view class="tree-symbol tree-symbol-users" />
            </view>
            <view class="tree-action-tile-copy">
              <text class="tree-action-tile-title">成员名册</text>
              <text class="tree-action-tile-desc">查看成员基本信息与绑定状态</text>
            </view>
            <text class="feature-arrow">›</text>
          </view>
          <view class="tree-action-tile clickable-tile" @click="openTree">
            <view class="tree-action-tile-icon">
              <view class="tree-symbol tree-symbol-folder" />
            </view>
            <view class="tree-action-tile-copy">
              <text class="tree-action-tile-title">私有家谱</text>
              <text class="tree-action-tile-desc">查看父母子女与配偶关系</text>
            </view>
            <text class="feature-arrow">›</text>
          </view>
        </view>
      </view>

      <view v-if="canManageFamily" class="tree-space">
        <view class="tree-space-head">
          <text class="tree-space-title">管理工作台</text>
          <text class="tree-space-subtitle">编辑家谱、处理申请和管理权限</text>
        </view>
        <view class="tree-space-body">
          <view class="tree-action-tile clickable-tile" @click="openManage">
            <view class="tree-action-tile-icon">
              <view class="tree-symbol tree-symbol-users" />
            </view>
            <view class="tree-action-tile-copy">
              <text class="tree-action-tile-title">编辑家谱</text>
              <text class="tree-action-tile-desc">新增、修改成员并维护家谱关系</text>
            </view>
            <text class="feature-arrow">›</text>
          </view>
          <view class="tree-action-tile clickable-tile" @click="openJoinRequests">
            <view class="tree-action-tile-icon">
              <view class="tree-symbol tree-symbol-users" />
            </view>
            <view class="tree-action-tile-copy">
              <text class="tree-action-tile-title">收到的加入申请</text>
              <text class="tree-action-tile-desc">查看并处理申请加入该家庭的记录</text>
            </view>
            <text class="feature-arrow">›</text>
          </view>
          <view class="tree-action-tile clickable-tile" @click="openSentInvitations">
            <view class="tree-action-tile-icon">
              <view class="tree-symbol tree-symbol-users" />
            </view>
            <view class="tree-action-tile-copy">
              <text class="tree-action-tile-title">发出的成员邀请</text>
              <text class="tree-action-tile-desc">查看、取消或重新生成成员邀请</text>
            </view>
            <text class="feature-arrow">›</text>
          </view>
          <view class="tree-action-tile clickable-tile" @click="openSettings">
            <view class="tree-action-tile-icon">
              <view class="tree-symbol tree-symbol-folder" />
            </view>
            <view class="tree-action-tile-copy">
              <text class="tree-action-tile-title">公开展示与权限</text>
              <text class="tree-action-tile-desc">公开信息、管理员与高风险操作</text>
            </view>
            <text class="feature-arrow">›</text>
          </view>
        </view>
      </view>

      <view v-if="family.role !== 'FOUNDER'" class="tree-space danger-space">
        <view class="tree-space-head">
          <text class="tree-space-title">退出家庭</text>
          <text class="tree-space-subtitle">账号退出后，家谱中的成员节点仍会保留</text>
        </view>
        <view class="tree-space-body section-pad">
          <MiniNotice tone="warm" title="退出说明">
            退出只解除你的账号绑定，不删除成员资料和亲属关系。再次加入需要重新申请或接受邀请。
          </MiniNotice>
          <MiniButton variant="danger" :loading="leaving" :disabled="leaving" @click="confirmLeave">
            退出该家庭
          </MiniButton>
        </view>
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
import MiniCard from '@/components/base/MiniCard.vue'
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
import MiniStatusTag from '@/components/base/MiniStatusTag.vue'
import { statusTagTone } from '@/components/base/formatStatus'
import { useSessionStore } from '@/stores/session'
import type { FamilyDetail } from '@/types/api'

const session = useSessionStore()
const familyId = ref('')
const family = ref<FamilyDetail | null>(null)
const loading = ref(false)
const errorMessage = ref('')
const authChecked = ref(false)
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
  authChecked.value = false
  family.value = null
  if (!familyId.value) {
    authChecked.value = true
    errorMessage.value = '缺少家庭信息。'
    return
  }
  const route = `/pages/family/detail?familyId=${encodeURIComponent(familyId.value)}`
  if (!session.requireLogin(route)) return
  authChecked.value = true
  loading.value = true
  errorMessage.value = ''
  try {
    family.value = await getFamilyDetail(familyId.value)
    fillProfileForm()
  } catch (error) {
    errorMessage.value = apiErrorMessage(error, '家庭详情加载失败。')
  } finally {
    loading.value = false
  }
}

function resetAuthView() {
  authChecked.value = false
  loading.value = false
  errorMessage.value = ''
  family.value = null
  editingProfile.value = false
  profileError.value = ''
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

function openMembers() {
  uni.navigateTo({
    url: `/pages/family/members?familyId=${encodeURIComponent(familyId.value)}`
  })
}

function openTree() {
  uni.navigateTo({
    url: `/pages/family/private-tree?familyId=${encodeURIComponent(familyId.value)}`
  })
}

function openManage() {
  uni.navigateTo({ url: `/pages/family/manage?familyId=${encodeURIComponent(familyId.value)}` })
}

function openSettings() {
  uni.navigateTo({ url: `/pages/family/settings?familyId=${encodeURIComponent(familyId.value)}` })
}

function openJoinRequests() {
  uni.navigateTo({
    url: `/pages/join/family?familyId=${encodeURIComponent(familyId.value)}`
  })
}

function openSentInvitations() {
  uni.navigateTo({
    url: `/pages/invite/sent?familyId=${encodeURIComponent(familyId.value)}`
  })
}

function confirmLeave() {
  if (!family.value || leaving.value) return
  uni.showModal({
    title: '确认退出家庭',
    content: '退出后你的账号将与当前成员节点解除绑定，家谱资料不会删除。确定继续吗？',
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
  switch (role) {
    case 'FOUNDER':
      return '创建者'
    case 'FAMILY_ADMIN':
      return '管理员'
    case 'MEMBER':
      return '成员'
    case 'ROOT_ADMIN':
      return '平台管理员'
    default:
      return role ? '未知角色' : '成员'
  }
}

function familyStatusText(status: string) {
  switch (status) {
    case 'NORMAL':
      return '正常'
    case 'DISABLED':
      return '已停用'
    case 'DISSOLVED':
      return '已解散'
    case 'DISSOLUTION_PENDING':
      return '解散待审核'
    case 'PENDING':
      return '待审核'
    case 'APPROVED':
      return '已通过'
    case 'REJECTED':
      return '已拒绝'
    case 'TAKEN_DOWN':
      return '已下架'
    case 'DELETED':
      return '已删除'
    default:
      return '未知状态'
  }
}

function publicStatusText(status: string) {
  switch (status) {
    case 'APPROVED':
      return '已公开'
    case 'PENDING':
      return '待审核'
    case 'REJECTED':
      return '已拒绝'
    case 'PRIVATE':
      return '未公开'
    case 'TAKEN_DOWN':
      return '已下架'
    default:
      return '未知状态'
  }
}

onLoad((options) => {
  familyId.value = String(options?.familyId || '')
})
onShow(loadFamily)
onHide(resetAuthView)
onUnload(resetAuthView)
</script>

<style scoped>
.tree-space-head {
  align-items: flex-start;
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

.section-pad :deep(.mini-notice) {
  margin-bottom: 18rpx;
}

.tree-page {
  min-height: auto;
  padding-bottom: calc(160rpx + env(safe-area-inset-bottom));
  background:
    radial-gradient(circle at 92% 0%, rgba(216, 175, 104, 0.14), transparent 260rpx),
    radial-gradient(circle at 0% 20%, rgba(24, 54, 83, 0.08), transparent 300rpx);
}

.dossier-hero {
  margin-bottom: 26rpx;
  padding-top: 34rpx;
  padding-bottom: 34rpx;
}

.dossier-hero :deep(.tree-page-title),
.dossier-hero .tree-page-title {
  color: #fff;
}

.dossier-eyebrow {
  display: block;
  margin-bottom: 8rpx;
  color: rgba(248, 231, 194, 0.92);
  font-size: 22rpx;
  font-weight: 600;
  letter-spacing: 2rpx;
}

.dossier-desc {
  display: block;
  margin-top: 16rpx;
  color: rgba(255, 255, 255, 0.76);
  line-height: 1.7;
}

.section-pad {
  padding: 8rpx 24rpx 16rpx;
}

.entry-panel {
  margin-top: 4rpx;
}

.clickable-tile {
  position: relative;
  margin-bottom: 16rpx;
  border: 1rpx solid rgba(255, 255, 255, 0.72);
  border-radius: 24rpx;
  background:
    radial-gradient(circle at 100% 0%, rgba(47, 107, 87, 0.08), transparent 150rpx),
    linear-gradient(135deg, rgba(255, 255, 255, 0.98) 0%, rgba(248, 250, 252, 0.96) 100%);
  box-shadow: 0 10rpx 26rpx rgba(24, 54, 83, 0.055);
}

.tree-space:nth-of-type(5) .clickable-tile {
  background:
    radial-gradient(circle at 100% 0%, rgba(216, 175, 104, 0.13), transparent 150rpx),
    linear-gradient(135deg, rgba(255, 252, 247, 0.98) 0%, rgba(244, 250, 247, 0.96) 100%);
}

.clickable-tile:active {
  opacity: 0.88;
}

.clickable-tile .tree-action-tile-copy {
  padding-right: 12rpx;
}

.clickable-tile .tree-action-tile-title {
  color: var(--tree-text-primary, #1e293b);
}

.clickable-tile .tree-action-tile-desc {
  color: var(--tree-text-secondary, #64748b);
}

.tree-space:last-child {
  margin-bottom: 48rpx;
}

.feature-arrow {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 40rpx;
  height: 40rpx;
  border-radius: 999rpx;
  background: rgba(24, 54, 83, 0.08);
  color: var(--tree-primary);
  font-size: 28rpx;
  font-weight: 700;
}

.tag-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8rpx;
  margin-top: 12rpx;
}

</style>
