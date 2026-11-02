<template>
  <view class="archive-page member-edit-page">
    <MiniBackHome :return-url="returnUrl" :return-label="returnLabel" />

    <view v-if="!authChecked" class="edit-state archive-form-panel">
      <MiniEmptyState symbol="…" title="正在确认登录状态" description="请稍候..." />
    </view>

    <view v-else-if="loading" class="edit-state archive-form-panel">
      <MiniEmptyState symbol="…" title="正在加载成员" description="正在读取成员资料..." />
    </view>

    <view v-else-if="loadError" class="edit-state archive-form-panel">
      <MiniNotice tone="warm" title="加载失败">{{ loadError }}</MiniNotice>
      <MiniButton variant="secondary" class="edit-action" @click="loadMember">重新加载</MiniButton>
    </view>

    <template v-else>
      <view class="edit-head archive-page-head">
        <view>
          <text class="archive-kicker">Member Record</text>
          <text class="archive-title">编辑成员资料</text>
          <text class="archive-subtitle">
            {{ form.name || '成员档案' }} · 修改姓名、性别、生卒年份、婚姻展示与简介
          </text>
        </view>
        <view class="archive-surname-stamp">{{ surnameLetter }}</view>
      </view>

      <view v-if="currentMember" class="edit-context">
        <text class="archive-chip">{{ memberStatusText(currentMember.status) }}</text>
        <text class="archive-chip">{{ bindingPolicyText(currentMember.userBindingPolicy) }}</text>
        <text v-if="currentMember.boundUserId" class="archive-chip">已绑定账号</text>
      </view>

      <view v-if="canEditCurrentMember" class="archive-form-panel">
        <view class="archive-section-head">
          <text class="archive-section-title">成员档案</text>
          <text class="archive-section-subtitle">字段变更会同步到家庭树成员与树状家谱节点</text>
        </view>
        <text class="tree-field-label">成员姓名</text>
        <input v-model.trim="form.name" class="tree-input" maxlength="80" placeholder="请输入成员姓名" />

        <text class="tree-field-label">性别</text>
        <picker :range="genderLabels" :value="genderIndex" @change="onGenderChange">
          <view class="field-picker">{{ genderLabels[genderIndex] }}</view>
        </picker>

        <view class="portrait-preview" aria-label="头像实时预览">
          <view class="portrait-preview-frame">
            <image
              :key="portraitPreviewUrl"
              class="portrait-preview-image"
              :src="portraitPreviewUrl"
              mode="aspectFit"
            />
          </view>
          <view class="portrait-preview-copy">
            <text class="portrait-preview-title">头像实时预览</text>
            <text class="portrait-preview-desc">选择完整形象后立即更新，保存后同步到家庭树。</text>
          </view>
        </view>

        <text class="tree-field-label">节点形象</text>
        <view class="portrait-mode-tabs" aria-label="选择节点形象类型">
          <view class="portrait-mode-tab" :class="{ selected: portraitCatalogMode === 'base' }" @tap="setPortraitCatalogMode('base')">不戴眼镜</view>
          <view class="portrait-mode-tab" :class="{ selected: portraitCatalogMode === 'glasses' }" @tap="setPortraitCatalogMode('glasses')">戴眼镜</view>
        </view>
        <view class="portrait-avatar-grid" aria-label="选择节点形象">
          <view class="portrait-avatar-option portrait-avatar-auto" :class="{ selected: form.portraitAvatarStyle === 'AUTO' }" @tap="selectPortraitAvatar('AUTO')">
            <text>自动</text>
          </view>
          <view v-for="option in portraitOptions" :key="option.style" class="portrait-avatar-option" :class="{ selected: form.portraitAvatarStyle === option.style }" @tap="selectPortraitAvatar(option.style)">
            <image class="portrait-avatar-option-image" :src="option.url" mode="aspectFit" />
          </view>
        </view>
        <text class="tree-field-tip">图册会随性别和年龄段自动筛选；“自动”会继续由系统为该成员搭配。</text>

        <template v-if="spouseRelationship && isSpouseMemberNode && canManageFamily">
          <text class="tree-field-label">婚姻展示身份</text>
          <picker :range="spouseDisplayLabels" :value="spouseDisplayIndex" @change="onSpouseDisplayChange">
            <view class="field-picker">{{ spouseDisplayLabels[spouseDisplayIndex] }}</view>
          </picker>
          <text class="tree-field-tip">用于家庭树和家谱导出；选择前妻、前夫、现任妻或现任夫时会显示对应标记。</text>
        </template>

        <template v-if="childCoParentOptions.length && canManageFamily">
          <text class="tree-field-label">另一位实际家长</text>
          <picker :range="childCoParentLabels" :value="childCoParentIndex" @change="onChildCoParentChange">
            <view class="field-picker">{{ childCoParentLabels[childCoParentIndex] }}</view>
          </picker>
          <text class="tree-field-tip">从已记录父或母的前任、现任配偶中选择。保存后会补齐孩子的另一条亲子关系。</text>
        </template>

        <text class="tree-field-label">出生年份</text>
        <input v-model="birthYearInput" class="tree-input" type="number" placeholder="可选，如 1988" />

        <label class="switch-row">
          <text>目前健在</text>
          <switch :checked="form.isAlive" color="#163353" @change="onAliveChange" />
        </label>

        <template v-if="!form.isAlive">
          <text class="tree-field-label">离世年份</text>
          <input v-model="deathYearInput" class="tree-input" type="number" placeholder="可选，如 2024" />
        </template>

        <text class="tree-field-label">{{ form.isAlive ? '成员简介' : '亲人简介 / 离世说明' }}</text>
        <textarea
          v-model.trim="form.description"
          class="tree-textarea"
          maxlength="500"
          :placeholder="form.isAlive ? '可选' : '可记录离世原因、生平简述或家人纪念文字'"
        />

        <label v-if="!form.isAlive" class="switch-row memorial-visibility-row">
          <view>
            <text>向家庭成员展示亲人简介</text>
            <text class="switch-row-desc">关闭后，仅创建人和管理员可查看与编辑这段简介。</text>
          </view>
          <switch :checked="form.memorialVisible" color="#A83B2D" @change="onMemorialVisibleChange" />
        </label>

        <text v-if="actionError" class="tree-field-error">{{ actionError }}</text>
        <MiniButton class="edit-action" :loading="saving" :disabled="saving" @click="saveMember">
          保存成员资料
        </MiniButton>
      </view>

      <view v-else class="archive-form-panel">
        <MiniNotice tone="security" title="当前为只读查看">
          仅成员本人、家庭创建者或管理员可以修改这份成员资料。
        </MiniNotice>
      </view>

      <view v-if="canDeleteCurrentMember || memberDeleteRestriction" class="archive-danger-panel">
        <text class="archive-danger-title">{{ canDeleteCurrentMember ? '删除成员节点' : '暂不能删除成员' }}</text>
        <text class="archive-danger-desc">{{ memberDeleteRestriction || '只删除家庭树节点，不删除平台账号。父母或配偶关系会随成员删除同步解除。' }}</text>
        <MiniButton
          v-if="canDeleteCurrentMember"
          class="edit-action"
          variant="danger"
          :loading="deleting"
          :disabled="deleting"
          @click="confirmDeleteMember"
        >
          删除该成员
        </MiniButton>
      </view>
    </template>
  </view>
