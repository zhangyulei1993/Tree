<template>
  <view class="page">
    <view class="card">
      <text class="title">亲属关系工具</text>
      <text class="muted">选择一层层亲属关系，查看常见称谓。称谓因地区和习惯可能不同，结果仅供参考。</text>
    </view>

    <view class="card">
      <text class="section-title">本人信息</text>
      <text class="muted field-hint">出生日期优先用于判断，年龄仅作为补充。</text>
      <view class="picker-row">
        <text class="label">本人性别</text>
        <view class="tag-group">
          <text
            v-for="item in genderOptions"
            :key="item.value"
            class="tag selectable"
            :class="{ active: self.gender === item.value }"
            @click="self.gender = item.value"
          >
            {{ item.label }}
          </text>
        </view>
      </view>
      <input v-model.trim="self.birthday" class="input" placeholder="本人出生日期（可选）" />
      <input v-model.trim="selfAgeInput" class="input" type="number" placeholder="本人年龄（可选）" />
    </view>

    <view class="card">
      <text class="section-title">当前路径</text>
      <text class="path-display">{{ pathDisplay }}</text>
      <text class="muted">当前层数：{{ steps.length }} / {{ maxDepth }}</text>
      <view v-if="steps.length > 0" class="step-list">
        <view v-for="(step, index) in steps" :key="index" class="step-item">
          <text>{{ index + 1 }}. {{ formatStepLabel(step) }}</text>
        </view>
      </view>
    </view>

    <view class="card">
      <text class="section-title">添加下一层关系</text>
      <view class="relation-grid">
        <button
          v-for="option in relationOptions"
          :key="option.relation"
          class="button"
          :class="{ secondary: !option.enabled, disabled: !option.enabled }"
          @click="selectRelation(option)"
        >
          {{ option.label }}
        </button>
      </view>
      <text v-if="selectedRelationLabel" class="muted">已选择：{{ selectedRelationLabel }}</text>

      <view v-if="pendingRelation" class="pending-form">
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
          <text class="label">长幼</text>
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
        <button class="button" @click="appendStep">添加到路径</button>
        <text v-if="appendError" class="error">{{ appendError }}</text>
      </view>
    </view>

    <view class="card">
      <text class="section-title">推导结果</text>
      <text v-if="resolution.status === 'resolved'" class="result-title">{{ resolution.primaryTitle }}</text>
      <text v-else-if="resolution.status === 'ambiguous'" class="result-title">
        {{ resolution.primaryTitle || '信息不足' }}
      </text>
      <text v-else class="result-title">暂未收录该关系的常用称谓</text>

      <text v-if="resolution.candidates?.length" class="muted">
        可能称谓：{{ resolution.candidates.join('、') }}
      </text>
      <text v-if="resolution.aliases.length" class="muted">别称：{{ resolution.aliases.join('、') }}</text>
      <text v-if="resolution.explanation" class="muted">{{ resolution.explanation }}</text>
      <text v-if="resolution.status === 'unsupported'" class="muted">你仍然可以保留完整关系路径。</text>
      <text class="muted">路径：{{ resolution.pathDescription }}</text>
    </view>

    <view class="card">
      <button class="button secondary" :disabled="steps.length === 0" @click="undoStep">撤销一步</button>
      <button class="button secondary" @click="resetAll">重新开始</button>
      <button class="button" @click="copyDescription">复制关系描述</button>
    </view>

    <view class="card self-check-card">
      <text class="self-check-text">规则自检：{{ selfCheckSummary }}</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'

import { canAppendRelation, getRelationOptions } from '@/features/kinship/allowedRelations'
import {
  formatPathDisplay,
  formatStepLabel,
  normalizePersonFacts
} from '@/features/kinship/helpers'
import { resolveKinship } from '@/features/kinship/resolveKinship'
import { runKinshipSelfChecks } from '@/features/kinship/testCases'
import type { Gender, KinshipContext, KinshipRelation, KinshipStep, PersonFacts, RelativeAge } from '@/features/kinship/types'
import { MAX_KINSHIP_DEPTH } from '@/features/kinship/types'

const maxDepth = MAX_KINSHIP_DEPTH
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

const context = computed<KinshipContext>(() => ({
  self: normalizePersonFacts({
    ...self.value,
    age: selfAgeInput.value ? Number(selfAgeInput.value) : undefined
  }),
  steps: steps.value,
  maxDepth
}))

const relationOptions = computed(() => getRelationOptions(context.value))
const resolution = computed(() => resolveKinship(context.value))
const pathDisplay = computed(() => formatPathDisplay(context.value))
const selfCheckResult = computed(() => runKinshipSelfChecks())
const selfCheckSummary = computed(() => {
  const result = selfCheckResult.value
  return `${result.passed} / ${result.total} 通过`
})

const selectedRelationLabel = computed(() => {
  if (!pendingRelation.value) return ''
  return relationOptions.value.find((item) => item.relation === pendingRelation.value)?.label || ''
})

function selectRelation(option: { relation: KinshipRelation; enabled: boolean; disabledReason?: string }) {
  if (!option.enabled) {
    uni.showToast({
      title: option.disabledReason || '当前不可选择该关系',
      icon: 'none',
      duration: 2500
    })
    return
  }
  pendingRelation.value = option.relation
  pendingPerson.value = { gender: 'unknown' }
  pendingAgeInput.value = ''
  pendingRelativeAge.value = 'unknown'
  appendError.value = ''
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
  if (!pendingRelation.value) {
    appendError.value = '请先选择关系类型。'
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
}

function undoStep() {
  if (steps.value.length === 0) return
  steps.value = steps.value.slice(0, -1)
  appendError.value = ''
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
}

function buildCopyText() {
  const path = resolution.value.pathDescription
  if (resolution.value.status === 'resolved' && resolution.value.primaryTitle) {
    return `${path}：${resolution.value.primaryTitle}`
  }
  if (resolution.value.status === 'ambiguous') {
    const candidates = resolution.value.candidates?.join('、') || resolution.value.primaryTitle || '多种称谓'
    return `${path}：可能为${candidates}，请补充性别或长幼信息。`
  }
  return `${path}：暂未收录该关系的常用称谓。`
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
.field-hint {
  display: block;
  margin-bottom: 12rpx;
}

.picker-row {
  margin-top: 16rpx;
}

.label {
  display: block;
  margin-bottom: 8rpx;
  color: #6b7280;
  font-size: 24rpx;
}

.tag-group {
  display: flex;
  flex-wrap: wrap;
  gap: 8rpx;
}

.tag.selectable {
  cursor: pointer;
}

.tag.active {
  background: #1f3a5f;
  color: #fff;
}

.path-display {
  display: block;
  margin-bottom: 12rpx;
  font-size: 28rpx;
  line-height: 1.7;
}

.step-list {
  margin-top: 16rpx;
}

.step-item {
  padding: 12rpx 0;
  border-top: 1rpx solid #e5e0d6;
  font-size: 26rpx;
}

.relation-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12rpx;
}

.button.disabled,
.button[disabled] {
  opacity: 0.55;
}

.pending-form {
  margin-top: 20rpx;
  padding-top: 20rpx;
  border-top: 1rpx solid #e5e0d6;
}

.result-title {
  display: block;
  margin-bottom: 12rpx;
  color: #2f6b57;
  font-size: 34rpx;
  font-weight: 700;
}

.error {
  display: block;
  margin-top: 12rpx;
  color: #c0392b;
  font-size: 24rpx;
}

.self-check-card {
  text-align: center;
}

.self-check-text {
  color: #9ca3af;
  font-size: 22rpx;
}
</style>
