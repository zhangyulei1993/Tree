<template>
  <view class="archive-page family-join-review-page">
    <MiniBackHome />

    <view v-if="familyName" class="review-head archive-page-head">
      <view>
        <text class="archive-kicker">Join Review</text>
        <text class="archive-title">收到的加入申请</text>
        <text class="archive-subtitle">{{ familyName }} · 审核外部用户提交给该家庭的加入申请</text>
      </view>
      <view class="archive-seal">{{ familySealLetter }}</view>
    </view>

    <view v-if="familyName" class="review-context">
      <text class="archive-chip">{{ familyRoleLabel }}</text>
      <text class="context-back" @click="openFamilyOverview">返回详情</text>
    </view>

    <view class="review-summary">
      <text v-if="requests.length" class="archive-chip">待处理 {{ pendingCount }}</text>
      <MiniButton variant="secondary" size="sm" :disabled="loading" :loading="loading" @click="loadRequests">
        刷新列表
      </MiniButton>
    </view>

    <MiniNotice tone="security" title="审核说明">
      仅家庭创建者和管理员可查看与处理。通过前请确认申请人在家庭树中的位置。
    </MiniNotice>

    <view v-if="loading && requests.length === 0" class="review-state archive-form-panel">
      <MiniEmptyState symbol="…" title="正在加载" description="正在加载加入申请..." />
    </view>

    <view v-else-if="loadError && requests.length === 0" class="review-state archive-form-panel">
      <MiniEmptyState
        symbol="!"
        title="加载失败"
        :description="loadError"
        action-text="重新加载"
        @action="loadRequests"
      />
    </view>

    <view v-else-if="requests.length === 0" class="review-state archive-form-panel">
      <MiniEmptyState
        symbol="申"
        title="暂无加入申请"
        description="有用户提交加入申请后，会在这里显示。"
      />
    </view>

    <template v-else>
      <view v-for="group in requestGroups" :key="group.key" class="review-group archive-form-panel">
        <view class="archive-section-head">
          <text class="archive-section-title">{{ group.title }}</text>
          <text class="archive-section-subtitle">{{ group.subtitle }}</text>
        </view>
        <view class="review-list archive-list">
          <view v-for="item in group.items" :key="item.requestId" class="review-item">
            <view class="review-item-head">
              <view class="archive-row-main">
                <text class="archive-row-title">{{ item.applicantRealName || '未填写姓名' }}</text>
                <text class="archive-row-desc">申请加入 {{ item.familyName || familyName || '当前家庭' }}</text>
              </view>
              <text class="archive-status-tag" :class="joinStatusClass(item.requestStatus)">
                {{ joinRequestStatusText(item.requestStatus) }}
              </text>
            </view>

            <view class="review-meta">
              <text>申请理由：{{ item.applicantMessage || '未填写' }}</text>
              <text>申请性别：{{ genderText(item.applicantGender) }}</text>
              <text class="review-meta-weak">提交时间：{{ formatDate(item.createdAt) }}</text>
              <text v-if="item.handleComment">处理结果：{{ item.handleComment }}</text>
              <text v-else-if="item.requestStatus !== 'PENDING'" class="review-meta-weak">
                更新时间：{{ formatDate(item.updatedAt) }}
              </text>
            </view>

            <view v-if="item.requestStatus === 'PENDING'" class="review-actions">
              <MiniButton
                size="sm"
                variant="secondary"
                :disabled="isActing(item)"
                @click="openResolvePanel(item, 'bind')"
              >
                绑定已有成员
              </MiniButton>
              <MiniButton size="sm" :disabled="isActing(item)" @click="openResolvePanel(item, 'locate')">
                创建并定位
              </MiniButton>
              <MiniButton
                size="sm"
                variant="secondary"
                :disabled="isActing(item)"
                @click="openResolvePanel(item, 'standalone')"
              >
                暂存未定位
              </MiniButton>
              <MiniButton
                size="sm"
                variant="secondary"
                :disabled="isActing(item)"
                :loading="actingId === item.requestId && actingType === 'reject'"
                @click="confirmReject(item)"
              >
                驳回
              </MiniButton>
            </view>

            <view
              v-if="activeRequestId === item.requestId && item.requestStatus === 'PENDING'"
              class="resolve-panel archive-form-panel"
            >
              <template v-if="resolveMode === 'bind'">
                <view class="archive-section-head">
                  <text class="archive-section-title">绑定到已有成员</text>
                  <text class="archive-section-subtitle">适合家庭树里已经有这个人，只是还没有绑定账号</text>
                </view>
                <picker
                  v-if="bindableMembers.length > 0"
                  mode="selector"
                  :range="bindableMemberLabels"
                  :value="selectedMemberIndex"
                  @change="onSelectExistingMember"
                >
                  <view class="field-picker">{{ selectedExistingMemberLabel }}</view>
                </picker>
                <MiniNotice v-else tone="warm">
                  当前没有可绑定成员。可先创建并定位，或暂存为未定位成员。
                </MiniNotice>
                <MiniButton
                  size="sm"
                  :disabled="bindableMembers.length === 0 || isActing(item)"
                  :loading="actingId === item.requestId && actingType === 'approve'"
                  @click="confirmBindExisting(item)"
                >
                  确认绑定并通过
                </MiniButton>
              </template>

              <template v-else-if="resolveMode === 'locate'">
                <view class="archive-section-head">
                  <text class="archive-section-title">创建新成员并放入家庭树</text>
                  <text class="archive-section-subtitle">适合知道申请人与某位已有成员的关系</text>
                </view>
                <text class="tree-field-label">新成员姓名</text>
                <input v-model.trim="resolveForm.name" class="tree-input" maxlength="100" placeholder="请输入成员姓名" />
                <text class="tree-field-label">新成员性别</text>
                <picker mode="selector" :range="genderLabels" :value="genderIndex" @change="onSelectGender">
                  <view class="field-picker">{{ genderLabels[genderIndex] }}</view>
                </picker>
                <text class="tree-field-label">基准成员</text>
                <picker
                  v-if="members.length > 0"
                  mode="selector"
                  :range="memberLabels"
                  :value="baseMemberIndex"
                  @change="onSelectBaseMember"
                >
                  <view class="field-picker">{{ selectedBaseMemberLabel }}</view>
                </picker>
                <MiniNotice v-else tone="warm">
                  当前家庭还没有可作为基准的成员，无法定位到家庭树。
                </MiniNotice>
                <text class="tree-field-label">与基准成员的关系</text>
                <picker mode="selector" :range="relationOptionLabels" :value="addTypeIndex" @change="onSelectAddType">
                  <view class="field-picker">{{ selectedAddTypeLabel }}</view>
                </picker>
                <text v-if="showParentRolePicker" class="tree-field-label">父母身份</text>
                <picker
                  v-if="showParentRolePicker"
                  mode="selector"
                  :range="parentRoleLabels"
                  :value="parentRoleIndex"
                  @change="onSelectParentRole"
                >
                  <view class="field-picker">{{ parentRoleLabels[parentRoleIndex] }}</view>
                </picker>
                <MiniNotice v-if="showParentRolePicker && selectedParentMemberType === 'SPOUSE'" tone="warm">
                  选择「本家成员的配偶」时，基准成员必须已有相反性别的本家成员父母。
                </MiniNotice>
                <MiniNotice :tone="placementTone">
                  {{ selectedPlacement.message }}
                </MiniNotice>
                <MiniButton
                  size="sm"
                  :disabled="orderedMembers.length === 0 || isActing(item)"
                  :loading="actingId === item.requestId && actingType === 'approve'"
                  @click="confirmCreateLocated(item)"
                >
                  创建并通过
                </MiniButton>
              </template>

              <template v-else>
                <view class="archive-section-head">
                  <text class="archive-section-title">暂存为未定位成员</text>
                  <text class="archive-section-subtitle">
                    只创建成员档案并绑定账号，后续需要在成员管理中补充亲属关系
                  </text>
                </view>
                <text class="tree-field-label">新成员姓名</text>
                <input v-model.trim="resolveForm.name" class="tree-input" maxlength="100" placeholder="请输入成员姓名" />
                <text class="tree-field-label">新成员性别</text>
                <picker mode="selector" :range="genderLabels" :value="genderIndex" @change="onSelectGender">
                  <view class="field-picker">{{ genderLabels[genderIndex] }}</view>
                </picker>
                <MiniButton
                  size="sm"
                  :disabled="isActing(item)"
                  :loading="actingId === item.requestId && actingType === 'approve'"
                  @click="confirmCreateStandalone(item)"
                >
                  暂存并通过
                </MiniButton>
              </template>
            </view>
          </view>
        </view>
      </view>
    </template>

    <text v-if="operationError" class="tree-field-error review-error">{{ operationError }}</text>
  </view>