</template>

<script setup lang="ts">
import { onLoad, onShow } from '@dcloudio/uni-app'
import { computed, reactive, ref } from 'vue'

import { apiErrorMessage } from '@/api/client'
import { getFamilyDetail } from '@/api/families'
import { deleteFamilyMember, getFamilyMember, updateFamilyMember } from '@/api/members'
import { placeExistingMember, updateRelationship } from '@/api/relationships'
import { getPrivateTree } from '@/api/tree'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
import { canDeleteMember } from '@/features/family/memberListActions'
import { resolveAvatarAgeBand } from '@/features/family-tree/avatarAgeBand'
import {
  resolveFemalePortraitVariant,
  resolvePortraitAvatarStyle,
  resolvePortraitVariant,
  resolveRegionalPortraitAsset,
  resolveSelectedPortraitAsset
} from '@/features/family-tree/portraitCatalog'
import { spouseDisplayOptions, spouseRelationLabel } from '@/features/family-tree/spouseRelationLabel'
import { useSessionStore } from '@/stores/session'
import type { FamilyDetail, FamilyMember, Gender, PortraitAvatarStyle, TreeEdge } from '@/types/api'
import { normalizeText, optionalText, validateTextFields } from '@/utils/inputValidation'

const session = useSessionStore()
const familyId = ref('')
const memberId = ref('')
const returnTo = ref('')
const family = ref<FamilyDetail | null>(null)
const currentMember = ref<FamilyMember | null>(null)
const authChecked = ref(false)
const loading = ref(false)
const saving = ref(false)
const deleting = ref(false)
const loadError = ref('')
const actionError = ref('')
const genderIndex = ref(2)
const birthYearInput = ref('')
const deathYearInput = ref('')
const spouseRelationship = ref<TreeEdge | null>(null)
const isSpouseMemberNode = ref(false)
const memberHasChildren = ref(false)
const spouseDisplayIndex = ref(0)
const childParentMemberId = ref<number | null>(null)
const childCoParentOptions = ref<Array<{ memberId: number; label: string }>>([])
const childCoParentIndex = ref(0)
const genderLabels = ['男', '女', '未知']
const genderValues: Gender[] = ['MALE', 'FEMALE', 'UNKNOWN']
const form = reactive({
  name: '',
  gender: 'UNKNOWN' as Gender,
  portraitAvatarStyle: 'AUTO' as PortraitAvatarStyle,
  isAlive: true,
  description: '',
  memorialVisible: true
})
const shouldKeepReturnTarget = computed(() => ['tree', 'health'].includes(returnTo.value))
const memberViewUrl = computed(() => {
  const query = `familyId=${encodeURIComponent(familyId.value)}&memberId=${encodeURIComponent(memberId.value)}`
  return `/pages/family/member-view?${query}${shouldKeepReturnTarget.value ? `&returnTo=${encodeURIComponent(returnTo.value)}` : ''}`
})
const returnUrl = computed(() => returnTo.value === 'tree'
  ? `/pages/family/private-tree?familyId=${encodeURIComponent(familyId.value)}`
  : returnTo.value === 'health'
    ? `/pages/family/health?familyId=${encodeURIComponent(familyId.value)}`
    : memberViewUrl.value)
