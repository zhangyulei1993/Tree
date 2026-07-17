<template>
  <view class="archive-page manage-page">
    <MiniBackHome />

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
            {{ family.familyName }} · 先确认关系，再选择添加、暂存或调整
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
          :class="{ active: manageSection === 'pending' }"
          @click="manageSection = 'pending'"
        >
          <text>添加暂存</text>
        </view>
        <view
          class="archive-segment-tab"
          :class="{ active: manageSection === 'place' }"
          @click="manageSection = 'place'"
        >
          <text>接入暂存</text>
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

      <view v-if="manageSection === 'node'" class="archive-form-panel">
        <view class="archive-section-head">
          <text class="archive-section-title">添加节点</text>
          <text class="archive-section-subtitle">已确认亲属关系时，直接添加到家庭树</text>
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
        <MiniNotice tone="security" title="关系预览">{{ relativePreview }}</MiniNotice>
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
          仅影响关系图纸签样式，不改变亲属关系；已有另一位家庭成员父/母时，可选择“家庭成员的配偶”。
        </text>
        <input v-model.trim="relativeName" class="tree-input" maxlength="80" placeholder="新成员姓名" />
        <picker :range="genderLabels" :value="relativeGenderIndex" @change="onRelativeGenderChange">
          <view class="field-picker">性别：{{ genderLabels[relativeGenderIndex] }}</view>
        </picker>
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
          此操作会同时创建新成员和关系，不用于连接两个已有成员。重复父亲、母亲或配偶由系统校验并给出提示。
        </MiniNotice>
        <text v-if="operationError" class="tree-field-error">{{ operationError }}</text>
        <MiniButton class="manage-action" :loading="submitting" :disabled="submitting" @click="submitRelative">
          创建并接入家庭树
        </MiniButton>
      </view>

      <view v-if="manageSection === 'pending'" class="archive-form-panel">
        <view class="archive-section-head">
          <text class="archive-section-title">添加暂存</text>
          <text class="archive-section-subtitle">知道姓名但分支身份还不确定时，先暂存为成员档案</text>
        </view>
        <input v-model.trim="unlocatedName" class="tree-input" maxlength="80" placeholder="成员姓名" />
        <picker :range="genderLabels" :value="unlocatedGenderIndex" @change="onUnlocatedGenderChange">
          <view class="field-picker">性别：{{ genderLabels[unlocatedGenderIndex] }}</view>
        </picker>
        <view class="switch-row">
          <text>目前健在</text>
          <switch :checked="unlocatedIsAlive" color="#163353" @change="onUnlocatedAliveChange" />
        </view>
        <MiniButton
          class="manage-action"
          variant="secondary"
          :loading="creatingUnlocated"
          :disabled="creatingUnlocated"
          @click="submitUnlocated"
        >
          暂存成员
        </MiniButton>
      </view>

      <view v-if="manageSection === 'place'" class="archive-form-panel">
        <view class="archive-section-head">
          <text class="archive-section-title">接入暂存</text>
          <text class="archive-section-subtitle">确认分支后，再把暂存成员接入家庭树</text>
        </view>
        <MiniEmptyState
          v-if="unlocatedMembers.length === 0"
          symbol="位"
          title="没有待接入成员"
          description="暂存但尚未建立任何关系的成员会显示在这里。"
        />
        <template v-else>
          <picker :range="unlocatedMemberLabels" :value="placeMemberIndex" @change="onPlaceMemberChange">
            <view class="field-picker">待接入：{{ selectedPlaceMember?.name || '选择成员' }}</view>
          </picker>
          <picker :range="placementBaseLabels" :value="placeBaseIndex" @change="onPlaceBaseChange">
            <view class="field-picker">和谁有关：{{ selectedPlaceBase?.name || '选择成员' }}</view>
          </picker>
          <picker :range="relationLabels" :value="placeRelationIndex" @change="onPlaceRelationChange">
            <view class="field-picker">确认关系为：{{ relationLabels[placeRelationIndex] }}</view>
          </picker>
          <MiniNotice tone="security" title="接入预览">{{ placementPreview }}</MiniNotice>
          <picker
            v-if="showPlaceParentRolePicker"
            :range="parentRoleLabels"
            :value="placeParentRoleIndex"
            @change="onPlaceParentRoleChange"
          >
            <view class="field-picker">父母身份（图示）：{{ parentRoleLabels[placeParentRoleIndex] }}</view>
          </picker>
          <text v-if="showPlaceParentRolePicker" class="field-help">
            仅影响关系图纸签样式，不改变亲属关系；已有另一位家庭成员父/母时，可选择“家庭成员的配偶”。
          </text>
          <MiniNotice v-if="unlocatedTypeAnomalies.length > 0" tone="warm" title="成员类型数据异常">
            以下暂存成员缺少 memberType，已从待接入列表排除：{{
              unlocatedTypeAnomalies.map((member) => member.name).join('、')
            }}
          </MiniNotice>
          <MiniNotice tone="security" title="接入说明">
            例如选择“子女”，表示这位暂存成员是所选成员的子女。
          </MiniNotice>
          <MiniButton class="manage-action" :loading="placingMember" :disabled="placingMember" @click="submitPlacement">
            接入家庭树
          </MiniButton>
        </template>
      </view>

      <view v-if="manageSection === 'relations'" class="archive-form-panel">
        <view class="archive-section-head">
          <text class="archive-section-title">调整关系</text>
          <text class="archive-section-subtitle">
            {{ relationScopeMember ? `围绕“${relationScopeMember.name}”调整当前关系` : '删除错误关系后可重新创建正确关系' }}
          </text>
        </view>
        <MiniNotice tone="security" title="操作影响">
          删除或调整关系会立即影响家庭树位置和称谓；如果只是还没确认分支，建议先使用「添加暂存」。
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
  </view>
