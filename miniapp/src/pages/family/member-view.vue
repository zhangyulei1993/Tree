<template>
  <view class="archive-page member-view-page">
    <MiniBackHome :return-url="returnUrl" :return-label="returnLabel" />

    <view v-if="!authChecked || loading" class="archive-form-panel view-state">
      <MiniEmptyState symbol="…" title="正在读取资料" description="请稍候..." />
    </view>
    <view v-else-if="loadError" class="archive-form-panel view-state">
      <MiniNotice tone="warm" title="加载失败">{{ loadError }}</MiniNotice>
      <MiniButton variant="secondary" @click="loadMember">重新加载</MiniButton>
    </view>

    <template v-else-if="member">
      <view class="archive-page-head">
        <view>
          <text class="archive-kicker">Member Record</text>
          <text class="archive-title">成员资料</text>
          <text class="archive-subtitle">查看档案、核实状态与维护入口</text>
        </view>
        <view class="archive-surname-stamp">{{ surnameLetter }}</view>
      </view>

      <view class="view-context">
        <text class="archive-chip">{{ memberStatusText(member.status) }}</text>
        <text class="archive-chip">{{ bindingPolicyText(member.userBindingPolicy) }}</text>
        <text v-if="member.boundUserId" class="archive-chip">已绑定账号</text>
      </view>

      <view class="archive-form-panel">
        <view class="archive-section-head">
          <text class="archive-section-title">档案信息</text>
          <text class="archive-section-subtitle">成员基本资料</text>
        </view>
        <view class="detail-list">
          <view class="detail-row"><text>姓名</text><text>{{ member.name || '未填写' }}</text></view>
          <view class="detail-row"><text>性别</text><text>{{ genderText(member.gender) }}</text></view>
          <view class="detail-row"><text>出生年份</text><text>{{ member.birthYear ? `${member.birthYear} 年` : '待补充' }}</text></view>
          <view v-if="member.isAlive === false" class="detail-row"><text>离世年份</text><text>{{ member.deathYear ? `${member.deathYear} 年` : '未记录' }}</text></view>
          <view class="detail-row"><text>当前状态</text><text>{{ member.isAlive === false ? '已离世' : '健在' }}</text></view>
        </view>
        <view v-if="member.description" class="description-block">
          <text class="description-label">成员简介</text>
          <text class="description-value">{{ member.description }}</text>
        </view>
      </view>

      <view v-if="canReviewOtherMember" class="archive-form-panel verification-panel">
        <view class="archive-section-head">
          <text class="archive-section-title">资料核实</text>
          <text class="archive-section-subtitle">{{ verificationDescription }}</text>
        </view>
        <text class="verification-note">核实操作只更新资料状态，不会修改当前档案内容。</text>
        <view class="verification-actions">
          <MiniButton size="sm" :loading="confirming" :disabled="confirming" @click="confirmProfile('FAMILY_CONFIRMED')">确认资料无误</MiniButton>
          <MiniButton size="sm" variant="secondary" :loading="confirming" :disabled="confirming" @click="confirmProfile('NEEDS_REVIEW')">标记待完善</MiniButton>
        </view>
      </view>

      <view v-if="canEditMember" class="member-view-actions"><MiniButton @click="openEditor">编辑资料</MiniButton></view>
      <text v-if="actionError" class="view-error">{{ actionError }}</text>
    </template>
  </view>
</template>

<script setup lang="ts">
import { onLoad, onShow } from '@dcloudio/uni-app'
import { computed, ref } from 'vue'
import { apiErrorMessage } from '@/api/client'
import { getFamilyDetail } from '@/api/families'
import { confirmFamilyMemberProfile, getFamilyMember } from '@/api/members'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
import { useSessionStore } from '@/stores/session'
import type { FamilyDetail, FamilyMember, Gender } from '@/types/api'

const session = useSessionStore()
const familyId = ref('')
const memberId = ref('')
const returnTo = ref('')
const family = ref<FamilyDetail | null>(null)
const member = ref<FamilyMember | null>(null)
const authChecked = ref(false)
const loading = ref(false)
const confirming = ref(false)
const loadError = ref('')
const actionError = ref('')
const familyDetailUrl = computed(() => `/pages/family/detail?familyId=${encodeURIComponent(familyId.value)}`)
const familyTreeUrl = computed(() => `/pages/family/private-tree?familyId=${encodeURIComponent(familyId.value)}&memberId=${encodeURIComponent(memberId.value)}`)
const familyHealthUrl = computed(() => `/pages/family/health?familyId=${encodeURIComponent(familyId.value)}`)
const returnUrl = computed(() => returnTo.value === 'tree' ? familyTreeUrl.value : returnTo.value === 'health' ? familyHealthUrl.value : familyDetailUrl.value)
const returnLabel = computed(() => returnTo.value === 'tree' ? '返回家庭树' : returnTo.value === 'health' ? '返回家庭健康' : '返回家庭详情')
const canManageFamily = computed(() => family.value?.role === 'FOUNDER' || family.value?.role === 'FAMILY_ADMIN')
const isCurrentUserMember = computed(() => Boolean(member.value?.boundUserId && member.value.boundUserId === session.user?.id))
const canReviewOtherMember = computed(() => canManageFamily.value && !isCurrentUserMember.value)
const canEditMember = computed(() => canManageFamily.value || isCurrentUserMember.value)
const surnameLetter = computed(() => member.value?.name?.trim().slice(0, 1) || family.value?.familySurname.slice(0, 1) || '员')
const verificationDescription = computed(() => {
  const status = member.value?.profileVerificationStatus
  if (status === 'SELF_CONFIRMED') return '本人已确认'
  if (status === 'FAMILY_CONFIRMED') return '家庭已确认'
  if (status === 'NEEDS_REVIEW') return '待完善'
  return '资料待核实'
})