const returnLabel = computed(() => returnTo.value === 'tree' ? '返回家庭树' : returnTo.value === 'health' ? '返回家庭健康' : '返回成员资料')

const canManageFamily = computed(() =>
  family.value?.role === 'FOUNDER' || family.value?.role === 'FAMILY_ADMIN'
)

const memberDeleteRestriction = computed(() => {
  if (!currentMember.value || !canManageFamily.value) return ''
  if (currentMember.value.memberId === family.value?.currentFounderMemberId) return '当前创建人需先完成转让，不能删除。'
  return memberHasChildren.value ? '该成员已有子女，不能删除。' : ''
})
const canDeleteCurrentMember = computed(() =>
  Boolean(currentMember.value)
    && canDeleteMember(currentMember.value!, canManageFamily.value)
    && !memberDeleteRestriction.value
)
const isCurrentUserMember = computed(() => Boolean(currentMember.value?.boundUserId && currentMember.value.boundUserId === session.user?.id))
const canEditCurrentMember = computed(() => canManageFamily.value || isCurrentUserMember.value)
const spouseDisplayOptionsForMember = computed(() => spouseDisplayOptions(currentMember.value?.gender))
const spouseDisplayLabels = computed(() => spouseDisplayOptionsForMember.value.map((option) => option.label))
const childCoParentLabels = computed(() => ['暂不补充另一位家长', ...childCoParentOptions.value.map((option) => option.label)])
const surnameLetter = computed(() => {
  const fromName = form.name.trim().slice(0, 1)
  if (fromName) return fromName
  return family.value?.familySurname.slice(0, 1) || '员'
})
const portraitCatalogMode = ref<'base' | 'glasses'>('base')
const portraitGender = computed(() => form.gender === 'FEMALE' ? 'female' : 'male')
const portraitAgeBand = computed(() => resolveAvatarAgeBand(
  currentMember.value?.birthDate,
  form.isAlive ? undefined : currentMember.value?.deathDate
))
const portraitOptions = computed(() => Array.from({ length: 12 }, (_, index) => {
  const variant = index + 1
  const style = `${portraitCatalogMode.value === 'glasses' ? 'GLASSES' : 'BASE'}_${variant}` as PortraitAvatarStyle
  return {
    style,
    url: resolveSelectedPortraitAsset({ gender: portraitGender.value, ageBand: portraitAgeBand.value, style })
  }
}))
const portraitPreviewUrl = computed(() => {
  const member = currentMember.value
  const selected = resolveSelectedPortraitAsset({
    gender: portraitGender.value,
    ageBand: portraitAgeBand.value,
    style: form.portraitAvatarStyle
  })
  return selected || resolveRegionalPortraitAsset({
    theme: 'north', gender: portraitGender.value, ageBand: portraitAgeBand.value, memberId: member?.memberId
  })
})

