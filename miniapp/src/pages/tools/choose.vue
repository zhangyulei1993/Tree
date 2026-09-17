<template>
  <view class="archive-page choose-page">
    <MiniBackHome />

    <view class="choose-head archive-page-head">
      <view>
        <text class="archive-kicker">Decision Tool</text>
        <text class="archive-title">该选什么</text>
        <text class="archive-subtitle">把选项交给随机，帮你做个决定</text>
      </view>
      <view class="archive-seal">选</view>
    </view>

    <view class="choose-panel archive-panel">
      <view class="archive-section-head">
        <text class="archive-section-title">输入选项</text>
        <text class="archive-section-subtitle">每行一个，也支持用逗号分隔</text>
      </view>
      <text class="tree-field-label">常用情景</text>
      <view class="scenario-list">
        <text
          v-for="scenario in scenarios"
          :key="scenario.name"
          class="scenario-chip"
          :class="{ active: activeScenarioId === scenario.id }"
          @click="applyScenario(scenario)"
        >
          {{ scenario.name }}
        </text>
      </view>
      <view class="shared-scenario-block">
        <view class="shared-scenario-head">
          <view class="shared-scenario-summary" @click="toggleSharedScenarios">
            <text class="tree-field-label">大家的场景</text>
            <text class="shared-scenario-hint">看看别人正在怎么选，也可以直接套用</text>
          </view>
          <text class="shared-toggle" @click="toggleSharedScenarios">{{ sharedExpanded ? '收起' : '展开' }}</text>
        </view>
        <view v-if="sharedExpanded" class="shared-scenario-content">
          <view class="shared-scenario-toolbar">
            <view class="period-list">
              <text
                v-for="period in sharedPeriods"
                :key="period.value"
                class="period-chip"
                :class="{ active: sharedPeriod === period.value }"
                @click="changeSharedPeriod(period.value)"
              >
                {{ period.label }}
              </text>
            </view>
            <text class="shared-refresh" @click="loadSharedScenarios">刷新</text>
          </view>
          <view v-if="sharedScenarios.length > 0" class="shared-scenario-list">
            <view
              v-for="item in sharedScenarios"
              :key="item.id"
              class="shared-scenario-item"
              @click="applySharedScenario(item)"
            >
              <view class="shared-scenario-copy">
                <text class="shared-scenario-title">{{ item.title }}</text>
                <text class="shared-scenario-options">{{ item.options.join(' · ') }}</text>
              </view>
              <text class="shared-scenario-action">套用 ›</text>
            </view>
          </view>
          <text v-else class="shared-empty">{{ sharedScenarioMessage || '这个时间段还没有共享场景' }}</text>
        </view>
      </view>
      <text class="tree-field-label title-label">随机标题</text>
      <input
        v-model="choiceTitle"
        class="choice-title-input"
        maxlength="20"
        placeholder="例如：下班去干什么"
        placeholder-class="choice-placeholder"
        @input="handleTitleInput"
      />
      <view class="choice-input-stage" :class="{ 'is-picking': isPicking, 'has-result': selectedOption }">
        <textarea
          v-model="rawOptions"
          class="choice-textarea"
          maxlength="500"
          auto-height
          :disabled="isPicking"
          placeholder="例如：\n火锅\n烧烤\n日料"
          placeholder-class="choice-placeholder"
          @focus="handleOptionsFocus"
          @input="handleOptionsInput"
        />
        <view v-if="isPicking || selectedOption" class="choice-random-overlay" aria-hidden="true">
          <text class="draw-kicker">{{ isPicking ? '正在随机' : '这次就选' }}</text>
          <text class="draw-option">{{ selectedOption }}</text>
        </view>
      </view>
      <view class="choice-meta">
        <text>{{ options.length }} / {{ MAX_CHOICE_OPTIONS }} 个选项</text>
        <text>单项不超过 30 字</text>
      </view>
      <text v-if="errorMessage" class="choice-error">{{ errorMessage }}</text>
      <text v-if="selectedOption && !isPicking" class="choice-edit-hint">点击输入框可继续修改选项</text>
      <MiniButton size="md" :loading="isPicking" @click="makeChoice">随机帮我选</MiniButton>
      <text v-if="!activeScenarioId" class="publish-scenario" @click="publishScenario">分享这个场景</text>
    </view>

    <view class="choice-examples archive-panel">
      <view class="history-head">
        <view>
          <text class="examples-title">最近选择</text>
          <text class="history-hint">仅保存在当前设备，最多保留 10 条</text>
        </view>
        <text v-if="history.length > 0" class="clear-history" @click="confirmClearHistory">清空</text>
      </view>
      <view v-if="history.length > 0" class="history-list">
        <view v-for="item in history" :key="item.id" class="history-item" @click="reuseHistory(item)">
          <view class="history-main">
            <text class="history-result">{{ item.result }}</text>
            <text class="history-scene">{{ item.scenario }} · {{ formatHistoryTime(item.createdAt) }}</text>
          </view>
          <text class="history-action">再选 ›</text>
        </view>
      </view>
      <view v-else class="history-empty">
        <text>还没有选择记录</text>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { onLoad, onUnload } from '@dcloudio/uni-app'
