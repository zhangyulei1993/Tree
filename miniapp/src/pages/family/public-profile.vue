<template>
  <view class="page">
    <view v-if="loadingFamily" class="card state-card">
      <text class="muted">正在加载公开家庭信息...</text>
    </view>
    <view v-else-if="familyError" class="card state-card">
      <text class="title">公开家庭</text>
      <text class="error">{{ familyError }}</text>
      <button class="button secondary" @click="loadFamily">重新加载</button>
    </view>
    <template v-else-if="family">
      <view class="card">
        <text class="title">{{ family.familyName }}</text>
        <view class="actions">
          <button class="button" @click="openPublicTree">查看公开家谱</button>
          <button class="button secondary" @click="openJoinApply">申请加入家庭</button>
        </view>
      </view>

      <view class="card">
        <text class="section-title">基本信息</text>
        <view class="info-row"><text class="label">家庭名称</text><text>{{ family.familyName }}</text></view>
        <view class="info-row"><text class="label">姓氏</text><text>{{ family.familySurname }}</text></view>
        <view class="info-row"><text class="label">籍贯</text><text>{{ family.nativePlace || '未设置' }}</text></view>
        <view class="info-row"><text class="label">地区</text><text>{{ family.regionText || '未设置' }}</text></view>
        <view class="info-block">
          <text class="label">简介</text>
          <text class="muted">{{ family.description || '该家庭暂未填写公开简介。' }}</text>
        </view>
      </view>

      <view class="card">
        <text class="section-title">公开联系方式</text>
        <template v-if="family.publicContactVisible && hasPublicContact">
          <view v-if="family.publicContactName" class="info-row">
            <text class="label">联系人</text><text>{{ family.publicContactName }}</text>
          </view>
          <view v-if="family.publicContactPhone" class="info-row">
            <text class="label">联系电话</text><text>{{ family.publicContactPhone }}</text>
          </view>
          <view v-if="family.publicContactWechat" class="info-row">
            <text class="label">微信</text><text>{{ family.publicContactWechat }}</text>
          </view>
          <view v-if="family.publicContactNote" class="info-block">
            <text class="label">备注</text>
            <text class="muted">{{ family.publicContactNote }}</text>
          </view>
        </template>
        <text v-else class="muted">该家庭暂未公开联系方式。</text>
      </view>

      <view class="card">
        <text class="section-title">公开留言</text>
        <text class="muted section-desc">审核通过的留言会显示在这里。</text>

        <view class="subsection">
          <text class="subsection-title">留言列表</text>
          <view v-if="loadingMessages" class="state-card">
            <text class="muted">正在加载留言...</text>
          </view>
          <view v-else-if="messagesError" class="state-card">
            <text class="error">{{ messagesError }}</text>
            <button class="button secondary" @click="loadMessages">重新加载留言</button>
          </view>
          <view v-else-if="messages.length === 0" class="message-empty">
            <text class="muted">暂无公开留言，审核通过的留言会显示在这里。</text>
          </view>
          <view v-for="message in messages" v-else :key="message.messageId" class="message">
            <text class="message-name">{{ message.visitorName || '匿名访客' }}</text>
            <text class="muted">{{ message.messageContent }}</text>
            <text class="message-time">{{ formatDate(message.reviewedAt || message.createdAt) }}</text>
          </view>
        </view>

        <view class="subsection">
          <text class="subsection-title">我要留言</text>
          <text class="muted section-desc">留言内容会经家庭管理员审核后公开显示。</text>
          <input v-model.trim="form.visitorName" class="input" maxlength="40" placeholder="你的称呼（可选）" />
          <text class="field-hint">联系电话和微信号不会公开展示，仅用于联系确认。</text>
          <input v-model.trim="form.visitorPhone" class="input" maxlength="30" placeholder="联系电话（可选）" />
          <input v-model.trim="form.visitorWechat" class="input" maxlength="100" placeholder="微信号（可选）" />
          <textarea
            v-model.trim="form.messageContent"
            class="textarea"
            maxlength="500"
            placeholder="请输入留言内容（必填）"
          />
          <text class="counter">{{ form.messageContent.length }}/500</text>
          <button class="button secondary" :disabled="submitting" :loading="submitting" @click="submitMessage">
            提交留言
          </button>
          <text v-if="formError" class="error">{{ formError }}</text>
          <text v-if="submitResult" class="result">{{ submitResult }}</text>
        </view>
      </view>
    </template>
  </view>