</template>

<script setup lang="ts">
import { onHide, onLoad, onShow, onUnload } from '@dcloudio/uni-app'
import { computed, reactive, ref } from 'vue'

import { apiErrorMessage } from '@/api/client'
import { getFamilyDetail } from '@/api/families'
import {
  approveJoinRequest,
  listFamilyJoinRequests,
  rejectJoinRequest
} from '@/api/joinRequests'
import { listFamilyMembers } from '@/api/members'
import { getPrivateTree } from '@/api/tree'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'
import { joinRequestStatusText } from '@/components/base/formatStatus'
import MiniNotice from '@/components/base/MiniNotice.vue'
import {
  buildApproveJoinLocation,
  parentRoleLabels,
  parentRoleValues
} from '@/features/join/approveJoinLocation'
import { useSessionStore } from '@/stores/session'
import type { FamilyMember, Gender, JoinRequest, RelationshipAddType, TreeEdge, TreeNode } from '@/types/api'
import { normalizeText, validateTextFields } from '@/utils/inputValidation'

interface PlacementState {
  blocked: boolean
  message: string
  parentLinkType?: string
  relationNoteType?: string
  relationNote?: string
}

interface RelationOption {
  addType: RelationshipAddType
  label: string
}

const session = useSessionStore()
const familyId = ref('')
const familyName = ref('')
const familySurname = ref('')
const familyRole = ref('')
const requests = ref<JoinRequest[]>([])
const members = ref<FamilyMember[]>([])
const treeNodes = ref<TreeNode[]>([])
const treeEdges = ref<TreeEdge[]>([])
const loading = ref(false)
const loadError = ref('')
const operationError = ref('')
const actingId = ref<number | string | null>(null)
const actingType = ref<'' | 'approve' | 'reject'>('')
const activeRequestId = ref<number | string | null>(null)
const resolveMode = ref<'bind' | 'locate' | 'standalone'>('bind')
const selectedMemberIndex = ref(0)
const baseMemberIndex = ref(0)
const genderIndex = ref(0)
const addTypeIndex = ref(1)
const parentRoleIndex = ref(0)
const resolveForm = reactive({
  name: '',
  gender: 'MALE' as Gender
})

