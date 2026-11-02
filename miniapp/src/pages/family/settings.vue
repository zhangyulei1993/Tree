<template>
  <view class="archive-page settings-page">
    <MiniBackHome :return-url="returnUrl" :return-label="returnLabel" />

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
          <text class="archive-kicker">Family Settings</text>
          <text class="archive-title">家庭设置</text>
          <text class="archive-subtitle">
            {{ family.familyName }} · 展示、权限、转让与解散
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
        <MiniNotice v-if="family.status === 'DISSOLVED'" tone="warm" title="家庭已完成解散">
          家庭已从用户端隐藏。如需恢复，请联系平台管理员。
        </MiniNotice>
        <view v-if="family.status === 'DISSOLVED'" class="archive-form-panel">
          <MiniEmptyState
            symbol="⏸"
            title="家庭已完成解散"
            description="当前仅可由平台管理员恢复。"
          />
        </view>

        <template v-else-if="family.status === 'DISSOLUTION_COOLDOWN'">
          <MiniNotice tone="warm" title="家庭处于恢复冷静期">
            当前仅可查看。创建者可恢复，或跳过冷静期解散。
          </MiniNotice>
          <view class="archive-form-panel">
            <view class="archive-section-head">
              <text class="archive-section-title">恢复冷静期</text>
              <text class="archive-section-subtitle">
                截止 {{ formatDateTime(family.dissolutionCooldownUntil) }} · {{ family.dissolutionCooldownDays || 0 }} 天
              </text>
            </view>
            <MiniButton
              v-if="family.role === 'FOUNDER'"
              class="settings-action"
              @click="confirmRestoreFromCooldown"
            >
              恢复家庭
            </MiniButton>
            <MiniButton
              v-if="family.role === 'FOUNDER'"
              class="settings-action"
              variant="secondary"
              @click="confirmFinalizeDissolution"
            >
              跳过冷静期并完成解散
            </MiniButton>
          </view>
        </template>

        <template v-else-if="family.status === 'DISSOLUTION_PENDING'">
          <MiniNotice tone="warm" title="家庭解散申请审核中">
            审核完成或取消申请前，家庭资料、角色和创建者转让暂不可调整。
          </MiniNotice>
          <view class="archive-form-panel">
            <view class="archive-section-head">
              <text class="archive-section-title">待审核申请</text>
              <text class="archive-section-subtitle">创建者可在审核前撤销</text>
            </view>
            <text class="settings-muted">
              {{ currentDissolution?.requestReason || '后台正在审核该申请。' }}
            </text>
            <MiniButton
              v-if="family.role === 'FOUNDER'"
              class="settings-action"
              variant="secondary"
              :disabled="!currentDissolution"
              @click="confirmCancelDissolution"
            >
              撤销解散申请
            </MiniButton>
          </view>
        </template>

        <template v-if="familyOperational">
          <view v-if="qualityEligibility && !customerPublicDisplayAvailable" class="archive-form-panel quality-eligibility-panel">
            <view class="archive-section-head">
              <text class="archive-section-title">优质家庭申请资格</text>
              <text class="archive-section-subtitle">达到条件后，平台会发出申请资格邀请</text>
            </view>
            <view class="quality-eligibility-list">
              <view v-for="requirement in qualityEligibility.requirements" :key="requirement" class="quality-eligibility-item">
                <text class="quality-eligibility-dot">·</text>
                <text>{{ requirement }}</text>
              </view>
            </view>
          </view>
          <view class="archive-segment-tabs">
            <view
              v-if="publicSettingsVisible"
              class="archive-segment-tab"
              :class="{ active: settingsSection === 'public' }"
              @click="settingsSection = 'public'"
            >
              <text>展示</text>
            </view>
            <view
              class="archive-segment-tab"
              :class="{ active: settingsSection === 'roles' }"
              @click="settingsSection = 'roles'"
            >
              <text>权限</text>
            </view>
            <view
              class="archive-segment-tab"
              :class="{ active: settingsSection === 'transfer' }"
              @click="settingsSection = 'transfer'"
            >
              <text>转让</text>
            </view>
            <view
              v-if="family.role === 'FOUNDER'"
              class="archive-segment-tab"
              :class="{ active: settingsSection === 'dissolution' }"
              @click="settingsSection = 'dissolution'"
            >
              <text>解散</text>
            </view>
            <view
              class="archive-segment-tab"
              :class="{ active: settingsSection === 'logs' }"
              @click="settingsSection = 'logs'"
            >
              <text>记录</text>
            </view>
          </view>
          <view class="settings-section-guide">
            <text class="settings-section-guide-title">{{ settingsSectionGuide.title }}</text>
            <text class="settings-section-guide-desc">{{ settingsSectionGuide.description }}</text>
          </view>

          <MiniNotice v-if="operationError" tone="warm" title="操作失败">
            {{ operationError }}
          </MiniNotice>

          <view v-if="settingsSection === 'transfer' && family.role !== 'FOUNDER'" class="archive-form-panel">
            <MiniNotice tone="security" title="仅创建者可操作">
              创建者转让属于高风险操作，家庭管理员只能查看角色与记录。
            </MiniNotice>
          </view>

          <view v-if="publicSettingsVisible && settingsSection === 'public'" class="archive-form-panel">
            <view class="archive-section-head">
              <text class="archive-section-title">展示权限</text>
              <text class="archive-section-subtitle">当前状态：{{ publicStatusText(family.publicDisplayStatus) }}</text>
            </view>
            <template v-if="family.publicDisplayStatus === 'APPROVED'">
              <MiniNotice tone="security" title="已获得展示权限">
                平台已将该家庭收录为优质家庭案例，审核通过后会自动展示；你仍可随时关闭展示。
              </MiniNotice>
              <view class="switch-row">
                <text>{{ family.publicDisplayEnabled ? '正在作为优质家庭展示' : '暂未展示' }}</text>
                <MiniButton
                  size="sm"
                  :variant="family.publicDisplayEnabled ? 'secondary' : 'primary'"
                  :loading="togglingPublicDisplay"
                  :disabled="!customerPublicDisplayAvailable || togglingPublicDisplay"
                  @click="family.publicDisplayEnabled ? confirmClosePublic() : enablePublicDisplay()"
                >
                  {{ family.publicDisplayEnabled ? '关闭展示' : '开启展示' }}
                </MiniButton>
              </view>
            </template>
            <view v-else-if="customerPublicDisplayAvailable && !pendingPublicApplication" class="quality-check-panel">
              <view class="quality-check-panel-head">
                <text class="archive-section-title">当前资格核对</text>
                <text class="archive-section-subtitle">提交申请和后台审核时都会按最新标准再次核对</text>
              </view>
              <view v-for="check in qualityEligibility?.checks || []" :key="check.key" class="quality-check-row">
                <view class="quality-check-copy">
                  <text class="quality-check-label">{{ check.label }}</text>
                  <text class="quality-check-value">当前 {{ check.current }} · {{ check.required }}</text>
                </view>
                <text class="quality-check-status" :class="check.met ? 'quality-check-met' : 'quality-check-unmet'">
                  {{ check.met ? '已满足' : '未满足' }}
                </text>
              </view>
            </view>
            <MiniButton
              v-if="customerPublicDisplayAvailable && !pendingPublicApplication && family.publicDisplayStatus !== 'APPROVED'"
              class="settings-action"
              :loading="publicSubmitting"
              :disabled="publicSubmitting"
              @click="submitPublic"
            >
              申请成为优质家庭
            </MiniButton>
            <MiniNotice v-if="pendingPublicApplication" tone="warm" title="申请待审核">
              {{ pendingPublicApplication.reason || '平台审核通过后，将作为优质家庭案例供其他用户学习；是否展示仍由家庭自行开启。' }}
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
            <view v-if="publicApplications.length > 0" class="settings-history archive-form-panel">
              <view class="archive-section-head">
                <text class="archive-section-title">申请历史</text>
                <text class="archive-section-subtitle">记录每次申请及平台审核结果</text>
              </view>
              <view v-for="item in publicApplications" :key="item.applicationId" class="application-history-item">
                <view class="application-history-head">
                  <text class="archive-row-title">优质家庭申请</text>
                  <text class="archive-status-tag" :class="applicationStatusClass(item.status)">{{ applicationStatusText(item.status) }}</text>
                </view>
                <text class="application-history-meta">申请于 {{ formatDate(item.createdAt) }}</text>
                <text v-if="item.reason" class="application-history-text">申请说明：{{ item.reason }}</text>
                <text v-if="item.reviewComment" class="application-history-text">审核意见：{{ item.reviewComment }}</text>
                <text v-if="item.reviewedAt" class="application-history-meta">审核于 {{ formatDate(item.reviewedAt) }}</text>
              </view>
            </view>
          </view>

          <view v-if="publicSettingsVisible && settingsSection === 'public' && customerPublicDisplayAvailable && family.publicDisplayStatus === 'APPROVED'" class="archive-form-panel">
            <view class="archive-section-head">
              <text class="archive-section-title">公开信息</text>
              <text class="archive-section-subtitle">优质家庭仅展示可学习的家谱内容，不提供陌生联系入口</text>
            </view>
            <label class="switch-row">
              <text>加入优质家庭学习列表</text>
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
            <view class="archive-form-panel settings-subscription-panel">
              <view class="archive-section-head">
                <text class="archive-section-title">成员变更提醒</text>
                <text class="archive-section-subtitle">成员退出家庭后提醒创建者和管理员；微信提醒需在弹窗中允许</text>
              </view>
              <MiniButton
                class="settings-action"
                variant="secondary"
                @click="enableMemberChangeNotice"
              >
                开启成员变更提醒
              </MiniButton>
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

          <view v-if="family.role === 'FOUNDER' && settingsSection === 'transfer'" class="archive-form-panel">
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

          <view v-if="settingsSection === 'dissolution' && family.role !== 'FOUNDER'" class="archive-form-panel">
            <MiniNotice tone="security" title="仅创建者可操作">
              家庭解散属于高风险操作，仅创建者可以发起或取消。
            </MiniNotice>
          </view>

          <view v-if="family.role === 'FOUNDER' && settingsSection === 'dissolution'" class="archive-danger-panel">
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
              <text class="archive-section-subtitle">展示最近的家庭成员、关系与邀请操作</text>
            </view>
            <MiniEmptyState
              v-if="visibleOperationLogs.length === 0"
              symbol="录"
              title="暂无记录"
              description="家庭重要操作会记录在这里。"
            />
            <view v-else class="archive-list">
              <view v-for="log in visibleOperationLogs" :key="log.id" class="archive-row">
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
  finalizeDissolution,
  getCurrentDissolution,
  restoreDissolution
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
  getQualityEligibility,
  listFamilyOperationLogs,
  listPublicApplications,
  submitPublicApplication
} from '@/api/publicApplications'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
import { requestSubscribeMessage } from '@/features/notifications/subscribe'
import { useSessionStore } from '@/stores/session'
import type {
  DissolutionRequest,
  FamilyDetail,
  FamilyMember,
  FamilyOperationLog,
  FounderTransferRequest,
  PublicApplication,
  QualityEligibility
} from '@/types/api'
import { optionalText, validateTextFields } from '@/utils/inputValidation'

