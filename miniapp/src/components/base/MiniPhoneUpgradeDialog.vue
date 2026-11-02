<template>
  <view v-if="visible" class="phone-upgrade-mask" @click="close">
    <view class="phone-upgrade-dialog" @click.stop>
      <text class="phone-upgrade-kicker">PHONE VERIFICATION</text>
      <text class="phone-upgrade-title">{{ props.title }}</text>
      <text class="phone-upgrade-desc">{{ props.description }}</text>
      <text v-if="errorMessage" class="tree-field-error">{{ errorMessage }}</text>
      <!-- #ifdef MP-WEIXIN -->
      <button
        :class="['phone-upgrade-action', { 'is-binding': binding }]"
        open-type="getPhoneNumber"
        :disabled="binding"
        @getphonenumber="bindWithWechatPhone"
      >
        {{ binding ? '绑定中...' : props.actionText }}
      </button>
      <!-- #endif -->
      <!-- #ifndef MP-WEIXIN -->
      <MiniNotice tone="info">请在微信小程序中完成手机号授权。</MiniNotice>
      <!-- #endif -->
      <view class="phone-upgrade-cancel" @click="close">暂不绑定</view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'

import { apiErrorMessage } from '@/api/client'
import MiniNotice from '@/components/base/MiniNotice.vue'
import { useSessionStore } from '@/stores/session'

const props = withDefaults(defineProps<{
  visible: boolean
  title?: string
  description?: string
  actionText?: string
}>(), {
  title: '绑定手机号，继续此操作',
  description: '为确保家庭成员与邀请操作可追溯，请授权获取微信手机号。',
  actionText: '授权并继续'
})

const emit = defineEmits<{
  'update:visible': [value: boolean]
  bound: []
}>()

const session = useSessionStore()
const binding = ref(false)
const errorMessage = ref('')

function close() {
  if (binding.value) return
  errorMessage.value = ''
  emit('update:visible', false)
}

type PhoneNumberEvent = { detail?: { code?: string; errMsg?: string } }

async function bindWithWechatPhone(event: PhoneNumberEvent) {
  const phoneCode = event.detail?.code || ''
  if (!phoneCode) {
    errorMessage.value = /deny|cancel|拒绝|取消/i.test(event.detail?.errMsg || '')
      ? '你已取消手机号授权。'
      : '手机号授权失败，请稍后重试。'
    return
  }
  binding.value = true
  errorMessage.value = ''
  try {
    await session.bindWechatPhone({ phoneCode })
    await session.refreshMe().catch(() => undefined)
    uni.showToast({ title: '手机号已绑定', icon: 'success' })
    emit('update:visible', false)
    emit('bound')
  } catch (error) {
    errorMessage.value = apiErrorMessage(error, '手机号绑定失败，请稍后重试。')
  } finally {
    binding.value = false
  }
}
</script>

<style scoped>
.phone-upgrade-mask { position:fixed; inset:0; z-index:1000; display:flex; align-items:flex-end; background:rgba(20,31,46,.42); }
.phone-upgrade-dialog { width:100%; box-sizing:border-box; border-radius:22rpx 22rpx 0 0; background:var(--archive-paper-light); padding:38rpx 34rpx calc(34rpx + env(safe-area-inset-bottom)); }
.phone-upgrade-kicker, .phone-upgrade-title, .phone-upgrade-desc { display:block; }
.phone-upgrade-kicker { color:var(--archive-cinnabar); font-size:19rpx; font-weight:800; letter-spacing:3rpx; }
.phone-upgrade-title { margin-top:10rpx; color:var(--archive-blue); font-family:'Songti SC','STSong',serif; font-size:38rpx; font-weight:800; line-height:1.25; }
.phone-upgrade-desc { margin-top:14rpx; color:var(--archive-ink-soft); font-size:25rpx; line-height:1.6; }
.phone-upgrade-action { display:flex; width:100%; min-height:90rpx; align-items:center; justify-content:center; margin-top:26rpx; border:0; border-radius:10rpx; background:var(--archive-blue); color:#fffdf6; font-size:28rpx; font-weight:700; }
.phone-upgrade-action::after { border:0; }
.phone-upgrade-action.is-binding { opacity:.6; }
.phone-upgrade-cancel { min-height:74rpx; display:flex; align-items:center; justify-content:center; margin-top:10rpx; color:var(--archive-ink-soft); font-size:25rpx; }
</style>