const genders: Gender[] = ['MALE', 'FEMALE']
const genderLabels = ['男', '女']
const relationFallback: RelationshipAddType = 'ADD_CHILD'

const bindableMembers = computed(() =>
  members.value.filter((member) =>
    member.status === 'ACTIVE'
    && !member.boundUserId
    && member.userBindingPolicy !== 'NOT_REQUIRED'
  )
)
const bindableMemberLabels = computed(() => bindableMembers.value.map(memberLabel))
const orderedMembers = computed(() =>
  [...members.value].sort((left, right) => left.name.localeCompare(right.name, 'zh-CN'))
)
const memberLabels = computed(() => orderedMembers.value.map(memberLabel))
const selectedExistingMemberLabel = computed(() =>
  bindableMemberLabels.value[selectedMemberIndex.value] || '请选择成员'
)
const selectedBaseMemberLabel = computed(() =>
  memberLabels.value[baseMemberIndex.value] || '请选择基准成员'
)
const selectedBaseMember = computed(() => orderedMembers.value[baseMemberIndex.value])
const relationOptions = computed(() => buildRelationOptions(resolveForm.gender, selectedBaseMember.value))
const relationOptionLabels = computed(() => relationOptions.value.map((option) => option.label))
const selectedRelationOption = computed(() => relationOptions.value[addTypeIndex.value] || relationOptions.value[0])
const selectedAddType = computed(() => selectedRelationOption.value?.addType || relationFallback)
const selectedAddTypeLabel = computed(() => selectedRelationOption.value?.label || '关系')
const showParentRolePicker = computed(() => {
  const addType = selectedAddType.value
  return addType === 'ADD_FATHER' || addType === 'ADD_MOTHER'
})
const selectedParentMemberType = computed(() => parentRoleValues[parentRoleIndex.value] || 'LINEAGE_MEMBER')
const familyRoleLabel = computed(() => familyRole.value === 'FOUNDER' ? '创建者' : '管理员')
const familySealLetter = computed(() => familySurname.value.slice(0, 1) || familyName.value.slice(0, 1) || '审')
const pendingCount = computed(() =>
  requests.value.filter((item) => item.requestStatus === 'PENDING').length
)
const treeNodeMap = computed(() => new Map(treeNodes.value.map((node) => [Number(node.memberId), node])))
const selectedPlacement = computed(() => placementState(selectedBaseMember.value, selectedAddType.value))
const placementTone = computed(() => selectedPlacement.value.blocked || selectedPlacement.value.relationNoteType ? 'warm' : 'security')
const requestGroups = computed(() => {
  const pending = requests.value.filter((item) => item.requestStatus === 'PENDING')
  const history = requests.value.filter((item) => item.requestStatus !== 'PENDING')
  return [
    { key: 'pending', title: '待我审核', subtitle: '需要确认申请人身份和家庭树位置', items: pending },
    { key: 'history', title: '历史申请', subtitle: '已通过或驳回的处理记录', items: history }
  ].filter((group) => group.items.length > 0)
})

