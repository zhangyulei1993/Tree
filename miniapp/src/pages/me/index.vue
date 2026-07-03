<template>
  <view class="tree-page me-page">
    <template v-if="session.isLoggedIn && session.user">
      <MiniCard variant="hero" class="profile-hero paper-surface">
        <ProfileHeader
          :name="session.user.nickname || '未设置昵称'"
          :subtitle="profileSubtitle"
          :avatar-text="avatarText"
          :avatar-url="avatarUrl"
          :tags="profileTags"
        />
      </MiniCard>

      <MiniCard v-if="showProfileIncompleteNotice" variant="soft" class="bind-notice-card">
        <MiniNotice tone="info">设置昵称后即可使用家庭、邀请与加入等功能；头像可选。</MiniNotice>
        <MiniButton class="btn-top" @click="go('/pages/me/profile?onboarding=1')">去完善资料</MiniButton>
      </MiniCard>

      <MiniCard class="directory-card directory-card--list">
        <MiniDirectoryTile
          title="我的家庭事务"
          desc="家庭、邀请与加入申请"
          icon-class="tree-symbol-home"
          @click="go('/pages/me/family-affairs')"
        />
        <MiniDirectoryTile
          title="账号与安全"
          desc="个人资料与账号管理"
          icon-class="tree-symbol-profile"
          @click="go('/pages/me/account-security')"
        />
        <MiniDirectoryTile
          title="关于 Tree"
          desc="协议、隐私与版本信息"
          icon-class="tree-symbol-doc"
          @click="go('/pages/me/about')"
        />
      </MiniCard>

      <view class="logout-wrap">
        <MiniButton
          variant="ghost"
          :disabled="loggingOut"
          :loading="loggingOut"
          @click="logout"
        >
          退出登录
        </MiniButton>
        <text v-if="errorMessage" class="tree-field-error">{{ errorMessage }}</text>
      </view>
    </template>

    <template v-else>
      <MiniCard variant="hero" class="profile-hero paper-surface">
        <ProfileHeader
          name="欢迎使用 Tree"
          subtitle="登录后可查看家庭、成员与邀请信息"
          avatar-text="访"
        />
        <view class="guest-actions">
          <MiniButton @click="go('/pages/auth/wechat-login')">微信登录</MiniButton>
        </view>
      </MiniCard>

      <MiniCard class="directory-card directory-card--list">
        <MiniDirectoryTile
          title="关于 Tree"
          desc="协议、隐私与版本信息"
          icon-class="tree-symbol-doc"
          @click="go('/pages/me/about')"
        />
      </MiniCard>
    </template>
  </view>
</template>

<script setup lang="ts">
import { onShow } from '@dcloudio/uni-app'
import { computed, ref } from 'vue'

import { apiErrorMessage, resolveAssetUrl } from '@/api/client'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniCard from '@/components/base/MiniCard.vue'
import MiniDirectoryTile from '@/components/base/MiniDirectoryTile.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
import { accountStatusText, statusTagTone } from '@/components/base/formatStatus'
import ProfileHeader from '@/components/profile/ProfileHeader.vue'
import { useSessionStore } from '@/stores/session'

const session = useSessionStore()
const loggingOut = ref(false)
const errorMessage = ref('')
const navigating = ref(false)

const profileSubtitle = computed(() =>
  session.user?.nickname?.trim() ? '微信账号已登录' : '请完善昵称以使用完整功能'
)

const avatarText = computed(() => {
  const name = session.user?.nickname?.trim()
  if (name) return name.slice(0, 1)
  return '我'
})

const avatarUrl = computed(() => resolveAssetUrl(session.user?.avatarUrl))

const showProfileIncompleteNotice = computed(() =>
  session.isLoggedIn && !session.isProfileComplete
)

const profileTags = computed(() => {
  if (!session.user) return []
  return [
    { label: accountStatusText(session.user.status), tone: statusTagTone(session.user.status) },
    {
      label: session.isProfileComplete ? '资料已完善' : '待完善昵称',
      tone: session.isProfileComplete ? 'active' : 'pending'
    }
  ] as Array<{ label: string; tone: 'active' | 'pending' | 'danger' | 'muted' }>
})

function go(url: string) {
  if (navigating.value) return
  navigating.value = true
  uni.navigateTo({
    url,
    complete: () => {
      navigating.value = false
    }
  })
}

async function logout() {
  loggingOut.value = true
  errorMessage.value = ''
  try {
    await session.logout()
    uni.showToast({ title: '已退出', icon: 'none' })
    setTimeout(() => uni.reLaunch({ url: '/pages/home/index' }), 200)
  } catch (error) {
    errorMessage.value = apiErrorMessage(error, '退出请求失败，本地登录状态已清理。')
  } finally {
    loggingOut.value = false
  }
}

onShow(() => {
  session.restoreSession()
  if (session.isLoggedIn) {
    session.refreshMe().catch(() => undefined)
  }
})
</script>

<style scoped>
.me-page {
  background:
    radial-gradient(circle at 92% 0%, rgba(216, 175, 104, 0.16), transparent 260rpx),
    radial-gradient(circle at 0% 18%, rgba(24, 54, 83, 0.08), transparent 300rpx);
}

.profile-hero {
  margin-bottom: 20rpx;
  padding-top: 34rpx;
  padding-bottom: 34rpx;
}

.profile-hero :deep(.profile-header) {
  align-items: flex-start;
}

.profile-hero :deep(.avatar) {
  background: rgba(255, 255, 255, 0.16);
  color: #fff;
  box-shadow:
    0 0 0 1rpx rgba(255, 255, 255, 0.22) inset,
    0 14rpx 34rpx rgba(0, 0, 0, 0.14);
}

.profile-hero :deep(.name) {
  color: #fff;
  font-size: 38rpx;
  font-weight: 800;
}

.profile-hero :deep(.subtitle) {
  color: rgba(255, 255, 255, 0.72);
}

.btn-top {
  margin-top: 20rpx;
}

.bind-notice-card {
  margin-bottom: 16rpx;
}

.bind-notice-card :deep(.mini-notice) {
  margin-bottom: 0;
}

.directory-card {
  margin-bottom: 12rpx;
}

.directory-card--list {
  padding-top: 8rpx;
  padding-bottom: 8rpx;
}

.guest-actions {
  display: flex;
  flex-direction: column;
  gap: 14rpx;
  margin-top: 28rpx;
}

.logout-wrap {
  padding: 8rpx 0 calc(24rpx + env(safe-area-inset-bottom));
}
</style>
