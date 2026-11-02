<template>
  <view class="archive-page create-page">
    <MiniBackHome />

    <view v-if="!authChecked" class="create-state archive-form-panel">
      <MiniEmptyState symbol="…" title="正在确认资料" description="请稍候..." />
    </view>

    <template v-else>
      <view class="create-head archive-page-head">
        <view>
          <text class="archive-kicker">New Family Tree</text>
          <text class="archive-title">创建家庭</text>
          <text class="archive-subtitle">填写姓氏与基本信息，建立家庭树</text>
        </view>
        <view class="archive-seal create-seal" :class="{ muted: !form.surname.trim() }">{{ coverSurnameLetter }}</view>
      </view>

      <view class="create-cover">
        <view class="create-cover-copy">
          <text class="create-surname">{{ coverSurnameLabel }}</text>
          <text class="create-name">{{ coverFamilyName }}</text>
          <text class="create-region">{{ coverRegionSummary }}</text>
          <view class="create-status-line">
            <text class="create-status-badge">新建册页</text>
            <text v-if="form.nativePlace" class="archive-chip">{{ form.nativePlace }}</text>
            <text v-if="form.regionText" class="archive-chip">{{ form.regionText }}</text>
          </view>
        </view>
        <view class="create-spine archive-book-spine">
          <text>新</text>
          <text>家</text>
          <text>树</text>
        </view>
      </view>

      <view class="archive-form-panel">
        <view class="archive-section-head">
          <text class="archive-section-title">家庭信息</text>
          <text class="archive-section-subtitle">创建后将自动生成创建者成员节点</text>
        </view>

        <text class="tree-field-label">家庭姓氏</text>
        <input v-model.trim="form.surname" class="tree-input" maxlength="20" placeholder="必填，如：张" />

        <text class="tree-field-label">你的姓名</text>
        <input v-model.trim="form.founderName" class="tree-input" maxlength="80" placeholder="必填，用于家庭树中的本人节点" />

        <text class="tree-field-label">你的性别</text>
        <picker mode="selector" :range="genderLabels" :value="genderIndex" @change="onGenderChange">
          <view class="field-picker">{{ genderLabels[genderIndex] }}</view>
        </picker>
        <text class="tree-field-label">出生年份</text>
        <input v-model.trim="form.founderBirthYear" class="tree-input" type="number" maxlength="4" placeholder="必填，如：1990" />
        <MiniNotice tone="info">姓名、性别和出生年份用于正确初始化家庭树中的本人节点，创建后仍可在成员资料中修改。</MiniNotice>

        <text class="tree-field-label">家庭名称</text>
        <input v-model.trim="form.familyName" class="tree-input" maxlength="100" placeholder="可选，默认生成“某氏家族”" />

        <FamilyRegionSelector v-model="selectedRegion" />

        <text class="tree-field-label">家庭简介</text>
        <textarea v-model.trim="form.description" class="tree-textarea" maxlength="500" placeholder="可选" />

        <MiniNotice v-if="draftRestored" tone="info">已恢复上次未提交的内容。</MiniNotice>
        <text v-if="errorMessage" class="tree-field-error">{{ errorMessage }}</text>
        <MiniButton v-if="hasDraft" class="clear-draft-action" variant="ghost" :disabled="submitting" @click="clearDraft">清空草稿</MiniButton>
        <MiniButton
          class="create-action"
          :loading="submitting || quotaChecking"
          :disabled="submitting || quotaChecking || quotaBlocked"
          @click="submit"
        >
          创建家庭
        </MiniButton>
      </view>
    </template>
  </view>
</template>

<script setup lang="ts">
import { onShow } from '@dcloudio/uni-app'
import { computed, reactive, ref, watch } from 'vue'

import { apiErrorMessage } from '@/api/client'
import { fetchCapabilities } from '@/api/capabilities'
import { createFamily } from '@/api/families'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
import FamilyRegionSelector, { type FamilyRegionValue } from '@/components/family/FamilyRegionSelector.vue'
import { markNewUserGuideMilestone } from '@/features/onboarding/newUserGuide'
import { useSessionStore } from '@/stores/session'
import type { Gender } from '@/types/api'
import { normalizeText, optionalText, validateTextFields } from '@/utils/inputValidation'

