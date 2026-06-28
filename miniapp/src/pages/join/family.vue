<template>
  <view class="tree-page">
    <MiniBackHome />
    <MiniSectionHeader title="加入申请" subtitle="查看申请加入该家庭的记录。" />

    <MiniCard>
      <MiniNotice tone="security">
        仅家庭创建者和管理员可查看与处理。通过前请确认申请人在家谱中的位置。
      </MiniNotice>
      <MiniButton variant="secondary" size="sm" :disabled="loading" :loading="loading" @click="loadRequests">
        刷新列表
      </MiniButton>
      <text v-if="operationError" class="tree-field-error">{{ operationError }}</text>
    </MiniCard>

    <MiniCard v-if="!authChecked">
      <view class="state-block">
        <text class="tree-muted">正在确认登录状态...</text>
      </view>
    </MiniCard>

    <MiniCard v-else-if="loading">
      <view class="state-block">
        <text class="tree-muted">正在加载加入申请...</text>
      </view>
    </MiniCard>

    <MiniCard v-else-if="loadError">
      <MiniEmptyState
        symbol="!"
        title="加载失败"
        :description="loadError"
        action-text="重新加载"
        @action="loadRequests"
      />
    </MiniCard>

    <MiniCard v-else-if="requests.length === 0">
      <MiniEmptyState
        symbol="申"
        title="暂无加入申请"
        description="有用户从公开家庭主页提交申请后，会在这里显示。"
      />
    </MiniCard>

    <template v-else>
      <MiniCard v-for="item in requests" :key="item.requestId" variant="soft">
        <view class="item-head">
          <view class="item-title-block">
            <text class="item-name">{{ item.applicantRealName || '未填写姓名' }}</text>
            <text class="tree-weak">申请加入 {{ item.familyName || '当前家庭' }}</text>
          </view>
          <MiniStatusTag :status="item.requestStatus" :label="joinRequestStatusText(item.requestStatus)" />
        </view>

        <text class="tree-muted item-meta">申请性别：{{ genderText(item.applicantGender) }}</text>
        <text class="tree-muted item-meta">申请说明：{{ item.applicantMessage || '未填写' }}</text>
        <text v-if="item.handleComment" class="tree-muted item-meta">处理意见：{{ item.handleComment }}</text>
        <text class="tree-weak item-meta">提交时间：{{ formatDate(item.createdAt) }}</text>
        <text v-if="item.requestStatus !== 'PENDING'" class="tree-weak item-meta">
          更新时间：{{ formatDate(item.updatedAt) }}
        </text>

        <view v-if="item.requestStatus === 'PENDING'" class="item-actions">
          <MiniButton
            variant="secondary"
            :disabled="isActing(item)"
            @click="openResolvePanel(item, 'bind')"
          >
            绑定已有成员
          </MiniButton>
          <MiniButton
            :disabled="isActing(item)"
            @click="openResolvePanel(item, 'locate')"
          >
            创建并定位
          </MiniButton>
          <MiniButton
            variant="secondary"
            :disabled="isActing(item)"
            @click="openResolvePanel(item, 'standalone')"
          >
            暂存未定位
          </MiniButton>
          <MiniButton
            variant="secondary"
            :disabled="isActing(item)"
            :loading="actingId === item.requestId && actingType === 'reject'"
            @click="confirmReject(item)"
          >
            驳回申请
          </MiniButton>
        </view>

        <view v-if="activeRequestId === item.requestId && item.requestStatus === 'PENDING'" class="resolve-panel">
          <template v-if="resolveMode === 'bind'">
            <text class="panel-title">绑定到已有成员</text>
            <text class="tree-muted panel-desc">
              适合家谱里已经有这个人，只是还没有绑定账号。
            </text>
            <picker
              v-if="bindableMembers.length > 0"
              mode="selector"
              :range="bindableMemberLabels"
              :value="selectedMemberIndex"
              @change="onSelectExistingMember"
            >
              <view class="picker-field">{{ selectedExistingMemberLabel }}</view>
            </picker>
            <MiniNotice v-else tone="warm">
              当前没有可绑定成员。可先创建并定位，或暂存为未定位成员。
            </MiniNotice>
            <MiniButton
              :disabled="bindableMembers.length === 0 || isActing(item)"
              :loading="actingId === item.requestId && actingType === 'approve'"
              @click="confirmBindExisting(item)"
            >
              确认绑定并通过
            </MiniButton>
          </template>

          <template v-else-if="resolveMode === 'locate'">
            <text class="panel-title">创建新成员并放入家谱</text>
            <text class="tree-muted panel-desc">
              适合知道申请人与某位已有成员的关系。
            </text>
            <text class="tree-field-label">新成员姓名</text>
            <input v-model.trim="resolveForm.name" class="input" maxlength="100" placeholder="请输入成员姓名" />
            <text class="tree-field-label">新成员性别</text>
            <picker mode="selector" :range="genderLabels" :value="genderIndex" @change="onSelectGender">
              <view class="picker-field">{{ genderLabels[genderIndex] }}</view>
            </picker>
            <text class="tree-field-label">基准成员</text>
            <picker
              v-if="members.length > 0"
              mode="selector"
              :range="memberLabels"
              :value="baseMemberIndex"
              @change="onSelectBaseMember"
            >
              <view class="picker-field">{{ selectedBaseMemberLabel }}</view>
            </picker>
            <MiniNotice v-else tone="warm">
              当前家庭还没有可作为基准的成员，无法定位到家谱。
            </MiniNotice>
            <text class="tree-field-label">与基准成员的关系</text>
            <picker mode="selector" :range="relationOptionLabels" :value="addTypeIndex" @change="onSelectAddType">
              <view class="picker-field">{{ selectedAddTypeLabel }}</view>
            </picker>
            <MiniNotice :tone="placementTone">
              {{ selectedPlacement.message }}
            </MiniNotice>
            <MiniButton
              :disabled="members.length === 0 || isActing(item)"
              :loading="actingId === item.requestId && actingType === 'approve'"
              @click="confirmCreateLocated(item)"
            >
              创建并通过
            </MiniButton>
          </template>

          <template v-else>
            <text class="panel-title">暂存为未定位成员</text>
            <text class="tree-muted panel-desc">
              只创建成员档案并绑定账号，后续需要在成员管理中补充亲属关系。
            </text>
            <text class="tree-field-label">新成员姓名</text>
            <input v-model.trim="resolveForm.name" class="input" maxlength="100" placeholder="请输入成员姓名" />
            <text class="tree-field-label">新成员性别</text>
            <picker mode="selector" :range="genderLabels" :value="genderIndex" @change="onSelectGender">
              <view class="picker-field">{{ genderLabels[genderIndex] }}</view>
            </picker>
            <MiniButton
              :disabled="isActing(item)"
              :loading="actingId === item.requestId && actingType === 'approve'"
              @click="confirmCreateStandalone(item)"
            >
              暂存并通过
            </MiniButton>
          </template>
        </view>
      </MiniCard>
    </template>
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
import MiniCard from '@/components/base/MiniCard.vue'
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'
import { joinRequestStatusText } from '@/components/base/formatStatus'
import MiniNotice from '@/components/base/MiniNotice.vue'
import MiniSectionHeader from '@/components/base/MiniSectionHeader.vue'
import MiniStatusTag from '@/components/base/MiniStatusTag.vue'
import { useSessionStore } from '@/stores/session'
import type { FamilyMember, Gender, JoinRequest, RelationshipAddType, TreeEdge, TreeNode } from '@/types/api'

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
const familySurname = ref('')
const requests = ref<JoinRequest[]>([])
const members = ref<FamilyMember[]>([])
const treeNodes = ref<TreeNode[]>([])
const treeEdges = ref<TreeEdge[]>([])
const loading = ref(false)
const loadError = ref('')
const operationError = ref('')
const authChecked = ref(false)
const actingId = ref<number | string | null>(null)
const actingType = ref<'' | 'approve' | 'reject'>('')
const activeRequestId = ref<number | string | null>(null)
const resolveMode = ref<'bind' | 'locate' | 'standalone'>('bind')
const selectedMemberIndex = ref(0)
const baseMemberIndex = ref(0)
const genderIndex = ref(0)
const addTypeIndex = ref(1)
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
const memberLabels = computed(() => members.value.map(memberLabel))
const selectedExistingMemberLabel = computed(() =>
  bindableMemberLabels.value[selectedMemberIndex.value] || '请选择成员'
)
const selectedBaseMemberLabel = computed(() =>
  memberLabels.value[baseMemberIndex.value] || '请选择基准成员'
)
const selectedBaseMember = computed(() => members.value[baseMemberIndex.value])
const relationOptions = computed(() => buildRelationOptions(resolveForm.gender, selectedBaseMember.value))
const relationOptionLabels = computed(() => relationOptions.value.map((option) => option.label))
const selectedRelationOption = computed(() => relationOptions.value[addTypeIndex.value] || relationOptions.value[0])
const selectedAddType = computed(() => selectedRelationOption.value?.addType || relationFallback)
const selectedAddTypeLabel = computed(() => selectedRelationOption.value?.label || '关系')
const treeNodeMap = computed(() => new Map(treeNodes.value.map((node) => [Number(node.memberId), node])))
const selectedPlacement = computed(() => placementState(selectedBaseMember.value, selectedAddType.value))
const placementTone = computed(() => selectedPlacement.value.blocked || selectedPlacement.value.relationNoteType ? 'warm' : 'security')

