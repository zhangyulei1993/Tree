<template>
  <view class="archive-page kinship-page">
    <MiniBackHome />

    <view class="kinship-head archive-page-head">
      <view>
        <text class="archive-kicker">Kinship Tool</text>
        <text class="archive-title">亲属关系工具</text>
        <text class="archive-subtitle">{{ maxDepthHint }}</text>
      </view>
      <view class="archive-seal">亲</view>
    </view>

    <MiniNotice tone="info" class="beta-notice">规则持续完善中</MiniNotice>
    <MiniNotice tone="security" class="local-only-notice" title="本地处理说明">
      本工具仅在当前设备本地推测称谓，不上传你的性别、年龄或关系路径信息至服务器。
    </MiniNotice>

    <view class="archive-form-panel kinship-panel">
      <view class="archive-section-head">
        <text class="archive-section-title">本人信息</text>
        <text class="archive-section-subtitle">出生日期优先用于判断长幼，年龄仅作为补充</text>
      </view>
      <text class="tree-field-label">本人性别</text>
      <view class="choice-row">
        <text
          v-for="item in selfGenderOptions"
          :key="item.value"
          class="choice-chip"
          :class="{ active: self.gender === item.value }"
          @click="self.gender = item.value"
        >
          {{ item.label }}
        </text>
      </view>
      <text v-if="self.gender === 'unknown'" class="disabled-hint">请先选择本人性别，再开始选择关系。</text>
      <text class="tree-field-label">出生日期（可选）</text>
      <input v-model.trim="self.birthday" class="tree-input" placeholder="如 1990-01-01" />
      <text class="tree-field-label">年龄（可选）</text>
      <input v-model.trim="selfAgeInput" class="tree-input" type="number" placeholder="请输入年龄" />
    </view>

    <view class="archive-form-panel kinship-panel">
      <view class="archive-section-head">
        <text class="archive-section-title">关系路径</text>
        <text class="archive-section-subtitle">从「我」出发，沿家庭关系推导</text>
      </view>
      <view class="path-tag-chain">
        <text class="path-tag path-tag-self">我</text>
        <template v-for="(step, index) in steps" :key="index">
          <text class="path-tag-link">—</text>
          <text class="path-tag">{{ formatStepLabel(step) }}</text>
        </template>
      </view>
      <text class="path-meta">当前代数：{{ generationDepth }} / {{ maxDepth }}</text>
    </view>

    <view class="archive-form-panel kinship-panel">
      <view class="archive-section-head">
        <text class="archive-section-title">选择下一层关系</text>
        <text class="archive-section-subtitle">点击纸签添加下一层亲属节点</text>
      </view>
      <view class="relation-grid">
        <text
          v-for="option in relationOptions"
          :key="option.relation"
          class="relation-chip"
          :class="{
            active: pendingRelation === option.relation,
            disabled: !option.enabled
          }"
          @click="selectRelation(option)"
        >
          {{ option.label }}
        </text>
      </view>
      <text v-if="lastDisabledReason" class="disabled-hint">{{ lastDisabledReason }}</text>
    </view>

    <view class="archive-form-panel kinship-panel">
      <view class="archive-section-head">
        <text class="archive-section-title">补充这个人的信息</text>
        <text class="archive-section-subtitle">性别、出生日期与长幼关系</text>
      </view>
      <view v-if="!pendingRelation" class="empty-hint">
        <text class="tree-muted">请先选择下一层关系。</text>
      </view>
      <view v-else class="pending-form">
        <text class="pending-relation-hint">正在为「{{ selectedRelationLabel }}」补充信息</text>
        <text class="tree-field-label">性别</text>
        <view class="choice-row">
          <text
            v-for="item in genderOptions"
            :key="item.value"
            class="choice-chip"
            :class="{ active: pendingPerson.gender === item.value }"
            @click="pendingPerson.gender = item.value"
          >
            {{ item.label }}
          </text>
        </view>
        <text class="tree-field-label">出生日期（可选）</text>
        <input v-model.trim="pendingPerson.birthday" class="tree-input" placeholder="如 1990-01-01" />
        <text class="tree-field-label">年龄（可选）</text>
        <input v-model.trim="pendingAgeInput" class="tree-input" type="number" placeholder="请输入年龄" />
        <view v-if="pendingRelation === 'sibling'" class="relative-age-block">
          <text class="tree-field-label">长幼（相对上一位）</text>
          <view class="choice-row">
            <text
              v-for="item in relativeAgeOptions"
              :key="item.value"
              class="choice-chip"
              :class="{ active: pendingRelativeAge === item.value }"
              @click="pendingRelativeAge = item.value"
            >
              {{ item.label }}
            </text>
          </view>
        </view>
        <MiniButton class="append-action" size="sm" @click="appendStep">添加到路径</MiniButton>
        <text v-if="appendError" class="tree-field-error">{{ appendError }}</text>
      </view>
    </view>

    <view class="kinship-result" :class="`result-${resolution.status}`">
      <view class="result-head">
        <text class="result-seal">称谓</text>
        <text class="result-kicker">批注结果</text>
      </view>
      <view class="result-body">
        <template v-if="resolution.status === 'resolved'">
          <text class="result-label">规范称谓</text>
          <text class="result-title">{{ resolution.primaryTitle }}</text>
        </template>
        <template v-else-if="resolution.incompleteInfo">
          <text class="result-label">信息不足</text>
          <text class="result-unsupported">关系信息不完整</text>
          <text class="tree-muted result-hint">请补充性别、长幼或父母方向等客观信息。</text>
        </template>
        <template v-else>
          <text class="result-label">客观关系路径</text>
          <text class="result-title muted-title">{{ resolution.pathDescription }}</text>
          <text class="tree-muted result-hint">当前路径暂无对应规范称谓。</text>
        </template>
        <text v-if="resolution.explanation" class="explanation-text">{{ resolution.explanation }}</text>
      </view>
    </view>

    <view class="kinship-actions">
      <MiniButton
        variant="secondary"
        size="sm"
        class="action-btn"
        :class="{ 'btn-weak': steps.length === 0 }"
        @click="undoStep"
      >
        撤销一步
      </MiniButton>
      <MiniButton variant="secondary" size="sm" class="action-btn" @click="confirmReset">
        重新开始
      </MiniButton>
      <MiniButton variant="ghost" size="sm" class="action-btn-wide" @click="copyDescription">
        复制关系描述
      </MiniButton>
    </view>

    <view class="footer-check">
      <text class="self-check-text">规则自检：{{ selfCheckSummary }}</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'

