<template>
  <view class="page">
    <view class="card intro-card">
      <text class="title">亲属关系工具</text>
      <text class="muted intro-text">推导关系结果仅供参考，最多支持 5 代关系推测。</text>
    </view>

    <view class="card">
      <text class="section-title">本人信息</text>
      <text class="muted field-hint">出生日期优先用于判断长幼，年龄仅作为补充。</text>
      <view class="picker-row">
        <text class="label">本人性别</text>
        <view class="tag-group">
          <text
            v-for="item in selfGenderOptions"
            :key="item.value"
            class="tag selectable"
            :class="{ active: self.gender === item.value }"
            @click="self.gender = item.value"
          >
            {{ item.label }}
          </text>
        </view>
      </view>
      <text v-if="self.gender === 'unknown'" class="disabled-hint">请先选择本人性别，再开始选择关系。</text>
      <input v-model.trim="self.birthday" class="input" placeholder="本人出生日期（可选，如 1990-01-01）" />
      <input v-model.trim="selfAgeInput" class="input" type="number" placeholder="本人年龄（可选）" />
    </view>

    <view class="card path-card">
      <text class="section-title">关系路径</text>
      <view class="path-pills">
        <text class="path-pill path-pill-self">我</text>
        <template v-for="(step, index) in steps" :key="index">
          <text class="path-separator">›</text>
          <text class="path-pill">{{ formatStepLabel(step) }}</text>
        </template>
      </view>
      <text class="path-meta">当前代数：{{ generationDepth }} / {{ maxDepth }}</text>
    </view>

    <view class="card">
      <text class="section-title">选择下一层关系</text>
      <view class="relation-grid">
        <view
          v-for="option in relationOptions"
          :key="option.relation"
          class="relation-btn"
          :class="{
            active: pendingRelation === option.relation,
            disabled: !option.enabled
          }"
          @click="selectRelation(option)"
        >
          <text class="relation-btn-label">{{ option.label }}</text>
        </view>
      </view>
      <text v-if="lastDisabledReason" class="disabled-hint">{{ lastDisabledReason }}</text>
    </view>

    <view class="card">
      <text class="section-title">补充这个人的信息</text>
      <view v-if="!pendingRelation" class="empty-hint">
        <text class="muted">请先选择下一层关系。</text>
      </view>
      <view v-else class="pending-form">
        <text class="pending-relation-hint">正在为「{{ selectedRelationLabel }}」补充信息</text>
        <view class="picker-row">
          <text class="label">性别</text>
          <view class="tag-group">
            <text
              v-for="item in genderOptions"
              :key="item.value"
              class="tag selectable"
              :class="{ active: pendingPerson.gender === item.value }"
              @click="pendingPerson.gender = item.value"
            >
              {{ item.label }}
            </text>
          </view>
        </view>
        <input v-model.trim="pendingPerson.birthday" class="input" placeholder="出生日期（可选）" />
        <input v-model.trim="pendingAgeInput" class="input" type="number" placeholder="年龄（可选）" />
        <view v-if="pendingRelation === 'sibling'" class="picker-row">
          <text class="label">长幼（相对上一位）</text>
          <view class="tag-group">
            <text
              v-for="item in relativeAgeOptions"
              :key="item.value"
              class="tag selectable"
              :class="{ active: pendingRelativeAge === item.value }"
              @click="pendingRelativeAge = item.value"
            >
              {{ item.label }}
            </text>
          </view>
        </view>
        <button class="button append-btn" @click="appendStep">添加到路径</button>
        <text v-if="appendError" class="error">{{ appendError }}</text>
      </view>
    </view>

    <view class="card result-card" :class="`result-${resolution.status}`">
      <template v-if="resolution.status === 'resolved'">
        <text class="result-label">常见称谓</text>
        <text class="result-title">{{ resolution.primaryTitle }}</text>
      </template>
      <template v-else-if="resolution.status === 'ambiguous'">
        <text class="result-label">可能称谓</text>
        <view v-if="resolution.candidates?.length" class="candidate-list">
          <text v-for="(item, index) in resolution.candidates" :key="index" class="candidate-tag">
            {{ item }}
          </text>
        </view>
        <text v-else-if="resolution.primaryTitle" class="result-title muted-title">
          {{ resolution.primaryTitle }}
        </text>
        <text class="muted result-hint">
          请补充性别、长幼或父系/母系方向等信息，以获得更明确的称谓。
        </text>
      </template>
      <template v-else>
        <text class="result-label">暂未收录</text>
        <text class="result-unsupported">暂未收录该关系的常用称谓</text>
        <text class="muted result-hint">
          你仍然可以保留完整关系路径，后续版本会继续补充称谓规则。
        </text>
      </template>

      <view v-if="resolution.aliases.length" class="aliases-block">
        <text class="aliases-label">其他常见叫法：</text>
        <text class="aliases-text">{{ resolution.aliases.join('、') }}</text>
      </view>

      <text v-if="resolution.explanation" class="explanation-text">{{ resolution.explanation }}</text>
    </view>

    <view class="card action-card">
      <view class="action-row">
        <button
          class="button secondary action-btn"
          :class="{ 'btn-disabled': steps.length === 0 }"
          @click="undoStep"
        >
          撤销一步
        </button>
        <button class="button secondary action-btn" @click="confirmReset">重新开始</button>
        <button class="button action-btn" @click="copyDescription">复制关系描述</button>
      </view>
    </view>

    <view class="footer-check">
      <text class="self-check-text">规则自检：{{ selfCheckSummary }}</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'

