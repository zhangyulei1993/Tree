<template>
  <view class="tree-page settings-page">
    <MiniBackHome />
    <FamilyContextHeader
      v-if="family"
      :family-name="family.familyName"
      section="公开展示与权限"
      subtitle="管理公开信息、家庭角色和高风险操作"
      :role-label="roleText(family.role)"
      @back="openFamilyOverview"
    />

    <MiniCard v-if="loading">
      <MiniEmptyState symbol="…" title="正在加载" description="正在读取家庭设置..." />
    </MiniCard>
    <MiniCard v-else-if="loadError">
      <MiniNotice tone="warm" title="加载失败">{{ loadError }}</MiniNotice>
      <MiniButton variant="secondary" @click="loadData">重新加载</MiniButton>
    </MiniCard>
    <MiniCard v-else-if="!canManage">
      <MiniNotice tone="warm" title="无管理权限">只有家庭创建者或家庭管理员可以进入家庭设置。</MiniNotice>
    </MiniCard>

    <template v-else-if="family">
      <MiniNotice v-if="family.status === 'DISSOLVED'" tone="warm" title="家庭已解散，等待恢复">
        该家庭已停止对外服务，设置暂不可编辑。请等待平台管理员恢复家庭后再继续操作。
      </MiniNotice>
      <MiniCard v-if="family.status === 'DISSOLVED'">
        <MiniEmptyState
          symbol="⏸"
          title="家庭已解散"
          description="当前仅可查看家庭概览，公开展示、角色与高风险操作需恢复后可用。"
        />
      </MiniCard>

      <MiniNotice v-else-if="family.status === 'DISSOLUTION_PENDING'" tone="warm" title="家庭解散申请审核中">
        审核完成或取消申请前，家庭资料、公开展示、角色和创建者转让暂不可调整。
      </MiniNotice>

      <template v-if="familyOperational">
      <view class="settings-tabs">
        <button :class="{ active: settingsSection === 'public' }" @click="settingsSection = 'public'">公开展示</button>
        <button :class="{ active: settingsSection === 'roles' }" @click="settingsSection = 'roles'">角色权限</button>
        <button :class="{ active: settingsSection === 'security' }" @click="settingsSection = 'security'">高风险操作</button>
      </view>
      <MiniNotice v-if="operationError" tone="warm" title="操作失败">
        {{ operationError }}
      </MiniNotice>

      <MiniCard v-if="settingsSection === 'security' && family.role !== 'FOUNDER'">
        <MiniNotice tone="security" title="仅创建者可操作">
          创建者转让和家庭解散属于高风险操作，家庭管理员只能查看其他设置。
        </MiniNotice>
      </MiniCard>

      <MiniCard v-if="familyOperational && settingsSection === 'public'">
        <MiniSectionHeader title="公开信息" subtitle="这些信息仅在家庭获准公开后对访客展示" />
        <input v-model.trim="publicForm.publicContactName" class="tree-input" maxlength="80" placeholder="公开联系人" />
        <input v-model.trim="publicForm.publicContactPhone" class="tree-input" maxlength="30" placeholder="公开联系电话" />
        <input v-model.trim="publicForm.publicContactWechat" class="tree-input" maxlength="80" placeholder="公开微信" />
        <textarea v-model.trim="publicForm.publicContactNote" class="tree-textarea" maxlength="300" placeholder="公开联系说明" />
        <label class="switch-row">
          <text>公开联系方式</text>
          <switch :checked="publicForm.publicContactVisible" color="#2F6B57" @change="onPublicContactVisibleChange" />
        </label>
        <label class="switch-row">
          <text>允许公开搜索</text>
          <switch :checked="publicForm.searchable" color="#2F6B57" @change="onSearchableChange" />
        </label>
        <MiniButton
          variant="secondary"
          :loading="savingPublicInfo"
          :disabled="savingPublicInfo"
          @click="savePublicInfo"
        >
          保存公开信息
        </MiniButton>
      </MiniCard>

      <MiniCard v-if="familyOperational && settingsSection === 'public'">
        <MiniSectionHeader title="公开展示申请" :subtitle="`当前状态：${publicStatusText(family.publicDisplayStatus)}`" />
        <template v-if="family.publicDisplayStatus === 'APPROVED'">
          <MiniNotice tone="security" title="家庭当前已公开">
            公开家庭主页和公开家谱可被访客访问。家庭主动关闭会立即生效，无需再次等待审核。
          </MiniNotice>
          <textarea v-model.trim="takeDownReason" class="tree-textarea" maxlength="500" placeholder="关闭说明（可选）" />
          <MiniButton variant="danger" :loading="takingDown" :disabled="takingDown" @click="confirmClosePublic">
            关闭公开展示
          </MiniButton>
        </template>
        <textarea
          v-else-if="!pendingPublicApplication"
          v-model.trim="publicReason"
          class="tree-textarea"
          maxlength="500"
          placeholder="申请说明（可选）"
        />
        <MiniButton
          v-if="!pendingPublicApplication && family.publicDisplayStatus !== 'APPROVED'"
          :loading="publicSubmitting"
          :disabled="publicSubmitting"
          @click="submitPublic"
        >
          提交公开申请
        </MiniButton>
        <MiniNotice v-if="pendingPublicApplication" tone="warm" title="申请待审核">
          {{ pendingPublicApplication.reason || '管理员审核通过后，家庭主页和公开家谱才会对外开放。' }}
        </MiniNotice>
        <MiniButton
          v-if="pendingPublicApplication"
          variant="secondary"
          :loading="publicSubmitting"
          :disabled="publicSubmitting"
          @click="confirmCancelPublic"
        >
          取消待审核申请
        </MiniButton>
        <view v-for="item in publicApplications.slice(0, 5)" :key="item.applicationId" class="record-row">
          <text>{{ applicationStatusText(item.status) }}</text>
          <text class="tree-muted">{{ formatDate(item.createdAt) }}</text>
        </view>
      </MiniCard>

      <MiniCard v-if="familyOperational && settingsSection === 'roles'">
        <MiniSectionHeader title="家庭管理员" subtitle="创建者可授予或取消家庭管理员角色" />
        <MiniNotice v-if="family.role !== 'FOUNDER'" tone="security" title="只读">
          只有家庭创建者可以调整管理员角色。
        </MiniNotice>
        <view v-for="member in boundMembers" :key="member.memberId" class="member-row">
          <view class="member-copy">
            <text class="member-name">{{ member.name }}</text>
            <text class="tree-muted">{{ roleText(member.boundFamilyRole) }}</text>
          </view>
          <MiniButton
            v-if="family.role === 'FOUNDER' && member.boundFamilyRole === 'MEMBER'"
            size="sm"
            variant="secondary"
            @click="confirmRoleChange(member, true)"
          >
            设为管理员
          </MiniButton>
          <MiniButton
            v-if="family.role === 'FOUNDER' && member.boundFamilyRole === 'FAMILY_ADMIN'"
            size="sm"
            variant="secondary"
            @click="confirmRoleChange(member, false)"
          >
            取消管理员
          </MiniButton>
        </view>
      </MiniCard>

      <MiniCard v-if="family.role === 'FOUNDER' && familyOperational && settingsSection === 'security'">
        <MiniSectionHeader title="转让创建者" subtitle="目标必须是已绑定账号的活跃家庭成员" />
        <template v-if="currentTransfer">
          <MiniNotice tone="warm" title="转让申请待审核">
            目标成员：{{ memberName(currentTransfer.toMemberId) }}。后台审核通过前可取消。
          </MiniNotice>
          <MiniButton variant="secondary" @click="confirmCancelTransfer">取消转让申请</MiniButton>
        </template>
        <template v-else>
          <picker :range="transferTargetNames" :value="transferTargetIndex" @change="transferTargetIndex = Number($event.detail.value)">
            <view class="field-picker">{{ selectedTransferTarget?.name || '选择接任成员' }}</view>
          </picker>
          <textarea v-model.trim="transferReason" class="tree-textarea" maxlength="500" placeholder="转让原因（可选）" />
          <MiniButton :disabled="!selectedTransferTarget" @click="confirmCreateTransfer">发起转让申请</MiniButton>
        </template>
      </MiniCard>

      <MiniCard v-if="family.role === 'FOUNDER' && settingsSection === 'security'" class="danger-zone">
        <MiniSectionHeader title="解散家庭" subtitle="高风险操作，审核通过后家庭将停止使用" />
        <template v-if="currentDissolution">
          <MiniNotice tone="warm" title="解散申请待审核">
            {{ currentDissolution.requestReason || '后台正在审核该申请。' }}
          </MiniNotice>
          <MiniButton variant="secondary" @click="confirmCancelDissolution">取消解散申请</MiniButton>
        </template>
        <template v-else>
          <textarea v-model.trim="dissolutionReason" class="tree-textarea" maxlength="500" placeholder="请填写解散原因" />
          <MiniButton variant="secondary" @click="confirmCreateDissolution">申请解散家庭</MiniButton>
        </template>
      </MiniCard>
      </template>
    </template>
  </view>