import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
import { canAppendRelation, getRelationOptions } from '@/features/kinship/allowedRelations'
import {
  formatStepLabel,
  normalizePersonFacts
} from '@/features/kinship/helpers'
import { resolveKinship } from '@/features/kinship/resolveKinship'
import { runKinshipSelfChecks } from '@/features/kinship/testCases'
import type { Gender, KinshipContext, KinshipRelation, KinshipStep, PersonFacts, RelativeAge } from '@/features/kinship/types'
import { MAX_KINSHIP_DEPTH } from '@/features/kinship/types'
import { validateDateInput } from '@/utils/inputValidation'

const maxDepth = MAX_KINSHIP_DEPTH
const maxDepthHint = `推导关系结果仅供参考，最多支持 ${MAX_KINSHIP_DEPTH} 层关系路径。`
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
  const selfBirthdayError = validateDateInput(self.value.birthday, '本人出生日期')
  if (selfBirthdayError) {
    appendError.value = selfBirthdayError
    return
  }
  if (!pendingRelation.value) {
    appendError.value = '请先选择下一层关系。'
    return
  }
  const pendingBirthdayError = validateDateInput(pendingPerson.value.birthday, '亲属出生日期')
  if (pendingBirthdayError) {
    appendError.value = pendingBirthdayError
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
  if (steps.value.length === 0) return '我'
  if (resolution.value.status === 'resolved' && resolution.value.primaryTitle) {
    return `${resolution.value.pathDescription}：${resolution.value.primaryTitle}`
  }
  if (resolution.value.incompleteInfo) {
    return `${resolution.value.pathDescription}：关系信息不完整`
  }
  return resolution.value.pathDescription
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
.kinship-page {
  padding-top: 28rpx;
}

.kinship-page :deep(.mini-back-home) {
  margin-bottom: 22rpx;
}

.beta-notice,
.local-only-notice {
  margin-bottom: 16rpx;
}

.kinship-panel {
  margin-bottom: 24rpx;
}

.choice-row {
  display: flex;
  flex-wrap: wrap;
  gap: 10rpx;
  margin-top: 8rpx;
}

.choice-chip {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 72rpx;
  min-height: 56rpx;
  box-sizing: border-box;
  border: 1rpx solid var(--archive-line);
  background: rgba(255, 248, 234, 0.58);
  color: var(--archive-ink-soft);
  padding: 0 18rpx;
  font-size: 24rpx;
}

.choice-chip.active {
  border-color: var(--archive-blue);
  background: rgba(22, 51, 83, 0.08);
  color: var(--archive-blue);
  font-weight: 650;
}

.path-tag-chain {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8rpx 6rpx;
  margin-top: 8rpx;
  padding: 18rpx 0;
  border-top: 1rpx solid var(--archive-line);
  border-bottom: 1rpx solid var(--archive-line);
}

.path-tag {
  display: inline-flex;
  align-items: center;
  min-height: 48rpx;
  border: 1rpx solid var(--archive-line-strong);
  background: rgba(255, 249, 236, 0.72);
  color: var(--archive-ink);
  padding: 0 16rpx;
  font-size: 24rpx;
  line-height: 1.3;
}

.path-tag-self {
  border-color: var(--archive-blue);
  background: rgba(22, 51, 83, 0.08);
  color: var(--archive-blue);
  font-weight: 700;
}

.path-tag-link {
  color: var(--archive-ink-soft);
  font-size: 22rpx;
}

.path-meta {
  display: block;
  margin-top: 12rpx;
  color: var(--archive-ink-soft);
  font-size: 22rpx;
}

.relation-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 10rpx;
  margin-top: 8rpx;
}

