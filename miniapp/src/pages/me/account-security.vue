<template>
  <view class="archive-page account-security-page">
    <MiniBackHome />

    <view class="security-head archive-page-head">
      <view>
        <text class="archive-kicker">Account Security</text>
        <text class="archive-title">账号与安全</text>
        <text class="archive-subtitle">管理登录方式、账号权益与高风险操作</text>
      </view>
      <view class="archive-seal">安</view>
    </view>

    <template v-if="session.isLoggedIn && session.user">
      <view class="security-status archive-form-panel">
        <view class="archive-section-head">
          <text class="archive-section-title">账号状态</text>
          <text class="archive-section-subtitle">当前平台账号与信任等级</text>
        </view>
        <view class="status-chips">
          <text class="archive-chip">{{ accountStatusLabel }}</text>
          <text class="archive-chip">{{ trustTierLabel }}</text>
          <text v-if="session.isProfileComplete" class="archive-chip">资料已完善</text>
          <text v-else class="archive-chip">资料待完善</text>
        </view>
      </view>

      <view class="security-directory archive-list">
        <view class="archive-row security-row static">
          <view class="archive-row-main">
            <text class="archive-row-title">微信登录</text>
            <text class="archive-row-desc">当前通过微信使用 Tree，昵称 {{ session.user.nickname || '未设置' }}</text>
          </view>
          <text class="archive-row-meta">已登录</text>
        </view>

        <view class="archive-row security-row" @click="openProfile">
          <view class="archive-row-main">
            <text class="archive-row-title">{{ session.isProfileComplete ? '更新资料' : '完善资料' }}</text>
            <text class="archive-row-desc">
              {{ session.isProfileComplete ? '更新昵称和头像' : '设置昵称后即可使用完整功能' }}
            </text>
          </view>
          <text class="archive-arrow">›</text>
        </view>

        <view class="archive-row security-row" @click="openPhoneBackup">
          <view class="archive-row-main">
            <text class="archive-row-title">手机号备用登录</text>
            <text class="archive-row-desc">{{ phoneBackupDesc }}</text>
          </view>
          <text class="archive-row-meta">{{ phoneBackupMeta }}</text>
          <text class="archive-arrow">›</text>
        </view>
      </view>

      <view v-if="capabilities" class="quota-panel archive-form-panel">
        <view class="archive-section-head">
          <text class="archive-section-title">账号权益</text>
          <text class="archive-section-subtitle">当前可用额度与用量</text>
        </view>
        <view class="quota-lines">
          <text v-for="line in capabilityLines" :key="line" class="quota-line">{{ line }}</text>
        </view>
        <MiniNotice v-if="showQuotaHint" tone="info">
          当前扩展能力尚未开放。达到基础上限时，仅会阻止新增、恢复或加入，不影响查看与编辑已有内容。
        </MiniNotice>
      </view>

      <view class="archive-danger-panel">
        <text class="archive-danger-title">注销账号</text>
        <text class="archive-danger-desc">注销后账号数据将无法恢复，请谨慎操作。</text>
        <MiniButton size="sm" variant="danger" @click="go('/pages/account/cancel')">注销账号</MiniButton>
      </view>
    </template>

    <view v-else class="security-state archive-form-panel">
      <MiniEmptyState
        symbol="!"
        title="请先登录"
        description="登录后可管理账号与安全设置。"
        action-text="去登录"
        @action="go('/pages/auth/wechat-login')"
      />
    </view>
  </view>
</template>

<script setup lang="ts">
import { onShow } from '@dcloudio/uni-app'
import { computed, ref } from 'vue'

import { fetchCapabilities } from '@/api/capabilities'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'
import { accountStatusText } from '@/components/base/formatStatus'
import MiniNotice from '@/components/base/MiniNotice.vue'
import { buildCapabilitiesSummary } from '@/features/quota/quotaDisplay'
import { useSessionStore } from '@/stores/session'
import type { UserCapabilities } from '@/types/api'
import { maskPhone } from '@/utils/maskPhone'

const session = useSessionStore()
const capabilities = ref<UserCapabilities | null>(null)

const capabilityLines = computed(() => (capabilities.value ? buildCapabilitiesSummary(capabilities.value) : []))
const accountStatusLabel = computed(() => accountStatusText(session.user?.status))
const trustTierLabel = computed(() => {
  const tier = capabilities.value?.trustTier
  if (tier === 'PHONE_BOUND') return '已启用备用登录'
  return '微信账号'
})
const phoneBackupDesc = computed(() => {
  if (session.user?.phoneLoginEnabled) {
    const masked = maskPhone(session.user.phone)
    return masked ? `已设置备用登录 · ${masked}` : '已设置备用登录密码'
  }
  return '可选设置手机号与登录密码，无法使用微信时备用'
})
const phoneBackupMeta = computed(() => (session.user?.phoneLoginEnabled ? '已设置' : '未设置'))

const showQuotaHint = computed(() => {
  if (!capabilities.value) return false
  const { limits, usage, trustTier } = capabilities.value
  return trustTier === 'WECHAT_ONLY' && (
    usage.ownedFamilies >= limits.maxOwnedFamilies ||
    usage.joinedFamilies >= limits.maxJoinedFamilies
  )
})

function go(url: string) {
  uni.navigateTo({ url })
}

function openProfile() {
  go(session.isProfileComplete ? '/pages/me/profile' : '/pages/me/profile?onboarding=1')
}

function openPhoneBackup() {
  go('/pages/me/profile')
}

onShow(async () => {
  session.restoreSession()
  if (session.isLoggedIn) {
    try {
      capabilities.value = await fetchCapabilities()
    } catch {
      capabilities.value = null
    }
  }
})
</script>

<style scoped>
.account-security-page {
  padding-top: 28rpx;
}

.account-security-page :deep(.mini-back-home) {
  margin-bottom: 22rpx;
}

.security-status,
.quota-panel,
.security-state {
  margin-bottom: 24rpx;
}

.status-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 10rpx;
}

.security-directory {
  margin-bottom: 24rpx;
}

.security-row.static {
  cursor: default;
}

.quota-lines {
  display: flex;
  flex-direction: column;
  gap: 10rpx;
  margin-bottom: 16rpx;
}

.quota-line {
  color: var(--archive-ink-soft);
  font-size: 24rpx;
  line-height: 1.55;
}

.quota-panel :deep(.mini-notice) {
  margin-top: 4rpx;
}

.archive-danger-panel :deep(.mini-button) {
  align-self: flex-start;
}
</style>