import { canAppendRelation, getRelationOptions } from '@/features/kinship/allowedRelations'
import {
  formatStepLabel,
  normalizePersonFacts
} from '@/features/kinship/helpers'
import { resolveKinship } from '@/features/kinship/resolveKinship'
import { runKinshipSelfChecks } from '@/features/kinship/testCases'
import type { Gender, KinshipContext, KinshipRelation, KinshipStep, PersonFacts, RelativeAge } from '@/features/kinship/types'
import { MAX_KINSHIP_DEPTH } from '@/features/kinship/types'

const maxDepth = MAX_KINSHIP_DEPTH
const selfGenderOptions = [
  { value: 'male' as Gender, label: '男' },
  { value: 'female' as Gender, label: '女' }
]
const genderOptions = [
  { value: 'male' as Gender, label: '男' },
  { value: 'female' as Gender, label: '女' },
  { value: 'unknown' as Gender, label: '未知' }
]
const relativeAgeOptions = [
  { value: 'older' as RelativeAge, label: '年长' },
  { value: 'younger' as RelativeAge, label: '年幼' },
  { value: 'same' as RelativeAge, label: '同龄' },
  { value: 'unknown' as RelativeAge, label: '未知' }
]

const self = ref<PersonFacts>({ gender: 'unknown' })
const selfAgeInput = ref('')
const steps = ref<KinshipStep[]>([])
const pendingRelation = ref<KinshipRelation | null>(null)
const pendingPerson = ref<PersonFacts>({ gender: 'unknown' })
const pendingAgeInput = ref('')
const pendingRelativeAge = ref<RelativeAge>('unknown')
const appendError = ref('')
const lastDisabledReason = ref('')

const context = computed<KinshipContext>(() => ({
  self: normalizePersonFacts({
    ...self.value,
    age: selfAgeInput.value ? Number(selfAgeInput.value) : undefined
  }),
  steps: steps.value,
  maxDepth
}))

const relationOptions = computed(() => {
  const options = getRelationOptions(context.value)
  if (self.value.gender !== 'unknown') return options
  return options.map((option) => ({
    ...option,
    enabled: false,
    disabledReason: '请先选择本人性别。'
  }))
})
const resolution = computed(() => resolveKinship(context.value))
const selfCheckResult = computed(() => runKinshipSelfChecks())
const selfCheckSummary = computed(() => {
  const result = selfCheckResult.value
  return `${result.passed} / ${result.total} 通过`
})

const selectedRelationLabel = computed(() => {
  if (!pendingRelation.value) return ''
  return relationOptions.value.find((item) => item.relation === pendingRelation.value)?.label || ''
})

const generationDepth = computed(() => steps.value.filter((step) => step.relation !== 'spouse').length)

function selectRelation(option: { relation: KinshipRelation; enabled: boolean; disabledReason?: string }) {
  if (self.value.gender === 'unknown') {
    lastDisabledReason.value = '请先选择本人性别。'
    uni.showToast({
      title: '请先选择本人性别',
      icon: 'none'
    })
    return
  }
  if (!option.enabled) {
    lastDisabledReason.value = option.disabledReason || '当前不可选择该关系'
    uni.showToast({
      title: option.disabledReason || '当前不可选择该关系',
      icon: 'none',
      duration: 2800
    })
    return
  }
  lastDisabledReason.value = ''
  pendingRelation.value = option.relation
  pendingPerson.value = { gender: getDefaultGenderForRelation(option.relation) }
  pendingAgeInput.value = ''
  pendingRelativeAge.value = 'unknown'
  appendError.value = ''
}

