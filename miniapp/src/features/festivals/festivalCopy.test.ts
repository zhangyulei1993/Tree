import assert from 'node:assert/strict'
import { test } from 'node:test'

import { festivalItems2026 } from './festivalData'
import {
  FESTIVAL_COPY_VARIANT_COUNT,
  festivalCopySceneOptions,
  generateFestivalCopy
} from './festivalCopy'

test('generates non-empty copy for every festival scene', () => {
  for (const festival of festivalItems2026) {
    for (const scene of festivalCopySceneOptions) {
      const copies = Array.from({ length: FESTIVAL_COPY_VARIANT_COUNT }, (_, index) =>
        generateFestivalCopy(festival, scene.value, index)
      )
      assert.equal(new Set(copies).size, FESTIVAL_COPY_VARIANT_COUNT)
      for (const copy of copies) {
        assert.ok(copy.length > 8)
        assert.match(copy, new RegExp(festival.name))
      }
      assert.equal(generateFestivalCopy(festival, scene.value, FESTIVAL_COPY_VARIANT_COUNT), copies[0])
    }
  }
})

test('keeps the three copy scenes meaningfully different', () => {
  const festival = festivalItems2026.find((item) => item.name === '中秋节')
  assert.ok(festival)

  const personalCopy = generateFestivalCopy(festival, 'personal')
  const groupCopy = generateFestivalCopy(festival, 'group')
  const momentsCopy = generateFestivalCopy(festival, 'moments')

  assert.notEqual(personalCopy, groupCopy)
  assert.notEqual(groupCopy, momentsCopy)
  assert.doesNotMatch(personalCopy, /亲爱的|家人们|领导|同事/)
  assert.doesNotMatch(groupCopy, /亲爱的|家人们|领导|同事/)
  assert.match(momentsCopy, /中秋节到了/)
})
