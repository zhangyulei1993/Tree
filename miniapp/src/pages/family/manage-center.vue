<template>
  <view class="tree-page manage-center-page">
    <MiniBackHome />
    <FamilyContextHeader
      v-if="family"
      :family-name="family.familyName"
      section="家庭管理"
      subtitle="申请、邀请与公开展示权限"
      :role-label="roleText(family.role)"
      back-label="返回详情"
      @back="openFamilyDetail"
    />

    <MiniFamilyPageSkeleton v-if="loading && !family" variant="manage" />

    <MiniCard v-else-if="errorMessage && !family">
      <MiniNotice tone="warm" title="加载失败">{{ errorMessage }}</MiniNotice>
      <MiniButton variant="secondary" @click="loadFamily">重新加载</MiniButton>
    </MiniCard>

    <MiniCard v-else-if="family && !canManageFamily">
      <MiniNotice tone="warm" title="无管理权限">只有家庭创建者或家庭管理员可以进入家庭管理。</MiniNotice>
      <MiniButton variant="secondary" @click="openFamilyDetail">返回家庭详情</MiniButton>
    </MiniCard>

    <MiniCard v-else-if="family" class="directory-card--list">
      <MiniDirectoryTile
        title="收到的加入申请"
        desc="查看并处理申请加入该家庭的记录"
        icon-class="tree-symbol-users"
        @click="openJoinRequests"
      />
      <MiniDirectoryTile
        title="发出的成员邀请"
        desc="查看、取消或重新生成成员邀请"
        icon-class="tree-symbol-mail"
        @click="openSentInvitations"
      />
      <MiniDirectoryTile
        title="公开展示与权限"
        desc="公开信息、管理员与高风险操作"
        icon-class="tree-symbol-folder"
        @click="openSettings"
      />
    </MiniCard>
  </view>
</template>

<script setup lang="ts">
import { onLoad, onShow, onUnload } from '@dcloudio/uni-app'
import { computed, ref } from 'vue'

import { apiErrorMessage } from '@/api/client'
import { getFamilyDetail } from '@/api/families'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniCard from '@/components/base/MiniCard.vue'
import MiniDirectoryTile from '@/components/base/MiniDirectoryTile.vue'
import MiniFamilyPageSkeleton from '@/components/base/MiniFamilyPageSkeleton.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
import FamilyContextHeader from '@/components/family/FamilyContextHeader.vue'
import { useSessionStore } from '@/stores/session'
import type { FamilyDetail } from '@/types/api'

const session = useSessionStore()
const familyId = ref('')
const family = ref<FamilyDetail | null>(null)
const loading = ref(false)
const errorMessage = ref('')

const canManageFamily = computed(() =>
  family.value?.role === 'FOUNDER' || family.value?.role === 'FAMILY_ADMIN'
)

async function loadFamily() {
  session.restoreSession()
  if (!familyId.value) {
    errorMessage.value = '缺少家庭信息。'
    return
  }
  const route = `/pages/family/manage-center?familyId=${encodeURIComponent(familyId.value)}`
  if (!session.isLoggedIn) {
    resetPageData()
    session.requireLogin(route)
    return
  }
  const isInitialLoad = !family.value
  loading.value = isInitialLoad
  errorMessage.value = ''
  try {
    family.value = await getFamilyDetail(familyId.value)
  } catch (error) {
    errorMessage.value = apiErrorMessage(error, '家庭管理加载失败。')
  } finally {
    loading.value = false
  }
}

function resetPageData() {
  loading.value = false
  errorMessage.value = ''
  family.value = null
}

function openFamilyDetail() {
  if (getCurrentPages().length > 1) {
    uni.navigateBack({ delta: 1 })
    return
  }
  uni.redirectTo({ url: `/pages/family/detail?familyId=${encodeURIComponent(familyId.value)}` })
}

function openJoinRequests() {
  uni.navigateTo({ url: `/pages/join/family?familyId=${encodeURIComponent(familyId.value)}` })
}

function openSentInvitations() {
  uni.navigateTo({ url: `/pages/invite/sent?familyId=${encodeURIComponent(familyId.value)}` })
}

function openSettings() {
  uni.navigateTo({ url: `/pages/family/settings?familyId=${encodeURIComponent(familyId.value)}` })
}

function roleText(role: string) {
  return ({ FOUNDER: '创建者', FAMILY_ADMIN: '家庭管理员', MEMBER: '普通成员' } as Record<string, string>)[role] || '成员'
}

onLoad((options) => {
  familyId.value = String(options?.familyId || '')
})
onShow(loadFamily)
onUnload(resetPageData)
</script>

<style scoped>
.manage-center-page {
  background:
    radial-gradient(circle at 92% 0%, rgba(216, 175, 104, 0.14), transparent 260rpx),
    radial-gradient(circle at 0% 18%, rgba(24, 54, 83, 0.08), transparent 300rpx);
}

.directory-card--list {
  padding-top: 8rpx;
  padding-bottom: 8rpx;
}
</style>
