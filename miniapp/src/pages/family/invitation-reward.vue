<template>
  <view class="archive-page reward-page">
    <MiniBackHome />
    <view class="archive-page-head"><view><text class="archive-kicker">Family Invitation</text><text class="archive-title">{{ referralCode ? '帮忙助力解锁名额' : '邀请家人解锁名额' }}</text><text class="archive-subtitle">{{ referralCode ? '邀请人请你完成一次有效助力，帮助家庭解锁更多成员名额。' : '有效邀请为家庭累计进度；奖励名额独立于购买扩容。' }}</text></view></view>
    <MiniNotice v-if="guideMode && !referralCode" tone="security">作为已绑定成员，也可以生成助力邀请。奖励归家庭所有，不占用你的个人名额。</MiniNotice>
    <view v-if="loading && !summary" class="reward-state archive-form-panel"><MiniEmptyState symbol="…" title="正在加载" description="正在同步家庭助力进度..." /></view>
    <view v-else-if="loadError && !summary" class="reward-state archive-form-panel"><MiniNotice tone="warm" title="加载失败">{{ loadError }}</MiniNotice><MiniButton variant="secondary" size="sm" @click="load">重新加载</MiniButton></view>
    <template v-else>
      <view v-if="summary" class="reward-card"><view class="reward-count"><view><text class="reward-number">{{ summary.validInvites }}</text><text class="reward-caption">有效邀请</text></view><view><text class="reward-number">+{{ summary.rewardedMembers }}</text><text class="reward-caption">已奖励名额</text></view></view><view class="milestone-line"><view class="milestone-track"><view class="milestone-fill" :style="{ width: `${progressPercent}%` }" /><view v-for="tier in summary.tiers" :key="tier.requiredInvites" class="milestone-point" :class="{ reached: summary.validInvites >= tier.requiredInvites }" :style="{ left: `${tier.requiredInvites / maxRequiredInvites * 100}%` }" /></view><view class="milestone-labels"><text v-for="tier in summary.tiers" :key="tier.requiredInvites">{{ tier.requiredInvites }} 人<br />+{{ tier.cumulativeRewardMembers }}</text></view></view><text v-if="summary.nextRequiredInvites" class="reward-tip">还差 {{ summary.nextRequiredInvites - summary.validInvites }} 位有效邀请，可累计再得 {{ summary.nextAddedMembers }} 位名额。</text></view>
      <view v-if="referralCode" class="helper-rules"><text class="helper-rules-title">这次助力怎样才会生效</text><text>1. 使用微信登录即可继续；头像和昵称可跳过。</text><text>2. 绑定手机号后，助力才会计入邀请人的家庭进度。</text><text>3. 同一位新用户只能为一个家庭完成一次有效助力。</text></view>
      <view v-if="!referralCode" class="share-actions"><button v-if="!shareCode" class="share-button" @tap="prepareShare">生成助力邀请</button><button v-else class="share-button" open-type="share">继续邀请家人助力</button></view><view v-else class="share-actions"><button class="share-button" :loading="qualifying" :disabled="qualifying" @tap="startQualify">帮他助力</button></view>
      <MiniNotice tone="info">{{ referralCode ? '助力成功后，奖励只会增加到邀请人的家庭，不会影响你的家庭名额。' : '已加入并绑定账号的家庭成员都可以发起助力邀请；同一新用户仅能为一个家庭计入一次。' }}</MiniNotice>
    </template>
    <MiniPhoneUpgradeDialog v-model:visible="phoneBindVisible" title="绑定手机号后完成助力" description="绑定手机号后，这次助力才会计入邀请人的家庭进度。" action-text="绑定微信手机号并完成助力" @bound="qualifyAfterPhoneBind" />
  </view>
