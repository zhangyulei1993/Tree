<template>
  <view class="archive-page health-page">
    <MiniBackHome :return-url="returnUrl" :return-label="returnLabel" />
    <view v-if="loading" class="archive-form-panel"><MiniEmptyState symbol="…" title="正在检查" description="正在核对成员资料和家庭关系…" /></view>
    <view v-else-if="errorMessage" class="archive-form-panel"><MiniNotice tone="warm" title="检查失败">{{ errorMessage }}</MiniNotice><MiniButton variant="secondary" @click="load">重新检查</MiniButton></view>
    <template v-else-if="family">
      <view v-if="errorMessage" class="health-refresh-error">
        <MiniNotice tone="warm" title="刷新未完成">{{ errorMessage }}</MiniNotice>
        <MiniButton size="sm" variant="secondary" :disabled="refreshing" @click="load">重新检查</MiniButton>
      </view>
      <view class="archive-page-head">
        <view><text class="archive-kicker">Family Health</text><text class="archive-title">家庭健康</text><text class="archive-subtitle">优先处理影响建树、称谓与邀请的事项</text></view>
        <view class="archive-seal">检</view>
      </view>
      <view class="health-score archive-form-panel">
        <view><text class="score-label">基础建树完整度</text><text class="score-note">{{ scoreNote }}</text></view>
        <text class="score-number">{{ score }}</text>
      </view>
      <view class="health-export archive-form-panel">
        <view class="archive-section-head">
          <text class="archive-section-title">家庭档案导出</text>
          <text class="archive-section-subtitle">生成前需达到 {{ exportScoreThreshold }} 分，确保长卷中的成员与关系足够可信。</text>
        </view>
        <MiniNotice v-if="canExport" tone="info" title="可以生成档案长卷">当前完整度 {{ score }} 分，可选择家庭留存、分享亲友或公开展示。</MiniNotice>
        <MiniNotice v-else tone="warm" title="暂不建议导出">当前 {{ score }} 分，还差 {{ exportScoreThreshold - score }} 分。请先处理上方待补充事项。</MiniNotice>
        <MiniButton :disabled="!canExport" @click="openExport">{{ canExport ? '生成家庭档案长卷' : '完成健康检查后可导出' }}</MiniButton>
      </view>
      <MiniNotice v-if="issues.length" tone="warm" title="建议先处理高优先级事项">补全后，称谓、邀请与统计更准确。</MiniNotice>
      <view v-if="issues.length" class="archive-list health-list">
        <view v-for="issue in issues" :key="issue.key" class="archive-row" @tap="openIssue(issue)">
          <view class="archive-row-main"><text class="archive-row-title">{{ issue.title }}</text><text class="archive-row-desc">{{ issue.description }}</text><view v-if="issue.members?.length" class="issue-members"><text v-for="member in issue.members" :key="member.memberId" class="issue-member" @tap.stop="openMemberIssue(issue, member)">{{ member.name }}</text></view></view>
          <text class="health-count" :class="`health-count--${issue.tone}`">{{ issue.count }}</text><text class="archive-arrow">›</text>
        </view>
      </view>
      <view v-else class="health-clear-state">
        <view class="health-clear-mark"><text>已</text></view>
        <view class="health-clear-copy">
          <text class="health-clear-title">基础检查已通过</text>
          <text class="health-clear-desc">当前没有需要优先补充的成员资料或关系问题。</text>
        </view>
      </view>
    </template>
  </view>
</template>
<script setup lang="ts">
import { computed, ref } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import { apiErrorMessage } from '@/api/client'
import { getFamilyDetail } from '@/api/families'
import { listFamilyMembers } from '@/api/members'
import { getPrivateTree } from '@/api/tree'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
import { calculateFamilyHealthScore, summarizeFamilyHealth } from '@/features/family-tree/familyHealthSummary'
import { inspectRelationshipHealth } from '@/features/family-tree/relationshipHealth'
import { useSessionStore } from '@/stores/session'
import type { FamilyDetail, FamilyMember, FamilyTreeResult } from '@/types/api'

