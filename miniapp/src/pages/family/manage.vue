<template>
  <view class="tree-page manage-page">
    <MiniBackHome />
    <FamilyContextHeader
      v-if="family"
      :family-name="family.familyName"
      section="编辑家谱"
      subtitle="添加亲属、定位成员并维护关系"
      :role-label="family.role === 'FOUNDER' ? '创建者' : '管理员'"
      @back="openFamilyOverview"
    />

    <MiniCard v-if="loading">
      <MiniEmptyState symbol="…" title="正在加载" description="正在读取家庭成员与关系..." />
    </MiniCard>
    <MiniCard v-else-if="loadError">
      <MiniNotice tone="warm" title="加载失败">{{ loadError }}</MiniNotice>
      <MiniButton variant="secondary" @click="loadData">重新加载</MiniButton>
    </MiniCard>
    <MiniCard v-else-if="!canManage">
      <MiniNotice tone="warm" title="无管理权限">只有家庭创建者或家庭管理员可以维护成员和关系。</MiniNotice>
    </MiniCard>

    <template v-else>
      <view class="manage-tabs">
        <button :class="{ active: manageSection === 'create' }" @click="manageSection = 'create'">添加亲属</button>
        <button :class="{ active: manageSection === 'locate' }" @click="manageSection = 'locate'">定位成员</button>
        <button :class="{ active: manageSection === 'relations' }" @click="manageSection = 'relations'">关系维护</button>
      </view>

      <MiniCard v-if="manageSection === 'create'">
        <MiniSectionHeader title="创建新亲属" subtitle="以现有成员为起点，创建一个新成员并放入家谱" />
        <picker :range="memberLabels" :value="baseMemberIndex" @change="onBaseMemberChange">
          <view class="field-picker">{{ selectedBaseMember?.name || '选择基准成员' }}</view>
        </picker>
        <picker :range="relationLabels" :value="relationIndex" @change="onRelationChange">
          <view class="field-picker">{{ relationLabels[relationIndex] }}</view>
        </picker>
        <input v-model.trim="relativeName" class="tree-input" maxlength="80" placeholder="新亲属姓名" />
        <picker :range="genderLabels" :value="relativeGenderIndex" @change="onRelativeGenderChange">
          <view class="field-picker">性别：{{ genderLabels[relativeGenderIndex] }}</view>
        </picker>
        <input v-model="relativeBirthYear" class="tree-input" type="number" placeholder="出生年份（可选）" />
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
        <MiniButton :loading="submitting" :disabled="submitting" @click="submitRelative">
          创建并放入家谱
        </MiniButton>
      </MiniCard>

      <MiniCard v-if="manageSection === 'locate'">
        <MiniSectionHeader title="暂存未定位成员" subtitle="只创建成员档案，稍后再确定其家谱位置" />
        <input v-model.trim="unlocatedName" class="tree-input" maxlength="80" placeholder="成员姓名" />
        <picker :range="genderLabels" :value="unlocatedGenderIndex" @change="onUnlocatedGenderChange">
          <view class="field-picker">性别：{{ genderLabels[unlocatedGenderIndex] }}</view>
        </picker>
        <MiniButton
          variant="secondary"
          :loading="creatingUnlocated"
          :disabled="creatingUnlocated"
          @click="submitUnlocated"
        >
          暂存成员
        </MiniButton>
      </MiniCard>

      <MiniCard v-if="manageSection === 'locate'">
        <MiniSectionHeader title="定位已有成员" subtitle="将暂存成员连接到一个现有成员，正式放入家谱" />
        <MiniEmptyState
          v-if="unlocatedMembers.length === 0"
          symbol="位"
          title="没有待定位成员"
          description="暂存但尚未建立任何关系的成员会显示在这里。"
        />
        <template v-else>
          <picker :range="unlocatedMemberLabels" :value="placeMemberIndex" @change="onPlaceMemberChange">
            <view class="field-picker">待定位：{{ selectedPlaceMember?.name || '选择成员' }}</view>
          </picker>
          <picker :range="placementBaseLabels" :value="placeBaseIndex" @change="onPlaceBaseChange">
            <view class="field-picker">基准成员：{{ selectedPlaceBase?.name || '选择基准成员' }}</view>
          </picker>
          <picker :range="relationLabels" :value="placeRelationIndex" @change="onPlaceRelationChange">
            <view class="field-picker">关系：{{ relationLabels[placeRelationIndex] }}</view>
          </picker>
          <MiniNotice tone="security" title="定位说明">
            关系以“基准成员”为起点，例如选择“子女”表示待定位成员是基准成员的子女。
          </MiniNotice>
          <MiniButton :loading="placingMember" :disabled="placingMember" @click="submitPlacement">
            放入家谱
          </MiniButton>
        </template>
      </MiniCard>

      <MiniCard v-if="manageSection === 'relations'">
        <MiniSectionHeader title="已有关系" subtitle="删除错误关系后可重新创建正确关系" />
        <MiniEmptyState
          v-if="tree.edges.length === 0"
          symbol="系"
          title="暂无关系"
          description="创建亲属关系后会显示在这里。"
        />
        <view v-for="edge in tree.edges" :key="edge.relationshipId" class="relation-row">
          <view class="relation-copy">
            <text class="relation-title">{{ relationSentence(edge) }}</text>
            <text class="tree-muted">{{ relationshipKind(edge) }}</text>
          </view>
          <view class="relation-actions">
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
      </MiniCard>
    </template>
  </view>
