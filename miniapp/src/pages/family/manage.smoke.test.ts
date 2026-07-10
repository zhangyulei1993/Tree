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

test('manage.vue separates node, pending and placement flows', () => {
  assert.match(manageSource, /manageSection === 'node'/)
  assert.match(manageSource, /manageSection === 'pending'/)
  assert.match(manageSource, /manageSection === 'place'/)
  assert.match(manageSource, /manageSection === 'relations'/)
  assert.match(manageSource, /添加节点/)
  assert.match(manageSource, /添加暂存/)
  assert.match(manageSource, /暂存转化/)
  assert.match(manageSource, /创建并接入家庭树/)
  assert.match(manageSource, /接入家庭树/)
  assert.match(manageSource, /requestedSection === 'create'/)
  assert.match(manageSource, /requestedSection === 'locate'/)
})

test('manage.vue wires parent memberType for create and place flows', () => {
  assert.match(manageSource, /parentRoleValues/)
  assert.match(manageSource, /showParentRolePicker/)
  assert.match(manageSource, /showPlaceParentRolePicker/)
  assert.match(manageSource, /新成员归属/)
  assert.match(manageSource, /defaultParentRoleIndex/)
  assert.match(manageSource, /hasLineageParentByGender/)
  assert.match(manageSource, /家庭成员的配偶/)
  assert.match(manageSource, /unlocatedTypeAnomalies/)
  assert.match(manageSource, /filterUnlocatedLineageMemberIds/)
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
