<template>
  <PageShell>
    <section class="container page-section">
      <div class="page-heading">
        <div>
          <span class="eyebrow">家庭成员</span>
          <h1>{{ family?.familyName || '成员管理' }}</h1>
          <p class="muted">当前身份：{{ roleText(family?.role) }}</p>
        </div>
        <div class="heading-actions">
          <RouterLink class="button secondary" :to="`/families/${familyId}`">家庭详情</RouterLink>
          <RouterLink v-if="canManage" class="button secondary" :to="`/families/${familyId}/manage`">申请与角色管理</RouterLink>
          <RouterLink class="button secondary" :to="`/families/${familyId}/tree`">查看家庭树</RouterLink>
        </div>
      </div>

      <section v-if="pageLoading" class="card state-panel">正在加载成员...</section>
      <section v-else-if="pageError" class="card state-panel error" role="alert">
        <strong>{{ forbidden ? '无权访问成员数据' : '成员数据加载失败' }}</strong>
        <span>{{ pageError }}</span>
        <button class="button secondary" @click="loadPage">重新加载</button>
      </section>
      <template v-else>
        <div v-if="canManage" class="forms-grid">
          <form class="card form-card" @submit.prevent="submitMember">
            <div>
              <h2>新增独立成员</h2>
              <p class="muted">只创建成员节点，不建立亲属关系。</p>
            </div>
            <label>
              <span>姓名</span>
              <input v-model.trim="memberForm.name" class="field" maxlength="100" placeholder="成员姓名" />
            </label>
            <div class="compact-grid">
              <label>
                <span>性别</span>
                <select v-model="memberForm.gender" class="field">
                  <option value="UNKNOWN">未知</option>
                  <option value="MALE">男</option>
                  <option value="FEMALE">女</option>
                </select>
              </label>
              <label>
                <span>出生年份</span>
                <input v-model.number="memberForm.birthYear" class="field" type="number" min="1" max="9999" placeholder="可选" />
              </label>
            </div>
            <label class="check-row">
              <input v-model="memberForm.isAlive" type="checkbox" />
              <span>当前健在</span>
            </label>
            <p v-if="memberError" class="feedback error" role="alert">{{ memberError }}</p>
            <button class="button" :disabled="memberSubmitting">
              {{ memberSubmitting ? '创建中...' : '新增成员' }}
            </button>
          </form>

          <form class="card form-card" @submit.prevent="submitRelationship">
            <div>
              <h2>添加新亲属</h2>
              <p class="notice">此操作会创建一个新成员并建立关系，不是连接两个已有成员。</p>
            </div>
            <label>
              <span>基础成员</span>
              <select v-model.number="relationshipForm.baseMemberId" class="field">
                <option :value="0" disabled>请选择成员</option>
                <option v-for="member in members" :key="member.memberId" :value="member.memberId">
                  {{ member.name }}
                </option>
              </select>
            </label>
            <div class="compact-grid">
              <label>
                <span>亲属类型</span>
                <select v-model="relationshipForm.addType" class="field">
                  <option value="ADD_FATHER">添加父亲</option>
                  <option value="ADD_MOTHER">添加母亲</option>
                  <option value="ADD_CHILD">添加子女</option>
                  <option value="ADD_SPOUSE">添加配偶</option>
                  <option value="ADD_SIBLING">添加兄弟姐妹</option>
                </select>
              </label>
              <label>
                <span>新亲属性别</span>
                <select v-model="relationshipForm.gender" class="field">
                  <option value="UNKNOWN">未知</option>
                  <option value="MALE">男</option>
                  <option value="FEMALE">女</option>
                </select>
              </label>
            </div>
            <label>
              <span>新亲属姓名</span>
              <input v-model.trim="relationshipForm.name" class="field" maxlength="100" placeholder="将创建的新成员姓名" />
            </label>
            <label v-if="relationshipForm.addType !== 'ADD_SPOUSE'">
              <span>父子关系属性</span>
              <select v-model="relationshipForm.parentLinkType" class="field">
                <option value="PRIMARY">主要关系</option>
                <option value="STEP">继亲</option>
                <option value="ADOPTIVE">收养</option>
                <option value="SUCCESSION">承继</option>
                <option value="NOTE_ONLY">仅备注</option>
                <option value="OTHER">其他</option>
              </select>
            </label>
            <p v-if="relationshipError" class="feedback error" role="alert">{{ relationshipError }}</p>
            <p v-if="relationshipSuccess" class="feedback success" role="status">{{ relationshipSuccess }}</p>
            <button class="button" :disabled="relationshipSubmitting || members.length === 0">
              {{ relationshipSubmitting ? '创建中...' : '创建新亲属及关系' }}
            </button>
          </form>
        </div>

        <form v-if="canManage" class="card form-card placement-card" @submit.prevent="submitPlacement">
          <div><h2>定位已有成员</h2><p class="muted">将暂存且尚未建立关系的成员连接到家谱。</p></div>
          <p v-if="unlocatedMembers.length === 0" class="notice">当前没有待定位成员。</p>
          <template v-else>
            <div class="compact-grid">
              <label><span>待定位成员</span><select v-model.number="placementForm.memberId" class="field"><option v-for="member in unlocatedMembers" :key="member.memberId" :value="member.memberId">{{ member.name }}</option></select></label>
              <label><span>基准成员</span><select v-model.number="placementForm.baseMemberId" class="field"><option v-for="member in placementBaseMembers" :key="member.memberId" :value="member.memberId">{{ member.name }}</option></select></label>
            </div>
            <label><span>与基准成员的关系</span><select v-model="placementForm.addType" class="field"><option value="ADD_FATHER">父亲</option><option value="ADD_MOTHER">母亲</option><option value="ADD_CHILD">子女</option><option value="ADD_SPOUSE">配偶</option><option value="ADD_SIBLING">兄弟姐妹</option></select></label>
            <p v-if="placementError" class="feedback error">{{ placementError }}</p>
            <button class="button" :disabled="placementSubmitting">{{ placementSubmitting ? '定位中...' : '放入家谱' }}</button>
          </template>
        </form>

        <section v-else class="card readonly-note">
          当前身份为家庭成员，仅可查看成员和家庭树。
        </section>

        <section class="member-section">
          <div class="section-heading">
            <h2>成员列表</h2>
            <button class="button secondary" :disabled="pageLoading" @click="loadPage">刷新</button>
          </div>
          <section v-if="members.length === 0" class="card state-panel">
            <strong>暂无成员</strong>
          </section>
          <div v-else class="member-list">
            <article v-for="member in members" :key="member.memberId" class="card member-card">
              <div class="member-title">
                <strong>{{ member.name }}</strong>
                <span class="status-tag">{{ memberStatusText(member.status) }}</span>
              </div>
              <dl>
                <div><dt>性别</dt><dd>{{ genderLabel(member.gender) }}</dd></div>
                <div><dt>家庭身份</dt><dd>{{ roleText(member.boundFamilyRole) }}</dd></div>
                <div><dt>健在状态</dt><dd>{{ aliveLabel(member) }}</dd></div>
                <div><dt>账号绑定</dt><dd>{{ bindingText(member) }}</dd></div>
              </dl>
              <div v-if="canManage" class="invite-actions">
                <button class="button secondary" @click="openEdit(member)">编辑成员</button>
                <button
                  v-if="canBindApplication(member)"
                  class="button secondary"
                  @click="openApplicationBinding(member)"
                >
                  绑定加入申请
                </button>
                <button v-if="canInvite(member)" class="button secondary" @click="openInvite(member.memberId)">
                  创建分享邀请
                </button>
                <RouterLink
                  v-else-if="pendingInvitation(member)"
                  class="button secondary"
                  :to="`/families/${familyId}/manage`"
                >
                  邀请待接受
                </RouterLink>
              </div>
              <form
                v-if="activeBindMemberId === member.memberId"
                class="invite-form"
                @submit.prevent="bindApplication(member)"
              >
                <label>
                  <span>选择性别一致的待处理申请</span>
                  <select v-model="selectedRequestByMember[String(member.memberId)]" class="field">
                    <option value="">请选择申请人</option>
                    <option
                      v-for="request in matchingJoinRequests(member)"
                      :key="request.requestId"
                      :value="String(request.requestId)"
                    >
                      {{ request.applicantRealName || '未填写姓名' }} · {{ request.applicantMessage || '无申请说明' }}
                    </option>
                  </select>
                </label>
                <p class="notice">通过后，申请账号会直接绑定到“{{ member.name }}”节点，不改变家谱结构和版本。</p>
                <div class="invite-buttons">
                  <button class="button" :disabled="bindingApplication">{{ bindingApplication ? '绑定中...' : '确认绑定并通过' }}</button>
                  <button class="button secondary" type="button" @click="activeBindMemberId = null">取消</button>
                </div>
              </form>
              <form
                v-if="activeEditMemberId === member.memberId"
                class="invite-form"
                @submit.prevent="saveMember(member)"
              >
                <div class="compact-grid">
                  <label><span>姓名</span><input v-model.trim="editForm.name" class="field" maxlength="100" /></label>
                  <label>
                    <span>性别</span>
                    <select v-model="editForm.gender" class="field">
                      <option value="MALE">男</option>
                      <option value="FEMALE">女</option>
                      <option value="UNKNOWN">未知</option>
                    </select>
                  </label>
                  <label><span>出生年份</span><input v-model.number="editForm.birthYear" class="field" type="number" min="1" max="9999" /></label>
                  <label class="check-row"><input v-model="editForm.isAlive" type="checkbox" /><span>目前健在</span></label>
                </div>
                <label><span>成员简介</span><textarea v-model.trim="editForm.description" class="field textarea" maxlength="500" /></label>
                <p v-if="editError" class="feedback error">{{ editError }}</p>
                <div class="invite-buttons">
                  <button class="button" :disabled="editingMember">{{ editingMember ? '保存中...' : '保存成员资料' }}</button>
                  <button class="button secondary" type="button" @click="closeEdit">取消</button>
                  <button class="button danger" type="button" :disabled="editingMember" @click="removeMember(member)">删除成员</button>
                </div>
              </form>
              <form
                v-if="canInvite(member) && activeInviteMemberId === member.memberId"
                class="invite-form"
                @submit.prevent="submitInvite(member)"
              >
                <label>
                  <span>邀请留言</span>
                  <textarea
                    v-model.trim="inviteMessage"
                    class="field textarea"
                    maxlength="300"
                    placeholder="可选，将随邀请展示"
                  />
                </label>
                <p class="notice">分享邀请接受后的身份为家庭成员，有效期为 7 天。</p>
                <p v-if="inviteError" class="feedback error" role="alert">{{ inviteError }}</p>
                <div v-if="inviteLink" class="invite-result">
                  <strong>邀请链接只在当前页面显示一次</strong>
                  <input class="field" :value="inviteLink" readonly />
                  <button class="button secondary" type="button" @click="copyInviteLink">
                    {{ copyResult || '复制链接' }}
                  </button>
                </div>
                <div class="invite-buttons">
                  <button class="button" :disabled="inviteSubmitting || Boolean(inviteLink)">
                    {{ inviteSubmitting ? '创建中...' : '创建分享邀请' }}
                  </button>
                  <button class="button secondary" type="button" @click="closeInvite">关闭</button>
                </div>
              </form>
            </article>
          </div>
        </section>

        <section v-if="canManage" class="member-section">
          <div class="section-heading">
            <div>
              <h2>关系维护</h2>
              <p class="muted">可调整父母子女关系性质，或删除错误关系后重新创建。</p>
            </div>
          </div>
          <section v-if="treeEdges.length === 0" class="card state-panel">暂无家谱关系</section>
          <div v-else class="member-list">
            <article v-for="edge in treeEdges" :key="edge.relationshipId" class="card member-card relation-card">
              <div>
                <strong>{{ relationshipSentence(edge) }}</strong>
                <p class="muted">{{ relationshipTypeText(edge.relationshipType) }}</p>
              </div>
              <div class="relation-controls">
                <select
                  v-if="edge.relationshipType === 'PARENT_CHILD'"
                  v-model="relationshipNature[String(edge.relationshipId)]"
                  class="field compact"
                >
                  <option value="PRIMARY">亲生</option>
                  <option value="STEP">继亲</option>
                  <option value="ADOPTIVE">收养</option>
                  <option value="SUCCESSION">承继</option>
                  <option value="NOTE_ONLY">仅备注</option>
                  <option value="OTHER">其他</option>
                </select>
                <button
                  v-if="edge.relationshipType === 'PARENT_CHILD'"
                  class="button secondary"
                  :disabled="relationshipActing"
                  @click="saveRelationship(edge)"
                >
                  保存性质
                </button>
                <button class="button danger" :disabled="relationshipActing" @click="removeRelationship(edge)">
                  删除关系
                </button>
              </div>
            </article>
          </div>
          <p v-if="relationshipManageError" class="feedback error">{{ relationshipManageError }}</p>
        </section>
      </template>
    </section>
  </PageShell>
