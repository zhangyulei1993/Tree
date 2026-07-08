import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'

const pageDir = path.dirname(fileURLToPath(import.meta.url))
const manageSource = readFileSync(path.join(pageDir, 'manage.vue'), 'utf8')

test('manage.vue imports useSessionStore from session store', () => {
  assert.match(manageSource, /import \{ useSessionStore \} from '@\/stores\/session'/)
  assert.match(manageSource, /const session = useSessionStore\(\)/)
})

test('manage.vue exposes three manage sections in template', () => {
  assert.match(manageSource, /manageSection === 'create'/)
  assert.match(manageSource, /manageSection === 'locate'/)
  assert.match(manageSource, /manageSection === 'relations'/)
  assert.match(manageSource, /创建并放入家谱/)
  assert.match(manageSource, /放入家谱/)
})

test('manage.vue wires parent memberType for create and place flows', () => {
  assert.match(manageSource, /parentRoleValues/)
  assert.match(manageSource, /showParentRolePicker/)
  assert.match(manageSource, /showPlaceParentRolePicker/)
  assert.match(manageSource, /unlocatedTypeAnomalies/)
  assert.match(manageSource, /filterUnlocatedLineageMemberIds/)
})

test('manage.vue uses session for auth gate', () => {
  assert.match(manageSource, /session\.requireLogin/)
})

test('manage.vue guards missing familyId before loading APIs', () => {
  const guardIndex = manageSource.indexOf('if (!familyId.value)')
  const routeIndex = manageSource.indexOf('const route = `/pages/family/manage')
  const firstApiIndex = manageSource.indexOf('getFamilyDetail(familyId.value)')
  assert.notEqual(guardIndex, -1)
  assert.ok(guardIndex < routeIndex)
  assert.ok(guardIndex < firstApiIndex)
})
