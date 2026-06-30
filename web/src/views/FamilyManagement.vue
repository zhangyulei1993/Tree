<template>
  <PageShell>
    <section class="container page-section">
      <div class="page-heading">
        <div><span class="eyebrow">家庭管理</span><h1>{{ family?.familyName || '家庭管理' }}</h1></div>
        <div class="record-actions">
          <RouterLink class="button secondary" :to="`/families/${familyId}/members`">成员与绑定</RouterLink>
          <RouterLink class="button secondary" :to="`/families/${familyId}`">返回家庭详情</RouterLink>
        </div>
      </div>

      <section v-if="loading" class="card state-panel">正在加载管理数据...</section>
      <section v-else-if="loadError" class="card state-panel error">
        <strong>管理数据加载失败</strong><span>{{ loadError }}</span>
        <button class="button secondary" @click="loadData">重新加载</button>
      </section>
      <template v-else-if="family">
        <p v-if="operationError" class="card operation-error" role="alert">操作失败：{{ operationError }}</p>

        <form v-if="canManage" class="card panel profile-form" @submit.prevent="saveFamily">
          <div class="panel-head">
            <div><h2>家庭资料</h2><p>维护家庭内部信息、公开页资料和搜索设置。</p></div>
          </div>
          <div class="form-grid">
            <label><span>家庭名称</span><input v-model.trim="familyForm.familyName" class="field" maxlength="120" /></label>
            <label><span>籍贯</span><input v-model.trim="familyForm.nativePlace" class="field" maxlength="120" /></label>
            <label><span>地区</span><input v-model.trim="familyForm.regionText" class="field" maxlength="120" /></label>
            <label><span>公开联系人</span><input v-model.trim="familyForm.publicContactName" class="field" maxlength="80" /></label>
            <label><span>公开联系电话</span><input v-model.trim="familyForm.publicContactPhone" class="field" maxlength="30" /></label>
            <label><span>公开微信</span><input v-model.trim="familyForm.publicContactWechat" class="field" maxlength="80" /></label>
          </div>
          <label><span>家庭简介</span><textarea v-model.trim="familyForm.description" class="field textarea" maxlength="1000" /></label>
          <label><span>公开联系说明</span><textarea v-model.trim="familyForm.publicContactNote" class="field textarea" maxlength="300" /></label>
          <div class="record-actions">
            <label class="check-row"><input v-model="familyForm.searchable" type="checkbox" /><span>允许在公开家庭搜索中展示</span></label>
            <label class="check-row"><input v-model="familyForm.publicContactVisible" type="checkbox" /><span>公开联系方式</span></label>
          </div>
          <button class="button" :disabled="acting">{{ acting ? '保存中...' : '保存家庭资料' }}</button>
        </form>

        <section v-if="canManage" class="card panel">
          <div class="panel-head"><div><h2>加入申请</h2><p>可绑定已有节点，也可创建节点并立即确定其家谱位置。</p></div></div>
          <p v-if="joinRequests.length === 0" class="empty">暂无加入申请</p>
          <article v-for="item in joinRequests" :key="item.requestId" class="record">
            <div>
              <strong>{{ item.applicantRealName || '未填写姓名' }}</strong>
              <span>{{ genderText(item.applicantGender) }} · {{ statusText(item.requestStatus) }}</span>
              <p>{{ item.applicantMessage || '未填写申请说明' }}</p>
            </div>
            <div v-if="item.requestStatus === 'PENDING'" class="record-actions">
              <select v-model="selectedMemberByRequest[String(item.requestId)]" class="field compact">
                <option value="">选择性别一致的未绑定节点</option>
                <option v-for="member in bindableMembersFor(item)" :key="member.memberId" :value="String(member.memberId)">
                  {{ member.name }}
                </option>
              </select>
              <button class="button secondary" :disabled="acting" @click="approveExisting(item)">绑定已有成员</button>
              <button class="button" :disabled="acting" @click="approveStandalone(item)">创建待定位成员</button>
              <button class="button" :disabled="acting" @click="startLocatedApproval(item)">创建并定位</button>
              <button class="button danger" :disabled="acting" @click="rejectRequest(item)">驳回</button>
            </div>
            <form
              v-if="item.requestStatus === 'PENDING' && activeLocationRequestId === String(item.requestId)"
              class="resolve-panel"
              @submit.prevent="approveLocated(item)"
            >
              <strong>创建节点并放入家谱</strong>
              <div class="form-grid">
                <label><span>成员姓名</span><input v-model.trim="locationForm.name" class="field" maxlength="100" /></label>
                <label><span>申请人性别</span><input class="field" :value="genderText(item.applicantGender)" disabled /></label>
                <label>
                  <span>基准成员</span>
                  <select v-model.number="locationForm.baseMemberId" class="field">
                    <option :value="0" disabled>请选择成员</option>
                    <option v-for="member in members" :key="member.memberId" :value="member.memberId">{{ member.name }}</option>
                  </select>
                </label>
                <label>
                  <span>与基准成员的关系</span>
                  <select v-model="locationForm.addType" class="field">
                    <option value="ADD_FATHER">父亲</option>
                    <option value="ADD_MOTHER">母亲</option>
                    <option value="ADD_CHILD">子女</option>
                    <option value="ADD_SPOUSE">配偶</option>
                    <option value="ADD_SIBLING">兄弟姐妹</option>
                  </select>
                </label>
                <label v-if="locationForm.addType !== 'ADD_SPOUSE'">
                  <span>关系性质</span>
                  <select v-model="locationForm.parentLinkType" class="field">
                    <option value="PRIMARY">亲生</option>
                    <option value="STEP">继亲</option>
                    <option value="ADOPTIVE">收养</option>
                    <option value="SUCCESSION">承继</option>
                    <option value="NOTE_ONLY">仅备注</option>
                    <option value="OTHER">其他</option>
                  </select>
                </label>
              </div>
              <p class="notice">申请人性别来自加入申请，不能在审核时改成另一性别。创建节点、建立关系和账号绑定在同一事务中完成。</p>
              <div class="record-actions">
                <button class="button" :disabled="acting">确认创建并通过</button>
                <button class="button secondary" type="button" @click="activeLocationRequestId = ''">取消</button>
              </div>
            </form>
          </article>
        </section>

        <section v-if="canManage" class="card panel">
          <div class="panel-head">
            <div><h2>已发邀请</h2><p>邀请链接只在生成时显示，不会长期保存到浏览器。</p></div>
            <RouterLink class="button secondary" :to="`/families/${familyId}/members`">前往成员列表发起邀请</RouterLink>
          </div>
          <div v-if="generatedInviteLink" class="notice">
            <strong>新邀请链接</strong><code>{{ generatedInviteLink }}</code>
            <button class="button secondary" @click="copyInviteLink">复制链接</button>
          </div>
          <p v-if="invitations.length === 0" class="empty">暂无已发邀请</p>
          <article v-for="item in invitations" :key="item.invitationId" class="record">
            <div><strong>{{ item.targetMemberName }}</strong><span>{{ statusText(item.status) }} · {{ item.inviteChannel === 'SHARE_LINK' ? '分享链接' : '站内邀请' }}</span></div>
            <div v-if="item.status === 'PENDING'" class="record-actions">
              <button class="button secondary" :disabled="acting" @click="regenerate(item)">重新生成</button>
              <button class="button danger" :disabled="acting" @click="cancelInvite(item)">取消邀请</button>
            </div>
          </article>
        </section>

        <section v-if="family.role === 'FOUNDER'" class="card panel">
          <div class="panel-head"><div><h2>家庭管理员</h2><p>只有创建者可以授予或取消管理员角色。</p></div></div>
          <article v-for="member in boundMembers" :key="member.memberId" class="record">
            <div><strong>{{ member.name }}</strong><span>{{ roleText(member.boundFamilyRole) }}</span></div>
            <div class="record-actions">
              <button v-if="member.boundFamilyRole === 'MEMBER'" class="button secondary" :disabled="acting" @click="changeAdmin(member, true)">设为管理员</button>
              <button v-if="member.boundFamilyRole === 'FAMILY_ADMIN'" class="button secondary" :disabled="acting" @click="changeAdmin(member, false)">取消管理员</button>
            </div>
          </article>
        </section>

        <section v-if="family.role === 'FOUNDER'" class="card panel">
          <div class="panel-head"><div><h2>创建者转让</h2><p>申请需平台管理员审核，审核前可以取消。</p></div></div>
          <div v-if="currentTransfer" class="notice">
            <span>待审核：转让给 {{ memberName(currentTransfer.toMemberId) }}</span>
            <button class="button secondary" :disabled="acting" @click="cancelTransfer">取消申请</button>
          </div>
          <div v-else class="inline-form">
            <select v-model="transferTargetId" class="field">
              <option value="">选择接任成员</option>
              <option v-for="member in transferTargets" :key="member.memberId" :value="String(member.memberId)">{{ member.name }}</option>
            </select>
            <button class="button" :disabled="!transferTargetId || acting" @click="createTransfer">发起转让</button>
          </div>
        </section>

        <section v-if="family.role === 'FOUNDER'" class="card panel danger-panel">
          <div class="panel-head"><div><h2>解散家庭</h2><p>高风险操作，审核通过后家庭停止使用。</p></div></div>
          <div v-if="currentDissolution" class="notice">
            <span>解散申请待审核：{{ currentDissolution.requestReason || '未填写原因' }}</span>
            <button class="button secondary" :disabled="acting" @click="cancelDissolutionRequest">取消申请</button>
          </div>
          <button v-else class="button danger" :disabled="acting" @click="createDissolutionRequest">申请解散家庭</button>
        </section>

        <section v-if="family.role !== 'FOUNDER'" class="card panel danger-panel">
          <div class="panel-head"><div><h2>退出家庭</h2><p>仅解除账号绑定，家谱成员节点与关系会保留。</p></div></div>
          <button class="button danger" :disabled="acting" @click="leave">退出该家庭</button>
        </section>
      </template>
    </section>
  </PageShell>
