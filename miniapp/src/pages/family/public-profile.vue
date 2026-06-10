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
        <text class="tag">姓氏：{{ family.familySurname }}</text>
        <text class="tag">籍贯：{{ family.nativePlace || '未设置' }}</text>
        <text class="tag">{{ family.regionText || '未设置地区' }}</text>
        <text class="muted">{{ family.description || '该家庭暂未填写公开简介。' }}</text>
        <view class="contact">
          <template v-if="family.publicContactVisible && hasPublicContact">
            <text v-if="family.publicContactName" class="muted">联系人：{{ family.publicContactName }}</text>
            <text v-if="family.publicContactPhone" class="muted">电话：{{ family.publicContactPhone }}</text>
            <text v-if="family.publicContactWechat" class="muted">微信：{{ family.publicContactWechat }}</text>
            <text v-if="family.publicContactNote" class="muted">{{ family.publicContactNote }}</text>
          </template>
          <text v-else class="muted">该家庭未设置公开联系方式，如需联系请通过平台协助。</text>
        </view>
        <button class="button" @click="openPublicTree">查看公开树</button>
        <button class="button secondary" @click="openJoinApply">申请加入</button>
      </view>

      <view class="card">
        <text class="section-title">游客留言</text>
        <view v-if="loadingMessages" class="state-card">
          <text class="muted">正在加载留言...</text>
        </view>
        <view v-else-if="messagesError" class="state-card">
          <text class="error">{{ messagesError }}</text>
          <button class="button secondary" @click="loadMessages">重新加载留言</button>
        </view>
        <view v-else-if="messages.length === 0" class="message">
          <text class="muted">暂无已审核公开留言。</text>
        </view>
        <view v-for="message in messages" v-else :key="message.messageId" class="message">
          <text>{{ message.visitorName || '匿名访客' }}</text>
          <text class="muted">{{ message.messageContent }}</text>
          <text class="muted">{{ formatDate(message.reviewedAt || message.createdAt) }}</text>
        </view>

        <input v-model.trim="form.visitorName" class="input" maxlength="40" placeholder="访客称呼（可选）" />
        <input v-model.trim="form.visitorPhone" class="input" maxlength="30" placeholder="联系电话（可选，不公开）" />
        <input v-model.trim="form.visitorWechat" class="input" maxlength="100" placeholder="微信号（可选，不公开）" />
        <textarea
          v-model.trim="form.messageContent"
          class="textarea"
          maxlength="500"
          placeholder="请输入留言内容"
        />
        <text class="counter">{{ form.messageContent.length }}/500</text>
        <button class="button secondary" :disabled="submitting" :loading="submitting" @click="submitMessage">
          提交留言
        </button>
        <text v-if="formError" class="error">{{ formError }}</text>
        <text v-if="submitResult" class="result">{{ submitResult }}</text>
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
    familyError.value = '缺少 familyId。'
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
    familyError.value = '缺少 familyId。'
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
      familyError.value = apiErrorMessage(error, '公开家庭信息不可访问。')
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
      messagesError.value = apiErrorMessage(error, '公开留言加载失败。')
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
  if (!form.messageContent) {
    formError.value = '请填写留言内容。'
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
    formError.value = apiErrorMessage(error, '留言提交失败。')
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
.contact {
  display: block;
  margin-top: 18rpx;
}

.message {
  padding: 16rpx 0;
  border-top: 1rpx solid #e5e0d6;
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
