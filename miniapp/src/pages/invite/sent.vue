<template>
  <view class="tree-page sent-page">
    <MiniBackHome />
    <FamilyContextHeader
      v-if="familyName"
      :family-name="familyName"
      section="发出的成员邀请"
      subtitle="管理该家庭发出的账号绑定邀请"
      :role-label="familyRoleLabel"
      @back="openFamilyOverview"
    />

    <MiniCard v-if="!authChecked">
      <MiniEmptyState symbol="…" title="正在确认权限" description="请稍候..." />
    </MiniCard>

    <template v-else>
      <template v-if="shareResult">
        <MiniCard variant="soft" class="share-result-card">
          <view class="share-result-head">
            <view>
              <text class="share-result-kicker">待发送</text>
              <text class="share-result-title">邀请 {{ shareResult.invitation.targetMemberName }} 绑定账号</text>
            </view>
            <MiniStatusTag status="PENDING" label="待接受" />
          </view>
          <MiniNotice tone="security" title="新邀请已生成">
            对方点开后可先查看完整邀请信息，再决定是否接受。此前生成的邀请链接已失效。
          </MiniNotice>
          <view class="share-summary">
            <view class="tree-info-row">
              <text class="tree-info-label">邀请加入</text>
              <text class="tree-info-value">{{ shareResult.invitation.familyName }}</text>
            </view>
            <view class="tree-info-row">
              <text class="tree-info-label">确认身份</text>
              <text class="tree-info-value">{{ shareResult.invitation.targetMemberName }}</text>
            </view>
            <view class="tree-info-row">
              <text class="tree-info-label">有效期至</text>
              <text class="tree-info-value">{{ formatDate(shareResult.invitation.expiredAt) }}</text>
            </view>
            <view v-if="shareResult.invitation.inviteMessage" class="share-message">
              <text class="share-message-label">给对方的说明</text>
              <text class="share-message-content">{{ shareResult.invitation.inviteMessage }}</text>
            </view>
          </view>
          <!-- #ifdef MP-WEIXIN -->
          <button class="wechat-share-button" open-type="share">发送给微信好友</button>
          <!-- #endif -->
          <MiniButton variant="secondary" @click="copyInviteLink">复制邀请链接</MiniButton>
          <MiniButton variant="ghost" @click="clearShareResult">返回邀请列表</MiniButton>
        </MiniCard>
      </template>

      <template v-else>
        <MiniCard v-if="loading">
          <MiniEmptyState symbol="…" title="正在加载" description="正在加载发出的成员邀请..." />
        </MiniCard>

        <MiniCard v-else-if="loadError">
          <MiniEmptyState
            symbol="!"
            title="加载失败"
            :description="loadError"
            action-text="重新加载"
            @action="loadInvitations"
          />
        </MiniCard>

        <MiniCard v-else-if="invitations.length === 0">
          <MiniEmptyState
            symbol="邀"
            title="暂无发出的成员邀请"
            description="可在成员列表中，对未绑定账号的成员发出邀请。"
            action-text="前往成员列表"
            @action="openMembers"
          />
        </MiniCard>

        <template v-else>
          <view v-for="group in invitationGroups" :key="group.key" class="invitation-group">
            <MiniSectionHeader :title="group.title" :subtitle="group.subtitle" />
            <MiniCard
              v-for="item in group.items"
              :key="item.invitationId"
              variant="soft"
              :class="{ 'actionable-card': canRegenerate(item) }"
            >
          <view class="item-head">
            <view>
              <text class="item-title">{{ item.targetMemberName }}</text>
              <text class="tree-muted item-desc">{{ item.inviteMessage || '未填写邀请说明' }}</text>
            </view>
            <MiniStatusTag :status="displayStatus(item)" :label="invitationStatusText(displayStatus(item))" />
          </view>
          <text class="tree-weak item-meta">创建时间：{{ formatDate(item.createdAt) }}</text>
          <text class="tree-weak item-meta">有效期至：{{ formatDate(item.expiredAt) }}</text>
          <view v-if="canRegenerate(item)" class="item-actions">
            <MiniButton
              :loading="actingId === item.invitationId && actingType === 'regenerate'"
              :disabled="Boolean(actingId)"
              @click="confirmRegenerate(item)"
            >
              重新生成并发送
            </MiniButton>
            <MiniButton
              v-if="item.status === 'PENDING' && !isExpired(item)"
              variant="secondary"
              :loading="actingId === item.invitationId && actingType === 'cancel'"
              :disabled="Boolean(actingId)"
              @click="confirmCancel(item)"
            >
              取消邀请
            </MiniButton>
          </view>
            </MiniCard>
          </view>
        </template>

        <text v-if="operationError" class="tree-field-error">{{ operationError }}</text>
      </template>
    </template>
  </view>
</template>

