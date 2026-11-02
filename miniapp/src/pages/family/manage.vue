<template>
  <view class="archive-page manage-page">
    <MiniBackHome :return-url="returnUrl" :return-label="returnLabel" />

    <view v-if="loading" class="manage-state archive-form-panel">
      <MiniEmptyState symbol="…" title="正在加载" description="正在读取家庭成员与关系..." />
    </view>
    <view v-else-if="loadError" class="manage-state archive-form-panel">
      <MiniNotice tone="warm" title="加载失败">{{ loadError }}</MiniNotice>
      <MiniButton variant="secondary" @click="loadData">重新加载</MiniButton>
    </view>

    <template v-else>
      <view v-if="family" class="manage-head archive-page-head">
        <view>
          <text class="archive-kicker">Family Tree Edit</text>
          <text class="archive-title">添加与调整</text>
          <text class="archive-subtitle">
            {{ family.familyName }} · 按关系添加或调整成员
          </text>
        </view>
        <view class="archive-seal">{{ family.familySurname.slice(0, 1) }}</view>
      </view>

      <view v-if="family" class="manage-context">
        <text class="context-back" @click="openPrivateTree">返回家庭树</text>
      </view>

      <view v-if="!canManage" class="manage-state archive-form-panel">
        <MiniNotice tone="warm" title="无管理权限">只有家庭创建者或家庭管理员可以维护成员和关系。</MiniNotice>
        <MiniButton variant="secondary" @click="openPrivateTree">返回家庭树</MiniButton>
      </view>

      <template v-else>
      <view class="archive-segment-tabs">
        <view
          class="archive-segment-tab"
          :class="{ active: manageSection === 'node' }"
          @click="manageSection = 'node'"
        >
          <text>添加节点</text>
        </view>
        <view
          class="archive-segment-tab"
          :class="{ active: manageSection === 'relations' }"
          @click="manageSection = 'relations'"
        >
          <text>调关系</text>
        </view>
      </view>
      <view class="manage-flow-note archive-panel">
        <text class="manage-flow-title">{{ sectionGuide.title }}</text>
        <text>{{ sectionGuide.description }}</text>
      </view>
      <MiniNotice v-if="guideMode" tone="security" class="manage-guide-note">
        {{ manageSection === 'node'
          ? '引导第 1 步：选择一位熟悉的家人和关系，填好姓名后即可直接接入家庭树。'
          : '引导第 2 步：先检查已有亲属是否已连线；需要补关系时切回“添加节点”，选择对应关系即可。' }}
      </MiniNotice>

      <view v-if="manageSection === 'node'" class="archive-form-panel">
        <view class="archive-section-head">
          <text class="archive-section-title">添加节点</text>
          <text class="archive-section-subtitle">选择成员和关系，直接加入家庭树</text>
        </view>
        <picker :range="memberLabels" :value="baseMemberIndex" @change="onBaseMemberChange">
          <view class="field-picker">
            <text class="field-picker-prefix">和谁有关：</text>{{ selectedBaseMember?.name || '选择成员' }}
          </view>
        </picker>
        <picker :range="relationLabels" :value="relationIndex" @change="onRelationChange">
          <view class="field-picker">
            <text class="field-picker-prefix">准备添加：</text>{{ relationLabels[relationIndex] }}
          </view>
        </picker>
        <MiniNotice
          :key="`${selectedBaseMember?.memberId || 'member'}-${relationValues[relationIndex]}`"
          tone="security"
          title="关系预览"
        >{{ relativePreview }}</MiniNotice>
        <picker v-if="showChildCoParentPicker" :range="childCoParentLabels" :value="childCoParentIndex" @change="onChildCoParentChange">
          <view class="field-picker">
            <text class="field-picker-prefix">另一位家长：</text>{{ childCoParentLabels[childCoParentIndex] }}
          </view>
        </picker>
        <text v-if="showChildCoParentPicker" class="field-help">请选择实际共同父母；前任、现任配偶均可选。未记录时可暂不选择，后续在子女资料中补充。</text>
        <picker
          v-if="showParentRolePicker"
          :range="parentRoleLabels"
          :value="parentRoleIndex"
          @change="onParentRoleChange"
        >
          <view class="field-picker">
            <text class="field-picker-prefix">父母身份（图示）：</text>{{ parentRoleLabels[parentRoleIndex] }}
          </view>
        </picker>
        <text v-if="showParentRolePicker" class="field-help">
          仅影响树状家谱成员卡样式，不改变亲属关系；已有另一位家庭成员父/母时，可选择“家庭成员的配偶”。
        </text>
        <input v-model.trim="relativeName" class="tree-input" maxlength="80" placeholder="新成员姓名" />
        <picker :range="genderLabels" :value="relativeGenderIndex" @change="onRelativeGenderChange">
          <view class="field-picker">性别：{{ genderLabels[relativeGenderIndex] }}</view>
        </picker>
        <picker v-if="showSpouseDisplayPicker" :range="spouseDisplayLabels" :value="spouseDisplayIndex" @change="onSpouseDisplayChange">
          <view class="field-picker"><text class="field-picker-prefix">婚姻身份：</text>{{ spouseDisplayLabels[spouseDisplayIndex] }}</view>
        </picker>
        <text v-if="showSpouseDisplayPicker" class="field-help">前妻、前夫仅记录历史婚姻；现任妻、现任夫用于当前婚姻展示。子女需另行选择实际共同父母。</text>
        <input v-model="relativeBirthYear" class="tree-input" type="number" placeholder="出生年份（可选）" />
        <view class="switch-row">
          <text>目前健在</text>
          <switch :checked="relativeIsAlive" color="#163353" @change="onRelativeAliveChange" />
        </view>
        <textarea
          v-model.trim="relationNote"
          class="tree-textarea"
          maxlength="200"
          placeholder="关系备注（可选）"
        />
        <MiniNotice tone="security" title="操作说明">
          会同时创建成员和关系。重复父母或配偶会被拦截。
        </MiniNotice>
        <MiniNotice v-if="relativeDraftRestored" tone="info">已恢复上次未提交的添加内容。</MiniNotice>
        <text v-if="operationError" class="tree-field-error">{{ operationError }}</text>
        <MiniButton v-if="hasRelativeDraft" size="sm" variant="ghost" :disabled="submitting" @click="clearRelativeDraft">清空草稿</MiniButton>
        <MiniButton class="manage-action" :loading="submitting" :disabled="submitting" @click="submitRelative">
          创建并接入家庭树
        </MiniButton>
      </view>

      <view v-if="manageSection === 'relations'" class="archive-form-panel">
        <view class="archive-section-head">
          <text class="archive-section-title">调整关系</text>
          <text class="archive-section-subtitle">
            {{ relationScopeMember ? `围绕“${relationScopeMember.name}”调整当前关系` : '删除错误关系后可重新创建正确关系' }}
          </text>
        </view>
        <MiniNotice tone="security" title="操作影响">
          删除或调整关系会立即影响家庭树位置和称谓，请确认后操作。
        </MiniNotice>
        <MiniEmptyState
          v-if="visibleRelationshipEdges.length === 0"
          symbol="系"
          title="暂无关系"
          description="该成员暂无可调整关系，或当前家庭尚未创建亲属关系。"
        />
        <view v-for="edge in visibleRelationshipEdges" :key="edge.relationshipId" class="archive-relation-row">
          <view class="archive-relation-copy">
            <text class="archive-relation-title">{{ relationSentence(edge) }}</text>
            <text class="tree-muted">{{ relationshipKind(edge) }}</text>
          </view>
          <view class="archive-relation-actions">
            <MiniButton
              v-if="edge.relationshipType === 'PARENT_CHILD'"
              size="sm"
              variant="secondary"
              @click="chooseRelationshipNature(edge)"
            >
              调整性质
            </MiniButton>
            <MiniButton size="sm" variant="secondary" @click="confirmDeleteRelationship(edge.relationshipId)">
              删除
            </MiniButton>
          </view>
        </view>
      </view>
      </template>
    </template>
    <MiniPhoneUpgradeDialog v-model:visible="phoneUpgradeVisible" @bound="continueAfterPhoneUpgrade" />
  </view>
