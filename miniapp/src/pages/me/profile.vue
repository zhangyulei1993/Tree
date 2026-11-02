<template>
  <view class="archive-page profile-page">
    <MiniBackHome v-if="!onboarding" />

    <view class="profile-head archive-page-head">
      <view>
        <text class="archive-kicker">{{ onboarding ? 'Profile Onboarding' : 'Personal File' }}</text>
        <text class="archive-title">{{ pageTitle }}</text>
        <text class="archive-subtitle">{{ profileDesc }}</text>
      </view>
      <view class="archive-seal">{{ sealLetter }}</view>
    </view>

    <view class="profile-folio">
      <view class="folio-avatar">
        <image v-if="avatarPreview" class="folio-avatar-image" :src="avatarPreview" mode="aspectFill" />
        <text v-else class="folio-avatar-text">{{ avatarText }}</text>
      </view>
      <view class="folio-copy">
        <text class="folio-name">{{ displayNickname }}</text>
        <view class="folio-meta">
          <text>平台账号</text>
          <text>ID {{ userId }}</text>
        </view>
        <text class="folio-status">{{ accountStatusLabel }}</text>
      </view>
    </view>

    <view class="archive-form-panel profile-form">
      <view class="archive-section-head">
        <text class="archive-section-title">基本资料</text>
        <text class="archive-section-subtitle">昵称和头像均可稍后补充</text>
      </view>

      <text class="tree-field-label">头像（可选）</text>
      <!-- #ifdef MP-WEIXIN -->
      <button
        class="avatar-picker"
        open-type="chooseAvatar"
        :disabled="uploading"
        @chooseavatar="onChooseAvatar"
      >
        <image v-if="avatarPreview" class="avatar-picker-image" :src="avatarPreview" mode="aspectFill" />
        <text v-else class="avatar-picker-text">选择微信头像</text>
      </button>
      <!-- #endif -->
      <!-- #ifndef MP-WEIXIN -->
      <MiniNotice tone="info">头像选择仅在微信小程序中可用。</MiniNotice>
      <view v-if="avatarPreview" class="avatar-static-wrap">
        <image class="avatar-picker-image" :src="avatarPreview" mode="aspectFill" />
      </view>
      <!-- #endif -->

      <text class="tree-field-label">昵称</text>
      <input
        v-model="nickname"
        class="tree-input"
        type="nickname"
        maxlength="30"
        placeholder="请输入昵称"
      />

      <text v-if="errorMessage" class="tree-field-error">{{ errorMessage }}</text>
      <text v-if="successMessage" class="tree-field-success">{{ successMessage }}</text>

      <MiniButton class="profile-save" :disabled="saving" :loading="saving" @click="saveNickname">
        {{ onboarding ? '保存并继续' : '保存资料' }}
      </MiniButton>
      <MiniButton
        v-if="onboarding && !profileRequired"
        class="profile-skip"
        variant="ghost"
        size="sm"
        :disabled="saving || uploading"
        @click="skipProfile"
      >
        暂时跳过，稍后再说
      </MiniButton>
    </view>

  </view>
</template>

<script setup lang="ts">
import { onLoad, onShow } from '@dcloudio/uni-app'
import { computed, ref } from 'vue'

import { apiErrorMessage, resolveAssetUrl } from '@/api/client'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import { accountStatusText } from '@/components/base/formatStatus'
import { markNewUserGuideMilestone } from '@/features/onboarding/newUserGuide'
import { isProfileComplete } from '@/features/session/profileComplete'
import { useSessionStore } from '@/stores/session'
import { validateTextFields } from '@/utils/inputValidation'

const session = useSessionStore()
const nickname = ref('')
const localAvatarPath = ref('')
const saving = ref(false)
const uploading = ref(false)
const errorMessage = ref('')
const successMessage = ref('')
const onboarding = ref(false)
const profileRequired = ref(false)

const displayNickname = computed(() => nickname.value.trim() || session.user?.nickname || '未设置昵称')
const userId = computed(() => session.user?.id || '—')
const profileComplete = computed(() => isProfileComplete(session.user))
const pageTitle = computed(() => {
  if (onboarding.value) return '完善资料'
  return profileComplete.value ? '个人资料' : '完善资料'
})
const profileDesc = computed(() =>
  onboarding.value
    ? profileRequired.value
      ? '当前操作需要可识别的资料，请先设置昵称；头像仍可稍后补充。'
      : '头像和昵称都可以稍后补充；涉及家庭身份时再按提示完善。'
    : profileComplete.value
      ? '更新昵称和头像，保持家人可识别'
      : '昵称和头像均可稍后补充。'
)
const avatarText = computed(() => {
  const name = nickname.value.trim() || session.user?.nickname?.trim()
  if (name) return name.slice(0, 1)
  return '档'
})
const sealLetter = computed(() => avatarText.value)
const accountStatusLabel = computed(() =>
  session.user ? accountStatusText(session.user.status) : ''
)

const avatarPreview = computed(() => {
  if (localAvatarPath.value) return localAvatarPath.value
  return resolveAssetUrl(session.user?.avatarUrl)
})

type ChooseAvatarEvent = {
  detail: {
    avatarUrl?: string
  }
}