</template>

<script setup lang="ts">
import { onLoad } from '@dcloudio/uni-app'
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue'

import { apiErrorMessage } from '@/api/client'
import { getPublicFamilyDetail } from '@/api/families'
import { createVisitorMessage, listPublicVisitorMessages } from '@/api/visitorMessages'
import type { PublicFamily, PublicVisitorMessage } from '@/types/api'

const familyId = ref('')
const family = ref<PublicFamily | null>(null)
const messages = ref<PublicVisitorMessage[]>([])
const loadingFamily = ref(false)
const loadingMessages = ref(false)
const submitting = ref(false)
const familyError = ref('')
const messagesError = ref('')
const formError = ref('')
const submitResult = ref('')
let requestVersion = 0
const form = reactive({
  visitorName: '',
  visitorPhone: '',
  visitorWechat: '',
  messageContent: ''
})

const hasPublicContact = computed(() => Boolean(
  family.value?.publicContactName ||
  family.value?.publicContactPhone ||
  family.value?.publicContactWechat ||
  family.value?.publicContactNote
))

function formatDate(value: string) {
  const time = new Date(value)
  return Number.isNaN(time.getTime()) ? value : time.toLocaleString('zh-CN')
}

function openPublicTree() {
  uni.navigateTo({
    url: `/pages/family/public-tree?familyId=${encodeURIComponent(familyId.value)}`
  })
}

function openJoinApply() {
  uni.navigateTo({
    url: `/pages/join/apply?familyId=${encodeURIComponent(familyId.value)}`
  })
}

function resetPageState() {
  family.value = null
  messages.value = []
  familyError.value = ''
  messagesError.value = ''
  formError.value = ''
  submitResult.value = ''
  loadingFamily.value = false
  loadingMessages.value = false
  submitting.value = false
}

function familyIdFromCurrentHash() {
  if (typeof window === 'undefined') return ''
  const [, query = ''] = window.location.hash.split('?')
  return new URLSearchParams(query).get('familyId') || ''
}

function isCurrentProfileRoute() {
  if (typeof window === 'undefined') return false
  return window.location.hash.split('?')[0] === '#/pages/family/public-profile'
}

function reloadForFamilyId(nextFamilyId: string) {
  requestVersion += 1
  familyId.value = nextFamilyId.trim()
  resetPageState()

  if (!familyId.value) {
    familyError.value = '家庭公开信息暂不可访问。'
    return
  }

  loadFamily()
  loadMessages()
}

function handleHashChange() {
  if (!isCurrentProfileRoute()) return
  const nextFamilyId = familyIdFromCurrentHash()
  if (nextFamilyId !== familyId.value) {
    reloadForFamilyId(nextFamilyId)
  }
}

async function loadFamily() {
  if (!familyId.value) {
    family.value = null
    familyError.value = '家庭公开信息暂不可访问。'
    return
  }
  const version = requestVersion
  const currentFamilyId = familyId.value
  family.value = null
  loadingFamily.value = true
  familyError.value = ''
  try {
    const result = await getPublicFamilyDetail(currentFamilyId)
    if (version === requestVersion && currentFamilyId === familyId.value) {
      family.value = result
    }
  } catch (error) {
    if (version === requestVersion && currentFamilyId === familyId.value) {
      family.value = null
      familyError.value = apiErrorMessage(error, '家庭公开信息暂不可访问。')
    }
  } finally {
    if (version === requestVersion && currentFamilyId === familyId.value) {
      loadingFamily.value = false
    }
  }
}