</template>

<script setup lang="ts">
import { onLoad, onShow } from '@dcloudio/uni-app'
import { computed, reactive, ref } from 'vue'

import { ApiError, apiErrorMessage, isRealApiMode } from '@/api/client'
import {
  cancelDissolution,
  createDissolution,
  getCurrentDissolution
} from '@/api/dissolutions'
import { setFamilyAdmin, unsetFamilyAdmin } from '@/api/familyRoles'
import { getFamilyDetail, updateFamily } from '@/api/families'
import {
  cancelFounderTransfer,
  createFounderTransfer,
  getCurrentFounderTransfer
} from '@/api/founderTransfers'
import { listFamilyMembers } from '@/api/members'
import {
  cancelPublicApplication,
  closePublicFamily,
  listPublicApplications,
  submitPublicApplication
} from '@/api/publicApplications'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniCard from '@/components/base/MiniCard.vue'
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
import MiniSectionHeader from '@/components/base/MiniSectionHeader.vue'
import FamilyContextHeader from '@/components/family/FamilyContextHeader.vue'
import { useSessionStore } from '@/stores/session'
import type {
  DissolutionRequest,
  FamilyDetail,
  FamilyMember,
  FounderTransferRequest,
  PublicApplication
} from '@/types/api'

const session = useSessionStore()
const familyId = ref('')
const family = ref<FamilyDetail | null>(null)
const members = ref<FamilyMember[]>([])
const publicApplications = ref<PublicApplication[]>([])
const currentTransfer = ref<FounderTransferRequest | null>(null)
const currentDissolution = ref<DissolutionRequest | null>(null)
const loading = ref(false)
const loadError = ref('')
const operationError = ref('')
const savingPublicInfo = ref(false)
const publicSubmitting = ref(false)
const publicReason = ref('')
const takeDownReason = ref('')
const takingDown = ref(false)
const transferTargetIndex = ref(0)
const transferReason = ref('')
const dissolutionReason = ref('')
const settingsSection = ref<'public' | 'roles' | 'security'>('public')
const publicForm = reactive({
  searchable: false,
  publicContactName: '',
  publicContactPhone: '',
  publicContactWechat: '',
  publicContactNote: '',
  publicContactVisible: false
})