</template>

<script setup lang="ts">
import axios from 'axios'
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { apiErrorMessage } from '@/api/client'
import {
  cancelDissolution,
  cancelFounderTransfer,
  createDissolution,
  createFounderTransfer,
  getCurrentDissolution,
  getCurrentFounderTransfer,
  setFamilyAdmin,
  unsetFamilyAdmin
} from '@/api/familyManagement'
import { getFamilyDetail, leaveFamily, updateFamily } from '@/api/families'
import { cancelInvitation, listFamilyInvitations, regenerateInvitation } from '@/api/invitations'
import { approveJoinRequest, listFamilyJoinRequests, rejectJoinRequest } from '@/api/joinRequests'
import { listMembers } from '@/api/members'
import PageShell from '@/components/PageShell.vue'
import type {
  DissolutionRequest,
  FamilyDetail,
  FamilyMember,
  FounderTransferRequest,
  Invitation,
  JoinRequest,
  RelationshipAddType
} from '@/types/api'

const route = useRoute()
const router = useRouter()
const familyId = computed(() => String(route.params.familyId))
const family = ref<FamilyDetail | null>(null)
const members = ref<FamilyMember[]>([])
const joinRequests = ref<JoinRequest[]>([])
const invitations = ref<Invitation[]>([])
const currentTransfer = ref<FounderTransferRequest | null>(null)
const currentDissolution = ref<DissolutionRequest | null>(null)
const selectedMemberByRequest = reactive<Record<string, string>>({})
const activeLocationRequestId = ref('')
const locationForm = reactive({
  name: '',
  baseMemberId: 0,
  addType: 'ADD_CHILD' as RelationshipAddType,
  parentLinkType: 'PRIMARY'
})
const familyForm = reactive({
  familyName: '',
  nativePlace: '',
  regionText: '',
  description: '',
  searchable: false,
  publicContactName: '',
  publicContactPhone: '',
  publicContactWechat: '',
  publicContactNote: '',
  publicContactVisible: false
})
const transferTargetId = ref('')
const generatedInviteLink = ref('')
const loading = ref(true)
const loadError = ref('')
const operationError = ref('')
const acting = ref(false)
const canManage = computed(() => family.value?.role === 'FOUNDER' || family.value?.role === 'FAMILY_ADMIN')
const boundMembers = computed(() => members.value.filter((item) => item.status === 'ACTIVE' && item.boundUserId))
const transferTargets = computed(() => boundMembers.value.filter((item) => item.boundFamilyRole !== 'FOUNDER'))