const session = useSessionStore()
const familyId = ref('')
const returnTo = ref('')
const family = ref<FamilyDetail | null>(null)
const members = ref<FamilyMember[]>([])
const qualityEligibility = ref<QualityEligibility | null>(null)
const publicApplications = ref<PublicApplication[]>([])
const operationLogs = ref<FamilyOperationLog[]>([])
const currentTransfer = ref<FounderTransferRequest | null>(null)
const currentDissolution = ref<DissolutionRequest | null>(null)
const returnUrl = computed(() => returnTo.value === 'tree'
  ? `/pages/family/private-tree?familyId=${encodeURIComponent(familyId.value)}`
  : `/pages/family/detail?familyId=${encodeURIComponent(familyId.value)}`)
const returnLabel = computed(() => returnTo.value === 'tree' ? '返回家庭树' : '返回家庭详情')
const loading = ref(false)
const loadError = ref('')
const operationError = ref('')
const savingPublicInfo = ref(false)
const publicSubmitting = ref(false)
const takeDownReason = ref('')
const togglingPublicDisplay = ref(false)
const transferTargetIndex = ref(0)
const transferReason = ref('')
const dissolutionReason = ref('')
const settingsSection = ref<'public' | 'roles' | 'transfer' | 'dissolution' | 'logs'>('roles')
const settingsSectionGuide = computed(() => ({
  public: { title: '展示与申优', description: '管理优质家庭申请及对外展示范围。' },
  roles: { title: '家庭权限与提醒', description: '设置管理员，并管理成员变更提醒。' },
  transfer: { title: '创建者转让', description: '高风险操作，提交后需平台审核。' },
  dissolution: { title: '解散家庭', description: '高风险操作，提交后需平台审核。' },
  logs: { title: '操作记录', description: '查看家庭成员、关系与邀请的重要操作。' }
})[settingsSection.value])
const customerPublicDisplayAvailable = computed(() => Boolean(qualityEligibility.value?.eligible))
const publicForm = reactive({
  searchable: false,
  publicContactName: '',
  publicContactPhone: '',
  publicContactWechat: '',
  publicContactNote: '',
  publicContactVisible: false
})
const publicOperationActions = [
  'SUBMIT_PUBLIC_APPLICATION',
  'CANCEL_PUBLIC_APPLICATION',
  'APPROVE_PUBLIC_APPLICATION',
  'REJECT_PUBLIC_APPLICATION',
  'ENABLE_PUBLIC_DISPLAY',
  'DISABLE_PUBLIC_DISPLAY',
  'CLOSE_PUBLIC_FAMILY'
]

