<template>
  <view class="tree-page">
    <MiniBackHome />

    <MiniCard v-if="!authChecked">
      <MiniEmptyState symbol="…" title="正在确认登录状态" description="请稍候..." />
    </MiniCard>

    <MiniCard v-else-if="loading">
      <MiniEmptyState symbol="…" title="正在加载成员" description="正在读取成员资料..." />
    </MiniCard>

    <MiniCard v-else-if="loadError">
      <MiniNotice tone="warm" title="加载失败">{{ loadError }}</MiniNotice>
      <MiniButton variant="secondary" @click="loadMember">重新加载</MiniButton>
    </MiniCard>

    <template v-else>
      <view class="edit-hero">
        <view class="edit-avatar">{{ form.name.slice(0, 1) || '员' }}</view>
        <view class="edit-hero-copy">
          <text class="edit-kicker">成员资料</text>
          <text class="edit-title">{{ form.name || '未命名成员' }}</text>
          <text class="edit-desc">修改姓名、性别、出生年份、简介和健在状态。</text>
        </view>
      </view>

      <MiniCard variant="soft" class="edit-card">
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
          <switch :checked="form.isAlive" color="#2F6B57" @change="onAliveChange" />
        </label>

        <text class="tree-field-label">成员简介</text>
        <textarea v-model.trim="form.description" class="tree-textarea" maxlength="500" placeholder="可选" />

        <text v-if="actionError" class="tree-field-error">{{ actionError }}</text>
        <MiniButton :loading="saving" :disabled="saving" @click="saveMember">保存成员资料</MiniButton>
      </MiniCard>

      <MiniCard variant="soft" class="danger-card">
        <view class="danger-copy">
          <text class="danger-title">删除成员节点</text>
          <text class="danger-desc">只删除家谱节点。存在亲属关系、创建者或管理员身份时，系统会拒绝删除。</text>
        </view>
        <MiniButton
          variant="danger"
          :loading="deleting"
          :disabled="deleting"
          @click="confirmDeleteMember"
        >
          删除该成员
        </MiniButton>
      </MiniCard>
    </template>
  </view>
</template>

<script setup lang="ts">
import { onLoad, onShow } from '@dcloudio/uni-app'
import { reactive, ref } from 'vue'

import { apiErrorMessage } from '@/api/client'
import { deleteFamilyMember, getFamilyMember, updateFamilyMember } from '@/api/members'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniCard from '@/components/base/MiniCard.vue'
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
import { useSessionStore } from '@/stores/session'
import type { Gender } from '@/types/api'

const session = useSessionStore()
const familyId = ref('')
const memberId = ref('')
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
    const member = await getFamilyMember(familyId.value, memberId.value)
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
  if (!form.name.trim()) {
    actionError.value = '请填写成员姓名。'
    return
  }
  saving.value = true
  actionError.value = ''
  try {
    await updateFamilyMember(familyId.value, memberId.value, {
      name: form.name,
      gender: form.gender,
      birthYear: yearValue(birthYearInput.value),
      isAlive: form.isAlive,
      description: form.description || undefined
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
.edit-hero {
  display: flex;
  align-items: center;
  gap: 18rpx;
  margin-bottom: 24rpx;
  border: 1rpx solid rgba(255, 255, 255, 0.20);
  border-radius: 34rpx;
  background:
    radial-gradient(circle at 92% 10%, rgba(216, 175, 104, 0.20), transparent 220rpx),
    linear-gradient(135deg, #17304c 0%, #245653 100%);
  padding: 30rpx;
  color: #fff;
  box-shadow: 0 24rpx 60rpx rgba(24, 54, 83, 0.18);
}

.edit-avatar {
  display: flex;
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  width: 78rpx;
  height: 78rpx;
  border-radius: 26rpx;
  background: rgba(255, 255, 255, 0.16);
  color: #fff;
  font-size: 32rpx;
  font-weight: 900;
}

.edit-hero-copy {
  display: flex;
  flex-direction: column;
  gap: 6rpx;
  min-width: 0;
}

.edit-kicker {
  color: rgba(248, 231, 194, 0.9);
  font-size: 22rpx;
  font-weight: 800;
}

.edit-title {
  color: #fff;
  font-size: 34rpx;
  font-weight: 900;
  line-height: 1.2;
}

.edit-desc {
  color: rgba(255, 255, 255, 0.72);
  font-size: 24rpx;
  line-height: 1.45;
}

.edit-card,
.danger-card {
  margin-bottom: 20rpx;
}

.field-picker {
  box-sizing: border-box;
  min-height: 84rpx;
  margin-bottom: 18rpx;
  border: 1rpx solid rgba(31, 58, 95, 0.10);
  border-radius: 18rpx;
  background: rgba(255, 255, 255, 0.9);
  color: var(--tree-text-primary, #1e293b);
  padding: 22rpx 24rpx;
  font-size: 27rpx;
  line-height: 1.4;
}

.switch-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 84rpx;
  margin-bottom: 18rpx;
  border: 1rpx solid rgba(31, 58, 95, 0.08);
  border-radius: 18rpx;
  background: rgba(255, 255, 255, 0.78);
  color: var(--tree-text-primary, #1e293b);
  padding: 0 24rpx;
  font-size: 27rpx;
}

.danger-card {
  border-color: rgba(180, 83, 58, 0.18);
  background:
    radial-gradient(circle at 100% 0%, rgba(180, 83, 58, 0.08), transparent 160rpx),
    rgba(255, 255, 255, 0.92);
}

.danger-copy {
  display: flex;
  flex-direction: column;
  gap: 8rpx;
  margin-bottom: 18rpx;
}

.danger-title {
  color: #9a3412;
  font-size: 28rpx;
  font-weight: 900;
}

.danger-desc {
  color: var(--tree-text-secondary, #64748b);
  font-size: 24rpx;
  line-height: 1.55;
}
</style>
