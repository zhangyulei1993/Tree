<template>
  <view class="archive-page account-security-page">
    <MiniBackHome />

    <view class="security-head archive-page-head">
      <view>
        <text class="archive-kicker">Account Security</text>
        <text class="archive-title">账号与安全</text>
        <text class="archive-subtitle">登录、账号权益与安全操作</text>
      </view>
      <view class="archive-seal">安</view>
    </view>

    <template v-if="session.isLoggedIn && session.user">
      <view class="security-status archive-form-panel">
        <view class="archive-section-head">
          <text class="archive-section-title">账号状态</text>
          <text class="archive-section-subtitle">当前平台账号与信任等级</text>
        </view>
        <view class="status-chips">
          <text class="archive-chip">{{ accountStatusLabel }}</text>
          <text class="archive-chip">{{ trustTierLabel }}</text>
          <text v-if="session.isProfileComplete" class="archive-chip">资料已完善</text>
          <text v-else class="archive-chip">资料待完善</text>
        </view>
      </view>

      <view class="security-directory archive-list">
        <view class="archive-row security-row static">
          <view class="archive-row-main">
            <text class="archive-row-title">微信登录</text>
            <text class="archive-row-desc">当前通过微信使用 Tree，昵称 {{ session.user.nickname || '未设置' }}</text>
          </view>
          <text class="archive-row-meta">已登录</text>
        </view>

        <view class="archive-row security-row" @click="openProfile">
          <view class="archive-row-main">
            <text class="archive-row-title">{{ session.isProfileComplete ? '更新资料' : '完善资料' }}</text>
            <text class="archive-row-desc">
              {{ session.isProfileComplete ? '更新昵称和头像' : '设置昵称后即可使用完整功能' }}
            </text>
          </view>
          <text class="archive-arrow">›</text>
        </view>

        <!-- #ifdef MP-WEIXIN -->
        <button
          v-if="!session.user.phoneVerified"
          class="archive-row security-row phone-bind-row"
          open-type="getPhoneNumber"
          :disabled="bindingWechatPhone"
          @getphonenumber="bindWithWechatPhone"
        >
          <view class="archive-row-main">
            <text class="archive-row-title">手机号绑定</text>
            <text class="archive-row-desc">{{ phoneBindDesc }}</text>
          </view>
          <text class="archive-row-meta">{{ bindingWechatPhone ? '绑定中...' : phoneBindMeta }}</text>
          <text class="archive-arrow">›</text>
        </button>
        <view v-else class="archive-row security-row static">
          <view class="archive-row-main">
            <text class="archive-row-title">手机号绑定</text>
            <text class="archive-row-desc">{{ phoneBindDesc }}</text>
          </view>
          <text class="archive-row-meta">{{ phoneBindMeta }}</text>
        </view>
        <!-- #endif -->
        <!-- #ifndef MP-WEIXIN -->
        <view class="archive-row security-row static">
          <view class="archive-row-main">
            <text class="archive-row-title">手机号绑定</text>
            <text class="archive-row-desc">请在微信小程序内完成手机号绑定</text>
          </view>
          <text class="archive-row-meta">仅小程序可用</text>
        </view>
        <!-- #endif -->
      </view>

      <view v-if="session.user.phoneVerified" class="archive-form-panel assistance-panel">
        <view class="archive-section-head"><text class="archive-section-title">邀请亲友协助</text><text class="archive-section-subtitle">生成短期协助码，仅限维护家庭树成员和关系</text></view>
        <view v-if="assistanceCode" class="assistance-code"><text>{{ assistanceCode }}</text><text>请立即告知可信亲友；离开后不再展示。</text></view>
        <view class="assistance-actions"><MiniButton size="sm" @click="generateAssistance">生成 24 小时协助码</MiniButton><MiniButton v-if="activeAssistance" size="sm" variant="secondary" @click="revokeAssistance">撤销当前协助</MiniButton></view>
        <text v-if="activeAssistance" class="archive-section-subtitle">当前有效至 {{ activeAssistance.expiresAt.replace('T', ' ').slice(0, 16) }}；重新生成会立即使其失效。</text>
      </view>

      <view v-if="capabilities" class="quota-panel archive-form-panel">
        <view class="archive-section-head">
          <text class="archive-section-title">账号权益</text>
          <text class="archive-section-subtitle">当前可用额度与用量</text>
        </view>
        <view class="quota-lines">
          <text v-for="line in capabilityLines" :key="line" class="quota-line">{{ line }}</text>
        </view>
        <MiniNotice v-if="showQuotaHint" tone="info">
          当前扩展能力尚未开放。达到基础上限时，仅会阻止新增、恢复或加入，不影响查看与编辑已有内容。
        </MiniNotice>
      </view>

      <view class="archive-danger-panel">
        <text class="archive-danger-title">注销账号</text>
        <text class="archive-danger-desc">注销后账号数据将无法恢复，请谨慎操作。</text>
        <MiniButton size="sm" variant="danger" @click="go('/pages/account/cancel')">注销账号</MiniButton>
      </view>
    </template>

    <view v-else class="security-state archive-form-panel">
      <MiniEmptyState
        symbol="!"
        title="请先登录"
        description="登录后可管理账号与安全设置。"
        action-text="去登录"
        @action="go('/pages/auth/wechat-login')"
      />
    </view>
  </view>
</template>

<script setup lang="ts">
import { onShow } from '@dcloudio/uni-app'
import { computed, ref } from 'vue'

