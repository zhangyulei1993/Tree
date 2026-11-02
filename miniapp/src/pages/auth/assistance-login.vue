<template>
  <view class="archive-page assistance-page">
    <MiniBackHome />
    <view class="archive-page-head"><view><text class="archive-kicker">FAMILY ASSISTANCE</text><text class="archive-title">协助登录</text><text class="archive-subtitle">受亲友委托，短期维护家庭树资料</text></view><view class="archive-seal">助</view></view>
    <view class="archive-form-panel"><view class="archive-section-head"><text class="archive-section-title">输入协助信息</text><text class="archive-section-subtitle">需由家庭成员提供；协助权限会在到期或撤销后立即失效。</text></view><text v-if="errorMessage" class="tree-field-error">{{ errorMessage }}</text><view class="field"><text>对方绑定的手机号</text><input v-model="phone" maxlength="11" type="number" placeholder="请输入 11 位手机号"/></view><view class="field"><text>8 位协助码</text><input v-model="code" maxlength="8" type="number" placeholder="请输入协助码"/></view><MiniButton :disabled="submitting || phone.length !== 11 || code.length !== 8" :loading="submitting" @click="submit">进入协助维护</MiniButton></view>
    <MiniNotice tone="info">协助模式仅可查看和维护家庭树、成员资料及关系；账号、邀请、订单、权限与家庭设置不可操作。</MiniNotice>
    <view class="back-login" @click="backToWechat">返回微信登录</view>
  </view>
</template>
<script setup lang="ts">
import { ref } from 'vue'
import { apiErrorMessage } from '@/api/client'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
import { useSessionStore } from '@/stores/session'
const session=useSessionStore(); const phone=ref(''); const code=ref(''); const submitting=ref(false); const errorMessage=ref('')
async function submit(){submitting.value=true;errorMessage.value='';try{await session.loginWithAssistanceCode(phone.value,code.value);uni.showToast({title:'已进入协助维护',icon:'success'});setTimeout(()=>session.routeAfterAuth(),300)}catch(error){errorMessage.value=apiErrorMessage(error,'手机号或协助码不正确，或协助已失效。')}finally{submitting.value=false}}
function backToWechat(){uni.navigateBack()}
</script>
<style scoped>
.assistance-page{padding-top:28rpx}.field{display:flex;flex-direction:column;gap:12rpx;margin:22rpx 0;color:var(--archive-ink);font-size:25rpx}.field input{padding:20rpx;border:1rpx solid var(--archive-line-strong);border-radius:10rpx;background:#fffdf7;font-size:28rpx}.back-login{padding:32rpx;color:var(--archive-ink-soft);font-size:25rpx;text-align:center}
</style>