const canManage = computed(() =>
  family.value?.role === 'FOUNDER' || family.value?.role === 'FAMILY_ADMIN'
)
const familyOperational = computed(() => family.value?.status === 'NORMAL')
const publicSettingsVisible = computed(() => customerPublicDisplayAvailable.value || publicApplications.value.length > 0 || family.value?.publicDisplayStatus === 'APPROVED')
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
const visibleOperationLogs = computed(() =>
  publicSettingsVisible.value
    ? operationLogs.value
    : operationLogs.value.filter((log) =>
      log.module !== 'FAMILY_PUBLIC_APPLICATION' &&
      !publicOperationActions.includes(log.action)
    )
)

function switchValue(event: Event) {
  const detail = (event as Event & { detail?: { value?: boolean | string | number } }).detail
  const value = detail?.value
  return value === true || value === 'true' || value === 1 || value === '1'
}

function onSearchableChange(event: Event) {
  publicForm.searchable = switchValue(event)
}

async function enableMemberChangeNotice() {
  const result = await requestSubscribeMessage('memberChange')
  if (result.status === 'accepted') {
    uni.showToast({ title: '已开启成员变更提醒', icon: 'success' })
    return
  }
  if (result.status === 'unsupported') {
    uni.showToast({ title: '请在微信小程序中开启提醒', icon: 'none' })
    return
  }
  uni.showToast({ title: '未开启，请在微信弹窗中允许提醒', icon: 'none' })
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
    const [memberResult, publicResult, logResult, transferResult, dissolutionResult, eligibilityResult] = await Promise.all([
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
        : Promise.resolve(null),
      isRealApiMode
        ? nullableRequest(() => getQualityEligibility(familyId.value))
        : Promise.resolve(null)
    ])
    members.value = memberResult
    qualityEligibility.value = eligibilityResult
    publicApplications.value = publicResult.items
    operationLogs.value = logResult.items
    currentTransfer.value = transferResult
    currentDissolution.value = dissolutionResult
    fillPublicForm(familyResult)
    // Eligibility is loaded together with the page data. Only normalize the
    // requested section after it has been applied, otherwise the申优入口 can
    // be reset to“角色”before the eligibility request finishes.
    if (!publicSettingsVisible.value && settingsSection.value === 'public') {
      settingsSection.value = 'roles'
    }
  } catch (error) {
    loadError.value = apiErrorMessage(error, '家庭设置加载失败。')
  } finally {
    loading.value = false
  }
}