</template>

<script setup lang="ts">
import axios from 'axios'
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute } from 'vue-router'

import { apiErrorMessage } from '@/api/client'
import { getFamilyDetail } from '@/api/families'
import { createInvitation, listFamilyInvitations } from '@/api/invitations'
import { approveJoinRequest, listFamilyJoinRequests } from '@/api/joinRequests'
import { createMember, deleteMember, listMembers, updateMember } from '@/api/members'
import {
  createRelationship,
  deleteRelationship,
  placeExistingMember,
  updateRelationship
} from '@/api/relationships'
import { getPrivateTree } from '@/api/tree'
import PageShell from '@/components/PageShell.vue'
import type {
  ApiResponse,
  FamilyDetail,
  FamilyMember,
  Gender,
  Invitation,
  JoinRequest,
  RelationshipAddType,
  TreeEdge
} from '@/types/api'

const route = useRoute()
const familyId = computed(() => String(route.params.familyId))
const family = ref<FamilyDetail | null>(null)
const members = ref<FamilyMember[]>([])
const pageLoading = ref(true)
const pageError = ref('')
const forbidden = ref(false)
const memberSubmitting = ref(false)
const relationshipSubmitting = ref(false)
const memberError = ref('')
const relationshipError = ref('')
const relationshipSuccess = ref('')
const relatedMemberIds = ref(new Set<number>())
const treeEdges = ref<TreeEdge[]>([])
const joinRequests = ref<JoinRequest[]>([])
const invitations = ref<Invitation[]>([])
const placementSubmitting = ref(false)
const placementError = ref('')
const activeInviteMemberId = ref<number | null>(null)
const inviteMessage = ref('')
const inviteLink = ref('')
const inviteError = ref('')
const inviteSubmitting = ref(false)
const copyResult = ref('')
const activeEditMemberId = ref<number | null>(null)
const editError = ref('')
const editingMember = ref(false)
const activeBindMemberId = ref<number | null>(null)
const selectedRequestByMember = reactive<Record<string, string>>({})
const bindingApplication = ref(false)
const relationshipActing = ref(false)
const relationshipManageError = ref('')
const relationshipNature = reactive<Record<string, string>>({})

