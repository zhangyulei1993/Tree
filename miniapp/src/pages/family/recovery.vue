<template>
  <view class="archive-page recovery-page">
    <MiniBackHome />
    <view class="archive-page-head">
      <view>
        <text class="archive-kicker">Recovery Center</text>
        <text class="archive-title">最近删除</text>
        <text class="archive-subtitle">仅家庭创建者可在删除后 7 天内恢复成员；账号绑定不会恢复。</text>
      </view>
    </view>
    <view v-if="errorMessage" class="archive-form-panel"><MiniNotice tone="warm" title="无法加载">{{ errorMessage }}</MiniNotice></view>
    <view v-else-if="loading" class="archive-form-panel"><MiniEmptyState symbol="…" title="正在读取恢复记录" /></view>
    <view v-else-if="!items.length" class="archive-form-panel"><MiniEmptyState symbol="✓" title="暂无可恢复成员" description="最近 7 天没有被删除的成员节点。" /></view>
    <view v-else class="archive-list">
      <view v-for="item in items" :key="item.memberId" class="archive-row">
        <view class="archive-row-main"><text class="archive-row-title">{{ item.name }}</text><text class="archive-row-desc">删除于 {{ formatDate(item.deletedAt) }} · {{ item.deleteReason || '未填写原因' }}</text></view>
        <view class="recovery-actions">
          <MiniButton size="sm" :loading="restoringId === item.memberId" :disabled="clearingId === item.memberId" @click="confirmRestore(item)">恢复</MiniButton>
          <MiniButton size="sm" variant="secondary" :loading="clearingId === item.memberId" :disabled="restoringId === item.memberId" @click="confirmDiscard(item)">清除</MiniButton>
        </view>
      </view>
    </view>
  </view>
</template>
<script setup lang="ts">
import { onLoad, onShow } from '@dcloudio/uni-app'
import { ref } from 'vue'
import { apiErrorMessage } from '@/api/client'
import { discardFamilyMemberRecovery, listRecoverableFamilyMembers, restoreFamilyMember, type RecoverableMember } from '@/api/members'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
const familyId = ref(''); const items = ref<RecoverableMember[]>([]); const loading = ref(false); const restoringId = ref<number | null>(null); const clearingId = ref<number | null>(null); const errorMessage = ref('')
const formatDate = (value: string) => new Date(value).toLocaleString('zh-CN', { hour12: false })
async function load() { if (!familyId.value) return; loading.value = true; errorMessage.value = ''; try { items.value = await listRecoverableFamilyMembers(familyId.value) } catch (e) { errorMessage.value = apiErrorMessage(e, '恢复记录加载失败。') } finally { loading.value = false } }
function confirmRestore(item: RecoverableMember) { uni.showModal({ title: '恢复成员', content: `恢复“${item.name}”及同次删除的关联关系？如原有位置已被新节点占用，系统会拒绝恢复，避免重复成员。已解绑的账号不会重新绑定。`, success: async result => { if (!result.confirm) return; restoringId.value = item.memberId; try { await restoreFamilyMember(familyId.value, item.memberId); uni.showToast({ title: '成员已恢复', icon: 'success' }); await load() } catch (e) { uni.showModal({ title: '恢复失败', content: apiErrorMessage(e, '恢复成员失败。'), showCancel: false }) } finally { restoringId.value = null } } }) }
function confirmDiscard(item: RecoverableMember) { uni.showModal({ title: '清除恢复记录', content: `清除“${item.name}”的恢复资格？该节点会继续保留为已删除状态和操作记录，但 7 天内也不能再恢复。`, confirmText: '确认清除', confirmColor: '#b84235', success: async result => { if (!result.confirm) return; clearingId.value = item.memberId; try { await discardFamilyMemberRecovery(familyId.value, item.memberId); uni.showToast({ title: '已清除', icon: 'success' }); await load() } catch (e) { uni.showModal({ title: '清除失败', content: apiErrorMessage(e, '清除恢复记录失败。'), showCancel: false }) } finally { clearingId.value = null } } }) }
onLoad(options => { familyId.value = String(options?.familyId || '') }); onShow(load)
</script>
<style scoped>.recovery-page{padding-top:28rpx}.archive-list{margin-top:24rpx}.archive-row{align-items:center}.archive-row-main{min-width:0}.recovery-actions{display:flex;align-items:center;gap:12rpx;flex:0 0 auto}</style>