function syncFamilyForm(value: FamilyDetail) {
  familyForm.familyName = value.familyName
  familyForm.nativePlace = value.nativePlace || ''
  familyForm.regionText = value.regionText || ''
  familyForm.description = value.description || ''
  familyForm.searchable = value.searchable
  familyForm.publicContactName = value.publicContactName || ''
  familyForm.publicContactPhone = value.publicContactPhone || ''
  familyForm.publicContactWechat = value.publicContactWechat || ''
  familyForm.publicContactNote = value.publicContactNote || ''
  familyForm.publicContactVisible = value.publicContactVisible
}

function bindableMembersFor(request: JoinRequest) {
  return members.value.filter((member) =>
    member.status === 'ACTIVE' &&
    !member.boundUserId &&
    member.userBindingPolicy !== 'NOT_REQUIRED' &&
    member.gender === request.applicantGender
  )
}

async function optional<T>(factory: () => Promise<T>) {
  try { return await factory() } catch (error) {
    if (axios.isAxiosError(error) && error.response?.status === 404) return null
    throw error
  }
}

async function loadData() {
  loading.value = true
  loadError.value = ''
  operationError.value = ''
  try {
    family.value = await getFamilyDetail(familyId.value)
    syncFamilyForm(family.value)
    members.value = await listMembers(familyId.value)
    if (canManage.value) {
      const [requests, sent] = await Promise.all([listFamilyJoinRequests(familyId.value), listFamilyInvitations(familyId.value)])
      joinRequests.value = requests
      invitations.value = sent
    }
    if (family.value.role === 'FOUNDER') {
      const [transfer, dissolution] = await Promise.all([
        optional(() => getCurrentFounderTransfer(familyId.value)),
        optional(() => getCurrentDissolution(familyId.value))
      ])
      currentTransfer.value = transfer
      currentDissolution.value = dissolution
    }
  } catch (error) {
    loadError.value = apiErrorMessage(error, '无法加载家庭管理数据。')
  } finally { loading.value = false }
}

