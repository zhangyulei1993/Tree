import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'

const featureDir = path.dirname(fileURLToPath(import.meta.url))
const settingsSource = readFileSync(path.join(featureDir, '../../pages/family/settings.vue'), 'utf8')

test('quality-family display uses explicit commands instead of an optimistic native switch', () => {
  assert.match(settingsSource, /family\.publicDisplayEnabled \? '关闭展示' : '开启展示'/)
  assert.match(settingsSource, /@click="family\.publicDisplayEnabled \? confirmClosePublic\(\) : enablePublicDisplay\(\)"/)
  assert.doesNotMatch(settingsSource, /onPublicDisplayToggle/)
})

test('closing quality-family display only calls the API after confirmation', () => {
  const closeApiIndex = settingsSource.indexOf('await closePublicFamily(')
  const confirmGuardIndex = settingsSource.indexOf('if (!result.confirm)', settingsSource.indexOf("title: '关闭优质家庭展示'"))

  assert.ok(confirmGuardIndex >= 0)
  assert.ok(closeApiIndex > confirmGuardIndex)
})

test('switch values do not treat the string false as enabled', () => {
  assert.match(settingsSource, /return value === true \|\| value === 'true' \|\| value === 1 \|\| value === '1'/)
})