import { computed, ref } from 'vue'

import {
  listSharedChoiceScenarios,
  publishSharedChoiceScenario,
  recordSharedChoiceScenarioUse,
  type ChoiceScenarioPeriod,
  type SharedChoiceScenario
} from '@/api/choiceScenarios'
import { apiErrorMessage } from '@/api/client'
import { ensureToolEnabled } from '@/api/toolConfigs'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import {
  CHOICE_OPTION_MAX_LENGTH,
  CHOICE_TITLE_MAX_LENGTH,
  MAX_CHOICE_OPTIONS,
  MAX_CHOICE_OPTIONS_TEXT_LENGTH,
  parseChoiceOptions,
  pickRandomChoice
} from '@/features/choice/choiceOptions'
import { useSessionStore } from '@/stores/session'

interface ChoiceScenario {
  id: string
  name: string
  options: string[]
}

interface ChoiceHistoryItem {
  id: string
  scenario: string
  result: string
  options: string[]
  createdAt: number
}

const CHOICE_HISTORY_STORAGE_KEY = 'tree_choice_history_v1'
const rawOptions = ref('')
const choiceTitle = ref('')
const selectedOption = ref('')
const errorMessage = ref('')
const activeScenarioId = ref('')
const history = ref<ChoiceHistoryItem[]>([])
const sharedScenarios = ref<SharedChoiceScenario[]>([])
const sharedScenarioMessage = ref('')
const sharedPeriod = ref<ChoiceScenarioPeriod>('today')
const sharedExpanded = ref(false)
const isPublishing = ref(false)
const isPicking = ref(false)
const session = useSessionStore()
let pickTimer: ReturnType<typeof setInterval> | null = null
const options = computed(() => parseChoiceOptions(rawOptions.value))
const sharedPeriods: Array<{ label: string; value: ChoiceScenarioPeriod }> = [
  { label: '今天', value: 'today' },
  { label: '近 7 天', value: '7d' },
  { label: '本月', value: 'month' }
]
const scenarios: ChoiceScenario[] = [
  { id: 'builtin_meal', name: '今天吃什么', options: ['家常菜', '火锅', '烧烤', '面食', '外卖'] },
  { id: 'builtin_after_work', name: '下班去干什么', options: ['打篮球', '去健身房', '逛街', '看电影', '回家休息'] },
  { id: 'builtin_weekend', name: '周末去哪儿', options: ['公园散步', '看电影', '逛街', '周边短途游', '在家休息'] },
  { id: 'builtin_movie', name: '今晚看什么', options: ['纪录片', '喜剧片', '电视剧', '综艺', '读书'] },
  { id: 'builtin_priority', name: '先做哪一件事', options: ['工作任务', '运动锻炼', '整理房间', '回复消息', '休息一下'] }
]