async function run(action: () => Promise<unknown>, success: string) {
  acting.value = true
  operationError.value = ''
  try { await action(); await loadData(); window.alert(success) }
  catch (error) { operationError.value = apiErrorMessage(error, '操作未完成。') }
  finally { acting.value = false }
}

async function saveFamily() {
  acting.value = true
  operationError.value = ''
  try {
    family.value = await updateFamily(familyId.value, { ...familyForm })
    syncFamilyForm(family.value)
    window.alert('家庭资料已保存')
  } catch (error) {
    operationError.value = apiErrorMessage(error, '家庭资料保存失败。')
  } finally {
    acting.value = false
  }
}

function approveExisting(item: JoinRequest) {
  const memberId = selectedMemberByRequest[String(item.requestId)]
  if (!memberId) { operationError.value = '请先选择一个未绑定成员。'; return }
  if (!window.confirm('确认将申请账号绑定到所选成员并通过申请？')) return
  void run(() => approveJoinRequest(familyId.value, item.requestId, {
    approveMode: 'BIND_EXISTING_MEMBER',
    memberId: Number(memberId)
  }), '申请已通过')
}

function approveStandalone(item: JoinRequest) {
  const name = window.prompt('请输入新成员姓名', item.applicantRealName || '')?.trim()
  if (!name) return
  const gender = item.applicantGender === 'FEMALE' ? 'FEMALE' : 'MALE'
  if (!window.confirm(`将创建“${name}”并暂存为待定位成员，确认通过？`)) return
  void run(() => approveJoinRequest(familyId.value, item.requestId, {
    approveMode: 'CREATE_NEW_MEMBER', newMember: { name, gender, isAlive: true, userBindingPolicy: 'OPTIONAL' }
  }), '申请已通过，成员待定位')
}