const canManage = computed(() =>
  family.value?.role === 'FOUNDER' || family.value?.role === 'FAMILY_ADMIN'
)
const familyOperational = computed(() => family.value?.status === 'NORMAL')
const boundMembers = computed(() =>
  members.value.filter((member) => member.status === 'ACTIVE' && member.boundUserId)
)
const transferTargets = computed(() =>
  boundMembers.value.filter((member) => member.boundFamilyRole !== 'FOUNDER')
)
const transferTargetNames = computed(() => transferTargets.value.map((member) => member.name))
const selectedTransferTarget = computed(() => transferTargets.value[transferTargetIndex.value])
const pendingPublicApplication = computed(() =>
  publicApplications.value.find((item) => item.status === 'PENDING')
)

function switchValue(event: Event) {
  const detail = (event as Event & { detail?: { value?: boolean | string | number } }).detail
  return Boolean(detail?.value)
}

function onPublicContactVisibleChange(event: Event) {
  publicForm.publicContactVisible = switchValue(event)
}

function onSearchableChange(event: Event) {
  publicForm.searchable = switchValue(event)
}

function nullableRequest<T>(requestFactory: () => Promise<T>) {
  return requestFactory().catch((error) => {
    if (error instanceof ApiError && error.statusCode === 404) return null
    throw error
  })
}