</template>

<script setup lang="ts">
import { onLoad } from '@dcloudio/uni-app'
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
import MiniCard from '@/components/base/MiniCard.vue'
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
import MiniSectionHeader from '@/components/base/MiniSectionHeader.vue'
import FamilyContextHeader from '@/components/family/FamilyContextHeader.vue'
import { useSessionStore } from '@/stores/session'
import type {
  FamilyDetail,
  FamilyMember,
  FamilyTreeResult,
  Gender,
  RelationshipAddType,
  TreeEdge
} from '@/types/api'

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
const baseMemberIndex = ref(0)
const relationIndex = ref(0)
const relativeName = ref('')
const relativeGenderIndex = ref(0)
const relativeBirthYear = ref('')
const relationNote = ref('')
const submitting = ref(false)
const unlocatedName = ref('')
const unlocatedGenderIndex = ref(0)
const creatingUnlocated = ref(false)
const placeMemberIndex = ref(0)
const placeBaseIndex = ref(0)
const placeRelationIndex = ref(0)
const placingMember = ref(false)
const manageSection = ref<'create' | 'locate' | 'relations'>('create')
const initialBaseMemberId = ref('')
const initialAddType = ref<RelationshipAddType | ''>('')
const initialSelectionApplied = ref(false)

const canManage = computed(() =>
  family.value?.role === 'FOUNDER' || family.value?.role === 'FAMILY_ADMIN'
)
const memberLabels = computed(() => members.value.map((member) => member.name))
const selectedBaseMember = computed(() => members.value[baseMemberIndex.value])
const relatedMemberIDs = computed(() => {
  const ids = new Set<number>()
  tree.value.edges.forEach((edge) => {
    ids.add(edge.fromMemberId)
    ids.add(edge.toMemberId)
  })
  return ids
})
const unlocatedMembers = computed(() => members.value.filter((member) => !relatedMemberIDs.value.has(member.memberId)))
const unlocatedMemberLabels = computed(() => unlocatedMembers.value.map((member) => member.name))
const selectedPlaceMember = computed(() => unlocatedMembers.value[placeMemberIndex.value])
const placementBaseMembers = computed(() => members.value.filter((member) => member.memberId !== selectedPlaceMember.value?.memberId))
const placementBaseLabels = computed(() => placementBaseMembers.value.map((member) => member.name))
const selectedPlaceBase = computed(() => placementBaseMembers.value[placeBaseIndex.value])

function asIndex(event: { detail: { value: string | number } }) {
  return Number(event.detail.value)
}

function onBaseMemberChange(event: { detail: { value: string | number } }) {
  baseMemberIndex.value = asIndex(event)
  applyRelationGenderDefault()
}