</template>

<script setup lang="ts">
import { onLoad, onShow } from '@dcloudio/uni-app'
import { computed, ref } from 'vue'

import { apiErrorMessage } from '@/api/client'
import { getFamilyDetail } from '@/api/families'
import {
  createFamilyMember,
  listFamilyMembers
} from '@/api/members'
import {
  createRelationship,
  deleteRelationship,
  placeExistingMember,
  updateRelationship
} from '@/api/relationships'
import { getPrivateTree } from '@/api/tree'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
import { filterUnlocatedLineageMemberIds } from '@/features/family-tree/graph'
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
const loading = ref(false)
const loadError = ref('')
const operationError = ref('')

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
const unlocatedName = ref('')
const unlocatedGenderIndex = ref(0)
const unlocatedIsAlive = ref(true)
const creatingUnlocated = ref(false)
const placeMemberIndex = ref(0)
const placeBaseIndex = ref(0)
const placeRelationIndex = ref(0)
const placeParentRoleIndex = ref(0)
const placingMember = ref(false)
const manageSection = ref<'node' | 'pending' | 'place' | 'relations'>('node')
const initialBaseMemberId = ref('')
const initialAddType = ref<RelationshipAddType | ''>('')
const initialSelectionApplied = ref(false)
const hasLoadedOnce = ref(false)
const selectedBaseMemberId = ref('')
const selectedPlaceMemberId = ref('')
const selectedPlaceBaseMemberId = ref('')