function startLocatedApproval(item: JoinRequest) {
  if (item.applicantGender !== 'MALE' && item.applicantGender !== 'FEMALE') {
    operationError.value = '申请人未填写有效性别，无法创建并定位节点。'
    return
  }
  activeLocationRequestId.value = String(item.requestId)
  locationForm.name = item.applicantRealName || ''
  locationForm.baseMemberId = members.value[0]?.memberId || 0
  locationForm.addType = 'ADD_CHILD'
  locationForm.parentLinkType = 'PRIMARY'
}

async function approveLocated(item: JoinRequest) {
  const name = locationForm.name.trim()
  if (!name || !locationForm.baseMemberId) {
    operationError.value = '请填写成员姓名并选择基准成员。'
    return
  }
  if (item.applicantGender !== 'MALE' && item.applicantGender !== 'FEMALE') {
    operationError.value = '申请人性别无效，不能创建成员节点。'
    return
  }
  if (!window.confirm(`确认创建“${name}”、放入家谱并绑定该申请账号？`)) return
  acting.value = true
  operationError.value = ''
  try {
    await approveJoinRequest(familyId.value, item.requestId, {
      approveMode: 'CREATE_NEW_MEMBER',
      newMember: {
        name,
        gender: item.applicantGender,
        isAlive: true,
        userBindingPolicy: 'OPTIONAL'
      },
      location: {
        baseMemberId: locationForm.baseMemberId,
        addType: locationForm.addType,
        relationship: locationForm.addType === 'ADD_SPOUSE'
          ? { relationshipType: 'SPOUSE' }
          : { relationshipType: 'PARENT_CHILD', parentLinkType: locationForm.parentLinkType }
      }
    })
    activeLocationRequestId.value = ''
    await loadData()
    window.alert('申请已通过，账号与家谱节点已绑定')
  } catch (error) {
    operationError.value = apiErrorMessage(error, '创建并定位成员失败。')
  } finally {
    acting.value = false
  }
}

function rejectRequest(item: JoinRequest) {
  const handleComment = window.prompt('请输入驳回原因')?.trim()
  if (!handleComment) return
  void run(() => rejectJoinRequest(familyId.value, item.requestId, { handleComment }), '申请已驳回')
}

function cancelInvite(item: Invitation) {
  if (!window.confirm(`确认取消发给“${item.targetMemberName}”的邀请？`)) return
  void run(() => cancelInvitation(item.invitationId, { reason: 'PC 家庭管理取消' }), '邀请已取消')
}

async function regenerate(item: Invitation) {
  if (!window.confirm('重新生成后旧邀请立即失效，确认继续？')) return
  acting.value = true; operationError.value = ''
  try {
    const result = await regenerateInvitation(item.invitationId)
    generatedInviteLink.value = `${window.location.origin}/invite/${encodeURIComponent(result.inviteToken)}`
    await loadData()
  } catch (error) { operationError.value = apiErrorMessage(error, '重新生成邀请失败。') }
  finally { acting.value = false }
}

async function copyInviteLink() { await navigator.clipboard.writeText(generatedInviteLink.value); window.alert('邀请链接已复制') }

function changeAdmin(member: FamilyMember, enabled: boolean) {
  if (!window.confirm(`确认${enabled ? '设置' : '取消'}“${member.name}”的管理员角色？`)) return
  void run(() => enabled ? setFamilyAdmin(familyId.value, member.memberId) : unsetFamilyAdmin(familyId.value, member.memberId), '角色已更新')
}

function createTransfer() {
  const reason = window.prompt('转让原因（可选）') || undefined
  if (!window.confirm('创建者转让需要平台审核，确认提交？')) return
  void run(() => createFounderTransfer(familyId.value, Number(transferTargetId.value), reason), '转让申请已提交')
}