</template>

<script setup lang="ts">
import { onLoad, onShow } from '@dcloudio/uni-app'
import { computed, ref, watch } from 'vue'

import { apiErrorMessage } from '@/api/client'
import { getFamilyDetail } from '@/api/families'
import { listFamilyMembers } from '@/api/members'
import {
  createRelationship,
  deleteRelationship,
  updateRelationship
} from '@/api/relationships'
import { getPrivateTree } from '@/api/tree'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
import MiniPhoneUpgradeDialog from '@/components/base/MiniPhoneUpgradeDialog.vue'
import { markNewUserGuideMilestone } from '@/features/onboarding/newUserGuide'
import { spouseDisplayOptions, spouseRelationLabel } from '@/features/family-tree/spouseRelationLabel'
import { useSessionStore } from '@/stores/session'
import type {
  FamilyDetail,
  FamilyMember,
  FamilyTreeResult,
  Gender,
  MemberType,
  RelationshipAddType,
  TreeEdge
} from '@/types/api'
import { normalizeText, optionalText, validateTextFields } from '@/utils/inputValidation'

const session = useSessionStore()
const familyId = ref('')
const returnTo = ref('')
const family = ref<FamilyDetail | null>(null)
const members = ref<FamilyMember[]>([])
const tree = ref<FamilyTreeResult>({
  familyId: 0,
  treeMode: 'LIST_TREE',
  graphVersion: 0,
  nodes: [],
  edges: [],
  tree: []
})
const returnUrl = computed(() => returnTo.value === 'tree'
  ? `/pages/family/private-tree?familyId=${encodeURIComponent(familyId.value)}`
  : `/pages/family/detail?familyId=${encodeURIComponent(familyId.value)}`)