function routePath() {
  const query = `familyId=${encodeURIComponent(familyId.value)}&memberId=${encodeURIComponent(memberId.value)}`
  return `/pages/family/member-view?${query}${['tree', 'health'].includes(returnTo.value) ? `&returnTo=${encodeURIComponent(returnTo.value)}` : ''}`
}
function genderText(gender?: Gender) { return gender === 'MALE' ? '男' : gender === 'FEMALE' ? '女' : '待补充' }
function memberStatusText(status?: string) { return ({ ACTIVE: '活跃', DISABLED: '已停用', DELETED: '已删除' } as Record<string, string>)[status || ''] || '未知状态' }
function bindingPolicyText(policy?: string) { return ({ OPTIONAL: '可选绑定', REQUIRED: '需绑定', NOT_REQUIRED: '无需绑定' } as Record<string, string>)[policy || ''] || '未知策略' }

async function loadMember() {
  if (!familyId.value || !memberId.value) { loadError.value = '缺少成员信息。'; return }
  loading.value = true; loadError.value = ''
  try {
    const [detail, currentMember] = await Promise.all([getFamilyDetail(familyId.value), getFamilyMember(familyId.value, memberId.value)])
    family.value = detail; member.value = currentMember
  } catch (error) { loadError.value = apiErrorMessage(error, '成员资料加载失败。') } finally { loading.value = false }
}

async function confirmProfile(status: 'SELF_CONFIRMED' | 'FAMILY_CONFIRMED' | 'NEEDS_REVIEW') {
  if (!member.value || confirming.value) return
  confirming.value = true; actionError.value = ''
  try {
    member.value = await confirmFamilyMemberProfile(familyId.value, memberId.value, { status, expectedRevision: member.value.revision })
    uni.showToast({ title: status === 'NEEDS_REVIEW' ? '已标记待完善' : '资料已确认', icon: 'success' })
  } catch (error) { actionError.value = apiErrorMessage(error, '资料核实失败。') } finally { confirming.value = false }
}

function openEditor() {
  const query = `familyId=${encodeURIComponent(familyId.value)}&memberId=${encodeURIComponent(memberId.value)}`
  const returnSuffix = ['tree', 'health'].includes(returnTo.value) ? `&returnTo=${encodeURIComponent(returnTo.value)}` : ''
  uni.navigateTo({ url: `/pages/family/member-edit?${query}${returnSuffix}` })
}
onLoad((options) => {
  familyId.value = String(options?.familyId || '')
  memberId.value = String(options?.memberId || '')
  returnTo.value = String(options?.returnTo || '')
})
onShow(() => { authChecked.value = session.requireLogin(routePath()); if (authChecked.value) void loadMember() })
</script>

<style scoped>
.member-view-page{padding-top:28rpx}.member-view-page :deep(.mini-back-home){margin-bottom:22rpx}.view-state :deep(.mini-notice){margin-bottom:18rpx}.view-context{display:flex;flex-wrap:wrap;gap:10rpx;margin-bottom:24rpx}.detail-list{margin-top:12rpx;border-top:1rpx solid var(--archive-line)}.detail-row{display:flex;justify-content:space-between;gap:24rpx;padding:18rpx 0;border-bottom:1rpx solid var(--archive-line);color:var(--archive-ink-soft);font-size:26rpx}.detail-row text:last-child{color:var(--archive-ink);font-weight:700;text-align:right}.description-block{margin-top:22rpx}.description-label,.description-value{display:block}.description-label{color:var(--archive-ink-soft);font-size:24rpx}.description-value{margin-top:8rpx;color:var(--archive-ink);font-size:27rpx;line-height:1.7;white-space:pre-wrap}.verification-panel{margin-top:22rpx}.verification-note{display:block;margin-top:12rpx;color:var(--archive-ink-soft);font-size:24rpx;line-height:1.6}.verification-actions{display:flex;flex-wrap:wrap;gap:14rpx;margin-top:18rpx}.member-view-actions{margin-top:24rpx}.view-error{display:block;margin-top:16rpx;color:var(--archive-cinnabar);font-size:24rpx}
</style>