const session = useSessionStore()
const authChecked = ref(false)
const submitting = ref(false)
const errorMessage = ref('')
const quotaChecking = ref(false)
const quotaBlocked = ref(false)
const genderIndex = ref(0)
const genders: Gender[] = ['MALE', 'FEMALE']
const genderLabels = ['男', '女']
const draftRestored = ref(false)
const hasDraft = ref(false)
const draftReady = ref(false)
const form = reactive({
  surname: '',
  founderName: '',
  founderGender: 'MALE' as Gender,
  founderBirthYear: '',
  familyName: '',
  nativePlace: '',
  regionCode: '',
  regionText: '',
  description: ''
})

function draftKey() {
  return `tree_miniapp_family_create_draft_${session.user?.id || 'guest'}`
}

function restoreDraft() {
  if (draftReady.value) return
  draftReady.value = true
  try {
    const stored = uni.getStorageSync(draftKey())
    const parsed = typeof stored === 'string' ? JSON.parse(stored) : stored
    if (!parsed || typeof parsed !== 'object') return
    const draft = parsed as Partial<typeof form>
    ;(['surname', 'founderName', 'founderBirthYear', 'familyName', 'nativePlace', 'regionCode', 'regionText', 'description'] as const).forEach((key) => {
      if (typeof draft[key] === 'string') form[key] = draft[key]
    })
    if (draft.founderGender === 'MALE' || draft.founderGender === 'FEMALE') {
      form.founderGender = draft.founderGender
      genderIndex.value = genders.indexOf(draft.founderGender)
    }
    hasDraft.value = true
    draftRestored.value = true
  } catch {
    uni.removeStorageSync(draftKey())
  }
}

function clearDraft() {
  uni.removeStorageSync(draftKey())
  hasDraft.value = false
  draftRestored.value = false
  Object.assign(form, { surname: '', founderName: '', founderGender: 'MALE', founderBirthYear: '', familyName: '', nativePlace: '', regionCode: '', regionText: '', description: '' })
  genderIndex.value = 0
}

watch(form, () => {
  if (!draftReady.value || submitting.value) return
  const hasContent = Object.values(form).some((value) => String(value).trim() !== '' && value !== 'MALE')
  if (!hasContent) {
    uni.removeStorageSync(draftKey())
    hasDraft.value = false
    return
  }
  uni.setStorageSync(draftKey(), JSON.stringify({ ...form }))
  hasDraft.value = true
}, { deep: true })

const coverSurnameLetter = computed(() => form.surname.trim().slice(0, 1) || '姓')

const coverSurnameLabel = computed(() => {
  const surname = form.surname.trim()
  return surname ? `${surname}氏` : '某氏'
})

const coverFamilyName = computed(() => {
  const customName = form.familyName.trim()
  if (customName) return customName
  const surname = form.surname.trim()
  return surname ? `${surname}氏家族` : '新家庭'
})

const coverRegionSummary = computed(() => {
  return form.regionText.trim() || form.nativePlace.trim() || '籍贯与地区（可选）'
})

const selectedRegion = computed<FamilyRegionValue>({
  get: () => ({ nativePlace: form.nativePlace, regionCode: form.regionCode, regionText: form.regionText }),
  set: (value) => Object.assign(form, value)
})

function onGenderChange(event: { detail: { value: number | string } }) {
  genderIndex.value = Number(event.detail.value) || 0
  form.founderGender = genders[genderIndex.value] || 'MALE'
}

async function submit() {
  if (quotaBlocked.value) {
    showQuotaBlockedModal()
    return
  }
  const validationMessage = validateTextFields([
    { value: form.surname, label: '家庭姓氏', kind: 'surname', required: true, maxLength: 20 },
    { value: form.founderName, label: '你的姓名', kind: 'name', required: true, maxLength: 80 },
    { value: form.familyName, label: '家庭名称', kind: 'name', maxLength: 100 },
    { value: form.nativePlace, label: '籍贯', maxLength: 100 },
    { value: form.regionText, label: '地区', maxLength: 100 },
    { value: form.description, label: '家庭简介', kind: 'multiLine', maxLength: 500 }
  ])
  if (validationMessage) {
    errorMessage.value = validationMessage
    return
  }
  if (!form.regionCode || !form.nativePlace || !form.regionText) {
    errorMessage.value = '请选择地理分区、省市籍贯。'
    return
  }
  const founderBirthYear = Number(form.founderBirthYear)
  const currentYear = new Date().getFullYear()
  if (!Number.isInteger(founderBirthYear) || founderBirthYear < 1 || founderBirthYear > currentYear) {
    errorMessage.value = `请输入 1 到 ${currentYear} 之间的出生年份。`
    return
  }
  submitting.value = true
  errorMessage.value = ''
  try {
    const family = await createFamily({
      surname: normalizeText(form.surname),
      founderName: normalizeText(form.founderName),
      founderGender: form.founderGender,
      founderBirthYear,
      familyName: optionalText(form.familyName),
      nativePlace: optionalText(form.nativePlace),
      regionCode: optionalText(form.regionCode),
      regionText: optionalText(form.regionText),
      description: optionalText(form.description)
    })
    uni.removeStorageSync(draftKey())
    hasDraft.value = false
    markNewUserGuideMilestone('FAMILY_CREATED')
    uni.redirectTo({ url: `/pages/family/detail?familyId=${encodeURIComponent(String(family.id))}` })
  } catch (error) {
    errorMessage.value = apiErrorMessage(error, '创建家庭失败。')
  } finally {
    submitting.value = false
  }
}