function getCurrentPersonGender(): Gender {
  if (steps.value.length === 0) return self.value.gender
  return steps.value[steps.value.length - 1].person.gender
}

function getOppositeGender(gender: Gender): Gender {
  if (gender === 'male') return 'female'
  if (gender === 'female') return 'male'
  return 'unknown'
}

function getDefaultGenderForRelation(relation: KinshipRelation): Gender {
  if (relation !== 'spouse') return 'unknown'
  return getOppositeGender(getCurrentPersonGender())
}

function buildPendingStep(): KinshipStep {
  const person = normalizePersonFacts({
    ...pendingPerson.value,
    age: pendingAgeInput.value ? Number(pendingAgeInput.value) : undefined
  })
  const step: KinshipStep = {
    relation: pendingRelation.value as KinshipRelation,
    person
  }
  if (pendingRelation.value === 'sibling') {
    step.relativeAge = pendingRelativeAge.value
  }
  return step
}

function appendStep() {
  appendError.value = ''
  if (self.value.gender === 'unknown') {
    appendError.value = '请先选择本人性别。'
    return
  }
  if (!pendingRelation.value) {
    appendError.value = '请先选择下一层关系。'
    return
  }
  const validation = canAppendRelation(context.value, pendingRelation.value)
  if (!validation.valid) {
    appendError.value = validation.reason || '当前不可添加该关系。'
    return
  }
  steps.value = [...steps.value, buildPendingStep()]
  pendingRelation.value = null
  pendingPerson.value = { gender: 'unknown' }
  pendingAgeInput.value = ''
  pendingRelativeAge.value = 'unknown'
  lastDisabledReason.value = ''
}

function undoStep() {
  if (steps.value.length === 0) {
    uni.showToast({ title: '当前没有可撤销的关系', icon: 'none' })
    return
  }
  steps.value = steps.value.slice(0, -1)
  appendError.value = ''
  lastDisabledReason.value = ''
}

function resetAll() {
  self.value = { gender: 'unknown' }
  selfAgeInput.value = ''
  steps.value = []
  pendingRelation.value = null
  pendingPerson.value = { gender: 'unknown' }
  pendingAgeInput.value = ''
  pendingRelativeAge.value = 'unknown'
  appendError.value = ''
  lastDisabledReason.value = ''
}

function confirmReset() {
  if (steps.value.length === 0 && self.value.gender === 'unknown' && !selfAgeInput.value && !self.value.birthday) {
    resetAll()
    return
  }
  uni.showModal({
    title: '重新开始',
    content: '是否清空当前关系路径？',
    confirmText: '清空',
    cancelText: '取消',
    success: (res) => {
      if (res.confirm) {
        resetAll()
      }
    }
  })
}

function buildPathText(): string {
  if (steps.value.length === 0) return '我'
  return ['我', ...steps.value.map((step) => formatStepLabel(step))].join(' > ')
}

function buildCopyText(): string {
  const path = buildPathText()
  if (steps.value.length === 0) return '我'
  if (resolution.value.status === 'resolved' && resolution.value.primaryTitle) {
    return `${path}：${resolution.value.primaryTitle}`
  }
  if (resolution.value.status === 'ambiguous') {
    const candidates = resolution.value.candidates?.join('、') || resolution.value.primaryTitle || '多种称谓'
    return `${path}：可能为 ${candidates}`
  }
  return `${path}：暂未收录该关系的常用称谓`
}

function copyDescription() {
  uni.setClipboardData({
    data: buildCopyText(),
    success: () => {
      uni.showToast({ title: '已复制关系描述', icon: 'none' })
    }
  })
}
</script>

<style scoped>
.intro-card {
  padding-bottom: 24rpx;
}

.intro-text {
  display: block;
}

.tip-text {
  display: block;
  margin-top: 16rpx;
  padding: 16rpx 20rpx;
  border-radius: 12rpx;
  background: #f4f1e8;
  color: #6b7280;
  font-size: 24rpx;
  line-height: 1.6;
}

.beta-tip {
  background: #eef4fb;
  color: #4b6a8a;
}

.field-hint {
  display: block;
  margin-bottom: 12rpx;
}

.picker-row {
  margin-top: 20rpx;
}

.label {
  display: block;
  margin-bottom: 10rpx;
  color: #6b7280;
  font-size: 24rpx;
}

.tag-group {
  display: flex;
  flex-wrap: wrap;
  gap: 10rpx;
}

.tag.selectable {
  cursor: pointer;
}

