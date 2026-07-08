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
        <text class="archive-section-subtitle">昵称必填，头像可选</text>
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
        {{ onboarding ? '完成并继续' : '保存资料' }}
      </MiniButton>
    </view>

    <view v-if="showPhoneBackupSection" class="archive-form-panel phone-backup-section">
      <view class="archive-section-head">
        <text class="archive-section-title">备用登录（可选）</text>
        <text class="archive-section-subtitle">
          设置手机号与登录密码，可在无法使用微信时登录。不使用短信验证，可随时跳过。
        </text>
      </view>

      <template v-if="phoneLoginEnabled">
        <text class="tree-field-label">已绑定手机号</text>
        <text class="phone-mask">{{ maskedPhone }}</text>

        <text class="tree-field-label">当前密码</text>
        <input
          v-model="currentPassword"
          class="tree-input"
          password
          maxlength="64"
          placeholder="请输入当前登录密码"
        />
        <text class="tree-field-label">新密码</text>
        <input
          v-model="newPassword"
          class="tree-input"
          password
          maxlength="64"
          placeholder="至少 6 位"
        />
        <MiniButton
          class="profile-action"
          size="sm"
          variant="secondary"
          :disabled="changingPassword"
          :loading="changingPassword"
          @click="submitPasswordChange"
        >
          更新登录密码
        </MiniButton>
      </template>

      <template v-else>
        <text class="tree-field-label">手机号</text>
        <input
          v-model.trim="bindPhone"
          class="tree-input"
          type="number"
          maxlength="11"
          placeholder="请输入手机号"
        />
        <text class="tree-field-label">登录密码</text>
        <input
          v-model="bindPassword"
          class="tree-input"
          password
          maxlength="64"
          placeholder="至少 6 位"
        />
        <text class="tree-field-label">确认密码</text>
        <input
          v-model="bindConfirmPassword"
          class="tree-input"
          password
          maxlength="64"
          placeholder="再次输入登录密码"
        />
        <MiniButton
          class="profile-action"
          size="sm"
          variant="secondary"
          :disabled="bindingPhone"
          :loading="bindingPhone"
          @click="submitPhoneBind"
        >
          设置备用登录
        </MiniButton>
        <MiniButton
          v-if="onboarding"
          class="profile-action"
          size="sm"
          variant="ghost"
          :disabled="bindingPhone || saving"
          @click="skipPhoneBackup"
        >
          暂时跳过
        </MiniButton>
      </template>

      <text v-if="phoneErrorMessage" class="tree-field-error">{{ phoneErrorMessage }}</text>
      <text v-if="phoneSuccessMessage" class="tree-field-success">{{ phoneSuccessMessage }}</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { onLoad, onShow } from '@dcloudio/uni-app'
import { computed, ref } from 'vue'

import { apiErrorMessage, resolveAssetUrl } from '@/api/client'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
import { accountStatusText } from '@/components/base/formatStatus'
import { isProfileComplete } from '@/features/session/profileComplete'
import { useSessionStore } from '@/stores/session'
import { validateTextFields } from '@/utils/inputValidation'
import { maskPhone } from '@/utils/maskPhone'

const session = useSessionStore()
const nickname = ref('')
const localAvatarPath = ref('')
const saving = ref(false)
const uploading = ref(false)
const errorMessage = ref('')
const successMessage = ref('')
const phoneErrorMessage = ref('')
const phoneSuccessMessage = ref('')
const onboarding = ref(false)
const bindPhone = ref('')
const bindPassword = ref('')
const bindConfirmPassword = ref('')
const currentPassword = ref('')
const newPassword = ref('')
const bindingPhone = ref(false)
const changingPassword = ref(false)

const displayNickname = computed(() => nickname.value.trim() || session.user?.nickname || '未设置昵称')
const userId = computed(() => session.user?.id || '—')
const profileComplete = computed(() => isProfileComplete(session.user))
const pageTitle = computed(() => {
  if (onboarding.value) return '完善资料'
  return profileComplete.value ? '个人资料' : '完善资料'
})
const profileDesc = computed(() =>
  onboarding.value
    ? '设置昵称后即可使用家庭、邀请与加入等功能；头像可选。'
    : profileComplete.value
      ? '更新昵称和头像，保持家人可识别'
      : '设置昵称后即可使用完整功能；头像可选。'
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

const phoneLoginEnabled = computed(() => Boolean(session.user?.phoneLoginEnabled))
const showPhoneBackupSection = computed(() => profileComplete.value && !onboarding.value)
const maskedPhone = computed(() => maskPhone(session.user?.phone))

function validatePhone(value: string) {
  return /^1[3-9]\d{9}$/.test(value)
}

function validatePassword(value: string) {
  return value.length >= 6
}

function clearPhoneMessages() {
  phoneErrorMessage.value = ''
  phoneSuccessMessage.value = ''
}

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
  clearPhoneMessages()
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

async function submitPhoneBind() {
  clearPhoneMessages()
  errorMessage.value = ''
  successMessage.value = ''

  const phoneValue = bindPhone.value.trim()
  if (!validatePhone(phoneValue)) {
    phoneErrorMessage.value = '请输入有效的手机号。'
    return
  }
  if (!validatePassword(bindPassword.value)) {
    phoneErrorMessage.value = '登录密码至少 6 位。'
    return
  }
  if (bindPassword.value !== bindConfirmPassword.value) {
    phoneErrorMessage.value = '两次输入的密码不一致。'
    return
  }

  bindingPhone.value = true
  try {
    await session.bindPhoneCredential({
      phone: phoneValue,
      password: bindPassword.value,
      confirmPassword: bindConfirmPassword.value
    })
    bindPhone.value = ''
    bindPassword.value = ''
    bindConfirmPassword.value = ''
    phoneSuccessMessage.value = '备用登录已设置。'
    uni.showToast({ title: '设置成功', icon: 'success' })
    if (onboarding.value) {
      setTimeout(() => session.finishProfile(), 300)
    }
  } catch (error) {
    phoneErrorMessage.value = apiErrorMessage(error, '备用登录设置失败，请稍后重试。')
  } finally {
    bindingPhone.value = false
  }
}

async function submitPasswordChange() {
  clearPhoneMessages()
  if (!currentPassword.value || !validatePassword(newPassword.value)) {
    phoneErrorMessage.value = '请填写当前密码，且新密码至少 6 位。'
    return
  }

  changingPassword.value = true
  try {
    await session.changePhoneLoginPassword({
      currentPassword: currentPassword.value,
      newPassword: newPassword.value
    })
    currentPassword.value = ''
    newPassword.value = ''
    phoneSuccessMessage.value = '登录密码已更新。'
    uni.showToast({ title: '密码已更新', icon: 'success' })
  } catch (error) {
    phoneErrorMessage.value = apiErrorMessage(error, '密码更新失败，请稍后重试。')
  } finally {
    changingPassword.value = false
  }
}

function skipPhoneBackup() {
  if (onboarding.value) {
    session.finishProfile()
  }
}

onLoad((options) => {
  onboarding.value = options?.onboarding === '1'
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

.profile-form,
.phone-backup-section {
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

.profile-action {
  margin-top: 16rpx;
  align-self: flex-start;
}

.phone-mask {
  display: block;
  margin-top: 8rpx;
  color: var(--archive-ink);
  font-size: 28rpx;
  font-weight: 650;
}
</style>