const returnLabel = computed(() => returnTo.value === 'tree' ? '返回家庭树' : '返回家庭详情')
const loading = ref(false)
const loadError = ref('')
const operationError = ref('')
const phoneUpgradeVisible = ref(false)
const pendingCapacityAction = ref<'create' | null>(null)

const relationValues: RelationshipAddType[] = [
  'ADD_FATHER',
  'ADD_MOTHER',
  'ADD_CHILD',
  'ADD_SPOUSE',
  'ADD_SIBLING'
]
const relationLabels = ['父亲', '母亲', '子女', '配偶', '兄弟姐妹']
const genderValues: Gender[] = ['MALE', 'FEMALE']
const genderLabels = ['男', '女']
const parentRoleValues: MemberType[] = ['LINEAGE_MEMBER', 'SPOUSE']
const parentRoleLabels = ['家庭成员', '家庭成员的配偶']
const baseMemberIndex = ref(0)
const relationIndex = ref(0)
const parentRoleIndex = ref(0)
const relativeName = ref('')
const relativeGenderIndex = ref(0)
const relativeBirthYear = ref('')
const relativeIsAlive = ref(true)
const relationNote = ref('')
const submitting = ref(false)
const manageSection = ref<'node' | 'relations'>('node')
const initialBaseMemberId = ref('')
const initialAddType = ref<RelationshipAddType | ''>('')
const guideMode = ref(false)
const initialSelectionApplied = ref(false)
const hasLoadedOnce = ref(false)
const selectedBaseMemberId = ref('')
const relativeDraftReady = ref(false)
const relativeDraftRestored = ref(false)
const hasRelativeDraft = ref(false)
const childCoParentIndex = ref(0)
const spouseDisplayIndex = ref(0)