<script setup lang="ts">
import { onHide, onLoad, onShareAppMessage, onShow, onUnload } from '@dcloudio/uni-app'
import { computed, ref } from 'vue'

import { apiErrorMessage } from '@/api/client'
import { getFamilyDetail } from '@/api/families'
import {
  cancelInvitation,
  listFamilyInvitations,
  regenerateInvitation
} from '@/api/invitations'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniCard from '@/components/base/MiniCard.vue'
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'
import { invitationStatusText } from '@/components/base/formatStatus'
import MiniNotice from '@/components/base/MiniNotice.vue'
import MiniSectionHeader from '@/components/base/MiniSectionHeader.vue'
import MiniStatusTag from '@/components/base/MiniStatusTag.vue'
import FamilyContextHeader from '@/components/family/FamilyContextHeader.vue'
import { useSessionStore } from '@/stores/session'
import type { CreatedInvitation, Invitation } from '@/types/api'

const session = useSessionStore()
const familyId = ref('')
const familyName = ref('')
const familyRole = ref('')
const invitations = ref<Invitation[]>([])
const loading = ref(false)
const authChecked = ref(false)
const loadError = ref('')
const operationError = ref('')
const actingId = ref<number | string | null>(null)
const actingType = ref<'' | 'cancel' | 'regenerate'>('')
const shareResult = ref<CreatedInvitation | null>(null)
const familyRoleLabel = computed(() => familyRole.value === 'FOUNDER' ? '创建者' : '管理员')
const invitationGroups = computed(() => {
  const current = invitations.value.filter((item) => displayStatus(item) === 'PENDING')
  const history = invitations.value.filter((item) => displayStatus(item) !== 'PENDING')
  return [
    { key: 'current', title: '待对方处理', subtitle: '当前仍有效的邀请', items: current },
    { key: 'history', title: '历史邀请', subtitle: '已接受、拒绝、取消或过期的记录', items: history }
  ].filter((group) => group.items.length > 0)
})

function currentRoute() {
  return `/pages/invite/sent?familyId=${encodeURIComponent(familyId.value)}`
}

async function loadInvitations() {
  authChecked.value = false
  invitations.value = []
  if (!familyId.value) {
    authChecked.value = true
    loadError.value = '缺少家庭信息。'
    return
  }
  if (!session.requireLogin(currentRoute())) return
  authChecked.value = true
  loading.value = true
  loadError.value = ''
  operationError.value = ''
  try {
    const [family, invitationRows] = await Promise.all([
      getFamilyDetail(familyId.value),
      listFamilyInvitations(familyId.value)
    ])
    familyName.value = family.familyName
    familyRole.value = family.role
    invitations.value = invitationRows
  } catch (error) {
    loadError.value = apiErrorMessage(error, '发出的成员邀请加载失败。')
  } finally {
    loading.value = false
  }
}

function formatDate(value: string) {
  const time = new Date(value)
  return Number.isNaN(time.getTime()) ? value : time.toLocaleString('zh-CN')
}

function isExpired(item: Invitation) {
  return item.status === 'PENDING' && new Date(item.expiredAt).getTime() <= Date.now()
}

function displayStatus(item: Invitation) {
  return isExpired(item) ? 'EXPIRED' : item.status
}

function canRegenerate(item: Invitation) {
  return item.inviteChannel === 'SHARE_LINK'
    && (item.status === 'PENDING' || item.status === 'EXPIRED')
}

function openMembers() {
  uni.navigateTo({ url: `/pages/family/members?familyId=${encodeURIComponent(familyId.value)}` })
}

function openFamilyOverview() {
  uni.redirectTo({ url: `/pages/family/detail?familyId=${encodeURIComponent(familyId.value)}` })
}

function confirmCancel(item: Invitation) {
  uni.showModal({
    title: '取消邀请',
    content: `确定取消发给「${item.targetMemberName}」的邀请吗？`,
    success: (result) => {
      if (result.confirm) cancel(item)
    }
  })
}

async function cancel(item: Invitation) {
  actingId.value = item.invitationId
  actingType.value = 'cancel'
  operationError.value = ''
  try {
    await cancelInvitation(item.invitationId, { reason: '小程序已发邀请页面取消' })
    await loadInvitations()
  } catch (error) {
    operationError.value = apiErrorMessage(error, '取消邀请失败。')
  } finally {
    actingId.value = null
    actingType.value = ''
  }
}

function confirmRegenerate(item: Invitation) {
  uni.showModal({
    title: '重新生成邀请',
    content: '重新生成后，旧邀请链接将立即失效。是否继续？',
    success: (result) => {
      if (result.confirm) regenerate(item)
    }
  })
}

async function regenerate(item: Invitation) {
  actingId.value = item.invitationId
  actingType.value = 'regenerate'
  operationError.value = ''
  try {
    shareResult.value = await regenerateInvitation(item.invitationId)
    await loadInvitations()
  } catch (error) {
    operationError.value = apiErrorMessage(error, '重新生成邀请失败。')
  } finally {
    actingId.value = null
    actingType.value = ''
  }
}

