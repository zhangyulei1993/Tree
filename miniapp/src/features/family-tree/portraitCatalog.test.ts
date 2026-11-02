import assert from 'node:assert/strict'
import test from 'node:test'

import {
  buildPortraitAppearanceMap,
  resolveFemalePortraitVariant,
  resolvePortraitAppearanceVariant,
  resolvePortraitAvatarStyle,
  resolvePortraitVariant,
  portraitAssetBaseURL,
  resolveRegionalPortraitAsset,
  resolveRegionalPortraitGlassesAsset,
  resolveSelectedPortraitAsset,
  resolveRegionalPortraitTheme
} from './portraitCatalog'

test('female portrait choices map to explicit bundled variants while auto stays stable', () => {
  assert.equal(resolveFemalePortraitVariant(15, 'BRAIDS', 'adult'), 1)
  assert.equal(resolveFemalePortraitVariant(15, 'LONG', 'adult'), 5)
  assert.equal(resolveFemalePortraitVariant(15, 'SHORT', 'adult'), 6)
  assert.equal(resolveFemalePortraitVariant(15, 'LONG', 'toddler'), 6)
  assert.equal(resolveFemalePortraitVariant(15, 'LONG', 'teen'), 5)
  assert.equal(resolveFemalePortraitVariant(15, 'LONG', 'youth'), 2)
  assert.equal(resolveFemalePortraitVariant(15, 'LONG', 'mature'), 5)
  assert.equal(resolveFemalePortraitVariant(15, 'LOW_BUN', 'senior'), 3)
  assert.equal(resolveFemalePortraitVariant(15, 'AUTO'), resolvePortraitVariant(15))
  assert.equal(
    resolveRegionalPortraitGlassesAsset({ theme: 'south', gender: 'female', ageBand: 'adult', variant: 5 }),
    `${portraitAssetBaseURL}/age-only/glasses/female-adult-5.png?v=age-only-editorial-20261014-112`
  )
})

test('complete portrait selections keep base and glasses as independent illustrations', () => {
  assert.deepEqual(resolvePortraitAvatarStyle('base_11'), { mode: 'base', variant: 11 })
  assert.deepEqual(resolvePortraitAvatarStyle('GLASSES_7'), { mode: 'glasses', variant: 7 })
  assert.equal(resolvePortraitAvatarStyle('GLASSES_13'), null)
  assert.equal(
    resolveSelectedPortraitAsset({ gender: 'male', ageBand: 'mature', style: 'BASE_11' }),
    `${portraitAssetBaseURL}/age-only/base/male-mature-11.png?v=age-only-editorial-20261014-112`
  )
  assert.equal(
    resolveSelectedPortraitAsset({ gender: 'female', ageBand: 'adult', style: 'GLASSES_7' }),
    `${portraitAssetBaseURL}/age-only/glasses/female-adult-7.png?v=age-only-editorial-20261014-112`
  )
})

test('base and glasses portrait URLs are permanently separated', () => {
  const base = resolveRegionalPortraitAsset({ theme: 'south', gender: 'female', ageBand: 'adult', variant: 2 })
  const glasses = resolveRegionalPortraitGlassesAsset({ theme: 'south', gender: 'female', ageBand: 'adult', variant: 2 })
  assert.match(base, /\/age-only\/base\/female-adult-2\.png\?v=/)
  assert.doesNotMatch(base, /age-only\/glasses/)
  assert.match(glasses, /\/age-only\/glasses\/female-adult-2\.png\?v=/)
  assert.notEqual(base, glasses)
})

test('regional portraits collapse hometowns into north and south styles', () => {
  assert.equal(resolveRegionalPortraitTheme('CN-NORTHEAST-LN', '辽宁省'), 'north')
  assert.equal(resolveRegionalPortraitTheme('CN-NORTHWEST-XJ', '新疆维吾尔自治区'), 'north')
  assert.equal(resolveRegionalPortraitTheme('CN-CENTRAL-HB', '湖北省'), 'south')
  assert.equal(resolveRegionalPortraitTheme('CN-EAST-ZJ', '浙江省'), 'south')
  assert.equal(resolveRegionalPortraitTheme(undefined, '山东省'), 'north')
  assert.equal(resolveRegionalPortraitTheme(undefined, '广西壮族自治区'), 'south')
})

test('portrait appearances cover a dozen stable sibling variations before repeating', () => {
  assert.deepEqual(
    Array.from({ length: 12 }, (_, index) => resolvePortraitAppearanceVariant(index)),
    [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12]
  )
  assert.equal(resolvePortraitVariant(0), 1)
  assert.equal(resolvePortraitVariant(1), 2)
  assert.equal(resolvePortraitVariant(2), 3)
  assert.equal(resolvePortraitVariant(3), 4)
  assert.equal(resolvePortraitVariant(4), 5)
  assert.equal(resolvePortraitVariant(5), 6)
  assert.equal(resolvePortraitVariant(6), 7)
  assert.equal(resolvePortraitAppearanceVariant(43), 8)
  assert.equal(resolvePortraitVariant(43), 8)
  assert.equal(
    resolveRegionalPortraitAsset({ theme: 'south', gender: 'female', ageBand: 'toddler', memberId: 43 }),
    `${portraitAssetBaseURL}/age-only/base/female-toddler-8.png?v=age-only-editorial-20261014-112`
  )
})

test('same-age peers receive distinct appearances even when member IDs collide modulo twelve', () => {
  const appearances = buildPortraitAppearanceMap([
    { memberId: 557, gender: 'FEMALE', birthDate: '1993-01-01', isLiving: true },
    { memberId: 564, gender: 'FEMALE', birthDate: '1987-01-01', isLiving: true },
    { memberId: 569, gender: 'FEMALE', birthDate: '1990-01-01', isLiving: true }
  ])

  assert.deepEqual(appearances, { 557: 6, 564: 1, 569: 7 })
  assert.deepEqual(
    Object.values(appearances).map((appearance) => resolvePortraitVariant(appearance - 1)),
    [6, 1, 7]
  )
})