type IssueMember = { memberId: number; name: string }
type Issue = { key: string; title: string; description: string; count: number; tone: 'risk' | 'attention' | 'neutral'; route: string; members?: IssueMember[] }
const session = useSessionStore(); const familyId = ref(''); const family = ref<FamilyDetail | null>(null); const members = ref<FamilyMember[]>([]); const tree = ref<FamilyTreeResult | null>(null); const loading = ref(false); const refreshing = ref(false); const navigating = ref(false); const errorMessage = ref('')
const returnTo = ref('')
const returnUrl = computed(() => returnTo.value === 'tree'
  ? `/pages/family/private-tree?familyId=${encodeURIComponent(familyId.value)}`
  : `/pages/family/detail?familyId=${encodeURIComponent(familyId.value)}`)
const returnLabel = computed(() => returnTo.value === 'tree' ? '返回家庭树' : '返回家庭详情')
const exportScoreThreshold = 90
const healthSummary = computed(() => summarizeFamilyHealth(family.value, members.value, tree.value))
const issues = computed<Issue[]>(() => {
  const treeUrl = `/pages/family/private-tree?familyId=${encodeURIComponent(familyId.value)}&returnTo=health`
  const list: Issue[] = []
  const unknownGender = members.value.filter((item) => item.gender !== 'MALE' && item.gender !== 'FEMALE')
  const missingBirth = members.value.filter((item) => !item.birthDate && !item.birthYear)
  const requiredUnbound = members.value.filter((item) => item.userBindingPolicy === 'REQUIRED' && !item.boundUserId)
  const relationshipHealth = inspectRelationshipHealth(tree.value)
  const missingSpouse = relationshipHealth.membersWithChildrenNoSpouse
  const birthOrderConflict = relationshipHealth.membersWithBirthOrderConflict
  const connected = new Set((tree.value?.edges || []).flatMap((edge) => [edge.fromMemberId, edge.toMemberId]))
  const isolated = members.value.filter((item) => members.value.length > 1 && !connected.has(item.memberId))
  if (healthSummary.value.unknownGender) list.push({ key:'gender', title:'确认成员性别', description:'用于称谓推导与头像展示。', count:healthSummary.value.unknownGender, tone:'risk', route:treeUrl, members: toIssueMembers(unknownGender) })
  if (healthSummary.value.isolated) list.push({ key:'relation', title:'连接孤立成员', description:'未纳入关系，无法推导称谓。', count:healthSummary.value.isolated, tone:'risk', route:treeUrl, members: toIssueMembers(isolated) })
  if (healthSummary.value.missingBirth) list.push({ key:'birth', title:'补充出生年份', description:'用于区分同辈顺序。', count:healthSummary.value.missingBirth, tone:'attention', route:treeUrl, members: toIssueMembers(missingBirth) })
  if (healthSummary.value.birthOrderConflict) list.push({ key:'birth-order', title:'核对亲子出生年份', description:'父母出生年份不早于子女，请核对资料或关系。', count:healthSummary.value.birthOrderConflict, tone:'risk', route:treeUrl, members: birthOrderConflict.map((item) => ({ memberId:item.memberId, name:item.displayName || `成员 #${item.memberId}` })) })
  if (healthSummary.value.missingSpouse) list.push({ key:'spouse', title:'补充配偶关系', description:'已有子女关系但未记录配偶；如信息已知，建议补充。', count:healthSummary.value.missingSpouse, tone:'attention', route:treeUrl, members: missingSpouse.map((item) => ({ memberId:item.memberId, name:item.displayName || `成员 #${item.memberId}` })) })
  if (healthSummary.value.requiredUnbound) list.push({ key:'invite', title:'邀请待绑定成员', description:'在对应成员节点中发起或继续邀请。', count:healthSummary.value.requiredUnbound, tone:'attention', route:treeUrl, members: toIssueMembers(requiredUnbound) })
  if (healthSummary.value.missingPlace) list.push({ key:'place', title:'补充家庭所在地', description:'完善家庭资料。', count:healthSummary.value.missingPlace, tone:'neutral', route:`/pages/family/profile?familyId=${encodeURIComponent(familyId.value)}` })
  return list
})
const score = computed(() => calculateFamilyHealthScore(healthSummary.value))
const scoreNote = computed(() => issues.value.length ? `${issues.value.length} 类事项待处理` : '已完成基础建树检查')
const canExport = computed(() => score.value >= exportScoreThreshold)
async function load() { if (loading.value || refreshing.value) return; session.restoreSession(); if (!familyId.value) { errorMessage.value='缺少家庭信息。'; return }; if (!session.isLoggedIn) { resetPageData(); session.requireLogin(`/pages/family/health?familyId=${encodeURIComponent(familyId.value)}`); return }; loading.value=!family.value; refreshing.value=true; errorMessage.value=''; try { const [detail, memberRows, treeResult] = await Promise.all([getFamilyDetail(familyId.value), listFamilyMembers(familyId.value), getPrivateTree(familyId.value)]); family.value=detail; members.value=memberRows; tree.value=treeResult } catch (error) { errorMessage.value=apiErrorMessage(error, '家庭健康检查失败。') } finally { loading.value=false; refreshing.value=false } }
function resetPageData() { loading.value=false; refreshing.value=false; navigating.value=false; errorMessage.value=''; family.value=null; members.value=[]; tree.value=null }
function navigateOnce(url: string) { if (navigating.value) return; navigating.value=true; uni.navigateTo({ url, complete: () => { navigating.value=false } }) }
function openIssue(issue: Issue) { navigateOnce(issue.route) }
function openExport() {
  if (!canExport.value) return
  navigateOnce(`/pages/family/export?familyId=${encodeURIComponent(familyId.value)}&returnTo=${returnTo.value === 'tree' ? 'tree' : 'health'}`)
}
function openMemberIssue(issue: Issue, member: IssueMember) {
  const memberId = encodeURIComponent(String(member.memberId))
  const route = ['gender', 'birth', 'birth-order'].includes(issue.key)
    ? `/pages/family/member-view?familyId=${encodeURIComponent(familyId.value)}&memberId=${memberId}&returnTo=health`
    : issue.key === 'invite'
      ? `/pages/invite/sent?familyId=${encodeURIComponent(familyId.value)}&mode=node&memberId=${memberId}&memberName=${encodeURIComponent(member.name)}&returnTo=health`
      : `/pages/family/private-tree?familyId=${encodeURIComponent(familyId.value)}&memberId=${memberId}&returnTo=health${issue.key === 'spouse' ? '&task=ADD_SPOUSE' : ''}`
  navigateOnce(route)
}
function toIssueMembers(rows: FamilyMember[]): IssueMember[] { return rows.map((item) => ({ memberId: item.memberId, name: item.name || `成员 #${item.memberId}` })) }
onLoad((options) => { familyId.value=String(options?.familyId || ''); returnTo.value=String(options?.returnTo || '') }); onShow(load)
</script>
<style scoped>.health-page{padding-top:28rpx}.health-refresh-error{display:flex;align-items:center;gap:16rpx;margin-bottom:18rpx}.health-refresh-error :deep(.mini-notice){flex:1}.health-score{display:flex;align-items:center;justify-content:space-between;margin-top:22rpx}.score-label,.score-note{display:block}.score-label{color:var(--archive-ink);font-size:28rpx;font-weight:800}.score-note{margin-top:8rpx;color:var(--archive-ink-soft);font-size:23rpx}.score-number{color:var(--archive-cinnabar);font-size:62rpx;font-weight:800;line-height:1}.health-export{margin-top:22rpx}.health-export :deep(.mini-notice){margin:0 0 18rpx}.health-list{margin-top:22rpx}.health-count{min-width:42rpx;font-size:28rpx;font-weight:800;text-align:center}.health-count--risk{color:var(--archive-cinnabar)}.health-count--attention{color:#9a5d14}.health-count--neutral{color:var(--archive-blue)}.issue-members{display:flex;flex-wrap:wrap;gap:10rpx;margin-top:12rpx}.issue-member{padding:5rpx 12rpx;border:1rpx solid rgba(31,58,95,.2);border-radius:6rpx;background:rgba(255,252,244,.78);color:var(--archive-blue);font-size:22rpx;line-height:1.35}.issue-member:active{border-color:var(--archive-cinnabar);color:var(--archive-cinnabar)}.health-clear-state{display:flex;align-items:center;gap:18rpx;margin-top:24rpx;padding:24rpx 4rpx 6rpx}.health-clear-mark{display:flex;flex:0 0 auto;align-items:center;justify-content:center;width:58rpx;height:58rpx;border:1rpx solid rgba(42,111,91,.28);border-radius:50%;background:rgba(42,111,91,.08);color:#2a6f5b;font-size:24rpx;font-weight:800}.health-clear-copy{display:flex;min-width:0;flex:1;flex-direction:column;gap:5rpx}.health-clear-title{color:var(--archive-ink);font-size:28rpx;font-weight:800}.health-clear-desc{color:var(--archive-ink-soft);font-size:24rpx;line-height:1.55}</style>
