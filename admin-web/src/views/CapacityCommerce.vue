<template>
  <div class="page-stack">
    <PageHeader title="容量商品与支付订单" description="查看家庭容量商品配置、支付结果与退款状态。商品金额单位为分。" />
    <el-alert v-if="errorMessage" :title="errorMessage" type="error" show-icon :closable="false" />
    <el-card shadow="never">
      <template #header><div class="card-header"><strong>容量商品</strong><el-button :loading="loadingProducts" @click="loadProducts">刷新</el-button></div></template>
      <el-table v-loading="loadingProducts" :data="products" border stripe>
        <el-table-column prop="productKey" label="商品标识" min-width="220" />
        <el-table-column prop="displayName" label="展示名称" min-width="180" />
        <el-table-column prop="addedMembers" label="增加成员额度" width="140" align="center" />
        <el-table-column label="价格" width="120" align="center"><template #default="{ row }">¥{{ (row.amountFen / 100).toFixed(2) }}</template></el-table-column>
        <el-table-column prop="refundWindowMinutes" label="退款窗口（分钟）" width="160" align="center" />
        <el-table-column label="状态" width="100" align="center"><template #default="{ row }"><el-tag :type="row.status === 'ACTIVE' ? 'success' : 'info'">{{ row.status === 'ACTIVE' ? '启用' : '停用' }}</el-tag></template></el-table-column>
        <el-table-column label="操作" width="100" fixed="right" align="center"><template #default="{ row }"><el-button type="primary" link @click="editProduct(row)">编辑</el-button></template></el-table-column>
      </el-table>
    </el-card>
    <el-card shadow="never">
      <template #header><div class="card-header"><strong>家庭邀请奖励</strong><el-button :loading="loadingReward" @click="loadReward">刷新</el-button></div></template>
      <el-alert type="info" :closable="false" show-icon title="奖励容量独立于付费扩容，不生成订单、不参与退款。仅在家庭剩余名额达到触发条件时开放分享。" />
      <el-form class="reward-form" label-width="190px">
        <el-form-item label="启用奖励"><el-switch v-model="rewardForm.enabled" active-text="启用" inactive-text="停用" /></el-form-item>
        <el-form-item label="开放分享的剩余名额"><el-input-number v-model="rewardForm.activationRemainingSlots" :min="0" :max="20" /> <el-text type="info">剩余名额不高于该值时开放</el-text></el-form-item>
        <el-form-item label="家庭累计奖励上限"><el-input-number v-model="rewardForm.maxRewardMembers" :min="1" :max="100" /> <el-text type="info">位成员</el-text></el-form-item>
        <el-form-item label="有效邀请条件"><el-switch v-model="rewardForm.requirePhoneBound" active-text="需绑定手机号" inactive-text="无需手机号" /></el-form-item>
        <el-form-item label="进度档位"><div class="tier-list"><div v-for="(tier,index) in rewardForm.tiers" :key="index" class="tier-row"><el-input-number v-model="tier.requiredInvites" :min="1" :max="100000" /><span>个有效邀请，累计奖励</span><el-input-number v-model="tier.cumulativeRewardMembers" :min="1" :max="rewardForm.maxRewardMembers" /><span>位</span><el-button type="danger" link :disabled="rewardForm.tiers.length <= 1" @click="rewardForm.tiers.splice(index, 1)">删除</el-button></div><el-button link type="primary" :disabled="rewardForm.tiers.length >= 10" @click="addTier">新增档位</el-button></div></el-form-item>
        <el-form-item><el-button type="primary" :loading="savingReward" @click="saveReward">保存邀请奖励配置</el-button></el-form-item>
      </el-form>
    </el-card>
    <el-dialog v-model="productDialogVisible" title="编辑容量商品" width="520px" destroy-on-close>
      <el-form label-width="150px">
        <el-form-item label="商品标识"><el-text>{{ productForm.productKey }}</el-text></el-form-item>
        <el-form-item label="展示名称"><el-input v-model="productForm.displayName" maxlength="80" show-word-limit /></el-form-item>
        <el-form-item label="增加成员人数"><el-input-number v-model="productForm.addedMembers" :min="1" :max="10000" /></el-form-item>
        <el-form-item label="价格（元）"><el-input-number v-model="productForm.amountYuan" :min="0.01" :max="999999.99" :precision="2" :step="0.01" /></el-form-item>
        <el-form-item label="退款窗口（分钟）"><el-input-number v-model="productForm.refundWindowMinutes" :min="1" :max="43200" /></el-form-item>
        <el-form-item label="商品状态"><el-switch v-model="productForm.active" active-text="启用" inactive-text="停用" /></el-form-item>
      </el-form>
      <template #footer><el-button @click="productDialogVisible = false">取消</el-button><el-button type="primary" :loading="savingProduct" @click="saveProduct">保存配置</el-button></template>
    </el-dialog>
    <el-card shadow="never">
      <template #header><div class="card-header"><strong>支付订单</strong><div class="order-tools"><el-radio-group v-model="statusFilter" size="small" @change="applyOrderFilters"><el-radio-button value="ALL">全部</el-radio-button><el-radio-button value="PENDING">待支付</el-radio-button><el-radio-button value="PAID">已支付</el-radio-button><el-radio-button value="REFUNDING">退款中</el-radio-button><el-radio-button value="REFUNDED">已退款</el-radio-button><el-radio-button value="CANCELLED">已取消</el-radio-button></el-radio-group><el-date-picker v-model="dateRange" type="daterange" value-format="YYYY-MM-DD" range-separator="至" start-placeholder="开始日期" end-placeholder="结束日期" clearable @change="applyOrderFilters" /><el-button :loading="loadingOrders" @click="loadOrders">刷新</el-button></div></div></template>
      <el-table v-loading="loadingOrders" :data="orders" border stripe>
        <el-table-column prop="orderNo" label="订单号" min-width="260" />
        <el-table-column prop="productName" label="商品" min-width="170" />
        <el-table-column prop="familyId" label="家庭 ID" width="100" align="center" />
        <el-table-column prop="buyerUserId" label="用户 ID" width="100" align="center" />
        <el-table-column label="金额" width="100" align="center"><template #default="{ row }">¥{{ (row.amountFen / 100).toFixed(2) }}</template></el-table-column>
        <el-table-column label="状态" width="120" align="center"><template #default="{ row }"><el-tag :type="statusTagType(row.status)">{{ statusLabel(row.status) }}</el-tag></template></el-table-column>
        <el-table-column prop="createdAt" label="创建时间" min-width="180" />
        <el-table-column prop="expiresAt" label="支付截止时间" min-width="180" />
        <el-table-column prop="paidAt" label="支付时间" min-width="180" />
        <el-table-column prop="closedAt" label="关闭时间" min-width="180" />
        <el-table-column label="支付与退款处理" min-width="210" fixed="right" align="center">
          <template #default="{ row }">
            <template v-if="['PENDING', 'CANCELLED'].includes(row.status)">
              <el-button type="primary" link @click="reconcilePayment(row)">查询支付结果</el-button>
            </template>
            <template v-if="row.refundStatus === 'PENDING' && row.refundNo">
              <el-button type="primary" link @click="approveRefund(row)">通过并退款</el-button>
              <el-button type="danger" link @click="rejectRefund(row)">拒绝</el-button>
            </template>
            <template v-else-if="row.refundNo && ['REVIEW_REQUIRED', 'PROCESSING', 'FAILED'].includes(row.refundStatus)">
              <el-tag :type="row.refundStatus === 'FAILED' ? 'info' : 'warning'">{{ refundStatusLabel(row.refundStatus) }}</el-tag>
              <div class="review-actions">
                <el-button type="primary" link @click="reconcileRefund(row)">查询微信结果</el-button>
                <el-button v-if="row.refundStatus === 'REVIEW_REQUIRED' && missingRefunds[row.refundNo]" type="warning" link @click="retryRefund(row)">同单号重试</el-button>
              </div>
            </template>
            <el-text v-else-if="row.refundStatus" type="info">{{ refundStatusLabel(row.refundStatus) }}</el-text>
            <el-text v-else type="info">无退款申请</el-text>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-if="!loadingOrders && orders.length === 0" description="暂无符合条件的订单" />
      <el-pagination v-if="orderTotal > 0" v-model:current-page="orderPage" v-model:page-size="orderPageSize" :total="orderTotal" :page-sizes="[20, 50, 100]" layout="total, sizes, prev, pager, next, jumper" @current-change="loadOrders" @size-change="changeOrderPageSize" />
    </el-card>
  </div>
