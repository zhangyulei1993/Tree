<template>
  <PageShell>
    <section class="container page-section">
      <div class="page-heading"><div><span class="eyebrow">账号安全</span><h1>账号设置</h1></div><RouterLink class="button secondary" to="/me">返回个人中心</RouterLink></div>
      <p v-if="operationError" class="card feedback error">{{ operationError }}</p>

      <form class="card panel" @submit.prevent="saveProfile">
        <h2>个人资料</h2>
        <label><span>昵称</span><input v-model.trim="nickname" class="field" maxlength="30" /></label>
        <button class="button" :disabled="acting">保存昵称</button>
      </form>

      <section class="card panel">
        <h2>更改手机号</h2>
        <p class="muted">当前手机号：{{ maskedPhone }}</p>
        <div class="code-row"><input v-model.trim="oldPhoneCode" class="field" maxlength="6" placeholder="当前手机号验证码" /><button class="button secondary" :disabled="acting" @click="sendOldCode">发送验证码</button></div>
        <input v-model.trim="newPhone" class="field" maxlength="11" placeholder="新手机号" />
        <div class="code-row"><input v-model.trim="newPhoneCode" class="field" maxlength="6" placeholder="新手机号验证码" /><button class="button secondary" :disabled="acting" @click="sendNewCode">发送验证码</button></div>
        <button class="button" :disabled="acting || !oldPhoneCode || !newPhoneCode || !newPhone" @click="submitPhone">确认更改</button>
      </section>

      <section class="card panel danger-panel">
        <h2>注销账号</h2>
        <p>注销前必须先退出所有家庭。家谱节点不会随账号注销而删除。</p>
        <div class="code-row"><input v-model.trim="cancelCode" class="field" maxlength="6" placeholder="当前手机号验证码" /><button class="button secondary" :disabled="acting" @click="sendCancelCode">发送验证码</button></div>
        <textarea v-model.trim="cancelReason" class="field textarea" maxlength="200" placeholder="注销原因（可选）" />
        <button class="button danger" :disabled="acting || !cancelCode" @click="submitCancel">确认注销账号</button>
      </section>
    </section>
  </PageShell>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { apiErrorMessage } from '@/api/client'
import { cancelAccount, changePhone, getMe, sendCode, updateProfile } from '@/api/auth'
import PageShell from '@/components/PageShell.vue'
import { useSessionStore } from '@/stores/session'

const session = useSessionStore()
const router = useRouter()
const nickname = ref('')
const oldPhoneCode = ref('')
const newPhone = ref('')
const newPhoneCode = ref('')
const cancelCode = ref('')
const cancelReason = ref('')
const operationError = ref('')
const acting = ref(false)
const maskedPhone = computed(() => (session.user?.phone || '').replace(/^(\d{3})\d{4}(\d{4})$/, '$1****$2') || '未绑定')

async function run(action: () => Promise<void>, success: string) {
  acting.value = true; operationError.value = ''
  try { await action(); window.alert(success) }
  catch (error) { operationError.value = apiErrorMessage(error, '操作失败。') }
  finally { acting.value = false }
}

function saveProfile() {
  if (!nickname.value) return
  void run(async () => { const user = await updateProfile(nickname.value); session.persistSession(session.token, user) }, '资料已保存')
}
function sendOldCode() { const phone = session.user?.phone; if (phone) void run(async () => { await sendCode({ phone, scene: 'CHANGE_PHONE_OLD', clientType: 'PC_WEB' }) }, '验证码已发送') }
function sendNewCode() { if (/^1\d{10}$/.test(newPhone.value)) void run(async () => { await sendCode({ phone: newPhone.value, scene: 'CHANGE_PHONE_NEW', clientType: 'PC_WEB' }) }, '验证码已发送'); else operationError.value = '请输入正确的新手机号。' }
function sendCancelCode() { const phone = session.user?.phone; if (phone) void run(async () => { await sendCode({ phone, scene: 'CANCEL_ACCOUNT', clientType: 'PC_WEB' }) }, '验证码已发送') }
function submitPhone() { void run(async () => { const result = await changePhone({ oldPhoneCode: oldPhoneCode.value, newPhone: newPhone.value, newPhoneCode: newPhoneCode.value }); session.persistSession(result.accessToken, result.user) }, '手机号已更改') }
async function submitCancel() {
  if (!window.confirm('注销后账号将无法继续登录，确认继续？')) return
  acting.value = true; operationError.value = ''
  try { await cancelAccount({ phoneCode: cancelCode.value, cancelReason: cancelReason.value || undefined }); session.clearSession(); await router.push('/') }
  catch (error) { operationError.value = apiErrorMessage(error, '注销失败。请确认已退出所有家庭。') }
  finally { acting.value = false }
}

onMounted(async () => {
  try { const user = await getMe(); session.persistSession(session.token, user); nickname.value = user.nickname || '' }
  catch (error) { operationError.value = apiErrorMessage(error, '账号资料加载失败。') }
})
</script>

<style scoped>
.page-section { max-width: 860px; padding: 44px 0; }
.page-heading, .code-row { display: flex; align-items: center; justify-content: space-between; gap: 12px; }
h1 { margin: 8px 0 0; } h2 { color: var(--color-primary); }
.eyebrow { color: var(--color-heritage-green); font-size: 13px; font-weight: 700; }
.panel { display: grid; gap: 14px; margin-top: 16px; padding: 24px; }
.panel label { display: grid; gap: 7px; }.code-row .field { flex: 1; }.textarea { min-height: 100px; }
.danger-panel { border-color: rgba(181, 71, 60, .3); }.feedback { padding: 14px; }.error { color: var(--color-danger); }
@media(max-width: 640px) { .page-heading, .code-row { align-items: stretch; flex-direction: column; } }
</style>