function fillPublicForm(value: FamilyDetail) {
  publicForm.searchable = value.searchable
  publicForm.publicContactName = ''
  publicForm.publicContactPhone = ''
  publicForm.publicContactWechat = ''
  publicForm.publicContactNote = ''
  publicForm.publicContactVisible = false
}

async function savePublicInfo() {
  if (savingPublicInfo.value) return
  savingPublicInfo.value = true
  operationError.value = ''
  try {
    family.value = await updateFamily(familyId.value, {
      searchable: publicForm.searchable,
      publicContactName: null,
      publicContactPhone: null,
      publicContactWechat: null,
      publicContactNote: null,
      publicContactVisible: false,
      expectedRevision: family.value?.revision
    })
    uni.showToast({ title: '公开信息已保存', icon: 'success' })
  } catch (error) {
    operationError.value = apiErrorMessage(error, '保存公开信息失败。')
  } finally {
    savingPublicInfo.value = false
  }
}

async function submitPublic() {
  if (!customerPublicDisplayAvailable) {
    uni.showToast({ title: '当前不可操作', icon: 'none' })
    return
  }
  publicSubmitting.value = true
  operationError.value = ''
  try {
    await requestSubscribeMessage('audit')
    await submitPublicApplication(familyId.value)
    uni.showToast({ title: '优质家庭申请已提交', icon: 'success' })
    await loadData()
  } catch (error) {
    operationError.value = apiErrorMessage(error, '提交优质家庭申请失败。')
  } finally {
    publicSubmitting.value = false
  }
}