async function checkCreateQuota() {
  quotaChecking.value = true
  quotaBlocked.value = false
  errorMessage.value = ''
  try {
    const capabilities = await fetchCapabilities()
    quotaBlocked.value = capabilities.usage.ownedFamilies >= capabilities.limits.maxOwnedFamilies
    if (quotaBlocked.value) {
      errorMessage.value = `当前账号可创建家庭 ${capabilities.usage.ownedFamilies}/${capabilities.limits.maxOwnedFamilies}，暂不能继续创建。`
      showQuotaBlockedModal(capabilities.usage.ownedFamilies, capabilities.limits.maxOwnedFamilies)
    }
  } catch (error) {
    errorMessage.value = apiErrorMessage(error, '无法确认创建资格，请稍后重试。')
    quotaBlocked.value = true
  } finally {
    quotaChecking.value = false
  }
}

function showQuotaBlockedModal(ownedFamilies?: number, maxOwnedFamilies?: number) {
  const usageText =
    typeof ownedFamilies === 'number' && typeof maxOwnedFamilies === 'number'
      ? `当前账号可创建家庭 ${ownedFamilies}/${maxOwnedFamilies}。`
      : '当前账号已达到可创建家庭数量上限。'
  uni.showModal({
    title: '暂不能创建家庭',
    content: `${usageText}冷静期家庭仍会占用创建名额，完成解散后才会释放。`,
    confirmText: '知道了',
    showCancel: false
  })
}

onShow(async () => {
  authChecked.value = session.requireProfileComplete('/pages/family/create')
  if (authChecked.value) {
    restoreDraft()
    await checkCreateQuota()
  }
})
</script>

<style scoped>
.create-page {
  padding-top: 28rpx;
}

.create-page :deep(.mini-back-home) {
  margin-bottom: 22rpx;
}

.clear-draft-action { margin-top: 18rpx; }

.create-cover {
  display: flex;
  min-height: 210rpx;
  margin-bottom: 24rpx;
  border-top: 1rpx solid var(--archive-line-strong);
  border-bottom: 1rpx solid var(--archive-line);
}

.create-cover-copy {
  flex: 1;
  min-width: 0;
  padding: 28rpx 28rpx 28rpx 0;
}

.create-surname,
.create-name,
.create-region {
  display: block;
}

.create-surname {
  color: var(--archive-cinnabar);
  font-size: 24rpx;
  font-weight: 700;
  letter-spacing: 2rpx;
}

.create-name {
  margin-top: 10rpx;
  color: var(--archive-blue);
  font-family: 'Songti SC', 'STSong', serif;
  font-size: 42rpx;
  font-weight: 800;
  letter-spacing: 2rpx;
  line-height: 1.28;
}

.create-region {
  margin-top: 12rpx;
  color: var(--archive-ink-soft);
  font-size: 23rpx;
  line-height: 1.55;
}

.create-status-line {
  display: flex;
  flex-wrap: wrap;
  gap: 10rpx;
  margin-top: 16rpx;
}

.create-status-badge {
  display: inline-flex;
  align-items: center;
  min-height: 42rpx;
  border: 1rpx dashed var(--archive-line-strong);
  background: rgba(255, 248, 234, 0.58);
  color: var(--archive-ink-soft);
  padding: 0 14rpx;
  font-size: 21rpx;
  line-height: 1.2;
}

.create-seal.muted {
  opacity: 0.72;
}

.create-spine {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 10rpx;
  width: 118rpx;
  flex-shrink: 0;
}

.create-spine text {
  color: rgba(255, 255, 255, 0.94);
  font-family: 'Songti SC', 'STSong', serif;
  font-size: 28rpx;
  font-weight: 800;
}

.create-page :deep(.mini-notice) {
  margin-top: 16rpx;
  margin-bottom: 16rpx;
}

.create-action {
  margin-top: 20rpx;
  align-self: flex-start;
}
</style>