const canManage = computed(() =>
  family.value?.role === 'FOUNDER' || family.value?.role === 'FAMILY_ADMIN'
)
const sectionGuide = computed(() => ({
  node: {
    title: '关系已确定',
    description: '选择一位已有成员，再添加其父母、子女、配偶或兄弟姐妹。新成员会立即进入家庭树。'
  },
  relations: {
    title: '修正已有关系',
    description: '删除或调整关系会立即影响家庭树位置和称谓，请确认后操作。'
  }
})[manageSection.value])
const memberLabels = computed(() => members.value.map((member) => member.name))
const selectedBaseMember = computed(() => members.value[baseMemberIndex.value])
const relativePreview = computed(() => {
  const baseName = selectedBaseMember.value?.name || '所选成员'
  return `将为“${baseName}”添加一位${relationLabels[relationIndex.value]}。`
})
const childCoParentOptions = computed(() => {
  const baseMemberId = selectedBaseMember.value?.memberId
  if (!baseMemberId) return []
  return tree.value.edges
    .filter((edge) => edge.relationshipType === 'SPOUSE' && (edge.fromMemberId === baseMemberId || edge.toMemberId === baseMemberId))
    .map((edge) => {
      const memberId = edge.fromMemberId === baseMemberId ? edge.toMemberId : edge.fromMemberId
      const member = members.value.find((item) => item.memberId === memberId)
      return member ? { memberId, label: `${member.name}（${spouseRelationLabel(edge.relationNoteType, member.gender)}）` } : null
    })
    .filter((item): item is { memberId: number; label: string } => Boolean(item))
})
const childCoParentLabels = computed(() => ['暂不记录另一位家长', ...childCoParentOptions.value.map((item) => item.label)])
const showChildCoParentPicker = computed(() => relationValues[relationIndex.value] === 'ADD_CHILD' && childCoParentOptions.value.length > 0)
const spouseDisplayOptionsForRelative = computed(() => spouseDisplayOptions(genderValues[relativeGenderIndex.value]))
const spouseDisplayLabels = computed(() => spouseDisplayOptionsForRelative.value.map((option) => option.label))
const showSpouseDisplayPicker = computed(() => relationValues[relationIndex.value] === 'ADD_SPOUSE')
const showParentRolePicker = computed(() => {
  const addType = relationValues[relationIndex.value]
  return addType === 'ADD_FATHER' || addType === 'ADD_MOTHER'
})
const relationScopeMember = computed(() =>
  members.value.find((member) => String(member.memberId) === initialBaseMemberId.value) || null
)
const visibleRelationshipEdges = computed(() => {
  const scopeMemberId = Number(initialBaseMemberId.value)
  if (!scopeMemberId || !Number.isFinite(scopeMemberId)) return tree.value.edges
  return tree.value.edges.filter((edge) =>
    edge.fromMemberId === scopeMemberId || edge.toMemberId === scopeMemberId
  )
})

function asIndex(event: { detail: { value: string | number } }) {
  return Number(event.detail.value)
}

function onBaseMemberChange(event: { detail: { value: string | number } }) {
  baseMemberIndex.value = asIndex(event)
  selectedBaseMemberId.value = String(selectedBaseMember.value?.memberId || '')
  applyRelationGenderDefault()
  childCoParentIndex.value = 0
  spouseDisplayIndex.value = 0
}

function onRelationChange(event: { detail: { value: string | number } }) {
  const nextIndex = asIndex(event)
  const nextType = relationValues[nextIndex]
  if (
    selectedBaseMember.value?.gender === 'UNKNOWN' &&
    (nextType === 'ADD_CHILD' || nextType === 'ADD_SPOUSE')
  ) {
    operationError.value = '该成员性别尚未确认，暂不能直接添加子女或配偶；可以先补充父母或编辑成员资料。'
    return
  }
  relationIndex.value = nextIndex
  childCoParentIndex.value = 0
  spouseDisplayIndex.value = 0
  parentRoleIndex.value = defaultParentRoleIndex(relationValues[relationIndex.value], selectedBaseMember.value?.memberId)
  applyRelationGenderDefault()
}

function onChildCoParentChange(event: { detail: { value: string | number } }) {
  childCoParentIndex.value = asIndex(event)
}

function onParentRoleChange(event: { detail: { value: string | number } }) {
  parentRoleIndex.value = asIndex(event)
}

function onRelativeGenderChange(event: { detail: { value: string | number } }) {
  relativeGenderIndex.value = asIndex(event)
  spouseDisplayIndex.value = 0
}

function onSpouseDisplayChange(event: { detail: { value: string | number } }) {
  spouseDisplayIndex.value = asIndex(event)
}

function onRelativeAliveChange(event: { detail: { value: boolean } }) {
  relativeIsAlive.value = Boolean(event.detail.value)
}

function applyRelationGenderDefault() {
  const addType = relationValues[relationIndex.value]
  if (addType === 'ADD_FATHER') relativeGenderIndex.value = 0
  if (addType === 'ADD_MOTHER') relativeGenderIndex.value = 1
  if (addType === 'ADD_SPOUSE') {
    relativeGenderIndex.value = selectedBaseMember.value?.gender === 'FEMALE' ? 0 : 1
  }
}