const memberForm = reactive({
  name: '',
  gender: 'UNKNOWN' as Gender,
  birthYear: undefined as number | undefined,
  isAlive: true
})

const relationshipForm = reactive({
  baseMemberId: 0,
  addType: 'ADD_CHILD' as RelationshipAddType,
  name: '',
  gender: 'UNKNOWN' as Gender,
  parentLinkType: 'PRIMARY'
})
const placementForm = reactive({ memberId: 0, baseMemberId: 0, addType: 'ADD_CHILD' as RelationshipAddType })
const editForm = reactive({
  name: '',
  gender: 'UNKNOWN' as Gender,
  birthYear: undefined as number | undefined,
  isAlive: true,
  description: ''
})

const canManage = computed(() => family.value?.role === 'FOUNDER' || family.value?.role === 'FAMILY_ADMIN')
const unlocatedMembers = computed(() => members.value.filter((member) => !relatedMemberIds.value.has(member.memberId)))
const placementBaseMembers = computed(() => members.value.filter((member) => member.memberId !== placementForm.memberId))

function errorText(requestError: unknown, fallback: string) {
  if (axios.isAxiosError<ApiResponse<unknown>>(requestError)) {
    forbidden.value = requestError.response?.status === 403
    if (requestError.response?.data?.code) {
      return `${requestError.response.data.code}：${requestError.response.data.message}`
    }
  }
  return apiErrorMessage(requestError, fallback)
}