function memberStatusText(status?: string) {
  return ({ ACTIVE: '活跃', DISABLED: '已停用', DELETED: '已删除' } as Record<string, string>)[status || ''] || '未知状态'
}

function bindingPolicyText(policy?: string) {
  return ({ OPTIONAL: '可选绑定', REQUIRED: '需绑定', NOT_REQUIRED: '无需绑定' } as Record<string, string>)[policy || ''] || '未知策略'
}

function routePath() {
  const query = `familyId=${encodeURIComponent(familyId.value)}&memberId=${encodeURIComponent(memberId.value)}`
  return `/pages/family/member-edit?${query}${shouldKeepReturnTarget.value ? `&returnTo=${encodeURIComponent(returnTo.value)}` : ''}`
}

function onGenderChange(event: { detail: { value: number | string } }) {
  genderIndex.value = Number(event.detail.value) || 0
  form.gender = genderValues[genderIndex.value] || 'UNKNOWN'
  form.portraitAvatarStyle = 'AUTO'
}

function setPortraitCatalogMode(mode: 'base' | 'glasses') {
  portraitCatalogMode.value = mode
}

function selectPortraitAvatar(style: PortraitAvatarStyle) {
  form.portraitAvatarStyle = style
}

function onSpouseDisplayChange(event: { detail: { value: number | string } }) {
  spouseDisplayIndex.value = Number(event.detail.value) || 0
}

function onChildCoParentChange(event: { detail: { value: number | string } }) {
  childCoParentIndex.value = Number(event.detail.value) || 0
}

function onAliveChange(event: Event) {
  const detail = (event as Event & { detail?: { value?: boolean | string | number } }).detail
  form.isAlive = Boolean(detail?.value)
  if (form.isAlive) {
    form.memorialVisible = true
  }
}

function onMemorialVisibleChange(event: Event) {
  const detail = (event as Event & { detail?: { value?: boolean | string | number } }).detail
  form.memorialVisible = Boolean(detail?.value)
}

function yearValue(value: string) {
  const parsed = Number(value)
  return Number.isInteger(parsed) && parsed > 0 ? parsed : undefined
}