function defaultParentRoleIndex(addType: RelationshipAddType | undefined, baseMemberId?: number) {
  if (addType === 'ADD_MOTHER' && hasLineageParentByGender(baseMemberId, 'MALE')) return 1
  if (addType === 'ADD_FATHER' && hasLineageParentByGender(baseMemberId, 'FEMALE')) return 1
  return 0
}

function hasParentByGender(childMemberId: number | undefined, gender: Gender) {
  if (!childMemberId) return false
  const parentIds = tree.value.edges
    .filter((edge) => edge.relationshipType === 'PARENT_CHILD' && edge.toMemberId === childMemberId)
    .map((edge) => edge.fromMemberId)
  return tree.value.nodes.some((node) =>
    parentIds.includes(node.memberId) &&
    node.gender === gender
  )
}

function hasLineageParentByGender(childMemberId: number | undefined, gender: Gender) {
  if (!childMemberId) return false
  const parentIds = tree.value.edges
    .filter((edge) => edge.relationshipType === 'PARENT_CHILD' && edge.toMemberId === childMemberId)
    .map((edge) => edge.fromMemberId)
  return tree.value.nodes.some((node) =>
    parentIds.includes(node.memberId) &&
    node.gender === gender &&
    node.memberType === 'LINEAGE_MEMBER'
  )
}

function syncCreateBaseSelection(preferredMemberId = selectedBaseMemberId.value) {
  const index = members.value.findIndex((member) => String(member.memberId) === preferredMemberId)
  if (index >= 0) {
    baseMemberIndex.value = index
  } else {
    baseMemberIndex.value = Math.min(baseMemberIndex.value, Math.max(0, members.value.length - 1))
  }
  selectedBaseMemberId.value = String(selectedBaseMember.value?.memberId || '')
}

function yearValue(value: string) {
  const parsed = Number(value)
  return Number.isInteger(parsed) && parsed > 0 ? parsed : undefined
}

function relativeDraftKey() {
  return `tree_miniapp_relative_draft_${session.user?.id || 'guest'}_${familyId.value}`
}

function restoreRelativeDraft() {
  if (relativeDraftReady.value || !members.value.length) return
  relativeDraftReady.value = true
  try {
    const stored = uni.getStorageSync(relativeDraftKey())
    const parsed = typeof stored === 'string' ? JSON.parse(stored) : stored
    if (!parsed || typeof parsed !== 'object') return
    const draft = parsed as Record<string, unknown>
    const baseIndex = members.value.findIndex((member) => String(member.memberId) === String(draft.baseMemberId || ''))
    if (baseIndex >= 0) baseMemberIndex.value = baseIndex
    const savedRelation = relationValues.indexOf(draft.relationType as RelationshipAddType)
    if (savedRelation >= 0) relationIndex.value = savedRelation
    if (typeof draft.name === 'string') relativeName.value = draft.name
    if (typeof draft.birthYear === 'string') relativeBirthYear.value = draft.birthYear
    if (typeof draft.note === 'string') relationNote.value = draft.note
    if (typeof draft.genderIndex === 'number' && draft.genderIndex >= 0 && draft.genderIndex < genderValues.length) relativeGenderIndex.value = draft.genderIndex
    if (typeof draft.parentRoleIndex === 'number' && draft.parentRoleIndex >= 0 && draft.parentRoleIndex < parentRoleValues.length) parentRoleIndex.value = draft.parentRoleIndex
    if (typeof draft.isAlive === 'boolean') relativeIsAlive.value = draft.isAlive
    selectedBaseMemberId.value = String(selectedBaseMember.value?.memberId || '')
    hasRelativeDraft.value = true
    relativeDraftRestored.value = true
  } catch {
    uni.removeStorageSync(relativeDraftKey())
  }
}

function clearRelativeDraft() {
  uni.removeStorageSync(relativeDraftKey())
  hasRelativeDraft.value = false
  relativeDraftRestored.value = false
  relativeName.value = ''
  relativeBirthYear.value = ''
  relativeIsAlive.value = true
  relationNote.value = ''
}