function genderLabel(gender: string) {
  return gender === 'MALE' ? '男' : gender === 'FEMALE' ? '女' : '未知'
}

function aliveLabel(member: FamilyMember) {
  if (member.isAlive === true) return '健在'
  if (member.isAlive === false) return '已故'
  return '未填写'
}

function roleText(role?: string | null) {
  if (role === 'FOUNDER') return '家庭创建者'
  if (role === 'FAMILY_ADMIN') return '家庭管理员'
  if (role === 'MEMBER') return '家庭成员'
  if (!role) return '未绑定身份'
  return '未知'
}

function memberStatusText(status?: string) {
  if (status === 'ACTIVE') return '正常'
  if (status === 'DELETED') return '已删除'
  return '未知'
}

function bindingText(member: FamilyMember) {
  if (member.boundUserId) return '已绑定账号'
  if (member.userBindingPolicy === 'NOT_REQUIRED') return '无需绑定'
  return '等待绑定'
}

function canInvite(member: FamilyMember) {
  return member.status === 'ACTIVE' &&
    !member.boundUserId &&
    member.userBindingPolicy !== 'NOT_REQUIRED' &&
    !pendingInvitation(member)
}

function pendingInvitation(member: FamilyMember) {
  return invitations.value.find((invitation) =>
    String(invitation.targetMemberId) === String(member.memberId) && invitation.status === 'PENDING'
  )
}

