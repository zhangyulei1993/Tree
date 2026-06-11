<template>
  <view class="page">
    <view class="card">
      <text class="title">家庭成员</text>
      <text class="muted">查看家庭中的成员、角色与绑定状态。</text>
      <text v-if="errorMessage" class="error">{{ errorMessage }}</text>
      <button v-if="errorMessage" class="button secondary" @click="loadMembers">重新加载</button>
    </view>
    <view v-if="loading" class="card state-card">
      <text class="muted">正在加载成员...</text>
    </view>
    <view v-else-if="!errorMessage && members.length === 0" class="card state-card">
      <text class="section-title">暂无成员</text>
      <text class="muted">当前家庭还没有可展示的成员。</text>
    </view>
    <view v-else-if="!errorMessage" class="summary-card">
      <text class="muted">共 {{ members.length }} 位成员</text>
    </view>
    <view v-for="member in members" :key="member.memberId" class="card">
      <view class="section-row">
        <text class="section-title">{{ member.name }}</text>
        <text class="tag">{{ roleText(member.boundFamilyRole) }}</text>
      </view>
      <view class="meta-row">
        <text class="tag">{{ genderText(member.gender) }}</text>
        <text class="tag">{{ memberStatusText(member.status) }}</text>
      </view>
      <view class="info-row">
        <text class="label">账号绑定</text>
        <text>{{ bindStatusText(member) }}</text>
      </view>
      <view class="info-row">
        <text class="label">绑定要求</text>
        <text>{{ bindingPolicyText(member.userBindingPolicy) }}</text>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { onLoad } from '@dcloudio/uni-app'
import { ref } from 'vue'

import { apiErrorMessage } from '@/api/client'
import { listFamilyMembers } from '@/api/members'
import { useSessionStore } from '@/stores/session'
import type { FamilyMember } from '@/types/api'

const session = useSessionStore()
const familyId = ref('')
const members = ref<FamilyMember[]>([])
const loading = ref(false)
const errorMessage = ref('')

async function loadMembers() {
  if (!familyId.value) {
    errorMessage.value = '缺少家庭信息。'
    return
  }
  const route = `/pages/family/members?familyId=${encodeURIComponent(familyId.value)}`
  if (!session.requireLogin(route)) return
  loading.value = true
  errorMessage.value = ''
  try {
    members.value = await listFamilyMembers(familyId.value)
  } catch (error) {
    errorMessage.value = apiErrorMessage(error, '成员列表加载失败。')
  } finally {
    loading.value = false
  }
}

function genderText(gender: string) {
  switch (gender) {
    case 'MALE':
      return '男'
    case 'FEMALE':
      return '女'
    default:
      return '未知性别'
  }
}

function memberStatusText(status: string) {
  switch (status) {
    case 'ACTIVE':
      return '正常'
    case 'DELETED':
      return '已删除'
    case 'DISABLED':
      return '已停用'
    case 'PENDING':
      return '待审核'
    case 'APPROVED':
      return '已通过'
    case 'REJECTED':
      return '已拒绝'
    default:
      return '未知状态'
  }
}

function roleText(role?: string | null) {
  switch (role) {
    case 'FOUNDER':
      return '创建者'
    case 'FAMILY_ADMIN':
      return '管理员'
    case 'MEMBER':
      return '成员'
    default:
      return role ? '未知角色' : '未绑定角色'
  }
}

function bindStatusText(member: FamilyMember) {
  if (member.userBindingPolicy === 'NOT_REQUIRED') return '无需绑定'
  return member.boundUserId ? '已绑定' : '未绑定'
}

function bindingPolicyText(policy: string) {
  switch (policy) {
    case 'REQUIRED':
      return '需要绑定'
    case 'OPTIONAL':
      return '可选择绑定'
    case 'NOT_REQUIRED':
      return '无需绑定'
    default:
      return '未设置'
  }
}

onLoad((options) => {
  familyId.value = String(options?.familyId || '')
  loadMembers()
})
</script>

<style scoped>
.state-card {
  text-align: center;
}

.error {
  display: block;
  margin-top: 16rpx;
  color: #c0392b;
  font-size: 24rpx;
}

.section-row,
.info-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16rpx;
}

.meta-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8rpx;
  margin-bottom: 8rpx;
}

.section-row .section-title {
  margin-bottom: 8rpx;
}

.info-row {
  padding: 12rpx 0;
  border-top: 1rpx solid #e5e0d6;
  font-size: 26rpx;
}

.label {
  color: #6b7280;
}

.summary-card {
  margin: 0 0 16rpx;
  padding: 0 8rpx;
}

</style>
