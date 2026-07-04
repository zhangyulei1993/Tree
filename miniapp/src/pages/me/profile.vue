<template>
  <view class="tree-page auth-page">
    <MiniBackHome v-if="!onboarding" />
    <view class="tree-auth-brand">
      <text class="tree-auth-brand-title">{{ profileTitle }}</text>
      <text class="tree-auth-brand-desc">{{ profileDesc }}</text>
    </view>

    <MiniCard variant="soft" class="tree-auth-card">
      <view class="avatar-section">
        <text class="tree-field-label">头像（可选）</text>
        <!-- #ifdef MP-WEIXIN -->
        <button class="avatar-picker" open-type="chooseAvatar" :disabled="uploading" @chooseavatar="onChooseAvatar">
          <image v-if="avatarPreview" class="avatar-image" :src="avatarPreview" mode="aspectFill" />
          <view v-else class="avatar-placeholder">选头像</view>
        </button>
        <!-- #endif -->
        <!-- #ifndef MP-WEIXIN -->
        <MiniNotice tone="info">头像选择仅在微信小程序中可用。</MiniNotice>
        <image v-if="avatarPreview" class="avatar-image avatar-image-static" :src="avatarPreview" mode="aspectFill" />
        <!-- #endif -->
      </view>

      <text class="tree-field-label">昵称</text>
      <input
        v-model="nickname"
        class="input"
        type="nickname"
        maxlength="30"
        placeholder="请输入昵称"
      />

      <text v-if="errorMessage" class="tree-field-error">{{ errorMessage }}</text>
      <text v-if="successMessage" class="tree-field-success">{{ successMessage }}</text>

      <MiniButton class="btn-top" :disabled="saving" :loading="saving" @click="saveNickname">
        {{ onboarding ? '完成并继续' : '保存资料' }}
      </MiniButton>

      <view v-if="showPhoneBackupSection" class="phone-backup-section">
        <text class="tree-section-title">备用登录（可选）</text>
        <text class="tree-muted phone-backup-desc">
          设置手机号与登录密码，可在无法使用微信时登录。不使用短信验证，可随时跳过。
        </text>

        <template v-if="phoneLoginEnabled">
          <text class="tree-field-label">已绑定手机号</text>
          <text class="phone-mask">{{ maskedPhone }}</text>

          <text class="tree-field-label">当前密码</text>
          <input
            v-model="currentPassword"
            class="input"
            password
            maxlength="64"
            placeholder="请输入当前登录密码"
          />
          <text class="tree-field-label">新密码</text>
          <input
            v-model="newPassword"
            class="input"
            password
            maxlength="64"
            placeholder="至少 6 位"
          />
          <MiniButton
            class="btn-top"
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
            class="input"
            type="number"
            maxlength="11"
            placeholder="请输入手机号"
          />
          <text class="tree-field-label">登录密码</text>
          <input
            v-model="bindPassword"
            class="input"
            password
            maxlength="64"
            placeholder="至少 6 位"
          />
          <text class="tree-field-label">确认密码</text>
          <input
            v-model="bindConfirmPassword"
            class="input"
            password
            maxlength="64"
            placeholder="再次输入登录密码"
          />
          <MiniButton
            class="btn-top"
            variant="secondary"
            :disabled="bindingPhone"
            :loading="bindingPhone"
            @click="submitPhoneBind"
          >
            设置备用登录
          </MiniButton>
          <MiniButton
            v-if="onboarding"
            class="btn-top"
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
    </MiniCard>
  </view>
</template>

<script setup lang="ts">
import { onLoad, onShow } from '@dcloudio/uni-app'
import { computed, ref } from 'vue'

import { apiErrorMessage, resolveAssetUrl } from '@/api/client'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniCard from '@/components/base/MiniCard.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
import { isProfileComplete } from '@/features/session/profileComplete'
import { useSessionStore } from '@/stores/session'
import { maskPhone } from '@/utils/maskPhone'

const session = useSessionStore()
const nickname = ref(session.user?.nickname || '')
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

const profileComplete = computed(() => isProfileComplete(session.user))
const profileTitle = computed(() => (onboarding.value ? '完善资料' : profileComplete.value ? '更新资料' : '完善资料'))
const profileDesc = computed(() =>
  onboarding.value
    ? '设置昵称后即可使用家庭、邀请与加入等功能；头像可选。'
    : profileComplete.value
      ? '更新昵称和头像，保持家人可识别'
      : '设置昵称后即可使用完整功能；头像可选。'
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
.avatar-section {
  margin-bottom: 24rpx;
}

.avatar-picker {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 160rpx;
  height: 160rpx;
  margin-top: 12rpx;
  padding: 0;
  border: none;
  border-radius: 50%;
  background: #eef4f1;
  overflow: hidden;
}

.avatar-picker::after {
  border: none;
}

.avatar-image,
.avatar-image-static {
  width: 160rpx;
  height: 160rpx;
  border-radius: 50%;
}

.avatar-placeholder {
  color: #2f6b4f;
  font-size: 26rpx;
}

.btn-top {
  margin-top: 28rpx;
}

.phone-backup-section {
  margin-top: 36rpx;
  padding-top: 28rpx;
  border-top: 1rpx solid var(--tree-border-subtle);
}

.phone-backup-desc {
  display: block;
  margin-top: 8rpx;
}

.phone-mask {
  display: block;
  margin-top: 8rpx;
  color: var(--tree-text-primary);
  font-size: 30rpx;
  font-weight: 600;
}
</style>