.tag.active {
  background: #1f3a5f;
  color: #fff;
}

.path-card {
  background: linear-gradient(180deg, #fff 0%, #faf8f4 100%);
}

.path-pills {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8rpx;
  margin-bottom: 16rpx;
}

.path-pill {
  display: inline-flex;
  align-items: center;
  border-radius: 999rpx;
  background: #eef4f1;
  color: #2f6b57;
  padding: 10rpx 20rpx;
  font-size: 26rpx;
  line-height: 1.4;
  max-width: 100%;
  word-break: break-all;
}

.path-pill-self {
  background: #1f3a5f;
  color: #fff;
  font-weight: 600;
}

.path-separator {
  color: #9ca3af;
  font-size: 24rpx;
  flex-shrink: 0;
}

.path-meta {
  color: #6b7280;
  font-size: 24rpx;
}

.relation-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 12rpx;
}

.relation-btn {
  flex: 1 1 calc(50% - 6rpx);
  min-width: 140rpx;
  box-sizing: border-box;
  border: 2rpx solid #1f3a5f;
  border-radius: 16rpx;
  background: #1f3a5f;
  padding: 24rpx 16rpx;
  text-align: center;
}

.relation-btn.active {
  border-color: #2f6b57;
  background: #2f6b57;
  box-shadow: 0 4rpx 12rpx rgba(47, 107, 87, 0.25);
}

.relation-btn.disabled {
  border-color: #e5e0d6;
  background: #f9f7f3;
  opacity: 1;
}

.relation-btn-label {
  color: #fff;
  font-size: 28rpx;
  font-weight: 600;
}

.relation-btn.disabled .relation-btn-label {
  color: #9ca3af;
}

.disabled-hint {
  display: block;
  margin-top: 16rpx;
  color: #9ca3af;
  font-size: 24rpx;
  line-height: 1.6;
}

.empty-hint {
  padding: 24rpx 0 8rpx;
}

.pending-form {
  padding-top: 8rpx;
}

.pending-relation-hint {
  display: block;
  margin-bottom: 16rpx;
  color: #2f6b57;
  font-size: 26rpx;
}

.append-btn {
  margin-top: 24rpx;
}

.error {
  display: block;
  margin-top: 12rpx;
  color: #c0392b;
  font-size: 24rpx;
}

.result-card {
  border-color: #d4e5de;
  background: linear-gradient(180deg, #f8fbf9 0%, #fff 100%);
}

.result-label {
  display: block;
  margin-bottom: 12rpx;
  color: #6b7280;
  font-size: 24rpx;
  font-weight: 600;
  letter-spacing: 1rpx;
}

.result-title {
  display: block;
  color: #2f6b57;
  font-size: 40rpx;
  font-weight: 700;
  line-height: 1.4;
}

.muted-title {
  font-size: 32rpx;
}

.result-unsupported {
  display: block;
  color: #6b7280;
  font-size: 30rpx;
  font-weight: 600;
  line-height: 1.5;
}

.result-hint {
  display: block;
  margin-top: 12rpx;
}

.candidate-list {
  display: flex;
  flex-wrap: wrap;
  gap: 10rpx;
  margin-bottom: 8rpx;
}

.candidate-tag {
  display: inline-flex;
  border-radius: 999rpx;
  background: #fff;
  border: 1rpx solid #b8d4c8;
  color: #2f6b57;
  padding: 10rpx 20rpx;
  font-size: 28rpx;
  font-weight: 600;
}

.aliases-block {
  margin-top: 20rpx;
  padding-top: 20rpx;
  border-top: 1rpx solid #e5e0d6;
}

.aliases-label {
  display: block;
  margin-bottom: 6rpx;
  color: #6b7280;
  font-size: 24rpx;
}

.aliases-text {
  color: #374151;
  font-size: 26rpx;
  line-height: 1.6;
}

.explanation-text {
  display: block;
  margin-top: 16rpx;
  color: #6b7280;
  font-size: 26rpx;
  line-height: 1.7;
}

.action-card {
  padding-bottom: 20rpx;
}

.action-row {
  display: flex;
  flex-wrap: wrap;
  gap: 12rpx;
}

.action-btn {
  flex: 1 1 calc(50% - 6rpx);
  min-width: 200rpx;
  margin: 0;
}

.action-btn:last-child {
  flex: 1 1 100%;
}

.btn-disabled {
  opacity: 0.5;
}

.footer-check {
  padding: 8rpx 0 32rpx;
  text-align: center;
}

.self-check-text {
  color: #c4c9d0;
  font-size: 20rpx;
}
</style>