async function loadMember() {
  if (!familyId.value || !memberId.value) {
    loadError.value = '缺少成员信息。'
    return
  }
  loading.value = true
  loadError.value = ''
  actionError.value = ''
  spouseRelationship.value = null
  isSpouseMemberNode.value = false
  memberHasChildren.value = false
  childParentMemberId.value = null
  childCoParentOptions.value = []
  childCoParentIndex.value = 0
  try {
    const [member, familyDetail, tree] = await Promise.all([
      getFamilyMember(familyId.value, memberId.value),
      getFamilyDetail(familyId.value),
      getPrivateTree(familyId.value).catch(() => null)
    ])
    family.value = familyDetail
    currentMember.value = member
    memberHasChildren.value = Boolean(tree?.edges.some((edge) =>
      edge.relationshipType === 'PARENT_CHILD' && edge.fromMemberId === member.memberId
    ))
    form.name = member.name
    form.gender = (member.gender || 'UNKNOWN') as Gender
    genderIndex.value = Math.max(0, genderValues.indexOf(form.gender))
    const savedPortrait = resolvePortraitAvatarStyle(member.portraitAvatarStyle)
    if (savedPortrait) {
      form.portraitAvatarStyle = member.portraitAvatarStyle as PortraitAvatarStyle
      portraitCatalogMode.value = savedPortrait.mode
    } else {
      const ageBand = resolveAvatarAgeBand(member.birthDate, member.isAlive === false ? member.deathDate : undefined)
      const variant = member.gender === 'FEMALE'
        ? resolveFemalePortraitVariant(member.memberId, member.portraitHairStyle, ageBand)
        : resolvePortraitVariant(member.memberId)
      form.portraitAvatarStyle = member.portraitGlassesStyle === 'ON'
        ? `GLASSES_${variant}` as PortraitAvatarStyle
        : 'AUTO'
      portraitCatalogMode.value = member.portraitGlassesStyle === 'ON' ? 'glasses' : 'base'
    }
    birthYearInput.value = member.birthYear ? String(member.birthYear) : ''
    deathYearInput.value = member.deathYear ? String(member.deathYear) : ''
    form.isAlive = member.isAlive !== false
    form.description = member.description || ''
    form.memorialVisible = member.memorialVisible !== false
    spouseRelationship.value = (tree?.edges || []).find((edge) =>
      edge.relationshipType === 'SPOUSE' &&
      (edge.fromMemberId === member.memberId || edge.toMemberId === member.memberId)
    ) || null
    isSpouseMemberNode.value = tree?.nodes.find((node) => node.memberId === member.memberId)?.memberType === 'SPOUSE'
    const parentEdges = (tree?.edges || []).filter((edge) => edge.relationshipType === 'PARENT_CHILD' && edge.toMemberId === member.memberId)
    const parentIds = new Set(parentEdges.map((edge) => edge.fromMemberId))
    const parentEdge = parentEdges.find((edge) => (tree?.edges || []).some((spouseEdge) =>
      spouseEdge.relationshipType === 'SPOUSE' && (spouseEdge.fromMemberId === edge.fromMemberId || spouseEdge.toMemberId === edge.fromMemberId)
    ))
    if (parentEdge && parentIds.size < 2) {
      childParentMemberId.value = parentEdge.fromMemberId
      const nodeMap = new Map((tree?.nodes || []).map((node) => [node.memberId, node]))
      childCoParentOptions.value = (tree?.edges || [])
        .filter((edge) => edge.relationshipType === 'SPOUSE' && (edge.fromMemberId === parentEdge.fromMemberId || edge.toMemberId === parentEdge.fromMemberId))
        .map((edge) => {
          const candidateId = edge.fromMemberId === parentEdge.fromMemberId ? edge.toMemberId : edge.fromMemberId
          const candidate = nodeMap.get(candidateId)
          return candidate && !parentIds.has(candidateId)
            ? { memberId: candidateId, label: `${candidate.displayName}（${spouseRelationLabel(edge.relationNoteType, candidate.gender)}）` }
            : null
        })
        .filter((item): item is { memberId: number; label: string } => Boolean(item))
    }
    const matchingDisplayIndex = spouseDisplayOptionsForMember.value.findIndex(
      (option) => option.noteType === (spouseRelationship.value?.relationNoteType || '')
    )
    spouseDisplayIndex.value = matchingDisplayIndex >= 0 ? matchingDisplayIndex : 0
  } catch (error) {
    loadError.value = apiErrorMessage(error, '成员资料加载失败。')
  } finally {
    loading.value = false
  }
}