function matchingJoinRequests(member: FamilyMember) {
  return joinRequests.value.filter((request) =>
    request.requestStatus === 'PENDING' && request.applicantGender === member.gender
  )
}

function canBindApplication(member: FamilyMember) {
  return canInvite(member) && matchingJoinRequests(member).length > 0
}

function openInvite(memberId: number) {
  activeInviteMemberId.value = memberId
  inviteMessage.value = ''
  inviteLink.value = ''
  inviteError.value = ''
  copyResult.value = ''
}

function closeInvite() {
  activeInviteMemberId.value = null
  inviteMessage.value = ''
  inviteLink.value = ''
  inviteError.value = ''
  copyResult.value = ''
}

async function submitInvite(member: FamilyMember) {
  inviteSubmitting.value = true
  inviteError.value = ''
  copyResult.value = ''
  try {
    const result = await createInvitation(familyId.value, member.memberId, {
      inviteChannel: 'SHARE_LINK',
      inviteMessage: inviteMessage.value || undefined,
      familyRoleAfterAccept: 'MEMBER'
    })
    inviteLink.value = `${window.location.origin}/invite/${encodeURIComponent(result.inviteToken)}`
  } catch (requestError) {
    inviteError.value = errorText(requestError, '创建分享邀请失败。')
  } finally {
    inviteSubmitting.value = false
  }
}

