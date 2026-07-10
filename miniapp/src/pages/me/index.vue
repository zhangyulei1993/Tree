<template>
  <view class="archive-page me-page">
    <template v-if="session.isLoggedIn && session.user">
      <view class="profile-archive">
        <view class="avatar-wrap">
          <image v-if="avatarUrl" class="avatar-image" :src="avatarUrl" mode="aspectFill" />
          <text v-else class="avatar-text">{{ avatarText }}</text>
        </view>
        <view class="profile-copy">
          <text class="archive-kicker">Personal File</text>
          <text class="profile-name">{{ session.user.nickname || '未设置昵称' }}</text>
          <view class="profile-meta">
            <text>世代待补</text>
            <text>平台账号</text>
            <text>ID {{ session.user.id }}</text>
          </view>
          <text class="profile-status">{{ accountStatusLabel }}</text>
        </view>
      </view>

      <view v-if="showProfileIncompleteNotice" class="profile-notice archive-panel">
        <MiniNotice tone="info">设置昵称后即可使用家庭、邀请与加入等功能；头像可选。</MiniNotice>
        <MiniButton class="btn-top" @click="go('/pages/me/profile?onboarding=1')">去完善资料</MiniButton>
      </view>

      <view class="me-directory archive-list">
        <view
          v-for="item in directoryItems"
          :key="item.title"
          class="archive-row"
          @click="go(item.url)"
        >
          <view class="archive-row-main">
            <text class="archive-row-title">{{ item.title }}</text>
            <text class="archive-row-desc">{{ item.desc }}</text>
          </view>
          <text v-if="item.count" class="archive-row-meta">{{ item.count }}</text>
          <text class="archive-arrow">›</text>
        </view>
      </view>

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
      <view class="profile-archive guest">
        <view class="avatar-wrap">
          <text class="avatar-text">访</text>
        </view>
        <view class="profile-copy">
          <text class="archive-kicker">Tree / Me</text>
          <text class="profile-name">欢迎使用 Tree</text>
          <text class="profile-status">登录后可查看家庭、成员与邀请信息</text>
        </view>
      </view>

      <view class="guest-actions">
        <MiniButton @click="go('/pages/auth/wechat-login')">去登录</MiniButton>
      </view>

      <view class="me-directory archive-list">
        <view class="archive-row" @click="go('/pages/me/about')">
          <view class="archive-row-main">
            <text class="archive-row-title">关于我们</text>
            <text class="archive-row-desc">协议、隐私与版本信息</text>
          </view>
          <text class="archive-arrow">›</text>
        </view>
      </view>
    </template>
  </view>
</template>

<script setup lang="ts">
import { onShow } from '@dcloudio/uni-app'
import { computed, ref } from 'vue'

import { apiErrorMessage, resolveAssetUrl } from '@/api/client'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
import { accountStatusText } from '@/components/base/formatStatus'
import { useSessionStore } from '@/stores/session'

const session = useSessionStore()
const loggingOut = ref(false)
const errorMessage = ref('')
const navigating = ref(false)

const avatarText = computed(() => {
  const name = session.user?.nickname?.trim()
  if (name) return name.slice(0, 1)
  return '我'
})

const avatarUrl = computed(() => resolveAssetUrl(session.user?.avatarUrl))

const showProfileIncompleteNotice = computed(() =>
  session.isLoggedIn && !session.isProfileComplete
)

const accountStatusLabel = computed(() =>
  session.user ? accountStatusText(session.user.status) : ''
)

const directoryItems = [
  { title: '个人资料', desc: '昵称、头像与基础资料', count: '', url: '/pages/me/profile' },
  { title: '关联成员', desc: '查看家庭事务与成员身份', count: '1', url: '/pages/me/family-affairs' },
  { title: '我的贡献', desc: '整理家庭树、邀请亲友的记录', count: '', url: '/pages/me/family-affairs' },
  { title: '家庭树记录', desc: '家庭邀请、加入申请与事务', count: '', url: '/pages/me/family-affairs' },
  { title: '设置', desc: '账号与安全、手机号登录', count: '', url: '/pages/me/account-security' },
  { title: '帮助与反馈', desc: '使用说明与联系方式', count: '', url: '/pages/me/about' },
  { title: '关于我们', desc: '协议、隐私与版本信息', count: '', url: '/pages/me/about' }
]

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
  if (loggingOut.value) return
  loggingOut.value = true
  errorMessage.value = ''
  try {
    await session.logout()
    uni.showToast({ title: '已退出登录', icon: 'success' })
    setTimeout(() => uni.reLaunch({ url: '/pages/home/index' }), 200)
  } catch (error) {
    errorMessage.value = apiErrorMessage(error, '退出登录失败。')
  } finally {
    loggingOut.value = false
  }
}

onShow(() => {
  session.restoreSession()
})
</script>

<style scoped>
.me-page {
  padding-top: 38rpx;
}

.profile-archive {
  display: flex;
  align-items: center;
  gap: 26rpx;
  margin-bottom: 34rpx;
  border-bottom: 1rpx solid var(--archive-line-strong);
  padding-bottom: 30rpx;
}

.profile-archive.guest {
  align-items: flex-start;
}

.avatar-wrap {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 136rpx;
  height: 136rpx;
  flex-shrink: 0;
  border: 1rpx solid var(--archive-line-strong);
  border-radius: 50%;
  background:
    radial-gradient(circle at 35% 28%, rgba(255, 255, 255, 0.9), transparent 34rpx),
    #eadfc8;
  overflow: hidden;
}

.avatar-image {
  width: 100%;
  height: 100%;
}

.avatar-text {
  color: var(--archive-blue);
  font-family: 'Songti SC', 'STSong', serif;
  font-size: 56rpx;
  font-weight: 800;
}

.profile-copy {
  flex: 1;
  min-width: 0;
}

.profile-name {
  display: block;
  margin-top: 8rpx;
  color: var(--archive-ink);
  font-family: 'Songti SC', 'STSong', serif;
  font-size: 42rpx;
  font-weight: 800;
  line-height: 1.3;
}

.profile-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 8rpx 14rpx;
  margin-top: 12rpx;
  color: var(--archive-ink-soft);
  font-size: 21rpx;
  line-height: 1.4;
}

.profile-status {
  display: block;
  margin-top: 10rpx;
  color: var(--archive-cinnabar);
  font-size: 22rpx;
  line-height: 1.4;
}

.profile-notice {
  margin-bottom: 24rpx;
  padding: 22rpx 0;
}

.profile-notice :deep(.mini-notice) {
  margin-bottom: 14rpx;
}

.btn-top {
  margin-top: 6rpx;
}

.guest-actions {
  margin: 26rpx 0;
}

.guest-actions :deep(.mini-button),
.logout-wrap :deep(.mini-button) {
  border-radius: 0;
  box-shadow: none;
}

.me-directory {
  margin-top: 10rpx;
}

.logout-wrap {
  padding: 26rpx 0 calc(24rpx + env(safe-area-inset-bottom));
}
</style>
