<template>
  <view class="tree-page account-security-page">
    <MiniBackHome />

    <template v-if="session.isLoggedIn && session.user">
      <MiniCard>
        <MiniSectionHeader title="账号与安全" subtitle="个人资料与账号管理" accent />
        <MiniActionList :items="securityItems" @select="onSecuritySelect" />
      </MiniCard>

      <MiniCard v-if="capabilities" class="quota-card">
        <MiniSectionHeader title="账号权益" subtitle="当前可用额度与用量" />
        <view class="quota-lines">
          <text v-for="line in capabilityLines" :key="line" class="quota-line">{{ line }}</text>
        </view>
        <MiniNotice v-if="showQuotaHint" tone="info">当前扩展能力尚未开放。达到基础上限时，仅会阻止新增、恢复或加入，不影响查看与编辑已有内容。</MiniNotice>
      </MiniCard>

      <MiniCard class="danger-zone">
        <MiniSectionHeader title="危险操作" subtitle="请谨慎操作，部分操作不可撤销" />
        <MiniButton variant="danger" @click="go('/pages/account/cancel')">注销账号</MiniButton>
      </MiniCard>
    </template>

    <MiniCard v-else>
      <MiniEmptyState
        symbol="!"
        title="请先登录"
        description="登录后可管理账号与安全设置。"
        action-text="去登录"
        @action="go('/pages/auth/wechat-login')"
      />
    </MiniCard>
  </view>
</template>

<script setup lang="ts">
import { onShow } from '@dcloudio/uni-app'
import { computed, ref } from 'vue'

import { fetchCapabilities } from '@/api/capabilities'
import MiniActionList from '@/components/base/MiniActionList.vue'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniCard from '@/components/base/MiniCard.vue'
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
import MiniSectionHeader from '@/components/base/MiniSectionHeader.vue'
import { buildCapabilitiesSummary } from '@/features/quota/quotaDisplay'
import type { UserCapabilities } from '@/types/api'
import { useSessionStore } from '@/stores/session'

const session = useSessionStore()
const capabilities = ref<UserCapabilities | null>(null)

const capabilityLines = computed(() => (capabilities.value ? buildCapabilitiesSummary(capabilities.value) : []))
const showQuotaHint = computed(() => {
  if (!capabilities.value) return false
  const { limits, usage, trustTier } = capabilities.value
  return trustTier === 'WECHAT_ONLY' && (
    usage.ownedFamilies >= limits.maxOwnedFamilies ||
    usage.joinedFamilies >= limits.maxJoinedFamilies
  )
})

const securityItems = computed(() => [
  {
    key: 'profile',
    title: session.isProfileComplete ? '更新资料' : '完善资料',
    desc: session.isProfileComplete ? '更新昵称和头像' : '设置昵称后即可使用完整功能'
  }
])

function go(url: string) {
  uni.navigateTo({ url })
}

function onSecuritySelect(key: string) {
  if (key === 'profile') {
    go(session.isProfileComplete ? '/pages/me/profile' : '/pages/me/profile?onboarding=1')
  }
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
  background:
    radial-gradient(circle at 92% 0%, rgba(216, 175, 104, 0.14), transparent 260rpx),
    radial-gradient(circle at 0% 18%, rgba(24, 54, 83, 0.08), transparent 300rpx);
}

.danger-zone {
  margin-top: 20rpx;
  border-color: rgba(181, 71, 60, 0.2);
  background: rgba(255, 248, 246, 0.92);
}

.quota-card {
  margin-top: 20rpx;
}

.quota-lines {
  display: flex;
  flex-direction: column;
  gap: 12rpx;
  margin-bottom: 16rpx;
}

.quota-line {
  color: #4f5d6b;
  font-size: 28rpx;
}
</style>