async function loadData() {
  const route = `/pages/family/settings?familyId=${encodeURIComponent(familyId.value)}`
  if (!session.requireLogin(route)) return
  loading.value = true
  loadError.value = ''
  try {
    const familyResult = await getFamilyDetail(familyId.value)
    family.value = familyResult
    if (familyResult.role !== 'FOUNDER' && familyResult.role !== 'FAMILY_ADMIN') return
    const [memberResult, publicResult, transferResult, dissolutionResult] = await Promise.all([
      listFamilyMembers(familyId.value),
      isRealApiMode
        ? listPublicApplications(familyId.value)
        : Promise.resolve({ items: [], page: 1, pageSize: 50, total: 0 }),
      isRealApiMode
        ? nullableRequest(() => getCurrentFounderTransfer(familyId.value))
        : Promise.resolve(null),
      isRealApiMode
        ? nullableRequest(() => getCurrentDissolution(familyId.value))
        : Promise.resolve(null)
    ])
    members.value = memberResult
    publicApplications.value = publicResult.items
    currentTransfer.value = transferResult
    currentDissolution.value = dissolutionResult
    fillPublicForm(familyResult)
  } catch (error) {
    loadError.value = apiErrorMessage(error, '家庭设置加载失败。')
  } finally {
    loading.value = false
  }
}

function fillPublicForm(value: FamilyDetail) {
  publicForm.searchable = value.searchable
  publicForm.publicContactName = value.publicContactName || ''
  publicForm.publicContactPhone = value.publicContactPhone || ''
  publicForm.publicContactWechat = value.publicContactWechat || ''
  publicForm.publicContactNote = value.publicContactNote || ''
  publicForm.publicContactVisible = value.publicContactVisible
}

async function savePublicInfo() {
  if (savingPublicInfo.value) return
  savingPublicInfo.value = true
  operationError.value = ''
  try {
    family.value = await updateFamily(familyId.value, { ...publicForm })
    uni.showToast({ title: '公开信息已保存', icon: 'success' })
  } catch (error) {
    operationError.value = apiErrorMessage(error, '保存公开信息失败。')
  } finally {
    savingPublicInfo.value = false
  }
}

async function submitPublic() {
  publicSubmitting.value = true
  operationError.value = ''
  try {
    await submitPublicApplication(familyId.value, publicReason.value)
    publicReason.value = ''
    uni.showToast({ title: '公开申请已提交', icon: 'success' })
    await loadData()
  } catch (error) {
    operationError.value = apiErrorMessage(error, '提交公开申请失败。')
  } finally {
    publicSubmitting.value = false
  }
}

function confirmCancelPublic() {
  const application = pendingPublicApplication.value
  if (!application) return
  uni.showModal({
    title: '取消公开申请',
    content: '确定取消当前待审核的公开展示申请吗？',
    success: async (result) => {
      if (!result.confirm) return
      try {
        await cancelPublicApplication(familyId.value, application.applicationId, '家庭管理主动取消')
        await loadData()
      } catch (error) {
        operationError.value = apiErrorMessage(error, '取消公开申请失败。')
      }
    }
  })
}