async function copyInviteLink() {
  if (!inviteLink.value) return
  try {
    await navigator.clipboard.writeText(inviteLink.value)
    copyResult.value = '已复制'
  } catch {
    copyResult.value = '复制失败，请手动复制'
  }
}

async function loadPage() {
  pageLoading.value = true
  pageError.value = ''
  forbidden.value = false
  try {
    family.value = await getFamilyDetail(familyId.value)
    const [memberResult, treeResult, requestResult, invitationResult] = await Promise.all([
      listMembers(familyId.value),
      getPrivateTree(familyId.value),
      canManage.value ? listFamilyJoinRequests(familyId.value) : Promise.resolve([]),
      canManage.value ? listFamilyInvitations(familyId.value) : Promise.resolve([])
    ])
    members.value = memberResult
    joinRequests.value = requestResult
    invitations.value = invitationResult
    treeEdges.value = treeResult.edges
    relatedMemberIds.value = new Set(treeResult.edges.flatMap((edge) => [edge.fromMemberId, edge.toMemberId]))
    Object.keys(relationshipNature).forEach((key) => delete relationshipNature[key])
    treeResult.edges.forEach((edge) => {
      relationshipNature[String(edge.relationshipId)] = edge.parentLinkType || 'PRIMARY'
    })
    if (memberResult.length > 0) relationshipForm.baseMemberId = memberResult[0].memberId
    normalizePlacement()
  } catch (requestError) {
    pageError.value = errorText(requestError, '无法加载家庭成员页面。')
  } finally {
    pageLoading.value = false
  }
}

function normalizePlacement() {
  if (!unlocatedMembers.value.some((member) => member.memberId === placementForm.memberId)) {
    placementForm.memberId = unlocatedMembers.value[0]?.memberId || 0
  }
  if (!placementBaseMembers.value.some((member) => member.memberId === placementForm.baseMemberId)) {
    placementForm.baseMemberId = placementBaseMembers.value[0]?.memberId || 0
  }
}

async function submitPlacement() {
  normalizePlacement()
  if (!placementForm.memberId || !placementForm.baseMemberId) {
    placementError.value = '请选择待定位成员和基准成员。'
    return
  }
  placementSubmitting.value = true
  placementError.value = ''
  try {
    await placeExistingMember(familyId.value, {
      memberId: placementForm.memberId,
      baseMemberId: placementForm.baseMemberId,
      addType: placementForm.addType,
      relationship: { parentLinkType: placementForm.addType === 'ADD_SPOUSE' ? undefined : 'PRIMARY' }
    })
    await loadPage()
  } catch (error) {
    placementError.value = errorText(error, '成员定位失败。')
  } finally { placementSubmitting.value = false }
}

async function submitMember() {
  memberError.value = ''
  if (!memberForm.name) {
    memberError.value = '成员姓名不能为空。'
    return
  }
  memberSubmitting.value = true
  try {
    await createMember(familyId.value, {
      name: memberForm.name,
      gender: memberForm.gender,
      birthYear: memberForm.birthYear || undefined,
      isAlive: memberForm.isAlive
    })
    memberForm.name = ''
    memberForm.gender = 'UNKNOWN'
    memberForm.birthYear = undefined
    memberForm.isAlive = true
    await loadPage()
  } catch (requestError) {
    memberError.value = errorText(requestError, '新增成员失败。')
  } finally {
    memberSubmitting.value = false
  }
}