watch(
  [selectedBaseMemberId, relationIndex, parentRoleIndex, relativeName, relativeGenderIndex, relativeBirthYear, relativeIsAlive, relationNote],
  () => {
    if (!relativeDraftReady.value || submitting.value) return
    const hasContent = Boolean(relativeName.value.trim() || relativeBirthYear.value.trim() || relationNote.value.trim())
    if (!hasContent) {
      uni.removeStorageSync(relativeDraftKey())
      hasRelativeDraft.value = false
      return
    }
    uni.setStorageSync(relativeDraftKey(), JSON.stringify({
      baseMemberId: selectedBaseMember.value?.memberId || '', relationType: relationValues[relationIndex.value],
      parentRoleIndex: parentRoleIndex.value, name: relativeName.value, genderIndex: relativeGenderIndex.value,
      birthYear: relativeBirthYear.value, isAlive: relativeIsAlive.value, note: relationNote.value
    }))
    hasRelativeDraft.value = true
  }
)

async function loadData() {
  if (!familyId.value) {
    loadError.value = '缺少家庭信息。'
    loading.value = false
    return
  }
  const route = `/pages/family/manage?familyId=${encodeURIComponent(familyId.value)}`
  if (!session.requireLogin(route)) return
  loading.value = true
  loadError.value = ''
  try {
    const previousBaseMemberId = selectedBaseMemberId.value || String(selectedBaseMember.value?.memberId || '')
    const shouldApplyInitialSelection = !initialSelectionApplied.value
    const [familyResult, memberResult, treeResult] = await Promise.all([
      getFamilyDetail(familyId.value),
      listFamilyMembers(familyId.value),
      getPrivateTree(familyId.value)
    ])
    family.value = familyResult
    members.value = memberResult
    tree.value = treeResult
    applyInitialSelection()
    syncCreateBaseSelection(shouldApplyInitialSelection ? selectedBaseMemberId.value : previousBaseMemberId)
    if (shouldApplyInitialSelection) {
      parentRoleIndex.value = defaultParentRoleIndex(
        relationValues[relationIndex.value],
        selectedBaseMember.value?.memberId
      )
      applyRelationGenderDefault()
    }
    restoreRelativeDraft()
    hasLoadedOnce.value = true
  } catch (error) {
    loadError.value = apiErrorMessage(error, '家庭管理数据加载失败。')
  } finally {
    loading.value = false
  }
}

function applyInitialSelection() {
  if (initialSelectionApplied.value) return
  initialSelectionApplied.value = true
  const memberIndex = members.value.findIndex(
    (member) => String(member.memberId) === initialBaseMemberId.value
  )
  if (memberIndex >= 0) {
    baseMemberIndex.value = memberIndex
    selectedBaseMemberId.value = String(members.value[memberIndex]?.memberId || '')
  }
  const addTypeIndex = relationValues.indexOf(initialAddType.value as RelationshipAddType)
  if (addTypeIndex >= 0) relationIndex.value = addTypeIndex
}

