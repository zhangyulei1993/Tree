<template>
  <view class="tree-page">
    <MiniBackHome />

    <MiniCard v-if="authChecked" variant="soft">
      <text class="tree-field-label">家庭姓氏</text>
      <input v-model.trim="form.surname" class="tree-input" maxlength="20" placeholder="必填，如：张" />

      <text class="tree-field-label">你的性别</text>
      <picker mode="selector" :range="genderLabels" :value="genderIndex" @change="onGenderChange">
        <view class="picker-field">{{ genderLabels[genderIndex] }}</view>
      </picker>
      <MiniNotice tone="info">性别用于正确初始化家谱中的本人节点，创建后仍可在成员资料中修改。</MiniNotice>

      <text class="tree-field-label">家庭名称</text>
      <input v-model.trim="form.familyName" class="tree-input" maxlength="100" placeholder="可选，默认生成“某氏家族”" />

      <text class="tree-field-label">籍贯</text>
      <input v-model.trim="form.nativePlace" class="tree-input" maxlength="100" placeholder="可选" />

      <text class="tree-field-label">地区</text>
      <input v-model.trim="form.regionText" class="tree-input" maxlength="100" placeholder="可选" />

      <text class="tree-field-label">家庭简介</text>
      <textarea v-model.trim="form.description" class="tree-textarea" maxlength="500" placeholder="可选" />

      <text v-if="errorMessage" class="tree-field-error">{{ errorMessage }}</text>
      <MiniButton :loading="submitting" :disabled="submitting" @click="submit">创建家庭</MiniButton>
    </MiniCard>
  </view>
</template>

<script setup lang="ts">
import { onShow } from '@dcloudio/uni-app'
import { reactive, ref } from 'vue'

import { apiErrorMessage } from '@/api/client'
import { createFamily } from '@/api/families'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniCard from '@/components/base/MiniCard.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
import { useSessionStore } from '@/stores/session'
import type { Gender } from '@/types/api'

const session = useSessionStore()
const authChecked = ref(false)
const submitting = ref(false)
const errorMessage = ref('')
const genderIndex = ref(0)
const genders: Gender[] = ['MALE', 'FEMALE']
const genderLabels = ['男', '女']
const form = reactive({
  surname: '',
  founderGender: 'MALE' as Gender,
  familyName: '',
  nativePlace: '',
  regionText: '',
  description: ''
})

function onGenderChange(event: { detail: { value: number | string } }) {
  genderIndex.value = Number(event.detail.value) || 0
  form.founderGender = genders[genderIndex.value] || 'MALE'
}

async function submit() {
  if (!form.surname) {
    errorMessage.value = '请填写家庭姓氏。'
    return
  }
  submitting.value = true
  errorMessage.value = ''
  try {
    const family = await createFamily({
      surname: form.surname,
      founderGender: form.founderGender,
      familyName: form.familyName || undefined,
      nativePlace: form.nativePlace || undefined,
      regionText: form.regionText || undefined,
      description: form.description || undefined
    })
    uni.redirectTo({ url: `/pages/family/detail?familyId=${encodeURIComponent(String(family.id))}` })
  } catch (error) {
    errorMessage.value = apiErrorMessage(error, '创建家庭失败。')
  } finally {
    submitting.value = false
  }
}

onShow(() => {
  authChecked.value = session.requireProfileComplete('/pages/family/create')
})
</script>

<style scoped>
.picker-field {
  box-sizing: border-box;
  width: 100%;
  min-height: 84rpx;
  margin-bottom: 18rpx;
  padding: 22rpx 24rpx;
  border: 1rpx solid var(--tree-border-warm, #ebe4d6);
  border-radius: 16rpx;
  background: #fff;
  color: var(--tree-text);
  font-size: 26rpx;
}
</style>