onLoad(() => {
  loadHistory()
  void loadSharedScenarios()
  void ensureToolEnabled('CHOOSE')
})

onUnload(() => {
  stopPicking()
})

function handleOptionsInput() {
  const wasPresetScenario = Boolean(activeScenarioId.value)
  selectedOption.value = ''
  errorMessage.value = ''
  activeScenarioId.value = ''
  if (wasPresetScenario) choiceTitle.value = ''
}

function handleOptionsFocus() {
  if (!isPicking.value) selectedOption.value = ''
}

function handleTitleInput() {
  selectedOption.value = ''
  errorMessage.value = ''
  const activeScenario = scenarios.find((item) => item.id === activeScenarioId.value)
  if (!activeScenario || choiceTitle.value.trim() !== activeScenario.name) activeScenarioId.value = ''
}

function applyScenario(scenario: ChoiceScenario) {
  rawOptions.value = scenario.options.join('\n')
  activeScenarioId.value = scenario.id
  choiceTitle.value = scenario.name
  selectedOption.value = ''
  errorMessage.value = ''
}

function applySharedScenario(scenario: SharedChoiceScenario) {
  rawOptions.value = scenario.options.join('\n')
  choiceTitle.value = scenario.title
  const builtinScenario = findBuiltinScenario(scenario.title, scenario.options)
  activeScenarioId.value = builtinScenario?.id || ''
  selectedOption.value = ''
  errorMessage.value = ''
  if (session.isLoggedIn) void recordSharedChoiceScenarioUse(scenario.id).catch(() => undefined)
}

function makeChoice() {
  if (isPicking.value) return
  errorMessage.value = ''
  if (options.value.length < 2) {
    errorMessage.value = '至少输入两个选项，才能开始选择。'
    uni.showToast({ title: '至少输入两个选项', icon: 'none' })
    return
  }
  if (!choiceTitle.value.trim()) {
    errorMessage.value = '请先填写随机标题，例如：下班去干什么。'
    uni.showToast({ title: '请先填写随机标题', icon: 'none' })
    return
  }
  startPicking([...options.value], selectedOption.value)
}

function startPicking(choiceOptions: string[], previousOption: string) {
  if (isPicking.value || choiceOptions.length < 2) return
  let currentOption = pickRandomChoice(choiceOptions, Math.random, previousOption)
  selectedOption.value = currentOption
  isPicking.value = true
  let ticks = 0
  pickTimer = setInterval(() => {
    currentOption = pickRandomChoice(choiceOptions, Math.random, currentOption)
    selectedOption.value = currentOption
    ticks += 1
    if (ticks >= 10) {
      stopPicking()
      saveHistory(currentOption)
      uni.showToast({ title: `已选：${currentOption}`, icon: 'none' })
    }
  }, 70)
}

function stopPicking() {
  if (pickTimer) {
    clearInterval(pickTimer)
    pickTimer = null
  }
  isPicking.value = false
}

function loadHistory() {
  try {
    const stored = uni.getStorageSync(CHOICE_HISTORY_STORAGE_KEY)
    const parsed = typeof stored === 'string' ? JSON.parse(stored) : stored
    if (!Array.isArray(parsed)) return
    history.value = parsed
      .filter((item): item is ChoiceHistoryItem => {
        return Boolean(
          item &&
          typeof item.id === 'string' &&
          typeof item.scenario === 'string' &&
          typeof item.result === 'string' &&
          Array.isArray(item.options) &&
          item.options.every((option: unknown) => typeof option === 'string') &&
          typeof item.createdAt === 'number'
        )
      })
      .slice(0, 10)
  } catch {
    history.value = []
  }
}

function saveHistory(result: string) {
  const item: ChoiceHistoryItem = {
    id: `${Date.now()}-${Math.floor(Math.random() * 10000)}`,
    scenario: choiceTitle.value.trim() || '自定义选择',
    result,
    options: [...options.value],
    createdAt: Date.now()
  }
  history.value = [item, ...history.value].slice(0, 10)
  uni.setStorageSync(CHOICE_HISTORY_STORAGE_KEY, JSON.stringify(history.value))
}