const canManage = computed(() =>
  family.value?.role === 'FOUNDER' || family.value?.role === 'FAMILY_ADMIN'
)
const sectionGuide = computed(() => ({
  node: {
    title: '关系已确定',
    description: '选择一位已有成员，再添加其父母、子女、配偶或兄弟姐妹。新成员会立即进入家庭树。'
  },
  pending: {
    title: '关系暂未确认',
    description: '先保存姓名和基础资料，不进入关系图；确认分支后再接入家庭树。'
  },
  place: {
    title: '把暂存成员接入家庭树',
    description: '选择待接入成员及其关系对象，确认后才会出现在对应分支。'
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
const relatedMemberIDs = computed(() => {
  const ids = new Set<number>()
  tree.value.edges.forEach((edge) => {
    ids.add(edge.fromMemberId)
    ids.add(edge.toMemberId)
  })
  return ids
})
const treeMemberTypeById = computed(() => {
  const map = new Map<number, string>()
  tree.value.nodes.forEach((node) => map.set(node.memberId, node.memberType))
  return map
})
const unlocatedTypeAnomalies = computed(() =>
  members.value.filter((member) => {
    if (relatedMemberIDs.value.has(member.memberId)) return false
    return !treeMemberTypeById.value.has(member.memberId)
  })
)
const unlocatedMembers = computed(() =>
  members.value.filter((member) =>
    filterUnlocatedLineageMemberIds(
      [member.memberId],
      relatedMemberIDs.value,
      treeMemberTypeById.value
    ).includes(member.memberId)
  )
)
const showParentRolePicker = computed(() => {
  const addType = relationValues[relationIndex.value]
  return addType === 'ADD_FATHER' || addType === 'ADD_MOTHER'
})
const showPlaceParentRolePicker = computed(() => {
  const addType = relationValues[placeRelationIndex.value]
  return addType === 'ADD_FATHER' || addType === 'ADD_MOTHER'
})
const unlocatedMemberLabels = computed(() => unlocatedMembers.value.map((member) => member.name))
const selectedPlaceMember = computed(() => unlocatedMembers.value[placeMemberIndex.value])
const placementBaseMembers = computed(() => members.value.filter((member) => member.memberId !== selectedPlaceMember.value?.memberId))
const placementBaseLabels = computed(() => placementBaseMembers.value.map((member) => member.name))
const selectedPlaceBase = computed(() => placementBaseMembers.value[placeBaseIndex.value])
const placementPreview = computed(() => {
  const memberName = selectedPlaceMember.value?.name || '该暂存成员'
  const baseName = selectedPlaceBase.value?.name || '所选成员'
  return `确认后，“${memberName}”将作为“${baseName}”的${relationLabels[placeRelationIndex.value]}接入家庭树。`
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
}

function onRelationChange(event: { detail: { value: string | number } }) {
  relationIndex.value = asIndex(event)
  parentRoleIndex.value = defaultParentRoleIndex(relationValues[relationIndex.value], selectedBaseMember.value?.memberId)
  applyRelationGenderDefault()
}

function onParentRoleChange(event: { detail: { value: string | number } }) {
  parentRoleIndex.value = asIndex(event)
}

function onRelativeGenderChange(event: { detail: { value: string | number } }) {
  relativeGenderIndex.value = asIndex(event)
}

function onRelativeAliveChange(event: { detail: { value: boolean } }) {
  relativeIsAlive.value = Boolean(event.detail.value)
}

function onUnlocatedGenderChange(event: { detail: { value: string | number } }) {
  unlocatedGenderIndex.value = asIndex(event)
}

function onUnlocatedAliveChange(event: { detail: { value: boolean } }) {
  unlocatedIsAlive.value = Boolean(event.detail.value)
}

function onPlaceMemberChange(event: { detail: { value: string | number } }) {
  placeMemberIndex.value = asIndex(event)
  selectedPlaceMemberId.value = String(selectedPlaceMember.value?.memberId || '')
  placeBaseIndex.value = 0
  selectedPlaceBaseMemberId.value = String(selectedPlaceBase.value?.memberId || '')
}

function onPlaceBaseChange(event: { detail: { value: string | number } }) {
  placeBaseIndex.value = asIndex(event)
  selectedPlaceBaseMemberId.value = String(selectedPlaceBase.value?.memberId || '')
}

function onPlaceRelationChange(event: { detail: { value: string | number } }) {
  placeRelationIndex.value = asIndex(event)
  placeParentRoleIndex.value = defaultParentRoleIndex(
    relationValues[placeRelationIndex.value],
    selectedPlaceBase.value?.memberId
  )
}

function onPlaceParentRoleChange(event: { detail: { value: string | number } }) {
  placeParentRoleIndex.value = asIndex(event)
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

function syncPlacementSelection(
  preferredMemberId = selectedPlaceMemberId.value,
  preferredBaseMemberId = selectedPlaceBaseMemberId.value
) {
  const memberIndex = unlocatedMembers.value.findIndex(
    (member) => String(member.memberId) === preferredMemberId
  )
  if (memberIndex >= 0) {
    placeMemberIndex.value = memberIndex
  } else {
    placeMemberIndex.value = Math.min(placeMemberIndex.value, Math.max(0, unlocatedMembers.value.length - 1))
  }
  selectedPlaceMemberId.value = String(selectedPlaceMember.value?.memberId || '')

  const baseIndex = placementBaseMembers.value.findIndex(
    (member) => String(member.memberId) === preferredBaseMemberId
  )
  if (baseIndex >= 0) {
    placeBaseIndex.value = baseIndex
  } else {
    placeBaseIndex.value = Math.min(placeBaseIndex.value, Math.max(0, placementBaseMembers.value.length - 1))
  }
  selectedPlaceBaseMemberId.value = String(selectedPlaceBase.value?.memberId || '')
}

function yearValue(value: string) {
  const parsed = Number(value)
  return Number.isInteger(parsed) && parsed > 0 ? parsed : undefined
}

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
    const previousPlaceMemberId = selectedPlaceMemberId.value || String(selectedPlaceMember.value?.memberId || '')
    const previousPlaceBaseMemberId = selectedPlaceBaseMemberId.value || String(selectedPlaceBase.value?.memberId || '')
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
    syncPlacementSelection(previousPlaceMemberId, previousPlaceBaseMemberId)
    if (shouldApplyInitialSelection) {
      parentRoleIndex.value = defaultParentRoleIndex(
        relationValues[relationIndex.value],
        selectedBaseMember.value?.memberId
      )
      placeParentRoleIndex.value = defaultParentRoleIndex(
        relationValues[placeRelationIndex.value],
        selectedPlaceBase.value?.memberId
      )
      applyRelationGenderDefault()
    }
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
  submitting.value = true
  operationError.value = ''
  try {
    const addType = relationValues[relationIndex.value]
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
      addType,
      newMember,
      relationship: {
        parentLinkType: relationValues[relationIndex.value] === 'ADD_SPOUSE' ? undefined : 'PRIMARY',
        relationNote: optionalText(relationNote.value)
      }
    })
    relativeName.value = ''
    relativeBirthYear.value = ''
    relativeIsAlive.value = true
    relationNote.value = ''
    await loadData()
    showContinueAddModal(result.createdMember?.memberId, newMember.name)
  } catch (error) {
    operationError.value = apiErrorMessage(error, '创建亲属失败。')
  } finally {
    submitting.value = false
  }
}

async function submitUnlocated() {
  if (!unlocatedName.value || creatingUnlocated.value) {
    operationError.value = '请填写成员姓名。'
    return
  }
  const validationMessage = validateTextFields([
    { value: unlocatedName.value, label: '成员姓名', kind: 'name', required: true, maxLength: 80 }
  ])
  if (validationMessage) {
    operationError.value = validationMessage
    return
  }
  creatingUnlocated.value = true
  operationError.value = ''
  try {
    await createFamilyMember(familyId.value, {
      name: normalizeText(unlocatedName.value),
      gender: genderValues[unlocatedGenderIndex.value],
      isAlive: unlocatedIsAlive.value,
      userBindingPolicy: 'OPTIONAL'
    })
    unlocatedName.value = ''
    unlocatedIsAlive.value = true
    await loadData()
    showPlacePendingModal()
  } catch (error) {
    operationError.value = apiErrorMessage(error, '暂存成员失败。')
  } finally {
    creatingUnlocated.value = false
  }
}

async function submitPlacement() {
  const member = selectedPlaceMember.value
  const base = selectedPlaceBase.value
  if (!member || !base || placingMember.value) {
    operationError.value = '请选择待接入成员和关联成员。'
    return
  }
  placingMember.value = true
  operationError.value = ''
  try {
    const addType = relationValues[placeRelationIndex.value]
    const payload: {
      baseMemberId: number
      memberId: number
      addType: RelationshipAddType
      memberType?: MemberType
      relationship: { parentLinkType?: string }
    } = {
      baseMemberId: base.memberId,
      memberId: member.memberId,
      addType,
      relationship: {
        parentLinkType: addType === 'ADD_SPOUSE' ? undefined : 'PRIMARY'
      }
    }
    if (addType === 'ADD_FATHER' || addType === 'ADD_MOTHER') {
      payload.memberType = parentRoleValues[placeParentRoleIndex.value]
    }
    await placeExistingMember(familyId.value, payload)
    placeMemberIndex.value = 0
    placeBaseIndex.value = 0
    await loadData()
    showContinueAddModal(member.memberId, member.name)
  } catch (error) {
    operationError.value = apiErrorMessage(error, '接入成员失败。')
  } finally {
    placingMember.value = false
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

function showPlacePendingModal() {
  uni.showModal({
    title: '成员已暂存',
    content: '该成员暂未进入关系图。确认身份后，可以把 TA 接入家庭树。',
    confirmText: '去接入',
    cancelText: '稍后',
    success: (result) => {
      if (result.confirm) manageSection.value = 'place'
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
  if (edge.relationshipType === 'SPOUSE') return '配偶关系'
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
          parentLinkType: values[result.tapIndex]
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
  if (getCurrentPages().length > 1) {
    uni.navigateBack()
    return
  }
  uni.redirectTo({
    url: `/pages/family/private-tree?familyId=${encodeURIComponent(familyId.value)}`
  })
}

onLoad((options) => {
  familyId.value = String(options?.familyId || '')
  initialBaseMemberId.value = String(options?.baseMemberId || '')
  const requestedType = String(options?.addType || '') as RelationshipAddType
  initialAddType.value = relationValues.includes(requestedType) ? requestedType : ''
  const requestedSection = String(options?.section || '')
  if (
    requestedSection === 'node' ||
    requestedSection === 'pending' ||
    requestedSection === 'place' ||
    requestedSection === 'relations'
  ) {
    manageSection.value = requestedSection
  } else if (requestedSection === 'create') {
    manageSection.value = 'node'
  } else if (requestedSection === 'locate') {
    manageSection.value = 'place'
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
