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

test('manage.vue separates node and relationship flows', () => {
  assert.match(manageSource, /manageSection === 'node'/)
  assert.match(manageSource, /manageSection === 'relations'/)
  assert.match(manageSource, /添加节点/)
  assert.match(manageSource, /创建并接入家庭树/)
  assert.match(manageSource, /requestedSection === 'create'/)
  assert.match(manageSource, /requestedSection === 'node'/)
})

test('manage.vue wires parent memberType only while directly creating a relationship', () => {
  assert.match(manageSource, /parentRoleValues/)
  assert.match(manageSource, /showParentRolePicker/)
  assert.match(manageSource, /父母身份/)
  assert.match(manageSource, /defaultParentRoleIndex/)
  assert.match(manageSource, /hasLineageParentByGender/)
  assert.match(manageSource, /家庭成员的配偶/)
  assert.doesNotMatch(manageSource, /showPlaceParentRolePicker|unlocatedTypeAnomalies|filterUnlocatedLineageMemberIds/)
})

test('manage.vue refreshes and preserves selected base member by id', () => {
  assert.match(manageSource, /onShow/)
  assert.match(manageSource, /selectedBaseMemberId/)
  assert.match(manageSource, /syncCreateBaseSelection/)
  assert.match(manageSource, /String\(selectedBaseMember\.value\?\.memberId/)
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
