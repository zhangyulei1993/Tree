<template>
  <view v-if="visible" class="privacy-modal-mask" @click.stop>
    <view class="privacy-modal card">
      <text class="privacy-modal-title">个人信息保护提示</text>
      <text class="privacy-modal-body">
        欢迎使用 Tree。我们将按照法律法规保护你的个人信息。请阅读并理解《用户协议》和《隐私政策》后决定是否同意。
      </text>
      <view class="privacy-links">
        <text class="privacy-link" @click="openUserAgreement">《用户协议》</text>
        <text class="privacy-link" @click="openPrivacyPolicy">《隐私政策》</text>
      </view>
      <view class="privacy-check-row" @click="toggleChecked">
        <checkbox :checked="checked" color="#2F6B57" @click.stop="toggleChecked" />
        <text>我已阅读并同意上述协议与隐私政策</text>
      </view>
      <view class="privacy-actions">
        <MiniButton variant="secondary" @click="decline">仅浏览基础功能</MiniButton>
        <MiniButton :disabled="!checked" @click="agree">同意并继续</MiniButton>
      </view>
      <text class="privacy-footnote">
        拒绝同意仍可浏览首页、展示家庭页面、经分享链接访问的公开家庭主页与本地亲属称谓工具；微信登录、完善昵称后加入家庭等功能需同意后方可使用。
      </text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'

import MiniButton from '@/components/base/MiniButton.vue'
import { hasPrivacyConsent, openPrivacyPolicy, openUserAgreement, promptPrivacyConsentIfNeeded, requestWechatPrivacyAuthorize, resolvePrivacyAuthorizeError, setPrivacyConsent, subscribePrivacyConsentPrompt } from '@/features/legal/privacyConsent'

const visible = ref(false)
const checked = ref(false)

function toggleChecked() {
  checked.value = !checked.value
}

function showPrompt() {
  if (!hasPrivacyConsent()) {
    visible.value = true
    checked.value = false
  }
}

function decline() {
  visible.value = false
}

async function agree() {
  if (!checked.value) return
  try {
    await requestWechatPrivacyAuthorize()
  } catch (error) {
    uni.showToast({ title: resolvePrivacyAuthorizeError(error), icon: 'none' })
    return
  }
  setPrivacyConsent()
  visible.value = false
}

let unsubscribe: (() => void) | null = null

onMounted(() => {
  unsubscribe = subscribePrivacyConsentPrompt(showPrompt)
  promptPrivacyConsentIfNeeded()
})

onUnmounted(() => {
  unsubscribe?.()
})

defineExpose({
  visible
})
</script>

<style scoped>
.privacy-modal-mask {
  position: fixed;
  inset: 0;
  z-index: 9999;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40rpx;
  background: rgba(24, 54, 83, 0.42);
}

.privacy-modal {
  width: 100%;
  max-width: 640rpx;
  padding: 36rpx 30rpx;
}

.privacy-modal-title {
  display: block;
  font-size: 34rpx;
  font-weight: 700;
  color: var(--tree-text-primary);
}

.privacy-modal-body {
  display: block;
  margin-top: 18rpx;
  color: var(--tree-text-secondary);
  font-size: 26rpx;
  line-height: 1.8;
}

.privacy-links {
  display: flex;
  gap: 20rpx;
  margin-top: 18rpx;
}

.privacy-link {
  color: var(--tree-green);
  font-size: 26rpx;
}

.privacy-check-row {
  display: flex;
  align-items: flex-start;
  gap: 12rpx;
  margin-top: 24rpx;
  color: var(--tree-text-secondary);
  font-size: 24rpx;
  line-height: 1.6;
}

.privacy-actions {
  display: flex;
  flex-direction: column;
  gap: 14rpx;
  margin-top: 28rpx;
}

.privacy-footnote {
  display: block;
  margin-top: 18rpx;
  color: var(--tree-text-weak);
  font-size: 22rpx;
  line-height: 1.7;
}
</style>