function confirmClosePublic() {
  uni.showModal({
    title: '关闭公开展示',
    content: '关闭后，公开主页和公开家谱将立即不可访问。重新公开时需要再次提交审核。确定继续吗？',
    success: async (result) => {
      if (!result.confirm) return
      takingDown.value = true
      operationError.value = ''
      try {
        await closePublicFamily(familyId.value, takeDownReason.value)
        takeDownReason.value = ''
        uni.showToast({ title: '公开展示已关闭', icon: 'success' })
        await loadData()
      } catch (error) {
        operationError.value = apiErrorMessage(error, '关闭公开展示失败。')
      } finally {
        takingDown.value = false
      }
    }
  })
}

function confirmRoleChange(member: FamilyMember, setAdmin: boolean) {
  uni.showModal({
    title: setAdmin ? '设置家庭管理员' : '取消家庭管理员',
    content: `确定${setAdmin ? '授予' : '取消'}“${member.name}”的家庭管理员权限吗？`,
    success: async (result) => {
      if (!result.confirm) return
      operationError.value = ''
      try {
        if (setAdmin) {
          await setFamilyAdmin(familyId.value, member.memberId, '家庭设置调整')
        } else {
          await unsetFamilyAdmin(familyId.value, member.memberId, '家庭设置调整')
        }
        await loadData()
      } catch (error) {
        operationError.value = apiErrorMessage(error, '调整管理员角色失败。')
      }
    }
  })
}

function confirmCreateTransfer() {
  const target = selectedTransferTarget.value
  if (!target) return
  uni.showModal({
    title: '确认转让创建者',
    content: `申请审核通过后，“${target.name}”将成为家庭创建者，你将变为普通成员。确定提交吗？`,
    success: async (result) => {
      if (!result.confirm) return
      try {
        currentTransfer.value = await createFounderTransfer(
          familyId.value,
          target.memberId,
          transferReason.value
        )
        transferReason.value = ''
      } catch (error) {
        operationError.value = apiErrorMessage(error, '发起转让申请失败。')
      }
    }
  })
}

function confirmCancelTransfer() {
  if (!currentTransfer.value) return
  uni.showModal({
    title: '取消转让申请',
    content: '确定取消当前待审核的创建者转让申请吗？',
    success: async (result) => {
      if (!result.confirm || !currentTransfer.value) return
      try {
        await cancelFounderTransfer(familyId.value, currentTransfer.value.requestId, '创建者主动取消')
        await loadData()
      } catch (error) {
        operationError.value = apiErrorMessage(error, '取消转让申请失败。')
      }
    }
  })
}

function confirmCreateDissolution() {
  if (!dissolutionReason.value) {
    operationError.value = '请先填写解散原因。'
    return
  }
  uni.showModal({
    title: '申请解散家庭',
    content: '这是高风险操作。后台审核通过后，家庭将停止使用。确定提交吗？',
    success: async (result) => {
      if (!result.confirm) return
      try {
        await createDissolution(familyId.value, dissolutionReason.value)
        dissolutionReason.value = ''
        await loadData()
      } catch (error) {
        operationError.value = apiErrorMessage(error, '提交解散申请失败。')
      }
    }
  })
}

function confirmCancelDissolution() {
  if (!currentDissolution.value) return
  uni.showModal({
    title: '取消解散申请',
    content: '确定取消当前待审核的家庭解散申请吗？',
    success: async (result) => {
      if (!result.confirm || !currentDissolution.value) return
      try {
        await cancelDissolution(familyId.value, currentDissolution.value.requestId, '创建者主动取消')
        await loadData()
      } catch (error) {
        operationError.value = apiErrorMessage(error, '取消解散申请失败。')
      }
    }
  })
}