async function submitRelationship() {
  relationshipError.value = ''
  relationshipSuccess.value = ''
  if (!relationshipForm.baseMemberId) {
    relationshipError.value = '请选择基础成员。'
    return
  }
  if (!relationshipForm.name) {
    relationshipError.value = '新亲属姓名不能为空。'
    return
  }

  relationshipSubmitting.value = true
  try {
    const result = await createRelationship(familyId.value, {
      baseMemberId: relationshipForm.baseMemberId,
      addType: relationshipForm.addType,
      newMember: {
        name: relationshipForm.name,
        gender: relationshipForm.gender,
        isAlive: true
      },
      relationship: relationshipForm.addType === 'ADD_SPOUSE'
        ? { relationshipType: 'SPOUSE' }
        : {
            relationshipType: 'PARENT_CHILD',
            parentLinkType: relationshipForm.parentLinkType
          }
    })
    relationshipSuccess.value = `创建成功，家谱版本已更新为第 ${result.graphVersion} 版。`
    relationshipForm.name = ''
    relationshipForm.gender = 'UNKNOWN'
    await loadPage()
  } catch (requestError) {
    relationshipError.value = errorText(requestError, '创建亲属关系失败。')
  } finally {
    relationshipSubmitting.value = false
  }
}

function openApplicationBinding(member: FamilyMember) {
  activeBindMemberId.value = member.memberId
  selectedRequestByMember[String(member.memberId)] = String(matchingJoinRequests(member)[0]?.requestId || '')
  activeEditMemberId.value = null
  closeInvite()
}

async function bindApplication(member: FamilyMember) {
  const requestId = selectedRequestByMember[String(member.memberId)]
  if (!requestId) {
    pageError.value = '请选择一个待处理加入申请。'
    return
  }
  if (!window.confirm(`确认通过申请，并将账号绑定到“${member.name}”节点？`)) return
  bindingApplication.value = true
  pageError.value = ''
  try {
    await approveJoinRequest(familyId.value, requestId, {
      approveMode: 'BIND_EXISTING_MEMBER',
      memberId: member.memberId
    })
    activeBindMemberId.value = null
    await loadPage()
  } catch (error) {
    pageError.value = errorText(error, '绑定加入申请失败。')
  } finally {
    bindingApplication.value = false
  }
}

function openEdit(member: FamilyMember) {
  activeEditMemberId.value = member.memberId
  activeBindMemberId.value = null
  closeInvite()
  editError.value = ''
  editForm.name = member.name
  editForm.gender = member.gender as Gender
  editForm.birthYear = member.birthYear || undefined
  editForm.isAlive = member.isAlive !== false
  editForm.description = member.description || ''
}

function closeEdit() {
  activeEditMemberId.value = null
  editError.value = ''
}

async function saveMember(member: FamilyMember) {
  if (!editForm.name.trim()) {
    editError.value = '成员姓名不能为空。'
    return
  }
  editingMember.value = true
  editError.value = ''
  try {
    await updateMember(familyId.value, member.memberId, {
      name: editForm.name,
      gender: editForm.gender,
      birthYear: editForm.birthYear || undefined,
      isAlive: editForm.isAlive,
      description: editForm.description || undefined
    })
    closeEdit()
    await loadPage()
  } catch (error) {
    editError.value = errorText(error, '成员资料保存失败。')
  } finally {
    editingMember.value = false
  }
}

async function removeMember(member: FamilyMember) {
  if (!window.confirm(`确定删除“${member.name}”吗？存在下级成员、绑定账号或创建者身份时，后端会拒绝。`)) return
  editingMember.value = true
  editError.value = ''
  try {
    await deleteMember(familyId.value, member.memberId, 'PC 家庭成员管理删除')
    closeEdit()
    await loadPage()
  } catch (error) {
    editError.value = errorText(error, '成员删除失败。')
  } finally {
    editingMember.value = false
  }
}

function memberName(memberId: number) {
  return members.value.find((member) => member.memberId === memberId)?.name || '未知成员'
}

function relationshipSentence(edge: TreeEdge) {
  const from = memberName(edge.fromMemberId)
  const to = memberName(edge.toMemberId)
  return edge.relationshipType === 'SPOUSE' ? `${from} 与 ${to} 是配偶` : `${from} 是 ${to} 的父母`
}