</template>
<script setup lang="ts">
import { onLoad, onShareAppMessage, onShow } from '@dcloudio/uni-app'
import { computed, ref } from 'vue'
import { apiErrorMessage } from '@/api/client'
import { getFamilyDetail } from '@/api/families'
import { createInvitationRewardShare, getInvitationReward, qualifyInvitationReward, type InvitationRewardSummary } from '@/api/invitationReward'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
import MiniPhoneUpgradeDialog from '@/components/base/MiniPhoneUpgradeDialog.vue'
import { useSessionStore } from '@/stores/session'
import { buildInvitationRewardSharePayload, loadShareConfigs } from '@/features/share/wechatShare'
import { markNewUserGuideMilestone } from '@/features/onboarding/newUserGuide'
const session = useSessionStore(); const familyId = ref(''); const referralCode = ref(''); const guideMode = ref(false); const shareCode = ref(''); const familyName = ref('家庭'); const summary = ref<InvitationRewardSummary>(); const qualifying = ref(false); const phoneBindVisible = ref(false); const loading = ref(false); const loadError = ref('')
let skipInitialShow = true
const maxRequiredInvites = computed(() => summary.value?.tiers[summary.value.tiers.length - 1]?.requiredInvites || 1)
const progressPercent = computed(() => Math.min(100, (summary.value?.validInvites || 0) / maxRequiredInvites.value * 100))
async function load() { if (loading.value) return; if (!familyId.value) { loadError.value = '缺少家庭信息。'; return }; loading.value = true; loadError.value = ''; try { summary.value = await getInvitationReward(familyId.value) } catch (error) { loadError.value = apiErrorMessage(error, '进度加载失败。') } finally { loading.value = false } }
async function prepareShare() { try { const result = await createInvitationRewardShare(familyId.value); shareCode.value = result.referralCode; summary.value = result.summary; if (guideMode.value) markNewUserGuideMilestone('INVITE_GUIDE_OPENED') } catch (error) { uni.showToast({ title: apiErrorMessage(error, '暂不能发起邀请'), icon: 'none' }) } }
function currentRewardRoute() { const query = [`familyId=${encodeURIComponent(familyId.value)}`]; if (referralCode.value) query.push(`referralCode=${encodeURIComponent(referralCode.value)}`); return `/pages/family/invitation-reward?${query.join('&')}` }
function startQualify() { if (!referralCode.value || qualifying.value) return; if (!session.requireLogin(currentRewardRoute())) return; if (!session.user?.phoneVerified) { phoneBindVisible.value = true; return }; void qualify() }
function qualifyAfterPhoneBind() { void qualify() }
async function qualify() { if (!referralCode.value || qualifying.value) return; qualifying.value = true; try { summary.value = await qualifyInvitationReward(referralCode.value); uni.showModal({ title: '助力成功', content: '这次有效助力已加入家庭邀请进度。', showCancel: false }) } catch (error) { uni.showModal({ title: '暂未计入', content: apiErrorMessage(error, '请完成微信登录和手机号绑定后重试。'), showCancel: false }) } finally { qualifying.value = false } }
onShareAppMessage(() => shareCode.value
  ? buildInvitationRewardSharePayload({ familyId: familyId.value, referralCode: shareCode.value, familyName: familyName.value })
  : ({ title: 'Tree：记录家庭关系', path: '/pages/home/index' }))
onLoad((options) => { familyId.value = String(options?.familyId || ''); referralCode.value = String(options?.referralCode || ''); guideMode.value = String(options?.guide || '') === '1'; session.restoreSession(); void loadShareConfigs(); if (!session.requireLogin(currentRewardRoute())) return; if (!referralCode.value) void getFamilyDetail(familyId.value).then((family) => { familyName.value = family.familyName || '家庭' }).catch(() => undefined); void load() }); onShow(() => { if (skipInitialShow) { skipInitialShow = false; return }; session.restoreSession(); if (session.isLoggedIn) void load() })
</script>
<style scoped>.reward-page{padding-top:28rpx}.reward-state{margin-top:24rpx}.reward-state :deep(.mini-notice){margin-bottom:18rpx}.reward-card,.helper-rules{margin-top:24rpx;padding:28rpx 24rpx;background:var(--archive-paper-deep);border:1px solid var(--archive-line);border-radius:14rpx}.reward-count{display:grid;grid-template-columns:1fr 1fr;gap:16rpx;padding-bottom:22rpx;border-bottom:1px solid var(--archive-line)}.reward-number,.reward-caption{display:block}.reward-number{color:var(--archive-ink);font-size:48rpx;font-weight:700}.reward-caption,.reward-tip,.helper-rules text{color:var(--archive-ink-soft);font-size:23rpx}.milestone-line{margin-top:34rpx}.milestone-track{position:relative;height:12rpx;background:rgba(31,58,95,.12);border-radius:99rpx}.milestone-fill{height:100%;background:var(--archive-cinnabar);border-radius:inherit}.milestone-point{position:absolute;top:50%;width:24rpx;height:24rpx;box-sizing:border-box;border:4rpx solid var(--archive-paper-deep);border-radius:50%;background:var(--archive-ink-soft);transform:translate(-50%,-50%)}.milestone-point.reached{background:var(--archive-cinnabar)}.milestone-labels{display:flex;justify-content:space-between;gap:8rpx;margin-top:18rpx;color:var(--archive-ink-soft);font-size:21rpx;line-height:1.45;text-align:center}.reward-tip{display:block;margin-top:26rpx;line-height:1.6}.helper-rules{display:grid;gap:10rpx;background:rgba(255,252,244,.7)}.helper-rules-title{color:var(--archive-ink)!important;font-size:26rpx!important;font-weight:700}.share-actions{margin-top:24rpx}.share-button{background:var(--archive-cinnabar);color:#fff;font-size:27rpx;border-radius:10rpx}.share-button::after{border:0}</style>