function onRelationChange(event: { detail: { value: string | number } }) {
  relationIndex.value = asIndex(event)
  applyRelationGenderDefault()
}

function onRelativeGenderChange(event: { detail: { value: string | number } }) {
  relativeGenderIndex.value = asIndex(event)
}

function onUnlocatedGenderChange(event: { detail: { value: string | number } }) {
  unlocatedGenderIndex.value = asIndex(event)
}

function onPlaceMemberChange(event: { detail: { value: string | number } }) {
  placeMemberIndex.value = asIndex(event)
  placeBaseIndex.value = 0
}

function onPlaceBaseChange(event: { detail: { value: string | number } }) {
  placeBaseIndex.value = asIndex(event)
}

function onPlaceRelationChange(event: { detail: { value: string | number } }) {
  placeRelationIndex.value = asIndex(event)
}

function applyRelationGenderDefault() {
  const addType = relationValues[relationIndex.value]
  if (addType === 'ADD_FATHER') relativeGenderIndex.value = 0
  if (addType === 'ADD_MOTHER') relativeGenderIndex.value = 1
  if (addType === 'ADD_SPOUSE') {
    relativeGenderIndex.value = selectedBaseMember.value?.gender === 'FEMALE' ? 0 : 1
  }
}

function yearValue(value: string) {
  const parsed = Number(value)
  return Number.isInteger(parsed) && parsed > 0 ? parsed : undefined
}