async function loadSharedScenarios() {
  sharedScenarioMessage.value = ''
  try {
    const loadedScenarios = await listSharedChoiceScenarios(sharedPeriod.value)
    sharedScenarios.value = loadedScenarios.filter((item) => !findBuiltinScenario(item.title, item.options))
  } catch {
    sharedScenarios.value = []
    sharedScenarioMessage.value = '共享场景暂时不可用，请稍后再试。'
  }
}

function toggleSharedScenarios() {
  sharedExpanded.value = !sharedExpanded.value
}

async function changeSharedPeriod(period: ChoiceScenarioPeriod) {
  if (sharedPeriod.value === period) return
  sharedPeriod.value = period
  await loadSharedScenarios()
}

async function publishScenario() {
  if (isPublishing.value) return
  errorMessage.value = ''
  if (!session.requireLogin('/pages/tools/choose')) return
  if (activeScenarioId.value) {
    uni.showToast({ title: '内置场景无需发布', icon: 'none' })
    return
  }
  const title = choiceTitle.value.trim()
  const optionTextLength = Array.from(rawOptions.value).length
  if (!title) {
    errorMessage.value = '请先填写随机标题，例如：下班去干什么。'
    uni.showToast({ title: '请先填写随机标题', icon: 'none' })
    return
  }
  if (Array.from(title).length > CHOICE_TITLE_MAX_LENGTH) {
    errorMessage.value = `标题最多 ${CHOICE_TITLE_MAX_LENGTH} 个字。`
    uni.showToast({ title: '标题太长了', icon: 'none' })
    return
  }
  if (options.value.length < 2) {
    errorMessage.value = '至少输入两个选项，才能发布场景。'
    uni.showToast({ title: '至少输入两个选项', icon: 'none' })
    return
  }
  if (options.value.some((option) => Array.from(option).length > CHOICE_OPTION_MAX_LENGTH)) {
    errorMessage.value = `每个选项最多 ${CHOICE_OPTION_MAX_LENGTH} 个字。`
    uni.showToast({ title: '有选项太长了', icon: 'none' })
    return
  }
  if (optionTextLength > MAX_CHOICE_OPTIONS_TEXT_LENGTH) {
    errorMessage.value = `选项总长度最多 ${MAX_CHOICE_OPTIONS_TEXT_LENGTH} 个字。`
    uni.showToast({ title: '选项内容太长了', icon: 'none' })
    return
  }
  isPublishing.value = true
  try {
    const created = await publishSharedChoiceScenario(title, [...options.value])
    sharedScenarios.value = [created, ...sharedScenarios.value.filter((item) => item.id !== created.id)].slice(0, 30)
    uni.showToast({ title: '已分享这个场景', icon: 'none' })
  } catch (error) {
    errorMessage.value = apiErrorMessage(error, '发布失败，请稍后重试。')
    uni.showToast({ title: errorMessage.value, icon: 'none' })
  } finally {
    isPublishing.value = false
  }
}

function formatHistoryTime(timestamp: number) {
  const date = new Date(timestamp)
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  return `${month}月${day}日 ${hours}:${minutes}`
}

function reuseHistory(item: ChoiceHistoryItem) {
  if (isPicking.value) return
  rawOptions.value = item.options.join('\n')
  choiceTitle.value = item.scenario === '自定义选择' ? '' : item.scenario
  const builtinScenario = findBuiltinScenario(choiceTitle.value, item.options)
  activeScenarioId.value = builtinScenario?.id || ''
  errorMessage.value = ''
  if (!choiceTitle.value) choiceTitle.value = '自定义选择'
  startPicking([...item.options], item.result)
}

function findBuiltinScenario(title: string, scenarioOptions: string[]) {
  return scenarios.find(
    (scenario) =>
      scenario.name === title &&
      scenario.options.length === scenarioOptions.length &&
      scenario.options.every((option, index) => option === scenarioOptions[index])
  )
}

