<template>
  <PageShell>
    <section class="container invite-page">
      <section v-if="loading" class="card state-card">正在加载邀请详情...</section>
      <section v-else-if="error && !invitation" class="card state-card error" role="alert">
        <strong>邀请详情不可访问</strong>
        <span>{{ error }}</span>
        <button class="button secondary" @click="loadInvitation">重新加载</button>
      </section>
      <div v-else-if="invitation" class="card invite-card">
        <span class="eyebrow">邀请确认</span>
        <h1>你收到一个家庭成员绑定邀请</h1>
        <dl>
          <div><dt>家庭</dt><dd>{{ invitation.familyName }}</dd></div>
          <div><dt>成员</dt><dd>{{ invitation.targetMemberName }}</dd></div>
          <div><dt>邀请方式</dt><dd>{{ inviteChannelText(invitation.inviteChannel) }}</dd></div>
          <div><dt>绑定后角色</dt><dd>{{ roleText(invitation.familyRoleAfterAccept) }}</dd></div>
          <div><dt>状态</dt><dd>{{ invitationStatusText(invitation.status) }}</dd></div>
          <div><dt>有效期至</dt><dd>{{ formatDate(invitation.expiredAt) }}</dd></div>
        </dl>
        <p v-if="invitation.inviteMessage" class="message">{{ invitation.inviteMessage }}</p>

        <template v-if="invitation.status === 'PENDING'">
          <div v-if="!session.isLoggedIn" class="notice">
            登录后可以接受或拒绝邀请。
            <div class="actions guest-actions">
              <button class="button" @click="goToLogin">登录后接受</button>
              <button class="button secondary" @click="goToLogin">登录后拒绝</button>
            </div>
          </div>
          <div v-else-if="!session.isPhoneBound" class="notice">
            当前账号尚未完成手机号验证。请先完成真实手机号绑定后再处理邀请。
          </div>
          <div v-else class="decision">
            <textarea
              v-model.trim="rejectReason"
              class="field textarea"
              maxlength="300"
              placeholder="拒绝原因（可选，仅拒绝时提交）"
            />
            <div class="actions">
              <button class="button" :disabled="Boolean(acting)" @click="accept">
                {{ acting === 'accept' ? '接受中...' : '接受绑定' }}
              </button>
              <button class="button secondary" :disabled="Boolean(acting)" @click="reject">
                {{ acting === 'reject' ? '拒绝中...' : '拒绝邀请' }}
              </button>
            </div>
          </div>
        </template>
        <p v-else class="notice">该邀请当前状态为“{{ invitationStatusText(invitation.status) }}”，不能继续处理。</p>

        <p v-if="error" class="feedback error" role="alert">{{ error }}</p>
        <p v-if="result" class="feedback success" role="status">{{ result }}</p>
      </div>
    </section>
  </PageShell>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { apiErrorMessage } from '@/api/client'
import {
  acceptInvitation,
  getInvitationDetail,
  rejectInvitation
} from '@/api/invitations'
import PageShell from '@/components/PageShell.vue'
import { useSessionStore } from '@/stores/session'
import type { Invitation } from '@/types/api'

const route = useRoute()
const router = useRouter()
const session = useSessionStore()
const inviteToken = computed(() => String(route.params.inviteToken || ''))
const invitation = ref<Invitation | null>(null)
const loading = ref(true)
const acting = ref<'' | 'accept' | 'reject'>('')
const rejectReason = ref('')
const error = ref('')
const result = ref('')

function formatDate(value: string) {
  return new Date(value).toLocaleString('zh-CN')
}

function invitationStatusText(status?: string) {
  if (status === 'PENDING') return '待处理'
  if (status === 'ACCEPTED') return '已接受'
  if (status === 'REJECTED') return '已拒绝'
  if (status === 'EXPIRED') return '已过期'
  if (status === 'CANCELLED') return '已取消'
  return '未知状态'
}

function inviteChannelText(channel?: string) {
  if (channel === 'SHARE_LINK') return '链接邀请'
  if (channel === 'IN_APP') return '站内邀请'
  return '其他方式'
}

function roleText(role?: string) {
  if (role === 'FOUNDER') return '家庭创建者'
  if (role === 'FAMILY_ADMIN') return '家庭管理员'
  if (role === 'MEMBER') return '家庭成员'
  return '未知角色'
}

function goToLogin() {
  router.push({ path: '/login', query: { redirect: route.fullPath } })
}

async function loadInvitation() {
  loading.value = true
  error.value = ''
  result.value = ''
  try {
    invitation.value = await getInvitationDetail(inviteToken.value)
  } catch (requestError) {
    invitation.value = null
    error.value = apiErrorMessage(requestError, '无法加载邀请详情。')
  } finally {
    loading.value = false
  }
}

async function accept() {
  if (!invitation.value) return
  acting.value = 'accept'
  error.value = ''
  result.value = ''
  try {
    invitation.value = await acceptInvitation(invitation.value.invitationId)
    result.value = '邀请已接受，你已绑定到该家庭成员。'
  } catch (requestError) {
    error.value = apiErrorMessage(requestError, '接受邀请失败。')
  } finally {
    acting.value = ''
  }
}

async function reject() {
  if (!invitation.value) return
  acting.value = 'reject'
  error.value = ''
  result.value = ''
  try {
    invitation.value = await rejectInvitation(invitation.value.invitationId, {
      reason: rejectReason.value || undefined
    })
    result.value = '邀请已拒绝。'
  } catch (requestError) {
    error.value = apiErrorMessage(requestError, '拒绝邀请失败。')
  } finally {
    acting.value = ''
  }
}

onMounted(loadInvitation)
watch(inviteToken, loadInvitation)
</script>

<style scoped>
.invite-page {
  display: grid;
  min-height: 560px;
  place-items: center;
  padding: 32px 0;
}

.invite-card,
.state-card {
  width: min(680px, 100%);
  padding: 24px;
}

.state-card {
  display: grid;
  min-height: 220px;
  place-items: center;
  gap: 12px;
  text-align: center;
}

.eyebrow {
  color: var(--color-success);
  font-weight: 700;
}

dl {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

dt {
  color: var(--color-text-secondary);
}

dd {
  margin: 4px 0 0;
  font-weight: 700;
}

.message,
.notice {
  border-left: 3px solid var(--color-warm-gold);
  background: #fbf7ed;
  padding: 12px;
  line-height: 1.7;
}

.notice,
.actions {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
  margin-top: 18px;
}

.guest-actions {
  margin-top: 0;
}

.decision {
  display: grid;
  gap: 12px;
  margin-top: 18px;
}

.feedback {
  font-weight: 700;
}

.error {
  color: var(--color-danger);
}

.success {
  color: var(--color-success);
}

.button:disabled {
  cursor: not-allowed;
  opacity: 0.6;
}

@media (max-width: 620px) {
  dl {
    grid-template-columns: 1fr;
  }
}
</style>
