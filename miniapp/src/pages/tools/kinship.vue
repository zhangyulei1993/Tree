<template>
  <view class="tree-page kinship-page">
    <MiniBackHome />
    <view class="tree-tool-banner kinship-banner">
      <view class="banner-copy">
        <text class="tree-tool-banner-title">亲属关系工具</text>
        <text class="tree-tool-banner-desc">{{ maxDepthHint }}</text>
      </view>
      <view class="tree-pedigree-mark" aria-hidden="true">
        <view class="node node-root" />
        <view class="line-v" />
        <view class="line-l" />
        <view class="line-r" />
        <view class="node node-branch node-left" />
        <view class="node node-branch node-right" />
        <view class="trunk" />
      </view>
    </view>

    <MiniNotice tone="info" class="beta-notice">规则持续完善中</MiniNotice>
    <MiniNotice tone="security" class="local-only-notice" title="本地处理说明">
      本工具仅在当前设备本地推测称谓，不上传你的性别、年龄或关系路径信息至服务器。
    </MiniNotice>

    <MiniCard>
      <MiniSectionHeader title="本人信息" subtitle="出生日期优先用于判断长幼，年龄仅作为补充。" />
      <view class="picker-row">
        <text class="tree-field-label">本人性别</text>
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
      <input v-model.trim="self.birthday" class="tree-input" placeholder="本人出生日期（可选，如 1990-01-01）" />
      <input v-model.trim="selfAgeInput" class="tree-input" type="number" placeholder="本人年龄（可选）" />
    </MiniCard>

    <MiniCard variant="soft" class="path-card">
      <MiniSectionHeader title="关系路径" subtitle="从「我」出发，沿谱系节点推导" />
      <view class="tree-lineage-chain">
        <view class="tree-lineage-node">
          <view class="tree-lineage-node-dot self">我</view>
          <text class="tree-lineage-node-label">本人</text>
        </view>
        <template v-for="(step, index) in steps" :key="index">
          <view class="tree-lineage-connector" />
          <view class="tree-lineage-node">
            <view class="tree-lineage-node-dot">{{ index + 1 }}</view>
            <text class="tree-lineage-node-label">{{ formatStepLabel(step) }}</text>
          </view>
        </template>
      </view>
      <text class="path-meta">当前代数：{{ generationDepth }} / {{ maxDepth }}</text>
    </MiniCard>

    <MiniCard>
      <MiniSectionHeader title="选择下一层关系" />
      <view class="relation-grid">
        <view
          v-for="option in relationOptions"
          :key="option.relation"
          class="tree-relation-card"
          :class="{
            active: pendingRelation === option.relation,
            disabled: !option.enabled
          }"
          @click="selectRelation(option)"
        >
          <text class="tree-relation-card-label">{{ option.label }}</text>
        </view>
      </view>
      <text v-if="lastDisabledReason" class="disabled-hint">{{ lastDisabledReason }}</text>
    </MiniCard>

    <MiniCard>
      <MiniSectionHeader title="补充这个人的信息" />
      <view v-if="!pendingRelation" class="empty-hint">
        <text class="tree-muted">请先选择下一层关系。</text>
      </view>
      <view v-else class="pending-form">
        <text class="pending-relation-hint">正在为「{{ selectedRelationLabel }}」补充信息</text>
        <view class="picker-row">
          <text class="tree-field-label">性别</text>
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
        <input v-model.trim="pendingPerson.birthday" class="tree-input" placeholder="出生日期（可选）" />
        <input v-model.trim="pendingAgeInput" class="tree-input" type="number" placeholder="年龄（可选）" />
        <view v-if="pendingRelation === 'sibling'" class="picker-row">
          <text class="tree-field-label">长幼（相对上一位）</text>
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
        <MiniButton @click="appendStep">添加到路径</MiniButton>
        <text v-if="appendError" class="tree-field-error">{{ appendError }}</text>
      </view>
    </MiniCard>

    <view class="tree-result-plaque result-card" :class="`result-${resolution.status}`">
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

    <MiniCard class="action-card">
      <view class="action-row">
        <MiniButton
          variant="secondary"
          size="sm"
          :block="false"
          class="action-btn"
          :class="{ 'btn-weak': steps.length === 0 }"
          @click="undoStep"
        >
          撤销一步
        </MiniButton>
        <MiniButton variant="secondary" size="sm" :block="false" class="action-btn" @click="confirmReset">
          重新开始
        </MiniButton>
        <MiniButton class="action-btn-wide" @click="copyDescription">复制关系描述</MiniButton>
      </view>
    </MiniCard>

    <view class="footer-check">
      <text class="self-check-text">规则自检：{{ selfCheckSummary }}</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'

import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniCard from '@/components/base/MiniCard.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
import MiniSectionHeader from '@/components/base/MiniSectionHeader.vue'
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
.path-card :deep(.mini-section-header) {
  margin-bottom: 12rpx;
}

.picker-row {
  margin-top: 12rpx;
}

.tag-group {
  display: flex;
  flex-wrap: wrap;
  gap: 10rpx;
}

.tag.selectable {
  min-width: 72rpx;
  min-height: 56rpx;
  align-items: center;
  justify-content: center;
  box-sizing: border-box;
  cursor: pointer;
}

.tag.active {
  background: var(--tree-primary, #1f3a5f);
  color: #fff;
}

.path-pills,
.path-pill,
.path-pill-self,
.path-separator {
  display: none;
}

.result-card {
  margin-bottom: 20rpx;
}

.path-meta {
  display: block;
  margin-top: 12rpx;
  color: var(--tree-text-secondary);
  font-size: 24rpx;
}

.relation-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 12rpx;
}

.tree-relation-card {
  min-height: 92rpx;
  align-items: center;
  justify-content: center;
}

.disabled-hint {
  display: block;
  margin-top: 16rpx;
  color: #9ca3af;
  font-size: 24rpx;
  line-height: 1.6;
}

.empty-hint {
  padding: 16rpx 0 8rpx;
}

.pending-form {
  padding-top: 4rpx;
}

.pending-relation-hint {
  display: block;
  margin-bottom: 16rpx;
  color: #2f6b57;
  font-size: 26rpx;
}

.result-card.result-card {
  border: none;
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
  padding-bottom: 12rpx;
}

.action-row {
  display: flex;
  flex-wrap: wrap;
  gap: 12rpx;
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

.beta-notice {
  margin-bottom: 16rpx;
}

.kinship-banner,
.genealogy-banner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16rpx;
}

.banner-copy {
  flex: 1;
  min-width: 0;
}

.footer-check {
  padding: 4rpx 0 24rpx;
  text-align: center;
}

.self-check-text {
  color: #d1d5db;
  font-size: 18rpx;
}
</style>