.relation-chip {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 108rpx;
  min-height: 72rpx;
  box-sizing: border-box;
  border: 1rpx solid var(--archive-line-strong);
  background: rgba(255, 249, 236, 0.42);
  color: var(--archive-ink);
  padding: 0 18rpx;
  font-size: 26rpx;
  line-height: 1.3;
  text-align: center;
}

.relation-chip.active {
  border-color: var(--archive-cinnabar);
  background: rgba(168, 59, 45, 0.08);
  color: var(--archive-cinnabar);
  font-weight: 650;
}

.relation-chip.disabled {
  opacity: 0.42;
}

.disabled-hint {
  display: block;
  margin-top: 16rpx;
  color: var(--archive-ink-soft);
  font-size: 24rpx;
  line-height: 1.6;
}

.empty-hint {
  padding: 8rpx 0;
}

.pending-form {
  padding-top: 4rpx;
}

.pending-relation-hint {
  display: block;
  margin-bottom: 12rpx;
  color: var(--archive-blue);
  font-size: 24rpx;
}

.relative-age-block {
  margin-top: 8rpx;
}

.append-action {
  margin-top: 18rpx;
  align-self: flex-start;
}

.kinship-result {
  margin-bottom: 24rpx;
  border-top: 1rpx solid rgba(168, 59, 45, 0.28);
  border-bottom: 1rpx solid rgba(168, 59, 45, 0.18);
  background: rgba(168, 59, 45, 0.04);
  padding: 24rpx 0 28rpx;
}

.result-head {
  display: flex;
  align-items: center;
  gap: 14rpx;
  margin-bottom: 16rpx;
}

.result-seal {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 56rpx;
  height: 56rpx;
  border: 2rpx solid var(--archive-cinnabar);
  color: var(--archive-cinnabar);
  font-family: 'Songti SC', 'STSong', serif;
  font-size: 24rpx;
  font-weight: 800;
}

.result-kicker {
  color: var(--archive-cinnabar);
  font-family: 'Songti SC', 'STSong', serif;
  font-size: 28rpx;
  font-weight: 700;
}

.result-body {
  border-top: 1rpx solid var(--archive-line);
  padding-top: 18rpx;
}

.result-label {
  display: block;
  margin-bottom: 8rpx;
  color: var(--archive-ink-soft);
  font-size: 22rpx;
  font-weight: 600;
}

.result-title {
  display: block;
  color: var(--archive-cinnabar);
  font-family: 'Songti SC', 'STSong', serif;
  font-size: 40rpx;
  font-weight: 800;
  line-height: 1.4;
}

.muted-title {
  color: var(--archive-ink);
  font-size: 32rpx;
  font-weight: 700;
}

.result-unsupported {
  display: block;
  color: var(--archive-ink-soft);
  font-size: 30rpx;
  font-weight: 650;
  line-height: 1.5;
}

.result-hint {
  display: block;
  margin-top: 12rpx;
}

.explanation-text {
  display: block;
  margin-top: 16rpx;
  padding-top: 16rpx;
  border-top: 1rpx solid var(--archive-line);
  color: var(--archive-ink-soft);
  font-size: 24rpx;
  line-height: 1.75;
}

.kinship-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 12rpx;
  margin-bottom: 20rpx;
}

.action-btn {
  flex: 1 1 calc(50% - 6rpx);
  min-width: 200rpx;
}

.action-btn-wide {
  flex: 1 1 100%;
}

.btn-weak {
  opacity: 0.5;
}

.footer-check {
  padding: 4rpx 0 24rpx;
  text-align: center;
}

.self-check-text {
  color: var(--archive-ink-soft);
  font-size: 18rpx;
  opacity: 0.55;
}
</style>