function joinStatusClass(status: string) {
  if (status === 'PENDING') return 'is-cinnabar'
  if (status === 'APPROVED') return 'is-ink'
  return 'is-muted'
}

function resetTransientUI() {
  operationError.value = ''
  actingId.value = null
  actingType.value = ''
  activeRequestId.value = null
}

function resetPageData() {
  loading.value = false
  loadError.value = ''
  requests.value = []
  members.value = []
  treeNodes.value = []
  treeEdges.value = []
  familySurname.value = ''
  familyName.value = ''
  familyRole.value = ''
  resetTransientUI()
}

function currentRoute() {
  return `/pages/join/family?familyId=${encodeURIComponent(familyId.value)}`
}

function formatDate(value: string) {
  const time = new Date(value)
  return Number.isNaN(time.getTime()) ? value : time.toLocaleString('zh-CN')
}

async function loadRequests() {
  session.restoreSession()
  if (!familyId.value) {
    loadError.value = '缺少家庭信息。'
    return
  }
  if (!session.isLoggedIn) {
    resetPageData()
    session.requireLogin(currentRoute())
    return
  }
  const isInitialLoad = requests.value.length === 0
  loading.value = isInitialLoad
  loadError.value = ''
  operationError.value = ''
  try {
    const [family, requestRows, memberRows] = await Promise.all([
      getFamilyDetail(familyId.value),
      listFamilyJoinRequests(familyId.value),
      listFamilyMembers(familyId.value)
    ])
    const tree = await getPrivateTree(familyId.value)
    familySurname.value = family.familySurname || ''
    familyName.value = family.familyName || '当前家庭'
    familyRole.value = family.role
    requests.value = requestRows
    members.value = memberRows
    treeNodes.value = tree.nodes
    treeEdges.value = tree.edges
    normalizePickerIndexes()
    normalizeRelationIndex()
  } catch (error) {
    loadError.value = apiErrorMessage(error, '加入申请加载失败。')
  } finally {
    loading.value = false
  }
}

function openFamilyOverview() {
  uni.redirectTo({ url: `/pages/family/detail?familyId=${encodeURIComponent(familyId.value)}` })
}