function resetAuthView() {
  authChecked.value = false
  loading.value = false
  loadError.value = ''
  operationError.value = ''
  requests.value = []
  members.value = []
  treeNodes.value = []
  treeEdges.value = []
  familySurname.value = ''
  activeRequestId.value = null
}

function currentRoute() {
  return `/pages/join/family?familyId=${encodeURIComponent(familyId.value)}`
}

function formatDate(value: string) {
  const time = new Date(value)
  return Number.isNaN(time.getTime()) ? value : time.toLocaleString('zh-CN')
}

async function loadRequests() {
  authChecked.value = false
  requests.value = []
  if (!familyId.value) {
    authChecked.value = true
    loadError.value = '缺少家庭信息。'
    return
  }
  if (!session.requireLogin(currentRoute())) return
  authChecked.value = true
  loading.value = true
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

function memberLabel(member: FamilyMember) {
  const gender = member.gender === 'MALE' ? '男' : member.gender === 'FEMALE' ? '女' : '性别未知'
  const binding = member.boundUserId ? '已绑定' : member.userBindingPolicy === 'NOT_REQUIRED' ? '无需绑定' : '未绑定'
  return `${member.name}（${gender} · ${binding}）`
}

function normalizePickerIndexes() {
  if (selectedMemberIndex.value >= bindableMembers.value.length) selectedMemberIndex.value = 0
  if (baseMemberIndex.value >= members.value.length) baseMemberIndex.value = 0
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
  normalizeRelationIndex()
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
  return value
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
      location: {
        baseMemberId,
        addType,
        relationship: addType === 'ADD_SPOUSE'
          ? {
              relationshipType: 'SPOUSE',
              relationNoteType: placement.relationNoteType,
              relationNote: placement.relationNote
            }
          : {
              relationshipType: 'PARENT_CHILD',
              parentLinkType: placement.parentLinkType || 'PRIMARY',
              relationNoteType: placement.relationNoteType,
              relationNote: placement.relationNote
            }
      },
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

onShow(() => {
  session.restoreSession()
  loadRequests()
})
onHide(resetAuthView)
onUnload(resetAuthView)
</script>

<style scoped>
.state-block {
  padding: 32rpx 0;
  text-align: center;
}

.item-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16rpx;
  margin-bottom: 12rpx;
}

.item-title-block {
  flex: 1;
  min-width: 0;
}

.item-name {
  display: block;
  color: #1f2937;
  font-size: 30rpx;
  font-weight: 700;
}

.item-meta {
  display: block;
  margin-top: 8rpx;
}

.item-actions {
  display: flex;
  flex-direction: column;
  gap: 12rpx;
  margin-top: 20rpx;
}

.resolve-panel {
  margin-top: 20rpx;
  padding: 20rpx;
  border: 1rpx solid var(--tree-border-subtle, #f0ebe3);
  border-radius: 20rpx;
  background: rgba(247, 250, 249, 0.76);
}

.panel-title {
  display: block;
  color: var(--tree-text, #1f2937);
  font-size: 28rpx;
  font-weight: 700;
}

.panel-desc {
  display: block;
  margin: 8rpx 0 16rpx;
  line-height: 1.6;
}

.picker-field {
  box-sizing: border-box;
  width: 100%;
  min-height: 84rpx;
  margin-bottom: 18rpx;
  padding: 22rpx 24rpx;
  border: 1rpx solid var(--tree-border-warm, #ebe4d6);
  border-radius: 16rpx;
  background: #fff;
  color: var(--tree-text, #1f2937);
  font-size: 26rpx;
}
</style>
