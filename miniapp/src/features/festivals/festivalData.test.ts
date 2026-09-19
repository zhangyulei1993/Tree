import assert from 'node:assert/strict'
import test from 'node:test'

import { getFestivalItems, getSolarTermItems } from './festivalData'

test('provides complete 2027 festival data', () => {
  const festivals = getFestivalItems(2027)

  assert.equal(festivals.length, 13)
  assert.equal(festivals.find((item) => item.name === '春节')?.date, '2027-02-06')
  assert.equal(festivals.find((item) => item.name === '中秋节')?.date, '2027-09-15')
  assert.ok(festivals.every((item) => item.date.startsWith('2027-')))
})

test('provides all twenty-four solar terms for 2027', () => {
  const solarTerms = getSolarTermItems(2027)

  assert.equal(solarTerms.length, 24)
  assert.equal(solarTerms.find((item) => item.name === '立春')?.date, '2027-02-04')
  assert.equal(solarTerms.find((item) => item.name === '冬至')?.date, '2027-12-22')
  assert.ok(solarTerms.every((item) => item.date.startsWith('2027-')))
})

test('returns copies so callers cannot alter the calendar source data', () => {
  const firstRead = getFestivalItems(2027)
  firstRead[0]!.name = '已修改'

  assert.equal(getFestivalItems(2027)[0]?.name, '腊八节')
})
