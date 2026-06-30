<template>
  <view class="tree-page auth-page">
    <MiniBackHome />
    <view class="tree-auth-brand">
      <text class="tree-auth-brand-title">{{ profileTitle }}</text>
      <text class="tree-auth-brand-desc">{{ profileDesc }}</text>
    </view>

    <MiniCard variant="soft" class="tree-auth-card">
      <view class="avatar-section">
        <text class="tree-field-label">头像</text>
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
        保存资料
      </MiniButton>
    </MiniCard>
  </view>
</template>

<script setup lang="ts">
import { onShow } from '@dcloudio/uni-app'
import { computed, ref } from 'vue'

import { apiErrorMessage, resolveAssetUrl } from '@/api/client'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniCard from '@/components/base/MiniCard.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
import { useSessionStore } from '@/stores/session'

const session = useSessionStore()
const nickname = ref(session.user?.nickname || '')
const localAvatarPath = ref('')
const saving = ref(false)
const uploading = ref(false)
const errorMessage = ref('')
const successMessage = ref('')

const profileComplete = computed(() =>
  Boolean(session.user?.nickname?.trim()) && Boolean(session.user?.avatarUrl)
)
const profileTitle = computed(() => (profileComplete.value ? '更新资料' : '完善资料'))
const profileDesc = computed(() =>
  profileComplete.value ? '更新昵称和头像，保持家人可识别' : '设置昵称和头像，方便家人识别你'
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

  saving.value = true
  try {
    await session.saveProfile({ nickname: value })
    await session.refreshMe().catch(() => undefined)
    successMessage.value = '昵称已保存。'
    uni.showToast({ title: '保存成功', icon: 'success' })
  } catch (error) {
    errorMessage.value = apiErrorMessage(error, '昵称保存失败，请稍后重试。')
  } finally {
    saving.value = false
  }
}

onShow(() => {
  session.restoreSession()
  if (!session.isLoggedIn) {
    uni.reLaunch({ url: '/pages/auth/wechat-login' })
    return
  }
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
</style>