function memberLabel(member: FamilyMember) {
  const gender = member.gender === 'MALE' ? '男' : member.gender === 'FEMALE' ? '女' : '性别未知'
  const binding = member.boundUserId ? '已绑定' : member.userBindingPolicy === 'NOT_REQUIRED' ? '无需绑定' : '未绑定'
  return `${member.name}（${gender} · ${binding}）`
}

function normalizePickerIndexes() {
  if (selectedMemberIndex.value >= bindableMembers.value.length) selectedMemberIndex.value = 0
  if (baseMemberIndex.value >= orderedMembers.value.length) baseMemberIndex.value = 0
}

function isActing(item: JoinRequest) {
  return actingId.value === item.requestId
}

function defaultApplicantName(item: JoinRequest) {
  return item.applicantRealName?.trim() || ''
}

function defaultApplicantGender(item: JoinRequest): Gender {
  return item.applicantGender === 'FEMALE' ? 'FEMALE' : 'MALE'
}

function genderText(value?: string | null) {
  if (value === 'MALE') return '男'
  if (value === 'FEMALE') return '女'
  return '未填写'
}

function openResolvePanel(item: JoinRequest, mode: 'bind' | 'locate' | 'standalone') {
  activeRequestId.value = item.requestId
  resolveMode.value = mode
  operationError.value = ''
  resolveForm.name = defaultApplicantName(item)
  resolveForm.gender = defaultApplicantGender(item)
  genderIndex.value = genders.indexOf(resolveForm.gender)
  if (genderIndex.value < 0) genderIndex.value = 0
  addTypeIndex.value = 1
  parentRoleIndex.value = 0
  normalizePickerIndexes()
  normalizeRelationIndex()
}

function onSelectExistingMember(event: { detail: { value: number | string } }) {
  selectedMemberIndex.value = Number(event.detail.value) || 0
}

function onSelectBaseMember(event: { detail: { value: number | string } }) {
  baseMemberIndex.value = Number(event.detail.value) || 0
  normalizeRelationIndex()
}

function onSelectGender(event: { detail: { value: number | string } }) {
  genderIndex.value = Number(event.detail.value) || 0
  resolveForm.gender = genders[genderIndex.value] || 'MALE'
  normalizeRelationIndex()
}

function onSelectAddType(event: { detail: { value: number | string } }) {
  addTypeIndex.value = Number(event.detail.value) || 0
  parentRoleIndex.value = 0
  normalizeRelationIndex()
}

function onSelectParentRole(event: { detail: { value: number | string } }) {
  parentRoleIndex.value = Number(event.detail.value) || 0
}

function normalizeRelationIndex() {
  if (addTypeIndex.value >= relationOptions.value.length) addTypeIndex.value = 0
  if (addTypeIndex.value < 0) addTypeIndex.value = 0
}

function buildRelationOptions(gender: Gender, base?: FamilyMember): RelationOption[] {
  const options: RelationOption[] = []
  if (gender === 'MALE') {
    options.push({ addType: 'ADD_FATHER', label: '父亲' })
    options.push({ addType: 'ADD_CHILD', label: '儿子' })
    options.push({ addType: 'ADD_SIBLING', label: '兄弟' })
    if (base?.gender === 'FEMALE') options.push({ addType: 'ADD_SPOUSE', label: '丈夫' })
  } else if (gender === 'FEMALE') {
    options.push({ addType: 'ADD_MOTHER', label: '母亲' })
    options.push({ addType: 'ADD_CHILD', label: '女儿' })
    options.push({ addType: 'ADD_SIBLING', label: '姐妹' })
    if (base?.gender === 'MALE') options.push({ addType: 'ADD_SPOUSE', label: '妻子' })
  }
  return options.length > 0 ? options : [{ addType: relationFallback, label: '子女' }]
}