</template>
<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus/es/components/message/index'
import { ElMessageBox } from 'element-plus/es/components/message-box/index'
import { getApiErrorMessage } from '@/api/client'
import { approveCapacityRefund, listCapacityOrders, listCapacityProducts, reconcileCapacityPayment, reconcileCapacityRefund, rejectCapacityRefund, retryCapacityRefund, updateCapacityProduct, type AdminCapacityOrder, type AdminCapacityProduct } from '@/api/capacity'
import { getInvitationRewardConfig, updateInvitationRewardConfig, type InvitationRewardConfig } from '@/api/invitationReward'
import PageHeader from '@/components/PageHeader.vue'
const products = ref<AdminCapacityProduct[]>([]); const orders = ref<AdminCapacityOrder[]>([]); const loadingProducts = ref(false); const loadingOrders = ref(false); const errorMessage = ref(''); const statusFilter = ref('ALL'); const dateRange = ref<[string, string] | null>(null); const orderPage = ref(1); const orderPageSize = ref(20); const orderTotal = ref(0); const productDialogVisible = ref(false); const savingProduct = ref(false); const productForm = ref({ productKey: '', displayName: '', addedMembers: 1, amountYuan: 0.01, refundWindowMinutes: 30, active: true })
const missingRefunds = ref<Record<string, boolean>>({})
const loadingReward = ref(false); const savingReward = ref(false); const rewardForm = ref<InvitationRewardConfig>({ enabled: true, activationRemainingSlots: 1, maxRewardMembers: 3, requirePhoneBound: true, tiers: [{ requiredInvites: 6, cumulativeRewardMembers: 1 }, { requiredInvites: 10, cumulativeRewardMembers: 2 }, { requiredInvites: 18, cumulativeRewardMembers: 3 }] })
function statusLabel(status: string) { return ({ PENDING: '待支付', PAID: '已支付', REFUNDING: '退款审核中', REFUNDED: '已退款', CANCELLED: '已取消' } as Record<string, string>)[status] || status }
function statusTagType(status: string) { return ({ PAID: 'success', REFUNDING: 'warning', REFUNDED: 'info', CANCELLED: 'info', PENDING: 'warning' } as Record<string, string>)[status] || 'info' }
function refundStatusLabel(status: string) { return ({ PENDING: '待审核', PROCESSING: '退款处理中', REVIEW_REQUIRED: '待人工核查', SUCCESS: '已退款', REJECTED: '已拒绝', FAILED: '退款失败' } as Record<string, string>)[status] || status }
async function loadProducts() { loadingProducts.value = true; try { products.value = await listCapacityProducts() } catch (error) { errorMessage.value = getApiErrorMessage(error) } finally { loadingProducts.value = false } }
async function loadOrders() { loadingOrders.value = true; try { const result = await listCapacityOrders({ page: orderPage.value, pageSize: orderPageSize.value, status: statusFilter.value, startDate: dateRange.value?.[0], endDate: dateRange.value?.[1] }); orders.value = result.items; orderTotal.value = result.total } catch (error) { errorMessage.value = getApiErrorMessage(error) } finally { loadingOrders.value = false } }
function applyOrderFilters() { orderPage.value = 1; void loadOrders() }
function changeOrderPageSize() { orderPage.value = 1; void loadOrders() }
function editProduct(row: AdminCapacityProduct) { productForm.value = { productKey: row.productKey, displayName: row.displayName, addedMembers: row.addedMembers, amountYuan: Number((row.amountFen / 100).toFixed(2)), refundWindowMinutes: row.refundWindowMinutes, active: row.status !== 'INACTIVE' }; productDialogVisible.value = true }
async function saveProduct() {
  if (!productForm.value.displayName.trim()) { errorMessage.value = '展示名称不能为空'; return }
  savingProduct.value = true
  try {
    await updateCapacityProduct(productForm.value.productKey, { displayName: productForm.value.displayName.trim(), addedMembers: productForm.value.addedMembers, amountFen: Math.round(productForm.value.amountYuan * 100), refundWindowMinutes: productForm.value.refundWindowMinutes, status: productForm.value.active ? 'ACTIVE' : 'INACTIVE' })
    ElMessage.success('商品配置已更新，新订单立即使用新配置')
    productDialogVisible.value = false
    await loadProducts()
  } catch (error) { errorMessage.value = getApiErrorMessage(error) } finally { savingProduct.value = false }
}
async function loadReward() { loadingReward.value = true; try { rewardForm.value = await getInvitationRewardConfig() } catch (error) { errorMessage.value = getApiErrorMessage(error) } finally { loadingReward.value = false } }
function addTier() { const last = rewardForm.value.tiers[rewardForm.value.tiers.length - 1]; rewardForm.value.tiers.push({ requiredInvites: (last?.requiredInvites || 0) + 1, cumulativeRewardMembers: Math.min(rewardForm.value.maxRewardMembers, (last?.cumulativeRewardMembers || 0) + 1) }) }
async function saveReward() { savingReward.value = true; try { const tiers = [...rewardForm.value.tiers].sort((a,b) => a.requiredInvites - b.requiredInvites); const last = tiers[tiers.length - 1]; if (!last || last.cumulativeRewardMembers !== rewardForm.value.maxRewardMembers) { errorMessage.value = '最后一个档位的累计奖励必须等于家庭累计奖励上限'; return }; rewardForm.value = await updateInvitationRewardConfig({ ...rewardForm.value, tiers }); ElMessage.success('邀请奖励配置已更新') } catch (error) { errorMessage.value = getApiErrorMessage(error) } finally { savingReward.value = false } }
async function approveRefund(row: AdminCapacityOrder) {
  if (!row.refundNo) return
  try {
    await ElMessageBox.confirm(`确认通过订单 ${row.orderNo} 的退款申请，并发起 ¥${(row.amountFen / 100).toFixed(2)} 原路退款？`, '审核退款', { type: 'warning', confirmButtonText: '通过并退款', cancelButtonText: '取消' })
    await approveCapacityRefund(row.refundNo)
    ElMessage.success('已通过审核，退款处理中')
    await loadOrders()
  } catch (error) {
    if (error !== 'cancel' && error !== 'close') {
      errorMessage.value = '退款提交结果未确认，请在微信支付商户平台核实后处理。'
      await loadOrders()
    }
  }
}
async function rejectRefund(row: AdminCapacityOrder) {
  if (!row.refundNo) return
  try {
    const result = await ElMessageBox.prompt('请输入拒绝原因，用户可据此了解处理结果。', '拒绝退款申请', { inputPattern: /\S+/, inputErrorMessage: '拒绝原因不能为空', confirmButtonText: '确认拒绝', cancelButtonText: '取消' })
    await rejectCapacityRefund(row.refundNo, result.value)
    ElMessage.success('已拒绝退款申请，容量已恢复')
    await loadOrders()
  } catch (error) {
    if (error !== 'cancel' && error !== 'close') errorMessage.value = getApiErrorMessage(error)
  }
}
async function reconcilePayment(row: AdminCapacityOrder) {
  try {
    const result = await reconcileCapacityPayment(row.orderNo)
    if (result.status === 'PAID') ElMessage.success('微信确认支付成功，家庭名额已到账')
    else if (result.status === 'NOTPAY' || result.status === 'NOT_FOUND') ElMessage.info('微信暂未确认支付成功')
    else ElMessage.info(`微信支付状态：${result.status}`)
    await loadOrders()
  } catch (error) { errorMessage.value = getApiErrorMessage(error) }
}
async function reconcileRefund(row: AdminCapacityOrder) {
  if (!row.refundNo) return
  try {
    const result = await reconcileCapacityRefund(row.refundNo)
    missingRefunds.value[row.refundNo] = result.status === 'NOT_FOUND'
    if (result.status === 'SUCCESS') ElMessage.success('微信确认退款成功，扩容额度已扣除')
    else if (result.status === 'FAILED') ElMessage.success('微信确认退款关闭，扩容额度已恢复')
    else if (result.status === 'NOT_FOUND') ElMessage.warning('微信未查到该退款单，额度仍冻结；可用原单号重试提交')
    else ElMessage.info('微信退款仍在处理中，扩容额度保持冻结')
    await loadOrders()
  } catch (error) {
    errorMessage.value = getApiErrorMessage(error)
  }
}
async function retryRefund(row: AdminCapacityOrder) {
  if (!row.refundNo) return
  try {
    await ElMessageBox.confirm('微信暂未查到该退款单。确认沿用原商户退款单号重新提交？', '重试退款', { type: 'warning', confirmButtonText: '重新提交', cancelButtonText: '取消' })
    const result = await retryCapacityRefund(row.refundNo)
    delete missingRefunds.value[row.refundNo]
    if (result.status === 'SUCCESS') ElMessage.success('微信已确认退款成功，扩容额度已扣除')
    else if (result.status === 'FAILED') ElMessage.success('微信确认退款关闭，扩容额度已恢复')
    else ElMessage.success('已使用原退款单号重新提交，等待微信结果')
    await loadOrders()
  } catch (error) {
    if (error !== 'cancel' && error !== 'close') errorMessage.value = getApiErrorMessage(error)
  }
}
onMounted(() => { loadProducts(); loadOrders(); loadReward() })
</script>
<style scoped>.card-header { display:flex; align-items:center; justify-content:space-between; gap:16px; }.order-tools { display:flex; align-items:center; justify-content:flex-end; gap:10px; flex-wrap:wrap; }.order-tools :deep(.el-radio-group) { display:flex; }.review-actions { display:flex; justify-content:center; flex-wrap:wrap; }.el-pagination { justify-content:flex-end; margin-top:16px; }.reward-form { margin-top:18px; max-width:860px; }.tier-list { display:grid; gap:10px; }.tier-row { display:flex; align-items:center; flex-wrap:wrap; gap:8px; }</style>
