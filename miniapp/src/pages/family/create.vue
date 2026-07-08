<template>
  <view class="archive-page create-page">
    <MiniBackHome />

    <view v-if="!authChecked" class="create-state archive-form-panel">
      <MiniEmptyState symbol="…" title="正在确认资料" description="请稍候..." />
    </view>

    <template v-else>
      <view class="create-head archive-page-head">
        <view>
          <text class="archive-kicker">New Genealogy</text>
          <text class="archive-title">创建家庭</text>
          <text class="archive-subtitle">填写姓氏与基本信息，开启新家谱册</text>
        </view>
        <view class="archive-seal create-seal" :class="{ muted: !form.surname.trim() }">{{ coverSurnameLetter }}</view>
      </view>

      <view class="create-cover">
        <view class="create-cover-copy">
          <text class="create-surname">{{ coverSurnameLabel }}</text>
          <text class="create-name">{{ coverFamilyName }}</text>
          <text class="create-region">{{ coverRegionSummary }}</text>
          <view class="create-status-line">
            <text class="create-status-badge">新建册页</text>
            <text v-if="form.nativePlace" class="archive-chip">{{ form.nativePlace }}</text>
            <text v-if="form.regionText" class="archive-chip">{{ form.regionText }}</text>
          </view>
        </view>
        <view class="create-spine archive-book-spine">
          <text>新</text>
          <text>谱</text>
          <text>册</text>
        </view>
      </view>

      <view class="archive-form-panel">
        <view class="archive-section-head">
          <text class="archive-section-title">家谱册信息</text>
          <text class="archive-section-subtitle">创建后将自动生成创建者成员节点</text>
        </view>

        <text class="tree-field-label">家庭姓氏</text>
        <input v-model.trim="form.surname" class="tree-input" maxlength="20" placeholder="必填，如：张" />

        <text class="tree-field-label">你的性别</text>
        <picker mode="selector" :range="genderLabels" :value="genderIndex" @change="onGenderChange">
          <view class="field-picker">{{ genderLabels[genderIndex] }}</view>
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
        <MiniButton class="create-action" :loading="submitting" :disabled="submitting" @click="submit">
          创建家庭
        </MiniButton>
      </view>
    </template>
  </view>
</template>

<script setup lang="ts">
import { onShow } from '@dcloudio/uni-app'
import { computed, reactive, ref } from 'vue'

import { apiErrorMessage } from '@/api/client'
import { createFamily } from '@/api/families'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
import { useSessionStore } from '@/stores/session'
import type { Gender } from '@/types/api'
import { normalizeText, optionalText, validateTextFields } from '@/utils/inputValidation'

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

const coverSurnameLetter = computed(() => form.surname.trim().slice(0, 1) || '姓')

const coverSurnameLabel = computed(() => {
  const surname = form.surname.trim()
  return surname ? `${surname}氏` : '某氏'
})

const coverFamilyName = computed(() => {
  const customName = form.familyName.trim()
  if (customName) return customName
  const surname = form.surname.trim()
  return surname ? `${surname}氏家族` : '新家谱册'
})

const coverRegionSummary = computed(() => {
  const parts = [form.nativePlace.trim(), form.regionText.trim()].filter(Boolean)
  return parts.length > 0 ? parts.join(' · ') : '籍贯与地区（可选）'
})

function onGenderChange(event: { detail: { value: number | string } }) {
  genderIndex.value = Number(event.detail.value) || 0
  form.founderGender = genders[genderIndex.value] || 'MALE'
}

async function submit() {
  const validationMessage = validateTextFields([
    { value: form.surname, label: '家庭姓氏', kind: 'surname', required: true, maxLength: 20 },
    { value: form.familyName, label: '家庭名称', kind: 'name', maxLength: 100 },
    { value: form.nativePlace, label: '籍贯', maxLength: 100 },
    { value: form.regionText, label: '地区', maxLength: 100 },
    { value: form.description, label: '家庭简介', kind: 'multiLine', maxLength: 500 }
  ])
  if (validationMessage) {
    errorMessage.value = validationMessage
    return
  }
  submitting.value = true
  errorMessage.value = ''
  try {
    const family = await createFamily({
      surname: normalizeText(form.surname),
      founderGender: form.founderGender,
      familyName: optionalText(form.familyName),
      nativePlace: optionalText(form.nativePlace),
      regionText: optionalText(form.regionText),
      description: optionalText(form.description)
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
.create-page {
  padding-top: 28rpx;
}

.create-page :deep(.mini-back-home) {
  margin-bottom: 22rpx;
}

.create-cover {
  display: flex;
  min-height: 210rpx;
  margin-bottom: 24rpx;
  border-top: 1rpx solid var(--archive-line-strong);
  border-bottom: 1rpx solid var(--archive-line);
}

.create-cover-copy {
  flex: 1;
  min-width: 0;
  padding: 28rpx 28rpx 28rpx 0;
}

.create-surname,
.create-name,
.create-region {
  display: block;
}

.create-surname {
  color: var(--archive-cinnabar);
  font-size: 24rpx;
  font-weight: 700;
  letter-spacing: 2rpx;
}

.create-name {
  margin-top: 10rpx;
  color: var(--archive-blue);
  font-family: 'Songti SC', 'STSong', serif;
  font-size: 42rpx;
  font-weight: 800;
  letter-spacing: 2rpx;
  line-height: 1.28;
}

.create-region {
  margin-top: 12rpx;
  color: var(--archive-ink-soft);
  font-size: 23rpx;
  line-height: 1.55;
}

.create-status-line {
  display: flex;
  flex-wrap: wrap;
  gap: 10rpx;
  margin-top: 16rpx;
}

.create-status-badge {
  display: inline-flex;
  align-items: center;
  min-height: 42rpx;
  border: 1rpx dashed var(--archive-line-strong);
  background: rgba(255, 248, 234, 0.58);
  color: var(--archive-ink-soft);
  padding: 0 14rpx;
  font-size: 21rpx;
  line-height: 1.2;
}

.create-seal.muted {
  opacity: 0.72;
}

.create-spine {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 10rpx;
  width: 118rpx;
  flex-shrink: 0;
}

.create-spine text {
  color: rgba(255, 255, 255, 0.94);
  font-family: 'Songti SC', 'STSong', serif;
  font-size: 28rpx;
  font-weight: 800;
}

.create-page :deep(.mini-notice) {
  margin-top: 16rpx;
  margin-bottom: 16rpx;
}

.create-action {
  margin-top: 20rpx;
  align-self: flex-start;
}
</style>
