<template>
  <view class="archive-page member-edit-page">
    <MiniBackHome />

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
            {{ form.name || '成员档案' }} · 修改姓名、性别、出生年份、简介和健在状态
          </text>
        </view>
        <view class="archive-surname-stamp">{{ surnameLetter }}</view>
      </view>

      <view v-if="currentMember" class="edit-context">
        <text class="archive-chip">{{ memberStatusText(currentMember.status) }}</text>
        <text class="archive-chip">{{ bindingPolicyText(currentMember.userBindingPolicy) }}</text>
        <text v-if="currentMember.boundUserId" class="archive-chip">已绑定账号</text>
      </view>

      <view class="archive-form-panel">
        <view class="archive-section-head">
          <text class="archive-section-title">成员档案</text>
          <text class="archive-section-subtitle">字段变更会同步到成员名册与家谱节点</text>
        </view>
        <text class="tree-field-label">成员姓名</text>
        <input v-model.trim="form.name" class="tree-input" maxlength="80" placeholder="请输入成员姓名" />

        <text class="tree-field-label">性别</text>
        <picker :range="genderLabels" :value="genderIndex" @change="onGenderChange">
          <view class="field-picker">{{ genderLabels[genderIndex] }}</view>
        </picker>

        <text class="tree-field-label">出生年份</text>
        <input v-model="birthYearInput" class="tree-input" type="number" placeholder="可选，如 1988" />

        <label class="switch-row">
          <text>目前健在</text>
          <switch :checked="form.isAlive" color="#163353" @change="onAliveChange" />
        </label>

        <text class="tree-field-label">成员简介</text>
        <textarea v-model.trim="form.description" class="tree-textarea" maxlength="500" placeholder="可选" />

        <text v-if="actionError" class="tree-field-error">{{ actionError }}</text>
        <MiniButton class="edit-action" :loading="saving" :disabled="saving" @click="saveMember">
          保存成员资料
        </MiniButton>
      </view>

      <view v-if="canDeleteCurrentMember" class="archive-danger-panel">
        <text class="archive-danger-title">删除成员节点</text>
        <text class="archive-danger-desc">
          只删除家谱节点，不删除平台账号。有后代、父母或配偶关系时不能直接删除；创建者和管理员身份亦不可删除。
        </text>
        <MiniButton
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
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
import { canDeleteMember } from '@/features/family/memberListActions'
import { useSessionStore } from '@/stores/session'
import type { FamilyDetail, FamilyMember, Gender } from '@/types/api'
import { normalizeText, optionalText, validateTextFields } from '@/utils/inputValidation'

const session = useSessionStore()
const familyId = ref('')
const memberId = ref('')
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
const genderLabels = ['男', '女', '未知']
const genderValues: Gender[] = ['MALE', 'FEMALE', 'UNKNOWN']

const form = reactive({
  name: '',
  gender: 'UNKNOWN' as Gender,
  isAlive: true,
  description: ''
})

const canManageFamily = computed(() =>
  family.value?.role === 'FOUNDER' || family.value?.role === 'FAMILY_ADMIN'
)

const canDeleteCurrentMember = computed(() =>
  currentMember.value ? canDeleteMember(currentMember.value, canManageFamily.value) : false
)

const surnameLetter = computed(() => {
  const fromName = form.name.trim().slice(0, 1)
  if (fromName) return fromName
  return family.value?.familySurname.slice(0, 1) || '员'
})

function memberStatusText(status?: string) {
  return ({ ACTIVE: '活跃', DISABLED: '已停用', DELETED: '已删除' } as Record<string, string>)[status || ''] || '未知状态'
}

function bindingPolicyText(policy?: string) {
  return ({ OPTIONAL: '可选绑定', REQUIRED: '需绑定', NOT_REQUIRED: '无需绑定' } as Record<string, string>)[policy || ''] || '未知策略'
}

function routePath() {
  return `/pages/family/member-edit?familyId=${encodeURIComponent(familyId.value)}&memberId=${encodeURIComponent(memberId.value)}`
}

function onGenderChange(event: { detail: { value: number | string } }) {
  genderIndex.value = Number(event.detail.value) || 0
  form.gender = genderValues[genderIndex.value] || 'UNKNOWN'
}

function onAliveChange(event: Event) {
  const detail = (event as Event & { detail?: { value?: boolean | string | number } }).detail
  form.isAlive = Boolean(detail?.value)
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
  try {
    const [member, familyDetail] = await Promise.all([
      getFamilyMember(familyId.value, memberId.value),
      getFamilyDetail(familyId.value)
    ])
    family.value = familyDetail
    currentMember.value = member
    form.name = member.name
    form.gender = (member.gender || 'UNKNOWN') as Gender
    genderIndex.value = Math.max(0, genderValues.indexOf(form.gender))
    birthYearInput.value = member.birthYear ? String(member.birthYear) : ''
    form.isAlive = member.isAlive !== false
    form.description = member.description || ''
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
    await updateFamilyMember(familyId.value, memberId.value, {
      name: normalizeText(form.name),
      gender: form.gender,
      birthYear: yearValue(birthYearInput.value),
      isAlive: form.isAlive,
      description: optionalText(form.description)
    })
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
  uni.showModal({
    title: '删除成员',
    content: `确定删除“${form.name || '该成员'}”吗？删除后不能在列表中继续查看。`,
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
  } finally {
    deleting.value = false
  }
}

onLoad((options) => {
  familyId.value = String(options?.familyId || '')
  memberId.value = String(options?.memberId || '')
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
</style>