function confirmCancelPublic() {
  const application = pendingPublicApplication.value
  if (!application) return
  uni.showModal({
    title: '取消优质家庭申请',
    content: '确定取消当前待审核申请吗？',
    success: async (result) => {
      if (!result.confirm) return
      try {
        await cancelPublicApplication(familyId.value, application.applicationId, '家庭管理主动取消')
        await loadData()
      } catch (error) {
    operationError.value = apiErrorMessage(error, '取消优质家庭申请失败。')
      }
    }
  })
}

async function enablePublicDisplay() {
  if (togglingPublicDisplay.value) return
  togglingPublicDisplay.value = true
  operationError.value = ''
  try {
    const result = await enablePublicFamily(familyId.value)
    if (family.value) family.value.publicDisplayEnabled = result.publicDisplayEnabled
        uni.showToast({ title: '优质家庭展示已开启', icon: 'success' })
    await loadData()
  } catch (error) {
    operationError.value = apiErrorMessage(error, '开启优质家庭展示失败。')
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
    title: '关闭优质家庭展示',
    content: '关闭后，优质家庭列表和学习页面将立即不可访问；入选资格仍保留，可稍后重新开启。确定继续吗？',
    success: async (result) => {
      if (!result.confirm) return
      togglingPublicDisplay.value = true
      operationError.value = ''
      try {
        await closePublicFamily(familyId.value, optionalText(takeDownReason.value) || '')
        takeDownReason.value = ''
        uni.showToast({ title: '优质家庭展示已关闭', icon: 'success' })
        await loadData()
      } catch (error) {
        operationError.value = apiErrorMessage(error, '关闭优质家庭展示失败。')
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

async function confirmCreateTransfer() {
  const target = selectedTransferTarget.value
  if (!target) return
  const validationMessage = validateTextFields([
    { value: transferReason.value, label: '转让原因', kind: 'multiLine', maxLength: 500 }
  ])
  if (validationMessage) {
    operationError.value = validationMessage
    return
  }
  await requestSubscribeMessage('audit')
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

async function confirmCreateDissolution() {
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
  await requestSubscribeMessage('audit')
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

function confirmFinalizeDissolution() {
  if (!family.value || family.value.role !== 'FOUNDER') return
  uni.showModal({
    title: '跳过冷静期',
    content: '跳过后家庭将直接完成解散，并继续保持对用户端隐藏。此后如需恢复，请联系后台处理。确认继续吗？',
    confirmColor: '#A83B2D',
    success: async (result) => {
      if (!result.confirm) return
      try {
        family.value = await finalizeDissolution(familyId.value)
        currentDissolution.value = null
        uni.showToast({ title: '已完成解散', icon: 'success' })
      } catch (error) {
        operationError.value = apiErrorMessage(error, '完成解散失败，请稍后重试。')
      }
    }
  })
}

function confirmRestoreFromCooldown() {
  if (!family.value || family.value.role !== 'FOUNDER') return
  uni.showModal({
    title: '恢复家庭',
    content: '恢复后家庭将重新回到正常状态，并恢复到家庭列表。确认恢复吗？',
    confirmColor: '#163353',
    success: async (result) => {
      if (!result.confirm) return
      try {
        family.value = await restoreDissolution(familyId.value)
        currentDissolution.value = null
        uni.showToast({ title: '家庭已恢复', icon: 'success' })
      } catch (error) {
        operationError.value = apiErrorMessage(error, '恢复家庭失败，请稍后重试。')
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

function applicationStatusClass(status: string) {
  if (status === 'APPROVED') return 'status-approved'
  if (status === 'REJECTED') return 'status-rejected'
  if (status === 'CANCELLED') return 'status-cancelled'
  return 'status-pending'
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
    FAMILY_PUBLIC_APPLICATION: '优质家庭申请',
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
    PLACE_EXISTING_MEMBER: '定位已有成员',
    SUBMIT_PUBLIC_APPLICATION: '提交展示权限申请',
    CANCEL_PUBLIC_APPLICATION: '取消展示权限申请',
    APPROVE_PUBLIC_APPLICATION: '展示权限审核通过',
    REJECT_PUBLIC_APPLICATION: '展示权限审核驳回',
    ENABLE_PUBLIC_DISPLAY: '开启优质家庭展示',
    DISABLE_PUBLIC_DISPLAY: '关闭优质家庭展示',
    CLOSE_PUBLIC_FAMILY: '关闭优质家庭展示',
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

function formatDateTime(value?: string | null) {
  if (!value) return '待确定'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '待确定'
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false
  })
}

function openFamilyOverview() {
  uni.redirectTo({ url: returnUrl.value })
}

onLoad((options) => {
  familyId.value = String(options?.familyId || '')
  returnTo.value = String(options?.returnTo || '')
  const section = String(options?.section || '')
  if (['public', 'roles', 'transfer', 'dissolution', 'logs'].includes(section)) {
    settingsSection.value = section as typeof settingsSection.value
  }
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

.settings-section-guide {
  display: flex;
  flex-direction: column;
  gap: 5rpx;
  margin: 16rpx 0 20rpx;
  padding: 14rpx 16rpx;
  border-left: 3rpx solid var(--archive-line-strong);
  background: rgba(255, 255, 255, 0.26);
}

.settings-section-guide-title {
  color: var(--archive-ink);
  font-size: 24rpx;
  font-weight: 700;
}

.settings-section-guide-desc {
  color: var(--archive-ink-soft);
  font-size: 22rpx;
  line-height: 1.45;
}

.settings-action {
  margin-top: 20rpx;
  align-self: flex-start;
}

.quality-eligibility-panel {
  margin-bottom: 20rpx;
}

.quality-eligibility-list {
  display: flex;
  flex-direction: column;
  gap: 10rpx;
  color: var(--archive-ink-soft);
  font-size: 24rpx;
  line-height: 1.5;
}

.quality-eligibility-item {
  display: flex;
  gap: 8rpx;
  align-items: flex-start;
}

.quality-eligibility-dot {
  color: var(--archive-cinnabar);
  font-size: 30rpx;
  line-height: 1.1;
}

.quality-check-panel {
  margin-top: 18rpx;
  padding: 18rpx 20rpx;
  border: 1rpx solid var(--archive-line);
  border-radius: 10rpx;
  background: rgba(255, 255, 255, 0.36);
}

.quality-check-panel-head {
  display: flex;
  flex-direction: column;
  gap: 6rpx;
  padding-bottom: 12rpx;
  border-bottom: 1rpx solid var(--archive-line);
}

.quality-check-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16rpx;
  padding: 14rpx 0;
  border-bottom: 1rpx solid rgba(31, 54, 79, 0.1);
}

.quality-check-row:last-child { border-bottom: 0; }

.quality-check-copy {
  display: flex;
  flex-direction: column;
  gap: 4rpx;
  min-width: 0;
}

.quality-check-label { color: var(--archive-ink); font-size: 25rpx; font-weight: 600; }
.quality-check-value { color: var(--archive-ink-soft); font-size: 22rpx; }
.quality-check-status { flex: 0 0 auto; font-size: 22rpx; font-weight: 600; }
.quality-check-met { color: var(--archive-green, #2f6b57); }
.quality-check-unmet { color: var(--archive-cinnabar); }

.settings-history {
  margin-top: 20rpx;
}

.application-history-item {
  padding: 18rpx 0;
  border-top: 1rpx solid var(--archive-line);
}

.application-history-item:first-of-type { border-top: 0; }
.application-history-head { display: flex; align-items: center; justify-content: space-between; gap: 16rpx; }
.application-history-meta,
.application-history-text { display: block; margin-top: 7rpx; color: var(--archive-ink-soft); font-size: 22rpx; line-height: 1.45; }
.application-history-text { color: var(--archive-ink); }
.status-approved { color: var(--archive-green, #2f6b57); background: rgba(47, 107, 87, 0.1); }
.status-rejected { color: var(--archive-cinnabar); background: rgba(168, 59, 45, 0.1); }
.status-cancelled { color: var(--archive-ink-soft); background: rgba(90, 98, 107, 0.1); }
.status-pending { color: #8b6420; background: rgba(184, 135, 47, 0.12); }

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