function copyInviteLink() {
  const token = shareResult.value?.inviteToken
  if (!token) return
  uni.setClipboardData({
    data: `https://tree.bigbigboy.cn/invite/${encodeURIComponent(token)}`,
    success: () => uni.showToast({ title: '邀请链接已复制', icon: 'success' })
  })
}

function clearShareResult() {
  shareResult.value = null
}

onShareAppMessage(() => {
  const invitation = shareResult.value?.invitation
  return {
    title: invitation
      ? `${invitation.familyName} 邀请你确认「${invitation.targetMemberName}」身份并加入家谱`
      : 'Tree 家脉亲缘',
    path: shareResult.value
      ? `/pages/invite/detail?inviteToken=${encodeURIComponent(shareResult.value.inviteToken)}`
      : '/pages/home/index',
    imageUrl: '/static/share/family-invitation.jpg'
  }
})

onLoad((options) => {
  familyId.value = String(options?.familyId || '')
})
onShow(loadInvitations)
onHide(() => {
  authChecked.value = false
  loading.value = false
  invitations.value = []
})
onUnload(() => {
  shareResult.value = null
})
</script>

<style scoped>
.sent-page {
  background:
    radial-gradient(circle at 92% 0%, rgba(216, 175, 104, 0.14), transparent 260rpx),
    radial-gradient(circle at 0% 18%, rgba(24, 54, 83, 0.08), transparent 300rpx);
}

.item-head,
.invite-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16rpx;
}

.item-title,
.item-desc,
.item-meta {
  display: block;
}

.item-title {
  color: var(--tree-text-primary);
  font-size: 32rpx;
  font-weight: 800;
}

.item-desc,
.item-meta {
  margin-top: 8rpx;
}

.item-actions {
  display: flex;
  flex-direction: column;
  gap: 12rpx;
  margin-top: 18rpx;
  padding-top: 16rpx;
  border-top: 1rpx solid rgba(226, 232, 240, 0.78);
}

.share-result-card {
  border-color: rgba(255, 255, 255, 0.72);
  background:
    radial-gradient(circle at 100% 0%, rgba(216, 175, 104, 0.16), transparent 180rpx),
    linear-gradient(135deg, rgba(255, 255, 255, 0.98) 0%, rgba(244, 250, 247, 0.94) 100%);
  box-shadow: 0 18rpx 44rpx rgba(24, 54, 83, 0.08);
}

.actionable-card {
  position: relative;
  border-color: rgba(255, 255, 255, 0.72);
  background:
    radial-gradient(circle at 100% 0%, rgba(59, 110, 168, 0.12), transparent 160rpx),
    linear-gradient(135deg, rgba(255, 255, 255, 0.98) 0%, rgba(241, 246, 252, 0.94) 100%);
  box-shadow: 0 16rpx 40rpx rgba(24, 54, 83, 0.07);
}

.actionable-card::before {
  content: '';
  position: absolute;
  left: 0;
  top: 26rpx;
  bottom: 26rpx;
  width: 7rpx;
  border-radius: 0 999rpx 999rpx 0;
  background: linear-gradient(180deg, var(--tree-primary) 0%, var(--tree-accent-blue) 100%);
}

.share-result-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16rpx;
  margin-bottom: 18rpx;
}

.share-result-kicker,
.share-result-title {
  display: block;
}

.share-summary {
  margin: 18rpx 0 8rpx;
  padding: 4rpx 18rpx;
  border: 1rpx solid rgba(226, 232, 240, 0.82);
  border-radius: 22rpx;
  background: rgba(255, 255, 255, 0.78);
}

.share-message {
  padding: 16rpx 0;
  border-top: 1rpx solid var(--tree-border-subtle, #e8ece9);
}

.share-message-label,
.share-message-content {
  display: block;
}

.share-message-label {
  margin-bottom: 8rpx;
  color: var(--tree-text-secondary);
  font-size: 22rpx;
}

.share-message-content {
  color: var(--tree-text);
  font-size: 26rpx;
  line-height: 1.6;
}

.share-result-kicker {
  margin-bottom: 6rpx;
  color: var(--tree-green, #2f6b57);
  font-size: 22rpx;
  font-weight: 800;
  letter-spacing: 2rpx;
}

.share-result-title {
  color: var(--tree-text);
  font-size: 30rpx;
  font-weight: 700;
  line-height: 1.4;
}

.wechat-share-button {
  margin: 16rpx 0 12rpx;
  border: 0;
  border-radius: 16rpx;
  background: var(--tree-green, #2f6b57);
  color: #fff;
  font-size: 28rpx;
  font-weight: 600;
}

.wechat-share-button::after {
  border: 0;
}
</style>