async function loadData() {
  const route = `/pages/family/manage?familyId=${encodeURIComponent(familyId.value)}`
  if (!session.requireLogin(route)) return
  loading.value = true
  loadError.value = ''
  try {
    const [familyResult, memberResult, treeResult] = await Promise.all([
      getFamilyDetail(familyId.value),
      listFamilyMembers(familyId.value),
      getPrivateTree(familyId.value)
    ])
    family.value = familyResult
    members.value = memberResult
    tree.value = treeResult
    baseMemberIndex.value = Math.min(baseMemberIndex.value, Math.max(0, memberResult.length - 1))
    applyInitialSelection()
    applyRelationGenderDefault()
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
  if (memberIndex >= 0) baseMemberIndex.value = memberIndex
  const addTypeIndex = relationValues.indexOf(initialAddType.value as RelationshipAddType)
  if (addTypeIndex >= 0) relationIndex.value = addTypeIndex
}

async function submitRelative() {
  const base = selectedBaseMember.value
  if (!base || !relativeName.value || submitting.value) {
    operationError.value = base ? '请填写新亲属姓名。' : '请先选择基准成员。'
    return
  }
  submitting.value = true
  operationError.value = ''
  try {
    await createRelationship(familyId.value, {
      baseMemberId: base.memberId,
      addType: relationValues[relationIndex.value],
      newMember: {
        name: relativeName.value,
        gender: genderValues[relativeGenderIndex.value],
        birthYear: yearValue(relativeBirthYear.value),
        isAlive: true,
        userBindingPolicy: 'OPTIONAL'
      },
      relationship: {
        parentLinkType: relationValues[relationIndex.value] === 'ADD_SPOUSE' ? undefined : 'PRIMARY',
        relationNote: relationNote.value || undefined
      }
    })
    relativeName.value = ''
    relativeBirthYear.value = ''
    relationNote.value = ''
    uni.showToast({ title: '已创建并放入家谱', icon: 'success' })
    await loadData()
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
  creatingUnlocated.value = true
  operationError.value = ''
  try {
    await createFamilyMember(familyId.value, {
      name: unlocatedName.value,
      gender: genderValues[unlocatedGenderIndex.value],
      isAlive: true,
      userBindingPolicy: 'OPTIONAL'
    })
    unlocatedName.value = ''
    uni.showToast({ title: '成员已暂存', icon: 'success' })
    await loadData()
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
    operationError.value = '请选择待定位成员和基准成员。'
    return
  }
  placingMember.value = true
  operationError.value = ''
  try {
    const addType = relationValues[placeRelationIndex.value]
    await placeExistingMember(familyId.value, {
      baseMemberId: base.memberId,
      memberId: member.memberId,
      addType,
      relationship: {
        parentLinkType: addType === 'ADD_SPOUSE' ? undefined : 'PRIMARY'
      }
    })
    uni.showToast({ title: '成员已放入家谱', icon: 'success' })
    placeMemberIndex.value = 0
    placeBaseIndex.value = 0
    await loadData()
  } catch (error) {
    operationError.value = apiErrorMessage(error, '成员定位失败。')
  } finally {
    placingMember.value = false
  }
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
    content: '删除后家谱结构会立即更新，确定继续吗？',
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

function openFamilyOverview() {
  uni.redirectTo({ url: `/pages/family/detail?familyId=${encodeURIComponent(familyId.value)}` })
}

onLoad((options) => {
  familyId.value = String(options?.familyId || '')
  initialBaseMemberId.value = String(options?.baseMemberId || '')
  const requestedType = String(options?.addType || '') as RelationshipAddType
  initialAddType.value = relationValues.includes(requestedType) ? requestedType : ''
  const requestedSection = String(options?.section || '')
  if (requestedSection === 'locate' || requestedSection === 'relations') {
    manageSection.value = requestedSection
  }
  if (!familyId.value) {
    loadError.value = '缺少家庭信息。'
    return
  }
  loadData()
})
</script>

<style scoped>
.manage-page {
  background:
    radial-gradient(circle at 92% 0%, rgba(216, 175, 104, 0.14), transparent 260rpx),
    radial-gradient(circle at 0% 18%, rgba(24, 54, 83, 0.08), transparent 300rpx);
}

.manage-tabs {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12rpx;
  margin-bottom: 22rpx;
  border: 1rpx solid rgba(255, 255, 255, 0.74);
  border-radius: 28rpx;
  background:
    linear-gradient(135deg, rgba(255, 255, 255, 0.78) 0%, rgba(244, 250, 247, 0.76) 100%);
  padding: 10rpx;
  box-shadow: 0 12rpx 32rpx rgba(24, 54, 83, 0.06);
}

.manage-tabs button {
  min-height: 72rpx;
  margin: 0;
  border: 0;
  border-radius: 20rpx;
  background: transparent;
  color: var(--tree-text-secondary);
  padding: 10rpx 8rpx;
  font-size: 24rpx;
  line-height: 1.3;
}

.manage-tabs button::after {
  border: 0;
}

.manage-tabs button.active {
  background: linear-gradient(135deg, var(--tree-primary) 0%, var(--tree-green) 100%);
  color: #fff;
  font-weight: 800;
  box-shadow: 0 12rpx 26rpx rgba(24, 54, 83, 0.16);
}

.field-picker,
.tree-input,
.tree-textarea {
  margin-top: 18rpx;
}

.field-picker {
  min-height: 88rpx;
  padding: 0 24rpx;
  border: 1px solid var(--tree-border);
  border-radius: 20rpx;
  background: rgba(255, 255, 255, 0.94);
  color: var(--tree-text);
  box-shadow: 0 6rpx 18rpx rgba(24, 54, 83, 0.035);
  line-height: 88rpx;
}

.relation-row {
  display: flex;
  align-items: center;
  gap: 16rpx;
}

.relation-row {
  justify-content: space-between;
  margin-top: 14rpx;
  border: 1rpx solid rgba(226, 232, 240, 0.82);
  border-radius: 22rpx;
  background:
    linear-gradient(135deg, rgba(255, 255, 255, 0.96) 0%, rgba(248, 250, 252, 0.88) 100%);
  padding: 20rpx;
  box-shadow: 0 8rpx 22rpx rgba(24, 54, 83, 0.04);
}

.relation-actions {
  display: flex;
  flex-direction: column;
  gap: 10rpx;
}

.relation-row:last-child {
  border-bottom: 1rpx solid rgba(226, 232, 240, 0.82);
}

.relation-copy {
  display: flex;
  min-width: 0;
  flex: 1;
  flex-direction: column;
  gap: 6rpx;
}

.relation-title {
  color: var(--tree-text);
  font-size: 28rpx;
  font-weight: 800;
}
</style>
