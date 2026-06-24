<template>
  <view class="tree-page">
    <MiniBackHome />
    <view class="tree-tool-banner members-banner">
      <view class="banner-copy">
        <text class="tree-tool-banner-title">家庭成员</text>
        <text class="tree-tool-banner-desc">查看成员基本信息与账号绑定</text>
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

    <MiniCard v-if="errorMessage">
      <MiniNotice tone="warm" title="加载失败">{{ errorMessage }}</MiniNotice>
      <MiniButton variant="secondary" @click="loadMembers">重新加载</MiniButton>
    </MiniCard>

    <template v-else>
      <MiniCard v-if="loading">
        <MiniEmptyState symbol="…" title="正在加载" description="正在加载成员..." />
      </MiniCard>

      <MiniCard v-else-if="members.length === 0">
        <MiniEmptyState
          symbol="员"
          title="暂无成员"
          description="当前家庭还没有可展示的成员。"
        />
      </MiniCard>

      <view v-else class="tree-space">
        <view class="tree-space-head">
          <text class="tree-space-title">成员名册</text>
          <text class="tree-space-subtitle">共 {{ members.length }} 位成员</text>
        </view>
        <view class="tree-space-body section-pad">
          <MemberMiniCard
            v-for="member in members"
            :key="member.memberId"
            :name="member.name"
            :age-text="formatMemberAge(member)"
            :gender-text="formatMemberGender(member.gender)"
            :living-text="formatMemberLiving(member.isAlive)"
            :binding-text="formatMemberBindingNeed(member)"
            :show-binding="true"
            :status-label="member.status !== 'ACTIVE' ? memberStatusText(member.status) : undefined"
            :status-tone="memberStatusTone(member.status)"
            :role-label="roleText(member.boundFamilyRole)"
          />
        </view>
      </view>
    </template>
  </view>
</template>

<script setup lang="ts">
import { onLoad } from '@dcloudio/uni-app'
import { ref } from 'vue'

import { apiErrorMessage } from '@/api/client'
import { listFamilyMembers } from '@/api/members'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniCard from '@/components/base/MiniCard.vue'
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
import { statusTagTone } from '@/components/base/formatStatus'
import MemberMiniCard from '@/components/family/MemberMiniCard.vue'
import { useSessionStore } from '@/stores/session'
import type { FamilyMember } from '@/types/api'
import {
  formatMemberAge,
  formatMemberBindingNeed,
  formatMemberGender,
  formatMemberLiving
} from '@/utils/memberFormat'

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

function memberStatusTone(status: string) {
  return statusTagTone(status)
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
      return role ? '未知角色' : ''
  }
}

onLoad((options) => {
  familyId.value = String(options?.familyId || '')
  loadMembers()
})
</script>

<style scoped>
.members-banner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16rpx;
  margin-bottom: 20rpx;
}

.banner-copy {
  flex: 1;
  min-width: 0;
}

.section-pad {
  padding: 8rpx 24rpx 16rpx;
}

.tree-page :deep(.mini-card.soft) {
  margin-bottom: 16rpx;
}
</style>
