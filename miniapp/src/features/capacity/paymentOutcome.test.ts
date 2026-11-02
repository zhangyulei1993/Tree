import assert from 'node:assert/strict'
import test from 'node:test'

import { capacityPaymentFeedback, isCapacityPaymentConfirmed } from './paymentOutcome'

test('only settled capacity orders can be reported as paid', () => {
  for (const status of ['PAID', 'REFUNDING', 'REFUNDED']) {
    assert.equal(isCapacityPaymentConfirmed(status), true)
  }
  for (const status of ['PENDING', 'CANCELLED', 'CLOSED', '']) {
    assert.equal(isCapacityPaymentConfirmed(status), false)
  }
})

test('payment feedback distinguishes settled, cancelled, and pending orders', () => {
  assert.equal(capacityPaymentFeedback('PAID'), '支付成功，容量已增加')
  assert.equal(capacityPaymentFeedback('CANCELLED'), '订单未支付，可重新下单')
  assert.equal(capacityPaymentFeedback('PENDING'), '支付结果待确认，请到订单页查看')
})