async function onChooseAvatar(event: ChooseAvatarEvent) {
  errorMessage.value = ''
  successMessage.value = ''
  const filePath = event.detail.avatarUrl
  if (!filePath) {
    errorMessage.value = '未选择头像。'
    return
  }

  localAvatarPath.value = filePath
  uploading.value = true
  try {
    await session.saveAvatar(filePath)
    await session.refreshMe().catch(() => undefined)
    successMessage.value = '头像已保存。'
    uni.showToast({ title: '头像已更新', icon: 'success' })
  } catch (error) {
    errorMessage.value = apiErrorMessage(error, '头像上传失败，请稍后重试。')
  } finally {
    uploading.value = false
  }
}

async function saveNickname() {
  errorMessage.value = ''
  successMessage.value = ''
  const value = nickname.value.trim()
  if (!value) {
    errorMessage.value = '请输入昵称。'
    return
  }
  const validationMessage = validateTextFields([
    { value, label: '昵称', kind: 'name', required: true, maxLength: 30 }
  ])
  if (validationMessage) {
    errorMessage.value = validationMessage
    return
  }

  saving.value = true
  try {
    await session.saveProfile({ nickname: value })
    await session.refreshMe().catch(() => undefined)
    markNewUserGuideMilestone('PROFILE_DONE')
    successMessage.value = '昵称已保存。'
    uni.showToast({ title: '保存成功', icon: 'success' })
    if (onboarding.value) {
      setTimeout(() => session.finishProfile(), 300)
    }
  } catch (error) {
    errorMessage.value = apiErrorMessage(error, '昵称保存失败，请稍后重试。')
  } finally {
    saving.value = false
  }
}

function skipProfile() {
  if (profileRequired.value || saving.value || uploading.value) return
  uni.showToast({ title: '已跳过，可在个人资料中补充', icon: 'none' })
  setTimeout(() => session.finishProfile(), 300)
}

onLoad((options) => {
  onboarding.value = options?.onboarding === '1'
  profileRequired.value = options?.required === '1'
})

onShow(() => {
  session.restoreSession()
  if (!session.isLoggedIn) {
    uni.reLaunch({ url: '/pages/auth/wechat-login' })
    return
  }
  if (onboarding.value && session.isProfileComplete) {
    session.finishProfile()
    return
  }
  nickname.value = session.user?.nickname || ''
  session.refreshMe().catch(() => undefined)
})
</script>

<style scoped>
.profile-page {
  padding-top: 28rpx;
}

.profile-page :deep(.mini-back-home) {
  margin-bottom: 22rpx;
}

.profile-skip {
  margin-top: 10rpx;
}

.profile-folio {
  display: flex;
  align-items: center;
  gap: 24rpx;
  margin-bottom: 24rpx;
  border-top: 1rpx solid var(--archive-line-strong);
  border-bottom: 1rpx solid var(--archive-line);
  background:
    linear-gradient(90deg, rgba(255, 255, 255, 0.42), transparent 38%),
    rgba(255, 252, 245, 0.52);
  padding: 24rpx 0;
}

.folio-avatar {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 128rpx;
  height: 128rpx;
  flex-shrink: 0;
  border: 1rpx solid var(--archive-line-strong);
  border-radius: 50%;
  background:
    radial-gradient(circle at 35% 28%, rgba(255, 255, 255, 0.9), transparent 34rpx),
    #eadfc8;
  overflow: hidden;
}

.folio-avatar-image {
  width: 100%;
  height: 100%;
}

.folio-avatar-text {
  color: var(--archive-blue);
  font-family: 'Songti SC', 'STSong', serif;
  font-size: 52rpx;
  font-weight: 800;
}

.folio-copy {
  flex: 1;
  min-width: 0;
}

.folio-name {
  display: block;
  color: var(--archive-ink);
  font-family: 'Songti SC', 'STSong', serif;
  font-size: 40rpx;
  font-weight: 800;
  line-height: 1.3;
}

.folio-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 8rpx 14rpx;
  margin-top: 10rpx;
  color: var(--archive-ink-soft);
  font-size: 21rpx;
  line-height: 1.4;
}

.folio-status {
  display: block;
  margin-top: 10rpx;
  color: var(--archive-cinnabar);
  font-size: 22rpx;
  line-height: 1.4;
}

.profile-form {
  margin-bottom: 24rpx;
}

.avatar-picker {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 120rpx;
  height: 120rpx;
  margin-top: 12rpx;
  padding: 0;
  border: 1rpx solid var(--archive-line-strong);
  border-radius: 50%;
  background: rgba(255, 249, 236, 0.42);
  overflow: hidden;
}

.avatar-picker::after {
  border: none;
}

.avatar-picker-image {
  width: 100%;
  height: 100%;
}

.avatar-picker-text {
  padding: 0 12rpx;
  color: var(--archive-ink-soft);
  font-size: 20rpx;
  line-height: 1.35;
  text-align: center;
}

.avatar-static-wrap {
  width: 120rpx;
  height: 120rpx;
  margin-top: 12rpx;
  border: 1rpx solid var(--archive-line-strong);
  border-radius: 50%;
  overflow: hidden;
}

.profile-save {
  margin-top: 22rpx;
  align-self: flex-start;
}

</style>
