<template>
  <view class="tree-page auth-page">
    <MiniBackHome />
    <view class="tree-auth-brand"><text class="tree-auth-brand-title">更改手机号</text><text class="tree-auth-brand-desc">分别验证当前手机号和新手机号</text></view>
    <MiniCard variant="soft" class="tree-auth-card">
      <text class="tree-field-label">当前手机号</text>
      <view class="input readonly">{{ session.user?.phone || '未绑定' }}</view>
      <view class="code-row"><input v-model.trim="oldPhoneCode" class="input" maxlength="6" type="number" placeholder="当前手机号验证码" /><MiniButton variant="secondary" :disabled="sendingOld" @click="sendOld">{{ sendingOld ? '发送中' : '获取验证码' }}</MiniButton></view>
      <text class="tree-field-label">新手机号</text>
      <input v-model.trim="newPhone" class="input" maxlength="11" type="number" placeholder="请输入新手机号" />
      <view class="code-row"><input v-model.trim="newPhoneCode" class="input" maxlength="6" type="number" placeholder="新手机号验证码" /><MiniButton variant="secondary" :disabled="sendingNew" @click="sendNew">{{ sendingNew ? '发送中' : '获取验证码' }}</MiniButton></view>
      <text v-if="errorMessage" class="tree-field-error">{{ errorMessage }}</text>
      <MiniButton :loading="submitting" :disabled="submitting || !oldPhoneCode || !newPhoneCode || !newPhone" @click="submit">确认更改</MiniButton>
    </MiniCard>
  </view>
</template>

<script setup lang="ts">
import { onShow } from '@dcloudio/uni-app'
import { ref } from 'vue'
import { apiErrorMessage } from '@/api/client'
import { changePhone, sendCode } from '@/api/auth'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniCard from '@/components/base/MiniCard.vue'
import { useSessionStore } from '@/stores/session'

const session = useSessionStore()
const oldPhoneCode = ref('')
const newPhone = ref('')
const newPhoneCode = ref('')
const sendingOld = ref(false)
const sendingNew = ref(false)
const submitting = ref(false)
const errorMessage = ref('')

async function sendOld() {
  const phone = session.user?.phone
  if (!phone) { errorMessage.value = '当前账号未绑定手机号。'; return }
  sendingOld.value = true; errorMessage.value = ''
  try { await sendCode(phone, 'CHANGE_PHONE_OLD'); uni.showToast({ title: '验证码已发送', icon: 'none' }) }
  catch (error) { errorMessage.value = apiErrorMessage(error, '验证码发送失败。') }
  finally { sendingOld.value = false }
}

async function sendNew() {
  if (!/^1\d{10}$/.test(newPhone.value)) { errorMessage.value = '请输入正确的新手机号。'; return }
  sendingNew.value = true; errorMessage.value = ''
  try { await sendCode(newPhone.value, 'CHANGE_PHONE_NEW'); uni.showToast({ title: '验证码已发送', icon: 'none' }) }
  catch (error) { errorMessage.value = apiErrorMessage(error, '验证码发送失败。') }
  finally { sendingNew.value = false }
}

async function submit() {
  submitting.value = true
  errorMessage.value = ''
  try {
    await changePhone({
      oldPhoneCode: oldPhoneCode.value,
      newPhone: newPhone.value,
      newPhoneCode: newPhoneCode.value
    })
    await session.refreshMe()
    uni.showToast({ title: '手机号已更改', icon: 'success' })
    setTimeout(() => uni.navigateBack(), 500)
  } catch (error) {
    errorMessage.value = apiErrorMessage(error, '手机号更改失败。')
  } finally {
    submitting.value = false
  }
}

onShow(() => {
  session.restoreSession()
  if (!session.isLoggedIn) {
    uni.reLaunch({ url: '/pages/auth/wechat-login' })
    return
  }
  if (!session.isPhoneBound) {
    uni.redirectTo({ url: '/pages/auth/bind-phone' })
  }
})
</script>

<style scoped>
.code-row { display: grid; grid-template-columns: minmax(0, 1fr) 220rpx; gap: 12rpx; align-items: center; margin-bottom: 20rpx; }
.readonly { color: var(--tree-muted); }
</style>
