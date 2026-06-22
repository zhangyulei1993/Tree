<template>
  <view class="tree-page">
    <MiniCard v-if="loadingFamily">
      <MiniEmptyState symbol="…" title="正在加载" description="正在加载公开家庭信息..." />
    </MiniCard>

    <MiniCard v-else-if="familyError">
      <MiniSectionHeader title="公开家庭" />
      <MiniNotice tone="warm" title="加载失败">{{ familyError }}</MiniNotice>
      <MiniButton variant="secondary" @click="loadFamily">重新加载</MiniButton>
    </MiniCard>

    <template v-else-if="family">
      <MiniCard variant="hero" class="public-hero tree-pedigree-watermark">
        <text class="tree-public-eyebrow">公开家族主页</text>
        <view class="tree-dossier-head">
          <view class="tree-dossier-seal">{{ family.familySurname.slice(0, 1) }}</view>
          <view class="tree-dossier-copy">
            <text class="public-surname">{{ family.familySurname }}氏</text>
            <text class="tree-page-title">{{ family.familyName }}</text>
          </view>
        </view>
        <text class="public-desc">{{ family.description || '该家庭暂未填写公开简介。' }}</text>
        <view class="tree-archive-ribbon">
          <text v-if="family.nativePlace" class="tree-archive-chip">{{ family.nativePlace }}</text>
          <text v-if="family.regionText" class="tree-archive-chip green">{{ family.regionText }}</text>
        </view>
        <view class="hero-actions">
          <MiniButton @click="openPublicTree">查看公开家谱</MiniButton>
          <MiniButton variant="secondary" @click="openJoinApply">申请加入家庭</MiniButton>
        </view>
      </MiniCard>

      <view v-if="family.publicContactVisible && hasPublicContact" class="tree-space contact-space">
        <view class="tree-space-head">
          <text class="tree-space-title">联系方式</text>
          <text class="tree-space-subtitle">家庭选择对外公开的联络方式</text>
        </view>
        <view class="tree-space-body tree-contact-card">
          <view v-if="family.publicContactName" class="tree-info-row">
            <text class="tree-info-label">联系人</text>
            <text class="tree-info-value">{{ family.publicContactName }}</text>
          </view>
          <view v-if="family.publicContactPhone" class="tree-info-row">
            <text class="tree-info-label">联系电话</text>
            <text class="tree-info-value">{{ family.publicContactPhone }}</text>
          </view>
          <view v-if="family.publicContactWechat" class="tree-info-row">
            <text class="tree-info-label">微信</text>
            <text class="tree-info-value">{{ family.publicContactWechat }}</text>
          </view>
          <view v-if="family.publicContactNote" class="info-block">
            <text class="tree-field-label">备注</text>
            <text class="tree-muted">{{ family.publicContactNote }}</text>
          </view>
        </view>
      </view>

      <view class="tree-space message-space">
        <view class="tree-space-head">
          <text class="tree-space-title">公开留言</text>
          <text class="tree-space-subtitle">审核通过的留言会显示在这里</text>
        </view>
        <view class="tree-space-body section-pad">
        <view class="subsection">
          <text class="subsection-title">留言列表</text>
          <MiniCard v-if="loadingMessages" variant="soft" :no-margin="true">
            <MiniEmptyState symbol="…" title="正在加载" description="正在加载留言..." />
          </MiniCard>
          <MiniCard v-else-if="messagesError" variant="soft" :no-margin="true">
            <MiniNotice tone="warm" title="加载失败">{{ messagesError }}</MiniNotice>
            <MiniButton variant="secondary" @click="loadMessages">重新加载留言</MiniButton>
          </MiniCard>
          <MiniEmptyState
            v-else-if="messages.length === 0"
            symbol="言"
            title="暂无公开留言"
            description="审核通过的留言会显示在这里。"
          />
          <view v-for="message in messages" v-else :key="message.messageId" class="tree-message-wall-item">
            <text class="tree-message-wall-name">{{ message.visitorName || '匿名访客' }}</text>
            <text class="tree-muted">{{ message.messageContent }}</text>
            <text class="tree-message-wall-time">{{ formatDate(message.reviewedAt || message.createdAt) }}</text>
          </view>
        </view>

        <view class="subsection">
          <text class="subsection-title">我要留言</text>
          <MiniNotice tone="security">
            留言内容会经家庭管理员审核后公开显示。联系电话和微信号不会公开展示，仅用于联系确认。
          </MiniNotice>
          <text class="tree-field-label">你的称呼（可选）</text>
          <input v-model.trim="form.visitorName" class="tree-input" maxlength="40" placeholder="你的称呼（可选）" />
          <text class="tree-field-label">联系电话（可选）</text>
          <input v-model.trim="form.visitorPhone" class="tree-input" maxlength="30" placeholder="联系电话（可选）" />
          <text class="tree-field-label">微信号（可选）</text>
          <input v-model.trim="form.visitorWechat" class="tree-input" maxlength="100" placeholder="微信号（可选）" />
          <text class="tree-field-label">留言内容（必填）</text>
          <textarea
            v-model.trim="form.messageContent"
            class="tree-textarea"
            maxlength="500"
            placeholder="请输入留言内容（必填）"
          />
          <text class="counter">{{ form.messageContent.length }}/500</text>
          <MiniButton variant="secondary" :disabled="submitting" :loading="submitting" @click="submitMessage">
            提交留言
          </MiniButton>
          <text v-if="formError" class="tree-field-error">{{ formError }}</text>
          <text v-if="submitResult" class="tree-field-success">{{ submitResult }}</text>
        </view>
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
import MiniButton from '@/components/base/MiniButton.vue'
import MiniCard from '@/components/base/MiniCard.vue'
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
import MiniSectionHeader from '@/components/base/MiniSectionHeader.vue'
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
.section-pad {
  padding: 8rpx 24rpx 16rpx;
}

.message-space {
  margin-bottom: 20rpx;
}

.public-surname {
  display: block;
  margin-bottom: 8rpx;
  color: var(--tree-gold-text);
  font-size: 24rpx;
  font-weight: 600;
  letter-spacing: 2rpx;
}

.public-desc {
  display: block;
  margin-top: 12rpx;
  color: var(--tree-text-secondary);
  font-size: 26rpx;
  line-height: 1.7;
}

.hero-actions {
  margin-top: 22rpx;
}

.info-block {
  padding: 14rpx 0;
  border-top: 1rpx solid var(--tree-border-subtle);
}

.subsection {
  margin-top: 22rpx;
  padding-top: 18rpx;
  border-top: 1rpx solid var(--tree-border-subtle);
}

.subsection-title {
  display: block;
  margin-bottom: 12rpx;
  color: var(--tree-text-primary);
  font-size: 28rpx;
  font-weight: 600;
}

.message {
  padding: 16rpx 0;
  border-top: 1rpx solid var(--tree-border-subtle);
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

.counter {
  display: block;
  margin-top: 8rpx;
  color: #9ca3af;
  font-size: 22rpx;
  text-align: right;
}
</style>
