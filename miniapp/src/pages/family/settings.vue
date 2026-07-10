<template>
  <view class="archive-page settings-page">
    <MiniBackHome />

    <view v-if="loading" class="settings-state archive-form-panel">
      <MiniEmptyState symbol="…" title="正在加载" description="正在读取家庭设置..." />
    </view>
    <view v-else-if="loadError" class="settings-state archive-form-panel">
      <MiniNotice tone="warm" title="加载失败">{{ loadError }}</MiniNotice>
      <MiniButton variant="secondary" class="settings-action" @click="loadData">重新加载</MiniButton>
    </view>

    <template v-else-if="family">
      <view class="settings-head archive-page-head">
        <view>
          <text class="archive-kicker">Family Permissions</text>
          <text class="archive-title">公开展示与权限</text>
          <text class="archive-subtitle">
            {{ family.familyName }} · 管理公开信息、家庭角色和高风险操作
          </text>
        </view>
        <view class="archive-seal">{{ family.familySurname.slice(0, 1) }}</view>
      </view>

      <view class="settings-context">
        <text class="context-back" @click="openFamilyOverview">返回详情</text>
      </view>

      <view v-if="!canManage" class="settings-state archive-form-panel">
        <MiniNotice tone="warm" title="无管理权限">只有家庭创建者或家庭管理员可以进入家庭设置。</MiniNotice>
        <MiniButton variant="secondary" class="settings-action" @click="openFamilyOverview">返回详情</MiniButton>
      </view>

      <template v-else>
        <MiniNotice v-if="family.status === 'DISSOLVED'" tone="warm" title="家庭已解散，等待恢复">
          该家庭已停止对外服务，设置暂不可编辑。请等待平台管理员恢复家庭后再继续操作。
        </MiniNotice>
        <view v-if="family.status === 'DISSOLVED'" class="archive-form-panel">
          <MiniEmptyState
            symbol="⏸"
            title="家庭已解散"
            description="当前仅可查看家庭概览，公开展示、角色与高风险操作需恢复后可用。"
          />
        </view>

        <MiniNotice v-else-if="family.status === 'DISSOLUTION_PENDING'" tone="warm" title="家庭解散申请审核中">
          审核完成或取消申请前，家庭资料、公开展示、角色和创建者转让暂不可调整。
        </MiniNotice>

        <template v-if="familyOperational">
          <view class="archive-segment-tabs">
            <view
              class="archive-segment-tab"
              :class="{ active: settingsSection === 'public' }"
              @click="settingsSection = 'public'"
            >
              <text>公开</text>
            </view>
            <view
              class="archive-segment-tab"
              :class="{ active: settingsSection === 'roles' }"
              @click="settingsSection = 'roles'"
            >
              <text>角色</text>
            </view>
            <view
              class="archive-segment-tab"
              :class="{ active: settingsSection === 'security' }"
              @click="settingsSection = 'security'"
            >
              <text>风险</text>
            </view>
            <view
              class="archive-segment-tab"
              :class="{ active: settingsSection === 'logs' }"
              @click="settingsSection = 'logs'"
            >
              <text>记录</text>
            </view>
          </view>

          <MiniNotice v-if="operationError" tone="warm" title="操作失败">
            {{ operationError }}
          </MiniNotice>

          <view v-if="settingsSection === 'security' && family.role !== 'FOUNDER'" class="archive-form-panel">
            <MiniNotice tone="security" title="仅创建者可操作">
              创建者转让和家庭解散属于高风险操作，家庭管理员只能查看其他设置。
            </MiniNotice>
          </view>

          <view v-if="settingsSection === 'public'" class="archive-form-panel">
            <view class="archive-section-head">
              <text class="archive-section-title">展示权限</text>
              <text class="archive-section-subtitle">当前状态：{{ publicStatusText(family.publicDisplayStatus) }}</text>
            </view>
            <template v-if="family.publicDisplayStatus === 'APPROVED'">
              <MiniNotice tone="security" title="已获得展示权限">
                平台已允许该家庭公开展示；是否对外展示由家庭自行开启或关闭。
              </MiniNotice>
              <label class="switch-row">
                <text>{{ family.publicDisplayEnabled ? '正在公开展示' : '暂未公开展示' }}</text>
                <switch :checked="family.publicDisplayEnabled" color="#163353" @change="onPublicDisplayToggle" />
              </label>
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
              class="settings-action"
              :loading="publicSubmitting"
              :disabled="publicSubmitting"
              @click="submitPublic"
            >
              提交公开申请
            </MiniButton>
            <MiniNotice v-if="pendingPublicApplication" tone="warm" title="申请待审核">
              {{ pendingPublicApplication.reason || '平台审核通过后，将获得展示权限；是否展示仍由家庭自行开启。' }}
            </MiniNotice>
            <MiniButton
              v-if="pendingPublicApplication"
              class="settings-action"
              variant="secondary"
              :loading="publicSubmitting"
              :disabled="publicSubmitting"
              @click="confirmCancelPublic"
            >
              取消待审核申请
            </MiniButton>
            <view v-if="publicApplications.length > 0" class="settings-history archive-list">
              <view v-for="item in publicApplications.slice(0, 5)" :key="item.applicationId" class="archive-row">
                <view class="archive-row-main">
                  <text class="archive-row-title">{{ applicationStatusText(item.status) }}</text>
                  <text v-if="item.reason" class="archive-row-desc">{{ item.reason }}</text>
                </view>
                <text class="archive-row-meta">{{ formatDate(item.createdAt) }}</text>
              </view>
            </view>
          </view>

          <view v-if="settingsSection === 'public' && family.publicDisplayStatus === 'APPROVED'" class="archive-form-panel">
            <view class="archive-section-head">
              <text class="archive-section-title">公开信息</text>
              <text class="archive-section-subtitle">开启展示时，访客会看到这些公开资料</text>
            </view>
            <input v-model.trim="publicForm.publicContactName" class="tree-input" maxlength="80" placeholder="公开联系人" />
            <input v-model.trim="publicForm.publicContactPhone" class="tree-input" maxlength="30" placeholder="公开联系电话" />
            <input v-model.trim="publicForm.publicContactWechat" class="tree-input" maxlength="80" placeholder="公开微信" />
            <textarea v-model.trim="publicForm.publicContactNote" class="tree-textarea" maxlength="300" placeholder="公开联系说明" />
            <label class="switch-row">
              <text>公开联系方式</text>
              <switch :checked="publicForm.publicContactVisible" color="#163353" @change="onPublicContactVisibleChange" />
            </label>
            <label class="switch-row">
              <text>展示家庭列表收录</text>
              <switch :checked="publicForm.searchable" color="#163353" @change="onSearchableChange" />
            </label>
            <MiniButton
              class="settings-action"
              variant="secondary"
              :loading="savingPublicInfo"
              :disabled="savingPublicInfo"
              @click="savePublicInfo"
            >
              保存公开信息
            </MiniButton>
          </view>

          <view v-if="settingsSection === 'roles'" class="archive-form-panel">
            <view class="archive-section-head">
              <text class="archive-section-title">家庭管理员</text>
              <text class="archive-section-subtitle">创建者可授予或取消家庭管理员角色</text>
            </view>
            <MiniNotice v-if="family.role !== 'FOUNDER'" tone="security" title="只读">
              只有家庭创建者可以调整管理员角色。
            </MiniNotice>
            <MiniEmptyState
              v-if="boundMembers.length === 0"
              symbol="员"
              title="暂无已绑定成员"
              description="只有已绑定账号的活跃成员可设为家庭管理员。"
            />
            <view v-for="member in boundMembers" :key="member.memberId" class="archive-relation-row">
              <view class="archive-relation-copy">
                <text class="archive-relation-title">{{ member.name }}</text>
                <text class="tree-muted">{{ roleText(member.boundFamilyRole) }}</text>
              </view>
              <view class="settings-row-actions">
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
            </view>
          </view>

          <view v-if="family.role === 'FOUNDER' && settingsSection === 'security'" class="archive-form-panel">
            <view class="archive-section-head">
              <text class="archive-section-title">转让创建者</text>
              <text class="archive-section-subtitle">目标必须是已绑定账号的活跃家庭成员，需后台审核</text>
            </view>
            <template v-if="currentTransfer">
              <MiniNotice tone="warm" title="转让申请待审核">
                目标成员：{{ memberName(currentTransfer.toMemberId) }}。后台审核通过前可取消。
              </MiniNotice>
              <MiniButton class="settings-action" variant="secondary" @click="confirmCancelTransfer">取消转让申请</MiniButton>
            </template>
            <template v-else>
              <picker :range="transferTargetNames" :value="transferTargetIndex" @change="transferTargetIndex = Number($event.detail.value)">
                <view class="field-picker">{{ selectedTransferTarget?.name || '选择接任成员' }}</view>
              </picker>
              <textarea v-model.trim="transferReason" class="tree-textarea" maxlength="500" placeholder="转让原因（可选）" />
              <MiniButton class="settings-action" :disabled="!selectedTransferTarget" @click="confirmCreateTransfer">
                发起转让申请
              </MiniButton>
            </template>
          </view>

          <view v-if="family.role === 'FOUNDER' && settingsSection === 'security'" class="archive-danger-panel">
            <text class="archive-danger-title">解散家庭</text>
            <text class="archive-danger-desc">高风险操作，审核通过后家庭将停止使用</text>
            <template v-if="currentDissolution">
              <MiniNotice tone="warm" title="解散申请待审核">
                {{ currentDissolution.requestReason || '后台正在审核该申请。' }}
              </MiniNotice>
              <MiniButton class="settings-action" variant="secondary" @click="confirmCancelDissolution">取消解散申请</MiniButton>
            </template>
            <template v-else>
              <textarea v-model.trim="dissolutionReason" class="tree-textarea" maxlength="500" placeholder="请填写解散原因" />
              <MiniButton class="settings-action" variant="secondary" @click="confirmCreateDissolution">申请解散家庭</MiniButton>
            </template>
          </view>

          <view v-if="settingsSection === 'logs'" class="archive-form-panel">
            <view class="archive-section-head">
              <text class="archive-section-title">操作记录</text>
              <text class="archive-section-subtitle">展示最近的家庭成员、关系、邀请与公开展示操作</text>
            </view>
            <MiniEmptyState
              v-if="operationLogs.length === 0"
              symbol="录"
              title="暂无记录"
              description="家庭重要操作会记录在这里。"
            />
            <view v-else class="archive-list">
              <view v-for="log in operationLogs" :key="log.id" class="archive-row">
                <view class="archive-row-main">
                  <text class="archive-row-title">{{ operationActionText(log.action) }}</text>
                  <text class="archive-row-desc">{{ operationModuleText(log.module) }} · {{ operationActorText(log.operatorType) }}</text>
                </view>
                <text class="archive-row-meta">{{ formatDate(log.createdAt) }}</text>
              </view>
            </view>
          </view>
        </template>
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
  enablePublicFamily,
  listFamilyOperationLogs,
  listPublicApplications,
  submitPublicApplication
} from '@/api/publicApplications'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
import { useSessionStore } from '@/stores/session'
import type {
  DissolutionRequest,
  FamilyDetail,
  FamilyMember,
  FamilyOperationLog,
  FounderTransferRequest,
  PublicApplication
} from '@/types/api'
import { optionalText, validateTextFields } from '@/utils/inputValidation'

