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
        <MiniSectionHeader title="我的事务" subtitle="家庭、邀请与加入申请" />
        <MiniActionList :items="affairItems" @select="onAffairSelect" />
      </MiniCard>

      <MiniCard>
        <MiniSectionHeader title="账号与安全" subtitle="手机号绑定与账号管理" />
        <MiniNotice v-if="showPasswordUnsetNotice" tone="info" class="password-notice">
          建议设置登录密码，方便电脑网页登录。
        </MiniNotice>
        <MiniActionList :items="securityItems" @select="onSecuritySelect" />
      </MiniCard>

      <MiniCard>
        <MiniSectionHeader title="协议与隐私" />
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
        <MiniSectionHeader title="协议与隐私" />
        <MiniActionList :items="legalItems" @select="onLegalSelect" />
      </MiniCard>
    </template>
  </view>
</template>

<script setup lang="ts">
import { onShow } from '@dcloudio/uni-app'
import { computed, ref } from 'vue'

import { apiErrorMessage, resolveAssetUrl } from '@/api/client'
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
  { key: 'family', title: '我的家庭', desc: '查看和管理家族资料' },
  { key: 'invite', title: '我的邀请', desc: '查看邀请状态' },
  { key: 'join', title: '我的加入申请', desc: '管理申请记录' }
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
    case 'invite':
      go('/pages/invite/my')
      break
    case 'join':
      go('/pages/join/my')
      break
  }
}

function onSecuritySelect(key: string) {
  switch (key) {
    case 'profile':
      go('/pages/me/profile')
      break
    case 'bind-phone':
      go('/pages/auth/bind-phone')
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
  margin-bottom: 20rpx;
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
</style>