function publicStatusText(status: string) {
  return ({ PRIVATE: '未公开', PENDING: '待审核', APPROVED: '已公开', REJECTED: '已驳回', TAKEN_DOWN: '已下架' } as Record<string, string>)[status] || '未知状态'
}

function applicationStatusText(status: string) {
  return ({ PENDING: '待审核', APPROVED: '已通过', REJECTED: '已驳回', CANCELLED: '已取消' } as Record<string, string>)[status] || '未知状态'
}

function roleText(role?: string | null) {
  return ({ FOUNDER: '创建者', FAMILY_ADMIN: '家庭管理员', MEMBER: '普通成员' } as Record<string, string>)[role || ''] || '未知角色'
}

function memberName(memberId: number | string) {
  return members.value.find((member) => String(member.memberId) === String(memberId))?.name || '目标成员'
}

function formatDate(value: string) {
  return value ? value.slice(0, 10) : ''
}

function openFamilyOverview() {
  uni.redirectTo({ url: `/pages/family/detail?familyId=${encodeURIComponent(familyId.value)}` })
}

onLoad((options) => {
  familyId.value = String(options?.familyId || '')
  if (!familyId.value) {
    loadError.value = '缺少家庭信息。'
  }
})
onShow(() => {
  if (familyId.value) loadData()
})
</script>

<style scoped>
.settings-page {
  background:
    radial-gradient(circle at 92% 0%, rgba(216, 175, 104, 0.14), transparent 260rpx),
    radial-gradient(circle at 0% 18%, rgba(24, 54, 83, 0.08), transparent 300rpx);
}

.settings-tabs {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12rpx;
  margin-bottom: 22rpx;
  border: 1rpx solid rgba(255, 255, 255, 0.74);
  border-radius: 28rpx;
  background:
    linear-gradient(135deg, rgba(255, 255, 255, 0.78) 0%, rgba(244, 250, 247, 0.76) 100%);
  padding: 10rpx;
  box-shadow: 0 12rpx 32rpx rgba(24, 54, 83, 0.06);
}

.settings-tabs button {
  min-height: 68rpx;
  margin: 0;
  border: 0;
  border-radius: 20rpx;
  background: transparent;
  color: var(--tree-text-secondary);
  padding: 10rpx 8rpx;
  font-size: 23rpx;
  line-height: 1.3;
}

.settings-tabs button::after {
  border: 0;
}

.settings-tabs button.active {
  background: linear-gradient(135deg, var(--tree-primary) 0%, var(--tree-green) 100%);
  color: #fff;
  font-weight: 800;
  box-shadow: 0 12rpx 26rpx rgba(24, 54, 83, 0.16);
}

.tree-input,
.tree-textarea,
.field-picker {
  margin-top: 18rpx;
}

.field-picker {
  min-height: 88rpx;
  padding: 0 24rpx;
  border: 1px solid var(--tree-border);
  border-radius: 20rpx;
  background: rgba(255, 255, 255, 0.94);
  box-shadow: 0 6rpx 18rpx rgba(24, 54, 83, 0.035);
  line-height: 88rpx;
}

.switch-row,
.member-row,
.record-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16rpx;
  margin-top: 14rpx;
  border: 1rpx solid rgba(226, 232, 240, 0.82);
  border-radius: 22rpx;
  background: rgba(248, 250, 252, 0.86);
  padding: 18rpx 20rpx;
}

.member-copy {
  display: flex;
  min-width: 0;
  flex: 1;
  flex-direction: column;
  gap: 6rpx;
}

.member-name {
  color: var(--tree-text);
  font-weight: 600;
}

.danger-zone {
  border-color: rgba(181, 71, 60, 0.24);
  background:
    radial-gradient(circle at 100% 0%, rgba(181, 71, 60, 0.08), transparent 160rpx),
    #fff8f6;
}
</style>
