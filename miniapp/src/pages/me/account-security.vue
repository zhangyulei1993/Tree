<template>
  <view class="tree-page account-security-page">
    <MiniBackHome />

    <template v-if="session.isLoggedIn && session.user">
      <MiniCard>
        <MiniSectionHeader title="账号与安全" subtitle="个人资料、手机号与账号管理" accent />
        <MiniNotice v-if="showPasswordUnsetNotice" tone="info" class="password-notice">
          建议设置登录密码，方便电脑网页登录。可在绑定手机号页面设置。
        </MiniNotice>
        <MiniActionList :items="securityItems" @select="onSecuritySelect" />
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
import { computed } from 'vue'

import MiniActionList from '@/components/base/MiniActionList.vue'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniCard from '@/components/base/MiniCard.vue'
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
import MiniSectionHeader from '@/components/base/MiniSectionHeader.vue'
import { useSessionStore } from '@/stores/session'

const session = useSessionStore()

const showProfileIncompleteNotice = computed(() => {
  if (!session.user) return false
  return !session.user.nickname?.trim() || !session.user.avatarUrl
})

const showPasswordUnsetNotice = computed(() =>
  session.isPhoneBound && session.user?.passwordSet === false
)

const securityItems = computed(() => [
  {
    key: 'profile',
    title: showProfileIncompleteNotice.value ? '完善资料' : '更新资料',
    desc: showProfileIncompleteNotice.value ? '设置昵称和头像' : '更新昵称和头像'
  },
  {
    key: 'bind-phone',
    title: session.user?.phoneVerified ? '更改手机号' : '绑定手机号',
    desc: session.user?.phoneVerified ? '当前手机号已验证，可重新绑定' : '完成验证后可使用完整功能'
  }
])

function go(url: string) {
  uni.navigateTo({ url })
}

function onSecuritySelect(key: string) {
  switch (key) {
    case 'profile':
      go('/pages/me/profile')
      break
    case 'bind-phone':
      go(session.user?.phoneVerified ? '/pages/account/change-phone' : '/pages/auth/bind-phone')
      break
  }
}

onShow(() => {
  session.restoreSession()
})
</script>

<style scoped>
.account-security-page {
  background:
    radial-gradient(circle at 92% 0%, rgba(216, 175, 104, 0.14), transparent 260rpx),
    radial-gradient(circle at 0% 18%, rgba(24, 54, 83, 0.08), transparent 300rpx);
}

.password-notice {
  margin-bottom: 20rpx;
}

.danger-zone {
  margin-top: 20rpx;
  border-color: rgba(181, 71, 60, 0.2);
  background: rgba(255, 248, 246, 0.92);
}
</style>