function placementState(base: FamilyMember | undefined, addType: RelationshipAddType): PlacementState {
  if (!base) return { blocked: true, message: '请先选择基准成员。' }
  if (addType === 'ADD_FATHER') return parentPlacement(base, 'MALE', '父亲', '继父', 'STEP_FATHER')
  if (addType === 'ADD_MOTHER') return parentPlacement(base, 'FEMALE', '母亲', '继母', 'STEP_MOTHER')
  if (addType === 'ADD_SPOUSE') return spousePlacement(base)
  if (addType === 'ADD_CHILD') {
    if (base.gender !== 'MALE' && base.gender !== 'FEMALE') {
      return { blocked: true, message: '请先完善基准成员性别，再添加子女。' }
    }
    return { blocked: false, message: '将创建子女关系。', parentLinkType: 'PRIMARY' }
  }
  if (addType === 'ADD_SIBLING') {
    const parents = parentNodes(base.memberId)
    if (parents.length === 0) {
      return { blocked: true, message: '该基准成员还没有父亲或母亲节点，请先创建父亲或母亲节点。' }
    }
    return { blocked: false, message: '将沿用基准成员已有父母，创建兄弟姐妹关系。', parentLinkType: 'PRIMARY' }
  }
  return { blocked: false, message: '' }
}

function parentPlacement(base: FamilyMember, gender: Gender, label: string, stepLabel: string, noteType: string) {
  const existing = parentNodes(base.memberId).filter((node) => node.gender === gender)
  if (existing.some(isLineageNode)) {
    return { blocked: true, message: `该成员已有本家族${label}，不能重复添加。` }
  }
  if (existing.length > 0) {
    return {
      blocked: false,
      message: `该成员已有配偶类${label}，继续创建将按${stepLabel}处理。`,
      parentLinkType: 'STEP',
      relationNoteType: noteType,
      relationNote: stepLabel
    }
  }
  return { blocked: false, message: `将创建${label}关系。`, parentLinkType: 'PRIMARY' }
}

function spousePlacement(base: FamilyMember) {
  if (base.gender !== 'MALE' && base.gender !== 'FEMALE') {
    return { blocked: true, message: '请先完善基准成员性别，再添加配偶。' }
  }
  const spouseGender: Gender = base.gender === 'MALE' ? 'FEMALE' : 'MALE'
  const label = spouseGender === 'MALE' ? '丈夫' : '妻子'
  const secondLabel = spouseGender === 'MALE' ? '再婚丈夫' : '再婚妻子'
  const noteType = spouseGender === 'MALE' ? 'SECOND_HUSBAND' : 'SECOND_WIFE'
  const existing = spouseNodes(base.memberId).filter((node) => node.gender === spouseGender)
  if (existing.some(isLineageNode)) {
    return { blocked: true, message: `该成员已有本家族${label}，不能重复添加。` }
  }
  if (existing.length > 0) {
    return {
      blocked: false,
      message: `该成员已有非本家族${label}，继续创建会将原${label}备注为${spouseGender === 'MALE' ? '前夫' : '前妻'}，新成员备注为${secondLabel}。`,
      relationNoteType: noteType,
      relationNote: secondLabel
    }
  }
  return { blocked: false, message: `将创建${label}关系。` }
}

function parentNodes(memberId: number) {
  return treeEdges.value
    .filter((edge) => edge.relationshipType === 'PARENT_CHILD' && Number(edge.toMemberId) === Number(memberId))
    .map((edge) => treeNodeMap.value.get(Number(edge.fromMemberId)))
    .filter((node): node is TreeNode => Boolean(node))
}

function spouseNodes(memberId: number) {
  return treeEdges.value
    .filter((edge) =>
      edge.relationshipType === 'SPOUSE'
      && (Number(edge.fromMemberId) === Number(memberId) || Number(edge.toMemberId) === Number(memberId))
    )
    .map((edge) => treeNodeMap.value.get(Number(edge.fromMemberId) === Number(memberId) ? Number(edge.toMemberId) : Number(edge.fromMemberId)))
    .filter((node): node is TreeNode => Boolean(node))
}

function isLineageNode(node: TreeNode) {
  if (node.memberType === 'SPOUSE') return false
  const expected = familySurname.value.trim()
  if (!expected) return true
  if (node.surname && node.surname.trim()) return node.surname.trim() === expected
  return node.displayName.trim().startsWith(expected)
}

