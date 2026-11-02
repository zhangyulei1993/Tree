import { isRealApiMode, request } from './client'
import { isCapacityPaymentConfirmed } from '@/features/capacity/paymentOutcome'

export interface CapacityProduct { productKey: string; displayName: string; addedMembers: number; amountFen: number; refundWindowMinutes: number }
export interface CapacityOrder { orderNo: string; familyId: number | string; productName: string; addedMembers: number; amountFen: number; status: string; createdAt?: string; expiresAt?: string; closedAt?: string; refundableUntil?: string; refundAttemptCount?: number; refundAttemptsRemaining?: number; lastRefundStatus?: string; lastRefundReason?: string; payment?: { timeStamp: string; nonceStr: string; package: string; signType: string; paySign: string } }
export interface CapacityOrderPage { items: CapacityOrder[]; total: number; page: number; pageSize: number }

export class CapacityPaymentPendingError extends Error {
  readonly order: CapacityOrder

  constructor(message: string, order: CapacityOrder) {
    super(message)
    this.name = 'CapacityPaymentPendingError'
    this.order = order
  }
}

export function listCapacityProducts() { return request<CapacityProduct[]>('/capacity-products', { public: true }) }
export function createCapacityOrder(familyId: number | string, productKey: string) { return request<CapacityOrder>(`/families/${familyId}/capacity/orders`, { method: 'POST', data: { productKey } }) }
export function requestCapacityRefund(familyId: number | string, orderNo: string, reason = '') { return request<{ status: string }>(`/families/${familyId}/capacity/orders/${encodeURIComponent(orderNo)}/refund`, { method: 'POST', data: { reason } }) }
export async function listCapacityOrders(familyId: number | string) { return (await listCapacityOrdersPage(familyId, 1, 50)).items }
export function listCapacityOrdersPage(familyId: number | string, page = 1, pageSize = 20, query: { status?: string; startDate?: string; endDate?: string } = {}) {
  const params = [`page=${page}`, `pageSize=${pageSize}`]
  if (query.status && query.status !== 'ALL') params.push(`status=${encodeURIComponent(query.status)}`)
  if (query.startDate) params.push(`startDate=${encodeURIComponent(query.startDate)}`)
  if (query.endDate) params.push(`endDate=${encodeURIComponent(query.endDate)}`)
  return request<CapacityOrderPage>(`/families/${familyId}/capacity/orders?${params.join('&')}`)
}
export function reconcileCapacityPayment(familyId: number | string, orderNo: string) { return request<{ status: string }>(`/families/${familyId}/capacity/orders/${encodeURIComponent(orderNo)}/reconcile`, { method: 'POST' }) }
export function resumeCapacityPayment(familyId: number | string, orderNo: string) { return request<CapacityOrder>(`/families/${familyId}/capacity/orders/${encodeURIComponent(orderNo)}/pay`, { method: 'POST' }) }

async function waitForCapacityOrder(familyId: number | string, orderNo: string) {
  let latest: CapacityOrder | undefined
  for (let attempt = 0; attempt < 4; attempt += 1) {
    try {
      latest = (await listCapacityOrders(familyId)).find((item) => item.orderNo === orderNo)
      if (latest && isCapacityPaymentConfirmed(latest.status)) return latest
    } catch {
      // The callback or network request may still be in flight.
    }
    if (attempt < 3) await new Promise((wait) => setTimeout(wait, 800))
  }
  return latest
}

async function reconcileThenReadOrder(familyId: number | string, orderNo: string) {
  try {
    const result = await reconcileCapacityPayment(familyId, orderNo)
    if (isCapacityPaymentConfirmed(result.status)) return (await listCapacityOrders(familyId)).find((item) => item.orderNo === orderNo)
  } catch {
    // The order list remains available even when the merchant query is delayed.
  }
  return undefined
}

async function requestCapacityPayment(familyId: number | string, order: CapacityOrder) {
  if (!isRealApiMode || !order.payment) return order
  await new Promise<void>((resolve, reject) => {
    uni.requestPayment({ ...order.payment!, success: () => resolve(), fail: async (error) => {
      // WeChat can report `cancel` when the user does not return to the merchant
      // page even though payment has already succeeded. Re-read the order before
      // treating the result as a real cancellation.
      const latest = await waitForCapacityOrder(familyId, order.orderNo) || await reconcileThenReadOrder(familyId, order.orderNo)
      if (latest && isCapacityPaymentConfirmed(latest.status)) return resolve()
      const message = error.errMsg?.includes('cancel') ? '支付结果待确认，请在订单中查询结果' : (error.errMsg || '支付未完成')
      reject(new CapacityPaymentPendingError(message, latest || order))
    } })
  })
  const latest = await waitForCapacityOrder(familyId, order.orderNo) || await reconcileThenReadOrder(familyId, order.orderNo)
  return latest || order
}

export async function payCapacityOrder(familyId: number | string, productKey: string) {
  return requestCapacityPayment(familyId, await createCapacityOrder(familyId, productKey))
}

export async function resumeCapacityOrderPayment(familyId: number | string, orderNo: string) {
  return requestCapacityPayment(familyId, await resumeCapacityPayment(familyId, orderNo))
}