async function submitRelative() {
  const base = selectedBaseMember.value
  if (!base || !relativeName.value || submitting.value) {
    operationError.value = base ? '请填写新亲属姓名。' : '请先选择关联成员。'
    return
  }
  const validationMessage = validateTextFields([
    { value: relativeName.value, label: '新亲属姓名', kind: 'name', required: true, maxLength: 80 },
    { value: relationNote.value, label: '关系说明', kind: 'multiLine', maxLength: 200 }
  ])
  if (validationMessage) {
    operationError.value = validationMessage
    return
  }
  const addType = relationValues[relationIndex.value]
  if (base.gender === 'UNKNOWN' && (addType === 'ADD_CHILD' || addType === 'ADD_SPOUSE')) {
    operationError.value = '该成员性别尚未确认，暂不能直接添加子女或配偶；请先编辑成员资料确认性别。'
    return
  }
  submitting.value = true
  operationError.value = ''
  try {
    if (addType === 'ADD_FATHER' && hasParentByGender(base.memberId, 'MALE')) {
      operationError.value = '该成员已有父亲节点，不能重复添加父亲。'
      return
    }
    if (addType === 'ADD_MOTHER' && hasParentByGender(base.memberId, 'FEMALE')) {
      operationError.value = '该成员已有母亲节点，不能重复添加母亲。'
      return
    }
    const newMember: {
      name: string
      gender: Gender
      birthYear?: number
      isAlive: boolean
      userBindingPolicy: 'OPTIONAL'
      memberType?: MemberType
    } = {
      name: normalizeText(relativeName.value),
      gender: genderValues[relativeGenderIndex.value],
      birthYear: yearValue(relativeBirthYear.value),
      isAlive: relativeIsAlive.value,
      userBindingPolicy: 'OPTIONAL'
    }
    if (addType === 'ADD_FATHER' || addType === 'ADD_MOTHER') {
      newMember.memberType = parentRoleValues[parentRoleIndex.value]
    }
    const result = await createRelationship(familyId.value, {
      baseMemberId: base.memberId,
      coParentMemberId: childCoParentIndex.value > 0 ? childCoParentOptions.value[childCoParentIndex.value - 1]?.memberId : undefined,
      addType,
      newMember,
      relationship: {
        parentLinkType: relationValues[relationIndex.value] === 'ADD_SPOUSE' ? undefined : 'PRIMARY',
        relationNoteType: addType === 'ADD_SPOUSE'
          ? spouseDisplayOptionsForRelative.value[spouseDisplayIndex.value]?.noteType || undefined
          : undefined,
        relationNote: optionalText(relationNote.value)
      }
    })
    relativeName.value = ''
    relativeBirthYear.value = ''
    relativeIsAlive.value = true
    relationNote.value = ''
    uni.removeStorageSync(relativeDraftKey())
    hasRelativeDraft.value = false
    relativeDraftRestored.value = false
    await loadData()
    showContinueAddModal(result.createdMember?.memberId, newMember.name)
  } catch (error) {
    if (isMemberLimitError(error)) {
      if (!session.user?.phoneVerified) {
        pendingCapacityAction.value = 'create'
        phoneUpgradeVisible.value = true
      } else {
        openCapacityPurchase()
      }
      return
    }
    operationError.value = apiErrorMessage(error, '创建亲属失败。')
  } finally {
    submitting.value = false
  }
}

function showContinueAddModal(memberId: number | undefined, memberName: string) {
  uni.showModal({
    title: '已接入家庭树',
    content: `“${memberName}”已在家庭树中。是否继续添加 TA 的亲属？`,
    confirmText: '继续添加',
    cancelText: '知道了',
    success: (result) => {
      if (!result.confirm || !memberId) return
      manageSection.value = 'node'
      syncCreateBaseSelection(String(memberId))
      relationIndex.value = 0
      parentRoleIndex.value = defaultParentRoleIndex(
        relationValues[relationIndex.value],
        selectedBaseMember.value?.memberId
      )
      applyRelationGenderDefault()
    }
  })
}

function memberName(memberId: number) {
  return members.value.find((member) => member.memberId === memberId)?.name || '未知成员'
}

function relationSentence(edge: TreeEdge) {
  const from = memberName(edge.fromMemberId)
  const to = memberName(edge.toMemberId)
  return edge.relationshipType === 'SPOUSE' ? `${from} 与 ${to}` : `${from} → ${to}`
}

function relationshipKind(edge: TreeEdge) {
  if (edge.relationshipType === 'SPOUSE') {
    const spouse = members.value.find((member) => member.memberId === edge.toMemberId)
      || members.value.find((member) => member.memberId === edge.fromMemberId)
    return `配偶关系 · 展示身份：${spouseRelationLabel(edge.relationNoteType, spouse?.gender)}`
  }
  const labels: Record<string, string> = {
    PRIMARY: '亲生父母子女',
    STEP: '继亲父母子女',
    ADOPTIVE: '收养父母子女',
    SUCCESSION: '承继关系',
    NOTE_ONLY: '备注关系',
    OTHER: '其他父母子女关系'
  }
  return labels[edge.parentLinkType || 'PRIMARY'] || '父母子女关系'
}

function confirmDeleteRelationship(relationshipId: number) {
  uni.showModal({
    title: '删除关系',
    content: '删除后家庭树会立即更新，确定继续吗？',
    success: (result) => {
      if (result.confirm) removeRelationship(relationshipId)
    }
  })
}

function chooseRelationshipNature(edge: TreeEdge) {
  const values = ['PRIMARY', 'STEP', 'ADOPTIVE']
  uni.showActionSheet({
    itemList: ['亲生父母子女', '继亲父母子女', '收养父母子女'],
    success: async (result) => {
      operationError.value = ''
      try {
        await updateRelationship(familyId.value, edge.relationshipId, {
          parentLinkType: values[result.tapIndex],
          expectedRevision: edge.revision
        })
        uni.showToast({ title: '关系性质已更新', icon: 'success' })
        await loadData()
      } catch (error) {
        operationError.value = apiErrorMessage(error, '调整关系性质失败。')
      }
    }
  })
}


