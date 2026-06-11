<template>
  <view class="page">
    <view v-if="loading" class="card state-card">
      <text class="muted">正在加载家庭详情...</text>
    </view>
    <view v-else-if="errorMessage" class="card state-card">
      <text class="error">{{ errorMessage }}</text>
      <button class="button secondary" @click="loadFamily">重新加载</button>
    </view>
    <template v-else-if="family">
      <view class="card">
        <text class="title">{{ family.familyName }}</text>
        <text class="tag">{{ familyStatusText(family.status) }}</text>
        <text class="tag">{{ publicStatusText(family.publicDisplayStatus) }}</text>
        <text class="muted">{{ family.description || '暂未填写家庭简介。' }}</text>
      </view>

      <view class="card">
        <text class="section-title">基本信息</text>
        <view class="info-row"><text class="label">姓氏</text><text>{{ family.familySurname }}</text></view>
        <view class="info-row"><text class="label">籍贯</text><text>{{ family.nativePlace || '未设置' }}</text></view>
        <view class="info-row"><text class="label">地区</text><text>{{ family.regionText || '未设置' }}</text></view>
      </view>

      <view class="card">
        <text class="section-title">公开展示</text>
        <view class="info-row"><text class="label">公开状态</text><text>{{ publicStatusText(family.publicDisplayStatus) }}</text></view>
        <view class="info-row"><text class="label">允许被搜索</text><text>{{ family.searchable ? '是' : '否' }}</text></view>
        <view class="info-row">
          <text class="label">公开联系方式</text>
          <text>{{ publicContactText }}</text>
        </view>
      </view>

      <view class="card">
        <text class="section-title">我的身份</text>
        <view class="info-row"><text class="label">家庭角色</text><text>{{ roleText(family.role) }}</text></view>
        <view class="info-row"><text class="label">家庭状态</text><text>{{ familyStatusText(family.status) }}</text></view>
      </view>

      <view class="card">
        <text class="section-title">操作入口</text>
        <button class="button" @click="openMembers">查看成员列表</button>
        <button class="button secondary" @click="openTree">查看私有家庭树</button>
      </view>
    </template>
  </view>
</template>

<script setup lang="ts">
import { onLoad } from '@dcloudio/uni-app'
import { computed, ref } from 'vue'

import { apiErrorMessage } from '@/api/client'
import { getFamilyDetail } from '@/api/families'
import { useSessionStore } from '@/stores/session'
import type { FamilyDetail } from '@/types/api'

const session = useSessionStore()
const familyId = ref('')
const family = ref<FamilyDetail | null>(null)
const loading = ref(false)
const errorMessage = ref('')
const publicContactText = computed(() => {
  if (!family.value?.publicContactVisible) return '未公开'
  const contacts = [
    family.value.publicContactName,
    family.value.publicContactPhone,
    family.value.publicContactWechat,
    family.value.publicContactNote
  ].filter(Boolean)
  return contacts.length > 0 ? contacts.join(' / ') : '未设置'
})

async function loadFamily() {
  if (!familyId.value) {
    errorMessage.value = '缺少家庭信息。'
    return
  }
  const route = `/pages/family/detail?familyId=${encodeURIComponent(familyId.value)}`
  if (!session.requireLogin(route)) return
  loading.value = true
  errorMessage.value = ''
  try {
    family.value = await getFamilyDetail(familyId.value)
  } catch (error) {
    errorMessage.value = apiErrorMessage(error, '家庭详情加载失败。')
  } finally {
    loading.value = false
  }
}

function openMembers() {
  uni.navigateTo({
    url: `/pages/family/members?familyId=${encodeURIComponent(familyId.value)}`
  })
}

function openTree() {
  uni.navigateTo({
    url: `/pages/family/private-tree?familyId=${encodeURIComponent(familyId.value)}`
  })
}

function roleText(role: string) {
  switch (role) {
    case 'FOUNDER':
      return '创建者'
    case 'FAMILY_ADMIN':
      return '管理员'
    case 'MEMBER':
      return '成员'
    case 'ROOT_ADMIN':
      return '平台管理员'
    default:
      return role ? '未知角色' : '成员'
  }
}

function familyStatusText(status: string) {
  switch (status) {
    case 'NORMAL':
      return '正常'
    case 'DISABLED':
      return '已停用'
    case 'DISSOLVED':
      return '已解散'
    case 'DISSOLUTION_PENDING':
      return '解散待审核'
    case 'PENDING':
      return '待审核'
    case 'APPROVED':
      return '已通过'
    case 'REJECTED':
      return '已拒绝'
    case 'TAKEN_DOWN':
      return '已下架'
    case 'DELETED':
      return '已删除'
    default:
      return '未知状态'
  }
}

function publicStatusText(status: string) {
  switch (status) {
    case 'APPROVED':
      return '已公开'
    case 'PENDING':
      return '待审核'
    case 'REJECTED':
      return '已拒绝'
    case 'PRIVATE':
      return '未公开'
    case 'TAKEN_DOWN':
      return '已下架'
    default:
      return '未知状态'
  }
}

onLoad((options) => {
  familyId.value = String(options?.familyId || '')
  loadFamily()
})
</script>

<style scoped>
.state-card {
  text-align: center;
}

.error {
  display: block;
  color: #c0392b;
  font-size: 24rpx;
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

.label {
  flex-shrink: 0;
  color: #6b7280;
}

.info-row text:last-child {
  text-align: right;
}
</style>
