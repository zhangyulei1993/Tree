import { apiClient, apiMode, unwrapData } from '@/api/client'

export interface AdminCapacityProduct { productKey: string; displayName: string; addedMembers: number; amountFen: number; refundWindowMinutes: number; status?: string }
export interface AdminCapacityOrder { orderNo: string; familyId: number; buyerUserId: number; productName: string; addedMembers: number; amountFen: number; status: string; transactionId?: string; createdAt: string; expiresAt?: string; paidAt?: string; closedAt?: string; refundNo?: string; refundStatus?: string; refundRequestedAt?: string; refundReason?: string; refundAttemptNo?: number; refundAttemptsRemaining?: number }
export interface AdminCapacityOrderPage { items: AdminCapacityOrder[]; total: number; page: number; pageSize: number }

export async function listCapacityProducts() {
  if (apiMode === 'mock') return [] as AdminCapacityProduct[]
  return unwrapData<AdminCapacityProduct[]>(await apiClient.get('/admin/capacity-products'))
}
export async function updateCapacityProduct(productKey: string, payload: { displayName: string; addedMembers: number; amountFen: number; refundWindowMinutes: number; status: string }) {
  return unwrapData<AdminCapacityProduct>(await apiClient.put(`/admin/capacity-products/${encodeURIComponent(productKey)}`, payload))
}
export async function listCapacityOrders(params: { page: number; pageSize: number; status?: string; startDate?: string; endDate?: string }) {
  if (apiMode === 'mock') return { items: [], total: 0, page: params.page, pageSize: params.pageSize } as AdminCapacityOrderPage
  return unwrapData<AdminCapacityOrderPage>(await apiClient.get('/admin/capacity-orders', { params }))
}

export async function reconcileCapacityPayment(orderNo: string) {
  return unwrapData<{ status: string }>(await apiClient.post(`/admin/capacity-orders/${encodeURIComponent(orderNo)}/reconcile`))
}

export async function approveCapacityRefund(refundNo: string, reason = '') {
  return unwrapData<{ status: string }>(await apiClient.post(`/admin/capacity-refunds/${encodeURIComponent(refundNo)}/approve`, { reason }))
}

export async function rejectCapacityRefund(refundNo: string, reason: string) {
  return unwrapData<{ status: string }>(await apiClient.post(`/admin/capacity-refunds/${encodeURIComponent(refundNo)}/reject`, { reason }))
}

export async function reconcileCapacityRefund(refundNo: string) {
  return unwrapData<{ status: string }>(await apiClient.post(`/admin/capacity-refunds/${encodeURIComponent(refundNo)}/resolve`))
}

export async function retryCapacityRefund(refundNo: string) {
  return unwrapData<{ status: string }>(await apiClient.post(`/admin/capacity-refunds/${encodeURIComponent(refundNo)}/retry`))
}