function relationshipTypeText(type: string) {
  return type === 'SPOUSE' ? '配偶关系' : '父母子女关系'
}

async function saveRelationship(edge: TreeEdge) {
  relationshipActing.value = true
  relationshipManageError.value = ''
  try {
    await updateRelationship(familyId.value, edge.relationshipId, {
      parentLinkType: relationshipNature[String(edge.relationshipId)] || 'PRIMARY'
    })
    await loadPage()
  } catch (error) {
    relationshipManageError.value = errorText(error, '关系性质保存失败。')
  } finally {
    relationshipActing.value = false
  }
}

async function removeRelationship(edge: TreeEdge) {
  if (!window.confirm(`确定删除“${relationshipSentence(edge)}”吗？`)) return
  relationshipActing.value = true
  relationshipManageError.value = ''
  try {
    await deleteRelationship(familyId.value, edge.relationshipId, 'PC 家庭成员管理删除错误关系')
    await loadPage()
  } catch (error) {
    relationshipManageError.value = errorText(error, '关系删除失败。')
  } finally {
    relationshipActing.value = false
  }
}

onMounted(loadPage)
</script>

<style scoped>
.page-section {
  padding: 32px 0;
}

.page-heading,
.section-heading,
.heading-actions,
.member-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.page-heading {
  align-items: flex-end;
  margin-bottom: 18px;
}

h1,
h2,
p {
  margin-top: 0;
}

.eyebrow {
  color: var(--color-heritage-green);
  font-size: 13px;
  font-weight: 700;
}

.forms-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
}

.form-card {
  display: grid;
  align-content: start;
  gap: 15px;
  padding: 20px;
}

.compact-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

label {
  display: grid;
  gap: 7px;
  font-size: 14px;
  font-weight: 600;
}

.check-row {
  display: flex;
  align-items: center;
}

.notice,
.readonly-note {
  border-left: 3px solid var(--color-warm-gold);
  background: #fbf7ed;
  padding: 12px;
  color: var(--color-text-secondary);
  font-size: 13px;
  line-height: 1.6;
}

.readonly-note {
  margin-bottom: 18px;
}

.feedback {
  margin: 0;
  font-weight: 600;
}

.error {
  color: var(--color-danger);
}

.success {
  color: var(--color-success);
}

.state-panel {
  display: grid;
  min-height: 170px;
  place-items: center;
  gap: 12px;
  padding: 24px;
  text-align: center;
}

.state-panel.error {
  color: var(--color-danger);
}

.member-section {
  margin-top: 24px;
}

.member-list {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px;
}

.member-card {
  padding: 18px;
}

.invite-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-top: 16px;
}

.invite-form,
.invite-result {
  display: grid;
  gap: 12px;
  margin-top: 16px;
  border-top: 1px solid var(--color-border);
  padding-top: 16px;
}

.invite-buttons {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.relation-card,
.relation-controls {
  display: flex;
  align-items: center;
  gap: 12px;
}

.relation-card {
  justify-content: space-between;
}

.relation-controls {
  flex-wrap: wrap;
}

.status-tag {
  border-radius: 6px;
  background: #edf4ef;
  padding: 4px 8px;
  color: var(--color-success);
  font-size: 12px;
  font-weight: 700;
}

dl {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
  margin-bottom: 0;
}

dt {
  color: var(--color-text-secondary);
  font-size: 12px;
}

dd {
  margin: 4px 0 0;
  font-weight: 600;
}

.button:disabled {
  cursor: not-allowed;
  opacity: 0.6;
}

@media (max-width: 760px) {
  .page-heading,
  .heading-actions {
    align-items: stretch;
    flex-direction: column;
  }

  .forms-grid,
  .member-list,
  .compact-grid,
  dl {
    grid-template-columns: 1fr;
  }

  .relation-card {
    align-items: stretch;
    flex-direction: column;
  }
}
</style>