const session = useSessionStore()
const familyId = ref('')
const family = ref<FamilyDetail | null>(null)
const members = ref<FamilyMember[]>([])
const publicApplications = ref<PublicApplication[]>([])
const operationLogs = ref<FamilyOperationLog[]>([])
const currentTransfer = ref<FounderTransferRequest | null>(null)
const currentDissolution = ref<DissolutionRequest | null>(null)
const loading = ref(false)
const loadError = ref('')
const operationError = ref('')
const savingPublicInfo = ref(false)
const publicSubmitting = ref(false)
const publicReason = ref('')
const takeDownReason = ref('')
const togglingPublicDisplay = ref(false)
const transferTargetIndex = ref(0)
const transferReason = ref('')
const dissolutionReason = ref('')
const settingsSection = ref<'public' | 'roles' | 'security' | 'logs'>('public')
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
  if (!familyId.value) {
    loadError.value = '缺少家庭信息。'
    loading.value = false
    return
  }
  const route = `/pages/family/settings?familyId=${encodeURIComponent(familyId.value)}`
  if (!session.requireLogin(route)) return
  loading.value = true
  loadError.value = ''
  try {
    const familyResult = await getFamilyDetail(familyId.value)
    family.value = familyResult
    if (familyResult.role !== 'FOUNDER' && familyResult.role !== 'FAMILY_ADMIN') return
    const [memberResult, publicResult, logResult, transferResult, dissolutionResult] = await Promise.all([
      listFamilyMembers(familyId.value),
      isRealApiMode
        ? listPublicApplications(familyId.value)
        : Promise.resolve({ items: [], page: 1, pageSize: 50, total: 0 }),
      isRealApiMode
        ? listFamilyOperationLogs(familyId.value)
        : Promise.resolve({ items: [], page: 1, pageSize: 20, total: 0 }),
      isRealApiMode
        ? nullableRequest(() => getCurrentFounderTransfer(familyId.value))
        : Promise.resolve(null),
      isRealApiMode
        ? nullableRequest(() => getCurrentDissolution(familyId.value))
        : Promise.resolve(null)
    ])
    members.value = memberResult
    publicApplications.value = publicResult.items
    operationLogs.value = logResult.items
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
  const validationMessage = validateTextFields([
    { value: publicForm.publicContactName, label: '公开联系人', kind: 'name', maxLength: 80 },
    { value: publicForm.publicContactPhone, label: '公开联系电话', maxLength: 30 },
    { value: publicForm.publicContactWechat, label: '公开微信', maxLength: 80 },
    { value: publicForm.publicContactNote, label: '公开联系说明', kind: 'multiLine', maxLength: 300 }
  ])
  if (validationMessage) {
    operationError.value = validationMessage
    return
  }
  savingPublicInfo.value = true
  operationError.value = ''
  try {
    family.value = await updateFamily(familyId.value, {
      ...publicForm,
      publicContactName: optionalText(publicForm.publicContactName),
      publicContactPhone: optionalText(publicForm.publicContactPhone),
      publicContactWechat: optionalText(publicForm.publicContactWechat),
      publicContactNote: optionalText(publicForm.publicContactNote)
    })
    uni.showToast({ title: '公开信息已保存', icon: 'success' })
  } catch (error) {
    operationError.value = apiErrorMessage(error, '保存公开信息失败。')
  } finally {
    savingPublicInfo.value = false
  }
}

