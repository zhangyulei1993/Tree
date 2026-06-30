<template>
  <view class="tree-page">
    <template v-if="session.isLoggedIn && session.user">
      <MiniCard variant="hero" class="profile-hero paper-surface">
        <ProfileHeader
          :name="session.user.nickname || '未设置昵称'"
          :subtitle="maskedPhone"
          :avatar-text="avatarText"
          :avatar-url="avatarUrl"
          :tags="profileTags"
        />
      </MiniCard>

      <MiniCard v-if="showProfileIncompleteNotice" variant="soft" class="bind-notice-card">
        <MiniNotice tone="info">
          完善头像和昵称，方便家人识别你。
        </MiniNotice>
        <MiniButton class="btn-top" @click="go('/pages/me/profile')">去完善资料</MiniButton>
      </MiniCard>

      <MiniCard v-if="showPhoneBindNotice" variant="soft" class="bind-notice-card">
        <MiniNotice tone="warm">
          请绑定手机号以使用家庭、邀请、加入申请等功能。
        </MiniNotice>
        <MiniButton class="btn-top" @click="go('/pages/auth/bind-phone')">去绑定手机号</MiniButton>
      </MiniCard>

      <MiniCard>
        <MiniSectionHeader
          title="我的家庭事务"
          subtitle="查看所属家庭和个人申请进度"
          accent
        />
        <MiniActionList :items="affairItems" @select="onAffairSelect" />
      </MiniCard>

      <MiniCard>
        <MiniSectionHeader title="账号与安全" subtitle="手机号绑定与账号管理" accent />
        <MiniNotice v-if="showPasswordUnsetNotice" tone="info" class="password-notice">
          建议设置登录密码，方便电脑网页登录。
        </MiniNotice>
        <MiniActionList :items="securityItems" @select="onSecuritySelect" />
      </MiniCard>

      <MiniCard>
        <MiniSectionHeader title="协议与隐私" accent />
        <MiniActionList :items="legalItems" @select="onLegalSelect" />
        <MiniButton
          variant="secondary"
          class="btn-top"
          :disabled="loggingOut"
          :loading="loggingOut"
          @click="logout"
        >
          退出登录
        </MiniButton>
        <text v-if="errorMessage" class="tree-field-error">{{ errorMessage }}</text>
      </MiniCard>

      <MiniCard variant="soft" class="build-card">
        <MiniSectionHeader title="版本信息" subtitle="用于确认当前小程序包和接口环境" accent />
        <view class="build-grid">
          <view>
            <text>小程序</text>
            <text>{{ buildInfo.version }}</text>
          </view>
          <view>
            <text>提交</text>
            <text>{{ buildInfo.commit }}{{ buildInfo.dirty ? ' dirty' : '' }}</text>
          </view>
          <view>
            <text>环境</text>
            <text>{{ buildInfo.environment }}</text>
          </view>
          <view>
            <text>接口</text>
            <text>{{ buildInfo.apiBaseUrl || '未配置' }}</text>
          </view>
        </view>
      </MiniCard>
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
          <MiniButton variant="secondary" @click="go('/pages/auth/phone-login')">手机号登录</MiniButton>
          <MiniButton variant="ghost" @click="go('/pages/auth/register-phone')">
            注册账号
          </MiniButton>
        </view>
      </MiniCard>

      <MiniCard>
        <MiniSectionHeader title="协议与隐私" accent />
        <MiniActionList :items="legalItems" @select="onLegalSelect" />
      </MiniCard>

      <MiniCard variant="soft" class="build-card">
        <MiniSectionHeader title="版本信息" subtitle="用于确认当前小程序包和接口环境" accent />
        <view class="build-grid">
          <view>
            <text>小程序</text>
            <text>{{ buildInfo.version }}</text>
          </view>
          <view>
            <text>提交</text>
            <text>{{ buildInfo.commit }}{{ buildInfo.dirty ? ' dirty' : '' }}</text>
          </view>
          <view>
            <text>环境</text>
            <text>{{ buildInfo.environment }}</text>
          </view>
          <view>
            <text>接口</text>
            <text>{{ buildInfo.apiBaseUrl || '未配置' }}</text>
          </view>
        </view>
      </MiniCard>
    </template>
  </view>
</template>

<script setup lang="ts">
import { onShow } from '@dcloudio/uni-app'
import { computed, ref } from 'vue'

import { apiErrorMessage, resolveAssetUrl } from '@/api/client'
import { buildInfo } from '@/buildInfo'
import MiniActionList from '@/components/base/MiniActionList.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniCard from '@/components/base/MiniCard.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
import MiniSectionHeader from '@/components/base/MiniSectionHeader.vue'
import { accountStatusText, statusTagTone } from '@/components/base/formatStatus'
import ProfileHeader from '@/components/profile/ProfileHeader.vue'
import { useSessionStore } from '@/stores/session'