function ensureResolveName() {
  const value = resolveForm.name.trim()
  if (!value) {
    operationError.value = '请先填写新成员姓名。'
    return ''
  }
  const validationMessage = validateTextFields([
    { value, label: '新成员姓名', kind: 'name', required: true, maxLength: 100 }
  ])
  if (validationMessage) {
    operationError.value = validationMessage
    return ''
  }
  return normalizeText(value)
}

function confirmBindExisting(item: JoinRequest) {
  const member = bindableMembers.value[selectedMemberIndex.value]
  if (!member) {
    operationError.value = '请选择一个未绑定的已有成员。'
    return
  }
  const genderWarning = item.applicantGender && member.gender !== item.applicantGender
    ? `\n\n注意：申请性别为${genderText(item.applicantGender)}，所选成员性别为${genderText(member.gender)}，请确认是否仍要绑定。`
    : ''
  uni.showModal({
    title: '绑定已有成员',
    content: `确定把申请人绑定到「${member.name}」并通过申请吗？${genderWarning}`,
    success: (result) => {
      if (result.confirm) approveWithMember(item, member.memberId, '绑定已有成员通过申请')
    }
  })
}

function confirmCreateLocated(item: JoinRequest) {
  const name = ensureResolveName()
  if (!name) return
  const baseMember = members.value[baseMemberIndex.value]
  if (!baseMember) {
    operationError.value = '请选择基准成员。'
    return
  }
  if (selectedPlacement.value.blocked) {
    operationError.value = selectedPlacement.value.message
    uni.showToast({ title: selectedPlacement.value.message, icon: 'none' })
    return
  }
  const warning = selectedPlacement.value.relationNoteType ? `\n\n${selectedPlacement.value.message}` : ''
  uni.showModal({
    title: '创建并定位',
    content: `确定创建「${name}」，并设置为「${baseMember.name}」的${selectedAddTypeLabel.value}后通过申请吗？${warning}`,
    success: (result) => {
      if (result.confirm) createLocatedAndApprove(item, name, baseMember.memberId)
    }
  })
}

function confirmCreateStandalone(item: JoinRequest) {
  const name = ensureResolveName()
  if (!name) return
  uni.showModal({
    title: '暂存未定位成员',
    content: `确定创建「${name}」为未定位成员，并通过申请吗？后续需要补充亲属关系。`,
    success: (result) => {
      if (result.confirm) createStandaloneAndApprove(item, name)
    }
  })
}

function confirmReject(item: JoinRequest) {
  uni.showModal({
    title: '驳回申请',
    content: `确定驳回「${item.applicantRealName || '该用户'}」的加入申请吗？`,
    success: (result) => {
      if (result.confirm) reject(item)
    }
  })
}

async function approveWithMember(item: JoinRequest, memberId: number | string, comment: string) {
  actingId.value = item.requestId
  actingType.value = 'approve'
  operationError.value = ''
  try {
    await approveJoinRequest(familyId.value, item.requestId, {
      approveMode: 'BIND_EXISTING_MEMBER',
      memberId,
      handleComment: comment
    })
    await loadRequests()
    uni.showToast({ title: '已通过', icon: 'success' })
  } catch (error) {
    operationError.value = apiErrorMessage(error, '通过加入申请失败。')
  } finally {
    actingId.value = null
    actingType.value = ''
  }
}

async function createLocatedAndApprove(item: JoinRequest, name: string, baseMemberId: number) {
  actingId.value = item.requestId
  actingType.value = 'approve'
  operationError.value = ''
  try {
    const addType = selectedAddType.value
    const placement = selectedPlacement.value
    await approveJoinRequest(familyId.value, item.requestId, {
      approveMode: 'CREATE_NEW_MEMBER',
      newMember: {
        name,
        gender: resolveForm.gender,
        isAlive: true,
        userBindingPolicy: 'REQUIRED'
      },
      location: buildApproveJoinLocation({
        baseMemberId,
        addType,
        parentMemberType: selectedParentMemberType.value,
        placement: {
          parentLinkType: placement.parentLinkType,
          relationNoteType: placement.relationNoteType,
          relationNote: placement.relationNote
        }
      }),
      handleComment: '小程序创建并定位成员后通过申请'
    })
    await loadRequests()
    uni.showToast({ title: '已通过', icon: 'success' })
  } catch (error) {
    operationError.value = apiErrorMessage(error, '创建并通过申请失败。')
  } finally {
    actingId.value = null
    actingType.value = ''
  }
}