function cancelTransfer() {
  if (!currentTransfer.value || !window.confirm('确认取消待审核的创建者转让申请？')) return
  void run(() => cancelFounderTransfer(familyId.value, currentTransfer.value!.requestId, '创建者主动取消'), '转让申请已取消')
}

function createDissolutionRequest() {
  const reason = window.prompt('请输入解散原因')?.trim()
  if (!reason || !window.confirm('解散属于高风险操作，确认提交审核？')) return
  void run(() => createDissolution(familyId.value, reason), '解散申请已提交')
}

function cancelDissolutionRequest() {
  if (!currentDissolution.value || !window.confirm('确认取消待审核的解散申请？')) return
  void run(() => cancelDissolution(familyId.value, currentDissolution.value!.requestId, '创建者主动取消'), '解散申请已取消')
}

async function leave() {
  if (!window.confirm('退出后账号会解除绑定，但家谱成员节点仍保留。确认退出？')) return
  acting.value = true
  try { await leaveFamily(familyId.value, '用户主动退出家庭'); await router.push('/me/families') }
  catch (error) { operationError.value = apiErrorMessage(error, '退出家庭失败。') }
  finally { acting.value = false }
}

function memberName(id: number | string) { return members.value.find((item) => String(item.memberId) === String(id))?.name || '目标成员' }
function genderText(value?: string | null) { return value === 'MALE' ? '男' : value === 'FEMALE' ? '女' : '性别未填写' }
function roleText(value?: string | null) { return value === 'FOUNDER' ? '创建者' : value === 'FAMILY_ADMIN' ? '管理员' : '成员' }
function statusText(value: string) {
  const labels: Record<string, string> = { PENDING: '待处理', APPROVED: '已通过', REJECTED: '已驳回', CANCELLED: '已取消', ACCEPTED: '已接受', EXPIRED: '已过期' }
  return labels[value] || '未知状态'
}

onMounted(loadData)
</script>

<style scoped>
.page-section { padding: 44px 0; }
.page-heading, .panel-head, .record, .record-actions, .inline-form, .notice { display: flex; gap: 14px; }
.page-heading, .panel-head, .record { align-items: flex-start; justify-content: space-between; }
.page-heading { margin-bottom: 18px; }
h1 { margin: 8px 0 0; } h2 { margin: 0; color: var(--color-primary); }
.eyebrow { color: var(--color-heritage-green); font-size: 13px; font-weight: 700; }
.panel { margin-top: 16px; padding: 24px; }
.profile-form, .resolve-panel, label { display: grid; gap: 10px; }
.form-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 14px; }
.textarea { min-height: 96px; resize: vertical; }
.check-row { display: flex; align-items: center; font-weight: 600; }
.resolve-panel { flex: 1 0 100%; margin-top: 14px; border: 1px solid var(--color-border); border-radius: 10px; padding: 18px; background: var(--color-surface-soft); }
.panel-head p, .record p { margin: 6px 0 0; color: var(--color-text-secondary); }
.record { padding: 18px 0; border-top: 1px solid var(--color-border); }
.record strong, .record span { display: block; } .record span { margin-top: 5px; color: var(--color-text-secondary); }
.record-actions, .inline-form, .notice { align-items: center; flex-wrap: wrap; }
.compact { min-width: 190px; padding: 9px 12px; }
.notice { margin: 14px 0; border-radius: 10px; background: var(--color-heritage-green-light); padding: 14px; }
.notice code { max-width: 100%; overflow-wrap: anywhere; }
.empty, .state-panel { color: var(--color-text-secondary); text-align: center; }
.state-panel { display: grid; min-height: 180px; place-items: center; gap: 12px; padding: 28px; }
.error, .operation-error { color: var(--color-danger); } .operation-error { padding: 14px 18px; }
.danger-panel { border-color: rgba(181, 71, 60, .26); }
@media (max-width: 760px) {
  .page-heading, .panel-head, .record { flex-direction: column; }
  .record-actions { width: 100%; }
  .form-grid { grid-template-columns: 1fr; }
}
</style>