async function saveMember() {
  const validationMessage = validateTextFields([
    { value: form.name, label: '成员姓名', kind: 'name', required: true, maxLength: 80 },
    { value: form.description, label: '成员说明', kind: 'multiLine', maxLength: 500 }
  ])
  if (validationMessage) {
    actionError.value = validationMessage
    return
  }
  saving.value = true
  actionError.value = ''
  try {
    const birthYear = yearValue(birthYearInput.value)
    const deathYear = form.isAlive ? undefined : yearValue(deathYearInput.value)
    await updateFamilyMember(familyId.value, memberId.value, {
      expectedRevision: currentMember.value?.revision,
      name: normalizeText(form.name),
      gender: form.gender,
      portraitAvatarStyle: form.portraitAvatarStyle,
      // An empty date field is the explicit API signal to clear an existing year.
      birthDate: birthYear === undefined ? '' : undefined,
      birthYear,
      deathDate: form.isAlive || deathYear === undefined ? '' : undefined,
      deathYear,
      isAlive: form.isAlive,
      description: optionalText(form.description),
      memorialVisible: form.isAlive ? true : form.memorialVisible
    })
    const selectedSpouseDisplay = spouseDisplayOptionsForMember.value[spouseDisplayIndex.value]
    if (
      spouseRelationship.value &&
      selectedSpouseDisplay &&
      selectedSpouseDisplay.noteType !== (spouseRelationship.value.relationNoteType || '')
    ) {
      await updateRelationship(familyId.value, spouseRelationship.value.relationshipId, {
        relationNoteType: selectedSpouseDisplay.noteType,
        expectedRevision: spouseRelationship.value.revision
      })
    }
    const selectedCoParent = childCoParentIndex.value > 0 ? childCoParentOptions.value[childCoParentIndex.value - 1] : null
    if (childParentMemberId.value && selectedCoParent) {
      await placeExistingMember(familyId.value, {
        baseMemberId: childParentMemberId.value,
        coParentMemberId: selectedCoParent.memberId,
        memberId: Number(memberId.value),
        addType: 'ADD_CHILD',
        relationship: { parentLinkType: 'PRIMARY' }
      })
    }
    uni.showToast({ title: '成员资料已保存', icon: 'success' })
    setTimeout(() => {
      uni.navigateBack()
    }, 350)
  } catch (error) {
    actionError.value = apiErrorMessage(error, '保存成员失败。')
  } finally {
    saving.value = false
  }
}

function confirmDeleteMember() {
  if (memberDeleteRestriction.value) {
    uni.showModal({ title: '暂不能删除', content: memberDeleteRestriction.value, showCancel: false })
    return
  }
  uni.showModal({
    title: '删除成员',
    content: `确定删除“${form.name || '该成员'}”吗？相关父母或配偶关系会一并解除。`,
    success: (result) => {
      if (result.confirm) deleteMember()
    }
  })
}

async function deleteMember() {
  deleting.value = true
  actionError.value = ''
  try {
    await deleteFamilyMember(familyId.value, memberId.value, '小程序成员资料页删除')
    uni.showToast({ title: '成员已删除', icon: 'success' })
    setTimeout(() => {
      uni.navigateBack()
    }, 350)
  } catch (error) {
    actionError.value = apiErrorMessage(error, '删除成员失败。')
    uni.showModal({
      title: '删除失败',
      content: actionError.value,
      showCancel: false
    })
  } finally {
    deleting.value = false
  }
}

onLoad((options) => {
  familyId.value = String(options?.familyId || '')
  memberId.value = String(options?.memberId || '')
  returnTo.value = String(options?.returnTo || '')
})

onShow(() => {
  authChecked.value = session.requireLogin(routePath())
  if (authChecked.value) loadMember()
})
</script>

<style scoped>
.member-edit-page {
  padding-top: 28rpx;
}

.member-edit-page :deep(.mini-back-home) {
  margin-bottom: 22rpx;
}