async function createStandaloneAndApprove(item: JoinRequest, name: string) {
  actingId.value = item.requestId
  actingType.value = 'approve'
  operationError.value = ''
  try {
    await approveJoinRequest(familyId.value, item.requestId, {
      approveMode: 'CREATE_NEW_MEMBER',
      newMember: {
        name,
        gender: resolveForm.gender,
        isAlive: true,
        userBindingPolicy: 'REQUIRED'
      },
      handleComment: '小程序暂存未定位成员后通过申请'
    })
    await loadRequests()
    uni.showToast({ title: '已通过', icon: 'success' })
  } catch (error) {
    operationError.value = apiErrorMessage(error, '暂存并通过申请失败。')
  } finally {
    actingId.value = null
    actingType.value = ''
  }
}

async function reject(item: JoinRequest) {
  actingId.value = item.requestId
  actingType.value = 'reject'
  operationError.value = ''
  try {
    await rejectJoinRequest(familyId.value, item.requestId, {
      handleComment: '小程序驳回加入申请'
    })
    await loadRequests()
    uni.showToast({ title: '已驳回', icon: 'none' })
  } catch (error) {
    operationError.value = apiErrorMessage(error, '驳回加入申请失败。')
  } finally {
    actingId.value = null
    actingType.value = ''
  }
}

onLoad((options) => {
  familyId.value = String(options?.familyId || '')
})

onShow(loadRequests)
onHide(resetTransientUI)
onUnload(resetPageData)
</script>

<style scoped>
.family-join-review-page {
  padding-top: 28rpx;
}

.family-join-review-page :deep(.mini-back-home) {
  margin-bottom: 22rpx;
}

.review-context {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16rpx;
  margin-bottom: 22rpx;
}

.context-back {
  color: var(--archive-cinnabar);
  font-size: 22rpx;
  line-height: 1.4;
}

.review-summary {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 10rpx;
  margin-bottom: 22rpx;
}

.family-join-review-page :deep(.mini-notice) {
  margin-bottom: 20rpx;
}

.review-group {
  margin-bottom: 24rpx;
}

.review-list {
  margin-top: 8rpx;
}

.review-item {
  border-bottom: 1rpx solid var(--archive-line);
  padding: 18rpx 0;
}

.review-item:last-child {
  border-bottom: 0;
}

.review-item-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16rpx;
}

.archive-status-tag {
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
  min-height: 42rpx;
  border: 1rpx solid var(--archive-line);
  padding: 0 12rpx;
  font-size: 20rpx;
  font-weight: 650;
  line-height: 1.2;
}

.archive-status-tag.is-cinnabar {
  border-color: rgba(168, 59, 45, 0.28);
  background: rgba(168, 59, 45, 0.08);
  color: var(--archive-cinnabar);
}

.archive-status-tag.is-ink {
  border-color: rgba(22, 51, 83, 0.22);
  background: rgba(22, 51, 83, 0.08);
  color: var(--archive-blue);
}

.archive-status-tag.is-muted {
  background: rgba(255, 248, 234, 0.58);
  color: var(--archive-ink-soft);
}

.review-meta {
  display: flex;
  flex-direction: column;
  gap: 6rpx;
  margin-top: 12rpx;
  color: var(--archive-ink-soft);
  font-size: 22rpx;
  line-height: 1.55;
}

.review-meta-weak {
  opacity: 0.88;
}

.review-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10rpx;
  margin-top: 16rpx;
}

.resolve-panel {
  margin-top: 18rpx;
  padding-top: 18rpx;
  border-top: 1rpx dashed var(--archive-line);
}

.resolve-panel :deep(.mini-notice) {
  margin: 12rpx 0;
}

.resolve-panel .archive-section-head {
  margin-bottom: 16rpx;
}

.review-error {
  display: block;
  margin-top: 16rpx;
}
</style>