const session = useSessionStore()
const loggingOut = ref(false)
const errorMessage = ref('')

const maskedPhone = computed(() => {
  const phone = session.user?.phone || ''
  if (!phone) return '未绑定手机号'
  if (phone.includes('*')) return phone
  return phone.length === 11 ? `${phone.slice(0, 3)}****${phone.slice(-4)}` : phone
})

const avatarText = computed(() => {
  const name = session.user?.nickname?.trim()
  if (name) return name.slice(0, 1)
  return '我'
})

const avatarUrl = computed(() => resolveAssetUrl(session.user?.avatarUrl))

const showProfileIncompleteNotice = computed(() => {
  if (!session.isLoggedIn || !session.user) return false
  const nicknameMissing = !session.user.nickname?.trim()
  const avatarMissing = !session.user.avatarUrl
  return nicknameMissing || avatarMissing
})

const profileTags = computed(() => {
  if (!session.user) return []
  return [
    {
      label: accountStatusText(session.user.status),
      tone: statusTagTone(session.user.status)
    },
    {
      label: session.user.phoneVerified ? '手机号已验证' : '手机号未验证',
      tone: session.user.phoneVerified ? 'active' : 'pending'
    }
  ] as Array<{ label: string; tone: 'active' | 'pending' | 'danger' | 'muted' }>
})

const showPasswordUnsetNotice = computed(() =>
  session.isPhoneBound && session.user?.passwordSet === false
)

const showPhoneBindNotice = computed(() =>
  session.isLoggedIn && !session.isPhoneBound
)

const affairItems = [
  { key: 'family', title: '我的家庭', desc: '查看和管理家庭资料' },
  { key: 'affairs', title: '邀请与加入申请', desc: '处理收到的邀请，查看我提交的申请' }
]

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
  },
  { key: 'cancel', title: '注销账号', desc: '注销后账号将不可继续登录', danger: true }
])

const legalItems = [
  { key: 'user-agreement', title: '用户协议', desc: '查看' },
  { key: 'privacy-policy', title: '隐私政策', desc: '查看' }
]

function go(url: string) {
  uni.navigateTo({ url })
}

function onAffairSelect(key: string) {
  switch (key) {
    case 'family':
      go('/pages/family/my')
      break
    case 'affairs':
      go('/pages/me/family-affairs')
      break
  }
}

function onSecuritySelect(key: string) {
  switch (key) {
    case 'profile':
      go('/pages/me/profile')
      break
    case 'bind-phone':
      go(session.user?.phoneVerified ? '/pages/account/change-phone' : '/pages/auth/bind-phone')
      break
    case 'cancel':
      go('/pages/account/cancel')
      break
  }
}

function onLegalSelect(key: string) {
  switch (key) {
    case 'user-agreement':
      go('/pages/legal/user-agreement')
      break
    case 'privacy-policy':
      go('/pages/legal/privacy-policy')
      break
  }
}

async function logout() {
  loggingOut.value = true
  errorMessage.value = ''
  try {
    await session.logout()
    uni.showToast({ title: '已退出', icon: 'none' })
    setTimeout(() => uni.reLaunch({ url: '/pages/auth/phone-login' }), 200)
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
.profile-hero {
  margin-bottom: 26rpx;
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
  margin-top: 24rpx;
}

.password-notice {
  margin-bottom: 20rpx;
}

.bind-notice-card {
  margin-bottom: 20rpx;
}

.guest-actions {
  display: flex;
  flex-direction: column;
  gap: 14rpx;
  margin-top: 28rpx;
}

.bind-notice-card :deep(.mini-notice) {
  margin-bottom: 20rpx;
}

.build-card {
  margin-top: 20rpx;
}

.build-grid {
  display: grid;
  gap: 14rpx;
}

.build-grid view {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 20rpx;
  border-bottom: 1rpx solid rgba(148, 163, 184, 0.14);
  padding-bottom: 14rpx;
}

.build-grid view:last-child {
  border-bottom: 0;
  padding-bottom: 0;
}

.build-grid text:first-child {
  flex: 0 0 96rpx;
  color: var(--tree-text-muted);
  font-size: 24rpx;
}

.build-grid text:last-child {
  min-width: 0;
  color: var(--tree-text);
  font-size: 24rpx;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  line-height: 1.45;
  overflow-wrap: anywhere;
  text-align: right;
}

.tree-page {
  background:
    radial-gradient(circle at 92% 0%, rgba(216, 175, 104, 0.16), transparent 260rpx),
    radial-gradient(circle at 0% 18%, rgba(24, 54, 83, 0.08), transparent 300rpx);
}
</style>