.edit-state :deep(.mini-notice) {
  margin-bottom: 18rpx;
}

.edit-context {
  display: flex;
  flex-wrap: wrap;
  gap: 10rpx;
  margin-bottom: 24rpx;
}

.edit-action {
  margin-top: 20rpx;
  align-self: flex-start;
}

.memorial-visibility-row {
  align-items: flex-start;
}

.switch-row-desc {
  display: block;
  margin-top: 8rpx;
  color: var(--archive-ink-soft);
  font-size: 24rpx;
  line-height: 1.7;
}

.tree-field-tip {
  display: block;
  margin-top: 10rpx;
  color: var(--archive-ink-soft);
  font-size: 23rpx;
  line-height: 1.55;
}

.portrait-mode-tabs {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10rpx;
  margin-bottom: 12rpx;
}

.portrait-mode-tab {
  padding: 15rpx 12rpx;
  border: 1rpx solid var(--archive-line, rgba(92, 74, 48, 0.2));
  border-radius: 4rpx;
  color: var(--archive-ink-soft);
  font-size: 24rpx;
  line-height: 1.2;
  text-align: center;
}

.portrait-mode-tab.selected {
  border-color: var(--archive-red, #a83b2d);
  background: rgba(168, 59, 45, 0.08);
  color: var(--archive-red, #a83b2d);
  font-weight: 700;
}

.portrait-avatar-grid {
  display: grid;
  grid-template-columns: repeat(5, minmax(0, 1fr));
  gap: 10rpx;
}

.portrait-avatar-option {
  position: relative;
  aspect-ratio: 1;
  overflow: hidden;
  border: 1rpx solid var(--archive-line, rgba(92, 74, 48, 0.2));
  border-radius: 50%;
  background: #f6eedc;
}

.portrait-avatar-option.selected {
  border: 3rpx solid var(--archive-red, #a83b2d);
}

.portrait-avatar-option-image {
  width: 100%;
  height: 100%;
}

.portrait-avatar-auto {
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 252, 245, 0.9);
  color: var(--archive-ink-soft);
  font-size: 23rpx;
}

.portrait-preview {
  display: flex;
  align-items: center;
  gap: 16rpx;
  margin: 18rpx 0 22rpx;
  padding: 14rpx 16rpx;
  border: 1rpx solid var(--archive-line, rgba(92, 74, 48, 0.2));
  border-radius: 4rpx;
  background: rgba(255, 252, 245, 0.56);
}

.portrait-preview-frame {
  position: relative;
  flex: 0 0 auto;
  width: 100rpx;
  height: 100rpx;
  overflow: hidden;
  border: 1rpx solid var(--archive-line-strong, rgba(92, 74, 48, 0.32));
  border-radius: 50%;
  background: #f6eedc;
}

.portrait-preview-image {
  width: 100%;
  height: 100%;
}

.portrait-preview-copy {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 5rpx;
}

.portrait-preview-title {
  color: var(--archive-ink, #193652);
  font-size: 27rpx;
  font-weight: 700;
}

.portrait-preview-desc {
  color: var(--archive-ink-soft);
  font-size: 22rpx;
  line-height: 1.55;
}

.portrait-style-option {
  display: flex;
  min-width: 0;
  min-height: 78rpx;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 3rpx;
  border: 1rpx solid var(--archive-line, rgba(92, 74, 48, 0.2));
  border-radius: 4rpx;
  background: rgba(255, 252, 245, 0.72);
  color: var(--archive-ink-soft);
}

.portrait-style-option.selected {
  border-color: var(--archive-red, #a83b2d);
  background: rgba(168, 59, 45, 0.08);
  color: var(--archive-red, #a83b2d);
}

.portrait-style-symbol {
  font-family: Georgia, 'Songti SC', serif;
  font-size: 27rpx;
  line-height: 1;
}

.portrait-style-name {
  font-size: 19rpx;
  line-height: 1.2;
}

.portrait-glasses-row {
  margin-top: 0;
}

</style>
