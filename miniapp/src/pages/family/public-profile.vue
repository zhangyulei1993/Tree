<template>
  <view class="page">
    <view class="card">
      <text class="title">{{ family.name }}</text>
      <text class="tag">姓氏：{{ family.surname }}</text>
      <text class="tag">籍贯：{{ family.nativePlace }}</text>
      <text class="tag">{{ family.regionText }}</text>
      <text class="muted">{{ family.description }}</text>
      <text class="muted contact">{{ family.contact || '该家庭未设置公开联系方式，如需联系请通过平台协助。' }}</text>
      <button class="button" @click="go('/pages/family/public-tree')">查看公开树</button>
      <button class="button secondary" @click="go('/pages/join/apply')">申请加入</button>
    </view>
    <view class="card">
      <text class="section-title">游客留言</text>
      <view v-for="message in localMessages" :key="message.id" class="message">
        <text>{{ message.visitorName }}</text>
        <text class="muted">{{ message.content }}</text>
      </view>
      <input v-model.trim="visitorName" class="input" maxlength="20" placeholder="访客称呼" />
      <textarea
        v-model.trim="messageContent"
        class="textarea"
        maxlength="300"
        placeholder="留言内容，mock 不提交真实后端"
      />
      <text class="counter">{{ messageContent.length }}/300</text>
      <button class="button secondary" @click="submitMessage">提交 mock 留言</button>
      <text v-if="result" class="result">{{ result }}</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'

import { families, messages } from '@/mock/data'

const family = families[0]
const localMessages = ref([...messages])
const visitorName = ref('')
const messageContent = ref('')
const result = ref('')

function go(url: string) {
  uni.navigateTo({ url })
}

function submitMessage() {
  result.value = ''
  if (!visitorName.value) {
    uni.showToast({ title: '请填写访客称呼', icon: 'none' })
    return
  }
  if (!messageContent.value) {
    uni.showToast({ title: '请填写留言内容', icon: 'none' })
    return
  }

  localMessages.value.unshift({
    id: `message_mock_${Date.now()}`,
    visitorName: visitorName.value,
    content: messageContent.value
  })
  visitorName.value = ''
  messageContent.value = ''
  result.value = '留言已提交，正式环境中需审核后公开。'
}
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
</style>
