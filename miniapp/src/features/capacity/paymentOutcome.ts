export function isCapacityPaymentConfirmed(status: string) {
  return status === 'PAID' || status === 'REFUNDING' || status === 'REFUNDED'
}

export function capacityPaymentFeedback(status: string) {
  if (isCapacityPaymentConfirmed(status)) return '支付成功，容量已增加'
  if (status === 'CANCELLED' || status === 'CLOSED') return '订单未支付，可重新下单'
  return '支付结果待确认，请到订单页查看'
}