import { apiErrorMessage } from '@/api/client'
import { createAssistanceCode, listAssistanceCodes, revokeAssistanceCode, type AssistanceCodeStatus } from '@/api/auth'
import { fetchCapabilities } from '@/api/capabilities'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'
import { accountStatusText } from '@/components/base/formatStatus'
import MiniNotice from '@/components/base/MiniNotice.vue'
import { buildCapabilitiesSummary } from '@/features/quota/quotaDisplay'
import { useSessionStore } from '@/stores/session'
import type { UserCapabilities } from '@/types/api'
import { maskPhone } from '@/utils/maskPhone'

const session = useSessionStore()
const capabilities = ref<UserCapabilities | null>(null)
const bindingWechatPhone = ref(false)
const assistanceCode = ref('')
const assistanceCodes = ref<AssistanceCodeStatus[]>([])
const activeAssistance = computed(() => assistanceCodes.value.find((item) => item.status === 'ACTIVE'))

const capabilityLines = computed(() => (capabilities.value ? buildCapabilitiesSummary(capabilities.value) : []))
const accountStatusLabel = computed(() => accountStatusText(session.user?.status))
const trustTierLabel = computed(() => {
  const tier = capabilities.value?.trustTier
  if (tier === 'PHONE_BOUND') return '已绑定手机号'
  return '微信账号'
})
const phoneBindDesc = computed(() => {
  if (session.user?.phoneVerified) {
    const masked = maskPhone(session.user.phone)
    return masked ? `已绑定 ${masked}` : '手机号已绑定'
  }
  return '绑定后可发起家庭邀请与节点绑定邀请'
})
const phoneBindMeta = computed(() => {
  if (session.user?.phoneVerified) return '已绑定'
  return '未设置'
})

const showQuotaHint = computed(() => {
  if (!capabilities.value) return false
  const { limits, usage, trustTier } = capabilities.value
  return trustTier === 'WECHAT_ONLY' && (
    usage.ownedFamilies >= limits.maxOwnedFamilies ||
    usage.joinedFamilies >= limits.maxJoinedFamilies
  )
})

function go(url: string) {
  uni.navigateTo({ url })
}

function openProfile() {
  go(session.isProfileComplete ? '/pages/me/profile' : '/pages/me/profile?onboarding=1')
}

type PhoneNumberEvent = {
  detail?: {
    code?: string
    errMsg?: string
  }
}

async function refreshCapabilities() {
  try {
    capabilities.value = await fetchCapabilities()
  } catch {
    capabilities.value = null
  }
}
async function refreshAssistance() { if (!session.user?.phoneVerified) return; try { assistanceCodes.value = await listAssistanceCodes() } catch { assistanceCodes.value = [] } }
async function generateAssistance() { try { const result = await createAssistanceCode(1); assistanceCode.value = result.code; await refreshAssistance(); uni.showModal({ title: '协助码已生成', content: result.replacedPrevious ? '新协助码已生成，上一枚已立即失效。请只告知可信亲友。' : '协助码仅展示本次，请只告知可信亲友。', showCancel: false }) } catch (error) { uni.showToast({ title: apiErrorMessage(error, '生成失败'), icon: 'none' }) } }
async function revokeAssistance() { if (!activeAssistance.value) return; try { await revokeAssistanceCode(activeAssistance.value.id); assistanceCode.value = ''; await refreshAssistance(); uni.showToast({ title: '协助码已撤销', icon: 'success' }) } catch (error) { uni.showToast({ title: apiErrorMessage(error, '撤销失败'), icon: 'none' }) } }

async function bindWithWechatPhone(event: PhoneNumberEvent) {
  const phoneCode = event.detail?.code || ''
  if (!phoneCode) {
    const errMsg = event.detail?.errMsg || ''
    uni.showToast({
      title: /deny|cancel|拒绝|取消/i.test(errMsg) ? '已取消手机号授权' : '手机号授权失败',
      icon: 'none'
    })
    return
  }

  bindingWechatPhone.value = true
  try {
    await session.bindWechatPhone({ phoneCode })
    await refreshCapabilities()
    await refreshAssistance()
    uni.showToast({ title: '手机号已绑定', icon: 'success' })
  } catch (error) {
    uni.showToast({ title: apiErrorMessage(error, '手机号绑定失败'), icon: 'none' })
  } finally {
    bindingWechatPhone.value = false
  }
}

onShow(async () => {
  session.restoreSession()
  if (session.isLoggedIn) {
    await refreshCapabilities()
  }
})
</script>

<style scoped>
.account-security-page {
  padding-top: 28rpx;
}

.account-security-page :deep(.mini-back-home) {
  margin-bottom: 22rpx;
}

.security-status,
.quota-panel,
.security-state {
  margin-bottom: 24rpx;
}

.status-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 10rpx;
}

.security-directory {
  margin-bottom: 24rpx;
}
.assistance-panel { margin-bottom: 24rpx; }
.assistance-actions { display: flex; flex-wrap: wrap; gap: 14rpx; margin: 14rpx 0; }
.assistance-code { display: flex; flex-direction: column; gap: 8rpx; padding: 18rpx; background: #fff8ea; color: #755336; }
.assistance-code text:first-child { font-size: 42rpx; font-weight: 700; letter-spacing: 8rpx; color: #a83b2d; }

.security-row.static {
  cursor: default;
}

.phone-bind-row {
  width: 100%;
  border: 0;
  border-radius: 0;
  background: transparent;
  color: inherit;
  font: inherit;
  padding: 0;
  text-align: left;
}

.phone-bind-row::after {
  border: 0;
}

.quota-lines {
  display: flex;
  flex-direction: column;
  gap: 10rpx;
  margin-bottom: 16rpx;
}

.quota-line {
  color: var(--archive-ink-soft);
  font-size: 24rpx;
  line-height: 1.55;
}

.quota-panel :deep(.mini-notice) {
  margin-top: 4rpx;
}

.archive-danger-panel :deep(.mini-button) {
  align-self: flex-start;
}
</style>