async function removeRelationship(relationshipId: number) {
  operationError.value = ''
  try {
    await deleteRelationship(familyId.value, relationshipId, '小程序家庭管理删除')
    uni.showToast({ title: '关系已删除', icon: 'success' })
    await loadData()
  } catch (error) {
    operationError.value = apiErrorMessage(error, '删除关系失败。')
  }
}

function openPrivateTree() {
  uni.redirectTo({
    url: `/pages/family/private-tree?familyId=${encodeURIComponent(familyId.value)}`
  })
}

function openCapacityPurchase() {
  uni.showActionSheet({ itemList: ['购买扩容', '邀请家人解锁奖励名额'], success: (result) => {
    const page = result.tapIndex === 0 ? 'capacity' : 'invitation-reward'
    const capacityReturn = page === 'capacity' && pendingCapacityAction.value
      ? 'add'
      : returnTo.value === 'tree'
        ? 'tree'
        : ''
    uni.navigateTo({ url: `/pages/family/${page}?familyId=${encodeURIComponent(familyId.value)}${capacityReturn ? `&returnTo=${capacityReturn}` : ''}` })
  } })
}

function continueAfterPhoneUpgrade() {
  const action = pendingCapacityAction.value
  pendingCapacityAction.value = null
  if (action === 'create') {
    void submitRelative()
    return
  }
}

function isMemberLimitError(error: unknown) {
  return typeof error === 'object' && error !== null && 'code' in error && (error as { code?: unknown }).code === 41504
}

onLoad((options) => {
  familyId.value = String(options?.familyId || '')
  returnTo.value = String(options?.returnTo || '')
  initialBaseMemberId.value = String(options?.baseMemberId || '')
  const requestedType = String(options?.addType || '') as RelationshipAddType
  initialAddType.value = relationValues.includes(requestedType) ? requestedType : ''
  const requestedSection = String(options?.section || '')
  guideMode.value = String(options?.guide || '') === '1'
  if (requestedSection === 'node' || requestedSection === 'relations') {
    manageSection.value = requestedSection
  } else if (requestedSection === 'create') {
    manageSection.value = 'node'
  }
  if (guideMode.value) {
    markNewUserGuideMilestone(manageSection.value === 'relations' ? 'RELATION_GUIDE_OPENED' : 'MEMBER_GUIDE_OPENED')
  }
  if (!familyId.value) {
    loadError.value = '缺少家庭信息。'
    return
  }
  loadData()
})

onShow(() => {
  if (!hasLoadedOnce.value || loading.value || !familyId.value) return
  loadData()
})
</script>

<style scoped>
.manage-page {
  padding-top: 28rpx;
}

.manage-page :deep(.mini-back-home) {
  margin-bottom: 22rpx;
}

.manage-state :deep(.mini-notice) {
  margin-bottom: 18rpx;
}

.manage-page :deep(.mini-notice) {
  margin-top: 16rpx;
  margin-bottom: 16rpx;
}

.manage-context {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16rpx;
  margin-bottom: 24rpx;
}

.context-back {
  color: var(--archive-ink-soft);
  font-size: 23rpx;
  line-height: 1.4;
}

.context-back:active {
  color: var(--archive-cinnabar);
}

.manage-flow-note {
  display: flex;
  flex-direction: column;
  gap: 8rpx;
  margin-bottom: 20rpx;
  padding: 18rpx 0;
  color: var(--archive-ink-soft);
  font-size: 23rpx;
  line-height: 1.6;
}

.manage-page .archive-form-panel + .archive-form-panel {
  margin-top: -8rpx;
}

.manage-action {
  margin-top: 20rpx;
  align-self: flex-start;
}

.field-picker-prefix {
  color: var(--archive-ink-soft);
}

.field-help {
  display: block;
  margin: -10rpx 0 16rpx;
  color: var(--archive-ink-soft);
  font-size: 23rpx;
  line-height: 1.7;
}

.manage-page .archive-relation-actions {
  flex-direction: row;
  flex-wrap: wrap;
  justify-content: flex-end;
}
</style>