async function loadMessages() {
  if (!familyId.value) return
  const version = requestVersion
  const currentFamilyId = familyId.value
  messages.value = []
  loadingMessages.value = true
  messagesError.value = ''
  try {
    const result = await listPublicVisitorMessages(currentFamilyId, { page: 1, pageSize: 20 })
    if (version === requestVersion && currentFamilyId === familyId.value) {
      messages.value = result.items
    }
  } catch (error) {
    if (version === requestVersion && currentFamilyId === familyId.value) {
      messages.value = []
      messagesError.value = apiErrorMessage(error, '公开留言加载失败，请稍后重试。')
    }
  } finally {
    if (version === requestVersion && currentFamilyId === familyId.value) {
      loadingMessages.value = false
    }
  }
}

async function submitMessage() {
  formError.value = ''
  submitResult.value = ''
  if (!form.messageContent.trim()) {
    formError.value = '请填写留言内容后再提交。'
    return
  }
  submitting.value = true
  try {
    await createVisitorMessage(familyId.value, {
      visitorName: form.visitorName || undefined,
      visitorPhone: form.visitorPhone || undefined,
      visitorWechat: form.visitorWechat || undefined,
      messageContent: form.messageContent
    })
    form.visitorName = ''
    form.visitorPhone = ''
    form.visitorWechat = ''
    form.messageContent = ''
    submitResult.value = '留言已提交，等待审核后公开。'
  } catch (error) {
    formError.value = apiErrorMessage(error, '留言提交失败，请稍后重试。')
  } finally {
    submitting.value = false
  }
}

onLoad((options) => {
  reloadForFamilyId(String(options?.familyId || ''))
})

onMounted(() => {
  if (typeof window !== 'undefined') {
    window.addEventListener('hashchange', handleHashChange)
  }
})

onUnmounted(() => {
  if (typeof window !== 'undefined') {
    window.removeEventListener('hashchange', handleHashChange)
  }
})
</script>

<style scoped>
.actions {
  margin-top: 20rpx;
}

.info-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24rpx;
  padding: 14rpx 0;
  border-top: 1rpx solid #e5e0d6;
  font-size: 26rpx;
}

.info-block {
  padding: 14rpx 0;
  border-top: 1rpx solid #e5e0d6;
}

.label {
  flex-shrink: 0;
  color: #6b7280;
  font-size: 24rpx;
}

.info-row text:last-child {
  text-align: right;
}

.info-block .label {
  display: block;
  margin-bottom: 8rpx;
}

.section-desc {
  display: block;
  margin-bottom: 16rpx;
}

.subsection {
  margin-top: 24rpx;
  padding-top: 20rpx;
  border-top: 1rpx solid #e5e0d6;
}

.subsection-title {
  display: block;
  margin-bottom: 12rpx;
  font-size: 28rpx;
  font-weight: 600;
}

.field-hint {
  display: block;
  margin: 8rpx 0 12rpx;
  color: #9ca3af;
  font-size: 22rpx;
}

.message {
  padding: 16rpx 0;
  border-top: 1rpx solid #e5e0d6;
}

.message-empty {
  padding: 16rpx 0;
}

.message-name {
  display: block;
  font-weight: 600;
}

.message-time {
  display: block;
  margin-top: 8rpx;
  color: #9ca3af;
  font-size: 22rpx;
}

.state-card {
  padding: 20rpx 0;
  text-align: center;
}

.counter {
  display: block;
  margin-top: 8rpx;
  color: #9ca3af;
  font-size: 22rpx;
  text-align: right;
}

.result {
  display: block;
  margin-top: 18rpx;
  color: #2f6b57;
  font-size: 26rpx;
  font-weight: 600;
}

.error {
  display: block;
  margin-top: 18rpx;
  color: #c0392b;
  font-size: 24rpx;
}

.button[disabled] {
  opacity: 0.55;
}
</style>