function confirmClearHistory() {
  uni.showModal({
    title: '清空选择历史',
    content: '只会清除当前设备上的选择记录。',
    confirmText: '清空',
    cancelText: '取消',
    success: (result) => {
      if (!result.confirm) return
      history.value = []
      uni.removeStorageSync(CHOICE_HISTORY_STORAGE_KEY)
    }
  })
}
</script>

<style scoped>
.choose-page {
  padding-top: 28rpx;
}

.choose-page :deep(.mini-back-home) {
  margin-bottom: 22rpx;
}

.choose-panel,
.choice-examples {
  margin-top: 18rpx;
  padding: 22rpx 20rpx;
}

.archive-section-head {
  margin-bottom: 12rpx;
}

.archive-section-subtitle {
  display: block;
  margin-top: 6rpx;
  color: var(--archive-ink-soft);
  font-size: 22rpx;
}

.scenario-list {
  display: flex;
  flex-wrap: wrap;
  gap: 10rpx;
  margin-top: 14rpx;
}

.tree-field-label {
  display: block;
  margin-top: 14rpx;
  color: var(--archive-cinnabar);
  font-size: 21rpx;
  letter-spacing: 1rpx;
}

.title-label {
  margin-top: 20rpx;
}

.scenario-chip {
  border: 1rpx solid var(--archive-line);
  background: rgba(255, 249, 236, 0.46);
  color: var(--archive-ink-soft);
  padding: 9rpx 14rpx;
  font-size: 22rpx;
}

.scenario-chip.active {
  border-color: var(--archive-blue);
  background: var(--archive-blue-soft);
  color: var(--archive-blue);
}

.shared-scenario-block {
  margin-top: 22rpx;
  border-top: 1rpx solid var(--archive-line);
  padding-top: 16rpx;
}

.shared-scenario-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16rpx;
}

.shared-scenario-summary {
  min-width: 0;
  flex: 1;
}

.shared-toggle {
  flex-shrink: 0;
  color: var(--archive-blue);
  font-size: 22rpx;
}

.shared-scenario-hint {
  display: block;
  margin-top: 5rpx;
  color: var(--archive-ink-soft);
  font-size: 20rpx;
}

.shared-refresh,
.publish-scenario {
  color: var(--archive-blue);
  font-size: 22rpx;
}

.shared-scenario-toolbar {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 16rpx;
}

.shared-scenario-toolbar .period-list {
  margin-top: 12rpx;
}

.period-list {
  display: flex;
  gap: 10rpx;
  margin-top: 12rpx;
}

.period-chip {
  border: 1rpx solid var(--archive-line);
  padding: 7rpx 14rpx;
  color: var(--archive-ink-soft);
  font-size: 20rpx;
}

.period-chip.active {
  border-color: var(--archive-blue);
  background: var(--archive-blue-soft);
  color: var(--archive-blue);
}

.shared-scenario-list {
  margin-top: 12rpx;
}

.shared-scenario-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14rpx;
  border-top: 1rpx solid var(--archive-line);
  padding: 13rpx 0;
}

.shared-scenario-copy {
  min-width: 0;
  flex: 1;
}

.shared-scenario-title,
.shared-scenario-options {
  display: block;
}

.shared-scenario-title {
  color: var(--archive-ink);
  font-size: 23rpx;
  font-weight: 650;
}

