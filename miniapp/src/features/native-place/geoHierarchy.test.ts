import assert from 'node:assert/strict'
import test from 'node:test'

import chinaTownshipData from './china-township-data.json'
import { formatFamilyRegionText, nativePlaceQueryText, resolveNativePlace, resolveNativePlaceKeyword, searchNativePlaceKeywords } from './geoHierarchy'

test('municipality uses province and district without repeating the city layer', () => {
  const result = resolveNativePlace({ province: '北京市', city: '北京市', district: '朝阳区' })
  assert.ok(result)
  assert.equal(result.nativePlace, '北京市朝阳区')
  assert.equal(result.regionText, '华北地区 · 北京市 · 朝阳区')
  assert.equal(result.kind, 'MUNICIPALITY_DISTRICT')
})

test('province city county preserves all administrative levels', () => {
  const result = resolveNativePlace({ province: '浙江省', city: '杭州市', district: '桐庐县' })
  assert.ok(result)
  assert.equal(result.nativePlace, '浙江省杭州市桐庐县')
  assert.equal(result.regionText, '华东地区 · 浙江省 · 杭州市 · 桐庐县')
  assert.equal(result.kind, 'PROVINCE_CITY_COUNTY')
})

test('keyword parser resolves complete municipality and province city district values', () => {
  const beijing = resolveNativePlaceKeyword('北京市朝阳区')
  const hangzhou = resolveNativePlaceKeyword('浙江省杭州市西湖区')
  const jiayu = resolveNativePlaceKeyword('嘉鱼县')
  assert.equal(beijing?.nativePlace, '北京市朝阳区')
  assert.equal(hangzhou?.nativePlace, '浙江省杭州市西湖区')
  assert.equal(hangzhou?.kind, 'PROVINCE_CITY_DISTRICT')
  assert.equal(jiayu?.regionText, '华中地区 · 湖北省 · 咸宁市 · 嘉鱼县')
})

test('township lookup resolves to its county, city, province and geographic zone', () => {
  const jiayuTownships = ['陆溪镇', '高铁岭镇', '官桥镇', '鱼岳镇', '新街镇', '渡普镇', '潘家湾镇', '牌洲湾镇', '头墩农场', '嘉鱼县经济开发区']
  jiayuTownships.forEach((township) => {
    const result = resolveNativePlaceKeyword(`嘉鱼县${township}`)
    assert.ok(result, `${township} should resolve`)
    assert.equal(nativePlaceQueryText(result), `华中地区 · 湖北省 · 咸宁市 · 嘉鱼县 · ${township}`)
    assert.equal(result.nativePlace, '湖北省咸宁市嘉鱼县')
  })
})

test('township suffix is optional for both bare and county-qualified input', () => {
  const bare = resolveNativePlaceKeyword('潘家湾')
  const qualified = resolveNativePlaceKeyword('嘉鱼县潘家湾')
  const shortNameCandidates = searchNativePlaceKeywords('官桥')
  const variant = resolveNativePlaceKeyword('簰洲湾')
  assert.equal(nativePlaceQueryText(bare!), '华中地区 · 湖北省 · 咸宁市 · 嘉鱼县 · 潘家湾镇')
  assert.equal(nativePlaceQueryText(qualified!), '华中地区 · 湖北省 · 咸宁市 · 嘉鱼县 · 潘家湾镇')
  assert.ok(shortNameCandidates.length > 1)
  assert.ok(shortNameCandidates.some((item) => nativePlaceQueryText(item) === '华中地区 · 湖北省 · 咸宁市 · 嘉鱼县 · 官桥镇'))
  assert.equal(nativePlaceQueryText(variant!), '华中地区 · 湖北省 · 咸宁市 · 嘉鱼县 · 牌洲湾镇')
})

test('reused township names return every candidate for the user to choose', () => {
  assert.equal(resolveNativePlaceKeyword('周庄镇'), null)
  assert.equal(resolveNativePlaceKeyword('官庄镇'), null)

  const candidates = searchNativePlaceKeywords('周庄镇')
  assert.ok(candidates.length > 1)
  assert.ok(candidates.some((item) => nativePlaceQueryText(item) === '华中地区 · 河南省 · 平顶山市 · 宝丰县 · 周庄镇'))

  const official = resolveNativePlaceKeyword('天津市蓟州区官庄镇')
  assert.equal(nativePlaceQueryText(official!), '华北地区 · 天津市 · 蓟州区 · 官庄镇')
})

test('every bundled township can be queried by its exact name', () => {
  let checked = 0
  Object.entries(chinaTownshipData as Record<string, Record<string, Record<string, string[]>>>).forEach(([province, cities]) => {
    Object.entries(cities).forEach(([city, districts]) => {
      Object.entries(districts).forEach(([district, townships]) => {
        townships.forEach((township) => {
          const candidates = searchNativePlaceKeywords(township)
          const normalizedCity = city === '市辖区' || city === '县' ? '' : city
          const lowest = district && district !== province && district !== city ? district : ''
          const expected = [province, normalizedCity, lowest, township]
            .filter(Boolean)
            .join(' · ')
          assert.ok(
            candidates.some((item) => nativePlaceQueryText(item).endsWith(expected)),
            `${township} should include ${expected}`
          )
          checked += 1
        })
      })
    })
  })
  assert.equal(checked, 41352)
})

test('ambiguous keyword is not guessed', () => {
  assert.equal(resolveNativePlaceKeyword('朝阳区'), null)
})

test('legacy city-only family regions are enriched for display', () => {
  assert.equal(formatFamilyRegionText({ regionText: '咸宁市' }), '华中地区 · 湖北省 · 咸宁市')
  assert.equal(
    formatFamilyRegionText({ nativePlace: '湖北省咸宁市嘉鱼县' }),
    '华中地区 · 湖北省 · 咸宁市'
  )
  assert.equal(
    formatFamilyRegionText({ regionText: '华中地区 · 湖北省 · 咸宁市 · 嘉鱼县' }),
    '华中地区 · 湖北省 · 咸宁市'
  )
  assert.equal(
    formatFamilyRegionText({ nativePlace: '北京市朝阳区' }),
    '华北地区 · 北京市'
  )
})