async function submitPublic() {
  const validationMessage = validateTextFields([
    { value: publicReason.value, label: '公开申请说明', kind: 'multiLine', maxLength: 500 }
  ])
  if (validationMessage) {
    operationError.value = validationMessage
    return
  }
  publicSubmitting.value = true
  operationError.value = ''
  try {
    await submitPublicApplication(familyId.value, optionalText(publicReason.value) || '')
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

function onPublicDisplayToggle(event: Event) {
  const nextValue = switchValue(event)
  if (nextValue) {
    enablePublicDisplay()
    return
  }
  confirmClosePublic()
}

async function enablePublicDisplay() {
  if (togglingPublicDisplay.value) return
  togglingPublicDisplay.value = true
  operationError.value = ''
  try {
    const result = await enablePublicFamily(familyId.value)
    if (family.value) family.value.publicDisplayEnabled = result.publicDisplayEnabled
    uni.showToast({ title: '公开展示已开启', icon: 'success' })
    await loadData()
  } catch (error) {
    operationError.value = apiErrorMessage(error, '开启公开展示失败。')
  } finally {
    togglingPublicDisplay.value = false
  }
}

function confirmClosePublic() {
  const validationMessage = validateTextFields([
    { value: takeDownReason.value, label: '关闭说明', kind: 'multiLine', maxLength: 500 }
  ])
  if (validationMessage) {
    operationError.value = validationMessage
    return
  }
  uni.showModal({
    title: '关闭公开展示',
    content: '关闭后，公开主页和公开家庭树将立即不可访问；展示权限仍保留，可稍后重新开启。确定继续吗？',
    success: async (result) => {
      if (!result.confirm) return
      togglingPublicDisplay.value = true
      operationError.value = ''
      try {
        await closePublicFamily(familyId.value, optionalText(takeDownReason.value) || '')
        takeDownReason.value = ''
        uni.showToast({ title: '公开展示已关闭', icon: 'success' })
        await loadData()
      } catch (error) {
        operationError.value = apiErrorMessage(error, '关闭公开展示失败。')
      } finally {
        togglingPublicDisplay.value = false
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
  const validationMessage = validateTextFields([
    { value: transferReason.value, label: '转让原因', kind: 'multiLine', maxLength: 500 }
  ])
  if (validationMessage) {
    operationError.value = validationMessage
    return
  }
  uni.showModal({
    title: '确认转让创建者',
    content: `申请审核通过后，“${target.name}”将成为家庭创建者，你将变为普通成员。确定提交吗？`,
    success: async (result) => {
      if (!result.confirm) return
      try {
        currentTransfer.value = await createFounderTransfer(
          familyId.value,
          target.memberId,
          optionalText(transferReason.value) || ''
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
  const validationMessage = validateTextFields([
    { value: dissolutionReason.value, label: '解散原因', kind: 'multiLine', required: true, maxLength: 500 }
  ])
  if (validationMessage) {
    operationError.value = validationMessage
    return
  }
  uni.showModal({
    title: '申请解散家庭',
    content: '这是高风险操作。后台审核通过后，家庭将停止使用。确定提交吗？',
    success: async (result) => {
      if (!result.confirm) return
      try {
        await createDissolution(familyId.value, optionalText(dissolutionReason.value) || '')
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
  return ({ PRIVATE: '未申请', PENDING: '待审核', APPROVED: '已获展示权限', REJECTED: '已驳回', TAKEN_DOWN: '已下架' } as Record<string, string>)[status] || '未知状态'
}

function applicationStatusText(status: string) {
  return ({ PENDING: '待审核', APPROVED: '已通过', REJECTED: '已驳回', CANCELLED: '已取消' } as Record<string, string>)[status] || '未知状态'
}

function roleText(role?: string | null) {
  return ({ FOUNDER: '创建者', FAMILY_ADMIN: '家庭管理员', MEMBER: '普通成员' } as Record<string, string>)[role || ''] || '未知角色'
}

function operationModuleText(module: string) {
  return ({
    FAMILY: '家庭资料',
    FAMILY_MEMBER: '成员资料',
    FAMILY_RELATIONSHIP: '家庭关系',
    FAMILY_INVITATION: '家庭邀请',
    FAMILY_JOIN_REQUEST: '加入申请',
    FAMILY_PUBLIC_APPLICATION: '公开展示',
    FAMILY_ROLE: '角色权限',
    FAMILY_DISSOLUTION: '家庭解散',
    FAMILY_FOUNDER_TRANSFER: '创建者转让'
  } as Record<string, string>)[module] || module
}

function operationActionText(action: string) {
  return ({
    CREATE_FAMILY: '创建家庭',
    UPDATE_FAMILY: '更新家庭资料',
    CREATE_MEMBER: '创建成员',
    UPDATE_MEMBER: '更新成员',
    DELETE_MEMBER: '删除成员',
    CREATE_RELATIONSHIP: '创建关系',
    UPDATE_RELATIONSHIP: '调整关系',
    DELETE_RELATIONSHIP: '删除关系',
    PLACE_EXISTING_MEMBER: '接入暂存成员',
    SUBMIT_PUBLIC_APPLICATION: '提交展示权限申请',
    CANCEL_PUBLIC_APPLICATION: '取消展示权限申请',
    APPROVE_PUBLIC_APPLICATION: '展示权限审核通过',
    REJECT_PUBLIC_APPLICATION: '展示权限审核驳回',
    ENABLE_PUBLIC_DISPLAY: '开启公开展示',
    DISABLE_PUBLIC_DISPLAY: '关闭公开展示',
    CLOSE_PUBLIC_FAMILY: '关闭公开展示',
    SET_FAMILY_ADMIN: '设置管理员',
    UNSET_FAMILY_ADMIN: '取消管理员'
  } as Record<string, string>)[action] || action
}

function operationActorText(operatorType: string) {
  return operatorType === 'ADMIN' ? '平台操作' : '家庭操作'
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
  padding-top: 28rpx;
}

.settings-page :deep(.mini-back-home) {
  margin-bottom: 22rpx;
}

.settings-state :deep(.mini-notice) {
  margin-bottom: 18rpx;
}

.settings-page :deep(.mini-notice) {
  margin-top: 16rpx;
  margin-bottom: 16rpx;
}

.settings-context {
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

.settings-page .archive-form-panel + .archive-form-panel {
  margin-top: -8rpx;
}

.settings-action {
  margin-top: 20rpx;
  align-self: flex-start;
}

.settings-history {
  margin-top: 20rpx;
}

.settings-row-actions {
  display: flex;
  flex-direction: row;
  flex-wrap: wrap;
  gap: 10rpx;
  justify-content: flex-end;
}

.settings-page .archive-relation-row {
  align-items: flex-start;
}
</style>