.shared-scenario-options {
  overflow: hidden;
  margin-top: 4rpx;
  color: var(--archive-ink-soft);
  font-size: 19rpx;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.shared-scenario-action {
  flex-shrink: 0;
  color: var(--archive-blue);
  font-size: 20rpx;
}

.shared-empty {
  display: block;
  border-top: 1rpx solid var(--archive-line);
  margin-top: 12rpx;
  padding-top: 13rpx;
  color: var(--archive-ink-soft);
  font-size: 20rpx;
}

.choice-input-stage {
  position: relative;
  margin-top: 16rpx;
}

.choice-textarea {
  width: 100%;
  min-height: 220rpx;
  box-sizing: border-box;
  margin-top: 0;
  border: 1rpx solid var(--archive-line-strong);
  border-radius: 0;
  background: rgba(255, 249, 236, 0.48);
  color: var(--archive-ink);
  padding: 18rpx;
  font-size: 27rpx;
  line-height: 1.7;
}

.choice-input-stage.is-picking .choice-textarea,
.choice-input-stage.has-result .choice-textarea {
  color: transparent;
  text-shadow: 0 0 0 transparent;
}

.choice-random-overlay {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-direction: column;
  pointer-events: none;
  padding: 18rpx;
  text-align: center;
}

.choice-input-stage.is-picking .choice-random-overlay {
  animation: choice-picking 0.14s ease-in-out infinite alternate;
}

.choice-placeholder {
  color: var(--archive-ink-soft);
}

.choice-title-input {
  width: 100%;
  min-height: 72rpx;
  box-sizing: border-box;
  margin-top: 8rpx;
  border-bottom: 1rpx solid var(--archive-line-strong);
  color: var(--archive-ink);
  padding: 12rpx 4rpx;
  font-size: 27rpx;
}

.choice-meta {
  display: flex;
  justify-content: space-between;
  gap: 16rpx;
  margin: 10rpx 0 18rpx;
  color: var(--archive-ink-soft);
  font-size: 21rpx;
}

.choice-error {
  display: block;
  margin: -4rpx 0 14rpx;
  color: var(--archive-cinnabar);
  font-size: 22rpx;
}

.choice-edit-hint {
  display: block;
  margin: -4rpx 0 12rpx;
  color: var(--archive-ink-soft);
  font-size: 20rpx;
  text-align: center;
}

.publish-scenario {
  display: block;
  margin-top: 16rpx;
  text-align: center;
}

.draw-kicker,
.draw-option {
  display: block;
}

.draw-kicker {
  color: var(--archive-cinnabar);
  font-size: 20rpx;
  letter-spacing: 2rpx;
}

.draw-option {
  overflow: hidden;
  margin-top: 8rpx;
  color: var(--archive-ink-soft);
  font-family: 'Songti SC', 'STSong', serif;
  font-size: 30rpx;
  font-weight: 650;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.choice-input-stage.has-result .draw-option {
  color: var(--archive-ink);
  font-size: 42rpx;
  line-height: 1.25;
  white-space: normal;
}

.choice-input-stage.is-picking .draw-option {
  color: var(--archive-cinnabar);
}

.choice-examples {
  border-top: 1rpx solid var(--archive-line-strong);
  border-bottom: 1rpx solid var(--archive-line);
}

.history-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16rpx;
}

.examples-title {
  display: block;
  color: var(--archive-ink);
  font-size: 24rpx;
  font-weight: 650;
}

.history-hint {
  display: block;
  margin-top: 5rpx;
  color: var(--archive-ink-soft);
  font-size: 20rpx;
}

.clear-history {
  color: var(--archive-cinnabar);
  font-size: 22rpx;
}

.history-list {
  margin-top: 14rpx;
}

.history-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14rpx;
  border-top: 1rpx solid var(--archive-line);
  padding: 16rpx 0;
}

.history-main {
  min-width: 0;
  flex: 1;
}

.history-result,
.history-scene {
  display: block;
}

.history-result {
  overflow: hidden;
  color: var(--archive-ink);
  font-size: 26rpx;
  font-weight: 650;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.history-scene {
  margin-top: 5rpx;
  color: var(--archive-ink-soft);
  font-size: 20rpx;
}

.history-action {
  flex-shrink: 0;
  color: var(--archive-blue);
  font-size: 21rpx;
}

.history-empty {
  margin-top: 14rpx;
  border-top: 1rpx solid var(--archive-line);
  padding-top: 16rpx;
  color: var(--archive-ink-soft);
  font-size: 22rpx;
}

@keyframes choice-picking {
  from {
    transform: translateY(0);
  }
  to {
    transform: translateY(-3rpx);
  }
}
</style>
